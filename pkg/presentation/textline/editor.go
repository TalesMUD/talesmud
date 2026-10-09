package textline

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Editor is the classic SSH line editor. Call it from one goroutine.
// It echoes, supports backspace, Ctrl-U, and a short history. Ctrl-C and
// Ctrl-D ask before quitting. A character-list choice turns a number into
// "selectcharacter <name>".
type Editor struct {
	prompt     string
	line       []rune
	saved      []rune
	history    []string
	hist       int
	maxHist    int
	maxLine    int
	confirm    bool
	esc        []byte
	choices    []string
	choiceKind string
}

// FeedResult is what one chunk of terminal input did.
type FeedResult struct {
	Out  []byte
	Line string
	Quit bool
}

// NewEditor builds an editor. history is the number of remembered lines.
// maxLine is the rune cap (512 for classic play).
func NewEditor(history, maxLine int) *Editor {
	if history < 1 {
		history = 20
	}
	if maxLine < 1 {
		maxLine = 512
	}
	return &Editor{
		prompt:  "> ",
		hist:    -1,
		maxHist: history,
		maxLine: maxLine,
	}
}

// SetChoices remembers a numbered list. kind "character" rewrites a bare
// number into a selectcharacter command. Any other kind leaves the number
// alone so a dialog option still submits as that number.
func (e *Editor) SetChoices(kind string, choices []string) {
	if e == nil {
		return
	}
	e.choiceKind = kind
	e.choices = append([]string(nil), choices...)
}

// PromptLine is the current prompt plus the text the player has typed.
func (e *Editor) PromptLine() string {
	if e == nil {
		return ""
	}
	if e.confirm {
		return "Quit? [Y/N] "
	}
	return e.prompt + string(e.line)
}

// PrintAbove returns bytes that clear the input row, show text, and redraw.
func (e *Editor) PrintAbove(text string) []byte {
	if e == nil {
		return nil
	}
	text = strings.TrimRight(text, "\r\n")
	if text == "" {
		return e.redraw()
	}
	var b strings.Builder
	b.WriteString("\r\x1b[2K")
	b.WriteString(strings.ReplaceAll(text, "\n", "\r\n"))
	b.WriteString("\r\n")
	b.Write(e.redraw())
	return []byte(b.String())
}

// Redraw returns the current input row.
func (e *Editor) Redraw() []byte {
	if e == nil {
		return nil
	}
	return e.redraw()
}

func (e *Editor) redraw() []byte {
	return []byte("\r\x1b[2K" + e.PromptLine())
}

// Feed handles raw terminal bytes.
func (e *Editor) Feed(data []byte) FeedResult {
	var out FeedResult
	if e == nil || len(data) == 0 {
		return out
	}
	var buf []byte
	for len(data) > 0 {
		if len(e.esc) > 0 || data[0] == 0x1b {
			ev, rest, ok := e.takeEscape(data)
			data = rest
			if !ok {
				if len(data) == 0 {
					break
				}
				continue
			}
			step, submit, quit := e.apply(ev)
			buf = append(buf, step...)
			if submit {
				out.Line = e.takeSubmit()
			}
			out.Quit = out.Quit || quit
			continue
		}
		b := data[0]
		data = data[1:]
		if b == '\n' && len(buf) > 0 && strings.HasSuffix(string(buf), "\r\n") {
			continue
		}
		var ev editEvent
		switch b {
		case '\r', '\n':
			ev.kind = evEnter
		case 0x7f, 0x08:
			ev.kind = evBackspace
		case 0x15:
			ev.kind = evKill
		case 0x03, 0x04:
			ev.kind = evQuit
		default:
			if b < 0x20 {
				continue
			}
			// ASCII fast path. Multi-byte UTF-8 is collected below.
			if b < 0x80 {
				ev.kind = evRune
				ev.r = rune(b)
				break
			}
			need := utf8RuneLen(b)
			raw := []byte{b}
			for len(raw) < need && len(data) > 0 && data[0] < 0x80 == false && data[0]&0xC0 == 0x80 {
				raw = append(raw, data[0])
				data = data[1:]
			}
			r, size := utf8.DecodeRune(raw)
			if r == utf8.RuneError || size != len(raw) {
				continue
			}
			ev.kind = evRune
			ev.r = r
		}
		step, submit, quit := e.apply(ev)
		buf = append(buf, step...)
		if submit {
			out.Line = e.takeSubmit()
		}
		out.Quit = out.Quit || quit
	}
	out.Out = buf
	return out
}

func utf8RuneLen(b byte) int {
	switch {
	case b&0xE0 == 0xC0:
		return 2
	case b&0xF0 == 0xE0:
		return 3
	case b&0xF8 == 0xF0:
		return 4
	default:
		return 1
	}
}

type editKind int

const (
	evNone editKind = iota
	evRune
	evEnter
	evBackspace
	evKill
	evQuit
	evHistUp
	evHistDown
)

type editEvent struct {
	kind editKind
	r    rune
}

func (e *Editor) takeEscape(data []byte) (editEvent, []byte, bool) {
	if len(data) == 0 {
		return editEvent{}, data, false
	}
	e.esc = append(e.esc, data[0])
	data = data[1:]
	seq := e.esc
	if len(seq) == 1 {
		return editEvent{}, data, false
	}
	// ESC [ A / B or ESC O A / B. Longer CSI is swallowed.
	if len(seq) >= 3 && (seq[1] == '[' || seq[1] == 'O') {
		switch seq[2] {
		case 'A':
			e.esc = nil
			return editEvent{kind: evHistUp}, data, true
		case 'B':
			e.esc = nil
			return editEvent{kind: evHistDown}, data, true
		case 'C', 'D':
			e.esc = nil
			return editEvent{}, data, true
		}
		if seq[1] == '[' && ((seq[2] >= '0' && seq[2] <= '9') || seq[2] == '?' || seq[2] == ';') {
			last := seq[len(seq)-1]
			if (last >= 'A' && last <= 'Z') || (last >= 'a' && last <= 'z') || last == '~' {
				e.esc = nil
				return editEvent{}, data, true
			}
			if len(seq) > 16 {
				e.esc = nil
				return editEvent{}, data, true
			}
			return editEvent{}, data, false
		}
	}
	if len(seq) == 2 && (seq[1] == '[' || seq[1] == 'O') {
		return editEvent{}, data, false
	}
	e.esc = nil
	return editEvent{}, data, true
}

func (e *Editor) apply(ev editEvent) (out []byte, submit, quit bool) {
	if e.confirm {
		switch ev.kind {
		case evRune:
			switch ev.r {
			case 'y', 'Y':
				e.confirm = false
				e.line = nil
				return []byte("\r\n"), false, true
			case 'n', 'N':
				e.confirm = false
				return e.redraw(), false, false
			}
		case evEnter:
			e.confirm = false
			return e.redraw(), false, false
		}
		return nil, false, false
	}
	switch ev.kind {
	case evQuit:
		e.confirm = true
		return e.redraw(), false, false
	case evEnter:
		return []byte("\r\n"), true, false
	case evBackspace:
		if len(e.line) > 0 {
			e.line = e.line[:len(e.line)-1]
			e.hist = -1
		}
		return e.redraw(), false, false
	case evKill:
		e.line = nil
		e.hist = -1
		return e.redraw(), false, false
	case evHistUp:
		e.historyMove(-1)
		return e.redraw(), false, false
	case evHistDown:
		e.historyMove(1)
		return e.redraw(), false, false
	case evRune:
		if ev.r < 0x20 || ev.r == 0x7f {
			return nil, false, false
		}
		if len(e.line) >= e.maxLine {
			return nil, false, false
		}
		e.line = append(e.line, ev.r)
		e.hist = -1
		return e.redraw(), false, false
	default:
		return nil, false, false
	}
}

func (e *Editor) historyMove(dir int) {
	if len(e.history) == 0 {
		return
	}
	if e.hist == -1 {
		e.saved = append([]rune(nil), e.line...)
		if dir < 0 {
			e.hist = len(e.history) - 1
		} else {
			return
		}
	} else {
		e.hist += dir
		if e.hist >= len(e.history) {
			e.hist = -1
			e.line = append([]rune(nil), e.saved...)
			return
		}
		if e.hist < 0 {
			e.hist = 0
		}
	}
	e.line = []rune(e.history[e.hist])
}

func (e *Editor) takeSubmit() string {
	line := Sanitize(strings.TrimSpace(string(e.line)))
	e.line = nil
	e.hist = -1
	e.saved = nil
	if line != "" {
		if len(e.history) == 0 || e.history[len(e.history)-1] != line {
			e.history = append(e.history, line)
			if len(e.history) > e.maxHist {
				e.history = e.history[len(e.history)-e.maxHist:]
			}
		}
	}
	if e.choiceKind == "character" {
		if n, err := strconv.Atoi(line); err == nil && n >= 1 && n <= len(e.choices) {
			return "selectcharacter " + e.choices[n-1]
		}
	}
	return line
}
