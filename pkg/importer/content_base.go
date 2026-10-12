package importer

import (
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

// ApplyContentBase writes the combat stats from enemyTrait.baseStats.
// NPCs without a content base are left alone, so editor-authored numbers
// stay exactly as entered. A full-health NPC stays full after rescaling.
// A wounded NPC keeps its current HP unless that is above the new maximum.
func ApplyContentBase(n *npc.NPC) {
	if n == nil || n.EnemyTrait == nil || n.EnemyTrait.BaseStats == nil {
		return
	}
	base := n.EnemyTrait.BaseStats
	hp, attack, defense := balance.ApplyEnemyMultipliers(
		base.MaxHitPoints,
		base.AttackPower,
		base.Defense,
		n.EnemyTrait.Difficulty,
		n.Name,
	)
	oldMax := n.MaxHitPoints
	full := oldMax <= 0 || n.CurrentHitPoints >= oldMax
	n.MaxHitPoints = hp
	n.EnemyTrait.AttackPower = attack
	n.EnemyTrait.Defense = defense
	if full || n.CurrentHitPoints > hp {
		n.CurrentHitPoints = hp
	}
}
