package items

// ItemTemplate ...
type ItemTemplate struct {
	Item   `bson:",inline"`
	Script *string `bson:"script,omitempty" json:"script"`
}

//ItemTemplates type
type ItemTemplates []*ItemTemplate
