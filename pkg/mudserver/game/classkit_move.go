package game

import (
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/skills"
)

// fillKitHotbars puts known class skills into empty hotbar slots and persists when it can.
func (c *CombatController) fillKitHotbars(players []*characters.Character) {
	if c == nil {
		return
	}
	for _, p := range players {
		if p == nil || p.Entity == nil || !skills.IsKitClass(p.Class.ID) {
			continue
		}
		filled := skills.FillHotbar(p.Class.ID, p.Level, p.EquippedSkills)
		if sameIDs(p.EquippedSkills, filled) {
			continue
		}
		p.EquippedSkills = filled
		if c.game == nil || c.game.Facade == nil || c.game.Facade.CharactersService() == nil {
			continue
		}
		_ = c.game.Facade.CharactersService().Modify(p.ID, func(ch *characters.Character) error {
			ch.EquippedSkills = append([]string(nil), filled...)
			return nil
		})
	}
}

func sameIDs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (c *CombatController) slipOneExit(instance *combat.CombatInstance, characterID string) {
	if c == nil || c.game == nil || c.game.Facade == nil || instance == nil {
		return
	}
	char, err := c.game.Facade.CharactersService().FindByID(characterID)
	if err != nil || char == nil {
		return
	}
	room, _ := c.game.Facade.RoomsService().FindByID(char.CurrentRoomID)
	exits := visibleExitTargets(room)
	if len(exits) == 0 {
		return
	}
	dest := exits[0]
	if _, ok := c.game.RelocateCharacter(char, char.BelongsUserID, dest); !ok {
		return
	}
	if p := instance.GetPlayerByID(characterID); p != nil {
		p.SlipTo = dest
		c.engine.UpdateCombatant(instance, p)
	}
}

func (c *CombatController) reelOneRoom(instance *combat.CombatInstance, targetID string) {
	if c == nil || c.game == nil || c.game.Facade == nil || instance == nil || targetID == "" {
		return
	}
	origin := instance.OriginRoomID
	if origin == "" {
		return
	}
	if char, err := c.game.Facade.CharactersService().FindByID(targetID); err == nil && char != nil {
		if char.CurrentRoomID == "" || char.CurrentRoomID == origin {
			return
		}
		if !roomsOneApart(c, char.CurrentRoomID, origin) {
			return
		}
		_, _ = c.game.RelocateCharacter(char, char.BelongsUserID, origin)
		return
	}
	if c.game.NPCManager == nil {
		return
	}
	inst := c.game.NPCManager.GetInstance(targetID)
	if inst == nil || inst.CurrentRoomID == "" || inst.CurrentRoomID == origin {
		return
	}
	if !roomsOneApart(c, inst.CurrentRoomID, origin) {
		return
	}
	c.game.NPCManager.MoveInstance(targetID, origin)
}

func roomsOneApart(c *CombatController, fromID, toID string) bool {
	if c == nil || c.game == nil || c.game.Facade == nil || fromID == "" || toID == "" {
		return false
	}
	from, _ := c.game.Facade.RoomsService().FindByID(fromID)
	to, _ := c.game.Facade.RoomsService().FindByID(toID)
	return roomExitsTo(from, toID) || roomExitsTo(to, fromID)
}

func roomExitsTo(room *rooms.Room, dest string) bool {
	if room == nil || room.Exits == nil || dest == "" {
		return false
	}
	for _, exit := range *room.Exits {
		if exit.Hidden || exit.Target == "" {
			continue
		}
		if exit.Target == dest {
			return true
		}
	}
	return false
}
