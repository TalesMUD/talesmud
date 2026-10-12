package importer

import (
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// ConvertedWorld is an import folder turned into entities, without a database.
// Room-placed items get the same CopyOnPickup adjustment Import applies.
type ConvertedWorld struct {
	Rooms      []*rooms.Room
	NPCs       []*npc.NPC
	Items      []*items.Item
	LootTables []*items.LootTable
	Spawners   []*npc.NPCSpawner
	Dialogs    []*dialogs.Dialog
	Quests     []*quests.Quest
	Scripts    []*scripts.Script
	Skills     []*skills.Skill
}

// LoadConverted reads importPath/data and converts it the same way Import does.
// repos are not used. Class catalogs are left to the caller.
func LoadConverted(importPath string) (*ConvertedWorld, error) {
	w := New(nil, importPath)
	if err := w.validateImportFolder(); err != nil {
		return nil, err
	}
	yamlScripts, err := w.loadScripts()
	if err != nil {
		return nil, err
	}
	yamlItems, err := w.loadItems()
	if err != nil {
		return nil, err
	}
	yamlLoot, err := w.loadLootTables()
	if err != nil {
		return nil, err
	}
	yamlNPCs, err := w.loadNPCs()
	if err != nil {
		return nil, err
	}
	yamlDialogs, err := w.loadDialogs()
	if err != nil {
		return nil, err
	}
	yamlRooms, err := w.loadRooms()
	if err != nil {
		return nil, err
	}
	yamlSpawners, err := w.loadSpawners()
	if err != nil {
		return nil, err
	}
	yamlQuests, err := w.loadQuests()
	if err != nil {
		return nil, err
	}
	yamlSkills, err := w.loadSkills()
	if err != nil {
		return nil, err
	}

	out := &ConvertedWorld{}
	for _, script := range yamlScripts {
		out.Scripts = append(out.Scripts, script.ToEntity())
	}
	placed := roomPlacedItemIDs(yamlRooms)
	for _, item := range yamlItems {
		entity := item.ToEntity()
		if placed[item.ID] {
			entity.CopyOnPickup = true
		}
		out.Items = append(out.Items, entity)
	}
	for _, table := range yamlLoot {
		out.LootTables = append(out.LootTables, table.ToEntity())
	}
	for _, n := range yamlNPCs {
		out.NPCs = append(out.NPCs, n.ToEntity())
	}
	for _, dialog := range yamlDialogs {
		out.Dialogs = append(out.Dialogs, dialog.ToEntity())
	}
	for _, room := range yamlRooms {
		out.Rooms = append(out.Rooms, room.ToEntity())
	}
	for _, spawner := range yamlSpawners {
		out.Spawners = append(out.Spawners, spawner.ToEntity())
	}
	for _, quest := range yamlQuests {
		out.Quests = append(out.Quests, quest.ToEntity())
	}
	for _, skill := range yamlSkills {
		out.Skills = append(out.Skills, skill.ToEntity())
	}
	return out, nil
}
