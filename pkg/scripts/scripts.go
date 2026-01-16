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
	ScriptTypeRoom              = "room"
	ScriptTypeQuest             = "quest"
	ScriptTypeNPC               = "npc"
	ScriptTypeEvent             = "event"
)

// ScriptLanguage defines the scripting language used
type ScriptLanguage string

const (
	ScriptLanguageLua        ScriptLanguage = "lua"
	ScriptLanguageJavaScript ScriptLanguage = "javascript"
)

// Script ...
type Script struct {
	*entities.Entity `bson:",inline"`

	Name        string         `bson:"name,omitempty" json:"name"`
	Description string         `bson:"description,omitempty" json:"description"`
	Code        string         `bson:"code,omitempty" json:"code"`
	Type        ScriptType     `bson:"type,omitempty" json:"type"`
	Language    ScriptLanguage `bson:"language,omitempty" json:"language"`
}

// GetLanguage returns the script language, defaulting to JavaScript for backward compatibility
func (s *Script) GetLanguage() ScriptLanguage {
	if s.Language == "" {
		return ScriptLanguageJavaScript
	}
	return s.Language
}

// IsLua returns true if the script is written in Lua
func (s *Script) IsLua() bool {
	return s.GetLanguage() == ScriptLanguageLua
}
