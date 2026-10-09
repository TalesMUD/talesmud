package exporter

import (
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/importer"
	"gopkg.in/yaml.v3"
)

// ContentCombatStats returns the max HP, attack, and defense a content YAML
// file should store. When the NPC has base stats from import, those are the
// YAML values so a later import does not apply difficulty multipliers twice.
// Without a content base, the stored combat stats are the authored values.
func ContentCombatStats(n *npc.NPC) (maxHP, attack, defense int32, fromBase bool) {
	if n == nil {
		return 0, 0, 0, false
	}
	if n.EnemyTrait != nil && n.EnemyTrait.BaseStats != nil {
		base := n.EnemyTrait.BaseStats
		return base.MaxHitPoints, base.AttackPower, base.Defense, true
	}
	if n.EnemyTrait != nil {
		attack = n.EnemyTrait.AttackPower
		defense = n.EnemyTrait.Defense
	}
	return n.MaxHitPoints, attack, defense, false
}

// ContentNPC builds the YAML shape of an NPC, using content base stats when
// they are present. Re-importing the result applies scaling once.
func ContentNPC(n *npc.NPC) *importer.YAMLNPC {
	if n == nil {
		return nil
	}
	maxHP, attack, defense, _ := ContentCombatStats(n)
	out := &importer.YAMLNPC{
		ID:           n.ID,
		Name:         n.Name,
		Description:  n.Description,
		Level:        n.Level,
		MaxHitPoints: maxHP,
		SpawnRoomId:  n.SpawnRoomID,
		DialogID:     n.DialogID,
	}
	if n.Race.ID != "" || n.Race.Name != "" {
		out.Race = importer.YAMLRace{ID: n.Race.ID, Name: n.Race.Name}
	}
	if n.Class.ID != "" || n.Class.Name != "" {
		out.Class = importer.YAMLClass{ID: n.Class.ID, Name: n.Class.Name}
	}
	if n.EnemyTrait == nil {
		return out
	}
	trait := n.EnemyTrait
	out.EnemyTrait = &importer.YAMLEnemyTrait{
		CreatureType:       trait.CreatureType,
		CombatStyle:        trait.CombatStyle,
		Difficulty:         trait.Difficulty,
		AttackPower:        attack,
		Defense:            defense,
		AttackSpeed:        trait.AttackSpeed,
		AggroRadius:        trait.AggroRadius,
		AggroOnSight:       trait.AggroOnSight,
		CallForHelp:        trait.CallForHelp,
		FleeThreshold:      trait.FleeThreshold,
		XPReward:           trait.XPReward,
		GoldDrop:           importer.YAMLRange{Min: trait.GoldDrop.Min, Max: trait.GoldDrop.Max},
		LootTableID:        trait.LootTableID,
		GuaranteedLoot:     append([]string{}, trait.GuaranteedLoot...),
		MaxDrops:           trait.MaxDrops,
		OnAggroScript:      trait.OnAggroScript,
		OnDeathScript:      trait.OnDeathScript,
		OnFleeScript:       trait.OnFleeScript,
		OnLowHealthScript:  trait.OnLowHealthScript,
		LowHealthThreshold: trait.LowHealthThreshold,
		ResetOnDisengage:   trait.ResetOnDisengage,
	}
	return out
}

// ContentNPCYAML marshals ContentNPC.
func ContentNPCYAML(n *npc.NPC) ([]byte, error) {
	return yaml.Marshal(ContentNPC(n))
}
