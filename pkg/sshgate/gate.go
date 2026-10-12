// Package sshgate is the TalesMUD SSH listener. It is generic: no world names,
// no password auth, and no listener at all unless ssh.enabled is set.
package sshgate

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/talesmud/talesmud/pkg/devicecode"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/presentation/textline"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/sshkeys"
)

const (
	serverVersion = "SSH-2.0-TalesMUD"
	// closeSessionReplaced matches the websocket close used when a newer
	// session takes the same account.
	closeSessionReplaced = 4001
	queueSize            = 128
)

// UserLookup is the account read the gate needs. *service users satisfy it.
type UserLookup interface {
	FindByID(id string) (*entities.User, error)
	FindByRefID(refID string) (*entities.User, error)
}

// Deps are the engine hooks. Nil guests or users fail closed.
// Door selects the 80x25 frame renderer. Classic play is the default.
// Keys, Pending, and Devices stay nil unless that feature is enabled.
type Deps struct {
	Mud     mudserver.MUDServer
	Guests  service.GuestService
	Users   UserLookup
	Door    bool
	Keys    *sshkeys.Store
	Pending *sshkeys.Pending
	Devices *devicecode.Store
}

// PublicInfo is the unauthenticated GET /api/ssh/info body.
type PublicInfo struct {
	Enabled             bool     `json:"enabled"`
	Host                string   `json:"host,omitempty"`
	Port                int      `json:"port,omitempty"`
	HostKeyFingerprints []string `json:"host_key_fingerprints,omitempty"`
	GuestEnabled        bool     `json:"guest_enabled"`
	ActivateURL         string   `json:"activate_url,omitempty"`
}

// Gate is one SSH listener. Close stops new connections.
type Gate struct {
	cfg     gamemode.SSHConfig
	deps    Deps
	enabled bool
	info    PublicInfo
	limits  *limits
	ln      net.Listener
	sshConf *ssh.ServerConfig
	ctx     context.Context
	cancel  context.CancelFunc
	playURL string
	// door selects the 80x25 frame renderer. Classic play is the default.
	door    bool
	once    sync.Once
	notesMu sync.Mutex
	notes   map[string]*authNote
	sessMu  sync.Mutex
	live    map[*liveSession]struct{}
}

// Listen starts the SSH listener when cfg.Enabled. A disabled config returns
// a gate that reports enabled:false and does not bind.
func Listen(cfg gamemode.SSHConfig, deps Deps) (*Gate, error) {
	cfg, err := prepare(cfg)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	g := &Gate{
		cfg:     cfg,
		deps:    deps,
		enabled: cfg.Enabled,
		limits:  newLimits(cfg),
		ctx:     ctx,
		cancel:  cancel,
		playURL: playURL(cfg),
		door:    deps.Door,
		notes:   map[string]*authNote{},
	}
	if !cfg.Enabled {
		return g, nil
	}
	signer, fp, err := LoadOrCreateHostKey(cfg.HostKeyPath)
	if err != nil {
		cancel()
		return nil, err
	}
	g.info = PublicInfo{
		Enabled:             true,
		Host:                cfg.PublicHost,
		Port:                cfg.PublicPort,
		HostKeyFingerprints: []string{fp},
		GuestEnabled:        cfg.Guest.Enabled,
		ActivateURL:         cfg.Device.ActivateURL,
	}
	g.sshConf = g.serverConfig(signer)
	ln, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		cancel()
		return nil, err
	}
	g.ln = ln
	go g.accept()
	// The host-key fingerprint is public server identity. User key
	// fingerprints stay masked; this one is printed once at startup.
	log.WithFields(log.Fields{
		"listen":      ln.Addr().String(),
		"fingerprint": fp,
	}).Info("ssh listening")
	return g, nil
}

// PublicInfo returns the public listener description.
func (g *Gate) PublicInfo() PublicInfo {
	if g == nil || !g.enabled {
		return PublicInfo{}
	}
	return g.info
}

// Addr is the bound address, including an ephemeral port.
// CloseUser ends live SSH sessions for this user id in this process.
func (g *Gate) CloseUser(id string) {
	g.closeLive(id, "")
}

// CloseFingerprint ends live SSH sessions that authenticated with this key.
func (g *Gate) CloseFingerprint(fp string) {
	g.closeLive("", fp)
}

func (g *Gate) closeLive(id, fp string) {
	if g == nil {
		return
	}
	id = strings.TrimSpace(id)
	fp = strings.TrimSpace(fp)
	if id == "" && fp == "" {
		return
	}
	g.sessMu.Lock()
	hit := make([]*liveSession, 0, 1)
	for s := range g.live {
		s.mu.Lock()
		match := (id != "" && s.userID == id) || (fp != "" && s.fp == fp)
		s.mu.Unlock()
		if match {
			hit = append(hit, s)
		}
	}
	g.sessMu.Unlock()
	for _, s := range hit {
		s.revoke()
	}
}

func (g *Gate) track(s *liveSession) {
	if g == nil || s == nil {
		return
	}
	g.sessMu.Lock()
	if g.live == nil {
		g.live = map[*liveSession]struct{}{}
	}
	g.live[s] = struct{}{}
	g.sessMu.Unlock()
}

func (g *Gate) untrack(s *liveSession) {
	if g == nil || s == nil {
		return
	}
	g.sessMu.Lock()
	delete(g.live, s)
	g.sessMu.Unlock()
}

func (g *Gate) Addr() string {
	if g == nil || g.ln == nil {
		return ""
	}
	return g.ln.Addr().String()
}

// Close stops the listener. Existing sessions end when their connection closes.
func (g *Gate) Close() error {
	if g == nil {
		return nil
	}
	g.cancel()
	if g.ln == nil {
		return nil
	}
	return g.ln.Close()
}

func (g *Gate) accept() {
	for {
		conn, err := g.ln.Accept()
		if err != nil {
			if g.ctx.Err() != nil {
				return
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}
		go g.handle(conn)
	}
}

func (g *Gate) handle(conn net.Conn) {
	defer conn.Close()
	ip := remoteIP(conn)
	if err := g.limits.acquire(ip, time.Now()); err != nil {
		log.WithFields(log.Fields{"ip": ip, "reason": err.Error()}).Info("ssh rejected")
		return
	}
	released := false
	defer func() {
		if !released {
			g.limits.release(ip)
		}
	}()
	timeout := g.cfg.AuthTimeout.Duration()
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	_ = conn.SetDeadline(time.Now().Add(timeout))
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, g.sshConf)
	_ = conn.SetDeadline(time.Time{})
	if err != nil {
		g.limits.authFail(ip, time.Now())
		log.WithFields(log.Fields{"ip": ip, "ok": false}).Info("ssh auth")
		return
	}
	log.WithFields(log.Fields{
		"ip":     ip,
		"method": methodOf(sshConn),
		"user":   clipUser(sshConn.User()),
		"ok":     true,
	}).Info("ssh auth")
	released = true
	g.serve(conn, sshConn, chans, reqs, ip)
}

func (g *Gate) serve(conn net.Conn, sshConn *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request, ip string) {
	defer g.limits.release(ip)
	defer sshConn.Close()
	go rejectGlobals(ip, reqs)
	sessions := 0
	for ch := range chans {
		kind := ch.ChannelType()
		if kind != "session" || sessions >= 1 {
			log.WithFields(log.Fields{"ip": ip, "type": kind}).Info("ssh channel rejected")
			_ = ch.Reject(ssh.Prohibited, "only one terminal session is available")
			continue
		}
		channel, requests, err := ch.Accept()
		if err != nil {
			continue
		}
		sessions++
		sess := newSession(g, conn, channel, ip)
		sess.applyAuth(sshConn)
		go sess.requests(requests)
	}
}

func (g *Gate) noneAuth(meta ssh.ConnMetadata) (*ssh.Permissions, error) {
	if !g.cfg.Guest.Enabled || meta.User() != "guest" {
		return nil, errors.New("authentication required")
	}
	return &ssh.Permissions{Extensions: map[string]string{"via": "guest"}}, nil
}

func rejectGlobals(ip string, reqs <-chan *ssh.Request) {
	for req := range reqs {
		switch req.Type {
		case "tcpip-forward", "cancel-tcpip-forward":
			log.WithFields(log.Fields{"ip": ip, "type": req.Type}).Info("ssh request rejected")
		}
		if req.WantReply {
			_ = req.Reply(false, nil)
		}
	}
}

func methodOf(conn *ssh.ServerConn) string {
	if conn == nil || conn.Permissions.Extensions == nil {
		return ""
	}
	return conn.Permissions.Extensions["via"]
}

func remoteIP(conn net.Conn) string {
	if conn == nil {
		return ""
	}
	return remoteHost(conn.RemoteAddr())
}

func remoteHost(addr net.Addr) string {
	if addr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	return host
}

func clipUser(name string) string {
	name = textline.Sanitize(strings.TrimSpace(name))
	if len(name) > 64 {
		name = name[:64]
	}
	return name
}

func playURL(cfg gamemode.SSHConfig) string {
	host := strings.TrimSpace(cfg.PublicHost)
	if host == "" {
		return "/play"
	}
	return "https://" + host + "/play"
}

func prepare(cfg gamemode.SSHConfig) (gamemode.SSHConfig, error) {
	cfg.Listen = strings.TrimSpace(cfg.Listen)
	cfg.HostKeyPath = strings.TrimSpace(cfg.HostKeyPath)
	cfg.PublicHost = strings.TrimSpace(cfg.PublicHost)
	cfg.Device.ActivateURL = strings.TrimSpace(cfg.Device.ActivateURL)
	if !cfg.Enabled {
		return cfg, nil
	}
	if cfg.Listen == "" {
		return cfg, errors.New("ssh enabled but listen is empty")
	}
	if cfg.HostKeyPath == "" {
		return cfg, errors.New("ssh enabled but host_key_path is empty")
	}
	if cfg.MaxConnections <= 0 {
		cfg.MaxConnections = 100
	}
	if cfg.MaxPerIP <= 0 {
		cfg.MaxPerIP = 5
	}
	if cfg.NewConnsPerIPPerMin <= 0 {
		cfg.NewConnsPerIPPerMin = 20
	}
	if cfg.AuthFailBan.Failures <= 0 {
		cfg.AuthFailBan.Failures = 10
	}
	if cfg.AuthFailBan.Window.Duration() <= 0 {
		cfg.AuthFailBan.Window = gamemode.Duration(10 * time.Minute)
	}
	if cfg.AuthFailBan.Ban.Duration() <= 0 {
		cfg.AuthFailBan.Ban = gamemode.Duration(15 * time.Minute)
	}
	if cfg.AuthTimeout.Duration() <= 0 {
		cfg.AuthTimeout = gamemode.Duration(30 * time.Second)
	}
	if cfg.IdleTimeout.Duration() <= 0 {
		cfg.IdleTimeout = gamemode.Duration(30 * time.Minute)
	}
	if cfg.Guest.MaxConcurrent <= 0 {
		cfg.Guest.MaxConcurrent = 10
	}
	if cfg.Guest.PerIPPerHour <= 0 {
		cfg.Guest.PerIPPerHour = 5
	}
	if cfg.Guest.MaxSession.Duration() <= 0 {
		cfg.Guest.MaxSession = gamemode.Duration(30 * time.Minute)
	}
	if cfg.Mud.History <= 0 {
		cfg.Mud.History = 20
	}
	return cfg, nil
}

func (g *Gate) serverConfig(signer ssh.Signer) *ssh.ServerConfig {
	guest := g.cfg.Guest.Enabled
	cfg := &ssh.ServerConfig{
		ServerVersion: serverVersion,
		MaxAuthTries:  6,
		NoClientAuth:  guest,
		PublicKeyAuthAlgorithms: []string{
			ssh.KeyAlgoED25519,
			ssh.KeyAlgoSKED25519,
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoRSASHA256,
			ssh.KeyAlgoRSASHA512,
		},
		Config: ssh.Config{
			KeyExchanges: []string{
				ssh.KeyExchangeMLKEM768X25519,
				ssh.KeyExchangeCurve25519,
			},
			Ciphers: []string{
				ssh.CipherChaCha20Poly1305,
				ssh.CipherAES128GCM,
				ssh.CipherAES256GCM,
			},
			MACs: []string{
				ssh.HMACSHA256ETM,
				ssh.HMACSHA512ETM,
				ssh.HMACSHA256,
				ssh.HMACSHA512,
			},
		},
		AuthLogCallback: func(meta ssh.ConnMetadata, method string, err error) {
			// method and ok only. The error text and the key stay out of the log.
			log.WithFields(log.Fields{
				"ip":     remoteHost(meta.RemoteAddr()),
				"method": method,
				"user":   clipUser(meta.User()),
				"ok":     err == nil,
			}).Info("ssh auth method")
		},
	}
	if guest {
		cfg.NoClientAuthCallback = g.noneAuth
	}
	// Callbacks stay nil unless the flag and the store are both set, so a
	// guest-only listener does not advertise publickey or keyboard-interactive.
	if g.cfg.Keys.Enabled && g.deps.Keys != nil {
		cfg.PublicKeyCallback = g.publicKeyAuth
		cfg.VerifiedPublicKeyCallback = g.verifiedKeyAuth
	}
	if g.cfg.Device.Enabled && g.deps.Devices != nil {
		cfg.KeyboardInteractiveCallback = g.deviceAuth
	}
	cfg.AddHostKey(signer)
	return cfg
}
