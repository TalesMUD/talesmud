package devicecode

import "testing"

func FuzzUserCode(f *testing.F) {
	f.Add("")
	f.Add("BCDF-GHJK")
	f.Add("bcdfghjk")
	f.Add("BCDF GHJK")
	f.Add("BCDFGHJK")
	f.Add("AAAA-AAAA")
	f.Add("BCDF-GHJ")
	f.Add("BCDF-GHJK-EXTRA")
	f.Add("bc**-****")
	f.Fuzz(func(t *testing.T, code string) {
		norm, ok := Normalize(code)
		if ok && len(norm) != 8 {
			t.Fatalf("normalised %q", norm)
		}
		formatted := Format(norm)
		if ok && len(formatted) != 9 {
			t.Fatalf("formatted %q", formatted)
		}
		if !ok && formatted != "" && norm != "" {
			t.Fatalf("rejected code still formatted")
		}
		_ = Hash(norm)
	})
}
