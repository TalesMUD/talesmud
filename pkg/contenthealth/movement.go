package contenthealth

import (
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

func (b *builder) addMovement() {
	check := func(kind, id string, cfg *npc.Movement, spawn string, n *npc.NPC) {
		if cfg == nil {
			return
		}
		issues := npc.ValidateMovement(cfg, b.snap.Rooms)
		if spawn != "" && (!npc.MovementRoomAllowed(cfg, b.snap.Rooms[spawn]) || (cfg.Mode == "route" && !npc.ContainsID(cfg.Route, spawn))) {
			issues = append(issues, "spawn room must be allowed and on the route")
		}
		if n != nil && n.EnemyTrait != nil && n.EnemyTrait.AggroOnSight && (n.Level < 10 || n.EnemyTrait.Difficulty == "boss" || npc.ContainsID(n.Tags, "boss")) {
			issues = append(issues, "aggressive wanderers require level >= 10 and non-boss difficulty/tags")
		}
		for _, issue := range issues {
			b.add(Hit{RuleID: RuleMovement, Severity: "error", EntityType: kind, EntityID: id, Field: "movement." + issue, Message: issue, FixHint: "Use adjacent ordinary exits and allowed rooms; correct the movement block."})
		}
	}
	for id, n := range b.snap.NPCs {
		check("npc", id, n.EffectiveMovement(), n.SpawnRoomID, n)
	}
	for id, s := range b.snap.Spawners {
		n := b.snap.NPCs[s.TemplateID]
		cfg := s.Movement
		if cfg == nil && n != nil {
			cfg = n.EffectiveMovement()
		}
		check("spawner", id, cfg, s.RoomID, n)
	}
}
