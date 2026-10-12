package doorkeys

import "testing"

func FuzzDoorKeys(f *testing.F) {
	f.Add("")
	f.Add("q")
	f.Add(":")
	f.Add("look")
	f.Add("look\r")
	f.Add("\r")
	f.Add("\r\n")
	f.Add("\n\r")
	f.Add("\x1b")
	f.Add("\x1b[A")
	f.Add("\x7f")
	f.Add("north")
	f.Add("?")
	f.Add("a line that is longer than forty eight letters and should stop")
	f.Fuzz(func(t *testing.T, data string) {
		line := New()
		line.OnFrame([]string{"line"}, "line", ">")
		_ = line.Feed(data)
		_ = line.Shown()
		hot := New()
		hot.OnFrame([]string{"hotkey"}, "hotkey", "Say:")
		_ = hot.Feed(data)
		_ = hot.Shown()
		if len(line.line) > 48 || len(hot.line) > 48 {
			t.Fatalf("line grew past 48")
		}
	})
}
