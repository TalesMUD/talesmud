package worldmap

import (
	"embed"
	"encoding/json"
	"strings"
	"unicode"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

//go:embed map_terrain.json
var terrainFS embed.FS

type terrainRule struct {
	Terrain string
	Words   []string
}
type terrainConfig struct {
	Default     string
	Terrains    []string
	Rules       []terrainRule
	Areas       map[string]string
	GenericTags map[string]string
}

var terrainMapping = func() terrainConfig {
	b, err := terrainFS.ReadFile("map_terrain.json")
	if err != nil {
		panic(err)
	}
	var c terrainConfig
	if err := json.Unmarshal(b, &c); err != nil {
		panic(err)
	}
	if c.Default == "" || len(c.Terrains) == 0 {
		panic("invalid map terrain config")
	}
	return c
}()

func terrainText(s string) string {
	return strings.Join(strings.Fields(strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, s)), " ")
}
func matchTerrain(text string, exact bool) string {
	text = terrainText(text)
	for _, rule := range terrainMapping.Rules {
		for _, word := range rule.Words {
			word = terrainText(word)
			if exact && text == word || !exact && strings.Contains(" "+text+" ", " "+word+" ") {
				return rule.Terrain
			}
		}
	}
	return ""
}

// ClassifyTerrain returns the terrain and whether the documented default was needed.
// Authoritative rules and precedence are documented in map_terrain.json.
func ClassifyTerrain(r *rooms.Room) (string, bool) {
	if r == nil {
		return terrainMapping.Default, true
	}
	for _, v := range []string{r.RoomType, r.AreaType} {
		if t := matchTerrain(v, true); t != "" {
			return t, false
		}
	}
	// Ordered rules win independently of authored tag order.
	for _, rule := range terrainMapping.Rules {
		for _, tag := range r.Tags {
			for _, word := range rule.Words {
				if terrainText(tag) == terrainText(word) {
					return rule.Terrain, false
				}
			}
		}
	}
	if t := matchTerrain(r.Name, false); t != "" {
		return t, false
	}
	for _, tag := range []string{"underground", "indoor"} {
		if hasTag(r.Tags, tag) {
			return terrainMapping.GenericTags[tag], false
		}
	}
	if t := terrainMapping.Areas[r.Area]; t != "" {
		return t, false
	}
	if t := matchTerrain(r.Area, false); t != "" {
		return t, false
	}
	if t := matchTerrain(r.Description+" "+r.Detail, false); t != "" {
		return t, false
	}
	b := inferBiome(r.Area, r.Tags)
	for _, v := range []string{b, inferKind(r.Tags, r.CanBind, b)} {
		if t := matchTerrain(v, true); t != "" {
			return t, false
		}
	}
	if hasTag(r.Tags, "outdoor") {
		return terrainMapping.GenericTags["outdoor"], false
	}
	return terrainMapping.Default, true
}
