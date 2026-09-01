package game

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestDefeatRespawnsAtBoundRoomAndDamagesArmor(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R0106", nil)
	storeTestRoom(t, facade, "R0203", nil)

	jerkin := &items.Item{
		Entity:     &entities.Entity{ID: "armor-1"},
		Name:       "Leather Jerkin",
		Type:       items.ItemTypeArmor,
		Slot:       items.ItemSlotChest,
		Attributes: map[string]interface{}{"defense": float64(8)},
	}
	character, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-death"},
		Name:             "Faller",
		BelongsUser:      *traits.BelongsToUser("user-death"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R0106"},
		BoundRoomID:      "R0203",
		MaxHitPoints:     20,
		CurrentHitPoints: 1,
		XP:               100,
		Gold:             5,
		EquippedItems:    map[items.ItemSlot]*items.Item{items.ItemSlotChest: jerkin},
	})
	if err != nil {
		t.Fatal(err)
	}

	forest, _ := facade.RoomsService().FindByID("R0106")
	forest.AddCharacter(character.ID)
	_ = facade.RoomsService().Update("R0106", forest)

	instance := &combat.CombatInstance{
		ID:           "combat-death",
		OriginRoomID: "R0106",
		State:        combat.CombatStateDefeat,
		Players: []combat.CombatantRef{
			{ID: character.ID, Name: character.Name, CurrentHP: 0, MaxHP: 20, IsAlive: false},
		},
	}

	g.CombatController.processCombatDefeat(instance)

	stored, err := facade.CharactersService().FindByID(character.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CurrentRoomID != "R0203" {
		t.Fatalf("expected respawn at bound room R0203, got %s", stored.CurrentRoomID)
	}
	if stored.EquippedItems[items.ItemSlotChest] == nil {
		t.Fatal("armor slot should remain equipped")
	}
	if stored.EquippedItems[items.ItemSlotChest].Durability >= stored.EquippedItems[items.ItemSlotChest].MaxDurability {
		t.Fatal("expected armor durability loss on death")
	}
	if stored.CurrentHitPoints != 10 {
		t.Fatalf("expected 50%% HP after defeat, got %d", stored.CurrentHitPoints)
	}

	var sawDefeat, sawBoundRoom bool
	for _, out := range drainGameMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok {
			if strings.Contains(rsp.Message, "Your armor is battered") {
				sawDefeat = true
			}
			if strings.Contains(rsp.Message, "back at") {
				sawBoundRoom = true
			}
		}
	}
	if !sawDefeat {
		t.Fatal("expected defeat message mentioning battered armor")
	}
	if !sawBoundRoom {
		t.Fatal("expected defeat message mentioning bound-room respawn")
	}

	hub, _ := facade.RoomsService().FindByID("R0203")
	if hub == nil || !hub.IsCharacterInRoom(character.ID) {
		t.Fatal("character should be present in bound room after respawn")
	}
	forestAfter, _ := facade.RoomsService().FindByID("R0106")
	if forestAfter != nil && forestAfter.IsCharacterInRoom(character.ID) {
		t.Fatal("character should be removed from death room")
	}
}

func drainGameMessages(ch <-chan interface{}) []interface{} {
	var result []interface{}
	for {
		select {
		case msg := <-ch:
			result = append(result, msg)
		default:
			return result
		}
	}
}
