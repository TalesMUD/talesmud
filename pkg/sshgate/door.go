package sshgate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/presentation/ansi"
	"github.com/talesmud/talesmud/pkg/presentation/doorkeys"
	"github.com/talesmud/talesmud/pkg/presentation/termout"
	"github.com/talesmud/talesmud/pkg/presentation/textline"
)

const (
	splashCap   = 256 * 1024
	doorGoodbye = "Goodbye.\r\n"
)

func (s *liveSession) doorLoop() {
	defer s.finish()
	defer s.flushClose()

	s.mu.Lock()
	env := copyEnv(s.env)
	s.mu.Unlock()
	cs := termout.Detect(env["TERM"], env["LANG"], env["LC_ALL"], env["LC_CTYPE"], env["TALES_CHARSET"], s.gate.cfg.Door.CharsetDefault)
	view := termout.NewScreen()
	view.SetFill(s.gate.cfg.Door.LetterboxFill)
	defer func() {
		if s.channel == nil {
			return
		}
		if p := view.Exit(); len(p) > 0 {
			_, _ = s.channel.Write(p)
		}
	}()

	splash := loadDoorSplash(s.gate.cfg)
	keys := doorkeys.New()
	var last ansi.Frame
	playing := false

	paint := func(force bool) bool {
		cols, rows := s.window()
		blob := splash
		if playing && last.ANSI != "" {
			blob = last.ANSI
		}
		if !s.writeRaw(view.Paint(blob, cols, rows, cs, force)) {
			return false
		}
		if playing && keys.Echoing() {
			if !s.writeRaw(termout.EchoLine(keys.Shown(), cols, rows, cs)) {
				return false
			}
			view.Invalidate(22)
		}
		return true
	}
	if !paint(true) {
		return
	}

	in := make(chan []byte, 16)
	go s.readInput(in)

	idleFor := s.gate.cfg.IdleTimeout.Duration()
	idle := time.NewTimer(idleFor)
	defer idle.Stop()
	var idleWarn <-chan time.Time
	var idleWarnTimer *time.Timer
	if idleFor > time.Minute {
		idleWarnTimer = time.NewTimer(idleFor - time.Minute)
		defer idleWarnTimer.Stop()
		idleWarn = idleWarnTimer.C
	}
	var maxEnd <-chan time.Time
	var maxWarn <-chan time.Time
	var maxTimer, maxWarnTimer *time.Timer
	defer func() {
		stopTimer(maxTimer)
		stopTimer(maxWarnTimer)
	}()
	startGuestMax := func() {
		maxFor := s.gate.cfg.Guest.MaxSession.Duration()
		if maxFor <= 0 {
			return
		}
		maxTimer = time.NewTimer(maxFor)
		maxEnd = maxTimer.C
		if maxFor > 5*time.Minute {
			maxWarnTimer = time.NewTimer(maxFor - 5*time.Minute)
			maxWarn = maxWarnTimer.C
		}
	}

	for {
		select {
		case <-s.stop:
			return
		case <-s.kick:
			return
		case <-s.repaint:
			if !paint(true) {
				return
			}
		case item := <-s.link.queue:
			if s.netConn != nil {
				_ = s.netConn.SetWriteDeadline(time.Time{})
			}
			if s.link.closed.Load() || !playing {
				if s.link.closed.Load() {
					return
				}
				continue
			}
			frame, ok := asFrame(item.msg)
			if !ok {
				continue
			}
			if frame.Logout {
				s.endDoor(view, doorGoodbye)
				return
			}
			last = frame
			keys.OnFrame(frame.Accepts, frame.InputMode, frame.Prompt)
			if !paint(false) {
				return
			}
		case chunk, ok := <-in:
			if !ok {
				return
			}
			resetTimer(idle, idleFor)
			cols, rows := s.window()
			if !playing {
				admit, quit, toggle, refresh := splashKeys(chunk)
				if quit {
					s.endDoor(view, "\r\n")
					return
				}
				if toggle {
					cs = toggleCharset(cs)
					refresh = true
				}
				if refresh && !paint(true) {
					return
				}
				if admit && termout.Fits(cols, rows) {
					if err := s.admitGuest(); err != nil {
						log.WithFields(log.Fields{"ip": s.ip, "method": "guest"}).Info("ssh guest refused")
						s.endDoor(view, guestUnavailable)
						return
					}
					playing = true
					startGuestMax()
				}
				continue
			}
			if !termout.Fits(cols, rows) {
				if quitChunk(chunk) {
					s.endDoor(view, "\r\n")
					return
				}
				if refreshChunk(chunk) && !paint(true) {
					return
				}
				continue
			}
			res := keys.Feed(string(chunk))
			if res.Send && s.input != nil {
				s.input(res.Key, true)
			}
			if refreshChunk(chunk) {
				if !paint(true) {
					return
				}
				continue
			}
			if res.Echo {
				if !s.writeRaw(termout.EchoLine(keys.Shown(), cols, rows, cs)) {
					return
				}
				view.Invalidate(22)
			}
		case <-idleWarn:
			s.writeRaw([]byte("\r\nIdle for too long. Disconnecting in 1 minute.\r\n"))
			idleWarn = nil
		case <-idle.C:
			s.endDoor(view, "\r\nIdle timeout. Goodbye.\r\n")
			return
		case <-maxWarn:
			s.writeRaw([]byte("\r\nYour guest session expires in 5 minutes.\r\n"))
			maxWarn = nil
		case <-maxEnd:
			s.endDoor(view, "\r\nYour guest session has expired. Goodbye.\r\n")
			return
		}
	}
}

func (s *liveSession) endDoor(view *termout.Screen, msg string) {
	if s.channel == nil {
		return
	}
	if view != nil {
		if p := view.Exit(); len(p) > 0 {
			_, _ = s.channel.Write(p)
		}
	}
	if msg != "" {
		_, _ = s.channel.Write([]byte(msg))
	}
}

func (s *liveSession) window() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cols, s.rows
}

func splashKeys(chunk []byte) (admit, quit, toggle, refresh bool) {
	for _, c := range chunk {
		switch c {
		case 'q', 'Q', 0x03, 0x04:
			return false, true, false, false
		case 'c', 'C':
			toggle = true
		case '\r', '\n':
			admit = true
		case 0x0c, 0x12:
			refresh = true
		}
	}
	return admit, false, toggle, refresh
}

func quitChunk(chunk []byte) bool {
	for _, c := range chunk {
		if c == 'q' || c == 'Q' || c == 0x03 || c == 0x04 {
			return true
		}
	}
	return false
}

func refreshChunk(chunk []byte) bool {
	return bytes.Contains(chunk, []byte{0x0c}) || bytes.Contains(chunk, []byte{0x12})
}

func toggleCharset(cs termout.Charset) termout.Charset {
	if cs == termout.CP437 {
		return termout.UTF8
	}
	return termout.CP437
}

func asFrame(v any) (ansi.Frame, bool) {
	switch f := v.(type) {
	case ansi.Frame:
		return f, f.Type != "" || f.ANSI != "" || f.Logout
	case *ansi.Frame:
		if f == nil {
			return ansi.Frame{}, false
		}
		return *f, f.Type != "" || f.ANSI != "" || f.Logout
	default:
		return ansi.Frame{}, false
	}
}

func copyEnv(env map[string]string) map[string]string {
	out := make(map[string]string, len(env))
	for k, v := range env {
		out[k] = v
	}
	return out
}

func stopTimer(t *time.Timer) {
	if t == nil {
		return
	}
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func loadDoorSplash(cfg gamemode.SSHConfig) string {
	path := resolvePackFile(gamemode.Current().WorldPack, cfg.Door.Splash)
	if path == "" {
		return termout.GenericSplash()
	}
	// Operator-chosen splash. resolvePackFile already rejected a relative path
	// that escapes the world pack. The bytes are painted, never logged.
	b, err := os.ReadFile(path) // #nosec G304 -- operator-configured splash path
	if err != nil {
		log.WithField("path", textline.Sanitize(path)).Info("ssh splash missing, using built-in")
		return termout.GenericSplash()
	}
	if len(b) > splashCap {
		b = b[:splashCap]
	}
	b = stripSAUCE(b)
	text := termout.DecodeANS(b)
	if strings.TrimSpace(stripANSI(text)) == "" {
		return termout.GenericSplash()
	}
	return text
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			j := i + 1
			if j < len(s) && s[j] == '[' {
				j++
				for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
					j++
				}
				if j < len(s) {
					j++
				}
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func stripSAUCE(b []byte) []byte {
	if len(b) >= 128 && string(b[len(b)-128:len(b)-128+5]) == "SAUCE" {
		return b[:len(b)-128]
	}
	return b
}

// resolvePackFile joins a pack-relative splash path onto the world pack.
// An absolute path is the operator's explicit choice. A relative path that
// escapes the pack, including through a symlink, is rejected.
func resolvePackFile(base, rel string) string {
	rel = strings.TrimSpace(rel)
	if rel == "" || strings.ContainsRune(rel, 0) {
		return ""
	}
	if filepath.IsAbs(rel) {
		return filepath.Clean(rel)
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return ""
	}
	base = strings.TrimSpace(base)
	if base == "" {
		return rel
	}
	base = filepath.Clean(base)
	full := filepath.Clean(filepath.Join(base, rel))
	sep := string(os.PathSeparator)
	if full != base && !strings.HasPrefix(full, base+sep) {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(full)
	if err != nil {
		return full
	}
	resolved = filepath.Clean(resolved)
	baseResolved := base
	if b, err := filepath.EvalSymlinks(base); err == nil {
		baseResolved = filepath.Clean(b)
	}
	if resolved != baseResolved && !strings.HasPrefix(resolved, baseResolved+sep) {
		return ""
	}
	return resolved
}
