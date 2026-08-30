package worldmap

import (
	"strings"
	"unicode"
)

func hasTag(tags []string, want string) bool {
	want = strings.ToLower(want)
	for _, t := range tags {
		if strings.EqualFold(t, want) {
			return true
		}
	}
	return false
}

func areaLooksLike(area, needle string) bool {
	return strings.Contains(strings.ToLower(area), strings.ToLower(needle))
}

func inferBiome(area string, tags []string) string {
	switch {
	case hasTag(tags, "water"), hasTag(tags, "creek"), areaLooksLike(area, "sewer"):
		return "water"
	case hasTag(tags, "forest"), areaLooksLike(area, "forest"):
		return "forest"
	case hasTag(tags, "underground"), hasTag(tags, "cave"), hasTag(tags, "dungeon"),
		hasTag(tags, "tutorial"), areaLooksLike(area, "catacomb"):
		return "dungeon"
	case hasTag(tags, "road"), hasTag(tags, "gate"), hasTag(tags, "town"),
		hasTag(tags, "inn"), hasTag(tags, "guards"), areaLooksLike(area, "oldtown"),
		areaLooksLike(area, "town"):
		return "settlement"
	case hasTag(tags, "outdoor"), areaLooksLike(area, "meadow"):
		return "meadow"
	default:
		return "wild"
	}
}

func inferKind(tags []string, canBind bool, biome string) string {
	switch {
	case hasTag(tags, "landmark"), hasTag(tags, "entry_point"), hasTag(tags, "gate"),
		hasTag(tags, "starting_room"), hasTag(tags, "inn"), canBind:
		return "landmark"
	case biome == "water":
		return "water"
	case biome == "dungeon":
		return "dungeon"
	case biome == "settlement", hasTag(tags, "road"):
		return "settlement"
	default:
		return "wild"
	}
}

func isLandmark(tags []string, canBind bool) bool {
	return hasTag(tags, "landmark") || hasTag(tags, "entry_point") ||
		hasTag(tags, "gate") || hasTag(tags, "starting_room") ||
		hasTag(tags, "inn") || canBind
}

func pathKind(dir string, fromTags, toTags []string, hidden bool) string {
	if hidden {
		return "hidden"
	}
	if isVertical(dir) {
		return "stair"
	}
	if hasTag(fromTags, "road") || hasTag(toTags, "road") ||
		hasTag(fromTags, "gate") || hasTag(toTags, "gate") {
		return "road"
	}
	if isCompass(dir) {
		return "trail"
	}
	return "passage"
}

func displayArea(area string) string {
	if area == "" {
		return "Unknown"
	}
	s := area
	if len(s) >= 3 && (s[0] == 'Z' || s[0] == 'z') && unicode.IsDigit(rune(s[1])) {
		if i := strings.Index(s, "_"); i > 0 {
			s = s[i+1:]
		}
	}
	s = strings.ReplaceAll(s, "_", " ")
	parts := strings.Fields(s)
	for i, p := range parts {
		runes := []rune(strings.ToLower(p))
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToTitle(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}

func layerID(z int) string {
	switch {
	case z < 0:
		return "lower"
	case z > 0:
		return "upper"
	default:
		return "overworld"
	}
}

func layerName(id string) string {
	switch id {
	case "lower":
		return "Lower"
	case "upper":
		return "Upper"
	default:
		return "Overworld"
	}
}

func layerKind(id string) string {
	return id
}
