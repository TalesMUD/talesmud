package textline

import "strings"

// Sanitize strips terminal controls from player-supplied text.
// Tab, CR, and LF are kept. ESC, other C0, DEL, and C1 are removed so a
// name or a chat line cannot set a title, write the clipboard, or recolor
// the rest of the screen. Pack ANSI art is not passed through here.
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
		case r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f):
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func needsSanitize(s string) bool {
	for _, r := range s {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
		if r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			return true
		}
	}
	return false
}
