package worldmap

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

type artRule struct {
	key   string
	words []string
}

var artRules = []artRule{
	{"forge", []string{"forge", "smithy", "blacksmith", "smithing", "smith"}},
	{"shrine", []string{"shrine", "temple", "chapel", "sanctuary"}},
	{"tavern", []string{"tavern", "inn", "taproom"}},
	{"shop", []string{"shop", "store", "merchant", "market"}},
	{"farm", []string{"farm", "farmstead", "farmland", "cropland", "wheat", "vineyard", "orchard"}},
	{"tower", []string{"guard", "guards", "guard post", "gate", "watchtower", "barracks"}},
	{"keep", []string{"keep", "castle", "citadel", "fortress"}},
	{"ruins", []string{"ruins", "ruined", "ruin", "broken walls"}},
	{"graveyard", []string{"graveyard", "cemetery", "headstones", "graves"}},
	{"dock", []string{"dock", "docks", "pier", "harbor", "harbour", "wharf"}},
	{"mine", []string{"mine", "mines", "adit", "mineshaft", "quarry"}},
	{"magic", []string{"magic", "magical", "arcane", "leyline", "ley line", "rune", "runes", "ritual", "portal", "standing stone"}},
	{"reeds", []string{"reeds", "marsh", "swamp", "bog"}},
	{"stump", []string{"stump", "stumps", "logging", "felled"}},
	{"flowers", []string{"flowers", "flower", "wildflower", "wildflowers", "garden"}},
}

func artMatch(text string, words []string) bool {
	text = " " + terrainText(text) + " "
	for _, word := range words {
		if strings.Contains(text, " "+terrainText(word)+" ") {
			return true
		}
	}
	return false
}

// ClassifyArt emits derived presentation hints, never descriptions or scripts.
// Tags/type/name/service actions win; description/detail supply a fallback.
func ClassifyArt(r *rooms.Room) (features []string, seed string) {
	if r == nil {
		return nil, ""
	}
	tags := append([]string(nil), r.Tags...)
	sort.Strings(tags)
	parts := []string{r.Name, r.RoomType, r.AreaType, strings.Join(tags, " ")}
	if r.Actions != nil {
		for _, a := range *r.Actions {
			parts = append(parts, a.Name)
		}
	}
	primary := strings.Join(parts, " ")
	for _, rule := range artRules {
		if artMatch(primary, rule.words) {
			features = append(features, rule.key)
		}
	}
	if len(features) == 0 {
		for _, rule := range artRules {
			if artMatch(r.Description+" "+r.Detail, rule.words) {
				features = append(features, rule.key)
				if len(features) >= 2 {
					break
				}
			}
		}
	}
	if len(features) == 0 && r.CanBind {
		features = append(features, "magic")
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(r.ID + "|" + r.Area + "|" + primary + "|" + r.Description + "|" + r.Detail))
	return features, fmt.Sprintf("%08x", h.Sum32())
}
func undergroundStyle(r *rooms.Room) string {
	primary := r.Name + " " + r.RoomType + " " + strings.Join(r.Tags, " ")
	if artMatch(primary, []string{"sewer", "sewers", "cistern", "sludge", "drain", "flooded"}) {
		return "sewer"
	}
	if artMatch(primary, []string{"crypt", "catacombs", "catacomb", "ossuary", "tomb", "reliquary", "burial"}) {
		return "crypt"
	}
	if artMatch(primary, []string{"cellar", "basement", "inn", "tavern", "vault", "ledger"}) {
		return "cellar"
	}
	return "cave"
}
