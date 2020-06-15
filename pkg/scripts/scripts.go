package scripts

// Script ...
type Script struct {
	ID          string `bson:"id,omitempty" json:"id"`
	Name        string `bson:"name,omitempty" json:"name"`
	Description string `bson:"description,omitempty" json:"description"`
	Code        string `bson:"code,omitempty" json:"code"`
}