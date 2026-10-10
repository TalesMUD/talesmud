package textline

import (
	"strings"
	"unicode"
)

// Sanitize strips terminal controls from player-supplied text.
// Tab, CR, and LF are kept. ESC, other C0, DEL, C1, and Unicode format
// controls (bidi overrides and isolates included) are removed so a name or
// a chat line cannot set a title, write the clipboard, or reorder the line.
// Pack ANSI art is not passed through here.
func Sanitize(s string) string {
	if s == "" || !needsSanitize(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t' || r == '\n' || r == '\r':
			b.WriteRune(r)
		case r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) || unicode.Is(unicode.Cf, r):
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SingleLine is Sanitize for a label that must stay on one line.
// CR and LF are removed before the other controls.
func SingleLine(s string) string {
	if strings.ContainsAny(s, "\r\n") {
		s = strings.Map(func(r rune) rune {
			if r == '\r' || r == '\n' {
				return -1
			}
			return r
		}, s)
	}
	return strings.TrimSpace(Sanitize(s))
}

func needsSanitize(s string) bool {
	for _, r := range s {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
		if r == 0x7f || (r >= 0x80 && r <= 0x9f) || unicode.Is(unicode.Cf, r) {
			return true
		}
	}
	return false
}
