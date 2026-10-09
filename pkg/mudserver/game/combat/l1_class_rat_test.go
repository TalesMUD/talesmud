package combat_test

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	combatentity "github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	combatengine "github.com/talesmud/talesmud/pkg/mudserver/game/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/combat/simutil"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
)

// TestL1ClassRatNest fights fresh level-1 create-screen characters against
// Catacomb Rats through the combat engine. Gear is the template's starting
// items, equipped the way guest creation equips them. The signed web create
// path copies HP and attributes and does not equip those items.
//
// rand.Seed below is live because gap_matrix_test.go sets
// //go:debug randseednop=0 for this test binary. Go rejects a second copy.
//
// Live attack pulls every other living enemy in the room only when the target's
// combat style is swarm. CallForHelp alone does not. Catacomb Rats are swarm,
// and the nest keeps three, so attacking one rat is a three-rat fight. That
// swarm is the primary row. One rat is the secondary row.
//
// The kit policy uses each level-1 button except Slip. Slip sets the fight to
// fled, so an Alley who is trying to kill the rat swings twice instead.
// If the player queues nothing, live combat waits DecisionWindowSeconds (5)
// and then doAutoAttack, which is a basic swing. The auto row is that path.
func TestL1ClassRatNest(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	prev := skills.AllSkills()
	skills.LoadFromDB(skills.SeedSkills())
	t.Cleanup(func() { skills.LoadFromDB(prev) })

	iters := l1Iterations(t)
	presets := characters.SystemCharacterTemplatePresets()
	if len(presets) < 5 {
		t.Fatalf("expected the five pack classes, got %d", len(presets))
	}

	rat := catacombRat()
	t.Logf("catacomb rat hp %d atk %d def %d style %s call %v",
		rat.MaxHitPoints, rat.EnemyTrait.AttackPower, rat.EnemyTrait.Defense,
		rat.EnemyTrait.CombatStyle, rat.EnemyTrait.CallForHelp)

	var single, swarm, auto []l1Summary
	for _, preset := range presets {
		sample := freshLevel1(preset)
		weapon := sample.GetWeaponDamage()
		mod := int32(sample.GetPrimaryAttackMod())
		sub := ""
		if w := sample.EquippedItems[items.ItemSlotMainHand]; w != nil {
			sub = string(w.SubType)
		}
		t.Logf("create %s id %s race %s hp %d weapon %d mod %d atk %d def %d str %d dex %d sta %d subtype %q skills %v",
			preset.Name, sample.Class.ID, sample.Race.ID, sample.MaxHitPoints,
			weapon, mod, weapon+mod, sample.GetArmorDefense(),
			sample.GetAttribute("STR"), sample.GetAttribute("DEX"), sample.GetAttribute("STA"),
			sub, sample.EquippedSkills)
		if sample.Class.ID == "ward" {
			// Live probe: dwarf Ward, sword equipped, HP 26, attack 6 (5+1), STR 12.
			// Rusty Sword is not blunt, so the dwarf bonus does not change that swing.
			if sample.MaxHitPoints != 26 || weapon != 5 || mod != 1 || sample.GetAttribute("STR") != 12 {
				t.Errorf("ward opener hp %d weapon %d mod %d str %d", sample.MaxHitPoints, weapon, mod, sample.GetAttribute("STR"))
			}
			if balance.RacialWeaponMultiplier("dwarf", sub) != 1 {
				t.Errorf("dwarf subtype %q multiplier %.2f", sub, balance.RacialWeaponMultiplier("dwarf", sub))
			}
		}

		three := runScenario(preset, iters, 3, true, false)
		one := runScenario(preset, iters, 1, true, false)
		plain := runScenario(preset, iters, 3, sample.Class.ID == "ward", true)
		swarm = append(swarm, three)
		single = append(single, one)
		auto = append(auto, plain)
		t.Logf("SWARM %s", three.line())
		t.Logf("SINGLE %s", one.line())
		t.Logf("AUTO %s", plain.line())
		// Signed-in web create does not equip starting items. This row is that
		// body: no main hand and no chest. It is context. The tune is the geared row.
		naked := freshLevel1(preset)
		stripWornGear(naked)
		t.Logf("bare create %s weapon %d mod %d atk %d def %d",
			preset.Name, naked.GetWeaponDamage(), naked.GetPrimaryAttackMod(),
			naked.GetWeaponDamage()+int32(naked.GetPrimaryAttackMod()), naked.GetArmorDefense())
		bare := runBareNest(preset, iters)
		t.Logf("BARE %s", bare.line())
		if sample.Class.ID == "ward" {
			t.Logf("WARD swarm split %s", three.split())
			t.Logf("WARD single split %s", one.split())
			t.Logf("WARD auto split %s", plain.split())
		}
	}

	// Longer fights, same create path then the real level-up, so the opener
	// can be compared with a fight that has time to stack Grit.
	for _, id := range []string{"warrior", "ward"} {
		preset := presetByClass(presets, id)
		if preset == nil {
			t.Fatalf("missing %s", id)
		}
		boss := runLeveled(preset, iters, 10, "boss")
		bear := runLeveled(preset, iters, 5, "bear")
		t.Logf("PEAK L10 boss %s", boss.line())
		t.Logf("PEAK L5 bear %s", bear.line())
		if id == "ward" {
			t.Logf("WARD L10 split %s", boss.split())
			t.Logf("WARD L5 split %s", bear.split())
		}
	}

	fenSwarm := summaryByID(swarm, "warrior")
	wardSwarm := summaryByID(swarm, "ward")
	wardSingle := summaryByID(single, "ward")
	if fenSwarm == nil || wardSwarm == nil || wardSingle == nil {
		t.Fatal("missing ward or warrior summary")
	}
	// Primary: the nest. Sensible kit play wins, stays near the plate class's
	// round count, and does not finish near dead.
	if wardSwarm.winRate < 0.95 {
		t.Errorf("ward swarm win %.1f%%", wardSwarm.winRate*100)
	}
	if fenSwarm.meanRounds > 0 {
		ratio := wardSwarm.meanRounds / fenSwarm.meanRounds
		if ratio < 0.85 || ratio > 1.20 {
			t.Errorf("ward/plate swarm rounds %.2f (ward %.2f plate %.2f)", ratio, wardSwarm.meanRounds, fenSwarm.meanRounds)
		}
	}
	if wardSwarm.hpLeft < 50 {
		t.Errorf("ward swarm hp left %.1f%%", wardSwarm.hpLeft)
	}
	// Secondary: one rat is still a clear. Round count is logged, not locked.
	if wardSingle.winRate < 0.95 {
		t.Errorf("ward single win %.1f%%", wardSingle.winRate*100)
	}
}

func l1Iterations(t *testing.T) int {
	t.Helper()
	if v := os.Getenv("L1_ITERS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			t.Fatalf("L1_ITERS %q", v)
		}
		return n
	}
	return 2000
}

func presetByClass(presets []*characters.CharacterTemplate, id string) *characters.CharacterTemplate {
	for _, p := range presets {
		if p != nil && p.Class.ID == id {
			return p
		}
	}
	return nil
}

func summaryByID(rows []l1Summary, id string) *l1Summary {
	for i := range rows {
		if rows[i].classID == id {
			return &rows[i]
		}
	}
	return nil
}

func catacombRat() *npc.NPC {
	n := simutil.CreateEnemy(simutil.EnemyConfig{
		Name: "Catacomb Rat", Level: 1, HP: 8, AttackPower: 1, Defense: 0, Difficulty: "trivial",
	})
	if n.EnemyTrait != nil {
		n.EnemyTrait.CombatStyle = npc.CombatStyleSwarm
		n.EnemyTrait.CallForHelp = true
	}
	return n
}

func freshLevel1(preset *characters.CharacterTemplate) *characters.Character {
	ch := &characters.Character{
		Entity:           entities.NewEntity(),
		Name:             preset.Name,
		Race:             preset.Race,
		Class:            preset.Class,
		Level:            preset.Level,
		CurrentHitPoints: preset.CurrentHitPoints,
		MaxHitPoints:     preset.MaxHitPoints,
		CurrentMana:      preset.CurrentMana,
		MaxMana:          preset.MaxMana,
		Attributes:       append(characters.Attributes(nil), preset.Attributes...),
	}
	if ch.Level < 1 {
		ch.Level = 1
	}
	ch.EquippedSkills = skills.FillHotbar(ch.Class.ID, ch.Level, append([]string(nil), preset.DefaultSkills...))
	if len(preset.StartingItems) > 0 {
		ch.EquippedItems = map[items.ItemSlot]*items.Item{}
		for _, si := range preset.StartingItems {
			tpl := items.StarterItemTemplateByName(si.ItemTemplateName)
			if tpl == nil || si.Slot == "" {
				continue
			}
			copyItem := *tpl
			copyItem.Entity = entities.NewEntity()
			copyItem.IsTemplate = false
			ch.EquippedItems[si.Slot] = &copyItem
		}
	}
	return ch
}

// stripWornGear drops the starter sword and chest. Unarmed damage is 1 and
// armor defense is 0, which is what a signed-in web create fights with.
func stripWornGear(ch *characters.Character) {
	if ch == nil || ch.EquippedItems == nil {
		return
	}
	delete(ch.EquippedItems, items.ItemSlotMainHand)
	delete(ch.EquippedItems, items.ItemSlotChest)
}

func runBareNest(preset *characters.CharacterTemplate, iters int) l1Summary {
	return runFights(preset.Class.ID, preset.Name+" bare", iters, false, false, func() (*characters.Character, []*npc.NPC) {
		ch := freshLevel1(preset)
		stripWornGear(ch)
		enemies := make([]*npc.NPC, 3)
		for i := range enemies {
			enemies[i] = catacombRat()
		}
		return ch, enemies
	})
}

// leveledFromCreate starts from the level-1 template, then applies the real
// level-up path and spends points on the class primary. Weapon and armor are
// the starter kit plus the gap-table appropriate bump.
func leveledFromCreate(preset *characters.CharacterTemplate, level int32) *characters.Character {
	ch := freshLevel1(preset)
	if level > 1 {
		leveling.ApplyLevelUp(ch, int(level-1))
	}
	autoSpendPrimary(ch)
	ch.CurrentHitPoints = ch.MaxHitPoints
	ch.CurrentMana = ch.MaxMana
	base := freshLevel1(preset)
	weapon := int(base.GetWeaponDamage()) + int(level)/2
	armor := int(base.GetArmorDefense()) + int(level)/3
	if weapon < 1 {
		weapon = 1
	}
	ch.EquippedItems[items.ItemSlotMainHand] = &items.Item{
		Entity: entities.NewEntity(),
		Name:   "Level Weapon",
		Type:   items.ItemTypeWeapon,
		Slot:   items.ItemSlotMainHand,
		Attributes: map[string]interface{}{
			"damage": weapon,
		},
	}
	ch.EquippedItems[items.ItemSlotChest] = &items.Item{
		Entity: entities.NewEntity(),
		Name:   "Level Armor",
		Type:   items.ItemTypeArmor,
		Slot:   items.ItemSlotChest,
		Attributes: map[string]interface{}{
			"armor": armor,
		},
	}
	return ch
}

func autoSpendPrimary(ch *characters.Character) {
	primary := strings.ToUpper(strings.TrimSpace(ch.Class.ID))
	if p := strings.ToUpper(strings.TrimSpace(leveling.GetPrimaryAttribute(ch.Class))); p != "" {
		primary = p
	}
	// Catalog primary wins when the class is one of the signed five.
	switch strings.ToLower(ch.Class.ID) {
	case "warrior", "ward", "rigger":
		primary = "STR"
	case "rogue":
		primary = "DEX"
	case "wizard":
		primary = "INT"
	}
	order := []string{primary, leveling.GetSecondaryAttribute(ch.Class), "STA", "STR", "DEX", "INT", "WIS"}
	seen := map[string]bool{}
	for _, attr := range order {
		if attr == "" || seen[attr] {
			continue
		}
		seen[attr] = true
		for ch.UnspentAttributePoints > 0 {
			if msg := leveling.ValidateAttributeSpend(ch, attr, 1); msg != "" {
				break
			}
			leveling.ApplyAttributeSpend(ch, attr, 1)
		}
	}
}

type l1Summary struct {
	label      string
	classID    string
	fights     int
	wins       int
	losses     int
	timeouts   int
	winRate    float64
	meanRounds float64
	medRounds  float64
	lossMean   float64
	lossMed    float64
	hpLeft     float64
	swing      float64
	slam       float64
	retaliate  float64
	retTotal   float64
	grit       [7]float64
	gritN      [7]int
	maxHP      int32
}

func (s l1Summary) line() string {
	msg := fmt.Sprintf("%s wins %.1f%% (%d/%d, loss %d timeout %d) rounds mean %.2f med %.1f hp-left %.1f%% maxhp %d",
		s.label, s.winRate*100, s.wins, s.fights, s.losses, s.timeouts, s.meanRounds, s.medRounds, s.hpLeft, s.maxHP)
	if s.losses > 0 {
		msg += fmt.Sprintf(" defeat-rounds mean %.2f med %.1f", s.lossMean, s.lossMed)
	}
	return msg
}

func (s l1Summary) split() string {
	return fmt.Sprintf("per-round swing %.2f slam %.2f retaliate %.2f; retaliate total/fight %.2f; grit r1-6 %.2f %.2f %.2f %.2f %.2f %.2f (n %d %d %d %d %d %d)",
		s.swing, s.slam, s.retaliate, s.retTotal,
		s.grit[1], s.grit[2], s.grit[3], s.grit[4], s.grit[5], s.grit[6],
		s.gritN[1], s.gritN[2], s.gritN[3], s.gritN[4], s.gritN[5], s.gritN[6])
}

func runScenario(preset *characters.CharacterTemplate, iters, rats int, track, auto bool) l1Summary {
	label := preset.Name
	if auto {
		label += " auto"
	}
	return runFights(preset.Class.ID, label, iters, track, auto, func() (*characters.Character, []*npc.NPC) {
		enemies := make([]*npc.NPC, rats)
		for i := range enemies {
			enemies[i] = catacombRat()
		}
		return freshLevel1(preset), enemies
	})
}

func runLeveled(preset *characters.CharacterTemplate, iters int, level int32, kind string) l1Summary {
	label := fmt.Sprintf("%s L%d %s", preset.Name, level, kind)
	return runFights(preset.Class.ID, label, iters, preset.Class.ID == "ward", false, func() (*characters.Character, []*npc.NPC) {
		ch := leveledFromCreate(preset, level)
		var enemy *npc.NPC
		switch kind {
		case "boss":
			enemy = simutil.CreateScaledEnemy("boss", level, "boss")
		default:
			enemy = simutil.CreateEnemy(simutil.EnemyConfig{
				Name: "Thornback Bear", Level: 5, HP: 80, AttackPower: 11, Defense: 4, Difficulty: "hard",
			})
		}
		return ch, []*npc.NPC{enemy}
	})
}

func runFights(classID, name string, iters int, track, auto bool, spawn func() (*characters.Character, []*npc.NPC)) l1Summary {
	// One seed per table cell, independent of which cell ran first.
	rand.Seed(20261006)
	var wins, losses, timeouts int
	var winRounds, lossRounds []int
	var hpSum float64
	var swing, slam, retaliate int64
	var roundSum int
	var gritSum [7]int
	var gritN [7]int
	var maxHP int32
	for i := 0; i < iters; i++ {
		ch, enemies := spawn()
		if i == 0 {
			maxHP = ch.MaxHitPoints
		}
		one := fightOnce(ch, enemies, track, auto)
		roundSum += one.rounds
		swing += one.swing
		slam += one.slam
		retaliate += one.retaliate
		for r := 1; r <= 6; r++ {
			if one.gritSeen[r] {
				gritSum[r] += one.grit[r]
				gritN[r]++
			}
		}
		switch one.state {
		case combatentity.CombatStateVictory:
			wins++
			winRounds = append(winRounds, one.rounds)
			if one.hpMax > 0 {
				hpSum += float64(one.hpEnd) / float64(one.hpMax) * 100
			}
		case combatentity.CombatStateDefeat:
			losses++
			lossRounds = append(lossRounds, one.rounds)
		default:
			timeouts++
		}
	}
	sum := l1Summary{
		label:    name,
		classID:  classID,
		fights:   iters,
		wins:     wins,
		losses:   losses,
		timeouts: timeouts,
		winRate:  float64(wins) / float64(iters),
		maxHP:    maxHP,
	}
	if wins > 0 {
		var total float64
		for _, r := range winRounds {
			total += float64(r)
		}
		sum.meanRounds = total / float64(wins)
		sum.medRounds = medianInts(winRounds)
		sum.hpLeft = hpSum / float64(wins)
	}
	if len(lossRounds) > 0 {
		var total float64
		for _, r := range lossRounds {
			total += float64(r)
		}
		sum.lossMean = total / float64(len(lossRounds))
		sum.lossMed = medianInts(lossRounds)
	}
	if roundSum > 0 {
		sum.swing = float64(swing) / float64(roundSum)
		sum.slam = float64(slam) / float64(roundSum)
		sum.retaliate = float64(retaliate) / float64(roundSum)
	}
	if iters > 0 {
		sum.retTotal = float64(retaliate) / float64(iters)
	}
	for r := 1; r <= 6; r++ {
		sum.gritN[r] = gritN[r]
		if gritN[r] > 0 {
			sum.grit[r] = float64(gritSum[r]) / float64(gritN[r])
		}
	}
	return sum
}

type fightOnceResult struct {
	state     combatentity.CombatState
	rounds    int
	hpMax     int32
	hpEnd     int32
	swing     int64
	slam      int64
	retaliate int64
	grit      [7]int
	gritSeen  [7]bool
}

func fightOnce(ch *characters.Character, enemies []*npc.NPC, track, auto bool) fightOnceResult {
	engine := combatengine.NewEngine(combatengine.NewManager(), nil)
	inst := engine.InitiateCombat("R0005", []*characters.Character{ch}, enemies)
	var hpMax int32
	for _, p := range inst.Players {
		hpMax += p.MaxHP
	}
	seen := [7]bool{}
	grit := [7]int{}
	const maxTurns = 300
	state := combatentity.CombatStateTimeout
	for turn := 0; turn < maxTurns; turn++ {
		current := inst.GetCurrentTurnCombatant()
		if current == nil {
			// NextTurn opens the next round when the tail is already dead, so
			// a nil current means no living combatant is left in the order.
			state = engine.CheckCombatEnd(inst)
			if state != combatentity.CombatStateActive {
				return finishFight(inst, ch.ID, state, hpMax, track, grit, seen)
			}
			if engine.NextTurn(inst) == nil {
				break
			}
			continue
		}
		if current.Type == combatentity.CombatantTypeNPC {
			action, targetID := engine.GetNPCAIAction(inst, current, nil)
			switch action {
			case combatentity.CombatActionAttack:
				engine.StepNPCAttack(inst, current.ID, targetID)
			case combatentity.CombatActionDefend:
				engine.ProcessDefend(inst, current.ID)
			}
		} else if player := inst.GetPlayerByID(current.ID); player != nil {
			if track && inst.Round >= 1 && inst.Round <= 6 && !seen[inst.Round] {
				seen[inst.Round] = true
				grit[inst.Round] = player.Grit
			}
			living := inst.GetLivingEnemies()
			if len(living) == 0 {
				state = engine.CheckCombatEnd(inst)
				if state == combatentity.CombatStateActive {
					state = combatentity.CombatStateTimeout
				}
				return finishFight(inst, ch.ID, state, hpMax, track, grit, seen)
			}
			if auto {
				engine.ProcessAttack(inst, player.ID, living[0].ID)
			} else {
				playerAct(engine, inst, player, living[0].ID)
			}
		}
		state = engine.CheckCombatEnd(inst)
		if state != combatentity.CombatStateActive {
			return finishFight(inst, ch.ID, state, hpMax, track, grit, seen)
		}
		engine.NextTurn(inst)
	}
	if state == combatentity.CombatStateActive {
		state = combatentity.CombatStateTimeout
	}
	return finishFight(inst, ch.ID, state, hpMax, track, grit, seen)
}

func finishFight(inst *combatentity.CombatInstance, playerID string, state combatentity.CombatState, hpMax int32, track bool, grit [7]int, seen [7]bool) fightOnceResult {
	out := fightOnceResult{
		state:    state,
		rounds:   inst.Round,
		hpMax:    hpMax,
		hpEnd:    0,
		grit:     grit,
		gritSeen: seen,
	}
	for _, p := range inst.Players {
		if p.CurrentHP > 0 {
			out.hpEnd += p.CurrentHP
		}
	}
	if !track {
		return out
	}
	for _, entry := range inst.Log {
		if entry.ActorID != playerID || entry.Damage <= 0 {
			continue
		}
		switch {
		case entry.Result == "grit":
			out.retaliate += int64(entry.Damage)
		case entry.Action == combatentity.CombatActionSkill && strings.Contains(strings.ToLower(entry.Message), "slam"):
			out.slam += int64(entry.Damage)
		case entry.Action == combatentity.CombatActionAttack:
			out.swing += int64(entry.Damage)
		}
	}
	return out
}

func playerAct(engine *combatengine.Engine, inst *combatentity.CombatInstance, player *combatentity.CombatantRef, targetID string) {
	switch player.ClassID {
	case "warrior":
		if used, keeps := useKit(engine, inst, player, "warrior_brace", targetID); used {
			if keeps {
				engine.ProcessAttack(inst, player.ID, targetID)
			}
			return
		}
		if used, _ := useKit(engine, inst, player, "warrior_slam", targetID); used {
			return
		}
	case "wizard":
		if used, keeps := useKit(engine, inst, player, "mage_inscribe", targetID); used {
			if keeps {
				engine.ProcessAttack(inst, player.ID, targetID)
			}
			return
		}
	case "ward":
		if used, keeps := useKit(engine, inst, player, "ward_guard", "self"); used {
			if keeps {
				engine.ProcessAttack(inst, player.ID, targetID)
			}
			return
		}
		if used, _ := useKit(engine, inst, player, "ward_slam", targetID); used {
			return
		}
	case "rigger":
		if player.BoltLeft > 0 {
			if used, keeps := useKit(engine, inst, player, "rigger_bolt", targetID); used {
				if keeps {
					engine.ProcessAttack(inst, player.ID, targetID)
				}
				return
			}
		}
		if player.RigLeft > 0 && inst.Rig == nil {
			if used, keeps := useKit(engine, inst, player, "rigger_rig", targetID); used {
				if keeps {
					engine.ProcessAttack(inst, player.ID, targetID)
				}
				return
			}
		}
	}
	engine.ProcessAttack(inst, player.ID, targetID)
}

func useKit(engine *combatengine.Engine, inst *combatentity.CombatInstance, player *combatentity.CombatantRef, skillID, targetID string) (bool, bool) {
	if player == nil || !hasSkill(player, skillID) {
		return false, false
	}
	if player.KitSpent != nil && player.KitSpent[skillID] {
		return false, false
	}
	if player.SkillCooldowns != nil && player.SkillCooldowns[skillID] > 0 {
		return false, false
	}
	res := engine.ProcessSkill(inst, player.ID, skillID, targetID)
	if !res.Success {
		return false, false
	}
	return true, res.KeepsSwing
}

func hasSkill(player *combatentity.CombatantRef, id string) bool {
	for _, sid := range player.EquippedSkills {
		if sid == id {
			return true
		}
	}
	return false
}

func medianInts(xs []int) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]int(nil), xs...)
	sort.Ints(s)
	n := len(s)
	if n%2 == 1 {
		return float64(s[n/2])
	}
	return float64(s[n/2-1]+s[n/2]) / 2
}
