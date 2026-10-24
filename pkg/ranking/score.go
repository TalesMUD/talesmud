// Package ranking scores equipped gear and orders characters for a public board.
// Weights come from the caller. The engine default lives in pkg/ruleset.
package ranking

import (
	"sort"
	"strconv"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/items"
)

// Weights is the item-score formula. Quality keys are lowercase.
// A missing quality uses Unknown. StatKeys are the only attribute keys that count.
type Weights struct {
	Quality  map[string]int
	Unknown  int
	StatKeys []string
}

// DefaultWeights is the engine formula when the ruleset omits ranking.
// piece = qualityWeight * (1 + level) + statBudget.
// qualityWeight rises normal < magic < rare < legendary < mythic.
func DefaultWeights() Weights {
	return Weights{
		Unknown: 10,
		Quality: map[string]int{
			"normal":    10,
			"magic":     16,
			"rare":      24,
			"legendary": 36,
			"mythic":    52,
		},
		StatKeys: []string{
			"damage", "defense", "armor",
			"strength", "agility", "intelligence", "wisdom", "stamina",
			"health", "mana",
			"spellPower", "holyDamage", "naturePower",
			"critChance", "critical", "criticalChance",
		},
	}
}

// Clone returns a copy the caller can edit.
func (w Weights) Clone() Weights {
	out := Weights{
		Unknown:  w.Unknown,
		Quality:  map[string]int{},
		StatKeys: append([]string(nil), w.StatKeys...),
	}
	for key, value := range w.Quality {
		out.Quality[key] = value
	}
	return out
}

// Piece is one equipped item. An empty ID is broken and does not score.
type Piece struct {
	ID         string
	Level      int
	Quality    string
	Attributes map[string]any
}

// ScorePieces sums unique equipped pieces. Nil and duplicate IDs are skipped,
// so a two-hand weapon stored in both hand slots counts once.
func ScorePieces(pieces []Piece, w Weights) int {
	if len(pieces) == 0 {
		return 0
	}
	seen := map[string]bool{}
	total := 0
	for _, piece := range pieces {
		id := strings.TrimSpace(piece.ID)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		total += scorePiece(piece, w)
	}
	return total
}

func scorePiece(piece Piece, w Weights) int {
	level := piece.Level
	if level < 0 {
		level = 0
	}
	quality := strings.ToLower(strings.TrimSpace(piece.Quality))
	weight, ok := w.Quality[quality]
	if !ok || weight < 0 {
		weight = w.Unknown
	}
	if weight < 0 {
		weight = 0
	}
	return weight*(1+level) + statBudget(piece.Attributes, w.StatKeys)
}

func statBudget(attrs map[string]any, keys []string) int {
	if len(attrs) == 0 || len(keys) == 0 {
		return 0
	}
	total := 0
	for _, key := range keys {
		value, ok := positiveStat(attrs[key])
		if ok {
			total += value
		}
	}
	return total
}

func positiveStat(raw any) (int, bool) {
	switch v := raw.(type) {
	case int:
		if v > 0 {
			return v, true
		}
	case int32:
		if v > 0 {
			return int(v), true
		}
	case int64:
		if v > 0 {
			return int(v), true
		}
	case float64:
		if v > 0 {
			return int(v), true
		}
	case float32:
		if v > 0 {
			return int(v), true
		}
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			return 0, false
		}
		if n, err := strconv.Atoi(text); err == nil && n > 0 {
			return n, true
		}
		if n, err := strconv.ParseFloat(text, 64); err == nil && n > 0 {
			return int(n), true
		}
	}
	return 0, false
}

// Worn slots in display order. Anything else still equipped is appended by name.
var wornSlots = []items.ItemSlot{
	items.ItemSlotHead,
	items.ItemSlotChest,
	items.ItemSlotLegs,
	items.ItemSlotBoots,
	items.ItemSlotHands,
	items.ItemSlotWaist,
	items.ItemSlotNeck,
	items.ItemSlotRing1,
	items.ItemSlotRing2,
	items.ItemSlotMainHand,
	items.ItemSlotOffHand,
}

var notWorn = map[items.ItemSlot]bool{
	items.ItemSlotInventory: true,
	items.ItemSlotContainer: true,
	items.ItemSlotPurse:     true,
}

// PiecesFromEquipped reads the character's worn map. Nil items are dropped.
// The same two-hand item in both hand slots is returned twice. ScorePieces counts it once.
func PiecesFromEquipped(equipped map[items.ItemSlot]*items.Item) []Piece {
	if len(equipped) == 0 {
		return nil
	}
	seen := map[items.ItemSlot]bool{}
	var slots []items.ItemSlot
	for _, slot := range wornSlots {
		if _, ok := equipped[slot]; ok {
			slots = append(slots, slot)
			seen[slot] = true
		}
	}
	extra := make([]string, 0)
	for slot := range equipped {
		if seen[slot] || notWorn[slot] {
			continue
		}
		extra = append(extra, string(slot))
	}
	sort.Strings(extra)
	for _, slot := range extra {
		slots = append(slots, items.ItemSlot(slot))
	}
	out := make([]Piece, 0, len(slots))
	for _, slot := range slots {
		item := equipped[slot]
		if item == nil || item.Entity == nil {
			continue
		}
		out = append(out, Piece{
			ID:         item.ID,
			Level:      int(item.Level),
			Quality:    string(item.Quality),
			Attributes: item.Attributes,
		})
	}
	return out
}
