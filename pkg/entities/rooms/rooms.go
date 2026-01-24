package rooms

import (
	"errors"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

// RoomActionType type
type RoomActionType string

const (
	RoomActionTypeResponse     RoomActionType = "response"
	RoomActionTypeRoomResponse RoomActionType = "room_response"
	RoomActionTypeScript       RoomActionType = "script"
)

// Action ... action that can be invoked by a player in thi room
type Action struct {
	Name        string                 `bson:"name,omitempty" json:"name,omitempty"`
	Description string                 `bson:"description,omitempty" json:"description,omitempty"`
	Response    string                 `bson:"response,omitempty" json:"response,omitempty"`
	Type        RoomActionType         `bson:"type,omitempty" json:"type,omitempty"`
	Params      map[string]interface{} `bson:"params,omitempty" json:"params"`
}

//Actions type
type Actions []Action

// RoomExitType type
type RoomExitType string

const (
	RoomExitTypeNormal    RoomExitType = "normal"
	RoomExitTypeDirection              = "direction"
	RoomExitTypeTeleport               = "teleport"
)

// Exit ... action that can be invoked by a player in thi room
type Exit struct {
	Name        string                 `bson:"name,omitempty" json:"name,omitempty"`
	Description string                 `bson:"description,omitempty" json:"description,omitempty"`
	Type        RoomExitType           `bson:"type,omitempty" json:"type,omitempty"`
	Hidden      bool                   `bson:"hidden,omitempty" json:"hidden,omitempty"`
	Target      string                 `bson:"target,omitempty" json:"target,omitempty"`
	Params      map[string]interface{} `bson:"params,omitempty" json:"params"`
}

//Exits type
type Exits []Exit

//Characters type
type Characters []string

//Items type
type Items []string

//NPCs type
type NPCs []string

//Room data
type Room struct {
	*entities.Entity `bson:",inline"`
	traits.LookAt    `bson:",inline"` // provides detail

	Name        string `bson:"name,omitempty" json:"name"`
	Description string `bson:"description,omitempty" json:"description"`
	//Detail      string `bson:"detail,omitempty" json:"detail"`
	RoomType string `bson:"roomType,omitempty" json:"roomType"`

	Area     string   `bson:"area,omitempty" json:"area"`
	AreaType string   `bson:"areaType,omitempty" json:"areaType"`
	Tags     []string `bson:"tags" json:"tags"`

	// Scripting hooks
	OnEnterScriptID string `bson:"onEnterScriptID,omitempty" json:"onEnterScriptID,omitempty"`

	Actions *Actions `bson:"actions,omitempty" json:"actions"`
	Exits   *Exits   `bson:"exits,omitempty" json:"exits"`

	// live data
	Items      *Items      `bson:"items,omitempty" json:"items"`
	Characters *Characters `bson:"characters,omitempty" json:"characters"`
	NPCs       *NPCs       `bson:"npcs,omitempty" json:"npcs"`

	// can be optionally used for MUDs that want to be grid based or need stricter maps
	Coords *struct {
		X int32 `bson:"x" json:"x"`
		Y int32 `bson:"y" json:"y"`
		Z int32 `bson:"z" json:"z"`
	} `bson:"coords,omitempty" json:"coords,omitempty"`

	// additional non game critical meta information to enhance player experience on client
	Meta *struct {
		// supply a mood to the client (optional)
		Mood string `bson:"mood,omitempty" json:"mood,omitempty"`
		// supply a background image id to the client (optional)
		Background string `bson:"background,omitempty" json:"background,omitempty"`
	} `bson:"meta,omitempty" json:"meta,omitempty"`

	// CanBind indicates if players can use /bind in this room to set their respawn point
	// Typically true for inns, temples, safe houses
	CanBind bool `bson:"canBind" json:"canBind"`
}

//Rooms type
type Rooms []*Room

//GetExit ...
func (room *Room) GetExit(exit string) (Exit, bool) {

	for _, e := range *room.Exits {
		if e.Name == exit {
			return e, true
		}
	}
	return Exit{}, false
}

//IsCharacterInRoom ,,,
func (room *Room) IsCharacterInRoom(character string) bool {

	// only check if there are characters in the room
	if room.Characters == nil || len(*room.Characters) == 0 {
		return false
	}

	for _, c := range *room.Characters {
		if c == character {
			return true
		}
	}
	return false
}

//AddCharacter ,,,
func (room *Room) AddCharacter(character string) error {

	if room.IsCharacterInRoom(character) {
		return errors.New("Character already in room")
	}

	// make sure room.Characters is not nil
	if room.Characters == nil {
		room.Characters = &Characters{}
	}

	modified := append(*room.Characters, character)
	room.Characters = &modified

	return nil
}

//RemoveCharacter ,,,
func (room *Room) RemoveCharacter(character string) error {

	if !room.IsCharacterInRoom(character) {
		return errors.New("Character is not room")
	}

	charactersNew := make(Characters, 0)

	// make sure to remove duplicates if for some reason the slice was altered
	// by hand or via the databases
	for _, c := range *room.Characters {
		if c != character {
			charactersNew = append(charactersNew, c)
		}
	}

	room.Characters = &charactersNew
	return nil
}

//IsItemInRoom checks if an item is in the room
func (room *Room) IsItemInRoom(itemID string) bool {
	if room.Items == nil || len(*room.Items) == 0 {
		return false
	}

	for _, id := range *room.Items {
		if id == itemID {
			return true
		}
	}
	return false
}

//AddItem adds an item ID to the room
func (room *Room) AddItem(itemID string) error {
	if room.IsItemInRoom(itemID) {
		return errors.New("Item already in room")
	}

	if room.Items == nil {
		room.Items = &Items{}
	}

	modified := append(*room.Items, itemID)
	room.Items = &modified

	return nil
}

//RemoveItem removes an item ID from the room
func (room *Room) RemoveItem(itemID string) error {
	if !room.IsItemInRoom(itemID) {
		return errors.New("Item is not in room")
	}

	itemsNew := make(Items, 0)

	for _, id := range *room.Items {
		if id != itemID {
			itemsNew = append(itemsNew, id)
		}
	}

	room.Items = &itemsNew
	return nil
}

//GetItemIDs returns a copy of the item IDs in the room
func (room *Room) GetItemIDs() []string {
	if room.Items == nil || len(*room.Items) == 0 {
		return []string{}
	}

	result := make([]string, len(*room.Items))
	copy(result, *room.Items)
	return result
}

//GetNPCIDs returns a copy of the NPC IDs in the room (residents)
func (room *Room) GetNPCIDs() []string {
	if room.NPCs == nil || len(*room.NPCs) == 0 {
		return []string{}
	}

	result := make([]string, len(*room.NPCs))
	copy(result, *room.NPCs)
	return result
}
