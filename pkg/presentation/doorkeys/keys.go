// Package doorkeys is the Go port of the web client's onData handler.
// One Feed call is one onData chunk: a multi-byte paste is not split into keys.
package doorkeys

// Parser is the hotkey / colon composer / line-mode state.
// Call it from one goroutine.
type Parser struct {
	mode      string
	composing bool
	line      string
	prompt    string
}

// Result is what one chunk did. Echo means the caller should redraw Shown.
type Result struct {
	Send bool
	Key  string
	Echo bool
}

// New starts in hotkey mode with the default prompt.
func New() *Parser {
	return &Parser{mode: "hotkey", prompt: ">"}
}

// Echoing reports whether the composer should be drawn after a frame.
func (p *Parser) Echoing() bool {
	if p == nil {
		return false
	}
	return p.mode == "line" || p.composing
}

// Shown is the prompt-row text, capped at 79 runes, matching echoLine.
func (p *Parser) Shown() string {
	if p == nil {
		return ""
	}
	lead := p.prompt
	if p.composing && p.mode != "line" {
		lead = ":"
	}
	return clip79(lead + " " + p.line)
}

// OnFrame applies the accepts / inputMode / prompt rules from a door frame.
func (p *Parser) OnFrame(accepts []string, inputMode, prompt string) Result {
	if p == nil {
		return Result{}
	}
	switch {
	case len(accepts) > 0 && accepts[0] == "line":
		p.mode = "line"
	case len(accepts) > 0 && accepts[0] == "hotkey":
		p.mode = "hotkey"
	case inputMode != "":
		p.mode = inputMode
	default:
		p.mode = "hotkey"
	}
	if p.mode == "hotkey" {
		p.line = ""
		p.composing = false
	}
	if prompt != "" {
		p.prompt = prompt
	} else {
		p.prompt = ">"
	}
	return Result{Echo: p.Echoing()}
}

// Feed handles one terminal chunk the way door.js onData handles one data event.
// A chunk that is only CR LF is treated as Enter, because terminals send that
// as a single newline.
func (p *Parser) Feed(data string) Result {
	if p == nil {
		return Result{}
	}
	if data == "\r\n" || data == "\n\r" {
		data = "\r"
	}
	if p.mode == "line" || p.composing {
		if data == "\r" {
			text := p.line
			p.line = ""
			local := p.composing && p.mode != "line"
			p.composing = false
			if local && text == "" {
				return Result{}
			}
			return Result{Send: true, Key: text}
		}
		if data == "\x7f" || data == "\b" {
			if len(p.line) > 0 {
				p.line = p.line[:len(p.line)-1]
			}
			return Result{Echo: true}
		}
		if data == "\x1b" {
			p.line = ""
			wasLine := p.mode == "line"
			p.composing = false
			if wasLine {
				return Result{Send: true, Key: "x"}
			}
			return Result{}
		}
		if len(data) == 1 && data[0] >= ' ' && data[0] <= '~' && len(p.line) < 48 {
			p.line += data
			return Result{Echo: true}
		}
		return Result{}
	}
	if data == "\r" || data == "\n" {
		return Result{Send: true, Key: ""}
	}
	if data == ":" {
		p.composing = true
		p.line = ""
		return Result{Echo: true}
	}
	if len(data) == 1 && isHotkey(data[0]) {
		return Result{Send: true, Key: data}
	}
	if len(data) == 1 && data[0] >= ' ' && data[0] <= '~' {
		return Result{Send: true, Key: ""}
	}
	return Result{}
}

func isHotkey(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '?'
}

func clip79(s string) string {
	r := []rune(s)
	if len(r) > 79 {
		return string(r[:79])
	}
	return s
}
