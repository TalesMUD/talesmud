package characters

import (
	"github.com/atla/owndnd/pkg/entities/items"
)

//Inventory data
type Inventory struct {
	Size  int32         `json:"size"`
	Items []*items.Item `json:"items"`
}
