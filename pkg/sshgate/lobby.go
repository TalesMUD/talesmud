package sshgate

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/presentation/termout"
	"github.com/talesmud/talesmud/pkg/presentation/textline"
)

const (
	linkReconnect = "Reconnect once to finish linking this key.\r\n"
	linkQuestion  = "Only press Y if that is your account."
)

func (s *liveSession) openClassic(in <-chan []byte) bool {
	switch s.via {
	case "guest":
		if err := s.admitGuest(); err != nil {
			log.WithFields(log.Fields{"ip": s.ip, "method": "guest"}).Info("ssh guest refused")
			_, _ = s.channel.Write([]byte(guestUnavailable))
			return false
		}
		return true
	case "key":
		if err := s.admitAccount(); err != nil {
			s.refuse(nil, "Sign-in was refused.\r\n")
			return false
		}
		return true
	case "link_pending":
		if !s.confirmPending(in, nil, nil) {
			return false
		}
		if err := s.admitAccount(); err != nil {
			s.refuse(nil, "Sign-in was refused.\r\n")
			return false
		}
		return true
	case "device":
		return s.deviceLobby(in, nil, nil)
	default:
		s.refuse(nil, "Authentication required.\r\n")
		return false
	}
}

func (s *liveSession) openDoorAuth(in <-chan []byte, view *termout.Screen, cs *termout.Charset) bool {
	switch s.via {
	case "key":
		if err := s.admitAccount(); err != nil {
			s.refuse(view, "Sign-in was refused.\r\n")
			return false
		}
		return true
	case "link_pending":
		if !s.confirmPending(in, view, cs) {
			return false
		}
		if err := s.admitAccount(); err != nil {
			s.refuse(view, "Sign-in was refused.\r\n")
			return false
		}
		return true
	case "device":
		return s.deviceLobby(in, view, cs)
	default:
		s.refuse(view, "Authentication required.\r\n")
		return false
	}
}

// deviceLobby shows the code and waits for the account owner. It does not
// attach a game session until the code is confirmed, or the player chooses guest.
func (s *liveSession) deviceLobby(in <-chan []byte, view *termout.Screen, cs *termout.Charset) bool {
	if s.gate.deps.Devices == nil {
		s.refuse(view, "Sign-in is unavailable.\r\n")
		return false
	}
	ttl := s.gate.cfg.Device.TTL.Duration()
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	id, code, expires, err := s.gate.deps.Devices.Begin(s.ip, s.mode(), s.clientVersion, s.offers)
	if err != nil {
		s.refuse(view, "Too many sign-in attempts from this address. Try again later.\r\n")
		return false
	}
	s.deviceID = id
	defer func() {
		if s.deviceID != "" && s.gate.deps.Devices != nil {
			s.gate.deps.Devices.Cancel(s.deviceID)
			s.deviceID = ""
		}
	}()
	url := s.activateURL()
	var tmpl string
	if view != nil {
		tmpl = loadActivateTemplate(s.gate.cfg)
	} else if !s.writeRaw([]byte(classicLobbyText(code, url, expires))) {
		return false
	}
	deadline := time.NewTimer(ttl)
	defer deadline.Stop()
	poll := time.NewTicker(750 * time.Millisecond)
	defer poll.Stop()
	painted := false
	var lastShown string
	paint := func() bool {
		if view == nil {
			return true
		}
		left := time.Until(expires).Round(time.Second)
		if left < 0 {
			left = 0
		}
		shown := left.String()
		text := termout.Substitute(tmpl, code, url, shown)
		if !strings.Contains(text, code) {
			text = termout.Substitute(genericActivateTemplate(), code, url, shown)
		}
		cols, rows := s.window()
		charset := termout.UTF8
		if cs != nil {
			charset = *cs
		}
		force := !painted || shown != lastShown
		painted = true
		lastShown = shown
		return s.writeRaw(view.Paint(text, cols, rows, charset, force))
	}
	if !paint() {
		return false
	}
	for {
		select {
		case <-s.stop:
			return false
		case <-s.kick:
			return false
		case <-deadline.C:
			s.refuse(view, "code not found or expired\r\n")
			return false
		case <-poll.C:
			switch s.gate.deps.Devices.State(id) {
			case "confirmed":
				return s.finishDevice(in, view, cs, id)
			case "pending":
				if !paint() {
					return false
				}
			case "denied":
				s.deviceID = ""
				s.refuse(view, "Sign-in was denied.\r\n")
				return false
			default:
				s.deviceID = ""
				s.refuse(view, "code not found or expired\r\n")
				return false
			}
		case chunk, ok := <-in:
			if !ok {
				return false
			}
			if isQuit(chunk) {
				s.refuse(view, "Goodbye.\r\n")
				return false
			}
			if isLetter(chunk, 'g') && s.gate.cfg.Guest.Enabled {
				s.gate.deps.Devices.Cancel(id)
				s.deviceID = ""
				if err := s.admitGuest(); err != nil {
					s.refuse(view, guestUnavailable)
					return false
				}
				return true
			}
			if view != nil && cs != nil && isLetter(chunk, 'c') {
				*cs = toggleCharset(*cs)
				painted = false
				if !paint() {
					return false
				}
			}
		}
	}
}

func (s *liveSession) finishDevice(in <-chan []byte, view *termout.Screen, cs *termout.Charset, id string) bool {
	userRef, ok := s.gate.deps.Devices.Take(id)
	s.deviceID = ""
	if !ok || userRef == "" {
		s.refuse(view, "code not found or expired\r\n")
		return false
	}
	s.userRef = userRef
	s.via = "device"
	if !s.offerLink(in, view, cs) {
		return false
	}
	if err := s.admitAccount(); err != nil {
		s.refuse(view, "Sign-in was refused.\r\n")
		return false
	}
	return true
}

// offerLink asks the account owner to remember the key that signed.
// Y stores a pending link only. The key row is written on the next connection,
// after the signature is verified and the owner answers Y again.
// N leaves the device-confirmed account in the game and stores nothing.
func (s *liveSession) offerLink(in <-chan []byte, view *termout.Screen, cs *termout.Charset) bool {
	fp := signedFingerprint(s.keyLine, s.offers)
	if fp == "" || s.gate.deps.Pending == nil || !s.gate.cfg.Keys.Enabled {
		return true
	}
	user, err := s.gate.deps.Users.FindByRefID(s.userRef)
	if err != nil || user == nil || user.IsBanned || user.IsGuest {
		s.refuse(view, "Sign-in was refused.\r\n")
		return false
	}
	prompt := fmt.Sprintf("Remember this computer? Link key %s to %s [Y/N]\r\n", MaskFingerprint(fp), accountLabel(user))
	answer, ok := s.askYN(in, prompt, view, cs)
	if !ok || answer == "q" {
		s.refuse(view, "Goodbye.\r\n")
		return false
	}
	if answer == "y" {
		if _, offered := s.gate.deps.Pending.Offer(fp, s.userRef); offered {
			if !s.writeRaw([]byte(linkReconnect)) {
				return false
			}
		}
		return true
	}
	return s.writeRaw([]byte("Key not linked.\r\n"))
}

// confirmPending asks the key holder to link the signed key to the named account.
// Y admits that account after the key row is written. N, a timeout, or quitting
// drops the pending row and disconnects without admitting it.
func (s *liveSession) confirmPending(in <-chan []byte, view *termout.Screen, cs *termout.Charset) bool {
	if s.gate.deps.Users == nil {
		s.refuse(view, "Sign-in was refused.\r\n")
		return false
	}
	user, err := s.gate.deps.Users.FindByRefID(s.userRef)
	if err != nil || user == nil || user.IsBanned || user.IsGuest {
		s.refuse(view, "Sign-in was refused.\r\n")
		return false
	}
	drop := func() {
		if s.gate.deps.Pending != nil && s.fp != "" {
			s.gate.deps.Pending.Drop(s.fp)
		}
	}
	prompt := fmt.Sprintf("Link key %s to account %s? %s [Y/N]\r\n", MaskFingerprint(s.fp), accountLabel(user), linkQuestion)
	answer, ok := s.askYN(in, prompt, view, cs)
	if !ok || answer != "y" {
		drop()
		if answer == "n" {
			s.refuse(view, "Key not linked.\r\n")
		} else {
			s.refuse(view, "Goodbye.\r\n")
		}
		return false
	}
	signed := signedFingerprint(s.keyLine, nil)
	if s.gate.deps.Keys == nil || signed == "" || signed != s.fp {
		drop()
		s.refuse(view, "The key was not linked.\r\n")
		return false
	}
	max := s.gate.cfg.Keys.MaxPerAccount
	if _, err := s.gate.deps.Keys.Add(s.userRef, s.keyLine, "ssh", "device", max); err != nil {
		log.WithFields(log.Fields{"ip": s.ip, "userId": user.ID, "method": "link_pending"}).Info("ssh link refused")
		drop()
		s.refuse(view, "The key was not linked.\r\n")
		return false
	}
	drop()
	if !s.writeRaw([]byte("Key linked.\r\n")) {
		return false
	}
	return true
}

func (s *liveSession) askYN(in <-chan []byte, prompt string, view *termout.Screen, cs *termout.Charset) (string, bool) {
	if view != nil {
		if !s.paintPrompt(view, cs, strings.TrimRight(prompt, "\r\n")) {
			return "", false
		}
	}
	if !s.writeRaw([]byte(prompt)) {
		return "", false
	}
	deadline := time.NewTimer(2 * time.Minute)
	defer deadline.Stop()
	for {
		select {
		case <-s.stop:
			return "", false
		case <-s.kick:
			return "", false
		case <-deadline.C:
			return "n", true
		case chunk, ok := <-in:
			if !ok {
				return "", false
			}
			if answer := exactYN(chunk); answer != "" {
				return answer, true
			}
		}
	}
}

func (s *liveSession) paintPrompt(view *termout.Screen, cs *termout.Charset, prompt string) bool {
	if view == nil {
		return true
	}
	lines := make([]string, 25)
	lines[10] = prompt
	lines[12] = "[Y] yes    [N] no    [Q] quit"
	charset := termout.UTF8
	if cs != nil {
		charset = *cs
	}
	cols, rows := s.window()
	return s.writeRaw(view.Paint(strings.Join(lines, "\r\n"), cols, rows, charset, true))
}

func (s *liveSession) refuse(view *termout.Screen, msg string) {
	if view != nil {
		s.endDoor(view, msg)
		return
	}
	if s.channel != nil && msg != "" {
		_, _ = s.channel.Write([]byte(msg))
	}
}

func (s *liveSession) mode() string {
	if s != nil && s.gate != nil && s.gate.door {
		return "door"
	}
	return "mud"
}

func (s *liveSession) activateURL() string {
	if s == nil || s.gate == nil {
		return "the activate page"
	}
	u := strings.TrimSpace(s.gate.cfg.Device.ActivateURL)
	if u == "" {
		return "the activate page"
	}
	return u
}

func classicLobbyText(code, url string, expires time.Time) string {
	left := time.Until(expires).Round(time.Second)
	if left < 0 {
		left = 0
	}
	return "\r\nTalesMUD SSH\r\nOpen " + url + "\r\nCode: " + code + "\r\nExpires in " + left.String() + ".\r\n[G] continue as guest    [Q] quit\r\n"
}

func genericActivateTemplate() string {
	lines := make([]string, 25)
	lines[6] = "TalesMUD"
	lines[8] = "Sign in with this code"
	lines[10] = "{{CODE}}"
	lines[12] = "{{URL}}"
	lines[14] = "expires {{EXPIRES}}"
	lines[18] = "[G] guest     [Q] quit"
	return strings.Join(lines, "\r\n")
}

func loadActivateTemplate(cfg gamemode.SSHConfig) string {
	path := resolvePackFile(gamemode.Current().WorldPack, cfg.Door.ActivateScreen)
	if path == "" {
		return genericActivateTemplate()
	}
	// Operator-chosen screen. resolvePackFile already rejected a relative path
	// that escapes the world pack. The bytes are painted, never logged.
	b, err := os.ReadFile(path) // #nosec G304 -- operator-configured activate screen path
	if err != nil {
		log.WithField("path", textline.Sanitize(path)).Info("ssh activate screen missing, using built-in")
		return genericActivateTemplate()
	}
	if len(b) > splashCap {
		b = b[:splashCap]
	}
	b = stripSAUCE(b)
	text := termout.DecodeANS(b)
	if strings.TrimSpace(stripANSI(text)) == "" {
		return genericActivateTemplate()
	}
	return text
}

// exactYN accepts only a whole chunk of y, n, or quit, after a trailing CR/LF.
func exactYN(chunk []byte) string {
	text := strings.TrimRight(string(chunk), "\r\n")
	switch text {
	case "y", "Y":
		return "y"
	case "n", "N":
		return "n"
	case "q", "Q", "\x03", "\x04":
		return "q"
	default:
		return ""
	}
}

func isLetter(chunk []byte, want byte) bool {
	upper := want
	if want >= 'a' && want <= 'z' {
		upper = want - ('a' - 'A')
	}
	lower := upper + ('a' - 'A')
	for _, b := range chunk {
		if b == upper || b == lower {
			return true
		}
	}
	return false
}

func isQuit(chunk []byte) bool {
	for _, b := range chunk {
		switch b {
		case 'q', 'Q', 3, 4:
			return true
		}
	}
	return false
}
