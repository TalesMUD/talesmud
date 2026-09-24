package balance

import "testing"

func TestLevelGapNeutralAndClamp(t *testing.T) {
	cfg := &CombatBalanceConfig{LevelGap: defaultLevelGap()}

	neutral := levelGapModifiers(cfg, 5, 5)
	if neutral.Gap != 0 || neutral.HitBonus != 0 || neutral.CritChanceDelta != 0 || neutral.DamageMultiplier != 1 {
		t.Fatalf("equal levels should be neutral, got %+v", neutral)
	}
	unset := levelGapModifiers(cfg, 0, 0)
	if unset.DamageMultiplier != 1 || unset.HitBonus != 0 {
		t.Fatalf("unset levels should be neutral, got %+v", unset)
	}

	up := levelGapModifiers(cfg, 10, 4) // raw +6
	if up.Gap != 6 || up.HitBonus != 4 {
		t.Fatalf("expected clamped +6 / hit +4, got %+v", up)
	}
	if up.DamageMultiplier <= 1 || up.CritChanceDelta <= 0 {
		t.Fatalf("advantage should raise damage and crit, got %+v", up)
	}

	capped := levelGapModifiers(cfg, 100, 1)
	if capped.Gap != up.Gap || capped.HitBonus != up.HitBonus || capped.DamageMultiplier != up.DamageMultiplier || capped.CritChanceDelta != up.CritChanceDelta {
		t.Fatalf("gap past max_levels should match +6, capped %+v up %+v", capped, up)
	}

	down := levelGapModifiers(cfg, 1, 7) // raw -6
	if down.Gap != -6 || down.HitBonus != -4 || down.CritChanceDelta >= 0 || down.DamageMultiplier >= 1 {
		t.Fatalf("disadvantage should lower hit, crit, and damage, got %+v", down)
	}
	floor := levelGapModifiers(cfg, 1, 100)
	if floor.Gap != down.Gap || floor.DamageMultiplier != down.DamageMultiplier {
		t.Fatalf("gap below -max should match -6, floor %+v down %+v", floor, down)
	}
}

func TestLevelGapPlusThreeFeel(t *testing.T) {
	cfg := &CombatBalanceConfig{LevelGap: defaultLevelGap()}
	// Attacker 3 levels above the defender: harder for the lower side, not a wipe by itself.
	plus := levelGapModifiers(cfg, 8, 5)
	if plus.Gap != 3 || plus.HitBonus != 2 {
		t.Fatalf("+3 gap: %+v", plus)
	}
	if plus.DamageDealtMult < 1.10 || plus.DamageDealtMult > 1.11 {
		t.Fatalf("damage dealt at +3 = %v, want ~1.105", plus.DamageDealtMult)
	}
	if plus.DamageTakenMult < 1.05 || plus.DamageTakenMult > 1.07 {
		t.Fatalf("damage taken at +3 = %v, want ~1.06", plus.DamageTakenMult)
	}
	if plus.DamageMultiplier < 1.16 || plus.DamageMultiplier > 1.19 {
		t.Fatalf("combined damage at +3 = %v, want ~1.17", plus.DamageMultiplier)
	}
	if plus.CritChanceDelta < 0.029 || plus.CritChanceDelta > 0.031 {
		t.Fatalf("crit delta at +3 = %v, want 0.03", plus.CritChanceDelta)
	}

	minus := levelGapModifiers(cfg, 5, 8)
	if minus.Gap != -3 || minus.HitBonus != -2 || minus.DamageMultiplier >= 1 {
		t.Fatalf("-3 gap: %+v", minus)
	}
	// +3 worth of damage dealt by the higher side and taken by the lower side
	// should be the inverse situation, not a second independent system.
	if minus.DamageMultiplier <= 0 || minus.DamageMultiplier >= 0.90 {
		t.Fatalf("-3 combined damage = %v, want about 0.84", minus.DamageMultiplier)
	}
}

func TestLevelGapCustomClamp(t *testing.T) {
	cfg := defaultLevelGap()
	cfg.MaxLevels = 4
	mods := levelGapModifiers(&CombatBalanceConfig{LevelGap: cfg}, 30, 1)
	if mods.Gap != 4 {
		t.Fatalf("custom max 4, got gap %d", mods.Gap)
	}
}

func TestLevelGapMissingSectionUsesDefaults(t *testing.T) {
	mods := levelGapModifiers(&CombatBalanceConfig{}, 10, 4)
	if mods.Gap != 6 || mods.DamageMultiplier <= 1 {
		t.Fatalf("missing level_gap should use defaults, got %+v", mods)
	}
}

func TestLevelGapConfigFileWired(t *testing.T) {
	if _, err := ReloadConfig(); err != nil {
		t.Fatalf("reload: %v", err)
	}
	cfg := GetConfig()
	if cfg == nil || cfg.LevelGap.MaxLevels != 6 {
		t.Fatalf("yaml max_levels = %+v", cfg)
	}
	if cfg.LevelGap.PerLevel.HitChance != 0.035 || cfg.LevelGap.PerLevel.DamageDealt != 0.035 {
		t.Fatalf("yaml per_level = %+v", cfg.LevelGap.PerLevel)
	}
	// Public API reads the same file.
	mods := LevelGapModifiers(9, 3)
	if mods.Gap != 6 {
		t.Fatalf("LevelGapModifiers gap = %d", mods.Gap)
	}
	scaled := ScaleDamage(9, 3, 10)
	want := int32(14) // round(10 * 1.3552) at the tuned +6 clamp
	if scaled != want {
		t.Fatalf("ScaleDamage(10) at +6 = %d, want %d (mult %v)", scaled, want, mods.DamageMultiplier)
	}
	if ScaleDamage(3, 3, 10) != 10 {
		t.Fatalf("neutral scale changed damage")
	}
	if ScaleDamage(1, 20, 1) < 1 {
		t.Fatal("scaled damage dropped below 1")
	}
}
