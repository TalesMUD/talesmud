package sshgate

import (
	"encoding/hex"
	"errors"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/presentation/textline"
)

// authNote is the unverified key material for one handshake.
// The public key line is stored only after the signature is verified.
type authNote struct {
	fps  []string
	line string
	at   time.Time
}

func (s *liveSession) applyAuth(conn *ssh.ServerConn) {
	if s == nil || conn == nil {
		return
	}
	if conn.Permissions.Extensions != nil {
		s.via = conn.Permissions.Extensions["via"]
		s.userRef = conn.Permissions.Extensions["user"]
		s.fp = conn.Permissions.Extensions["fp"]
		s.linkID = conn.Permissions.Extensions["link_pending"]
	}
	note := s.gate.takeNote(conn)
	s.offers = note.fps
	s.keyLine = note.line
	s.clientVersion = clipUser(string(conn.ClientVersion()))
}

// publicKeyAuth decides whether a signature is worth checking.
// An unknown key returns a nil error so the library requires a signature.
// The fingerprint is recorded only in verifiedKeyAuth, after that signature
// checks. A query that is never signed is not an offer.
func (g *Gate) publicKeyAuth(meta ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	if g == nil || key == nil || g.deps.Keys == nil {
		return nil, errors.New("unknown key")
	}
	if g.cfg.Guest.Enabled && meta != nil && meta.User() == "guest" {
		return nil, errors.New("authentication required")
	}
	fp := ssh.FingerprintSHA256(key)
	linked, err := g.deps.Keys.ByFingerprint(fp)
	if err != nil {
		return nil, errors.New("unavailable")
	}
	if linked != nil {
		return &ssh.Permissions{Extensions: map[string]string{
			"fp":   fp,
			"user": linked.UserRefID,
			"via":  "key",
		}}, nil
	}
	if g.deps.Pending != nil {
		if userRef, id, ok := g.deps.Pending.Get(fp); ok {
			return &ssh.Permissions{Extensions: map[string]string{
				"fp":           fp,
				"user":         userRef,
				"link_pending": id,
				"via":          "link_pending",
			}}, nil
		}
	}
	return &ssh.Permissions{Extensions: map[string]string{
		"fp":  fp,
		"via": "proof",
	}}, nil
}

// verifiedKeyAuth runs only after the client proves it holds the private key.
func (g *Gate) verifiedKeyAuth(meta ssh.ConnMetadata, key ssh.PublicKey, perms *ssh.Permissions, _ string) (*ssh.Permissions, error) {
	if g == nil || perms == nil || perms.Extensions == nil {
		return nil, errors.New("refused")
	}
	if perms.Extensions["via"] == "proof" {
		if key != nil {
			fp := ssh.FingerprintSHA256(key)
			line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
			g.noteSigned(meta, fp, line)
		}
		return nil, errors.New("unlinked key")
	}
	if g.deps.Users == nil {
		return nil, errors.New("refused")
	}
	userRef := perms.Extensions["user"]
	user, err := g.deps.Users.FindByRefID(userRef)
	if err != nil || user == nil || user.IsBanned || user.IsGuest {
		return nil, errors.New("refused")
	}
	if key != nil {
		signed := ssh.FingerprintSHA256(key)
		if want := perms.Extensions["fp"]; want != "" && want != signed {
			return nil, errors.New("refused")
		}
		g.setLine(meta, strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key))))
	}
	return perms, nil
}

// deviceAuth admits the connection to the lobby only. It asks no questions.
// The username guest uses none-auth when guests are enabled, so this method
// refuses that name and does not open a second path into a guest session.
func (g *Gate) deviceAuth(meta ssh.ConnMetadata, _ ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
	if g == nil || !g.cfg.Device.Enabled || g.deps.Devices == nil {
		return nil, errors.New("device login disabled")
	}
	if g.cfg.Guest.Enabled && meta != nil && meta.User() == "guest" {
		return nil, errors.New("authentication required")
	}
	return &ssh.Permissions{Extensions: map[string]string{"via": "device"}}, nil
}

func (s *liveSession) admitAccount() error {
	if s == nil || s.userRef == "" || s.gate.deps.Users == nil || s.gate.deps.Mud == nil {
		return errors.New("refused")
	}
	user, err := s.gate.deps.Users.FindByRefID(s.userRef)
	if err != nil || user == nil || user.IsBanned || user.IsGuest {
		return errors.New("refused")
	}
	s.userID = user.ID
	input, done := s.gate.deps.Mud.AttachExternal(user, s.link)
	s.input = input
	s.done = done
	if s.via == "key" && s.fp != "" && s.gate.deps.Keys != nil {
		_ = s.gate.deps.Keys.Touch(s.fp, s.ip, time.Now())
	}
	logAdmit(s.ip, user.ID, s.via)
	return nil
}

func logAdmit(ip, userID, method string) {
	if method == "" {
		method = "none"
	}
	log.WithFields(log.Fields{"ip": ip, "userId": userID, "method": method}).Info("ssh admit")
}

func accountLabel(user *entities.User) string {
	name := ""
	if user != nil {
		name = user.Nickname
		if strings.TrimSpace(name) == "" {
			name = user.Username
		}
	}
	name = textline.SingleLine(name)
	if name == "" {
		return "this account"
	}
	runes := []rune(name)
	if len(runes) > 32 {
		name = string(runes[:32])
	}
	return name
}

// noteSigned records a fingerprint only after its signature verified.
func (g *Gate) noteSigned(meta ssh.ConnMetadata, fp, line string) {
	if g == nil || fp == "" {
		return
	}
	key := sessionKey(meta)
	g.notesMu.Lock()
	defer g.notesMu.Unlock()
	g.purgeNotesLocked()
	if g.notes == nil {
		g.notes = map[string]*authNote{}
	}
	if len(g.notes) > 1024 && g.notes[key] == nil {
		return
	}
	n := g.notes[key]
	if n == nil {
		n = &authNote{at: time.Now()}
		g.notes[key] = n
	}
	if line != "" {
		n.line = line
	}
	for _, have := range n.fps {
		if have == fp {
			return
		}
	}
	if len(n.fps) >= 5 {
		return
	}
	n.fps = append(n.fps, fp)
}

// signedFingerprint is the key that produced a signature. An unsigned offer
// is not used.
func signedFingerprint(line string, offers []string) string {
	if strings.TrimSpace(line) != "" {
		pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(strings.TrimSpace(line)))
		if err == nil && pub != nil {
			return ssh.FingerprintSHA256(pub)
		}
	}
	if len(offers) > 0 {
		return offers[0]
	}
	return ""
}

func (g *Gate) setLine(meta ssh.ConnMetadata, line string) {
	if g == nil || line == "" {
		return
	}
	key := sessionKey(meta)
	g.notesMu.Lock()
	defer g.notesMu.Unlock()
	if g.notes == nil {
		g.notes = map[string]*authNote{}
	}
	n := g.notes[key]
	if n == nil {
		n = &authNote{at: time.Now()}
		g.notes[key] = n
	}
	n.line = line
}

func (g *Gate) takeNote(meta ssh.ConnMetadata) authNote {
	if g == nil {
		return authNote{}
	}
	key := sessionKey(meta)
	g.notesMu.Lock()
	defer g.notesMu.Unlock()
	n := g.notes[key]
	delete(g.notes, key)
	if n == nil {
		return authNote{}
	}
	return *n
}

func (g *Gate) purgeNotesLocked() {
	if g == nil || len(g.notes) == 0 {
		return
	}
	cut := time.Now().Add(-2 * time.Minute)
	for key, note := range g.notes {
		if note == nil || note.at.Before(cut) {
			delete(g.notes, key)
		}
	}
}

func sessionKey(meta ssh.ConnMetadata) string {
	if meta == nil {
		return ""
	}
	if id := meta.SessionID(); len(id) > 0 {
		return hex.EncodeToString(id)
	}
	return remoteHost(meta.RemoteAddr()) + "\n" + meta.User()
}
