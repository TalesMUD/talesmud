package templates

import "github.com/atla/owndnd/pkg/entities/items"

//ItemTemplatePropertyType type
type ItemTemplatePropertyType int

//ItemTemplateAttributeType type
type ItemTemplateAttributeType int

//ItemTemplateProperties type
type ItemTemplateProperties []ItemTemplateProperty

//ItemTemplateAttributes type
type ItemTemplateAttributes []ItemTemplateAttribute

const (
	itemTemplatePropertyTypeString = iota + 1
	itemTemplatePropertyTypeInteger
	itemTemplatePropertyTypeDouble
)
const (
	itemTemplateAttributeTypeString = iota + 1
	itemTemplateAttributeTypeInteger
	itemTemplateAttributeTypeDouble
)

//ItemTemplateProperty data
type ItemTemplateProperty struct {
	//	ID    primitive.ObjectID       `bson:"_id,omitempty" json:"id,omitempty"`
	Key   string                   `yaml:"key" bson:"key" json:"key,omitempty"`
	Value string                   `yaml:"value" bson:"value" json:"value,omitempty"`
	Type  ItemTemplatePropertyType `yaml:"type" bson:"type,omitempty" json:"type,omitempty"`
}

//ItemTemplateAttribute data
type ItemTemplateAttribute struct {
	//ID    primitive.ObjectID       `bson:"_id,omitempty" json:"id,omitempty"`
	Key   string                   `yaml:"key" bson:"key" json:"key,omitempty"`
	Value string                   `yaml:"value" bson:"value" json:"value,omitempty"`
	Type  ItemTemplatePropertyType `yaml:"type" bson:"type,omitempty" json:"type,omitempty"`
}

type T struct {
	A string
	B struct {
		RenamedC int   `yaml:"c"`
		D        []int `yaml:",flow"`
	}
}

//ItemTemplate data
type ItemTemplate struct {
	//ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TemplateID  string         `yaml:"templateID" json:"templateID"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ItemType    items.ItemType `yaml:"itemType" json:"itemType"`
	ItemSlot    items.ItemSlot `yaml:"itemSlot" json:"itemSlot"`

	Properties map[string]string `yaml:",flow" json:"properties"`
	Attributes map[string]string `yaml:",flow" json:"attributes"`

	// General properties of the item template (interpreted during creation, effects?)
	//Properties map[string]string `yaml:"properties,omitempty" json:"properties"`

	// Generic attributes of the item created (copied over from template)
	//Attributes map[string]interface{} `yaml:"attributes,omitempty" json:"attributes"`
}

//ItemTemplates type
type ItemTemplates []*ItemTemplate
