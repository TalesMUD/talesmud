// Package ansi paints a fixed 80x25 text frame for the text client.
// It does not know prices, combat math, or world names.
package ansi

import (
	"strings"
	"unicode/utf8"
)

const (
	Cols = 80
	Rows = 25

	ansiReset = "\x1b[0m"
	ansiTitle = "\x1b[1;36m"
	ansiDim   = "\x1b[0;37m"
	ansiBar   = "\x1b[30;46m"
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
	lines[0] = titleBar(page.Title, "")
	lines[1] = fit(ansiTitle+page.Location+ansiReset, Cols)
	lines[2] = fit(ansiDim+strings.Repeat("-", Cols)+ansiReset, Cols)
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
	if visibleLen(s) > width {
		s = clipVisible(s, width)
	}
	pad := width - visibleLen(s)
	if pad < 0 {
		pad = 0
	}
	return s + strings.Repeat(" ", pad)
}

func visibleLen(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := skipANSI(s, i)
			i = j
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		if size < 1 {
			size = 1
		}
		n++
		i += size
	}
	return n
}

func clipVisible(s string, width int) string {
	var b strings.Builder
	vis := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := skipANSI(s, i)
			b.WriteString(s[i:j])
			i = j
			continue
		}
		if vis >= width {
			break
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		if size < 1 {
			size = 1
		}
		b.WriteString(s[i : i+size])
		vis++
		i += size
	}
	return b.String()
}

func skipANSI(s string, i int) int {
	j := i + 2
	for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
		j++
	}
	if j < len(s) {
		j++
	}
	return j
}
