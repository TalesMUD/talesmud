package combat

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

func skillText(got SkillResult) string {
	return strings.Join(got.Messages, " ")
}

func hitUntil(t *testing.T, e *Engine, inst *entcombat.CombatInstance, attackerID, targetID string) AttackResult {
	t.Helper()
	for i := 0; i < 80; i++ {
		res := e.ProcessAttack(inst, attackerID, targetID)
		if res.Hit {
			return res
		}
	}
	t.Fatal("no hit landed")
	return AttackResult{}
}

func boost(inst *entcombat.CombatInstance, id string, power int32) {
	c := inst.GetCombatantByID(id)
	c.AttackPower = power
	c.STRMod = 30
	c.Defense = 0
}

func TestWardGritAndRetaliate(t *testing.T) {
	e := NewEngine(NewManager(), DefaultConfig())
	mk := func(name string) *characters.Character {
		return &characters.Character{
			Entity:           entities.NewEntity(),
			Name:             name,
			Level:            5,
			Class:            characters.ClassWard,
			MaxHitPoints:     800,
			CurrentHitPoints: 800,
		}
	}
	a := mk("A")
	b := mk("B")
	enemy := &npc.NPC{
		Entity:           entities.NewEntity(),
		Name:             "Dummy",
		Level:            5,
		MaxHitPoints:     5000,
		CurrentHitPoints: 5000,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 1, Defense: 0, Difficulty: "normal"},
	}
	inst := e.InitiateCombat("ward", []*characters.Character{a, b}, []*npc.NPC{enemy})
	aID, bID := a.Entity.ID, b.Entity.ID
	boost(inst, aID, 100)
	boost(inst, bID, 100)

	if inst.GetCombatantByID(aID).Grit != balance.WardOpeningGrit || inst.GetCombatantByID(bID).Grit != balance.WardOpeningGrit {
		t.Fatalf("opening grit a=%d b=%d", inst.GetCombatantByID(aID).Grit, inst.GetCombatantByID(bID).Grit)
	}
	first := hitUntil(t, e, inst, aID, bID)
	openBack := balance.WardRetaliateDamage(first.Damage, balance.WardOpeningGrit)
	if openBack < 1 || !strings.Contains(first.Message, "Grit throws") {
		t.Fatalf("opening grit did not throw: back=%d %q", openBack, first.Message)
	}
	if inst.GetCombatantByID(bID).Grit != balance.WardOpeningGrit+1 {
		t.Fatalf("grit after first hit %d", inst.GetCombatantByID(bID).Grit)
	}
	if inst.GetCombatantByID(aID).Grit != balance.WardOpeningGrit || inst.GetCombatantByID(aID).CurrentHP != 800-openBack {
		t.Fatalf("retaliate on the attacker grit %d hp %d", inst.GetCombatantByID(aID).Grit, inst.GetCombatantByID(aID).CurrentHP)
	}

	before := inst.GetCombatantByID(aID).CurrentHP
	second := hitUntil(t, e, inst, aID, bID)
	back := balance.WardRetaliateDamage(second.Damage, balance.WardOpeningGrit+1)
	if back < 1 || !strings.Contains(second.Message, "Grit throws") {
		t.Fatalf("second hit back=%d msg=%q", back, second.Message)
	}
	if inst.GetCombatantByID(aID).CurrentHP != before-back {
		t.Fatalf("throwback hp %d want %d", inst.GetCombatantByID(aID).CurrentHP, before-back)
	}
	if inst.GetCombatantByID(aID).Grit != balance.WardOpeningGrit {
		t.Fatal("retaliate granted grit")
	}
	if inst.GetCombatantByID(bID).Grit != balance.WardOpeningGrit+2 {
		t.Fatalf("grit %d", inst.GetCombatantByID(bID).Grit)
	}

	inst.GetCombatantByID(bID).Grit = 5
	capped := hitUntil(t, e, inst, aID, bID)
	want := balance.WardRetaliateDamage(capped.Damage, 5)
	gotBack := before - back - inst.GetCombatantByID(aID).CurrentHP
	if gotBack != want || want < 1 {
		t.Fatalf("cap throw %d want %d (hit %d)", gotBack, want, capped.Damage)
	}
	if inst.GetCombatantByID(bID).Grit != 5 || inst.GetCombatantByID(aID).Grit != balance.WardOpeningGrit {
		t.Fatalf("cap grit b=%d a=%d", inst.GetCombatantByID(bID).Grit, inst.GetCombatantByID(aID).Grit)
	}

	inst.GetCombatantByID(bID).Grit = 1
	inst.GetCombatantByID(aID).AttackPower = 1
	hp := inst.GetCombatantByID(aID).CurrentHP
	tiny := hitUntil(t, e, inst, aID, bID)
	if tiny.Damage != 1 || balance.WardRetaliateDamage(tiny.Damage, 1) != 0 {
		t.Fatalf("tiny hit %d", tiny.Damage)
	}
	if strings.Contains(tiny.Message, "Grit throws") || inst.GetCombatantByID(aID).CurrentHP != hp {
		t.Fatalf("tiny throw %q hp %d", tiny.Message, inst.GetCombatantByID(aID).CurrentHP)
	}
	if inst.GetCombatantByID(bID).Grit != 2 {
		t.Fatalf("tiny grit %d", inst.GetCombatantByID(bID).Grit)
	}
}

func slamFor(t *testing.T, e *Engine, inst *entcombat.CombatInstance, heroID, enemyID string, grit int, want int32) {
	t.Helper()
	inst.GetCombatantByID(heroID).Grit = grit
	inst.GetCombatantByID(heroID).AttackPower = 200
	for i := 0; i < 40; i++ {
		hero := inst.GetCombatantByID(heroID)
		if hero.SkillCooldowns == nil {
			hero.SkillCooldowns = map[string]int{}
		}
		hero.SkillCooldowns["ward_slam"] = 0
		got := e.ProcessSkill(inst, heroID, "ward_slam", enemyID)
		if !got.Success || got.KeepsSwing {
			t.Fatalf("slam %+v", got)
		}
		if got.HitsLanded == 1 && got.TotalDamage == want {
			if !strings.Contains(skillText(got), "Grit "+itoa(grit)) {
				t.Fatalf("slam text %q", skillText(got))
			}
			again := e.ProcessSkill(inst, heroID, "ward_slam", enemyID)
			if again.Success {
				t.Fatal("slam ignored cooldown")
			}
			return
		}
	}
	t.Fatalf("no slam at grit %d for %d", grit, want)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [4]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func TestWardSlamScalesWithGrit(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassWard, 0)
	equipKit(inst, heroID, "ward_slam")
	slamFor(t, e, inst, heroID, enemyID, 0, 200)
	slamFor(t, e, inst, heroID, enemyID, 2, 280)
	slamFor(t, e, inst, heroID, enemyID, 5, 400)

	stored := characters.Class{ID: "hitch", Name: "Hitch"}
	e, inst, heroID, enemyID = newFight(stored, 0)
	if inst.GetCombatantByID(heroID).ClassID != "ward" {
		t.Fatalf("stored hitch class id %s", inst.GetCombatantByID(heroID).ClassID)
	}
	equipKit(inst, heroID, "ward_slam")
	got := e.ProcessSkill(inst, heroID, "ward_slam", enemyID)
	if !got.Success {
		t.Fatalf("hitch id could not slam %+v", got)
	}
}

func TestWardGuardRedirectAndSelf(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassWard, 0)
	equipKit(inst, heroID, "ward_guard", "ward_slam")
	ally := &characters.Character{
		Entity:           entities.NewEntity(),
		Name:             "Ally",
		Level:            5,
		Class:            characters.ClassRogue,
		MaxHitPoints:     400,
		CurrentHitPoints: 400,
	}
	if !e.JoinCombat(inst, ally) {
		t.Fatal("ally did not join")
	}
	allyID := ally.Entity.ID
	boost(inst, enemyID, 60)

	miss := e.ProcessSkill(inst, heroID, "ward_guard", "nobody")
	if miss.Success || !strings.Contains(skillText(miss), "Nobody to guard.") {
		t.Fatalf("unknown guard %+v", miss)
	}

	guard := e.ProcessSkill(inst, heroID, "ward_guard", "")
	if !guard.Success || !guard.KeepsSwing || !strings.Contains(skillText(guard), "You guard Ally") {
		t.Fatalf("guard %+v", guard)
	}
	if again := e.ProcessSkill(inst, heroID, "ward_guard", allyID); again.Success {
		t.Fatal("guard twice")
	}
	hero := inst.GetCombatantByID(heroID)
	if hero.GuardCharges != 1 || hero.GuardRounds != 2 || hero.GuardSelf || hero.GuardTargetID != allyID {
		t.Fatalf("guard state %+v", hero)
	}

	direct := hitUntil(t, e, inst, enemyID, heroID)
	hero = inst.GetCombatantByID(heroID)
	if hero.GuardCharges != 1 || hero.Grit != balance.WardOpeningGrit+1 {
		t.Fatalf("direct hit spent the ally guard: charges %d grit %d msg %q", hero.GuardCharges, hero.Grit, direct.Message)
	}
	allyHP := inst.GetCombatantByID(allyID).CurrentHP
	wardHP := hero.CurrentHP
	redirected := hitUntil(t, e, inst, enemyID, allyID)
	if inst.GetCombatantByID(allyID).CurrentHP != allyHP {
		t.Fatalf("ally took the hit %d -> %d (%q)", allyHP, inst.GetCombatantByID(allyID).CurrentHP, redirected.Message)
	}
	hero = inst.GetCombatantByID(heroID)
	if hero.CurrentHP >= wardHP || hero.GuardCharges != 0 || hero.Grit != balance.WardOpeningGrit+2 {
		t.Fatalf("redirect ward hp %d charges %d grit %d", hero.CurrentHP, hero.GuardCharges, hero.Grit)
	}
}

func TestWardSelfGuardAndMiss(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassWard, 0)
	equipKit(inst, heroID, "ward_guard")
	boost(inst, enemyID, 50)
	got := e.ProcessSkill(inst, heroID, "ward_guard", "self")
	if !got.Success || !got.KeepsSwing || !strings.Contains(skillText(got), "two Grit") {
		t.Fatalf("self guard %+v", got)
	}
	hero := inst.GetCombatantByID(heroID)
	hero.SmokeMiss = true
	smoked := e.ProcessAttack(inst, enemyID, heroID)
	if smoked.Hit {
		t.Fatalf("smoke should miss: %q", smoked.Message)
	}
	hero = inst.GetCombatantByID(heroID)
	if hero.GuardCharges != 1 || !hero.GuardSelf || hero.Grit != balance.WardOpeningGrit {
		t.Fatalf("miss spent self guard %+v", hero)
	}
	hitUntil(t, e, inst, enemyID, heroID)
	hero = inst.GetCombatantByID(heroID)
	if hero.Grit != balance.WardOpeningGrit+2 || hero.GuardCharges != 0 || hero.GuardSelf {
		t.Fatalf("self soak grit %d charges %d self %v", hero.Grit, hero.GuardCharges, hero.GuardSelf)
	}
	hitUntil(t, e, inst, enemyID, heroID)
	if inst.GetCombatantByID(heroID).Grit != balance.WardOpeningGrit+3 {
		t.Fatalf("next hit grit %d", inst.GetCombatantByID(heroID).Grit)
	}
}

func TestWardGuardWindowExpires(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassWard, 0)
	equipKit(inst, heroID, "ward_guard")
	if got := e.ProcessSkill(inst, heroID, "ward_guard", "me"); !got.Success {
		t.Fatalf("self %+v", got)
	}
	hero := inst.GetCombatantByID(heroID)
	e.tickKitRounds(inst, hero)
	hero = inst.GetCombatantByID(heroID)
	if hero.GuardRounds != 1 || hero.GuardCharges != 1 {
		t.Fatalf("after one round %+v", hero)
	}
	e.tickKitRounds(inst, hero)
	hero = inst.GetCombatantByID(heroID)
	if hero.GuardRounds != 0 || hero.GuardCharges != 0 || hero.GuardSelf {
		t.Fatalf("window left %+v", hero)
	}
	boost(inst, enemyID, 40)
	hitUntil(t, e, inst, enemyID, heroID)
	if inst.GetCombatantByID(heroID).Grit != balance.WardOpeningGrit+1 {
		t.Fatalf("expired self guard still doubled grit: %d", inst.GetCombatantByID(heroID).Grit)
	}
}

func TestWardLevel1BasicSwingClearsRatBreakpoint(t *testing.T) {
	hitFor := func(level int32, class characters.Class, power int32) AttackResult {
		t.Helper()
		e := NewEngine(NewManager(), DefaultConfig())
		hero := &characters.Character{
			Entity:           entities.NewEntity(),
			Name:             "Hero",
			Level:            level,
			Class:            class,
			MaxHitPoints:     40,
			CurrentHitPoints: 40,
		}
		enemy := &npc.NPC{
			Entity:           entities.NewEntity(),
			Name:             "Catacomb Rat",
			Level:            level,
			MaxHitPoints:     20,
			CurrentHitPoints: 20,
			EnemyTrait:       &npc.EnemyTrait{AttackPower: 3, Defense: 0, Difficulty: "trivial"},
		}
		inst := e.InitiateCombat("R0005", []*characters.Character{hero}, []*npc.NPC{enemy})
		inst.GetCombatantByID(hero.Entity.ID).AttackPower = power
		for i := 0; i < 80; i++ {
			res := e.ProcessAttack(inst, hero.Entity.ID, enemy.Entity.ID)
			if res.Hit && !res.Critical {
				return res
			}
			if !inst.GetCombatantByID(enemy.Entity.ID).IsAlive {
				inst.GetCombatantByID(enemy.Entity.ID).CurrentHP = 20
				inst.GetCombatantByID(enemy.Entity.ID).IsAlive = true
			}
		}
		t.Fatal("no non-crit hit")
		return AttackResult{}
	}
	// Sword 5 + STR 12 is 6. 0.95 rounds to 6, and the level-1 chip makes 7.
	// Three of those clear a 20 HP rat. Fenwatch starter (7) is not beaten.
	if got := hitFor(1, characters.ClassWard, 6); got.Damage != 7 {
		t.Fatalf("level-1 ward swing %d", got.Damage)
	}
	if got := hitFor(5, characters.ClassWard, 6); got.Damage != 6 {
		t.Fatalf("later ward swing %d, starter chip leaked", got.Damage)
	}
	if got := hitFor(1, characters.ClassWarrior, 7); got.Damage != 7 {
		t.Fatalf("fenwatch starter %d", got.Damage)
	}
}
