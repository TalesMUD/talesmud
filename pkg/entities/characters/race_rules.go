package characters

import (
	"errors"
	"strings"
)

const HumanStartingGoldBonus int64 = 15

var (
	ErrRaceRequired        = errors.New("race is required")
	ErrRaceNotAllowed      = errors.New("race is not allowed for this class")
	ErrGuestPickIncomplete = errors.New("guest class and race must both be chosen")
)

// CanonicalRaceID maps the old elve id onto elf. Other ids pass through lowercased.
// New creates store the canonical id. Reads of saved elve rows still resolve as elf.
func CanonicalRaceID(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	switch id {
	case "elve", "elves", "elf":
		return "elf"
	default:
		return id
	}
}

// RaceByID resolves a race id, including the elve read alias.
func RaceByID(id string) (Race, bool) {
	switch CanonicalRaceID(id) {
	case "human":
		return RaceHuman, true
	case "dwarf":
		return RaceDwarf, true
	case "elf":
		return RaceElf, true
	case "construct":
		return RaceConstruct, true
	default:
		return Race{}, false
	}
}

func rosterKey(id string) string {
	switch strings.ToLower(strings.TrimSpace(id)) {
	case "warrior", "fenwatch", "tpl-fenwatch":
		return "warrior"
	case "rogue", "alley", "tpl-alley", "ranger", "hunter":
		return "rogue"
	case "wizard", "mage", "runehand", "rune_hand", "rune hand", "tpl-runehand":
		return "mage"
	case "ward", "tpl-ward", "hitch", "tpl-hitch":
		return "ward"
	case "rigger", "tpl-rigger":
		return "rigger"
	default:
		return ""
	}
}

// AllowedRaceIDs is the create allow-list for a class id or template id.
func AllowedRaceIDs(classOrTemplate string) []string {
	switch rosterKey(classOrTemplate) {
	case "warrior":
		return []string{"human", "dwarf"}
	case "rogue":
		return []string{"human", "dwarf", "elf"}
	case "mage":
		return []string{"human", "elf"}
	case "ward":
		return []string{"human", "dwarf"}
	case "rigger":
		return []string{"construct"}
	default:
		return nil
	}
}

// RaceAllowed reports whether raceID may be chosen for that class or template.
func RaceAllowed(classOrTemplate, raceID string) bool {
	raceID = CanonicalRaceID(raceID)
	if raceID == "" {
		return false
	}
	for _, id := range AllowedRaceIDs(classOrTemplate) {
		if id == raceID {
			return true
		}
	}
	return false
}

// ResolveCreateRace copies the client's race onto a new character.
// The template race is not kept when the client sent another allowed race.
// A missing race fails. elve is stored as elf.
func ResolveCreateRace(template *CharacterTemplate, raceID string) (Race, error) {
	if template == nil {
		return Race{}, ErrRaceNotAllowed
	}
	if strings.TrimSpace(raceID) == "" {
		return Race{}, ErrRaceRequired
	}
	race, ok := RaceByID(raceID)
	if !ok {
		return Race{}, ErrRaceNotAllowed
	}
	classID := ""
	templateID := ""
	if template.Class.ID != "" {
		classID = template.Class.ID
	}
	if template.Entity != nil {
		templateID = template.Entity.ID
	}
	if !RaceAllowed(classID, race.ID) && !RaceAllowed(templateID, race.ID) {
		return Race{}, ErrRaceNotAllowed
	}
	return race, nil
}

// StartingGold is the template baseline. Humans add 15. Other races do not.
func StartingGold(templateGold int64, raceID string) int64 {
	if CanonicalRaceID(raceID) == "human" {
		return templateGold + HumanStartingGoldBonus
	}
	return templateGold
}

// GuestPick is an optional guest class and race. Empty means random.
type GuestPick struct {
	Random   bool
	Template *CharacterTemplate
	Race     Race
}

// ResolveGuestPick accepts an empty pair (random) or a legal templateId and race.
// One field without the other, or an illegal pair, is rejected.
func ResolveGuestPick(templateID, raceID string) (GuestPick, error) {
	templateID = strings.TrimSpace(templateID)
	raceID = strings.TrimSpace(raceID)
	if templateID == "" && raceID == "" {
		return GuestPick{Random: true}, nil
	}
	if templateID == "" || raceID == "" {
		return GuestPick{}, ErrGuestPickIncomplete
	}
	template := PresetByID(templateID)
	if template == nil {
		return GuestPick{}, ErrRaceNotAllowed
	}
	race, err := ResolveCreateRace(template, raceID)
	if err != nil {
		return GuestPick{}, err
	}
	return GuestPick{Template: template, Race: race}, nil
}

// RandomAllowedRace picks one allow-list race for a class. n is rand.Intn-shaped.
func RandomAllowedRace(classOrTemplate string, n func(int) int) Race {
	ids := AllowedRaceIDs(classOrTemplate)
	if len(ids) == 0 {
		return RaceHuman
	}
	if n == nil {
		return RaceHuman
	}
	race, ok := RaceByID(ids[n(len(ids))])
	if !ok {
		return RaceHuman
	}
	return race
}
