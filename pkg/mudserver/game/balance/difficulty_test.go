package balance

import (
	"testing"
)

func TestApplyEnemyMultipliersNamedOverride(t *testing.T) {
	ReloadConfig()
	// Hollow Knight named override should differ from generic boss tier
	bossHP, bossATK, bossDEF := ApplyEnemyMultipliers(55, 8, 3, "boss", "Burrow Brute")
	hkHP, hkATK, hkDEF := ApplyEnemyMultipliers(150, 13, 6, "boss", "The Hollow Knight")
	hk2HP, _, _ := ApplyEnemyMultipliers(150, 13, 6, "boss", "Hollow Knight")

	if bossHP < 90 || bossHP > 110 {
		t.Errorf("Burrow Brute HP %d outside expected ~101", bossHP)
	}
	if hkHP < 140 || hkHP > 160 {
		t.Errorf("Hollow Knight HP %d outside expected ~150", hkHP)
	}
	if hkHP != hk2HP {
		t.Errorf("name alias mismatch: %d vs %d", hkHP, hk2HP)
	}
	if hkATK >= bossATK*2 {
		t.Errorf("HK ATK %d unexpectedly huge vs brute %d", hkATK, bossATK)
	}
	_ = bossDEF
	_ = hkDEF
}

func TestTrivialTrashHP(t *testing.T) {
	ReloadConfig()
	hp, atk, def := ApplyEnemyMultipliers(8, 1, 0, "trivial", "Catacomb Rat")
	if hp < 18 || hp > 24 {
		t.Errorf("Catacomb Rat HP %d want ~20", hp)
	}
	if atk != 3 {
		t.Errorf("Catacomb Rat ATK %d want 3", atk)
	}
	if def != 0 {
		t.Errorf("Catacomb Rat DEF %d want 0", def)
	}
}

func TestEarlyNormalCurveOnce(t *testing.T) {
	ReloadConfig()
	// Level 0 is the no-curve path. An imported Alley Thug used to show 48 HP
	// because normal.hp is 2.7. The early curve applies only when the level is known.
	hp, atk, def := ApplyEnemyMultipliers(18, 4, 1, "normal", "Alley Thug")
	if hp != 48 || atk != 5 || def != 1 {
		t.Fatalf("level-unknown thug %d/%d/%d, want 48/5/1", hp, atk, def)
	}
	hp, atk, def = ApplyEnemyCombat(18, 4, 1, "normal", "Alley Thug", 2)
	if hp != 20 || atk != 7 || def != 1 {
		t.Fatalf("level-2 thug %d/%d/%d, want 20/7/1", hp, atk, def)
	}
	// A second call on the already-scaled numbers is the double-apply bug.
	again, _, _ := ApplyEnemyCombat(hp, atk, def, "normal", "Alley Thug", 2)
	if again == hp {
		t.Fatalf("scaled thug HP %d did not move; the test no longer catches a double apply", hp)
	}
	hp, atk, def = ApplyEnemyCombat(30, 8, 2, "normal", "Marsh Spider", 9)
	if hp != 81 || atk != 10 || def != 3 {
		t.Fatalf("level-9 spider %d/%d/%d, want 81/10/3", hp, atk, def)
	}
	hp, atk, def = ApplyEnemyCombat(80, 11, 4, "hard", "Thornback Bear", 5)
	if hp != 72 || atk != 11 || def != 4 {
		t.Fatalf("bear %d/%d/%d, want 72/11/4", hp, atk, def)
	}
	hp, atk, def = ApplyEnemyCombat(8, 1, 0, "trivial", "Catacomb Rat", 1)
	if hp != 20 || atk != 3 || def != 0 {
		t.Fatalf("rat %d/%d/%d, want 20/3/0", hp, atk, def)
	}
}
