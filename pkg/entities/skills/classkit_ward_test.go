package skills

import "testing"

func TestWardHotbarMigratesHitch(t *testing.T) {
	got := FillHotbar("hitch", 1, []string{"hitch_pin"})
	if len(got) != 2 || got[0] != "ward_guard" || got[1] != "ward_slam" {
		t.Fatalf("level 1 %v", got)
	}
	got = FillHotbar("hitch", 8, []string{"hitch_pin", "hitch_hobble", "hitch_reel"})
	if len(got) != 2 || got[0] != "ward_guard" || got[1] != "ward_slam" {
		t.Fatalf("level 8 %v", got)
	}
	if !IsKitClass("hitch") || !IsKitClass("ward") {
		t.Fatal("ward is not a kit class")
	}
	var guard, slam *Skill
	for _, s := range ClassKit() {
		if s == nil || s.Entity == nil {
			continue
		}
		switch s.Entity.ID {
		case "ward_guard":
			guard = s
		case "ward_slam":
			slam = s
		case "hitch_pin", "hitch_hobble", "hitch_reel":
			t.Fatalf("retired skill still in the kit: %s", s.Entity.ID)
		}
	}
	if guard == nil || guard.LevelRequired != 1 || !guard.KeepsSwing || !guard.OncePerFight || guard.Target != TargetAlly || guard.Kit != KitGuard {
		t.Fatalf("guard %+v", guard)
	}
	if slam == nil || slam.LevelRequired != 1 || slam.KeepsSwing || slam.CooldownRounds != 4 || slam.SwingMult != 1 || slam.Kit != KitSlam {
		t.Fatalf("slam %+v", slam)
	}
}
