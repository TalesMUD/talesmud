package validation

import (
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service"
)

type WorldSnapshot struct {
	Rooms      map[string]*rooms.Room
	Items      map[string]*items.Item
	NPCs       map[string]*npc.NPC
	Dialogs    map[string]*dialogs.Dialog
	LootTables map[string]*items.LootTable
	Spawners   map[string]*npc.NPCSpawner
	Quests     map[string]*quests.Quest
	Scripts    map[string]*scripts.Script

	RoomIDs      map[string]bool
	ItemIDs      map[string]bool
	NPCIDs       map[string]bool
	DialogIDs    map[string]bool
	LootTableIDs map[string]bool
	SpawnerIDs   map[string]bool
	QuestIDs     map[string]bool
	ScriptIDs    map[string]bool
}

func NewWorldSnapshot() WorldSnapshot {
	return WorldSnapshot{
		Rooms:        map[string]*rooms.Room{},
		Items:        map[string]*items.Item{},
		NPCs:         map[string]*npc.NPC{},
		Dialogs:      map[string]*dialogs.Dialog{},
		LootTables:   map[string]*items.LootTable{},
		Spawners:     map[string]*npc.NPCSpawner{},
		Quests:       map[string]*quests.Quest{},
		Scripts:      map[string]*scripts.Script{},
		RoomIDs:      map[string]bool{},
		ItemIDs:      map[string]bool{},
		NPCIDs:       map[string]bool{},
		DialogIDs:    map[string]bool{},
		LootTableIDs: map[string]bool{},
		SpawnerIDs:   map[string]bool{},
		QuestIDs:     map[string]bool{},
		ScriptIDs:    map[string]bool{},
	}
}

func (s WorldSnapshot) HasRoom(id string) bool {
	return id == "" || s.RoomIDs[id]
}

func (s WorldSnapshot) HasItem(id string) bool {
	return id == "" || s.ItemIDs[id]
}

func (s WorldSnapshot) HasNPC(id string) bool {
	return id == "" || s.NPCIDs[id]
}

func (s WorldSnapshot) HasDialog(id string) bool {
	return id == "" || s.DialogIDs[id]
}

func (s WorldSnapshot) HasLootTable(id string) bool {
	return id == "" || s.LootTableIDs[id]
}

func (s WorldSnapshot) HasQuest(id string) bool {
	return id == "" || s.QuestIDs[id]
}

func (s WorldSnapshot) HasScript(id string) bool {
	return id == "" || s.ScriptIDs[id]
}

func BuildSnapshot(f service.Facade) (WorldSnapshot, error) {
	snapshot := NewWorldSnapshot()

	roomsList, err := f.RoomsService().FindAll()
	if err != nil {
		return snapshot, err
	}
	for _, room := range roomsList {
		snapshot.Rooms[room.ID] = room
		snapshot.RoomIDs[room.ID] = true
	}

	itemsList, err := f.ItemsService().FindAll(repository.ItemsQuery{})
	if err != nil {
		return snapshot, err
	}
	for _, item := range itemsList {
		snapshot.Items[item.ID] = item
		snapshot.ItemIDs[item.ID] = true
	}

	npcsList, err := f.NPCsService().FindAll()
	if err != nil {
		return snapshot, err
	}
	for _, n := range npcsList {
		snapshot.NPCs[n.ID] = n
		snapshot.NPCIDs[n.ID] = true
	}

	dialogsList, err := f.DialogsService().FindAll()
	if err != nil {
		return snapshot, err
	}
	for _, dialog := range dialogsList {
		snapshot.Dialogs[dialog.ID] = dialog
		snapshot.DialogIDs[dialog.ID] = true
	}

	lootTables, err := f.LootTablesService().FindAll()
	if err != nil {
		return snapshot, err
	}
	for _, table := range lootTables {
		snapshot.LootTables[table.ID] = table
		snapshot.LootTableIDs[table.ID] = true
	}

	spawners, err := f.NPCSpawnersService().FindAll()
	if err != nil {
		return snapshot, err
	}
	for _, spawner := range spawners {
		snapshot.Spawners[spawner.ID] = spawner
		snapshot.SpawnerIDs[spawner.ID] = true
	}

	questsList, err := f.QuestsService().FindAll()
	if err != nil {
		return snapshot, err
	}
	for _, quest := range questsList {
		snapshot.Quests[quest.ID] = quest
		snapshot.QuestIDs[quest.ID] = true
	}

	scriptsList, err := f.ScriptsService().FindAll()
	if err != nil {
		return snapshot, err
	}
	for _, script := range scriptsList {
		snapshot.Scripts[script.ID] = script
		snapshot.ScriptIDs[script.ID] = true
	}

	return snapshot, nil
}
