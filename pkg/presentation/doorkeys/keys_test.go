package doorkeys

import "testing"

func TestOnDataParity(t *testing.T) {
	p := New()
	cases := []struct {
		in        string
		send      bool
		key       string
		echo      bool
		shown     string
		composing bool
	}{
		{in: "n", send: true, key: "n"},
		{in: "N", send: true, key: "N"},
		{in: "5", send: true, key: "5"},
		{in: "?", send: true, key: "?"},
		{in: "\r", send: true, key: ""},
		{in: "\n", send: true, key: ""},
		{in: "\r\n", send: true, key: ""},
		{in: " ", send: true, key: ""},
		{in: "!", send: true, key: ""},
		{in: "\x1b[A"},
		{in: "ab"},
		{in: "\x1b"},
		{in: ":", echo: true, shown: ": ", composing: true},
	}
	for _, tc := range cases {
		got := p.Feed(tc.in)
		if got.Send != tc.send || got.Key != tc.key || got.Echo != tc.echo || p.composing != tc.composing {
			t.Fatalf("hotkey %q → %+v composing %v", tc.in, got, p.composing)
		}
		if tc.echo && p.Shown() != tc.shown {
			t.Fatalf("shown %q want %q", p.Shown(), tc.shown)
		}
	}

	if got := p.Feed("n"); got.Send || !got.Echo || p.Shown() != ": n" {
		t.Fatalf("compose %q %+v", p.Shown(), got)
	}
	if got := p.Feed("\r"); !got.Send || got.Key != "n" || p.composing {
		t.Fatalf("compose enter %+v", got)
	}

	p.Feed(":")
	if got := p.Feed("\r"); got.Send || p.composing {
		t.Fatalf("empty local enter %+v composing %v", got, p.composing)
	}

	p.OnFrame([]string{"line"}, "", "Say")
	if p.Shown() != "Say " {
		t.Fatal(p.Shown())
	}
	if got := p.Feed("a"); got.Send || p.Shown() != "Say a" {
		t.Fatalf("line %q %+v", p.Shown(), got)
	}
	if p.Feed("b"); p.line != "ab" {
		t.Fatal(p.line)
	}
	if got := p.Feed("\x7f"); p.line != "a" || !got.Echo {
		t.Fatalf("backspace %q", p.line)
	}
	if got := p.Feed("\r"); !got.Send || got.Key != "a" {
		t.Fatalf("line enter %+v", got)
	}
	if got := p.Feed("\r"); !got.Send || got.Key != "" {
		t.Fatalf("empty line enter %+v", got)
	}
	p.Feed("z")
	if got := p.Feed("\x1b"); !got.Send || got.Key != "x" || p.line != "" {
		t.Fatalf("line esc %+v line %q", got, p.line)
	}

	p.OnFrame([]string{"hotkey"}, "line", "")
	if p.mode != "hotkey" || p.prompt != ">" || p.line != "" {
		t.Fatalf("frame reset %+v", p)
	}
	p.Feed(":")
	p.line = "hi"
	if got := p.Feed("\x1b"); got.Send || p.composing || p.line != "" {
		t.Fatalf("compose esc %+v", got)
	}

	p.OnFrame(nil, "", "")
	if p.mode != "hotkey" || p.prompt != ">" {
		t.Fatal("default frame")
	}
	p.OnFrame(nil, "line", ">")
	long := ""
	for i := 0; i < 48; i++ {
		long += "x"
		if got := p.Feed("x"); got.Send {
			t.Fatal("capped send")
		}
	}
	if p.line != long {
		t.Fatalf("len %d", len(p.line))
	}
	if got := p.Feed("y"); got.Echo || p.line != long {
		t.Fatal("49th char")
	}
}
