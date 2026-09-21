package door

import (
	"strings"
	"unicode/utf8"
)

const (
	Cols = 80
	Rows = 25

	ansiReset = "\x1b[0m"
	ansiTitle = "\x1b[1;36m"
	ansiKey   = "\x1b[1;33m"
	ansiDim   = "\x1b[0;37m"
	ansiGreen = "\x1b[1;32m"
	ansiRed   = "\x1b[1;31m"
	ansiBar   = "\x1b[30;46m"
)

// Frame is one full-screen ANSI page pushed over the existing WebSocket.
// Field names follow the door_frame contract: type, cols, rows, ansi, prompt, screen, accepts, keys.
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

type view struct {
	Location  string
	Body      []string
	Prompt    string
	Footer    string
	InputMode string
	ScreenID  string
	Keys      map[string]string
	// Lines, when set to Rows entries, is the full grid (prompt and footer already placed).
	Lines []string
}

func (v view) frame() Frame {
	mode := v.InputMode
	if mode == "" {
		mode = "hotkey"
	}
	accepts := []string{"hotkey"}
	if mode == "line" {
		accepts = []string{"line"}
	}
	if len(v.Lines) == Rows {
		lines := make([]string, Rows)
		for i, ln := range v.Lines {
			lines[i] = fitPlain(ln, Cols)
		}
		return finishFrame(v, mode, accepts, lines)
	}
	lines := make([]string, Rows)
	lines[0] = titleBar("AETHERMOOR DOOR", "VEILSPAN")
	lines[1] = fitPlain(ansiTitle+v.Location+ansiReset, Cols)
	lines[2] = fitPlain(ansiDim+strings.Repeat("-", Cols)+ansiReset, Cols)
	body := v.Body
	const bodyRows = 19
	if len(body) > bodyRows {
		body = body[:bodyRows]
	}
	for i := 0; i < bodyRows; i++ {
		src := ""
		if i < len(body) {
			src = body[i]
		}
		lines[3+i] = fitPlain(src, Cols)
	}
	lines[22] = fitPlain(v.Prompt, Cols)
	lines[23] = fitPlain(ansiDim+v.Footer+ansiReset, Cols)
	lines[24] = fitPlain("", Cols)

	return finishFrame(v, mode, accepts, lines)
}

func finishFrame(v view, mode string, accepts []string, lines []string) Frame {
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
		Prompt:    stripANSI(v.Prompt),
		Screen:    v.ScreenID,
		Accepts:   accepts,
		Keys:      v.Keys,
	}
}

func titleBar(left, right string) string {
	gap := Cols - visibleLen(left) - visibleLen(right) - 2
	if gap < 1 {
		gap = 1
	}
	plain := " " + left + strings.Repeat(" ", gap) + right + " "
	if visibleLen(plain) > Cols {
		plain = plain[:Cols]
	}
	if visibleLen(plain) < Cols {
		plain += strings.Repeat(" ", Cols-visibleLen(plain))
	}
	return ansiBar + plain + ansiReset
}

func fitPlain(s string, width int) string {
	if visibleLen(s) > width {
		s = clipVisible(s, width)
	}
	if !strings.Contains(s, "\x1b") {
		if len(s) > width {
			s = s[:width]
		}
		return s + strings.Repeat(" ", width-len(s))
	}
	pad := width - visibleLen(s)
	if pad < 0 {
		pad = 0
	}
	return s + ansiReset + strings.Repeat(" ", pad)
}

func clipVisible(s string, width int) string {
	if width <= 0 {
		return ""
	}
	var b strings.Builder
	vis := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && !isFinal(s[j]) {
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
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			size = 1
		}
		b.WriteString(s[i : i+size])
		vis++
		i += size
	}
	b.WriteString(ansiReset)
	return b.String()
}

func visibleLen(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && !isFinal(s[i]) {
				i++
			}
			if i < len(s) {
				i++
			}
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

func isFinal(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && !isFinal(s[i]) {
				i++
			}
			if i < len(s) {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func wrapText(s string, width int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	words := strings.Fields(s)
	var lines []string
	cur := ""
	for _, w := range words {
		if cur == "" {
			cur = w
			continue
		}
		if len(cur)+1+len(w) > width {
			lines = append(lines, cur)
			cur = w
			continue
		}
		cur += " " + w
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

func hotkey(key, label, hint string) string {
	text := "  " + ansiKey + "(" + key + ")" + ansiReset + " " + label
	if hint != "" {
		text += "  " + ansiDim + hint + ansiReset
	}
	return text
}

func hpBar(cur, max int) string {
	const w = 10
	if max < 1 {
		max = 1
	}
	if cur < 0 {
		cur = 0
	}
	if cur > max {
		cur = max
	}
	filled := cur * w / max
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", w-filled) + "]"
}
