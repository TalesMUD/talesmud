package exporter

import (
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// Data is the canonical JSON shape for export/import.
type Data struct {
	Rooms         []*rooms.Room           `json:"rooms"`
	Items         []*items.Item           `json:"items"`
	ItemTemplates []*items.ItemTemplate   `json:"itemTemplates"`
	Characters    []*characters.Character `json:"characters"`
	Scripts       []*scripts.Script       `json:"scripts"`
	Users         []*e.User               `json:"users"`
	NPCs          []*npc.NPC              `json:"npcs"`
	Dialogs       []*dialogs.Dialog       `json:"dialogs"`
	Parties       []*e.Party              `json:"parties"`
}
