package service

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

func TestQuestXPBanksWhenTrainerMode(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	ruleset.SetLevelUpMode(ruleset.ModeTrainer)

	facade := newTestFacade(t)
	exits := rooms.Exits{}
	chars := rooms.Characters{}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity: &entities.Entity{ID: "R0001"}, Name: "Test Room", Area: "Z00_test",
		Exits: &exits, Characters: &chars,
	}); err != nil {
		t.Fatal(err)
	}
	need := leveling.GetXPRequired(2)
	character := &characters.Character{
		Entity:           &entities.Entity{ID: "char-trainer"},
		Name:             "Novice",
		BelongsUser:      *traits.BelongsToUser("user-1"),
		Level:            1,
		XP:               0,
		MaxHitPoints:     20,
		CurrentHitPoints: 20,
		Class:            characters.ClassWarrior,
		Attributes: []characters.Attribute{
			{Short: "STR", Value: 10},
			{Short: "DEX", Value: 10},
			{Short: "INT", Value: 10},
			{Short: "WIS", Value: 10},
			{Short: "STA", Value: 10},
		},
		Inventory: items.Inventory{Size: 10},
	}
	if _, err := facade.CharactersService().Store(character); err != nil {
		t.Fatal(err)
	}
	quest := &quests.Quest{
		Entity:      &entities.Entity{ID: "quest-trainer"},
		Name:        "Lesson",
		Description: "Enough XP to level.",
		Source:      quests.QuestSource{Type: "auto"},
		Objectives: []quests.Objective{{
			ID: "obj", Type: quests.ObjectiveVisit, Description: "go", TargetID: "R0001", Amount: 1,
		}},
		Rewards: quests.Reward{XP: need, Gold: 2},
	}
	if _, err := facade.QuestsService().Store(quest); err != nil {
		t.Fatal(err)
	}
	_, levelUp, err := facade.QuestsService().GrantQuestRewards(character.ID, quest.ID)
	if err != nil {
		t.Fatal(err)
	}
	if levelUp != nil {
		t.Fatalf("trainer mode leveled from quest XP: %+v", levelUp)
	}
	stored, err := facade.CharactersService().FindByID(character.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Level != 1 || stored.XP != need || stored.Gold != 2 {
		t.Fatalf("banked state level=%d xp=%d gold=%d", stored.Level, stored.XP, stored.Gold)
	}
	if got := leveling.ApplyPendingLevels(stored); got == nil || got.NewLevel != 2 {
		t.Fatalf("pending levels = %+v", got)
	}
}
