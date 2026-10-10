package sshgate

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/talesmud/talesmud/pkg/presentation/textline"
)

const (
	lineMax          = 512
	envValueMax      = 64
	replaceNotice    = "\r\nSession moved to another client.\r\n"
	revokeNotice     = "\r\nThis session was closed.\r\n"
	useTerminal      = "use: ssh -t ...\r\n"
	guestUnavailable = "Guest login is unavailable.\r\n"
)

type outbound struct {
	msg any
}

// sshLink is the mudserver transport. Send never blocks the game loop.
// A full queue or Close marks the link closed and wakes the session goroutine,
// which is the only goroutine that writes to the SSH channel.
type sshLink struct {
	ip      string
	queue   chan outbound
	closed  atomic.Bool
	code    atomic.Int32
	fail    func(code int)
	playURL string
}

func (l *sshLink) Send(v any) error {
	if l == nil || l.closed.Load() {
		return errors.New("ssh closed")
	}
	select {
	case l.queue <- outbound{msg: v}:
		return nil
	default:
		if l.fail != nil {
			l.fail(0)
		}
		return errors.New("ssh slow reader")
	}
}

func (l *sshLink) Close(code int, reason string) {
	if l == nil || l.fail == nil {
		return
	}
	l.fail(code)
	_ = reason
}

func (l *sshLink) RemoteIP() string {
	if l == nil {
		return ""
	}
	return l.ip
}

func (l *sshLink) Kind() string { return "ssh" }

type liveSession struct {
	gate          *Gate
	netConn       net.Conn
	channel       ssh.Channel
	ip            string
	stop          chan struct{}
	kick          chan struct{}
	kickOnce      sync.Once
	once          sync.Once
	link          *sshLink
	editor        *textline.Editor
	input         func(string, bool)
	done          func()
	guestHeld     bool
	userID        string
	via           string
	userRef       string
	fp            string
	linkID        string
	offers        []string
	keyLine       string
	clientVersion string
	deviceID      string
	pty           bool
	started       atomic.Bool
	mu            sync.Mutex
	cols          int
	rows          int
	env           map[string]string
	repaint       chan struct{}
	notice        string
}

func newSession(g *Gate, conn net.Conn, channel ssh.Channel, ip string) *liveSession {
	s := &liveSession{
		gate:    g,
		netConn: conn,
		channel: channel,
		ip:      ip,
		stop:    make(chan struct{}),
		kick:    make(chan struct{}),
		repaint: make(chan struct{}, 1),
		cols:    80,
		rows:    24,
		env:     map[string]string{},
	}
	s.link = &sshLink{
		ip:      ip,
		queue:   make(chan outbound, queueSize),
		playURL: g.playURL,
		fail:    s.fail,
	}
	g.track(s)
	return s
}

func (s *liveSession) fail(code int) {
	if code > 0 && code <= math.MaxInt32 {
		s.link.code.Store(int32(code))
	}
	s.link.closed.Store(true)
	s.kickOnce.Do(func() { close(s.kick) })
	s.nudge()
}

func (s *liveSession) nudge() {
	if s.netConn != nil {
		_ = s.netConn.SetWriteDeadline(time.Now().Add(200 * time.Millisecond))
	}
}

func (s *liveSession) revoke() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.notice = revokeNotice
	s.mu.Unlock()
	s.fail(0)
}

func (s *liveSession) onKick() {
	if s == nil || s.channel == nil {
		return
	}
	s.mu.Lock()
	n := s.notice
	s.notice = ""
	s.mu.Unlock()
	if n == "" {
		return
	}
	_, _ = s.channel.Write([]byte(n))
}

func (s *liveSession) finish() {
	s.once.Do(func() {
		if s.gate != nil {
			s.gate.untrack(s)
		}
		s.link.closed.Store(true)
		close(s.stop)
		if s.channel != nil {
			_ = s.channel.Close()
		}
		if s.done != nil {
			s.done()
		}
		if s.guestHeld {
			s.gate.limits.releaseGuest(s.ip)
		}
		method := s.via
		if s.guestHeld {
			method = "guest"
		}
		if method == "" {
			method = "none"
		}
		log.WithFields(log.Fields{"ip": s.ip, "userId": s.userID, "method": method}).Info("ssh close")
	})
}

func (s *liveSession) requests(reqs <-chan *ssh.Request) {
	defer s.finish()
	for req := range reqs {
		switch req.Type {
		case "pty-req":
			if term, cols, rows, ok := parsePty(req.Payload); ok {
				s.mu.Lock()
				s.pty = true
				s.cols, s.rows = cols, rows
				s.env["TERM"] = textline.Sanitize(term)
				s.mu.Unlock()
			} else {
				s.mu.Lock()
				s.pty = true
				s.mu.Unlock()
			}
			_ = req.Reply(true, nil)
		case "shell":
			if !s.hasPty() {
				log.WithFields(log.Fields{"ip": s.ip, "type": "shell"}).Info("ssh request rejected")
				_, _ = s.channel.Write([]byte(useTerminal))
				_ = req.Reply(false, nil)
				return
			}
			if !s.started.CompareAndSwap(false, true) {
				_ = req.Reply(false, nil)
				continue
			}
			_ = req.Reply(true, nil)
			if s.gate.door {
				go s.doorLoop()
			} else {
				go s.classicLoop()
			}
		case "window-change":
			if cols, rows, ok := parseWindow(req.Payload); ok {
				s.mu.Lock()
				s.cols, s.rows = cols, rows
				s.mu.Unlock()
				s.signalRepaint()
			}
			_ = req.Reply(true, nil)
		case "env":
			_ = req.Reply(s.acceptEnv(req.Payload), nil)
		case "exec", "subsystem", "x11-req", "auth-agent-req@openssh.com":
			log.WithFields(log.Fields{"ip": s.ip, "type": req.Type}).Info("ssh request rejected")
			_ = req.Reply(false, nil)
		default:
			if req.WantReply {
				_ = req.Reply(false, nil)
			}
		}
	}
}

func (s *liveSession) signalRepaint() {
	if s == nil || s.repaint == nil {
		return
	}
	select {
	case s.repaint <- struct{}{}:
	default:
	}
}

func (s *liveSession) hasPty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pty
}

func (s *liveSession) classicLoop() {
	defer s.finish()
	defer s.flushClose()
	if s.link.closed.Load() {
		return
	}
	history := s.gate.cfg.Mud.History
	if history < 1 {
		history = 20
	}
	s.editor = textline.NewEditor(history, lineMax)
	in := make(chan []byte, 16)
	go s.readInput(in)
	if !s.openClassic(in) {
		return
	}
	_, _ = s.channel.Write(s.editor.Redraw())

	idleFor := s.gate.cfg.IdleTimeout.Duration()
	idle := time.NewTimer(idleFor)
	defer idle.Stop()
	var idleWarn <-chan time.Time
	if idleFor > time.Minute {
		warn := time.NewTimer(idleFor - time.Minute)
		defer warn.Stop()
		idleWarn = warn.C
	}
	maxFor, _ := s.maxLimit()
	var maxEnd <-chan time.Time
	var maxWarn <-chan time.Time
	if maxFor > 0 {
		end := time.NewTimer(maxFor)
		defer end.Stop()
		maxEnd = end.C
		if maxFor > 5*time.Minute {
			warn := time.NewTimer(maxFor - 5*time.Minute)
			defer warn.Stop()
			maxWarn = warn.C
		}
	}

	for {
		select {
		case <-s.stop:
			return
		case <-s.kick:
			s.onKick()
			return
		case item := <-s.link.queue:
			if s.netConn != nil {
				_ = s.netConn.SetWriteDeadline(time.Time{})
			}
			if s.link.closed.Load() {
				return
			}
			if !s.writeMsg(item.msg) {
				return
			}
		case chunk, ok := <-in:
			if !ok {
				return
			}
			resetTimer(idle, idleFor)
			if !s.onBytes(chunk) {
				return
			}
		case <-idleWarn:
			s.writeRaw([]byte("\r\nIdle for too long. Disconnecting in 1 minute.\r\n"))
			s.redraw()
			idleWarn = nil
		case <-idle.C:
			_, _ = s.channel.Write([]byte("\r\nIdle timeout. Goodbye.\r\n"))
			return
		case <-maxWarn:
			s.writeRaw([]byte(s.maxNotice(true)))
			s.redraw()
			maxWarn = nil
		case <-maxEnd:
			_, _ = s.channel.Write([]byte(s.maxNotice(false)))
			return
		}
	}
}

// maxLimit is the hard cap for this session. Zero means no cap.
// Guests use ssh.guest.max_session. Other admitted sessions use ssh.max_session.
func (s *liveSession) maxLimit() (time.Duration, bool) {
	if s == nil || s.gate == nil {
		return 0, false
	}
	if s.guestHeld || s.via == "guest" {
		return s.gate.cfg.Guest.MaxSession.Duration(), true
	}
	if s.via == "" {
		return 0, false
	}
	return s.gate.cfg.MaxSession.Duration(), false
}

func (s *liveSession) maxNotice(warn bool) string {
	guest := s != nil && (s.guestHeld || s.via == "guest")
	if warn {
		if guest {
			return "\r\nYour guest session expires in 5 minutes.\r\n"
		}
		return "\r\nYour session expires in 5 minutes.\r\n"
	}
	if guest {
		return "\r\nYour guest session has expired. Goodbye.\r\n"
	}
	return "\r\nYour session has expired. Goodbye.\r\n"
}

func (s *liveSession) admitGuest() error {
	if !s.gate.cfg.Guest.Enabled {
		return errors.New("guest disabled")
	}
	if err := s.gate.limits.acquireGuest(s.ip, time.Now()); err != nil {
		return err
	}
	s.guestHeld = true
	if s.gate.deps.Guests == nil || s.gate.deps.Users == nil || s.gate.deps.Mud == nil {
		return errors.New("guest unavailable")
	}
	token, err := s.gate.deps.Guests.CreateGuestSessionPick(s.ip, "", "")
	if err != nil {
		return err
	}
	uid, err := s.gate.deps.Guests.ValidateGuestToken(token)
	if err != nil {
		return err
	}
	user, err := s.gate.deps.Users.FindByID(uid)
	if err != nil || user == nil || !user.IsGuest || user.IsBanned || user.IsCreator() {
		return errors.New("guest refused")
	}
	s.mu.Lock()
	s.userID = user.ID
	s.mu.Unlock()
	input, done := s.gate.deps.Mud.AttachExternal(user, s.link)
	s.input = input
	s.done = done
	log.WithFields(log.Fields{"ip": s.ip, "userId": user.ID, "method": "guest"}).Info("ssh admit")
	return nil
}

func (s *liveSession) readInput(in chan<- []byte) {
	defer close(in)
	buf := make([]byte, 1024)
	for {
		n, err := s.channel.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			select {
			case in <- chunk:
			case <-s.stop:
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (s *liveSession) onBytes(chunk []byte) bool {
	if s.editor == nil {
		return false
	}
	res := s.editor.Feed(chunk)
	if len(res.Out) > 0 && !s.writeRaw(res.Out) {
		return false
	}
	if res.Submitted && s.input != nil {
		s.input(res.Line, false)
	}
	if res.Quit {
		_, _ = s.channel.Write([]byte("\r\n"))
		return false
	}
	return true
}

func (s *liveSession) writeMsg(v any) bool {
	if s.editor == nil {
		return false
	}
	rendered := textline.Render(v, textline.Options{PlayURL: s.link.playURL})
	if rendered.ChoiceKind != "" {
		s.editor.SetChoices(rendered.ChoiceKind, rendered.Choices)
	}
	if rendered.Text == "" {
		return true
	}
	return s.writeRaw(s.editor.PrintAbove(rendered.Text))
}

func (s *liveSession) redraw() {
	if s.editor == nil {
		return
	}
	s.writeRaw(s.editor.Redraw())
}

func (s *liveSession) writeRaw(b []byte) bool {
	if len(b) == 0 || s.link.closed.Load() {
		return !s.link.closed.Load()
	}
	_, err := s.channel.Write(b)
	if err != nil {
		return false
	}
	if s.netConn != nil {
		_ = s.netConn.SetWriteDeadline(time.Time{})
	}
	return true
}

func (s *liveSession) flushClose() {
	s.onKick()
	if s.link.code.Load() != closeSessionReplaced {
		return
	}
	if s.netConn != nil {
		_ = s.netConn.SetWriteDeadline(time.Now().Add(time.Second))
	}
	_, _ = s.channel.Write([]byte(replaceNotice))
}

func (s *liveSession) acceptEnv(payload []byte) bool {
	name, rest, ok := sshString(payload)
	if !ok {
		return false
	}
	value, _, ok := sshString(rest)
	if !ok {
		return false
	}
	switch name {
	case "LANG", "LC_ALL", "LC_CTYPE", "TALES_CHARSET":
		value = textline.Sanitize(value)
		if len(value) > envValueMax {
			value = value[:envValueMax]
		}
		s.mu.Lock()
		s.env[name] = value
		s.mu.Unlock()
		return true
	default:
		return false
	}
}

func resetTimer(t *time.Timer, d time.Duration) {
	if d <= 0 {
		return
	}
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(d)
}

func parsePty(payload []byte) (term string, cols, rows int, ok bool) {
	term, rest, ok := sshString(payload)
	if !ok || len(rest) < 8 {
		return "", 80, 24, false
	}
	cols = int(binary.BigEndian.Uint32(rest[0:4]))
	rows = int(binary.BigEndian.Uint32(rest[4:8]))
	cols, rows = clampWindow(cols, rows)
	return term, cols, rows, true
}

func parseWindow(payload []byte) (cols, rows int, ok bool) {
	if len(payload) < 8 {
		return 0, 0, false
	}
	cols = int(binary.BigEndian.Uint32(payload[0:4]))
	rows = int(binary.BigEndian.Uint32(payload[4:8]))
	cols, rows = clampWindow(cols, rows)
	return cols, rows, true
}

func clampWindow(cols, rows int) (int, int) {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	if cols > 512 {
		cols = 512
	}
	if rows > 512 {
		rows = 512
	}
	return cols, rows
}

func sshString(b []byte) (string, []byte, bool) {
	if len(b) < 4 {
		return "", b, false
	}
	n := int(binary.BigEndian.Uint32(b[:4]))
	if n < 0 || len(b) < 4+n {
		return "", b, false
	}
	return string(b[4 : 4+n]), b[4+n:], true
}
