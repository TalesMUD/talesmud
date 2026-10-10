package termout

import "testing"

func FuzzTermoutEncode(f *testing.F) {
	f.Add("", "utf-8")
	f.Add("Harbor", "ascii")
	f.Add("░▒▓█", "cp437")
	f.Add("\x1b[2J\x1b[HHello", "cp437")
	f.Add("\x1b]0;title\x07", "ascii")
	f.Add("café", "ascii")
	f.Add("{{CODE}} at {{URL}} until {{EXPIRES}}", "utf-8")
	f.Add("\x1b", "cp437")
	f.Fuzz(func(t *testing.T, text, cs string) {
		charset, _ := parseCharset(cs)
		_ = Encode(text, charset)
		_ = Encode(text, CP437)
		_ = Encode(text, ASCII)
		_ = DecodeANS([]byte(text))
		_ = Substitute(text, "BCDFGHJK", "http://127.0.0.1/activate", "10m")
	})
}
