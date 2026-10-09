package items

import "testing"

func TestDisplayEffectsAuthoredWins(t *testing.T) {
	item := &Item{OnHitScriptID: "SCR1", Effects: []ItemEffect{{Trigger: EffectTriggerOnHit, Name: "Burn", Text: "2 damage."}}}
	got := item.DisplayEffects(func(string) string { return "Burn Script" })
	if len(got) != 1 || got[0].Name != "Burn" {
		t.Fatalf("authored effect should be the only line, got %+v", got)
	}
	if line := FormatEffect(got[0]); line != "On hit — Burn: 2 damage." {
		t.Fatalf("line = %q", line)
	}
}

func TestDisplayEffectsScriptFallback(t *testing.T) {
	item := &Item{OnHitScriptID: "SCR1", OnUseScriptID: "SCR2"}
	got := item.DisplayEffects(func(id string) string {
		if id == "SCR1" {
			return "Vigil Burn On-Hit"
		}
		return ""
	})
	if len(got) != 2 {
		t.Fatalf("want two fallback lines, got %+v", got)
	}
	if got[0].Text != SpecialEffectText || got[0].Name != "Vigil Burn On-Hit" || got[0].Trigger != EffectTriggerOnHit {
		t.Fatalf("on-hit fallback = %+v", got[0])
	}
	if FormatEffect(got[1]) != "On use — Special effect" {
		t.Fatalf("on-use fallback = %q", FormatEffect(got[1]))
	}
	if (&Item{}).DisplayEffects(nil) != nil {
		t.Fatal("plain item has no effects")
	}
}

func TestCopyEffectsIsIndependent(t *testing.T) {
	src := []ItemEffect{{Text: "a"}}
	dst := CopyEffects(src)
	dst[0].Text = "b"
	if src[0].Text != "a" {
		t.Fatal("copy shares backing array")
	}
}
