package game

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/service"
)

func setupCellarRatsQuest(t *testing.T, g *Game, facade service.Facade) (*characters.Character, string) {
	t.Helper()
	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity:     &entities.Entity{ID: "NPC0001"},
		Name:       "Mira",
		IsTemplate: false,
	}); err != nil {
		t.Fatalf("store Mira: %v", err)
	}
	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity:     &entities.Entity{ID: "ENM0008"},
		Name:       "Sewer Rat",
		IsTemplate: true,
		Level:      2,
	}); err != nil {
		t.Fatalf("store ENM0008: %v", err)
	}
	storedQuest, err := facade.QuestsService().Store(&quests.Quest{
		Entity:      &entities.Entity{ID: "QST0203"},
		Name:        "Cellar Rats",
		Description: "Mira's cellar has a rat problem.",
		Category:    "side",
		Level:       1,
		Source:      quests.QuestSource{Type: "npc", NPCID: "NPC0001"},
		Objectives: []quests.Objective{{
			ID:          "kill_cellar_rats",
			Type:        quests.ObjectiveKill,
			Description: "Kill 5 Sewer Rats in the cellar",
			TargetID:    "ENM0008",
			TargetName:  "Sewer Rat",
			Amount:      5,
			Order:       1,
		}},
	})
	if err != nil {
		t.Fatalf("store quest: %v", err)
	}

	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:      &entities.Entity{ID: "char-marcus"},
		Name:        "Marcus",
		BelongsUser: *traits.BelongsToUser("user-marcus"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0203"},
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}
	if _, err := facade.QuestsService().AcceptQuest(char.ID, storedQuest.ID); err != nil {
		t.Fatalf("accept quest: %v", err)
	}
	return char, storedQuest.ID
}

func cellarRatsProgress(t *testing.T, facade service.Facade, characterID, questID string) quests.ObjectiveProgress {
	t.Helper()
	progress, err := facade.QuestsService().GetProgress(characterID, questID)
	if err != nil || progress == nil || len(progress.Objectives) == 0 {
		t.Fatalf("missing quest progress: %v %#v", err, progress)
	}
	return progress.Objectives[0]
}

func TestOnNPCKilledCountsClonedCellarSewerRat(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R0203", rooms.Exits{{Name: "down", Target: "R0215", Type: "direction"}})
	storeTestRoom(t, facade, "R0215", rooms.Exits{{Name: "up", Target: "R0203"}})
	cellar, err := facade.RoomsService().FindByID("R0215")
	if err != nil {
		t.Fatal(err)
	}
	cellar.Tags = []string{"instance", "cellar"}
	if err := facade.RoomsService().Update("R0215", cellar); err != nil {
		t.Fatal(err)
	}

	char, questID := setupCellarRatsQuest(t, g, facade)

	spawned, err := g.NPCManager.SpawnInstanceDirect("ENM0008", "R0215")
	if err != nil {
		t.Fatalf("spawn sewer rat: %v", err)
	}
	if spawned.TemplateID != "ENM0008" {
		t.Fatalf("spawned template id = %q", spawned.TemplateID)
	}

	clonedRoom, err := g.RoomInstances.Enter(char.ID, "R0203", "R0215")
	if err != nil {
		t.Fatal(err)
	}
	if clonedRoom == "R0215" {
		t.Fatal("expected private cellar clone")
	}
	clones := g.NPCManager.GetInstancesInRoom(clonedRoom)
	if len(clones) != 1 {
		t.Fatalf("expected 1 cloned rat, got %d", len(clones))
	}
	if clones[0].TemplateID != "ENM0008" {
		t.Fatalf("cloned rat template id = %q (id=%s)", clones[0].TemplateID, clones[0].ID)
	}

	g.QuestTracker.OnNPCKilled(char.ID, char.BelongsUserID, clones[0])
	got := cellarRatsProgress(t, facade, char.ID, questID)
	if got.Current != 1 {
		t.Fatalf("expected 1/5 after killing cloned sewer rat, got %d/%d", got.Current, got.Required)
	}
}

func TestOnNPCKilledCountsCellarRatByDisplayNameIfTemplateIDMissing(t *testing.T) {
	g, facade := newNPCTestGame(t)
	char, questID := setupCellarRatsQuest(t, g, facade)

	unnamedInstance := &npc.NPC{
		Entity:     &entities.Entity{ID: "uuid-no-template"},
		Name:       "Sewer Rat",
		TemplateID: "",
	}
	g.QuestTracker.OnNPCKilled(char.ID, char.BelongsUserID, unnamedInstance)
	got := cellarRatsProgress(t, facade, char.ID, questID)
	if got.Current != 1 {
		t.Fatalf("expected name match to increment, got %d", got.Current)
	}

	g.QuestTracker.OnNPCKilled(char.ID, char.BelongsUserID, &npc.NPC{
		Entity: &entities.Entity{ID: "enm0001-inst"},
		Name:   "Catacomb Rat",
	})
	got = cellarRatsProgress(t, facade, char.ID, questID)
	if got.Current != 1 {
		t.Fatalf("catacomb rat must not count, got %d", got.Current)
	}
}

func TestOnNPCKilledCompletesCellarRatsAtFive(t *testing.T) {
	g, facade := newNPCTestGame(t)
	char, questID := setupCellarRatsQuest(t, g, facade)
	rat := &npc.NPC{
		Entity:     &entities.Entity{ID: "rat-1"},
		Name:       "Sewer Rat",
		TemplateID: "ENM0008",
	}
	for i := 0; i < 5; i++ {
		g.QuestTracker.OnNPCKilled(char.ID, char.BelongsUserID, rat)
	}
	got := cellarRatsProgress(t, facade, char.ID, questID)
	if !got.Completed || got.Current != 5 {
		t.Fatalf("expected complete 5/5, got current=%d completed=%v", got.Current, got.Completed)
	}
}

func TestOnRoomEnterCountsPrivateCellarClone(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R0215", nil)
	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity: &entities.Entity{ID: "NPC0001"},
		Name:   "Mira",
	}); err != nil {
		t.Fatal(err)
	}
	storedQuest, err := facade.QuestsService().Store(&quests.Quest{
		Entity:      &entities.Entity{ID: "QST-visit"},
		Name:        "See the cellar",
		Description: "Go downstairs.",
		Source:      quests.QuestSource{Type: "npc", NPCID: "NPC0001"},
		Objectives: []quests.Objective{{
			ID:          "visit_cellar",
			Type:        quests.ObjectiveVisit,
			Description: "Enter the cellar",
			TargetID:    "R0215",
			Amount:      1,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:      &entities.Entity{ID: "char-visit"},
		Name:        "Visitor",
		BelongsUser: *traits.BelongsToUser("user-visit"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := facade.QuestsService().AcceptQuest(char.ID, storedQuest.ID); err != nil {
		t.Fatal(err)
	}

	g.QuestTracker.OnRoomEnter(char.ID, char.BelongsUserID, &rooms.Room{
		Entity: &entities.Entity{ID: "R0215~deadbeef"},
		Name:   "The Weary Wanderer - Cellar",
	})
	progress, err := facade.QuestsService().GetProgress(char.ID, storedQuest.ID)
	if err != nil || progress == nil {
		t.Fatalf("progress: %v", err)
	}
	if !progress.Objectives[0].Completed {
		t.Fatalf("expected visit objective to complete in private cellar clone, got %+v", progress.Objectives[0])
	}
}

func TestProcessCombatVictoryCreditsKillAfterNPCRemoved(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R0215~clone", nil)
	char, questID := setupCellarRatsQuest(t, g, facade)

	instance := &combat.CombatInstance{
		ID:           "combat-cellar",
		OriginRoomID: "R0215~clone",
		State:        combat.CombatStateVictory,
		Players: []combat.CombatantRef{
			{ID: char.ID, Name: char.Name, IsAlive: true, CurrentHP: 10, MaxHP: 10},
		},
		Enemies: []combat.CombatantRef{
			{ID: "gone-rat", Name: "Sewer Rat", TemplateID: "ENM0008", IsAlive: false, CurrentHP: 0, MaxHP: 12},
		},
	}
	g.CombatController.processCombatVictory(instance)

	got := cellarRatsProgress(t, facade, char.ID, questID)
	if got.Current != 1 {
		t.Fatalf("expected kill credit from combatant snapshot after NPC was removed, got %d", got.Current)
	}
}
