package service

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/repository"
)

func TestRepeatResetBlocksUntilDailyWindow(t *testing.T) {
	start := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
	quests.SetNow(func() time.Time { return start })
	t.Cleanup(func() { quests.SetNow(nil) })

	facade := newTestFacade(t)
	char, quest := storeKillQuest(t, facade, &quests.Quest{
		Name:        "Alley Contract",
		Description: "Clear the alley.",
		Repeatable:  true,
		RepeatReset: quests.RepeatResetDaily,
		Source:      quests.QuestSource{Type: "npc", NPCID: "clerk"},
		Objectives:  []quests.Objective{{ID: "kill", Type: quests.ObjectiveKill, Description: "Kill a thug", TargetID: "thug", Amount: 1}},
	})

	if _, err := facade.QuestsService().AcceptQuest(char.ID, quest.ID); err != nil {
		t.Fatalf("accept: %v", err)
	}
	finishQuest(t, facade, char.ID, quest.ID)
	if _, err := facade.QuestsService().AcceptQuest(char.ID, quest.ID); err == nil || !strings.Contains(err.Error(), "Available again in") {
		t.Fatalf("re-accept before reset: %v", err)
	}
	available, err := facade.QuestsService().GetAvailableQuests(char.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range available {
		if q.ID == quest.ID {
			t.Fatal("waiting quest was listed as available")
		}
	}

	quests.SetNow(func() time.Time { return time.Date(2026, 10, 18, 4, 0, 0, 0, time.UTC) })
	if _, err := facade.QuestsService().AcceptQuest(char.ID, quest.ID); err != nil {
		t.Fatalf("accept after daily reset: %v", err)
	}
}

func TestRepeatResetSurvivesRestart(t *testing.T) {
	start := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
	quests.SetNow(func() time.Time { return start })
	t.Cleanup(func() { quests.SetNow(nil) })

	dir := t.TempDir()
	path := filepath.Join(dir, "quests.db")
	open := func() Facade {
		t.Helper()
		client, err := sqlite.Open(path)
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		t.Cleanup(func() { _ = client.Close() })
		return NewFacade(repository.NewSQLiteFactory(client), nil)
	}

	facade := open()
	char, quest := storeKillQuest(t, facade, &quests.Quest{
		Name:        "Alley Contract",
		Description: "Clear the alley.",
		Repeatable:  true,
		RepeatReset: quests.RepeatResetDaily,
		Source:      quests.QuestSource{Type: "npc", NPCID: "clerk"},
		Objectives:  []quests.Objective{{ID: "kill", Type: quests.ObjectiveKill, Description: "Kill a thug", TargetID: "thug", Amount: 1}},
	})
	if _, err := facade.QuestsService().AcceptQuest(char.ID, quest.ID); err != nil {
		t.Fatal(err)
	}
	finishQuest(t, facade, char.ID, quest.ID)

	reopened := open()
	if _, err := reopened.QuestsService().AcceptQuest(char.ID, quest.ID); err == nil || !strings.Contains(err.Error(), "Available again in") {
		t.Fatalf("restart lost the window: %v", err)
	}
	logEntries, err := reopened.QuestsService().BuildQuestLog(char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(logEntries) != 1 || !logEntries[0].Repeats || logEntries[0].RepeatReset != quests.RepeatResetDaily || logEntries[0].NextAvailableAt == "" {
		t.Fatalf("log payload %+v", logEntries)
	}
}

func TestCooldownResetAndPlainRepeatable(t *testing.T) {
	start := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
	quests.SetNow(func() time.Time { return start })
	t.Cleanup(func() { quests.SetNow(nil) })

	facade := newTestFacade(t)
	char, cool := storeKillQuest(t, facade, &quests.Quest{
		Name:                  "Timed Contract",
		Description:           "Come back later.",
		Repeatable:            true,
		RepeatReset:           quests.RepeatResetCooldown,
		RepeatCooldownSeconds: 3600,
		Source:                quests.QuestSource{Type: "npc", NPCID: "clerk"},
		Objectives:            []quests.Objective{{ID: "kill", Type: quests.ObjectiveKill, Description: "Kill a thug", TargetID: "thug", Amount: 1}},
	})
	if _, err := facade.QuestsService().AcceptQuest(char.ID, cool.ID); err != nil {
		t.Fatal(err)
	}
	finishQuest(t, facade, char.ID, cool.ID)
	if _, err := facade.QuestsService().AcceptQuest(char.ID, cool.ID); err == nil {
		t.Fatal("cooldown accepted immediately")
	}
	quests.SetNow(func() time.Time { return start.Add(time.Hour) })
	if _, err := facade.QuestsService().AcceptQuest(char.ID, cool.ID); err != nil {
		t.Fatalf("accept after cooldown: %v", err)
	}

	_, plain := storeKillQuest(t, facade, &quests.Quest{
		Name:        "Open Contract",
		Description: "Again, right away.",
		Repeatable:  true,
		Source:      quests.QuestSource{Type: "npc", NPCID: "clerk"},
		Objectives:  []quests.Objective{{ID: "kill", Type: quests.ObjectiveKill, Description: "Kill a thug", TargetID: "thug", Amount: 1}},
	})
	if _, err := facade.QuestsService().AcceptQuest(char.ID, plain.ID); err != nil {
		t.Fatal(err)
	}
	finishQuest(t, facade, char.ID, plain.ID)
	if _, err := facade.QuestsService().AcceptQuest(char.ID, plain.ID); err != nil {
		t.Fatalf("plain repeatable blocked: %v", err)
	}
}

func TestNonRepeatableStillCompletesOnce(t *testing.T) {
	facade := newTestFacade(t)
	char, quest := storeKillQuest(t, facade, &quests.Quest{
		Name:        "Once",
		Description: "Only once.",
		Source:      quests.QuestSource{Type: "npc", NPCID: "clerk"},
		Objectives:  []quests.Objective{{ID: "kill", Type: quests.ObjectiveKill, Description: "Kill a thug", TargetID: "thug", Amount: 1}},
	})
	if _, err := facade.QuestsService().AcceptQuest(char.ID, quest.ID); err != nil {
		t.Fatal(err)
	}
	finishQuest(t, facade, char.ID, quest.ID)
	if _, err := facade.QuestsService().AcceptQuest(char.ID, quest.ID); err == nil || !strings.Contains(err.Error(), "already completed") {
		t.Fatalf("non-repeatable: %v", err)
	}
}

func storeKillQuest(t *testing.T, facade Facade, quest *quests.Quest) (*characters.Character, *quests.Quest) {
	t.Helper()
	if _, err := facade.NPCsService().Import(&npc.NPC{Entity: &entities.Entity{ID: "clerk"}, Name: "Clerk"}); err != nil && !strings.Contains(err.Error(), "UNIQUE") && !strings.Contains(err.Error(), "exists") {
		// A second quest in the same facade reuses the NPCs. Ignore a duplicate import.
		if existing, findErr := facade.NPCsService().FindByID("clerk"); findErr != nil || existing == nil {
			t.Fatalf("store clerk: %v", err)
		}
	}
	if _, err := facade.NPCsService().Import(&npc.NPC{Entity: &entities.Entity{ID: "thug"}, Name: "Thug"}); err != nil {
		if existing, findErr := facade.NPCsService().FindByID("thug"); findErr != nil || existing == nil {
			t.Fatalf("store thug: %v", err)
		}
	}
	char, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Wanderer",
		BelongsUser: *traits.BelongsToUser("user-1"),
	})
	if err != nil {
		// Same character may already exist when the helper is called twice. Find one.
		all, listErr := facade.CharactersService().FindAll()
		if listErr != nil || len(all) == 0 {
			t.Fatalf("store character: %v", err)
		}
		char = all[0]
	}
	stored, err := facade.QuestsService().Store(quest)
	if err != nil {
		t.Fatalf("store quest: %v", err)
	}
	return char, stored
}

func finishQuest(t *testing.T, facade Facade, characterID, questID string) {
	t.Helper()
	if _, err := facade.QuestsService().IncrementObjective(characterID, questID, "kill", 1); err != nil {
		t.Fatalf("increment: %v", err)
	}
	if _, err := facade.QuestsService().CompleteQuest(characterID, questID); err != nil {
		t.Fatalf("complete: %v", err)
	}
}
