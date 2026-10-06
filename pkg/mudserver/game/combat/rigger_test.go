package combat

import (
	"math"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
)

func scrapClass(t *testing.T) characters.Class {
	t.Helper()
	c, ok := characters.ClassByID("rigger")
	if !ok {
		t.Fatal("scrap class missing from catalog")
	}
	return c
}

func logText(inst *entcombat.CombatInstance) string {
	var b strings.Builder
	for _, entry := range inst.Log {
		b.WriteString(entry.Message)
		b.WriteByte('\n')
	}
	return b.String()
}

func TestBoltReturnsDamageAndBreaks(t *testing.T) {
	e, inst, heroID, enemyID := newFight(scrapClass(t), 0)
	hero := inst.GetCombatantByID(heroID)
	if hero.BoltLeft != 1 || hero.RigLeft != 1 {
		t.Fatalf("charges bolt=%d rig=%d", hero.BoltLeft, hero.RigLeft)
	}
	enemyBefore := inst.GetCombatantByID(enemyID).CurrentHP
	heroBefore := hero.CurrentHP
	msg := e.ProcessBolt(inst, heroID, enemyID)
	if !strings.Contains(msg, "bolt scrap") {
		t.Fatalf("arm message %q", msg)
	}
	hero = inst.GetCombatantByID(heroID)
	enemy := inst.GetCombatantByID(enemyID)
	if hero.BoltLeft != 0 || !enemy.ScrapArmed {
		t.Fatalf("armed bolt=%d scrap=%v", hero.BoltLeft, enemy.ScrapArmed)
	}
	if hero.CurrentHP != heroBefore || enemy.CurrentHP != enemyBefore {
		t.Fatal("arming changed hp")
	}
	if again := e.ProcessBolt(inst, heroID, enemyID); again != "" {
		t.Fatalf("second bolt did something: %q", again)
	}
	hero = inst.GetCombatantByID(heroID)
	enemy = inst.GetCombatantByID(enemyID)
	if hero.BoltLeft != 0 || !enemy.ScrapArmed {
		t.Fatalf("second bolt changed charges bolt=%d scrap=%v", hero.BoltLeft, enemy.ScrapArmed)
	}

	var result AttackResult
	landed := false
	for i := 0; i < 80; i++ {
		result = e.ProcessAttack(inst, heroID, enemyID)
		if result.Hit {
			landed = true
			break
		}
	}
	if !landed {
		t.Fatal("could not land a hit on the scrapped target")
	}
	hero = inst.GetCombatantByID(heroID)
	enemy = inst.GetCombatantByID(enemyID)
	if enemyBefore-enemy.CurrentHP != result.Damage {
		t.Fatalf("target lost %d, hit was %d", enemyBefore-enemy.CurrentHP, result.Damage)
	}
	if heroBefore-hero.CurrentHP != result.Damage {
		t.Fatalf("attacker lost %d, hit was %d", heroBefore-hero.CurrentHP, result.Damage)
	}
	if enemy.ScrapArmed {
		t.Fatal("scrap still armed")
	}
	text := logText(inst)
	if !strings.Contains(text, "Scrap returns") || !strings.Contains(text, "breaks") {
		t.Fatalf("return line missing: %s", text)
	}
	if !strings.Contains(strings.ToLower(text), "hero") {
		t.Fatalf("hit log missing: %s", text)
	}
	if strings.Contains(strings.ToLower(text), "absorb") {
		t.Fatalf("absorb line: %s", text)
	}
	for _, se := range hero.StatusEffects {
		if strings.Contains(strings.ToLower(se.Stat+" "+se.Name), "shield") {
			t.Fatalf("shield status %+v", se)
		}
	}
	if e.ProcessBolt(inst, heroID, enemyID) != "" || inst.GetCombatantByID(enemyID).ScrapArmed {
		t.Fatal("bolt after the return armed another scrap")
	}
}

func TestRigTwoRoundsThenDespawns(t *testing.T) {
	e, inst, heroID, enemyID := newFight(scrapClass(t), 0)
	inst.OriginRoomID = "gear-yard"
	hero := inst.GetCombatantByID(heroID)
	hero.AttackPower = 40
	hero.WeaponSubType = "sword"
	hero.RaceID = "construct"
	enemy := inst.GetCombatantByID(enemyID)
	enemy.Defense = 0
	swing := e.CalculateDamage(hero, enemy, false)
	want := int32(math.Round(float64(swing) * 0.50))
	if want < 1 {
		want = 1
	}
	if swing == want {
		t.Fatalf("half damage collapsed swing=%d", swing)
	}
	before := enemy.CurrentHP
	msg := e.ProcessRig(inst, heroID)
	if !strings.Contains(msg, "You drop a rig") || !strings.Contains(msg, "The rig hits") {
		t.Fatalf("drop/hit %q", msg)
	}
	if e.ProcessRig(inst, heroID) != "" {
		t.Fatal("second rig spawned")
	}
	enemy = inst.GetCombatantByID(enemyID)
	if before-enemy.CurrentHP != want {
		t.Fatalf("first rig hit %d, want %d (swing %d)", before-enemy.CurrentHP, want, swing)
	}
	if inst.Rig == nil || inst.Rig.RoundsLeft != 1 || inst.Rig.RoomID != "gear-yard" || inst.Rig.Follows {
		t.Fatalf("turret %+v", inst.Rig)
	}
	rigID := inst.Rig.ID
	for _, p := range inst.Players {
		if p.ID == rigID {
			t.Fatal("rig is a player")
		}
	}
	for _, en := range inst.Enemies {
		if en.ID == rigID {
			t.Fatal("rig is an enemy")
		}
	}
	for _, turn := range inst.TurnOrder {
		if turn.ID == rigID {
			t.Fatal("rig is in turn order")
		}
	}
	inst.Rig.Follows = true
	inst.OriginRoomID = "somewhere-else"
	e.ProcessRoundStart(inst)
	enemy = inst.GetCombatantByID(enemyID)
	if before-enemy.CurrentHP != want*2 {
		t.Fatalf("two rig hits lost %d, want %d", before-enemy.CurrentHP, want*2)
	}
	if inst.Rig != nil {
		t.Fatalf("turret still up %+v", inst.Rig)
	}
	text := logText(inst)
	if strings.Count(text, "The rig hits") != 2 {
		t.Fatalf("hit lines %d\n%s", strings.Count(text, "The rig hits"), text)
	}
	if !strings.Contains(text, "The rig falls apart.") {
		t.Fatalf("despawn line missing\n%s", text)
	}
	if strings.Contains(text, "somewhere-else") {
		t.Fatal("rig followed a room change")
	}
	held := enemy.CurrentHP
	e.ProcessRoundStart(inst)
	if e.ProcessRig(inst, heroID) != "" || inst.Rig != nil {
		t.Fatal("rig came back")
	}
	if inst.GetCombatantByID(enemyID).CurrentHP != held {
		t.Fatal("rig hit after despawn")
	}
	if inst.GetCombatantByID(heroID).RigLeft != 0 {
		t.Fatal("rig charge returned")
	}
}

func TestConstructPoisonNeverApplies(t *testing.T) {
	e, inst, heroID, enemyID := newFight(scrapClass(t), 0)
	enemy := inst.GetCombatantByID(enemyID)
	enemy.RaceID = "construct"
	e.ApplyStatusEffectFromScript(inst, enemyID, entcombat.StatusEffect{
		SkillID:  "rogue_poison_strike",
		Name:     "Poison",
		Type:     "dot",
		Stat:     "poison",
		Duration: 3,
		Value:    4,
	})
	enemy = inst.GetCombatantByID(enemyID)
	if len(enemy.StatusEffects) != 0 {
		t.Fatalf("poison stuck %+v", enemy.StatusEffects)
	}
	e.ApplyStatusEffectFromScript(inst, enemyID, entcombat.StatusEffect{
		SkillID:  "inscribe",
		Name:     "Inscribe",
		Type:     "dot",
		Stat:     "hp",
		Duration: 3,
		Value:    4,
	})
	enemy = inst.GetCombatantByID(enemyID)
	if len(enemy.StatusEffects) != 1 || enemy.StatusEffects[0].SkillID != "inscribe" {
		t.Fatalf("inscribe %+v", enemy.StatusEffects)
	}

	e2, inst2, _, enemy2 := newFight(characters.ClassRogue, 0)
	inst2.GetCombatantByID(enemy2).RaceID = "elf"
	e2.ApplyStatusEffectFromScript(inst2, enemy2, entcombat.StatusEffect{
		SkillID:  "rogue_poison_strike",
		Name:     "Poison",
		Type:     "dot",
		Duration: 3,
		Value:    4,
	})
	if len(inst2.GetCombatantByID(enemy2).StatusEffects) != 1 {
		t.Fatal("elf should take poison")
	}
	_ = heroID
}

func TestRacialBonusInDamagePath(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassWarrior, 0)
	hero := inst.GetCombatantByID(heroID)
	enemy := inst.GetCombatantByID(enemyID)
	hero.ClassID = "warrior"
	hero.Level = 5
	hero.AttackPower = 20
	enemy.Defense = 0
	enemy.Level = 5
	enemy.ClassID = ""

	hero.RaceID = "dwarf"
	hero.WeaponSubType = "mace"
	if got := e.CalculateDamage(hero, enemy, false); got != 22 {
		t.Fatalf("dwarf mace %d", got)
	}
	hero.WeaponSubType = "sword"
	if got := e.CalculateDamage(hero, enemy, false); got != 20 {
		t.Fatalf("dwarf sword %d", got)
	}
	hero.RaceID = "human"
	hero.WeaponSubType = "mace"
	if got := e.CalculateDamage(hero, enemy, false); got != 20 {
		t.Fatalf("human mace %d", got)
	}
	hero.RaceID = "elf"
	hero.WeaponSubType = "bow"
	if got := e.CalculateDamage(hero, enemy, false); got != 22 {
		t.Fatalf("elf bow %d", got)
	}
	hero.RaceID = "elve"
	if got := e.CalculateDamage(hero, enemy, false); got != 22 {
		t.Fatalf("elve bow %d", got)
	}
	hero.RaceID = "construct"
	hero.WeaponSubType = "mace"
	if got := e.CalculateDamage(hero, enemy, false); got != 20 {
		t.Fatalf("construct mace %d", got)
	}
	hero.RaceID = "orc"
	if got := e.CalculateDamage(hero, enemy, false); got != 20 {
		t.Fatalf("orc mace %d", got)
	}
}
