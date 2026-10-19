package simutil

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	combatentity "github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	combatengine "github.com/talesmud/talesmud/pkg/mudserver/game/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
)

// EarlySeed is the fixed draw for the early-game table. Each cell reseeds it.
const EarlySeed int64 = 20261018

// EarlyEnemy is one content enemy at its authored (pre-multiplier) stats.
// CreateEnemy applies the live difficulty multipliers, which is what import stores.
type EarlyEnemy struct {
	ID         string
	Name       string
	Level      int32
	HP         int32
	Attack     int32
	Defense    int32
	Difficulty string
}

// EarlyPackEnemies is every L1–6 non-boss with a live spawner in the
// talesmud-rpg-1 pack (content 68d287d). Stats are the YAML numbers.
// Sample Bear is difficulty hard (elite). The rest are trash tiers.
func EarlyPackEnemies() []EarlyEnemy {
	return []EarlyEnemy{
		{"ENM0001", "Sample Rat", 1, 8, 1, 0, "trivial"},
		{"ENM0002", "Sample Wolf", 2, 18, 4, 1, "normal"},
		{"ENM0003", "Sample Boar", 2, 25, 5, 2, "normal"},
		{"ENM0007", "Sample Thug", 2, 18, 4, 1, "normal"},
		{"ENM0008", "Sample Pest", 2, 12, 3, 0, "normal"},
		{"ENM0010", "Sample Burrower", 2, 15, 3, 2, "normal"},
		{"ENM0004", "Sample Bandit", 3, 22, 5, 2, "normal"},
		{"ENM0006", "Sample Shade", 4, 30, 7, 0, "normal"},
		{"ENM0012", "Sample Servitor", 4, 40, 7, 3, "normal"},
		{"ENM0005", "Sample Bear", 5, 80, 11, 4, "hard"},
	}
}

// EarlyRow is one class × level × enemy cell.
type EarlyRow struct {
	ClassID    string
	ClassName  string
	Level      int32
	EnemyID    string
	EnemyName  string
	EnemyLevel int32
	Difficulty string
	Gap        int
	Fights     int
	Wins       int
	Losses     int
	Timeouts   int
	WinRate    float64
	MedRounds  float64
	MedHPLeft  float64
	PlayerHP   int32
	EnemyHP    int32
	EnemyATK   int32
	EnemyDEF   int32
}

// EarlyLevels are the fresh-character levels in the early-game table.
func EarlyLevels() []int32 { return []int32{1, 3, 5} }

// RunEarlyTable fights each pack class at L1, L3, and L5 in starter gear
// against every early enemy. iters below 1 becomes 80.
func RunEarlyTable(iters int) []EarlyRow {
	if iters < 1 {
		iters = 80
	}
	var rows []EarlyRow
	for _, preset := range characters.SystemCharacterTemplatePresets() {
		if preset == nil {
			continue
		}
		for _, level := range EarlyLevels() {
			for _, enemy := range EarlyPackEnemies() {
				rows = append(rows, RunEarlyCell(preset, level, enemy, iters))
			}
		}
	}
	return rows
}

// RunEarlyCell reseeds EarlySeed and fights one matchup.
func RunEarlyCell(preset *characters.CharacterTemplate, level int32, enemy EarlyEnemy, iters int) EarlyRow {
	if iters < 1 {
		iters = 80
	}
	rand.Seed(EarlySeed)
	sample := EarlyCharacter(preset, level)
	foe := enemyNPC(enemy)
	row := EarlyRow{
		ClassID:    sample.Class.ID,
		ClassName:  preset.Name,
		Level:      level,
		EnemyID:    enemy.ID,
		EnemyName:  enemy.Name,
		EnemyLevel: enemy.Level,
		Difficulty: enemy.Difficulty,
		Gap:        int(enemy.Level - level),
		Fights:     iters,
		PlayerHP:   sample.MaxHitPoints,
		EnemyHP:    foe.MaxHitPoints,
		EnemyATK:   foe.EnemyTrait.AttackPower,
		EnemyDEF:   foe.EnemyTrait.Defense,
	}
	var rounds []int
	var hpLeft []float64
	for i := 0; i < iters; i++ {
		ch := EarlyCharacter(preset, level)
		one := fightEarly(ch, []*npc.NPC{enemyNPC(enemy)})
		switch one.state {
		case combatentity.CombatStateVictory:
			row.Wins++
			rounds = append(rounds, one.rounds)
			if one.hpMax > 0 {
				hpLeft = append(hpLeft, float64(one.hpEnd)/float64(one.hpMax)*100)
			}
		case combatentity.CombatStateDefeat:
			row.Losses++
			rounds = append(rounds, one.rounds)
		default:
			row.Timeouts++
		}
	}
	row.WinRate = float64(row.Wins) / float64(iters)
	row.MedRounds = medianInts(rounds)
	row.MedHPLeft = medianFloats(hpLeft)
	return row
}

// RunBountyFight fights one enemy at playerLevel. healBetween is unused for a
// single fight. count fights in a row. When heal is true, HP and mana refill
// between fights (out-of-combat rest). A chain counts as a win only if every
// fight is a victory.
func RunBountyFight(preset *characters.CharacterTemplate, playerLevel int32, enemy EarlyEnemy, count int, heal bool, gear string, iters int) EarlyRow {
	if count < 1 {
		count = 1
	}
	if iters < 1 {
		iters = 80
	}
	rand.Seed(EarlySeed)
	sample := EarlyCharacter(preset, playerLevel)
	if gear == GearAppropriate || gear == GearGood {
		applyGearScores(sample, playerLevel, gear)
	}
	foe := enemyNPC(enemy)
	row := EarlyRow{
		ClassID:    sample.Class.ID,
		ClassName:  preset.Name,
		Level:      playerLevel,
		EnemyID:    enemy.ID,
		EnemyName:  enemy.Name,
		EnemyLevel: enemy.Level,
		Difficulty: enemy.Difficulty,
		Gap:        int(enemy.Level - playerLevel),
		Fights:     iters,
		PlayerHP:   sample.MaxHitPoints,
		EnemyHP:    foe.MaxHitPoints,
		EnemyATK:   foe.EnemyTrait.AttackPower,
		EnemyDEF:   foe.EnemyTrait.Defense,
	}
	var rounds []int
	var hpLeft []float64
	for i := 0; i < iters; i++ {
		ch := EarlyCharacter(preset, playerLevel)
		if gear == GearAppropriate || gear == GearGood {
			applyGearScores(ch, playerLevel, gear)
		}
		won := 0
		var last fightOutcome
		totalRounds := 0
		for n := 0; n < count; n++ {
			if n > 0 && heal {
				ch.CurrentHitPoints = ch.MaxHitPoints
				ch.CurrentMana = ch.MaxMana
			}
			last = fightEarly(ch, []*npc.NPC{enemyNPC(enemy)})
			totalRounds += last.rounds
			if last.state != combatentity.CombatStateVictory {
				break
			}
			won++
			ch.CurrentHitPoints = last.hpEnd
		}
		if won == count {
			row.Wins++
			rounds = append(rounds, totalRounds)
			if last.hpMax > 0 {
				hpLeft = append(hpLeft, float64(last.hpEnd)/float64(last.hpMax)*100)
			}
		} else if last.state == combatentity.CombatStateDefeat {
			row.Losses++
			rounds = append(rounds, totalRounds)
		} else {
			row.Timeouts++
		}
	}
	row.WinRate = float64(row.Wins) / float64(iters)
	row.MedRounds = medianInts(rounds)
	row.MedHPLeft = medianFloats(hpLeft)
	return row
}

// EarlyCharacter is a create-screen character at level, still in starter gear.
// Levels above 1 use ApplyLevelUp and spend points on the class primary.
// The hotbar is the kit skills unlocked at that level.
func EarlyCharacter(preset *characters.CharacterTemplate, level int32) *characters.Character {
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
	if level > ch.Level {
		leveling.ApplyLevelUp(ch, int(level-ch.Level))
	}
	autoSpendPrimaryAttributes(ch)
	ch.CurrentHitPoints = ch.MaxHitPoints
	ch.CurrentMana = ch.MaxMana
	ch.EquippedSkills = skills.FillHotbar(ch.Class.ID, ch.Level, append([]string(nil), preset.DefaultSkills...))
	return ch
}

func applyGearScores(ch *characters.Character, level int32, gear string) {
	if ch == nil {
		return
	}
	// Same bump the gap table uses, measured from this character's starter kit.
	weapon := int(ch.GetWeaponDamage()) + int(level)/2
	armor := int(ch.GetArmorDefense()) + int(level)/3
	if weapon < 1 {
		weapon = 1
	}
	if gear == GearGood {
		weapon = int(float64(weapon)*2.2 + 0.5)
		armor = int(float64(armor)*2.2 + 0.5)
	}
	if ch.EquippedItems == nil {
		ch.EquippedItems = map[items.ItemSlot]*items.Item{}
	}
	// Keep the starter subtype so racial weapon rules still apply.
	if hand := ch.EquippedItems[items.ItemSlotMainHand]; hand != nil {
		if hand.Attributes == nil {
			hand.Attributes = map[string]interface{}{}
		}
		hand.Attributes["damage"] = weapon
	}
	if chest := ch.EquippedItems[items.ItemSlotChest]; chest != nil {
		if chest.Attributes == nil {
			chest.Attributes = map[string]interface{}{}
		}
		chest.Attributes["armor"] = armor
	}
}

func enemyNPC(e EarlyEnemy) *npc.NPC {
	return CreateEnemy(EnemyConfig{
		Name:        e.Name,
		Level:       e.Level,
		HP:          e.HP,
		AttackPower: e.Attack,
		Defense:     e.Defense,
		Difficulty:  e.Difficulty,
	})
}

type fightOutcome struct {
	state  combatentity.CombatState
	rounds int
	hpMax  int32
	hpEnd  int32
}

func fightEarly(ch *characters.Character, enemies []*npc.NPC) fightOutcome {
	engine := combatengine.NewEngine(combatengine.NewManager(), nil)
	inst := engine.InitiateCombat("sim-room", []*characters.Character{ch}, enemies)
	var hpMax int32
	for _, p := range inst.Players {
		hpMax += p.MaxHP
	}
	const maxTurns = 300
	state := combatentity.CombatStateTimeout
	for turn := 0; turn < maxTurns; turn++ {
		current := inst.GetCurrentTurnCombatant()
		if current == nil {
			state = engine.CheckCombatEnd(inst)
			if state != combatentity.CombatStateActive {
				break
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
			living := inst.GetLivingEnemies()
			if len(living) == 0 {
				state = engine.CheckCombatEnd(inst)
				break
			}
			actEarly(engine, inst, player, living[0].ID)
		}
		state = engine.CheckCombatEnd(inst)
		if state != combatentity.CombatStateActive {
			break
		}
		engine.NextTurn(inst)
	}
	if state == combatentity.CombatStateActive {
		state = engine.CheckCombatEnd(inst)
		if state == combatentity.CombatStateActive {
			state = combatentity.CombatStateTimeout
		}
	}
	var hpEnd int32
	for _, p := range inst.Players {
		if p.CurrentHP > 0 {
			hpEnd += p.CurrentHP
		}
	}
	return fightOutcome{state: state, rounds: inst.Round, hpMax: hpMax, hpEnd: hpEnd}
}

// actEarly uses the level-1 kit the way a new player does, plus the damage
// buttons that unlock by L5. Slip ends the fight, so the rogue swings instead.
func actEarly(engine *combatengine.Engine, inst *combatentity.CombatInstance, player *combatentity.CombatantRef, targetID string) {
	try := func(skillID, target string) (bool, bool) {
		return useEarlyKit(engine, inst, player, skillID, target)
	}
	switch player.ClassID {
	case "warrior":
		if used, keeps := try("warrior_brace", targetID); used {
			if keeps {
				engine.ProcessAttack(inst, player.ID, targetID)
			}
			return
		}
		if used, _ := try("warrior_slam", targetID); used {
			return
		}
	case "wizard":
		if used, keeps := try("mage_inscribe", targetID); used {
			if keeps {
				engine.ProcessAttack(inst, player.ID, targetID)
			}
			return
		}
		if used, _ := try("mage_sear", targetID); used {
			return
		}
	case "ward":
		if used, keeps := try("ward_guard", "self"); used {
			if keeps {
				engine.ProcessAttack(inst, player.ID, targetID)
			}
			return
		}
		if used, _ := try("ward_slam", targetID); used {
			return
		}
	case "rigger":
		if player.BoltLeft > 0 {
			if used, keeps := try("rigger_bolt", targetID); used {
				if keeps {
					engine.ProcessAttack(inst, player.ID, targetID)
				}
				return
			}
		}
		if player.RigLeft > 0 && inst.Rig == nil {
			if used, keeps := try("rigger_rig", targetID); used {
				if keeps {
					engine.ProcessAttack(inst, player.ID, targetID)
				}
				return
			}
		}
	case "rogue":
		if used, _ := try("rogue_nick", targetID); used {
			return
		}
	}
	engine.ProcessAttack(inst, player.ID, targetID)
}

func useEarlyKit(engine *combatengine.Engine, inst *combatentity.CombatInstance, player *combatentity.CombatantRef, skillID, targetID string) (bool, bool) {
	if player == nil {
		return false, false
	}
	has := false
	for _, sid := range player.EquippedSkills {
		if sid == skillID {
			has = true
			break
		}
	}
	if !has {
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

func medianFloats(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

// FormatEarlyMarkdown renders the early-game table grouped by enemy.
func FormatEarlyMarkdown(rows []EarlyRow) string {
	var b []byte
	b = append(b, []byte("# Early-game starter-gear table\n\n")...)
	b = append(b, []byte("Fresh pack classes at L1, L3, and L5 in starter gear. ")...)
	b = append(b, []byte("Enemy HP/ATK/DEF are the live values after difficulty multipliers. ")...)
	b = append(b, []byte("Gap is enemy level minus player level. ")...)
	b = append(b, []byte("Win% is the player. Rounds and HP left are medians. HP left is on wins.\n\n")...)
	type key struct{ id, name string }
	var order []key
	seen := map[string]bool{}
	for _, row := range rows {
		if seen[row.EnemyID] {
			continue
		}
		seen[row.EnemyID] = true
		order = append(order, key{row.EnemyID, row.EnemyName})
	}
	for _, k := range order {
		var sample EarlyRow
		for _, row := range rows {
			if row.EnemyID == k.id {
				sample = row
				break
			}
		}
		b = append(b, []byte(fmt.Sprintf("## %s (%s) L%d %s — live %d HP / %d ATK / %d DEF\n\n",
			k.name, k.id, sample.EnemyLevel, sample.Difficulty, sample.EnemyHP, sample.EnemyATK, sample.EnemyDEF))...)
		b = append(b, []byte("| Class | Lvl | Gap | Win% | Med rounds | Med HP left |\n")...)
		b = append(b, []byte("| --- | --- | --- | --- | --- | --- |\n")...)
		for _, row := range rows {
			if row.EnemyID != k.id {
				continue
			}
			b = append(b, []byte(fmt.Sprintf("| %s | %d | %+d | %.0f%% | %.1f | %.0f%% |\n",
				row.ClassName, row.Level, row.Gap, row.WinRate*100, row.MedRounds, row.MedHPLeft))...)
		}
		b = append(b, '\n')
	}
	return string(b)
}
