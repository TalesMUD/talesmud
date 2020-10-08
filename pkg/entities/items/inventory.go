package items

//Inventory data
type Inventory struct {
	Size  int32   `json:"size"`
	Items []*Item `json:"items"`
}
