package game

import (
	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/instances"
)

type roomInstanceAdapter struct {
	game *Game
	mgr  *instances.Manager
}

func (a *roomInstanceAdapter) Enter(characterID, hubID, destID string) (string, error) {
	fresh := a.mgr.ClonesForCharacter(characterID) == nil
	cloned, err := a.mgr.Enter(a.game.Facade.RoomsService(), characterID, hubID, destID)
	if err != nil {
		return "", err
	}
	if fresh {
		a.cloneNPCs(characterID)
	}
	return cloned, nil
}

func (a *roomInstanceAdapter) NoteLeave(characterID, fromRoomID, toRoomID string) {
	deleted := a.mgr.NoteLeave(a.game.Facade.RoomsService(), characterID, fromRoomID, toRoomID)
	a.dropNPCs(deleted)
}

func (a *roomInstanceAdapter) DestroyCharacterInstance(characterID string) {
	deleted := a.mgr.DestroyCharacterInstance(a.game.Facade.RoomsService(), characterID)
	a.dropNPCs(deleted)
}

func (a *roomInstanceAdapter) IsClone(roomID string) bool {
	return a.mgr.IsClone(roomID)
}

func (a *roomInstanceAdapter) cloneNPCs(characterID string) {
	clones := a.mgr.ClonesForCharacter(characterID)
	if clones == nil || a.game.NPCManager == nil {
		return
	}
	for templateRoom, cloneRoom := range clones {
		for _, n := range a.game.NPCManager.GetInstancesInRoom(templateRoom) {
			if n == nil || n.Entity == nil {
				continue
			}
			dup := *n
			dup.Entity = &entities.Entity{ID: n.ID + "~" + cloneRoom}
			dup.CurrentRoom = traits.CurrentRoom{CurrentRoomID: cloneRoom}
			dup.SpawnRoomID = cloneRoom
			if dup.TemplateID == "" {
				dup.TemplateID = n.ID
			}
			a.game.NPCManager.RegisterExistingNPC(&dup, cloneRoom)
		}
	}
}

func (a *roomInstanceAdapter) dropNPCs(cloneRooms []string) {
	if a.game.NPCManager == nil {
		return
	}
	for _, rid := range cloneRooms {
		for _, n := range a.game.NPCManager.GetInstancesInRoom(rid) {
			if n != nil && n.Entity != nil {
				a.game.NPCManager.UpdateInstance(n.ID, func(nn *npc.NPC) {
					nn.SpawnRoomID = ""
					nn.RespawnTime = 0
				})
				a.game.NPCManager.KillInstance(n.ID)
			}
		}
	}
}
