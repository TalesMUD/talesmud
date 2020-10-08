package dialogoues

// Dialogue ...
type Dialogue struct {
	ID      string               `bson:"id,omitempty" json:"id,omitempty"`
	Text    string               `bson:"text,omitempty" json:"text,omitempty"`
	Options map[string]*Dialogue `bson:"options,omitempty" json:"options,omitempty"`
}

// FindDialogue ...
func (d *Dialogue) FindDialogue(id string) *Dialogue {

	if d.ID == id {
		return d
	}

	for _, v := range d.Options {
		if r := v.FindDialogue(id); r != nil {
			return r
		}
	}
	return nil
}
