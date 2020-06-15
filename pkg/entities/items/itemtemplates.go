package items

import "github.com/talesmud/talesmud/pkg/scripts"

// ItemTemplate ...
type ItemTemplate struct {
	Item          `bson:",inline"`
	OnAfterCreate *scripts.Script `bson:"script,omitempty" json:"script"`
}
