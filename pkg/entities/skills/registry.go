package skills

import (
	"strings"
	"sync"
)

// normalizeClassID maps class ID aliases to canonical IDs used in skill definitions.
// e.g. "wizard" → "mage" (the entity class is ClassWizard with ID "wizard")
func normalizeClassID(classID string) string {
	classLower := strings.ToLower(classID)
	if classLower == "wizard" {
		return "mage"
	}
	return classLower
}

var (
	skillCache []*Skill
	cacheMu    sync.RWMutex
)

// LoadFromDB populates the in-memory skill cache from database records.
// Called once at server startup by the SkillsService.
func LoadFromDB(skills []*Skill) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	skillCache = skills
}

// RefreshCache replaces the cache with updated data.
// Called after any CRUD operation on skills.
func RefreshCache(skills []*Skill) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	skillCache = skills
}

// AllSkills returns all skill definitions
func AllSkills() []*Skill {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	result := make([]*Skill, len(skillCache))
	copy(result, skillCache)
	return result
}

// SkillByID returns the skill with the given ID, or nil
func SkillByID(id string) *Skill {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	for _, s := range skillCache {
		if s.Entity != nil && s.Entity.ID == id {
			return s
		}
	}
	return nil
}

// SkillByName returns the first skill matching the given name (case-insensitive), or nil
func SkillByName(name string) *Skill {
	nameLower := strings.ToLower(name)
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	for _, s := range skillCache {
		if strings.ToLower(s.Name) == nameLower {
			return s
		}
	}
	return nil
}

// SkillsForClass returns all skills a class can learn (regardless of level)
func SkillsForClass(classID string) []*Skill {
	classLower := normalizeClassID(classID)
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	var result []*Skill
	for _, s := range skillCache {
		if s.HasClass(classLower) {
			result = append(result, s)
		}
	}
	return result
}

// AvailableSkills returns skills unlocked at the given class and level
func AvailableSkills(classID string, level int32) []*Skill {
	classLower := normalizeClassID(classID)
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	var result []*Skill
	for _, s := range skillCache {
		if s.HasClass(classLower) && s.LevelRequired <= level {
			result = append(result, s)
		}
	}
	return result
}

// MaxSkillSlots returns the number of skill slots available for a class at a given level
func MaxSkillSlots(classID string, level int32) int {
	classLower := normalizeClassID(classID)

	switch classLower {
	case "mage", "cleric", "druid":
		// Casters: L1=2, L15=3, L30=4
		switch {
		case level >= 30:
			return 4
		case level >= 15:
			return 3
		default:
			return 2
		}
	default:
		// Warrior, Rogue, Ranger: L1=1, L10=2, L20=3, L30=4
		switch {
		case level >= 30:
			return 4
		case level >= 20:
			return 3
		case level >= 10:
			return 2
		default:
			return 1
		}
	}
}

// IsCasterClass returns true if the class uses mana
func IsCasterClass(classID string) bool {
	classLower := normalizeClassID(classID)
	return classLower == "mage" || classLower == "cleric" || classLower == "druid"
}
