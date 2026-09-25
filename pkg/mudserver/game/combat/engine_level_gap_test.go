package combat

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

func TestResolveAttackRollLevelGap(t *testing.T) {
	// roll 1 always misses, even with a large level bonus.
	hit, crit, toHit := ResolveAttackRoll(1, 0, 10, 0, 0.5, 0.05, 0)
	if hit || crit || toHit != 11 {
		t.Fatalf("natural 1 should miss, hit=%v crit=%v toHit=%d", hit, crit, toHit)
	}

	// Equal level: 10 vs AC 15 misses. +5 from level gap hits.
	hit, crit, _ = ResolveAttackRoll(10, 0, 0, 15, 0.05, 0.05, 0.99)
	if hit || crit {
		t.Fatal("expected miss at equal level")
	}
	hit, crit, toHit = ResolveAttackRoll(10, 0, 5, 15, 0.05, 0.05, 0.99)
	if !hit || crit || toHit != 15 {
		t.Fatalf("level bonus should turn the miss into a hit, hit=%v crit=%v toHit=%d", hit, crit, toHit)
	}

	// Natural 20 always hits. Full base crit chances crit. A crushed crit chance often does not.
	hit, crit, _ = ResolveAttackRoll(20, 0, -6, 30, 0.05, 0.05, 0.99)
	if !hit || !crit {
		t.Fatalf("natural 20 at base crit should crit, hit=%v crit=%v", hit, crit)
	}
	hit, crit, _ = ResolveAttackRoll(20, 0, -6, 30, 0, 0.05, 0.5)
	if !hit || crit {
		t.Fatalf("natural 20 with zero crit chance should hit and not crit, hit=%v crit=%v", hit, crit)
	}
	hit, crit, _ = ResolveAttackRoll(20, 0, 0, 10, 0.01, 0.05, 0.1)
	if !hit || !crit {
		t.Fatalf("extraRoll 0.1 < 0.01/0.05 should still crit, hit=%v crit=%v", hit, crit)
	}
	hit, crit, _ = ResolveAttackRoll(20, 0, 0, 10, 0.01, 0.05, 0.5)
	if !hit || crit {
		t.Fatalf("extraRoll 0.5 should not crit at reduced chance, hit=%v crit=%v", hit, crit)
	}

	// Extra crit on a normal hit when the attacker is ahead.
	hit, crit, _ = ResolveAttackRoll(15, 0, 0, 10, 0.25, 0.05, 0.1)
	if !hit || !crit {
		t.Fatalf("expected bonus crit, hit=%v crit=%v", hit, crit)
	}
	hit, crit, _ = ResolveAttackRoll(15, 0, 0, 10, 0.25, 0.05, 0.5)
	if !hit || crit {
		t.Fatalf("expected non-crit hit, hit=%v crit=%v", hit, crit)
	}
}

func TestCalculateDamageLevelGap(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	e := NewEngine(nil, DefaultConfig())
	attacker := &entcombat.CombatantRef{AttackPower: 10, Level: 4}
	target := &entcombat.CombatantRef{Defense: 0, Level: 4}

	equal := e.CalculateDamage(attacker, target, false)
	if equal != 10 {
		t.Fatalf("equal-level damage = %d, want 10 (ATK - DEF/2)", equal)
	}
	if crit := e.CalculateDamage(attacker, target, true); crit != 20 {
		t.Fatalf("equal-level crit = %d, want 20", crit)
	}

	attacker.Level = 12
	target.Level = 1
	high := e.CalculateDamage(attacker, target, false)
	if high != balance.ScaleDamage(12, 1, 10) {
		t.Fatalf("high damage %d != ScaleDamage %d", high, balance.ScaleDamage(12, 1, 10))
	}
	if high <= equal {
		t.Fatalf("higher-level attacker should deal more than %d, got %d", equal, high)
	}
	if crit := e.CalculateDamage(attacker, target, true); crit != high*2 {
		t.Fatalf("crit should double gap-scaled damage, got %d want %d", crit, high*2)
	}

	attacker.Level = 1
	target.Level = 12
	low := e.CalculateDamage(attacker, target, false)
	if low != balance.ScaleDamage(1, 12, 10) || low >= equal {
		t.Fatalf("lower-level damage = %d, equal = %d", low, equal)
	}

	// Clamp: 100 vs 1 matches 7 vs 1 (max gap 6).
	if balance.ScaleDamage(100, 1, 10) != balance.ScaleDamage(7, 1, 10) {
		t.Fatal("damage scale did not clamp at +6")
	}
	if balance.ScaleDamage(1, 100, 10) != balance.ScaleDamage(1, 7, 10) {
		t.Fatal("damage scale did not clamp at -6")
	}

	// Defense still applies before the gap multiplier.
	target.Defense = 4
	target.Level = 4
	attacker.Level = 4
	withDef := e.CalculateDamage(attacker, target, false)
	if withDef != 8 { // 10 - 4/2
		t.Fatalf("defense reduction = %d, want 8", withDef)
	}
}

func TestCreateCombatantCopiesLevel(t *testing.T) {
	e := NewEngine(nil, DefaultConfig())
	char := &characters.Character{
		Entity:           entities.NewEntity(),
		Name:             "Hero",
		Level:            7,
		MaxHitPoints:     20,
		CurrentHitPoints: 20,
	}
	pref := e.CreateCombatantFromCharacter(char)
	if pref.Level != 7 {
		t.Fatalf("player combatant level = %d", pref.Level)
	}

	n := &npc.NPC{
		Entity:           entities.NewEntity(),
		Name:             "Rat",
		Level:            4,
		MaxHitPoints:     10,
		CurrentHitPoints: 10,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 3, Defense: 1},
	}
	nref := e.CreateCombatantFromNPC(n)
	if nref.Level != 4 {
		t.Fatalf("npc combatant level = %d", nref.Level)
	}
}

func TestProcessAttackHigherLevelHitsHarder(t *testing.T) {
	e := NewEngine(nil, DefaultConfig())
	const n = 200
	equal := averageAttackDamage(t, e, 6, n)
	high := averageAttackDamage(t, e, 12, n)
	low := averageAttackDamage(t, e, 1, n)
	if high <= equal*1.5 {
		t.Fatalf("level 12 avg damage %.2f should be > 1.5× level 6 avg %.2f", high, equal)
	}
	if low >= equal*0.6 {
		t.Fatalf("level 1 avg damage %.2f should be < 0.6× level 6 avg %.2f", low, equal)
	}
}

func averageAttackDamage(t *testing.T, e *Engine, attackerLevel int32, n int) float64 {
	t.Helper()
	var total int32
	for i := 0; i < n; i++ {
		inst := entcombat.NewCombatInstance("r")
		inst.Players = []entcombat.CombatantRef{{
			ID: "p", Name: "P", Type: entcombat.CombatantTypePlayer, IsAlive: true,
			AttackPower: 20, Level: attackerLevel, MaxHP: 100, CurrentHP: 100,
		}}
		inst.Enemies = []entcombat.CombatantRef{{
			ID: "e", Name: "E", Type: entcombat.CombatantTypeNPC, IsAlive: true,
			Defense: 0, Level: 6, MaxHP: 100000, CurrentHP: 100000,
		}}
		res := e.ProcessAttack(inst, "p", "e")
		total += res.Damage
	}
	return float64(total) / float64(n)
}

func TestSkillDamageFollowsLevelGap(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	prev := skills.AllSkills()
	skills.LoadFromDB(skills.SeedSkills())
	t.Cleanup(func() { skills.LoadFromDB(prev) })
	e := NewEngine(nil, DefaultConfig())

	equal := castPowerStrike(e, 6, 6)
	if !equal.Success || equal.TotalDamage != 3 {
		t.Fatalf("equal-level Power Strike = %+v, want 3 damage", equal)
	}

	highDmg := balance.ScaleDamage(12, 6, 3)
	if highDmg <= 3 {
		t.Fatalf("expected +6 skill scale above 3, got %d", highDmg)
	}
	high := castPowerStrike(e, 12, 6)
	if high.TotalDamage < highDmg {
		t.Fatalf("higher-level skill damage %d < scaled %d", high.TotalDamage, highDmg)
	}

	lowDmg := balance.ScaleDamage(1, 12, 3)
	if lowDmg >= 3 {
		t.Fatalf("expected -6 skill scale below 3, got %d", lowDmg)
	}
	landed := int32(-1)
	misses := 0
	for i := 0; i < 40; i++ {
		res := castPowerStrike(e, 1, 12)
		if res.TotalDamage > 0 {
			landed = res.TotalDamage
			break
		}
		misses++
	}
	if landed < 0 {
		if misses != 40 {
			t.Fatalf("miss accounting %d", misses)
		}
		mods := balance.LevelGapModifiers(1, 12)
		if mods.HitChanceDelta >= 0 {
			t.Fatal("40 misses but hit chance was not reduced")
		}
		return
	}
	if landed != lowDmg {
		t.Fatalf("lower-level skill hit for %d, want scaled %d", landed, lowDmg)
	}
}

func castPowerStrike(e *Engine, casterLevel, targetLevel int32) SkillResult {
	inst := entcombat.NewCombatInstance("r")
	inst.Players = []entcombat.CombatantRef{{
		ID: "p", Name: "P", Type: entcombat.CombatantTypePlayer, IsAlive: true,
		Level: casterLevel, MaxHP: 100, CurrentHP: 100,
		EquippedSkills: []string{"warrior_power_strike"},
		SkillCooldowns: map[string]int{},
	}}
	inst.Enemies = []entcombat.CombatantRef{{
		ID: "e", Name: "E", Type: entcombat.CombatantTypeNPC, IsAlive: true,
		Level: targetLevel, Defense: 0, MaxHP: 500, CurrentHP: 500,
	}}
	return e.ProcessSkill(inst, "p", "warrior_power_strike", "e")
}
