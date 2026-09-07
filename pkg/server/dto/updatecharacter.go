package dto

// UpdateCharacterDTO is the allowlisted player-facing character update body.
// Combat stats, inventory, room, flags, and level caps are not client-writable.
type UpdateCharacterDTO struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
