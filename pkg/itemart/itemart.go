package itemart

import (
	"path"
	"strings"
)

const PublicPath = "/api/item-art"

// FileName is the on-disk / URL filename for an item sprite.
func FileName(id, templateID string) string {
	key := strings.TrimSpace(templateID)
	if key == "" {
		key = strings.TrimSpace(id)
	}
	if i := strings.LastIndex(key, "~"); i > 0 {
		key = key[:i]
	}
	if key == "" {
		return ""
	}
	return key + ".png"
}

// URL is the guest-public item art URL for a specific template/instance.
func URL(id, templateID string) string {
	name := FileName(id, templateID)
	if name == "" {
		return ""
	}
	return path.Join(PublicPath, name)
}

// GenericKey maps item type/subtype to a generic sprite bucket.
func GenericKey(itemType, subType string) string {
	t := strings.ToLower(strings.TrimSpace(itemType))
	s := strings.ToLower(strings.TrimSpace(subType))
	if strings.Contains(s, "torch") {
		return "torch"
	}
	switch t {
	case "weapon":
		return "weapon"
	case "armor":
		return "armor"
	case "quest":
		return "quest"
	case "currency":
		return "currency"
	case "consumable":
		return "consumable"
	case "collectible", "crafting_material":
		return "junk"
	default:
		return "default"
	}
}

// GenericFileName is the filename for a type-based generic sprite.
func GenericFileName(itemType, subType string) string {
	return "generic-" + GenericKey(itemType, subType) + ".png"
}

// GenericURL is the guest-public URL for a generic type sprite.
func GenericURL(itemType, subType string) string {
	return path.Join(PublicPath, GenericFileName(itemType, subType))
}

// SkillGenericFile is the item-art filename for an equipped skill / spell bind.
func SkillGenericFile(skillID, skillName string) string {
	return SkillGenericKey(skillID, skillName) + ".png"
}

// SkillGenericURL is the guest-public URL for a skill hotbar icon.
func SkillGenericURL(skillID, skillName string) string {
	return path.Join(PublicPath, SkillGenericFile(skillID, skillName))
}

// SkillGenericKey maps a combat skill id or display name to a generic icon stem.
func SkillGenericKey(skillID, skillName string) string {
	id := strings.ToLower(strings.TrimSpace(skillID))
	name := strings.ToLower(strings.TrimSpace(skillName))
	key := id
	if key == "" {
		key = name
	}
	switch id {
	case "warrior_power_strike", "warrior_cleave", "rogue_backstab", "rogue_flurry":
		return "generic-action-melee"
	case "warrior_shield_bash", "ranger_pin_down":
		return "generic-spell-stun"
	case "warrior_battle_cry", "warrior_berserker_rage":
		return "generic-spell-strength"
	case "rogue_poison_strike":
		return "generic-spell-poison"
	case "rogue_evasion", "mage_mana_shield", "cleric_shield_of_faith", "druid_barkskin":
		return "generic-spell-shield"
	case "rogue_shadow_strike":
		return "generic-spell-curse"
	case "mage_fireball":
		return "generic-spell-fire"
	case "mage_frost_shield":
		return "generic-spell-ice"
	case "mage_lightning_bolt":
		return "generic-spell-lightning"
	case "mage_arcane_burst":
		return "generic-spell-arcane"
	case "cleric_heal", "cleric_divine_light", "ranger_natures_gift", "druid_rejuvenation":
		return "generic-spell-heal"
	case "cleric_holy_strike", "cleric_smite", "druid_starfire":
		return "generic-spell-holy"
	case "ranger_aimed_shot", "ranger_volley":
		return "generic-action-ranged"
	case "druid_wrath", "druid_entangle":
		return "generic-spell-nature"
	}
	blob := id + " " + name
	switch {
	case strings.Contains(blob, "fire") || strings.Contains(blob, "flame"):
		return "generic-spell-fire"
	case strings.Contains(blob, "frost") || strings.Contains(blob, "ice"):
		return "generic-spell-ice"
	case strings.Contains(blob, "lightning") || strings.Contains(blob, "bolt"):
		return "generic-spell-lightning"
	case strings.Contains(blob, "heal") || strings.Contains(blob, "rejuven") || strings.Contains(blob, "divine"):
		return "generic-spell-heal"
	case strings.Contains(blob, "shield") || strings.Contains(blob, "bark") || strings.Contains(blob, "evasion"):
		return "generic-spell-shield"
	case strings.Contains(blob, "poison"):
		return "generic-spell-poison"
	case strings.Contains(blob, "holy") || strings.Contains(blob, "smite") || strings.Contains(blob, "starfire"):
		return "generic-spell-holy"
	case strings.Contains(blob, "curse") || strings.Contains(blob, "shadow") || strings.Contains(blob, "dark"):
		return "generic-spell-curse"
	case strings.Contains(blob, "rage") || strings.Contains(blob, "cry") || strings.Contains(blob, "strength"):
		return "generic-spell-strength"
	case strings.Contains(blob, "stun") || strings.Contains(blob, "sleep") || strings.Contains(blob, "bash") || strings.Contains(blob, "pin"):
		return "generic-spell-stun"
	case strings.Contains(blob, "arcane"):
		return "generic-spell-arcane"
	case strings.Contains(blob, "wrath") || strings.Contains(blob, "entangle") || strings.Contains(blob, "nature"):
		return "generic-spell-nature"
	case strings.Contains(blob, "shot") || strings.Contains(blob, "volley") || strings.Contains(blob, "bow"):
		return "generic-action-ranged"
	case strings.Contains(blob, "strike") || strings.Contains(blob, "cleave") || strings.Contains(blob, "slash"):
		return "generic-action-melee"
	default:
		return "generic-spell-arcane"
	}
}

// ActionGenericKey maps a bindable action id to a generic icon stem.
func ActionGenericKey(actionID string) string {
	switch strings.ToLower(strings.TrimSpace(actionID)) {
	case "melee", "attack":
		return "generic-action-melee"
	case "look":
		return "generic-action-look"
	case "rest":
		return "generic-action-rest"
	case "flee", "run", "escape":
		return "generic-action-flee"
	case "search", "examine":
		return "generic-action-search"
	case "talk", "use", "say":
		return "generic-action-talk"
	case "ranged", "shoot":
		return "generic-action-ranged"
	default:
		return "generic-default"
	}
}
