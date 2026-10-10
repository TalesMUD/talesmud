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
// of terminal controls so a value cannot repaint the screen. A framed line
// keeps its visible width: a short value is padded, and a long value wraps
// onto following blank interior lines.
func Substitute(screen, code, url, expires string) string {
	return SubstituteSignup(screen, code, url, expires, "")
}

// SubstituteSignup is Substitute plus the new-player URL. {{SIGNUP_URL}} is
// filled first so it is not confused with {{URL}}.
func SubstituteSignup(screen, code, url, expires, signupURL string) string {
	lines, sep := splitScreen(screen)
	fields := []struct{ token, value string }{
		{"{{SIGNUP_URL}}", cleanSub(signupURL)},
		{"{{URL}}", cleanSub(url)},
		{"{{CODE}}", cleanSub(code)},
		{"{{EXPIRES}}", cleanSub(expires)},
	}
	for _, field := range fields {
		for i := 0; i < len(lines); i++ {
			if !strings.Contains(lines[i], field.token) {
				continue
			}
			var rest string
			lines[i], rest = fitField(lines[i], field.token, field.value)
			if rest != "" {
				placeLeftover(lines, i+1, rest)
			}
		}
	}
	return strings.Join(lines, sep)
}

func splitScreen(s string) ([]string, string) {
	if strings.Contains(s, "\r\n") {
		return strings.Split(s, "\r\n"), "\r\n"
	}
	return strings.Split(s, "\n"), "\n"
}

func fitField(line, token, value string) (string, string) {
	idx := strings.Index(line, token)
	if idx < 0 {
		return line, ""
	}
	tokenVis := visibleLen(token)
	valVis := visibleLen(value)
	if !lineIsFramed(line) {
		repl := value
		if valVis < tokenVis {
			repl += strings.Repeat(" ", tokenVis-valVis)
		}
		next := line[:idx] + repl + line[idx+len(token):]
		if !fieldOnly(line, token) || visibleLen(next) <= frameCols {
			return next, ""
		}
		room := frameCols - visibleLen(line[:idx]) - visibleLen(line[idx+len(token):])
		if room < 1 {
			room = 1
		}
		head, rest := takeHead(value, room)
		if gap := room - visibleLen(head); gap > 0 {
			head += strings.Repeat(" ", gap)
		}
		return line[:idx] + head + line[idx+len(token):], rest
	}
	if valVis <= tokenVis {
		repl := value + strings.Repeat(" ", tokenVis-valVis)
		return line[:idx] + repl + line[idx+len(token):], ""
	}
	return spliceFramed(line, idx, token, value)
}

func fieldOnly(line, token string) bool {
	rest := strings.Replace(line, token, "", 1)
	for i := 0; i < len(rest); {
		if rest[i] == 0x1b {
			i = skipESC([]byte(rest), i)
			continue
		}
		r, size := utf8.DecodeRuneInString(rest[i:])
		if size < 1 {
			return false
		}
		i += size
		if r != ' ' && r != '\t' {
			return false
		}
	}
	return true
}

func lineIsFramed(line string) bool {
	if visibleLen(line) >= frameCols {
		return true
	}
	for i := 0; i < len(line); {
		if line[i] == 0x1b {
			i = skipESC([]byte(line), i)
			continue
		}
		r, size := utf8.DecodeRuneInString(line[i:])
		if size < 1 {
			return false
		}
		i += size
		if r >= 0x2500 && r <= 0x259F {
			return true
		}
	}
	return false
}

type lineSeg struct {
	b0, b1 int
	csi    bool
	r      rune
}

func parseSegs(s string) []lineSeg {
	var out []lineSeg
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			j := skipESC([]byte(s), i)
			out = append(out, lineSeg{b0: i, b1: j, csi: true})
			i = j
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if size < 1 {
			size = 1
			r = rune(s[i])
		}
		out = append(out, lineSeg{b0: i, b1: i + size, r: r})
		i += size
	}
	return out
}

func spliceFramed(line string, idx int, token, value string) (string, string) {
	segs := parseSegs(line)
	tokEnd := idx + len(token)
	ts, te := -1, -1
	for i, sg := range segs {
		if sg.b1 <= idx || sg.b0 >= tokEnd {
			continue
		}
		if ts < 0 {
			ts = i
		}
		te = i + 1
	}
	if ts < 0 {
		return line, ""
	}
	var leftSpaces, rightSpaces []int
	for i := ts; i > 0; {
		sg := segs[i-1]
		if sg.csi {
			i--
			continue
		}
		if sg.r != ' ' {
			break
		}
		leftSpaces = append(leftSpaces, i-1)
		i--
	}
	for i := te; i < len(segs); {
		sg := segs[i]
		if sg.csi {
			i++
			continue
		}
		if sg.r != ' ' {
			break
		}
		rightSpaces = append(rightSpaces, i)
		i++
	}
	tokenVis := visibleLen(token)
	avail := len(leftSpaces) + len(rightSpaces)
	head, rest := takeHead(value, tokenVis+avail)
	steal := visibleLen(head) - tokenVis
	if rest != "" {
		steal = avail
	}
	if steal < 0 {
		steal = 0
	}
	if steal > avail {
		steal = avail
	}
	leftSteal := steal / 2
	if leftSteal > len(leftSpaces) {
		leftSteal = len(leftSpaces)
	}
	rightSteal := steal - leftSteal
	if rightSteal > len(rightSpaces) {
		rightSteal = len(rightSpaces)
		leftSteal = steal - rightSteal
		if leftSteal > len(leftSpaces) {
			leftSteal = len(leftSpaces)
		}
	}
	startSeg := ts
	if leftSteal > 0 {
		startSeg = leftSpaces[leftSteal-1]
	}
	endSeg := te
	if rightSteal > 0 {
		endSeg = rightSpaces[rightSteal-1] + 1
	}
	b0 := segs[startSeg].b0
	b1 := segs[endSeg-1].b1
	var lead, trail strings.Builder
	for _, sg := range segs[startSeg:endSeg] {
		if !sg.csi {
			continue
		}
		if sg.b0 < idx {
			lead.WriteString(line[sg.b0:sg.b1])
		} else if sg.b0 >= tokEnd {
			trail.WriteString(line[sg.b0:sg.b1])
		}
	}
	slotVis := tokenVis + leftSteal + rightSteal
	if gap := slotVis - visibleLen(head); gap > 0 {
		head += strings.Repeat(" ", gap)
	}
	return line[:b0] + lead.String() + head + trail.String() + line[b1:], rest
}

func takeHead(value string, width int) (string, string) {
	if width < 1 {
		return "", value
	}
	rs := []rune(value)
	if len(rs) <= width {
		return value, ""
	}
	if width >= 8 {
		lo := width - 20
		if lo < 8 {
			lo = 8
		}
		for i := width - 1; i >= lo; i-- {
			switch rs[i] {
			case '/', '?', '&', '=':
				return string(rs[:i+1]), string(rs[i+1:])
			}
		}
	}
	return string(rs[:width]), string(rs[width:])
}

func placeLeftover(lines []string, from int, leftover string) {
	for i := from; i < len(lines) && leftover != ""; i++ {
		if !padLine(lines[i]) {
			return
		}
		next, rest := fillInterior(lines[i], leftover)
		if next == lines[i] {
			return
		}
		lines[i] = next
		leftover = rest
	}
}

func padLine(line string) bool {
	if strings.Contains(line, "{{") {
		return false
	}
	spaces := 0
	for i := 0; i < len(line); {
		if line[i] == 0x1b {
			i = skipESC([]byte(line), i)
			continue
		}
		r, size := utf8.DecodeRuneInString(line[i:])
		if size < 1 {
			return false
		}
		i += size
		switch {
		case r == ' ' || r == '\t':
			spaces++
		case r >= 0x2500 && r <= 0x259F:
		default:
			return false
		}
	}
	return spaces > 0 || line == "" || visibleLen(line) == 0
}

func fillInterior(line, leftover string) (string, string) {
	if leftover == "" {
		return line, ""
	}
	bestStart, bestLen := -1, 0
	for i := 0; i < len(line); {
		if line[i] == 0x1b {
			i = skipESC([]byte(line), i)
			continue
		}
		if line[i] == ' ' {
			j := i
			for j < len(line) && line[j] == ' ' {
				j++
			}
			if j-i > bestLen {
				bestStart, bestLen = i, j-i
			}
			i = j
			continue
		}
		_, size := utf8.DecodeRuneInString(line[i:])
		if size < 1 {
			size = 1
		}
		i += size
	}
	if bestLen < 1 {
		if line != "" {
			return line, leftover
		}
		head, rest := takeHead(leftover, frameCols)
		return head, rest
	}
	head, rest := takeHead(leftover, bestLen)
	if gap := bestLen - visibleLen(head); gap > 0 {
		head += strings.Repeat(" ", gap)
	}
	return line[:bestStart] + head + line[bestStart+bestLen:], rest
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
