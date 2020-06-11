package rooms

import (
	"errors"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

// RoomActionType type
type RoomActionType string

const (
	RoomActionTypeResponse     RoomActionType = "response"
	RoomActionTypeRoomResponse RoomActionType = "room_response"
	RoomActionTypeScript       RoomActionType = "Script"
)

// Action ... action that can be invoked by a player in thi room
type Action struct {
	Name        string                 `bson:"name,omitempty" json:"name,omitempty"`
	Description string                 `bson:"description,omitempty" json:"description,omitempty"`
	Response    string                 `bson:"response,omitempty" json:"response,omitempty"`
	Type        RoomActionType         `bson:"type,omitempty" json:"type,omitempty"`
	Params      map[string]interface{} `bson:"params,omitempty" json:"params,omitempty"`
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
	Tags     []string `bson:"tags,omitempty" json:"tags"`

	Actions    Actions     `bson:"actions,omitempty" json:"actions"`
	Exits      Exits       `bson:"exits,omitempty" json:"exits"`
	Characters Characters  `bson:"characters,omitempty" json:"characters"`
	Items      items.Items `bson:"items,omitempty" json:"items"`

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
	} `bson:"meta,omitempty" meta:"coords,omitempty"`
}

//GetExit ...
func (room *Room) GetExit(exit string) (Exit, bool) {

	for _, e := range room.Exits {
		if e.Name == exit {
			return e, true
		}
	}
	return Exit{}, false
}

//IsCharacterInRoom ,,,
func (room *Room) IsCharacterInRoom(character string) bool {

	for _, c := range room.Characters {
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

	room.Characters = append(room.Characters, character)

	return nil
}

//RemoveCharacter ,,,
func (room *Room) RemoveCharacter(character string) error {

	if !room.IsCharacterInRoom(character) {
		return errors.New("Character is not room")
	}

	var charactersNew Characters

	// make sure to remove duplilcates if for some reason the slice was altered
	// by hand or via the databases
	for _, c := range room.Characters {
		if c != character {
			charactersNew = append(charactersNew, c)
		}
	}
	room.Characters = charactersNew
	return nil
}
