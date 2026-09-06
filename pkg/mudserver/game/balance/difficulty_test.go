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
