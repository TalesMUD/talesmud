// Package ansi paints a fixed 80x25 text frame for the text client.
// It does not know prices, combat math, or world names.
package ansi

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	Cols = 80
	Rows = 25

	ansiReset = "\x1b[0m"
	ansiTitle = "\x1b[1;36m"
	ansiDim   = "\x1b[0;37m"
	ansiBar   = "\x1b[30;46m"
	ansiRails = "\x1b[0;36m"
	ansiHot   = "\x1b[1;37m"

	// StyleRails is the pack-selected frame: a dithered location rail,
	// a status row the caller already coloured, and a half-block rule.
	StyleRails = "rails"
)

// Frame is one full-screen page pushed over the existing WebSocket.
type Frame struct {
	Type      string            `json:"type"`
	ANSI      string            `json:"ansi"`
	InputMode string            `json:"inputMode"`
	Cols      int               `json:"cols"`
	Rows      int               `json:"rows"`
	Prompt    string            `json:"prompt"`
	Screen    string            `json:"screen,omitempty"`
	Accepts   []string          `json:"accepts,omitempty"`
	Keys      map[string]string `json:"keys,omitempty"`
	Logout    bool              `json:"logout,omitempty"`
}

// Page is the text the caller wants on screen.
type Page struct {
	Title     string
	Location  string
	Body      []string
	Prompt    string
	Footer    string
	InputMode string
	ScreenID  string
	Keys      map[string]string
	// Style selects the chrome. Empty keeps the classic bar.
	Style string
	// Status is the precoloured status row for StyleRails.
	// The caller owns the words. This package only fits the row.
	Status string
}

// Render builds one frame. An empty title leaves the left side of the bar blank.
func Render(page Page) Frame {
	mode := page.InputMode
	if mode == "" {
		mode = "hotkey"
	}
	accepts := []string{"hotkey"}
	if mode == "line" {
		accepts = []string{"line"}
	}
	lines := make([]string, Rows)
	if page.Style == StyleRails {
		lines[0] = railsBar(page.Location, framePlain(page.Title))
		lines[1] = fit(page.Status, Cols)
		lines[2] = fit(railsRule(), Cols)
	} else {
		lines[0] = titleBar(framePlain(page.Title), "")
		lines[1] = fit(ansiTitle+page.Location+ansiReset, Cols)
		lines[2] = fit(ansiDim+strings.Repeat("-", Cols)+ansiReset, Cols)
	}
	const bodyRows = 19
	for i := 0; i < bodyRows; i++ {
		src := ""
		if i < len(page.Body) {
			src = page.Body[i]
		}
		lines[3+i] = fit(src, Cols)
	}
	lines[22] = fit(page.Prompt, Cols)
	lines[23] = fit(ansiDim+page.Footer+ansiReset, Cols)
	lines[24] = fit("", Cols)

	var b strings.Builder
	b.WriteString("\x1b[2J\x1b[H")
	for i, ln := range lines {
		if i > 0 {
			b.WriteString("\r\n")
		}
		b.WriteString(ln)
	}
	return Frame{
		Type:      "door_frame",
		ANSI:      b.String(),
		InputMode: mode,
		Cols:      Cols,
		Rows:      Rows,
		Prompt:    page.Prompt,
		Screen:    page.ScreenID,
		Accepts:   accepts,
		Keys:      page.Keys,
	}
}

func railsBar(location, title string) string {
	loc := strings.TrimSpace(location)
	right := strings.TrimSpace(title)
	tail := 0
	if right != "" {
		tail = 1 + utf8.RuneCountInString(right)
	}
	// "░▒▓ " + name + " ▓▒░" is 8 runes plus the name.
	maxLoc := Cols - 8 - tail
	if maxLoc < 1 {
		right = ""
		tail = 0
		maxLoc = Cols - 8
	}
	if utf8.RuneCountInString(loc) > maxLoc {
		loc = clipRunes(loc, maxLoc)
	}
	gap := Cols - (8 + utf8.RuneCountInString(loc)) - tail
	if gap < 0 {
		gap = 0
	}
	var b strings.Builder
	b.WriteString(ansiRails)
	b.WriteString("░▒▓ ")
	b.WriteString(ansiHot)
	b.WriteString(loc)
	b.WriteString(ansiRails)
	b.WriteString(" ▓▒░")
	b.WriteString(strings.Repeat("░", gap))
	if right != "" {
		b.WriteString(ansiDim)
		b.WriteByte(' ')
		b.WriteString(right)
	}
	b.WriteString(ansiReset)
	return fit(b.String(), Cols)
}

func railsRule() string {
	return ansiRails + strings.Repeat("▀", Cols) + ansiReset
}

func titleBar(left, right string) string {
	gap := Cols - utf8.RuneCountInString(left) - utf8.RuneCountInString(right) - 2
	if gap < 1 {
		gap = 1
	}
	plain := " " + left + strings.Repeat(" ", gap) + right + " "
	if utf8.RuneCountInString(plain) > Cols {
		plain = clipRunes(plain, Cols)
	}
	if utf8.RuneCountInString(plain) < Cols {
		plain += strings.Repeat(" ", Cols-utf8.RuneCountInString(plain))
	}
	return ansiBar + plain + ansiReset
}

func clipRunes(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= width {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= width {
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

func fit(s string, width int) string {
	s = clipFrame(s, width)
	pad := width - frameWidth(s)
	if pad < 0 {
		pad = 0
	}
	return s + strings.Repeat(" ", pad)
}

func frameWidth(s string) int {
	_, n := filterFrame(s, -1, false)
	return n
}

func clipFrame(s string, width int) string {
	out, _ := filterFrame(s, width, true)
	return out
}

// framePlain drops every escape sequence. The title bar adds its own SGR.
func framePlain(s string) string {
	out, _ := filterFrame(s, -1, true)
	return stripSGR(out)
}

func stripSGR(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			j, _ := takeESC(s, i)
			i = j
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// filterFrame copies text for one frame row. CSI SGR is kept and has no
// width. OSC, DCS, other escapes, C1, CR, LF, and format controls are dropped.
func filterFrame(s string, width int, copy bool) (string, int) {
	var b strings.Builder
	vis := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			j, sgr := takeESC(s, i)
			if copy && sgr {
				b.WriteString(s[i:j])
			}
			i = j
			continue
		}
		if s[i] >= 0x80 && s[i] <= 0x9f {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if size < 1 {
			i++
			continue
		}
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		if dropFrameRune(r) {
			i += size
			continue
		}
		if width >= 0 && vis >= width {
			break
		}
		if copy {
			b.WriteString(s[i : i+size])
		}
		vis++
		i += size
	}
	if !copy {
		return "", vis
	}
	return b.String(), vis
}

func dropFrameRune(r rune) bool {
	if r == '\n' || r == '\r' || r == 0x7f {
		return true
	}
	if r < 0x20 && r != '\t' {
		return true
	}
	if r >= 0x80 && r <= 0x9f || unicode.Is(unicode.Cf, r) {
		return true
	}
	return false
}

// takeESC returns the index after this escape and whether it is CSI SGR.
func takeESC(s string, i int) (int, bool) {
	if i+1 >= len(s) {
		return len(s), false
	}
	switch s[i+1] {
	case '[':
		j := i + 2
		for j < len(s) {
			c := s[j]
			j++
			if c >= 0x40 && c <= 0x7e {
				return j, c == 'm'
			}
		}
		return j, false
	case ']':
		return skipST(s, i+2), false
	case 'P', 'X', '^', '_':
		return skipST(s, i+2), false
	default:
		if i+2 > len(s) {
			return len(s), false
		}
		return i + 2, false
	}
}

func skipST(s string, j int) int {
	for j < len(s) {
		if s[j] == 0x07 {
			return j + 1
		}
		if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
			return j + 2
		}
		j++
	}
	return j
}
