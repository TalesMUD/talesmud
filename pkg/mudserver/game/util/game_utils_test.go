package util

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/service"
)

// stubRoomGame answers the two GameCtrl calls CreateRoomDescription makes.
// The embedded interfaces stay nil; unused methods are not called.
type stubRoomGame struct{ def.GameCtrl }

func (stubRoomGame) GetRoomPlayers(string, string) []def.OnlinePlayer { return nil }

func (stubRoomGame) GetNPCInstanceManager() def.NPCInstanceCtrl { return stubRoomNPCs{} }

type stubRoomNPCs struct{ def.NPCInstanceCtrl }

func (stubRoomNPCs) GetInstancesInRoom(string) []*npc.NPC { return nil }

func TestCreateRoomDescriptionNilExits(t *testing.T) {
	room := &rooms.Room{
		Entity:      &entities.Entity{ID: "room-nil-exits"},
		Name:        "Bare Cell",
		Description: "Stone and nothing else.",
	}
	got := CreateRoomDescription(room, &entities.User{LastCharacter: "char-1"}, stubRoomGame{})
	if !strings.Contains(got, "[Bare Cell]") || !strings.Contains(got, "Stone and nothing else.") {
		t.Fatalf("room text missing: %q", got)
	}
	if !strings.Contains(got, "- The visible exits are:") {
		t.Fatalf("exit header missing: %q", got)
	}
	if strings.Contains(got, " + [") {
		t.Fatalf("nil exits listed a direction: %q", got)
	}
}

func TestCreateRoomDescriptionListsVisibleExits(t *testing.T) {
	exits := rooms.Exits{
		{Name: "north", Description: "A corridor"},
		{Name: "secret", Description: "A crack", Hidden: true},
	}
	room := &rooms.Room{
		Entity:      &entities.Entity{ID: "room-exits"},
		Name:        "Landing",
		Description: "A landing.",
		Exits:       &exits,
	}
	got := CreateRoomDescription(room, &entities.User{}, stubRoomGame{})
	if !strings.Contains(got, " + [north] A corridor") {
		t.Fatalf("visible exit missing: %q", got)
	}
	if strings.Contains(got, "secret") || strings.Contains(got, "crack") {
		t.Fatalf("hidden exit listed: %q", got)
	}
}

func TestRemoveStringFromSliceRemovesLastElement(t *testing.T) {
	got := RemoveStringFromSlice([]string{"north", "south", "east"}, "east")
	want := []string{"north", "south"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}

func TestRemoveStringFromSliceRemovesOnlyMatch(t *testing.T) {
	got := RemoveStringFromSlice([]string{"north", "south", "east"}, "south")
	want := []string{"north", "east"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}

func TestRoomItemPayloadFieldsPresentWhenSet(t *testing.T) {
	row := roomItemPayload(&items.Item{
		Entity:       &entities.Entity{ID: "potion"},
		Name:         "Weak Health Potion",
		Type:         items.ItemTypeConsumable,
		SubType:      "potion",
		Quality:      items.ItemQualityRare,
		Unique:       true,
		Quantity:     3,
		NoPickup:     true,
		CopyOnPickup: true,
		Attributes:   map[string]interface{}{"healthRestore": 25},
	})
	data, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"id":           "potion",
		"name":         "Weak Health Potion",
		"noPickup":     true,
		"copyOnPickup": true,
		"type":         "consumable",
		"subType":      "potion",
		"quality":      "rare",
		"unique":       true,
		"quantity":     float64(3),
		"effect":       "Heals 25 HP",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("payload = %#v", got)
	}
}

func TestRoomItemPayloadUsesExamineEffectLine(t *testing.T) {
	row := roomItemPayload(&items.Item{
		Entity: &entities.Entity{ID: "blade"},
		Name:   "Ember Blade",
		Type:   items.ItemTypeWeapon,
		Effects: []items.ItemEffect{{
			Trigger: items.EffectTriggerOnHit,
			Name:    "Burn",
			Text:    "2 damage.",
		}},
	})
	if row.Effect != "On hit — Burn: 2 damage." {
		t.Fatalf("effect = %q", row.Effect)
	}
	data, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"effect":"On hit — Burn: 2 damage."`) {
		t.Fatalf("effect missing from json: %s", data)
	}
}

func TestRoomItemPayloadOmitsEmptyFields(t *testing.T) {
	dropped := roomItemPayload(&items.Item{
		Entity:         &entities.Entity{ID: "rock"},
		Name:           "Rock",
		TemplateID:     "ITM0099",
		InstanceSuffix: "abcd",
	})
	data, err := json.Marshal(dropped)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"id":"rock","name":"Rock"}` {
		t.Fatalf("empty payload = %s", data)
	}

	blueprint := roomItemPayload(&items.Item{
		Entity: &entities.Entity{ID: "torch"},
		Name:   "Dusty Torch",
	})
	raw, err := json.Marshal(blueprint)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"id":"torch","name":"Dusty Torch","copyOnPickup":true}` {
		t.Fatalf("blueprint payload = %s", raw)
	}
}

func TestGetRoomItemsKeepsOldFieldsAndSkipsBound(t *testing.T) {
	catalog := &items.Item{
		Entity: &entities.Entity{ID: "coin"},
		Name:   "Copper Bits",
		Type:   items.ItemTypeCurrency,
	}
	dropped := &items.Item{
		Entity:         &entities.Entity{ID: "loot"},
		Name:           "Rusty Dagger",
		TemplateID:     "ITM0009",
		InstanceSuffix: "abcd",
		Type:           items.ItemTypeWeapon,
	}
	bound := &items.Item{
		Entity:             &entities.Entity{ID: "bound"},
		Name:               "Leather Armor",
		BoundToCharacterID: "char-1",
	}
	ids := rooms.Items{"coin", "loot", "bound", "missing"}
	room := &rooms.Room{Items: &ids}
	game := itemLookupGame{byID: map[string]*items.Item{
		"coin": catalog, "loot": dropped, "bound": bound,
	}}

	got := GetRoomItems(room, game, nil)
	if len(got) != 2 {
		t.Fatalf("rows = %+v", got)
	}
	if got[0].ID != "coin" || got[0].Name != "Copper Bits" || got[0].Type != items.ItemTypeCurrency || !got[0].CopyOnPickup {
		t.Fatalf("currency row = %+v", got[0])
	}
	if got[1].ID != "loot" || got[1].CopyOnPickup || got[1].NoPickup {
		t.Fatalf("dropped loot must stay takeable: %+v", got[1])
	}
}

type itemLookupGame struct {
	def.GameCtrl
	byID map[string]*items.Item
}

func (g itemLookupGame) GetFacade() service.Facade {
	return itemLookupFacade{items: itemLookupService{byID: g.byID}}
}

type itemLookupFacade struct {
	service.Facade
	items service.ItemsService
}

func (f itemLookupFacade) ItemsService() service.ItemsService { return f.items }

type itemLookupService struct {
	service.ItemsService
	byID map[string]*items.Item
}

func (s itemLookupService) FindByID(id string) (*items.Item, error) {
	item, ok := s.byID[id]
	if !ok {
		return nil, errors.New("missing")
	}
	return item, nil
}
