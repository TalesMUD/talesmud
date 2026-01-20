package entities

import (
	"time"
)

// CharacterRace type
type CharacterRace int

const (
	crHuman CharacterRace = iota + 1
	crDwarf
	crElve
)

func (cr CharacterRace) String() string {
	return [...]string{"human", "dwarf", "elve"}[cr]
}

// Party data
type Party struct {
	*Entity
	Name       string    `json:"name"`
	Created    time.Time `json:"created,omitempty"`
	Characters []string  `json:"characters,omitempty"`
}
