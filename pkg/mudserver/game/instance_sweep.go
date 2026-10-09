package game

import (
	"errors"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/instances"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/worldmap"
)

// errSweepSkip tells Modify not to write a character who does not move.
var errSweepSkip = errors.New("instance sweep: skip")

// InstanceSweepStats is the one-line summary of a startup or idle sweep.
type InstanceSweepStats struct {
	RoomsDeleted    int
	NPCsDeleted     int
	SpawnersDeleted int
	CharactersMoved int
}

// SweepInstanceRooms deletes persisted instance copies and moves characters
// who were saved in one, or in a room that no longer exists.
// NPC copies are memory-only in normal play. A row whose id contains the
// instance marker is still removed, as is a spawner aimed at a copy.
// Ground items are stored on the room row and disappear with it.
// Call this once before the process accepts players. A second call is a no-op
// when the first one finished. There is no SIGTERM hook; the next start runs
// this again.
func (g *Game) SweepInstanceRooms() InstanceSweepStats {
	var stats InstanceSweepStats
	if g == nil || g.Facade == nil {
		return stats
	}
	all, err := loadRoomMap(g.Facade.RoomsService())
	if err != nil {
		log.WithError(err).Error("instance sweep: failed to list rooms")
		return stats
	}
	startID := service.ResolveStartRoomID(g.Facade.ServerSettingsService(), g.Facade.RoomsService())
	if !roomPresent(all, startID) {
		startID = ""
	}

	chars, err := g.Facade.CharactersService().FindAll()
	if err != nil {
		log.WithError(err).Error("instance sweep: failed to list characters")
	} else {
		for _, ch := range chars {
			if ch == nil || ch.ID == "" {
				continue
			}
			if g.moveSweptCharacter(all, ch, startID) {
				stats.CharactersMoved++
			}
		}
	}

	stats.NPCsDeleted = g.deleteCloneNPCs()
	stats.SpawnersDeleted = g.deleteCloneSpawners()
	stats.RoomsDeleted = g.deleteCloneRooms(all)

	log.WithFields(log.Fields{
		"roomsDeleted":    stats.RoomsDeleted,
		"npcsDeleted":     stats.NPCsDeleted,
		"spawnersDeleted": stats.SpawnersDeleted,
		"charactersMoved": stats.CharactersMoved,
	}).Info("instance room sweep")
	return stats
}

func (g *Game) moveSweptCharacter(all map[string]*rooms.Room, ch *characters.Character, startID string) bool {
	var roomName string
	var destID string
	err := g.Facade.CharactersService().Modify(ch.ID, func(stored *characters.Character) error {
		if stored == nil || !needsSweepMove(all, stored.CurrentRoomID) {
			return errSweepSkip
		}
		fallback := ""
		if roomPresent(all, stored.BoundRoomID) {
			fallback = stored.BoundRoomID
		}
		destID = instances.RelocationDest(all, stored.CurrentRoomID, startID, fallback)
		dest := all[destID]
		if dest == nil || destID == "" || destID == stored.CurrentRoomID {
			log.WithFields(log.Fields{
				"characterID": stored.ID,
				"roomID":      stored.CurrentRoomID,
			}).Warn("instance sweep: no room for character")
			return errSweepSkip
		}
		stored.CurrentRoomID = destID
		if g.CombatController == nil || !g.CombatController.IsPlayerInCombat(stored.ID) {
			stored.InCombat = false
			stored.CombatInstanceID = ""
		}
		worldmap.MarkOn(stored, dest)
		roomName = dest.Name
		if roomName == "" {
			roomName = dest.ID
		}
		return nil
	})
	if errors.Is(err, errSweepSkip) || err != nil {
		if err != nil && !errors.Is(err, errSweepSkip) {
			log.WithError(err).WithField("characterID", ch.ID).Warn("instance sweep: failed to move character")
		}
		return false
	}
	placeOffline(g.Facade.RoomsService(), all, ch.ID, destID)
	g.noteRelocation(ch.ID, roomName)
	ch.CurrentRoomID = destID
	return true
}

func needsSweepMove(all map[string]*rooms.Room, roomID string) bool {
	if roomID == "" {
		return false
	}
	if instances.IsCloneID(roomID) {
		return true
	}
	return all[roomID] == nil
}

func placeOffline(roomsSvc service.RoomsService, all map[string]*rooms.Room, charID, destID string) {
	room := all[destID]
	if room == nil || roomsSvc == nil || room.IsCharacterInRoom(charID) {
		return
	}
	if err := room.AddCharacter(charID); err != nil {
		return
	}
	if err := roomsSvc.Update(destID, room); err != nil {
		log.WithError(err).WithField("roomID", destID).Warn("instance sweep: failed to place character in room")
	}
}

func (g *Game) deleteCloneNPCs() int {
	if g.Facade.NPCsService() == nil {
		return 0
	}
	npcs, err := g.Facade.NPCsService().FindAll()
	if err != nil {
		log.WithError(err).Warn("instance sweep: failed to list npcs")
		return 0
	}
	n := 0
	for _, npcRow := range npcs {
		if npcRow == nil || npcRow.Entity == nil || !instances.IsCloneID(npcRow.ID) {
			continue
		}
		if err := g.Facade.NPCsService().Delete(npcRow.ID); err != nil {
			log.WithError(err).WithField("npcID", npcRow.ID).Warn("instance sweep: failed to delete npc copy")
			continue
		}
		n++
	}
	return n
}

func (g *Game) deleteCloneSpawners() int {
	if g.Facade.NPCSpawnersService() == nil {
		return 0
	}
	spawners, err := g.Facade.NPCSpawnersService().FindAll()
	if err != nil {
		log.WithError(err).Warn("instance sweep: failed to list spawners")
		return 0
	}
	n := 0
	for _, sp := range spawners {
		if sp == nil || sp.Entity == nil {
			continue
		}
		if !instances.IsCloneID(sp.ID) && !instances.IsCloneID(sp.RoomID) {
			continue
		}
		if err := g.Facade.NPCSpawnersService().Delete(sp.ID); err != nil {
			log.WithError(err).WithField("spawnerID", sp.ID).Warn("instance sweep: failed to delete spawner")
			continue
		}
		n++
	}
	return n
}

func (g *Game) deleteCloneRooms(all map[string]*rooms.Room) int {
	ids := make([]string, 0)
	for id := range all {
		if instances.IsCloneID(id) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	n := 0
	for _, id := range ids {
		if err := g.Facade.RoomsService().Delete(id); err != nil {
			log.WithError(err).WithField("roomID", id).Warn("instance sweep: failed to delete room")
			continue
		}
		delete(all, id)
		n++
	}
	return n
}

func loadRoomMap(roomsSvc service.RoomsService) (map[string]*rooms.Room, error) {
	if roomsSvc == nil {
		return nil, errors.New("no rooms service")
	}
	list, err := roomsSvc.FindAll()
	if err != nil {
		return nil, err
	}
	all := make(map[string]*rooms.Room, len(list))
	for _, room := range list {
		if room != nil && room.Entity != nil && room.ID != "" {
			all[room.ID] = room
		}
	}
	return all, nil
}

func roomPresent(all map[string]*rooms.Room, id string) bool {
	return id != "" && !instances.IsCloneID(id) && all[id] != nil
}

func (g *Game) noteRelocation(charID, roomName string) {
	if g == nil || charID == "" || roomName == "" {
		return
	}
	g.relocMu.Lock()
	defer g.relocMu.Unlock()
	if g.relocNotice == nil {
		g.relocNotice = map[string]string{}
	}
	g.relocNotice[charID] = fmt.Sprintf("You find yourself back at %s.", roomName)
}

// TakeRelocationNotice returns the one login line for charID and forgets it.
// A second call is empty.
func (g *Game) TakeRelocationNotice(charID string) string {
	if g == nil || charID == "" {
		return ""
	}
	g.relocMu.Lock()
	defer g.relocMu.Unlock()
	line := g.relocNotice[charID]
	delete(g.relocNotice, charID)
	return line
}
