// Package worldindex builds an in-memory reference graph for a world snapshot.
// The same snapshot can be filled from live repositories or from an import folder.
package worldindex

import (
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// Kind is an entity type in the reference graph.
type Kind string

const (
	KindRoom              Kind = "room"
	KindNPC               Kind = "npc"
	KindItem              Kind = "item"
	KindLootTable         Kind = "lootTable"
	KindSpawner           Kind = "spawner"
	KindDialog            Kind = "dialog"
	KindQuest             Kind = "quest"
	KindScript            Kind = "script"
	KindSkill             Kind = "skill"
	KindClass             Kind = "class"
	KindCharacterTemplate Kind = "characterTemplate"
)

// Ref is a typed entity id.
type Ref struct {
	Type Kind   `json:"type"`
	ID   string `json:"id"`
}

// Edge is one directed reference: from (type, id, field) to (type, id).
// How is a short label such as "exit north", "hidden exit", or "revealExit".
// ExitName is set for revealExit edges.
type Edge struct {
	FromType Kind   `json:"fromType"`
	FromID   string `json:"fromId"`
	Field    string `json:"field"`
	ToType   Kind   `json:"toType"`
	ToID     string `json:"toId"`
	How      string `json:"how"`
	ExitName string `json:"exitName,omitempty"`
}

// ClassInfo is the slice of a class kit the index needs. The catalog stays outside this package.
type ClassInfo struct {
	ID         string
	Name       string
	TemplateID string
	SkillIDs   []string
}

// Snapshot is the world the index is built from.
type Snapshot struct {
	StartRoomID        string
	ContentCommit      string
	Rooms              map[string]*rooms.Room
	NPCs               map[string]*npc.NPC
	Items              map[string]*items.Item
	LootTables         map[string]*items.LootTable
	Spawners           map[string]*npc.NPCSpawner
	Dialogs            map[string]*dialogs.Dialog
	Quests             map[string]*quests.Quest
	Scripts            map[string]*scripts.Script
	Skills             map[string]*skills.Skill
	CharacterTemplates map[string]*characters.CharacterTemplate
	Classes            []ClassInfo
}

// NewSnapshot returns an empty snapshot with initialized maps.
func NewSnapshot() Snapshot {
	return Snapshot{
		Rooms:              map[string]*rooms.Room{},
		NPCs:               map[string]*npc.NPC{},
		Items:              map[string]*items.Item{},
		LootTables:         map[string]*items.LootTable{},
		Spawners:           map[string]*npc.NPCSpawner{},
		Dialogs:            map[string]*dialogs.Dialog{},
		Quests:             map[string]*quests.Quest{},
		Scripts:            map[string]*scripts.Script{},
		Skills:             map[string]*skills.Skill{},
		CharacterTemplates: map[string]*characters.CharacterTemplate{},
	}
}
