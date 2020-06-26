package scripts

import "github.com/talesmud/talesmud/pkg/entities"

//ScriptType type
type ScriptType string

//ScriptTypes ...
type ScriptTypes []ScriptType

const (
	ScriptTypeNone   ScriptType = "none"
	ScriptTypeCustom            = "custom"
	ScriptTypeItem              = "item"
	ScriptTypeRoom              = "rooom"
	ScriptTypeQuest             = "quest"
	ScriptTypeNPC               = "npc"
)

// Script ...
type Script struct {
	//ID          string `bson:"id,omitempty" json:"id"`
	*entities.Entity `bson:",inline"`

	Name        string     `bson:"name,omitempty" json:"name"`
	Description string     `bson:"description,omitempty" json:"description"`
	Code        string     `bson:"code,omitempty" json:"code"`
	Type        ScriptType `bson:"type,omitempty" json:"type"`
}
