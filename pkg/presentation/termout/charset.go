// Package termout adapts an 80x25 text frame to the SSH client's terminal.
// It is generic: no world names and no pack art.
package termout

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Charset is how frame text is written to the client.
type Charset string

const (
	UTF8  Charset = "utf8"
	CP437 Charset = "cp437"
	ASCII Charset = "ascii"
)

// cp437Runes is the CP437 mapping for bytes 0x80 through 0xFF, in order.
const cp437Runes = "ÇüéâäàåçêëèïîìÄÅ" +
	"ÉæÆôöòûùÿÖÜ¢£¥₧ƒ" +
	"áíóúñÑªº¿⌐¬½¼¡«»" +
	"░▒▓│┤╡╢╖╕╣║╗╝╜╛┐" +
	"└┴┬├─┼╞╟╚╔╩╦╠═╬╧" +
	"╨╤╥╙╘╒╓╫╪┘┌█▄▌▐▀" +
	"αßΓπΣσµτΦΘΩδ∞φε∩" +
	"≡±≥≤⌠⌡÷≈°∙·√ⁿ²■\u00a0"

var cp437Rev map[rune]byte

func init() {
	cp437Rev = make(map[rune]byte, utf8.RuneCountInString(cp437Runes))
	i := 0
	for _, r := range cp437Runes {
		cp437Rev[r] = byte(0x80 + i)
		i++
	}
}

// Detect picks a charset. An explicit TALES_CHARSET wins. Otherwise TERM
// ansi / syncterm / pcansi selects CP437, dumb / vt52 selects ASCII, a UTF-8
// locale or a modern TERM selects UTF-8, and anything else uses fallback.
func Detect(term, lang, lcAll, lcCtype, tales, fallback string) Charset {
	if cs, ok := parseCharset(tales); ok {
		return cs
	}
	term = strings.ToLower(strings.TrimSpace(term))
	switch {
	case term == "ansi" || strings.HasPrefix(term, "ansi-") ||
		term == "syncterm" || strings.HasPrefix(term, "syncterm") ||
		term == "pcansi" || strings.HasPrefix(term, "pcansi"):
		return CP437
	case term == "dumb" || term == "vt52" || strings.HasPrefix(term, "vt52"):
		return ASCII
	}
	if containsUTF8(lcAll) || containsUTF8(lcCtype) || containsUTF8(lang) {
		return UTF8
	}
	if strings.HasPrefix(term, "xterm") || strings.HasPrefix(term, "screen") ||
		strings.HasPrefix(term, "tmux") || strings.HasPrefix(term, "rxvt") ||
		strings.HasPrefix(term, "alacritty") || strings.HasPrefix(term, "kitty") ||
		term == "linux" {
		return UTF8
	}
	if cs, ok := parseCharset(fallback); ok {
		return cs
	}
	return UTF8
}

func parseCharset(s string) (Charset, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "utf8", "utf-8":
		return UTF8, true
	case "cp437", "ibm437":
		return CP437, true
	case "ascii":
		return ASCII, true
	default:
		return "", false
	}
}

func containsUTF8(s string) bool {
	s = strings.ToUpper(s)
	return strings.Contains(s, "UTF-8") || strings.Contains(s, "UTF8")
}

// Encode writes s for the client. CSI is kept so pack .ans art can clear,
// home, and color. OSC, DCS, other escapes, C1, and format controls are
// dropped. Unmappable runes become '?'.
func Encode(s string, cs Charset) []byte {
	if s == "" {
		return nil
	}
	if cs == "" {
		cs = UTF8
	}
	raw := []byte(s)
	out := make([]byte, 0, len(raw))
	for i := 0; i < len(raw); {
		if raw[i] == 0x1b {
			j, keep := takeOutputESC(raw, i)
			if keep {
				out = append(out, raw[i:j]...)
			}
			i = j
			continue
		}
		r, size := utf8.DecodeRune(raw[i:])
		if size < 1 || (r == utf8.RuneError && size == 1) {
			i++
			continue
		}
		if (r >= 0x80 && r <= 0x9f) || unicode.Is(unicode.Cf, r) {
			i += size
			continue
		}
		if cs == UTF8 {
			out = append(out, raw[i:i+size]...)
		} else {
			out = append(out, mapRune(r, cs))
		}
		i += size
	}
	return out
}

// takeOutputESC keeps CSI, including the cursor and erase commands in pack
// art. OSC and DCS are dropped, including the DCS payload.
func takeOutputESC(b []byte, i int) (int, bool) {
	if i+1 >= len(b) {
		return len(b), false
	}
	switch b[i+1] {
	case '[':
		return skipESC(b, i), true
	case ']':
		return skipESC(b, i), false
	case 'P', 'X', '^', '_':
		return skipUntilST(b, i+2), false
	default:
		return skipESC(b, i), false
	}
}

func skipUntilST(b []byte, j int) int {
	for j < len(b) {
		if b[j] == 0x07 {
			return j + 1
		}
		if b[j] == 0x1b && j+1 < len(b) && b[j+1] == '\\' {
			return j + 2
		}
		j++
	}
	return j
}

func mapRune(r rune, cs Charset) byte {
	if r == '\t' {
		return '\t'
	}
	if r >= 0x20 && r < 0x7f {
		return byte(r)
	}
	if cs == ASCII {
		switch r {
		case '░':
			return '.'
		case '▒':
			return ':'
		case '▓':
			return '#'
		case '█':
			return '@'
		default:
			return '?'
		}
	}
	if b, ok := cp437Rev[r]; ok {
		return b
	}
	return '?'
}

// DecodeANS interprets a .ans blob as CP437 plus ESC sequences and returns
// Unicode text. C0 bytes other than CR, LF, and tab are dropped.
func DecodeANS(b []byte) string {
	var sb strings.Builder
	sb.Grow(len(b))
	for i := 0; i < len(b); {
		if b[i] == 0x1b {
			j := skipESC(b, i)
			sb.Write(b[i:j])
			i = j
			continue
		}
		switch b[i] {
		case '\r', '\n', '\t':
			sb.WriteByte(b[i])
			i++
			continue
		}
		if b[i] < 0x80 {
			if b[i] >= 0x20 && b[i] < 0x7f {
				sb.WriteByte(b[i])
			}
			i++
			continue
		}
		sb.WriteRune(cp437Rune(b[i]))
		i++
	}
	return sb.String()
}

func cp437Rune(b byte) rune {
	if b < 0x80 {
		return rune(b)
	}
	n := int(b - 0x80)
	i := 0
	for _, r := range cp437Runes {
		if i == n {
			return r
		}
		i++
	}
	return '?'
}

// Substitute fills the device-code placeholders. Replacement text is stripped
// of terminal controls so a value cannot repaint the screen.
func Substitute(screen, code, url, expires string) string {
	screen = strings.ReplaceAll(screen, "{{CODE}}", cleanSub(code))
	screen = strings.ReplaceAll(screen, "{{URL}}", cleanSub(url))
	screen = strings.ReplaceAll(screen, "{{EXPIRES}}", cleanSub(expires))
	return screen
}

func cleanSub(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			i = skipESC([]byte(s), i)
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if size < 1 {
			size = 1
			i += size
			continue
		}
		i += size
		if r == '\t' || (r >= 0x20 && r != 0x7f && !(r >= 0x80 && r <= 0x9f)) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func skipESC(b []byte, i int) int {
	if i >= len(b) || b[i] != 0x1b {
		if i+1 > len(b) {
			return len(b)
		}
		return i + 1
	}
	if i+1 >= len(b) {
		return len(b)
	}
	switch b[i+1] {
	case '[':
		j := i + 2
		for j < len(b) {
			c := b[j]
			j++
			if c >= 0x40 && c <= 0x7e {
				return j
			}
		}
		return j
	case ']':
		j := i + 2
		for j < len(b) {
			if b[j] == 0x07 {
				return j + 1
			}
			if b[j] == 0x1b && j+1 < len(b) && b[j+1] == '\\' {
				return j + 2
			}
			j++
		}
		return j
	default:
		if i+2 > len(b) {
			return len(b)
		}
		return i + 2
	}
}
