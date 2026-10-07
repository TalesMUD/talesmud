package game

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	combatpkg "github.com/talesmud/talesmud/pkg/mudserver/game/combat"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// enemyHookView is the Lua snapshot for one combatant. luar exposes ID/id,
// HP/hp, MaxHP/maxHp, and Alive/alive. HP is the fight's current HP.
type enemyHookView struct {
	ID    string
	Name  string
	HP    int32
	MaxHP int32
	Alive bool
}

func enemyHookViewFrom(ref *combat.CombatantRef) enemyHookView {
	if ref == nil {
		return enemyHookView{}
	}
	return enemyHookView{
		ID:    ref.ID,
		Name:  ref.Name,
		HP:    ref.CurrentHP,
		MaxHP: ref.MaxHP,
		Alive: ref.IsAlive && !ref.HasFled,
	}
}

func enemyHookScriptID(ref *combat.CombatantRef, hook string) string {
	if ref == nil {
		return ""
	}
	switch hook {
	case "onAggro":
		return ref.OnAggroScript
	case "onDeath":
		return ref.OnDeathScript
	case "onFlee":
		return ref.OnFleeScript
	case "onLowHealth":
		return ref.OnLowHealthScript
	default:
		return ""
	}
}

func enemyHookOpponents(instance *combat.CombatInstance) []enemyHookView {
	out := make([]enemyHookView, 0)
	if instance == nil {
		return out
	}
	for i := range instance.Players {
		out = append(out, enemyHookViewFrom(&instance.Players[i]))
	}
	return out
}

func enemyHookAllies(instance *combat.CombatInstance, selfID string) []enemyHookView {
	out := make([]enemyHookView, 0)
	if instance == nil {
		return out
	}
	for i := range instance.Enemies {
		if instance.Enemies[i].ID == selfID {
			continue
		}
		out = append(out, enemyHookViewFrom(&instance.Enemies[i]))
	}
	return out
}

// runEnemyHooksOnEnter fires onAggro once for each enemy that joined with this start.
func (c *CombatController) runEnemyHooksOnEnter(instance *combat.CombatInstance) {
	if c == nil || instance == nil {
		return
	}
	for i := range instance.Enemies {
		c.runEnemyHook(instance, &instance.Enemies[i], "onAggro")
	}
}

// MaxSummonsPerFight is the engine cap for tales.combat.summon. It is not content.
const MaxSummonsPerFight = 3

// flushEnemyLowHealthHooks fires onLowHealth once when an enemy's HP first
// crosses below its threshold while still alive. A hit that goes from
// above the line to dead does not fire (onDeath covers that).
func (c *CombatController) flushEnemyLowHealthHooks(instance *combat.CombatInstance) {
	if c == nil || instance == nil {
		return
	}
	beforeSummons := instance.SummonsUsed
	ids := make([]string, len(instance.Enemies))
	for i := range instance.Enemies {
		ids[i] = instance.Enemies[i].ID
	}
	for _, id := range ids {
		e := instance.GetEnemyByID(id)
		if e == nil {
			continue
		}
		if !e.LowHealthSeen {
			e.LowHealthSeen = true
			e.LowHealthHP = e.CurrentHP
			continue
		}
		before := e.LowHealthHP
		after := e.CurrentHP
		e.LowHealthHP = after
		if !e.IsAlive || e.HasFled || after <= 0 {
			continue
		}
		threshold := e.LowHealthThreshold
		if threshold <= 0 || threshold >= 1 {
			threshold = npc.DefaultLowHealthFraction
		}
		line := threshold * float64(e.MaxHP)
		if float64(before) >= line && float64(after) < line {
			c.runEnemyHook(instance, instance.GetEnemyByID(id), "onLowHealth")
		}
	}
	c.emitSummonRosterIfGrew(instance, beforeSummons)
}

// flushEnemyDeathHooks fires onDeath once for each enemy that is already dead.
// Safe to call again. It does not grant loot or XP.
func (c *CombatController) flushEnemyDeathHooks(instance *combat.CombatInstance) {
	if c == nil || instance == nil {
		return
	}
	beforeSummons := instance.SummonsUsed
	for i := range instance.Enemies {
		if instance.Enemies[i].IsAlive {
			continue
		}
		c.runEnemyHook(instance, &instance.Enemies[i], "onDeath")
	}
	c.emitSummonRosterIfGrew(instance, beforeSummons)
}

// runEnemyHook runs one sandboxed script. An empty script id is ignored.
// The hook is marked before the script runs, so a failing or re-entrant call
// still counts as once. Errors and panics are logged and swallowed.
func (c *CombatController) runEnemyHook(instance *combat.CombatInstance, enemy *combat.CombatantRef, hook string) {
	if c == nil || instance == nil || enemy == nil || enemy.ID == "" {
		return
	}
	scriptID := enemyHookScriptID(enemy, hook)
	if scriptID == "" {
		return
	}
	if instance.HookOnce == nil {
		instance.HookOnce = map[string]bool{}
	}
	key := enemy.ID + "|" + hook
	if instance.HookOnce[key] {
		return
	}
	instance.HookOnce[key] = true

	defer func() {
		if rec := recover(); rec != nil {
			log.WithFields(log.Fields{"hook": hook, "npc": enemy.ID, "panic": rec}).Error("enemy combat hook panicked")
		}
	}()

	if c.game == nil || c.game.Facade == nil || c.game.Facade.Runner() == nil {
		log.WithField("script", scriptID).Warn("enemy combat hook skipped; no script runner")
		return
	}
	facade := c.game.Facade
	script, err := facade.ScriptsService().FindByID(scriptID)
	if err != nil || script == nil {
		log.WithField("script", scriptID).WithError(err).Warn("enemy combat hook script not found")
		return
	}
	ctx := scripts.NewScriptContext()
	ctx.Set("hook", hook)
	ctx.SetCombatHook(hook, enemy.Name)
	ctx.Set("roomId", instance.OriginRoomID)
	ctx.Set("npc", enemyHookViewFrom(enemy))
	ctx.Set("opponents", enemyHookOpponents(instance))
	ctx.Set("allies", enemyHookAllies(instance, enemy.ID))
	run := facade.Runner().RunWithResult(*script, ctx)
	if run == nil || !run.Success {
		errText := ""
		if run != nil {
			errText = run.Error
		}
		log.WithFields(log.Fields{"script": script.Name, "hook": hook, "error": errText}).Warn("enemy combat hook failed")
	}
}

// holdNPCAttack spends a slow enemy's attack action without swinging.
// The caller sends the line. This does not take c.mu.
func (c *CombatController) holdNPCAttack(instance *combat.CombatInstance, actorID string) (bool, string) {
	if c == nil || instance == nil || c.engine == nil {
		return false, ""
	}
	actor := instance.GetCombatantByID(actorID)
	if actor == nil || actor.Type != combat.CombatantTypeNPC {
		return false, ""
	}
	if !combatpkg.EnemyHoldsAttack(actor.AttackSpeed, actor.AttackActions) {
		return false, ""
	}
	actor.AttackActions++
	c.engine.UpdateCombatant(instance, actor)
	msg := fmt.Sprintf("%s is slow to swing.", actor.Name)
	instance.AddLogEntry(combat.CombatLogEntry{
		ActorID:   actor.ID,
		ActorName: actor.Name,
		Action:    combat.CombatActionAttack,
		Result:    "hold",
		Message:   msg,
	})
	return true, msg
}

// HealCombatNPC restores HP on a living enemy in the active fight.
// It does not take c.mu, grant loot, or grant XP. A dead combatant returns 0.
func (c *CombatController) HealCombatNPC(npcID string, amount int32) int32 {
	if c == nil || c.manager == nil || c.engine == nil || npcID == "" || amount <= 0 {
		return 0
	}
	instance := c.manager.GetInstanceByNPCID(npcID)
	if instance == nil {
		return 0
	}
	ref := instance.GetCombatantByID(npcID)
	if ref == nil || ref.Type != combat.CombatantTypeNPC || !ref.IsAlive || ref.HasFled {
		return 0
	}
	missing := ref.MaxHP - ref.CurrentHP
	if missing <= 0 {
		return 0
	}
	healed := amount
	if healed > missing {
		healed = missing
	}
	ref.CurrentHP += healed
	c.engine.UpdateCombatant(instance, ref)
	if c.game != nil && c.game.NPCManager != nil {
		c.game.NPCManager.UpdateInstance(npcID, func(n *npc.NPC) {
			n.CurrentHitPoints = ref.CurrentHP
		})
	}
	c.notifyPlayersInCombat(instance, fmt.Sprintf("%s is healed for %d.", ref.Name, healed))
	return healed
}

// ApplyCombatEffect applies an existing buff or debuff by skill id.
// Damage, heal, and damage-over-time skills return false. Duration 0 lasts one round.
// This does not take c.mu.
func (c *CombatController) ApplyCombatEffect(targetID, effectID string) bool {
	if c == nil || c.manager == nil || c.engine == nil || c.game == nil || c.game.Facade == nil {
		return false
	}
	if targetID == "" || effectID == "" {
		return false
	}
	skill, err := c.game.Facade.SkillsService().FindByID(effectID)
	if err != nil || skill == nil || skill.Entity == nil {
		return false
	}
	if skill.Effect != skills.EffectBuff && skill.Effect != skills.EffectDebuff {
		return false
	}
	instance := c.manager.GetInstanceByPlayerID(targetID)
	if instance == nil {
		instance = c.manager.GetInstanceByNPCID(targetID)
	}
	if instance == nil {
		return false
	}
	duration := skill.Duration
	if duration < 1 {
		duration = 1
	}
	se := combat.StatusEffect{
		SkillID:  skill.Entity.ID,
		Name:     skill.Name,
		Type:     string(skill.Effect),
		Stat:     skill.BuffStat,
		Percent:  skill.BuffPercent,
		Duration: duration,
		SourceID: "script",
	}
	return c.engine.ApplyStatusEffectFromScript(instance, targetID, se)
}

// clampSummonCount limits one tales.combat.summon call to 1..3 and the remaining fight cap.
func clampSummonCount(count, remaining int) int {
	if remaining <= 0 {
		return 0
	}
	if count < 1 {
		count = 1
	}
	if count > MaxSummonsPerFight {
		count = MaxSummonsPerFight
	}
	if count > remaining {
		return remaining
	}
	return count
}

// SummonCombatAllies spawns enemy-template adds into the summoner's current fight.
// Adds use the spawner instance path (HP and stats from the template) but are not
// registered on a spawner. They grant no loot, XP, or quest credit, and they are
// removed when the fight ends. Each new add may run onAggro once. The per-fight
// cap stops a summon loop. Returns how many this call spawned.
func (c *CombatController) SummonCombatAllies(summonerNPCID, templateID string, count int) int {
	if c == nil || c.manager == nil || c.engine == nil || c.game == nil || c.game.NPCManager == nil {
		log.Warn("summon: no active fight")
		return 0
	}
	instance := c.manager.GetInstanceByNPCID(summonerNPCID)
	if instance == nil || instance.State != combat.CombatStateActive || instance.OriginRoomID == "" {
		log.WithFields(log.Fields{"npc": summonerNPCID, "template": templateID}).Warn("summon: no active fight")
		return 0
	}
	if c.game.Facade == nil {
		log.WithField("template", templateID).Warn("summon: unknown template")
		return 0
	}
	template, err := c.game.Facade.NPCsService().FindByID(templateID)
	if err != nil || template == nil || !template.IsTemplate {
		log.WithField("template", templateID).WithError(err).Warn("summon: unknown template")
		return 0
	}
	if template.EnemyTrait == nil {
		log.WithField("template", templateID).Warn("summon: not an enemy")
		return 0
	}
	remaining := MaxSummonsPerFight - instance.SummonsUsed
	want := clampSummonCount(count, remaining)
	if want == 0 {
		log.WithFields(log.Fields{"npc": summonerNPCID, "template": templateID, "used": instance.SummonsUsed}).Warn("summon: cap reached")
		return 0
	}

	currentID := ""
	if instance.CurrentTurnIdx >= 0 && instance.CurrentTurnIdx < len(instance.TurnOrder) {
		currentID = instance.TurnOrder[instance.CurrentTurnIdx].ID
	}

	spawned := 0
	for spawned < want {
		if instance.SummonsUsed >= MaxSummonsPerFight {
			log.WithFields(log.Fields{"template": templateID, "used": instance.SummonsUsed}).Warn("summon: cap reached")
			break
		}
		inst, spawnErr := c.game.NPCManager.SpawnInstanceDirect(templateID, instance.OriginRoomID)
		if spawnErr != nil || inst == nil || inst.Entity == nil {
			log.WithField("template", templateID).WithError(spawnErr).Warn("summon: unknown template")
			break
		}
		c.game.NPCManager.UpdateInstance(inst.Entity.ID, func(n *npc.NPC) {
			n.InCombat = true
			n.CombatInstanceID = instance.ID
			n.State = "combat"
		})
		ref := c.engine.CreateCombatantFromNPC(inst)
		ref.Summoned = true
		c.engine.RollInitiative(&ref)
		instance.Enemies = append(instance.Enemies, ref)
		instance.SummonsUsed++
		c.manager.RegisterNPC(inst.Entity.ID, instance.ID)
		c.engine.BuildTurnOrder(instance)
		if currentID != "" {
			for i := range instance.TurnOrder {
				if instance.TurnOrder[i].ID == currentID {
					instance.CurrentTurnIdx = i
					break
				}
			}
		}
		spawned++
		if added := instance.GetEnemyByID(inst.Entity.ID); added != nil {
			c.runEnemyHook(instance, added, "onAggro")
		}
	}
	return spawned
}
