package simutil

import (
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	combatentity "github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	combatengine "github.com/talesmud/talesmud/pkg/mudserver/game/combat"
)

func init() {
	// Suppress combat engine log spam during simulations
	log.SetLevel(log.WarnLevel)
}

const (
	DefaultIterations = 500
	MaxSimRounds      = 100
)

// MatchupConfig describes one simulation matchup
type MatchupConfig struct {
	PlayerConfigs []ClassConfig // Player characters (usually 1)
	PlayerLevel   int32
	EnemyConfigs  []EnemyConfig // Enemy NPCs (usually 1)
	Iterations    int
}

// SingleResult holds the outcome of one combat simulation
type SingleResult struct {
	State       combatentity.CombatState
	Rounds      int
	PlayerHPMax int32
	PlayerHPEnd int32
	EnemyHPMax  int32
	EnemyHPEnd  int32
}

// SimResult holds aggregated results from N simulations of one matchup
type SimResult struct {
	Label       string // e.g. "Warrior L5 vs Meadow Wolf"
	PlayerName  string
	PlayerLevel int32
	EnemyName   string
	EnemyLevel  int32
	Iterations  int

	// Combat stat snapshots (from first simulation, for debugging)
	PlayerAttackPower int32
	PlayerDefense     int32
	PlayerMaxHP       int32
	PlayerSTRMod      int
	EnemyAttackPower  int32
	EnemyDefense      int32
	EnemyMaxHP        int32

	// Aggregated outcomes
	PlayerWins int
	EnemyWins  int
	Timeouts   int
	WinRate    float64

	// Averages
	AvgRounds          float64
	AvgPlayerHPPercent float64 // Average remaining HP % for player wins
	AvgEnemyHPPercent  float64 // Average remaining HP % for enemy wins
}

// RunSimulation runs a single combat to completion using the real combat engine.
// Players always attack the first living enemy. NPCs use the engine's AI.
func RunSimulation(players []*characters.Character, enemies []*npc.NPC) SingleResult {
	manager := combatengine.NewManager()
	engine := combatengine.NewEngine(manager, nil)

	instance := engine.InitiateCombat("sim-room", players, enemies)

	var totalPlayerMaxHP, totalEnemyMaxHP int32
	for _, p := range instance.Players {
		totalPlayerMaxHP += p.MaxHP
	}
	for _, e := range instance.Enemies {
		totalEnemyMaxHP += e.MaxHP
	}

	for round := 0; round < MaxSimRounds; round++ {
		current := instance.GetCurrentTurnCombatant()
		if current == nil {
			break
		}

		if current.Type == combatentity.CombatantTypeNPC {
			action, targetID := engine.GetNPCAIAction(instance, current, nil)
			switch action {
			case combatentity.CombatActionAttack:
				engine.ProcessAttack(instance, current.ID, targetID)
			case combatentity.CombatActionDefend:
				engine.ProcessDefend(instance, current.ID)
			}
		} else {
			// Skill-aware player AI
			livingEnemies := instance.GetLivingEnemies()
			if len(livingEnemies) > 0 {
				targetID := livingEnemies[0].ID
				skillUsed := false

				// Try to use a skill
				player := instance.GetPlayerByID(current.ID)
				if player != nil && len(player.EquippedSkills) > 0 {
					skillUsed = tryPlayerSkill(engine, instance, player, targetID)
				}

				if !skillUsed {
					engine.ProcessAttack(instance, current.ID, targetID)
				}
			}
		}

		endState := engine.CheckCombatEnd(instance)
		if endState != combatentity.CombatStateActive {
			return SingleResult{
				State:       endState,
				Rounds:      instance.Round,
				PlayerHPMax: totalPlayerMaxHP,
				PlayerHPEnd: sumHP(instance.Players),
				EnemyHPMax:  totalEnemyMaxHP,
				EnemyHPEnd:  sumHP(instance.Enemies),
			}
		}

		engine.NextTurn(instance)
	}

	return SingleResult{
		State:       combatentity.CombatStateTimeout,
		Rounds:      MaxSimRounds,
		PlayerHPMax: totalPlayerMaxHP,
		PlayerHPEnd: sumHP(instance.Players),
		EnemyHPMax:  totalEnemyMaxHP,
		EnemyHPEnd:  sumHP(instance.Enemies),
	}
}

// RunMatchup runs N iterations of a matchup and aggregates results
func RunMatchup(config MatchupConfig) *SimResult {
	iterations := config.Iterations
	if iterations <= 0 {
		iterations = DefaultIterations
	}

	result := &SimResult{
		PlayerName:  config.PlayerConfigs[0].Name,
		PlayerLevel: config.PlayerLevel,
		EnemyName:   config.EnemyConfigs[0].Name,
		EnemyLevel:  config.EnemyConfigs[0].Level,
		Iterations:  iterations,
	}

	// Capture stat snapshots from a single character/enemy
	sampleChar := CreateCharacter(config.PlayerConfigs[0], config.PlayerLevel)
	result.PlayerAttackPower = sampleChar.GetWeaponDamage() + int32(sampleChar.GetPrimaryAttackMod())
	if result.PlayerAttackPower < 1 {
		result.PlayerAttackPower = 1
	}
	result.PlayerDefense = sampleChar.GetArmorDefense()
	result.PlayerMaxHP = sampleChar.MaxHitPoints
	result.PlayerSTRMod = sampleChar.GetPrimaryAttackMod()
	result.EnemyAttackPower = config.EnemyConfigs[0].AttackPower
	result.EnemyDefense = config.EnemyConfigs[0].Defense
	result.EnemyMaxHP = config.EnemyConfigs[0].HP

	result.Label = formatLabel(result.PlayerName, result.PlayerLevel, result.EnemyName, result.EnemyLevel)

	var totalRounds float64
	var totalPlayerHPPercent float64
	var totalEnemyHPPercent float64
	var playerWinCount, enemyWinCount int

	for i := 0; i < iterations; i++ {
		// Create fresh characters and enemies each iteration
		players := make([]*characters.Character, len(config.PlayerConfigs))
		for j, pc := range config.PlayerConfigs {
			players[j] = CreateCharacter(pc, config.PlayerLevel)
		}
		enemies := make([]*npc.NPC, len(config.EnemyConfigs))
		for j, ec := range config.EnemyConfigs {
			enemies[j] = CreateEnemy(ec)
		}

		single := RunSimulation(players, enemies)
		totalRounds += float64(single.Rounds)

		switch single.State {
		case combatentity.CombatStateVictory:
			result.PlayerWins++
			playerWinCount++
			if single.PlayerHPMax > 0 {
				totalPlayerHPPercent += float64(single.PlayerHPEnd) / float64(single.PlayerHPMax) * 100
			}
		case combatentity.CombatStateDefeat:
			result.EnemyWins++
			enemyWinCount++
			if single.EnemyHPMax > 0 {
				totalEnemyHPPercent += float64(single.EnemyHPEnd) / float64(single.EnemyHPMax) * 100
			}
		default:
			result.Timeouts++
		}
	}

	result.WinRate = float64(result.PlayerWins) / float64(iterations)
	result.AvgRounds = totalRounds / float64(iterations)
	if playerWinCount > 0 {
		result.AvgPlayerHPPercent = totalPlayerHPPercent / float64(playerWinCount)
	}
	if enemyWinCount > 0 {
		result.AvgEnemyHPPercent = totalEnemyHPPercent / float64(enemyWinCount)
	}

	return result
}

// RunAllMatchups runs multiple matchup configurations and returns all results
func RunAllMatchups(configs []MatchupConfig) []*SimResult {
	results := make([]*SimResult, 0, len(configs))
	for _, c := range configs {
		results = append(results, RunMatchup(c))
	}
	return results
}

// BuildStandardMatchups generates the standard test matrix:
// All 6 classes at levels 1, 5, 10 vs all 11 enemies
func BuildStandardMatchups(iterations int) []MatchupConfig {
	classes := AllClassConfigs()
	enemies := AllEnemyConfigs()
	levels := []int32{1, 5, 10}

	configs := make([]MatchupConfig, 0, len(classes)*len(enemies)*len(levels))

	for _, cls := range classes {
		for _, lvl := range levels {
			for _, enm := range enemies {
				configs = append(configs, MatchupConfig{
					PlayerConfigs: []ClassConfig{cls},
					PlayerLevel:   lvl,
					EnemyConfigs:  []EnemyConfig{enm},
					Iterations:    iterations,
				})
			}
		}
	}

	return configs
}

func sumHP(combatants []combatentity.CombatantRef) int32 {
	var total int32
	for _, c := range combatants {
		if c.CurrentHP > 0 {
			total += c.CurrentHP
		}
	}
	return total
}

func formatLabel(playerName string, playerLevel int32, enemyName string, enemyLevel int32) string {
	return playerName + " L" + itoa(int(playerLevel)) + " vs " + enemyName + " L" + itoa(int(enemyLevel))
}

// tryPlayerSkill attempts to use an available skill. Returns true if a skill was used.
// AI priority: damage skills first (highest base power), then buffs if none active, then heals if HP low.
func tryPlayerSkill(engine *combatengine.Engine, instance *combatentity.CombatInstance, player *combatentity.CombatantRef, targetID string) bool {
	var bestDamageSkill *skills.Skill
	var bestBuffSkill *skills.Skill
	var bestHealSkill *skills.Skill

	for _, sid := range player.EquippedSkills {
		skill := skills.SkillByID(sid)
		if skill == nil {
			continue
		}

		// Check cooldown
		if cd, ok := player.SkillCooldowns[sid]; ok && cd > 0 {
			continue
		}

		// Check mana
		if skill.ResourceType == skills.ResourceMana && player.CurrentMana < skill.ManaCost {
			continue
		}

		switch skill.Effect {
		case skills.EffectDamage:
			if bestDamageSkill == nil || skill.BasePower > bestDamageSkill.BasePower {
				bestDamageSkill = skill
			}
		case skills.EffectBuff:
			if bestBuffSkill == nil {
				bestBuffSkill = skill
			}
		case skills.EffectHeal, skills.EffectHot:
			if bestHealSkill == nil {
				bestHealSkill = skill
			}
		case skills.EffectDebuff, skills.EffectDot:
			// Treat debuffs/DoTs as secondary damage options
			if bestDamageSkill == nil {
				bestDamageSkill = skill
			}
		}
	}

	// Decision logic:
	// 1. If HP below 40% and we have a heal, heal
	if bestHealSkill != nil && player.MaxHP > 0 {
		hpPercent := float64(player.CurrentHP) / float64(player.MaxHP)
		if hpPercent < 0.4 {
			engine.ProcessSkill(instance, player.ID, bestHealSkill.ID, player.ID)
			return true
		}
	}

	// 2. If we have a buff and no active buffs from this skill, use it (round 1 priority)
	if bestBuffSkill != nil && !hasActiveBuffFromSkill(player, bestBuffSkill.ID) {
		engine.ProcessSkill(instance, player.ID, bestBuffSkill.ID, player.ID)
		return true
	}

	// 3. Use best damage skill
	if bestDamageSkill != nil {
		tgt := targetID
		if bestDamageSkill.Target == skills.TargetSelf {
			tgt = player.ID
		}
		engine.ProcessSkill(instance, player.ID, bestDamageSkill.ID, tgt)
		return true
	}

	return false
}

// hasActiveBuffFromSkill checks if the player already has an active status effect from a specific skill
func hasActiveBuffFromSkill(player *combatentity.CombatantRef, skillID string) bool {
	for _, se := range player.StatusEffects {
		if se.SkillID == skillID && se.Duration > 0 {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
