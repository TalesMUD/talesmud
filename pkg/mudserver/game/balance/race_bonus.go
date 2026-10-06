package balance

import (
	"math"
	"strings"
)

func canonicalRaceID(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	switch id {
	case "elve", "elves", "elf":
		return "elf"
	default:
		return id
	}
}

func isBluntWeapon(sub string) bool {
	switch strings.ToLower(strings.TrimSpace(sub)) {
	case "mace", "club", "hammer", "maul", "flail":
		return true
	default:
		return false
	}
}

func isBowWeapon(sub string) bool {
	return strings.ToLower(strings.TrimSpace(sub)) == "bow"
}

// RacialWeaponMultiplier is the one racial weapon bonus.
// Dwarf blunt +10%. Elf bow +10%. Human, construct, and any other race add nothing.
func RacialWeaponMultiplier(raceID, weaponSubType string) float64 {
	switch canonicalRaceID(raceID) {
	case "dwarf":
		if isBluntWeapon(weaponSubType) {
			return 1.10
		}
	case "elf":
		if isBowWeapon(weaponSubType) {
			return 1.10
		}
	}
	return 1
}

// ApplyRacialWeaponBonus applies RacialWeaponMultiplier once. Non-bonus races
// and the wrong weapon kind leave damage unchanged.
func ApplyRacialWeaponBonus(raceID, weaponSubType string, damage int32) int32 {
	if damage <= 0 {
		return damage
	}
	mult := RacialWeaponMultiplier(raceID, weaponSubType)
	if mult == 1 {
		return damage
	}
	out := int32(math.Round(float64(damage) * mult))
	if out < 1 {
		return 1
	}
	return out
}
