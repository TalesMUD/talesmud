package contenthealth

import (
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/classkit"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service/validation"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

// World is the content snapshot a health run checks.
// Runtime copies are left out: room ids containing "~", NPC instances, and item instances.
type World struct {
	StartRoomID        string
	ContentCommit      string
	Rooms              []*rooms.Room
	NPCs               []*npc.NPC
	Items              []*items.Item
	LootTables         []*items.LootTable
	Spawners           []*npc.NPCSpawner
	Dialogs            []*dialogs.Dialog
	Quests             []*quests.Quest
	Scripts            []*scripts.Script
	Skills             []*skills.Skill
	CharacterTemplates []*characters.CharacterTemplate
	Classes            []worldindex.ClassInfo
}

// CatalogClasses reads the class kit currently loaded in this process.
func CatalogClasses() []worldindex.ClassInfo {
	var out []worldindex.ClassInfo
	for _, def := range classkit.Playable() {
		if def == nil {
			continue
		}
		info := worldindex.ClassInfo{ID: def.ID, Name: def.Name}
		if def.Template != nil {
			info.TemplateID = def.Template.ID
		}
		for _, skill := range def.Skills {
			if skill.ID != "" {
				info.SkillIDs = append(info.SkillIDs, skill.ID)
			}
		}
		out = append(out, info)
	}
	return out
}

func (w World) indexSnapshot() worldindex.Snapshot {
	snap := worldindex.NewSnapshot()
	snap.StartRoomID = w.StartRoomID
	snap.ContentCommit = w.ContentCommit
	snap.Classes = w.Classes
	for _, room := range w.Rooms {
		if id := entityID(roomEntity(room)); id != "" && !isRuntimeRoom(id) {
			snap.Rooms[id] = room
		}
	}
	for _, n := range w.NPCs {
		if n == nil || isRuntimeNPC(n) {
			continue
		}
		if id := entityID(n.Entity); id != "" {
			snap.NPCs[id] = n
		}
	}
	for _, item := range w.Items {
		if item == nil || !item.IsTemplate {
			continue
		}
		if id := entityID(item.Entity); id != "" {
			snap.Items[id] = item
		}
	}
	for _, table := range w.LootTables {
		if id := entityID(lootEntity(table)); id != "" {
			snap.LootTables[id] = table
		}
	}
	for _, spawner := range w.Spawners {
		if id := entityID(spawnerEntity(spawner)); id != "" {
			snap.Spawners[id] = spawner
		}
	}
	for _, dialog := range w.Dialogs {
		if id := entityID(dialogEntity(dialog)); id != "" {
			snap.Dialogs[id] = dialog
		}
	}
	for _, quest := range w.Quests {
		if id := entityID(questEntity(quest)); id != "" {
			snap.Quests[id] = quest
		}
	}
	for _, script := range w.Scripts {
		if id := entityID(scriptEntity(script)); id != "" {
			snap.Scripts[id] = script
		}
	}
	for _, skill := range w.Skills {
		if id := entityID(skillEntity(skill)); id != "" {
			snap.Skills[id] = skill
		}
	}
	for _, tmpl := range w.CharacterTemplates {
		if id := entityID(templateEntity(tmpl)); id != "" {
			snap.CharacterTemplates[id] = tmpl
		}
	}
	if snap.StartRoomID == "" || snap.Rooms[snap.StartRoomID] == nil {
		if snap.Rooms["R0001"] != nil {
			snap.StartRoomID = "R0001"
		}
	}
	return snap
}

func (w World) validationSnapshot(snap worldindex.Snapshot) validation.WorldSnapshot {
	out := validation.NewWorldSnapshot()
	for id, room := range snap.Rooms {
		out.Rooms[id] = room
		out.RoomIDs[id] = true
	}
	for id, item := range snap.Items {
		out.Items[id] = item
		out.ItemIDs[id] = true
	}
	for id, n := range snap.NPCs {
		out.NPCs[id] = n
		out.NPCIDs[id] = true
	}
	for id, dialog := range snap.Dialogs {
		out.Dialogs[id] = dialog
		out.DialogIDs[id] = true
	}
	for id, table := range snap.LootTables {
		out.LootTables[id] = table
		out.LootTableIDs[id] = true
	}
	for id, spawner := range snap.Spawners {
		out.Spawners[id] = spawner
		out.SpawnerIDs[id] = true
	}
	for id, quest := range snap.Quests {
		out.Quests[id] = quest
		out.QuestIDs[id] = true
	}
	for id, script := range snap.Scripts {
		out.Scripts[id] = script
		out.ScriptIDs[id] = true
	}
	return out
}

func isRuntimeRoom(id string) bool {
	return strings.Contains(id, "~")
}

func isRuntimeNPC(n *npc.NPC) bool {
	if n == nil || n.Entity == nil {
		return true
	}
	if n.TemplateID != "" || strings.Contains(n.ID, "~") {
		return true
	}
	return false
}

func entityID(e *entities.Entity) string {
	if e == nil {
		return ""
	}
	return e.ID
}

func roomEntity(room *rooms.Room) *entities.Entity {
	if room == nil {
		return nil
	}
	return room.Entity
}

func lootEntity(table *items.LootTable) *entities.Entity {
	if table == nil {
		return nil
	}
	return table.Entity
}

func spawnerEntity(spawner *npc.NPCSpawner) *entities.Entity {
	if spawner == nil {
		return nil
	}
	return spawner.Entity
}

func dialogEntity(dialog *dialogs.Dialog) *entities.Entity {
	if dialog == nil {
		return nil
	}
	return dialog.Entity
}

func questEntity(quest *quests.Quest) *entities.Entity {
	if quest == nil {
		return nil
	}
	return quest.Entity
}

func scriptEntity(script *scripts.Script) *entities.Entity {
	if script == nil {
		return nil
	}
	return script.Entity
}

func skillEntity(skill *skills.Skill) *entities.Entity {
	if skill == nil {
		return nil
	}
	return skill.Entity
}

func templateEntity(tmpl *characters.CharacterTemplate) *entities.Entity {
	if tmpl == nil {
		return nil
	}
	return tmpl.Entity
}

func sortedIDs[T any](m map[string]T) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
