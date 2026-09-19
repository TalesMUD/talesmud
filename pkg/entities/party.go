package entities

import (
	"time"
)

// MaxPartySize is the soft/hard cap for party membership (invite + accept).
const MaxPartySize = 5

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
	Name               string    `json:"name"`
	Created            time.Time `json:"created,omitempty"`
	Characters         []string  `json:"characters,omitempty"`
	LeaderCharacterID  string    `json:"leaderCharacterId,omitempty"`
}

// EnsureLeader sets LeaderCharacterID to the first member when empty/invalid.
func (p *Party) EnsureLeader() {
	if p == nil || len(p.Characters) == 0 {
		return
	}
	if p.LeaderCharacterID == "" {
		p.LeaderCharacterID = p.Characters[0]
		return
	}
	for _, id := range p.Characters {
		if id == p.LeaderCharacterID {
			return
		}
	}
	p.LeaderCharacterID = p.Characters[0]
}

// IsLeader reports whether characterID is the party leader.
func (p *Party) IsLeader(characterID string) bool {
	if p == nil || characterID == "" {
		return false
	}
	p.EnsureLeader()
	return p.LeaderCharacterID == characterID
}
