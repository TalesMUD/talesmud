package characters

import (
	"github.com/talesmud/talesmud/pkg/entities/items"
)

//Inventory data
type Inventory struct {
	Size  int32         `json:"size"`
	Items []*items.Item `json:"items"`
}
