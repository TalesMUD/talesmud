package ranking

import (
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

func TestScoreEmptyGearIsZero(t *testing.T) {
	if got := ScorePieces(nil, DefaultWeights()); got != 0 {
		t.Fatalf("nil %d", got)
	}
	if got := ScorePieces([]Piece{}, DefaultWeights()); got != 0 {
		t.Fatalf("empty %d", got)
	}
	if got := ScorePieces(PiecesFromEquipped(nil), DefaultWeights()); got != 0 {
		t.Fatalf("no map %d", got)
	}
}

func TestScoreRarityIsMonotonic(t *testing.T) {
	w := DefaultWeights()
	order := []string{"normal", "magic", "rare", "legendary", "mythic"}
	prev := -1
	for _, quality := range order {
		got := ScorePieces([]Piece{{ID: "a", Level: 4, Quality: quality}}, w)
		if got <= prev {
			t.Fatalf("%s scored %d after %d", quality, got, prev)
		}
		prev = got
	}
}

func TestScoreTwoHandCountsOnce(t *testing.T) {
	w := DefaultWeights()
	one := ScorePieces([]Piece{{
		ID: "sword", Level: 5, Quality: "rare", Attributes: map[string]any{"damage": 8},
	}}, w)
	bothHands := ScorePieces([]Piece{
		{ID: "sword", Level: 5, Quality: "rare", Attributes: map[string]any{"damage": 8}},
		{ID: "sword", Level: 5, Quality: "rare", Attributes: map[string]any{"damage": 8}},
	}, w)
	if bothHands != one {
		t.Fatalf("two-hand %d one-hand %d", bothHands, one)
	}
	shield := ScorePieces([]Piece{
		{ID: "sword", Level: 5, Quality: "rare", Attributes: map[string]any{"damage": 8}},
		{ID: "shield", Level: 5, Quality: "normal", Attributes: map[string]any{"armor": 2}},
	}, w)
	if shield <= one {
		t.Fatalf("weapon and shield %d should beat the two-hand once %d", shield, one)
	}

	weapon := &items.Item{
		Entity: &entities.Entity{ID: "sword"}, Level: 5, Quality: items.ItemQualityRare,
		Attributes: map[string]any{"damage": int32(8)},
	}
	fromMap := ScorePieces(PiecesFromEquipped(map[items.ItemSlot]*items.Item{
		items.ItemSlotMainHand: weapon,
		items.ItemSlotOffHand:  weapon,
	}), w)
	if fromMap != one {
		t.Fatalf("equipped map %d want %d", fromMap, one)
	}
}

func TestScoreIgnoresBrokenItems(t *testing.T) {
	w := DefaultWeights()
	got := ScorePieces([]Piece{
		{},
		{ID: " ", Quality: "mythic", Level: 40},
		{ID: "ok", Quality: "not-a-quality", Level: 0, Attributes: map[string]any{
			"damage":     "fast",
			"useMessage": "glow",
			"defense":    -3,
			"armor":      "4",
			"notes":      true,
		}},
	}, w)
	// unknown quality 10 * (1+0) + armor 4
	if got != 14 {
		t.Fatalf("got %d", got)
	}
	if ScorePieces(PiecesFromEquipped(map[items.ItemSlot]*items.Item{
		items.ItemSlotHead:  nil,
		items.ItemSlotChest: &items.Item{Name: "ghost"},
	}), w) != 0 {
		t.Fatal("nil and id-less item should score 0")
	}
}

func TestScoreLevelAndStatBudget(t *testing.T) {
	w := DefaultWeights()
	low := ScorePieces([]Piece{{ID: "a", Level: 1, Quality: "normal", Attributes: map[string]any{"damage": 1.0}}}, w)
	high := ScorePieces([]Piece{{ID: "a", Level: 3, Quality: "normal", Attributes: map[string]any{"damage": float64(1)}}}, w)
	if high <= low {
		t.Fatalf("level %d vs %d", high, low)
	}
	// normal 10 * (1+1) + damage 1 = 21
	if low != 21 {
		t.Fatalf("low %d", low)
	}
}

func TestPlaceOrdersLevelScoreCreatedName(t *testing.T) {
	early := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	late := early.Add(time.Hour)
	rows := Place([]Entry{
		{ID: "b", Name: "Bea", Level: 5, ItemScore: 10, Created: late},
		{ID: "c", Name: "Ced", Level: 9, ItemScore: 0, Created: late},
		{ID: "a", Name: "ada", Level: 5, ItemScore: 40, Created: late},
		{ID: "d", Name: "Dee", Level: 5, ItemScore: 10, Created: early},
		{ID: "e", Name: "Eve", Level: 5, ItemScore: 10},
		{ID: "f", Name: "Bea", Level: 5, ItemScore: 10, Created: late},
	})
	want := []string{"Ced", "ada", "Dee", "Bea", "Bea", "Eve"}
	if len(rows) != len(want) {
		t.Fatalf("len %d", len(rows))
	}
	for i, name := range want {
		if rows[i].Name != name || rows[i].Rank != i+1 {
			t.Fatalf("row %d %#v", i, rows[i])
		}
	}
	// The earlier Bea id "b" sorts before id "f" when the name ties.
	if rows[3].ID != "b" || rows[4].ID != "f" {
		t.Fatalf("tie ids %s %s", rows[3].ID, rows[4].ID)
	}
}

func TestCollectSkipsGuestsAdminsAndMissingUsers(t *testing.T) {
	player := &entities.User{Entity: &entities.Entity{ID: "u-play"}, Role: entities.RolePlayer, IsOnline: false}
	guest := &entities.User{Entity: &entities.Entity{ID: "u-guest"}, Role: entities.RolePlayer, IsGuest: true}
	admin := &entities.User{Entity: &entities.Entity{ID: "u-admin"}, Role: entities.RoleAdmin}
	maker := &entities.User{Entity: &entities.Entity{ID: "u-maker"}, Role: entities.RoleCreator}
	chars := []*characters.Character{
		{Entity: &entities.Entity{ID: "offline"}, BelongsUser: traits.BelongsUser{BelongsUserID: "u-play"}, Name: "Offline", Level: 3},
		{Entity: &entities.Entity{ID: "ghost"}, BelongsUser: traits.BelongsUser{BelongsUserID: "u-guest"}, Name: "Ghost", Level: 99},
		{Entity: &entities.Entity{ID: "root"}, BelongsUser: traits.BelongsUser{BelongsUserID: "u-admin"}, Name: "Root", Level: 98},
		{Entity: &entities.Entity{ID: "maker"}, BelongsUser: traits.BelongsUser{BelongsUserID: "u-maker"}, Name: "Maker", Level: 2, Class: characters.Class{ID: "mage", Name: "Mage"}, Race: characters.Race{ID: "human", Name: "Human"}},
		{Entity: &entities.Entity{ID: "orphan"}, BelongsUser: traits.BelongsUser{BelongsUserID: "missing"}, Name: "Orphan", Level: 50},
		nil,
	}
	rows := Collect(chars, []*entities.User{player, guest, admin, maker, nil}, DefaultWeights())
	if len(rows) != 2 || rows[0].Name != "Offline" || rows[1].Name != "Maker" {
		t.Fatalf("%#v", rows)
	}
	if rows[1].Class != "Mage" || rows[1].Race != "Human" {
		t.Fatalf("labels %#v", rows[1])
	}
}

func TestFindSelfByIDThenName(t *testing.T) {
	rows := Place([]Entry{
		{ID: "1", Name: "Ada", Level: 2},
		{ID: "2", Name: "Bea", Level: 1},
	})
	if got := Find(rows, "2"); got == nil || got.Name != "Bea" || got.Rank != 2 {
		t.Fatalf("%#v", got)
	}
	if got := Find(rows, "ada"); got == nil || got.Rank != 1 {
		t.Fatalf("%#v", got)
	}
	if Find(rows, "") != nil || Find(rows, "nope") != nil {
		t.Fatal("expected no match")
	}
}

func TestFormatIncludesOwnRank(t *testing.T) {
	rows := Place([]Entry{{ID: "1", Name: "Ada", Class: "Warrior", Race: "Human", Level: 4, ItemScore: 21}})
	you := rows[0]
	text := Format(rows, &you, 10)
	if !strings.Contains(text, "1  Ada") || !strings.Contains(text, "Your rank: 1") {
		t.Fatalf("%s", text)
	}
	out := Format(nil, nil, 10)
	if !strings.Contains(out, "No ranked characters") || !strings.Contains(out, "not ranked") {
		t.Fatalf("%s", out)
	}
}
