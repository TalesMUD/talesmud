// Package ansi paints a fixed 80x25 text frame for the text client.
// It does not know prices, combat math, or world names.
package ansi

import (
	"strings"
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
	gap := Cols - len(left) - len(right) - 2
	if gap < 1 {
		gap = 1
	}
	plain := " " + left + strings.Repeat(" ", gap) + right + " "
	if len(plain) > Cols {
		plain = plain[:Cols]
	}
	if len(plain) < Cols {
		plain += strings.Repeat(" ", Cols-len(plain))
	}
	return ansiBar + plain + ansiReset
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
			j := i + 2
			for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
				j++
			}
			if j < len(s) {
				j++
			}
			i = j
			continue
		}
		n++
		i++
	}
	return n
}

func clipVisible(s string, width int) string {
	var b strings.Builder
	vis := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
				j++
			}
			if j < len(s) {
				j++
			}
			b.WriteString(s[i:j])
			i = j
			continue
		}
		if vis >= width {
			break
		}
		b.WriteByte(s[i])
		vis++
		i++
	}
	return b.String()
}
