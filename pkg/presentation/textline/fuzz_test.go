package textline

import "testing"

func FuzzLineEditor(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("look\r"))
	f.Add([]byte("say hi\n"))
	f.Add([]byte("\x1b[A"))
	f.Add([]byte("\x1b[B"))
	f.Add([]byte("\x1bOA"))
	f.Add([]byte("\x1b[200~pasted\x1b[201~"))
	f.Add([]byte("\x1b["))
	f.Add([]byte("\x7f\x08\x15\x03\x04"))
	f.Add([]byte("café\r"))
	f.Add([]byte{0x1b, '[', '1', ';', '5', 'H', 'x', '\r'})
	f.Fuzz(func(t *testing.T, data []byte) {
		e := NewEditor(4, 32)
		_ = e.Feed(data)
		e = NewEditor(4, 32)
		for i := 0; i < len(data); i++ {
			_ = e.Feed(data[i : i+1])
			if len(e.esc) > 32 {
				t.Fatalf("escape buffer %d", len(e.esc))
			}
		}
		_ = e.Redraw()
	})
}

func FuzzEscapeParser(f *testing.F) {
	f.Add([]byte("\x1b[A"))
	f.Add([]byte("\x1b[B"))
	f.Add([]byte("\x1b[C"))
	f.Add([]byte("\x1b[D"))
	f.Add([]byte("\x1bOA"))
	f.Add([]byte("\x1b["))
	f.Add([]byte("\x1b"))
	f.Add([]byte("\x1b[200~"))
	f.Add([]byte("\x1b[?25h"))
	f.Add([]byte("\x1b[1;1H"))
	f.Add([]byte{0x1b, 0x00, 'A'})
	f.Fuzz(func(t *testing.T, data []byte) {
		e := NewEditor(4, 32)
		rest := data
		for len(rest) > 0 {
			_, next, _ := e.takeEscape(rest)
			if len(next) > len(rest) {
				t.Fatal("escape parser grew the buffer")
			}
			if len(e.esc) > 32 {
				t.Fatalf("escape buffer %d", len(e.esc))
			}
			if len(next) == len(rest) {
				rest = rest[1:]
				continue
			}
			rest = next
		}
	})
}

func FuzzSanitize(f *testing.F) {
	f.Add("")
	f.Add("hello")
	f.Add("line\r\n\tmore")
	f.Add("\x1b]0;title\x07name")
	f.Add("\x1b[31mred\x1b[0m")
	f.Add("caf\u00e9")
	f.Add("\u009b31m")
	f.Add("del\x7fend")
	f.Fuzz(func(t *testing.T, s string) {
		out := Sanitize(s)
		for _, r := range out {
			if r == 0x1b || r == 0x7f || (r < 0x20 && r != '\t' && r != '\n' && r != '\r') || (r >= 0x80 && r <= 0x9f) {
				t.Fatalf("control survived: %q", out)
			}
		}
	})
}
