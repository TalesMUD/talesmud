package balance

import (
	"testing"
)

func TestApplyEnemyMultipliersNamedOverride(t *testing.T) {
	ReloadConfig()
	// Sample Warden named override should differ from generic boss tier
	bossHP, bossATK, bossDEF := ApplyEnemyMultipliers(55, 8, 3, "boss", "Sample Brute")
	hkHP, hkATK, hkDEF := ApplyEnemyMultipliers(150, 13, 6, "boss", "Sample Warden")
	hk2HP, _, _ := ApplyEnemyMultipliers(150, 13, 6, "boss", "Sample Warden")

	if bossHP < 90 || bossHP > 110 {
		t.Errorf("Sample Brute HP %d outside expected ~101", bossHP)
	}
	if hkHP < 140 || hkHP > 160 {
		t.Errorf("Sample Warden HP %d outside expected ~150", hkHP)
	}
	if hkHP != hk2HP {
		t.Errorf("name alias mismatch: %d vs %d", hkHP, hk2HP)
	}
	if hkATK >= bossATK*2 {
		t.Errorf("Warden ATK %d unexpectedly huge vs brute %d", hkATK, bossATK)
	}
	_ = bossDEF
	_ = hkDEF
}

func TestTrivialTrashHP(t *testing.T) {
	ReloadConfig()
	hp, atk, def := ApplyEnemyMultipliers(8, 1, 0, "trivial", "Sample Rat")
	if hp < 18 || hp > 24 {
		t.Errorf("Sample Rat HP %d want ~20", hp)
	}
	if atk != 3 {
		t.Errorf("Sample Rat ATK %d want 3", atk)
	}
	if def != 0 {
		t.Errorf("Sample Rat DEF %d want 0", def)
	}
}
