package termout

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	frameCols = 80
	frameRows = 25

	enterAlt = "\x1b[?1049h\x1b[?25l"
	leaveAlt = "\x1b[?25h\x1b[0m\x1b[?1049l"
	syncOn   = "\x1b[?2026h"
	syncOff  = "\x1b[?2026l"
	clearSeq = "\x1b[2J\x1b[H"
)

// Screen remembers the last painted rows so a later frame can skip them.
// Use it from one goroutine.
type Screen struct {
	cols     int
	rows     int
	charset  Charset
	prev     [frameRows][]byte
	dirty    [frameRows]bool
	alt      bool
	ready    bool
	tooSmall bool
	fill     rune
}

// NewScreen starts on the normal screen. The first Paint enters the alt screen.
func NewScreen() *Screen { return &Screen{} }

// SetFill paints one character outside the centred frame. Empty and control
// text are ignored, which leaves the terminal's black background.
func (s *Screen) SetFill(fill string) {
	if s == nil {
		return
	}
	s.fill = 0
	if fill == "" {
		return
	}
	r, size := utf8.DecodeRuneInString(fill)
	if size != len(fill) || r == utf8.RuneError || r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
		return
	}
	s.fill = r
}

// Fits reports whether the window can hold an 80x25 frame.
func Fits(cols, rows int) bool { return cols >= frameCols && rows >= frameRows }

// Offsets is the 0-based letterbox origin. Negative when the window is smaller
// than the frame.
func Offsets(cols, rows int) (top, left int) {
	return (rows - frameRows) / 2, (cols - frameCols) / 2
}

// Invalidate marks one 0-based frame row so the next Paint rewrites it.
func (s *Screen) Invalidate(row int) {
	if s == nil || row < 0 || row >= frameRows {
		return
	}
	s.dirty[row] = true
}

// Exit leaves the alternate screen. A second call writes nothing.
func (s *Screen) Exit() []byte {
	if s == nil || !s.alt {
		return nil
	}
	s.alt = false
	s.ready = false
	return []byte(leaveAlt)
}

// Paint centres the frame, or the too-small notice when the window cannot
// hold it. force rewrites every row. A nil return means nothing changed.
func (s *Screen) Paint(ansi string, cols, rows int, cs Charset, force bool) []byte {
	if s == nil {
		return nil
	}
	if cols <= 0 {
		cols = frameCols
	}
	if rows <= 0 {
		rows = 24
	}
	if !Fits(cols, rows) {
		return s.paintSmall(cols, rows, force)
	}
	if cs == "" {
		cs = UTF8
	}
	rowsText := splitFrame(ansi)
	var encoded [frameRows][]byte
	for i, row := range rowsText {
		encoded[i] = Encode(row, cs)
	}
	top, left := Offsets(cols, rows)
	if !s.ready || s.tooSmall || s.cols != cols || s.rows != rows || s.charset != cs {
		force = true
	}
	var enter []byte
	if !s.alt {
		enter = []byte(enterAlt)
		s.alt = true
		force = true
	}
	var ops bytes.Buffer
	if force {
		ops.WriteString("\x1b[2J")
		if s.fill != 0 {
			cell := Encode(string(s.fill), cs)
			for r := 1; r <= rows; r++ {
				fmt.Fprintf(&ops, "\x1b[%d;1H", r)
				for c := 0; c < cols; c++ {
					ops.Write(cell)
				}
			}
		}
	}
	for i := 0; i < frameRows; i++ {
		if !force && !s.dirty[i] && bytes.Equal(encoded[i], s.prev[i]) {
			continue
		}
		fmt.Fprintf(&ops, "\x1b[%d;%dH", top+1+i, left+1)
		ops.Write(encoded[i])
		s.dirty[i] = false
	}
	s.prev = encoded
	s.cols, s.rows, s.charset = cols, rows, cs
	s.tooSmall = false
	s.ready = true
	if ops.Len() == 0 {
		return enter
	}
	var out bytes.Buffer
	out.Write(enter)
	out.WriteString(syncOn)
	out.Write(ops.Bytes())
	out.WriteString(syncOff)
	return out.Bytes()
}

func (s *Screen) paintSmall(cols, rows int, force bool) []byte {
	if s.ready && s.tooSmall && s.cols == cols && s.rows == rows && !force {
		return nil
	}
	msg := fmt.Sprintf("This door is 80x25. Your window is %dx%d. Make it a little bigger, or press Q to leave.", cols, rows)
	var enter []byte
	if !s.alt {
		enter = []byte(enterAlt)
		s.alt = true
	}
	var out bytes.Buffer
	out.Write(enter)
	out.WriteString(syncOn)
	out.WriteString("\x1b[2J\x1b[H")
	out.WriteString(msg)
	out.WriteString(syncOff)
	s.cols, s.rows = cols, rows
	s.tooSmall = true
	s.ready = true
	for i := range s.prev {
		s.prev[i] = nil
		s.dirty[i] = true
	}
	return out.Bytes()
}

// EchoLine draws the composer on the frame's prompt row (frame row 23).
func EchoLine(shown string, cols, rows int, cs Charset) []byte {
	if !Fits(cols, rows) {
		return nil
	}
	top, left := Offsets(cols, rows)
	var b bytes.Buffer
	fmt.Fprintf(&b, "\x1b[%d;%dH\x1b[2K", top+23, left+1)
	b.Write(Encode(shown, cs))
	return b.Bytes()
}

// GenericSplash is the built-in screen used when the operator did not
// configure a pack file, or the file cannot be read. It names no world.
func GenericSplash() string {
	lines := make([]string, frameRows)
	lines[8] = "TalesMUD"
	lines[10] = "do you see ░▒▓█ correctly?"
	lines[12] = "[Enter] yes    [C] CP437    [Q] quit"
	return strings.Join(lines, "\r\n")
}

func splitFrame(ansi string) [frameRows]string {
	s := ansi
	if strings.HasPrefix(s, clearSeq) {
		s = s[len(clearSeq):]
	}
	var parts []string
	switch {
	case strings.Contains(s, "\r\n"):
		parts = strings.Split(s, "\r\n")
	default:
		parts = strings.Split(s, "\n")
	}
	var rows [frameRows]string
	for i := 0; i < frameRows; i++ {
		src := ""
		if i < len(parts) {
			src = parts[i]
		}
		rows[i] = padClip(src, frameCols)
	}
	return rows
}

// WrapText breaks text on spaces so each line's visible width stays within
// width. A single word longer than width is split by rune.
func WrapText(s string, width int) []string {
	s = strings.TrimRight(s, "\r\n")
	if width < 8 {
		width = 8
	}
	var lines []string
	for _, para := range strings.Split(s, "\n") {
		para = strings.TrimRight(para, "\r")
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		cur := ""
		for _, w := range words {
			for visibleLen(w) > width {
				if cur != "" {
					lines = append(lines, cur)
					cur = ""
				}
				chunk := clipVisible(w, width)
				lines = append(lines, chunk)
				w = w[len(chunk):]
			}
			if w == "" {
				continue
			}
			if cur == "" {
				cur = w
				continue
			}
			if visibleLen(cur)+1+visibleLen(w) <= width {
				cur += " " + w
				continue
			}
			lines = append(lines, cur)
			cur = w
		}
		if cur != "" {
			lines = append(lines, cur)
		}
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func padClip(s string, width int) string {
	s = clipVisible(s, width)
	n := visibleLen(s)
	if n < width {
		s += strings.Repeat(" ", width-n)
	}
	return s
}

func visibleLen(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			i = skipESC([]byte(s), i)
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
	if width <= 0 {
		return ""
	}
	var b strings.Builder
	vis := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			j := skipESC([]byte(s), i)
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
