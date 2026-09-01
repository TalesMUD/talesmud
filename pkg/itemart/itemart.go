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
