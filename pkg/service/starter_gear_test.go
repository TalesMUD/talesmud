package service

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/server/dto"
)

func TestCreateNewCharacterEquipsStartingItems(t *testing.T) {
	facade := newStarterGearFacade(t)

	for _, preset := range characters.SystemCharacterTemplatePresets() {
		if preset == nil || preset.Entity == nil {
			t.Fatal("preset missing")
		}
		if len(preset.StartingItems) == 0 {
			t.Fatalf("%s has no starting items", preset.ID)
		}
		for _, si := range preset.StartingItems {
			if si.Slot == "" || items.StarterItemTemplateByName(si.ItemTemplateName) == nil {
				t.Fatalf("%s starter %+v does not resolve", preset.ID, si)
			}
		}
	}

	created, err := facade.CharactersService().CreateNewCharacter(&dto.CreateCharacterDTO{
		TemplateID:  "tpl-ward",
		Name:        "WebWard",
		Description: "signed in",
		Race:        "human",
		UserID:      "user-web",
	})
	if err != nil {
		t.Fatal(err)
	}
	stored, err := facade.CharactersService().FindByID(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertStarterGear(t, stored, "Rusty Sword", "Leather Armor")
	if len(stored.Inventory.Items) != 0 {
		t.Fatalf("inventory %+v", stored.Inventory.Items)
	}

	ward := characters.PresetByID("tpl-ward")
	custom := &characters.CharacterTemplate{
		Entity:           &entities.Entity{ID: "tpl-db-only"},
		Name:             "Saved",
		Class:            ward.Class,
		Race:             ward.Race,
		Level:            1,
		CurrentHitPoints: ward.CurrentHitPoints,
		MaxHitPoints:     ward.MaxHitPoints,
		StartingItems: []characters.StartingItem{
			{Slot: items.ItemSlotMainHand, ItemTemplateName: "Rusty Sword", ItemTemplateID: "itm-sword"},
			{Slot: items.ItemSlotChest, ItemTemplateName: "Leather Armor", ItemTemplateID: "itm-leather"},
			{Slot: items.ItemSlotOffHand, ItemTemplateName: "Not A Real Blade"},
			{Slot: "", ItemTemplateName: "Worn Dagger"},
		},
	}
	if _, err := facade.CharacterTemplatesRepo().Import(custom); err != nil {
		t.Fatal(err)
	}
	fromDB, err := facade.CharactersService().CreateNewCharacter(&dto.CreateCharacterDTO{
		TemplateID: "tpl-db-only",
		Name:       "DbCreate",
		Race:       "human",
		UserID:     "user-web",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertStarterGear(t, fromDB, "Rusty Sword", "Leather Armor")
	if fromDB.EquippedItems[items.ItemSlotOffHand] != nil {
		t.Fatal("unknown starter name was equipped")
	}
	if fromDB.ID == created.ID {
		t.Fatal("db create reused the preset character")
	}
	if fromDB.EquippedItems[items.ItemSlotMainHand].ID == created.EquippedItems[items.ItemSlotMainHand].ID {
		t.Fatal("item id was reused across characters")
	}
}

func TestGuestCreateStillEquipsStartingItems(t *testing.T) {
	facade := newStarterGearFacade(t)
	token, err := facade.GuestService().CreateGuestSessionPick("127.0.0.1", "tpl-ward", "human")
	if err != nil {
		t.Fatal(err)
	}
	userID, err := facade.GuestService().ValidateGuestToken(token)
	if err != nil {
		t.Fatal(err)
	}
	user, err := facade.UsersService().FindByID(userID)
	if err != nil || user == nil || user.LastCharacter == "" {
		t.Fatalf("guest user %+v %v", user, err)
	}
	guest, err := facade.CharactersService().FindByID(user.LastCharacter)
	if err != nil {
		t.Fatal(err)
	}
	assertStarterGear(t, guest, "Rusty Sword", "Leather Armor")
	if guest.MaxLevelCap != GuestMaxLevel {
		t.Fatalf("guest cap %d", guest.MaxLevelCap)
	}
	if len(guest.Inventory.Items) != 0 {
		t.Fatalf("guest inventory %+v", guest.Inventory.Items)
	}

	web, err := facade.CharactersService().CreateNewCharacter(&dto.CreateCharacterDTO{
		TemplateID: "tpl-ward",
		Name:       "BesideGuest",
		Race:       "human",
		UserID:     "user-web",
	})
	if err != nil {
		t.Fatal(err)
	}
	if web.EquippedItems[items.ItemSlotMainHand].Name != guest.EquippedItems[items.ItemSlotMainHand].Name ||
		web.EquippedItems[items.ItemSlotChest].Name != guest.EquippedItems[items.ItemSlotChest].Name {
		t.Fatalf("guest %+v web %+v", guest.EquippedItems, web.EquippedItems)
	}
	if web.EquippedItems[items.ItemSlotMainHand].ID == guest.EquippedItems[items.ItemSlotMainHand].ID {
		t.Fatal("guest and web shared an item id")
	}
	if web.MaxLevelCap != 0 {
		t.Fatalf("web cap %d", web.MaxLevelCap)
	}
}

func TestEquipStartingItemsLeavesEmptyListAlone(t *testing.T) {
	ch := &characters.Character{}
	equipStartingItems(ch, nil)
	if ch.EquippedItems != nil {
		t.Fatalf("nil list allocated %+v", ch.EquippedItems)
	}
	equipStartingItems(ch, []characters.StartingItem{
		{Slot: items.ItemSlotMainHand, ItemTemplateName: "Missing Blade"},
	})
	if len(ch.EquippedItems) != 0 {
		t.Fatalf("unknown name equipped %+v", ch.EquippedItems)
	}
}

func newStarterGearFacade(t *testing.T) Facade {
	t.Helper()
	client, err := sqlite.Open(filepath.Join(t.TempDir(), "starter.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return NewFacade(repository.NewSQLiteFactory(client), nil)
}

func assertStarterGear(t *testing.T, ch *characters.Character, weapon, armor string) {
	t.Helper()
	if ch == nil || ch.EquippedItems == nil {
		t.Fatal("character has no equipped items")
	}
	w := ch.EquippedItems[items.ItemSlotMainHand]
	a := ch.EquippedItems[items.ItemSlotChest]
	if w == nil || a == nil || w.Name != weapon || a.Name != armor {
		t.Fatalf("gear weapon=%v armor=%v", w, a)
	}
	if w.IsTemplate || a.IsTemplate || w.ID == "" || a.ID == "" || w.ID == a.ID {
		t.Fatalf("copies template=%v/%v ids=%q %q", w.IsTemplate, a.IsTemplate, w.ID, a.ID)
	}
	srcW := items.StarterItemTemplateByName(weapon)
	srcA := items.StarterItemTemplateByName(armor)
	if srcW == nil || srcA == nil || !srcW.IsTemplate || !srcA.IsTemplate || srcW.Entity != nil || srcA.Entity != nil {
		t.Fatal("starter catalog template was changed")
	}
}
