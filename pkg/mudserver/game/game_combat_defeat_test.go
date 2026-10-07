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
	"github.com/talesmud/talesmud/pkg/ruleset"
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

	var sawDefeat, sawBoundRoom, sawOutcome bool
	for _, out := range drainGameMessages(g.SendMessage()) {
		switch msg := out.(type) {
		case *messages.CombatEndMessage:
			if msg.Outcome != "defeat" {
				t.Fatalf("expected outcome defeat, got %q", msg.Outcome)
			}
			sawOutcome = true
			if strings.Contains(msg.Message, "Your armor is battered") {
				sawDefeat = true
			}
			if strings.Contains(msg.Message, "back at") {
				sawBoundRoom = true
			}
			if msg.Defeat == nil {
				t.Fatal("defeat payload missing")
			}
			if msg.Defeat.XPLost != 10 || msg.Defeat.GoldLost != 1 {
				t.Fatalf("defeat losses %+v", msg.Defeat)
			}
			if msg.Defeat.RespawnRoomID != "R0203" || msg.Defeat.RespawnRoom == "" {
				t.Fatalf("respawn %+v", msg.Defeat)
			}
			if msg.Defeat.HP != 10 || msg.Defeat.MaxHP != 20 {
				t.Fatalf("awaken hp %+v", msg.Defeat)
			}
			if len(msg.Defeat.Armor) == 0 || !strings.Contains(msg.Defeat.Armor[0], "Leather Jerkin") {
				t.Fatalf("armor %+v", msg.Defeat.Armor)
			}
			if msg.Rewards != nil {
				t.Fatalf("defeat must not carry victory rewards: %+v", msg.Rewards)
			}
		case messages.MessageResponse:
			if strings.Contains(msg.Message, "Your armor is battered") {
				sawDefeat = true
			}
			if strings.Contains(msg.Message, "back at") {
				sawBoundRoom = true
			}
		}
	}
	if !sawOutcome {
		t.Fatal("expected machine-readable combatEnd with outcome:defeat")
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

func TestNextResetDeathDoesNotRelocate(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	ruleset.SetDeath(ruleset.DeathPolicy{
		XPLossPercent:    10,
		GoldLossFlat:     1,
		Respawn:          ruleset.RespawnNextReset,
		RespawnHPPercent: 0,
		DamageArmor:      false,
	})

	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "wild", nil)
	storeTestRoom(t, facade, "town", nil)
	character, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-wait"},
		Name:             "Waiter",
		BelongsUser:      *traits.BelongsToUser("user-wait"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "wild"},
		BoundRoomID:      "town",
		MaxHitPoints:     20,
		CurrentHitPoints: 1,
		XP:               100,
		Gold:             5,
	})
	if err != nil {
		t.Fatal(err)
	}
	wild, _ := facade.RoomsService().FindByID("wild")
	wild.AddCharacter(character.ID)
	_ = facade.RoomsService().Update("wild", wild)

	g.CombatController.processCombatDefeat(&combat.CombatInstance{
		ID:           "combat-wait",
		OriginRoomID: "wild",
		State:        combat.CombatStateDefeat,
		Players: []combat.CombatantRef{
			{ID: character.ID, Name: character.Name, CurrentHP: 0, MaxHP: 20, IsAlive: false},
		},
	})

	stored, err := facade.CharactersService().FindByID(character.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CurrentRoomID != "wild" || !stored.AwaitingReset || stored.CurrentHitPoints != 0 {
		t.Fatalf("stay put: room=%s reset=%v hp=%d", stored.CurrentRoomID, stored.AwaitingReset, stored.CurrentHitPoints)
	}
	if stored.XP != 90 || stored.Gold != 4 {
		t.Fatalf("xp=%d gold=%d", stored.XP, stored.Gold)
	}
	still, _ := facade.RoomsService().FindByID("wild")
	if still == nil || !still.IsCharacterInRoom(character.ID) {
		t.Fatal("character left the death room")
	}
}

func TestFledPlayerKeepsHPGoldAndRoom(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)

	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "nest", nil)
	storeTestRoom(t, facade, "alley", nil)
	storeTestRoom(t, facade, "town", nil)

	fled, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-fled"},
		Name:             "Slipper",
		BelongsUser:      *traits.BelongsToUser("user-fled"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "alley"},
		BoundRoomID:      "town",
		MaxHitPoints:     21,
		CurrentHitPoints: 21,
		XP:               100,
		Gold:             5,
		InCombat:         true,
		CombatInstanceID: "combat-mixed",
	})
	if err != nil {
		t.Fatal(err)
	}
	dead, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-dead"},
		Name:             "Faller",
		BelongsUser:      *traits.BelongsToUser("user-dead"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "nest"},
		BoundRoomID:      "town",
		MaxHitPoints:     21,
		CurrentHitPoints: 21,
		XP:               100,
		Gold:             5,
		InCombat:         true,
		CombatInstanceID: "combat-mixed",
	})
	if err != nil {
		t.Fatal(err)
	}
	alley, _ := facade.RoomsService().FindByID("alley")
	alley.AddCharacter(fled.ID)
	_ = facade.RoomsService().Update("alley", alley)
	nest, _ := facade.RoomsService().FindByID("nest")
	nest.AddCharacter(dead.ID)
	_ = facade.RoomsService().Update("nest", nest)

	instance := &combat.CombatInstance{
		ID:           "combat-mixed",
		OriginRoomID: "nest",
		State:        combat.CombatStateDefeat,
		Players: []combat.CombatantRef{
			{ID: fled.ID, Name: fled.Name, CurrentHP: 21, MaxHP: 21, IsAlive: true, HasFled: true},
			{ID: dead.ID, Name: dead.Name, CurrentHP: 0, MaxHP: 21, IsAlive: false},
		},
	}

	// cleanup calls processCombatDefeat, then syncs HP and clears InCombat.
	g.CombatController.cleanupCombatInstance(instance, combat.CombatStateDefeat)

	storedFled, err := facade.CharactersService().FindByID(fled.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedFled.CurrentRoomID != "alley" || storedFled.CurrentHitPoints != 21 || storedFled.Gold != 5 || storedFled.XP != 100 {
		t.Fatalf("fled player changed: room=%s hp=%d gold=%d xp=%d", storedFled.CurrentRoomID, storedFled.CurrentHitPoints, storedFled.Gold, storedFled.XP)
	}
	if storedFled.InCombat || storedFled.CombatInstanceID != "" {
		t.Fatalf("fled combat flags: inCombat=%v instance=%q", storedFled.InCombat, storedFled.CombatInstanceID)
	}

	storedDead, err := facade.CharactersService().FindByID(dead.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedDead.CurrentRoomID != "town" || storedDead.CurrentHitPoints != 10 || storedDead.Gold != 4 || storedDead.XP != 90 {
		t.Fatalf("dead player: room=%s hp=%d gold=%d xp=%d", storedDead.CurrentRoomID, storedDead.CurrentHitPoints, storedDead.Gold, storedDead.XP)
	}
	if storedDead.InCombat || storedDead.CombatInstanceID != "" {
		t.Fatalf("dead combat flags: inCombat=%v instance=%q", storedDead.InCombat, storedDead.CombatInstanceID)
	}

	alleyAfter, _ := facade.RoomsService().FindByID("alley")
	if alleyAfter == nil || !alleyAfter.IsCharacterInRoom(fled.ID) {
		t.Fatal("fled player should stay in the slip room")
	}
	town, _ := facade.RoomsService().FindByID("town")
	if town == nil || !town.IsCharacterInRoom(dead.ID) || town.IsCharacterInRoom(fled.ID) {
		t.Fatal("only the dead player should be in the respawn room")
	}

	var fledEscaped, fledDefeat, deadDefeat bool
	for _, out := range drainGameMessages(g.SendMessage()) {
		msg, ok := out.(*messages.CombatEndMessage)
		if !ok {
			continue
		}
		switch msg.AudienceID {
		case "user-fled":
			if msg.Outcome == "defeat" || strings.Contains(msg.Message, "DEFEAT") || msg.Defeat != nil {
				fledDefeat = true
			}
			if msg.Outcome == "fled" && strings.Contains(msg.Message, "ESCAPED") {
				fledEscaped = true
			}
		case "user-dead":
			if msg.Outcome == "defeat" && strings.Contains(msg.Message, "DEFEAT") {
				deadDefeat = true
			}
		}
	}
	if !fledEscaped || fledDefeat {
		t.Fatalf("fled notices: escaped=%v defeat=%v", fledEscaped, fledDefeat)
	}
	if !deadDefeat {
		t.Fatal("dead player should still receive the defeat notice")
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
