package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/portraits"
)

// CreateRoomDescription ...
func CreateRoomDescription(room *rooms.Room, user *entities.User, game def.GameCtrl) string {
	description := "\n[" + room.Name + "]\n"
	description += room.Description

	// Characters - only show online players
	if room != nil {
		onlinePlayers := game.GetRoomPlayers(room.ID, user.LastCharacter)
		onlineChars := make([]string, 0, len(onlinePlayers))
		for _, player := range onlinePlayers {
			name := player.CharacterName
			if player.IsYou {
				name += "(you)"
			}
			onlineChars = append(onlineChars, name)
		}
		if len(onlineChars) > 0 {
			description += "\n"
			charResult := "- In the area: "
			for i, name := range onlineChars {
				if i > 0 {
					charResult += ", "
				}
				charResult += name
			}
			description += charResult
		}
	}

	// NPCs in the room from the instance manager (spawned NPCs are in memory, not database)
	npcs := game.GetNPCInstanceManager().GetInstancesInRoom(room.ID)
	if len(npcs) > 0 {
		var enemies []string
		var friendlyNPCs []string

		// Build display names with numbers for duplicates
		displayNames := BuildNPCDisplayNames(npcs)

		for _, n := range npcs {
			displayName := displayNames[n.Entity.ID]
			if n.IsEnemy() {
				enemies = append(enemies, displayName)
			} else {
				friendlyNPCs = append(friendlyNPCs, displayName)
			}
		}

		if len(enemies) > 0 {
			description += "\n- Enemies: "
			for i, name := range enemies {
				if i > 0 {
					description += ", "
				}
				description += name
			}
		}

		if len(friendlyNPCs) > 0 {
			description += "\n- NPCs: "
			for i, name := range friendlyNPCs {
				if i > 0 {
					description += ", "
				}
				description += name
			}
		}
	}

	// Actions
	if room.Actions != nil && len(*room.Actions) > 0 {
		description += "\n- You can:"
		for _, action := range *room.Actions {
			if action.Description != "" {
				description += "\n  > " + action.Description + " [" + action.Name + "]"
			}
		}
		description += "\n"
	}

	// Exits. A room loaded without an exit list has a nil pointer.
	description += "\n"
	description += "- The visible exits are:\n"

	if room.Exits != nil {
		for _, exit := range *room.Exits {
			if !exit.Hidden {
				description += " + [" + exit.Name + "] " + exit.Description + "\n"
			}
		}
	}

	return description
}

// RoomCharacter represents player character data for frontend UI rendering
type RoomCharacter struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	IsYou bool   `json:"isYou"`
}

// GetRoomCharacters returns online player character data for frontend rendering.
// The requesting user's own character is always included.
func GetRoomCharacters(room *rooms.Room, user *entities.User, game def.GameCtrl) []RoomCharacter {
	if room == nil || user == nil {
		return []RoomCharacter{}
	}

	players := game.GetRoomPlayers(room.ID, user.LastCharacter)
	result := make([]RoomCharacter, 0, len(players))
	for _, player := range players {
		result = append(result, RoomCharacter{
			ID:    player.CharacterID,
			Name:  player.CharacterName,
			IsYou: player.IsYou,
		})
	}

	return result
}

// GetRoomPresenceCharacters returns online active player characters for room-wide
// presence broadcasts. The client marks its own character locally.
func GetRoomPresenceCharacters(room *rooms.Room, game def.GameCtrl) []RoomCharacter {
	if room.Characters == nil || len(*room.Characters) == 0 {
		return []RoomCharacter{}
	}

	result := make([]RoomCharacter, 0, len(*room.Characters))
	for _, charID := range *room.Characters {
		character, err := game.GetFacade().CharactersService().FindByID(charID)
		if err != nil {
			continue
		}
		charUser, err := game.GetFacade().UsersService().FindByID(character.BelongsUserID)
		if err != nil || !charUser.IsOnline || charUser.LastCharacter != character.ID {
			continue
		}
		result = append(result, RoomCharacter{
			ID:    character.ID,
			Name:  character.Name,
			IsYou: false,
		})
	}

	return result
}

// RoomItem represents item data for frontend UI rendering.
// Fields after copyOnPickup are additive. Empty values stay off the wire.
type RoomItem struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	NoPickup     bool              `json:"noPickup,omitempty"`
	CopyOnPickup bool              `json:"copyOnPickup,omitempty"`
	Type         items.ItemType    `json:"type,omitempty"`
	SubType      items.ItemSubType `json:"subType,omitempty"`
	Quality      items.ItemQuality `json:"quality,omitempty"`
	Unique       bool              `json:"unique,omitempty"`
	Quantity     int32             `json:"quantity,omitempty"`
	Effect       string            `json:"effect,omitempty"`
}

// roomItemPayload copies the display fields a pickup row needs.
// Effect is one line: the first examine FormatEffect line, or "Heals N HP"
// when the item only restores health.
func roomItemPayload(item *items.Item) RoomItem {
	if item == nil {
		return RoomItem{}
	}
	row := RoomItem{
		ID:           item.ID,
		Name:         item.Name,
		NoPickup:     item.NoPickup,
		CopyOnPickup: item.CopyOnPickup || item.IsRoomBlueprint(),
		Type:         item.Type,
		SubType:      item.SubType,
		Quality:      item.Quality,
		Unique:       item.Unique,
		Effect:       roomItemEffect(item),
	}
	if item.Quantity > 0 {
		row.Quantity = item.Quantity
	}
	return row
}

func roomItemEffect(item *items.Item) string {
	if item == nil {
		return ""
	}
	if effects := item.DisplayEffects(nil); len(effects) > 0 {
		if line := strings.TrimSpace(items.FormatEffect(effects[0])); line != "" {
			return line
		}
	}
	for _, key := range []string{"healthRestore", "healAmount"} {
		if n, ok := attrInt(item.Attributes, key); ok && n > 0 {
			return fmt.Sprintf("Heals %d HP", n)
		}
	}
	return ""
}

func attrInt(attrs map[string]interface{}, key string) (int32, bool) {
	if len(attrs) == 0 {
		return 0, false
	}
	val, ok := attrs[key]
	if !ok || val == nil {
		return 0, false
	}
	switch v := val.(type) {
	case int32:
		return v, true
	case int:
		return int32(v), true
	case int64:
		return int32(v), true
	case float64:
		return int32(v), true
	case float32:
		return int32(v), true
	case string:
		text := strings.TrimSpace(v)
		if n, err := strconv.Atoi(text); err == nil {
			return int32(n), true
		}
		if n, err := strconv.ParseFloat(text, 64); err == nil {
			return int32(n), true
		}
	}
	return 0, false
}

// GetRoomItems returns item data for frontend rendering.
// If char is non-nil, CopyOnPickup items already collected by this character are hidden.
func GetRoomItems(room *rooms.Room, game def.GameCtrl, char *characters.Character) []RoomItem {
	if room.Items == nil || len(*room.Items) == 0 {
		return []RoomItem{}
	}

	result := make([]RoomItem, 0, len(*room.Items))
	for _, itemID := range *room.Items {
		item, err := game.GetFacade().ItemsService().FindByID(itemID)
		if err != nil {
			continue
		}

		// Safety net: never show bound items on the ground
		if item.IsBound() {
			continue
		}

		// Hide room blueprints this character has already copied
		if item.IsRoomBlueprint() && char != nil {
			templateID := item.TemplateID
			if templateID == "" {
				templateID = item.ID
			}
			if char.HasCollectedCopyItem(templateID) {
				continue
			}
		}

		result = append(result, roomItemPayload(item))
	}

	return result
}

// RoomNPC represents NPC data for frontend UI rendering
type RoomNPC struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"displayName"`
	IsEnemy       bool   `json:"isEnemy"`
	IsMerchant    bool   `json:"isMerchant"`
	IsQuestGiver  bool   `json:"isQuestGiver"`
	HasDialog     bool   `json:"hasDialog"`
	HasIdleDialog bool   `json:"hasIdleDialog"`
	CurrentHP     int32  `json:"currentHp,omitempty"`
	MaxHP         int32  `json:"maxHp,omitempty"`
	Level         int32  `json:"level,omitempty"`
	AggroOnSight  bool   `json:"aggroOnSight,omitempty"`
	State         string `json:"state"`
	Portrait      string `json:"portrait,omitempty"`
	TemplateID    string `json:"templateId,omitempty"`
}

// GetRoomNPCs returns NPC data for frontend rendering
func GetRoomNPCs(room *rooms.Room, game def.GameCtrl) []RoomNPC {
	npcs := game.GetNPCInstanceManager().GetInstancesInRoom(room.ID)
	if len(npcs) == 0 {
		return []RoomNPC{}
	}

	displayNames := BuildNPCDisplayNames(npcs)
	result := make([]RoomNPC, 0, len(npcs))

	for _, n := range npcs {
		roomNPC := RoomNPC{
			ID:            n.Entity.ID,
			Name:          n.Name,
			DisplayName:   displayNames[n.Entity.ID],
			IsEnemy:       n.IsEnemy(),
			IsMerchant:    n.IsMerchant(),
			IsQuestGiver:  isQuestGiver(n, game),
			HasDialog:     n.HasDialog(),
			HasIdleDialog: n.HasIdleDialog(),
			CurrentHP:     n.CurrentHitPoints,
			MaxHP:         n.MaxHitPoints,
			Level:         n.Level,
			AggroOnSight:  n.EnemyTrait != nil && n.EnemyTrait.AggroOnSight,
			State:         n.State,
			Portrait:      portraits.ForNPC(n),
			TemplateID:    n.TemplateID,
		}
		result = append(result, roomNPC)
	}

	return result
}

func isQuestGiver(n *npc.NPC, game def.GameCtrl) bool {
	if n == nil || n.Entity == nil {
		return false
	}
	ids := []string{n.Entity.ID}
	if n.TemplateID != "" {
		ids = append(ids, n.TemplateID)
	}
	for _, id := range ids {
		quests, err := game.GetFacade().QuestsService().FindBySourceNPC(id)
		if err == nil && len(quests) > 0 {
			return true
		}
	}
	return false
}

// BuildNPCDisplayNames creates display names for NPCs, adding numbers when duplicates exist
// Returns a map of NPC ID -> display name (e.g., "Rat #1", "Rat #2" or just "Rat" if unique)
func BuildNPCDisplayNames(npcs []*npc.NPC) map[string]string {
	result := make(map[string]string)

	// Count NPCs by base name
	nameCounts := make(map[string]int)
	for _, n := range npcs {
		nameCounts[n.Name]++
	}

	// Track current index for each name
	nameIndex := make(map[string]int)

	// Assign display names
	for _, n := range npcs {
		baseName := n.Name
		if nameCounts[baseName] > 1 {
			// Multiple NPCs with same name - add number
			nameIndex[baseName]++
			result[n.Entity.ID] = baseName + "#" + strconv.Itoa(nameIndex[baseName])
		} else {
			// Unique name - no number needed
			result[n.Entity.ID] = baseName
		}
	}

	return result
}

// StripHiddenExits returns a shallow room copy whose Exits list omits hidden
// exits. Revealed exits should already have Hidden=false via RoomWithCharacterReveals.
func StripHiddenExits(room *rooms.Room) *rooms.Room {
	if room == nil || room.Exits == nil {
		return room
	}
	visible := make(rooms.Exits, 0, len(*room.Exits))
	for _, exit := range *room.Exits {
		if exit.Hidden {
			continue
		}
		visible = append(visible, exit)
	}
	if len(visible) == len(*room.Exits) {
		return room
	}
	roomCopy := *room
	roomCopy.Exits = &visible
	return &roomCopy
}

// RoomWithCharacterReveals returns a room view with character-specific hidden exits revealed.
// If the character has no reveals for this room, the original room pointer is returned (no copy).
// Otherwise a shallow copy with modified exits is returned.
func RoomWithCharacterReveals(room *rooms.Room, char *characters.Character) *rooms.Room {
	if char == nil || char.RevealedExits == nil || room.Exits == nil {
		return room
	}
	revealed, ok := char.RevealedExits[room.ID]
	if !ok || len(revealed) == 0 {
		return room
	}

	// Make a shallow copy of the room and its exits
	roomCopy := *room
	exitsCopy := make(rooms.Exits, len(*room.Exits))
	copy(exitsCopy, *room.Exits)

	for i := range exitsCopy {
		if exitsCopy[i].Hidden {
			for _, name := range revealed {
				if strings.EqualFold(exitsCopy[i].Name, name) {
					exitsCopy[i].Hidden = false
					break
				}
			}
		}
	}

	roomCopy.Exits = &exitsCopy
	return &roomCopy
}

// RemoveStringFromSlice ...
func RemoveStringFromSlice(slice []string, inst string) []string {

	for i, elem := range slice {
		if elem == inst {
			if i == len(slice)-1 {
				return slice[:i]
			}
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
