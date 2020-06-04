package rooms

import "github.com/atla/owndnd/pkg/entities"

// RoomActionType type
type RoomActionType int8

const (
	RoomActionTypeDirection RoomActionType = iota + 1
)

func (it RoomActionType) String() string {
	return [...]string{"direction"}[it]
}

// RoomExitType type
type RoomExitType int8

const (
	RoomExitTypeNormal RoomExitType = iota + 1
)

func (it RoomExitType) String() string {
	return [...]string{"normal"}[it]
}

// Action ... action that can be invoked by a player in thi room
type Action struct {
	Name        string                 `bson:"name,omitempty" json:"name,omitempty"`
	Description string                 `bson:"description,omitempty" json:"description,omitempty"`
	Type        RoomActionType         `bson:"type,omitempty" json:"type,omitempty"`
	Params      map[string]interface{} `bson:"params,omitempty" json:"params"`
}

//Actions type
type Actions []Action

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

	Name        string     `bson:"name,omitempty" json:"name"`
	Description string     `bson:"description,omitempty" json:"description"`
	Detail      string     `bson:"detail,omitempty" json:"detail"`
	Actions     Actions    `bson:"actions,omitempty" json:"actions"`
	Exits       Exits      `bson:"exits,omitempty" json:"exits"`
	Characters  Characters `bson:"characters,omitempty" json:"characters"`
}
