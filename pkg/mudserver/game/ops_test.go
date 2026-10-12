package game

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/classkit"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/service"
)

func storeRoom(t *testing.T, facade service.Facade, id, name string) {
	t.Helper()
	exits := rooms.Exits{}
	room := &rooms.Room{Entity: &entities.Entity{ID: id}, Name: name, Area: "Zone", Exits: &exits}
	if _, err := facade.RoomsService().Import(room); err != nil {
		t.Fatalf("room %s: %v", id, err)
	}
}

func storeHero(t *testing.T, facade service.Facade, userID, charID, name, roomID string, online bool) *characters.Character {
	t.Helper()
	user := &entities.User{
		Entity:   &entities.Entity{ID: userID},
		RefID:    userID,
		Nickname: name + "-user",
		Role:     entities.RolePlayer,
	}
	if _, err := facade.UsersService().Import(user); err != nil {
		t.Fatalf("user: %v", err)
	}
	char := &characters.Character{
		Entity:           &entities.Entity{ID: charID},
		Name:             name,
		BelongsUser:      *traits.BelongsToUser(userID),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: roomID},
		Class:            characters.Class{ID: "warrior", Name: "Warrior"},
		Level:            1,
		CurrentHitPoints: 20,
		MaxHitPoints:     20,
	}
	if _, err := facade.CharactersService().Import(char); err != nil {
		t.Fatalf("character: %v", err)
	}
	_ = online
	fresh, err := facade.CharactersService().FindByID(charID)
	if err != nil {
		t.Fatal(err)
	}
	return fresh
}

func TestOpTeleportOnlineAndOfflineRoundTrip(t *testing.T) {
	g, facade := newSessionTestGame(t)
	storeRoom(t, facade, "room-a", "Alpha")
	storeRoom(t, facade, "room-b", "Beta")
	offline := storeHero(t, facade, "user-off", "char-off", "Offline", "room-a", false)
	res, err := g.OpTeleport(offline.ID, "room-b", false)
	if err != nil || !res.Undoable {
		t.Fatalf("offline teleport: %+v %v", res, err)
	}
	got, _ := facade.CharactersService().FindByID(offline.ID)
	if got.CurrentRoomID != "room-b" {
		t.Fatalf("offline room %s", got.CurrentRoomID)
	}
	if _, err := g.ApplyInverse(res.Inverse); err != nil {
		t.Fatalf("undo offline: %v", err)
	}
	got, _ = facade.CharactersService().FindByID(offline.ID)
	if got.CurrentRoomID != "room-a" {
		t.Fatalf("undo room %s", got.CurrentRoomID)
	}

	online := storeHero(t, facade, "user-on", "char-on", "Online", "room-a", false)
	user, _ := facade.UsersService().FindByID("user-on")
	char, _ := facade.CharactersService().FindByID(online.ID)
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)
	if _, err := g.OpTeleport(online.ID, "room-b", false); err != nil {
		t.Fatalf("online teleport: %v", err)
	}
	players := g.GetRoomPlayers("room-b", "")
	if len(players) != 1 || players[0].CharacterID != online.ID {
		t.Fatalf("session room: %+v", players)
	}
	if _, err := g.OpTeleport(online.ID, "room-a", false); err != nil {
		t.Fatal(err)
	}
	// Put them in a fight and refuse the move.
	enemy := &npc.NPC{Entity: &entities.Entity{ID: "rat"}, Name: "Rat", CurrentHitPoints: 10, MaxHitPoints: 10}
	g.NPCManager.RegisterExistingNPC(enemy, "room-a")
	hero, _ := facade.CharactersService().FindByID(online.ID)
	if inst := g.CombatController.InitiateCombat("room-a", []*characters.Character{hero}, []*npc.NPC{enemy}); inst == nil {
		t.Fatal("fight did not start")
	}
	if _, err := g.OpTeleport(online.ID, "room-b", false); err == nil {
		t.Fatal("expected combat block")
	}
	forced, err := g.OpTeleport(online.ID, "room-b", true)
	if err != nil {
		t.Fatalf("force: %v", err)
	}
	hero, _ = facade.CharactersService().FindByID(online.ID)
	if hero.InCombat || hero.CurrentRoomID != "room-b" {
		t.Fatalf("after force: combat %v room %s", hero.InCombat, hero.CurrentRoomID)
	}
	if forced.Detail["endedCombat"] != true {
		t.Fatalf("detail %+v", forced.Detail)
	}
}

func TestOpGiveAndTakeInverse(t *testing.T) {
	g, facade := newSessionTestGame(t)
	storeRoom(t, facade, "room-a", "Alpha")
	hero := storeHero(t, facade, "user-1", "char-1", "Hero", "room-a", false)
	tpl := &items.Item{Entity: &entities.Entity{ID: "tpl-bread"}, Name: "Bread", IsTemplate: true, Stackable: false}
	if _, err := facade.ItemsService().Import(tpl); err != nil {
		t.Fatalf("template: %v", err)
	}
	unique := &items.Item{Entity: &entities.Entity{ID: "tpl-relic"}, Name: "Relic", IsTemplate: true, Unique: true}
	if _, err := facade.ItemsService().Import(unique); err != nil {
		t.Fatalf("unique: %v", err)
	}
	given, err := g.OpGiveItem(hero.ID, "tpl-bread", 2)
	if err != nil {
		t.Fatalf("give: %v", err)
	}
	got, _ := facade.CharactersService().FindByID(hero.ID)
	if got.Inventory.Count() != 2 {
		t.Fatalf("inventory %d", got.Inventory.Count())
	}
	if _, err := g.ApplyInverse(given.Inverse); err != nil {
		t.Fatalf("undo give: %v", err)
	}
	got, _ = facade.CharactersService().FindByID(hero.ID)
	if got.Inventory.Count() != 0 {
		t.Fatalf("after undo %d", got.Inventory.Count())
	}

	again, err := g.OpGiveItem(hero.ID, "tpl-bread", 1)
	if err != nil {
		t.Fatal(err)
	}
	got, _ = facade.CharactersService().FindByID(hero.ID)
	instID := got.Inventory.Items[0].ID
	taken, err := g.OpTakeItem(hero.ID, instID, "", 1)
	if err != nil {
		t.Fatalf("take: %v", err)
	}
	got, _ = facade.CharactersService().FindByID(hero.ID)
	if got.Inventory.Count() != 0 {
		t.Fatalf("taken inventory %d", got.Inventory.Count())
	}
	if _, err := g.ApplyInverse(again.Inverse); err == nil {
		t.Fatal("expected give undo conflict after the item was taken")
	}
	if _, err := g.ApplyInverse(taken.Inverse); err != nil {
		t.Fatalf("undo take: %v", err)
	}
	got, _ = facade.CharactersService().FindByID(hero.ID)
	if got.Inventory.Count() != 1 || got.Inventory.Items[0].ID != instID {
		t.Fatalf("restored item %+v", got.Inventory.Items)
	}

	if _, err := g.OpGiveItem(hero.ID, "tpl-relic", 2); err == nil {
		t.Fatal("unique quantity should fail")
	}
	if _, err := g.OpGiveItem(hero.ID, "tpl-relic", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := g.OpGiveItem(hero.ID, "tpl-relic", 1); err == nil {
		t.Fatal("second unique should fail")
	}
	got, _ = facade.CharactersService().FindByID(hero.ID)
	if got.CountOfTemplate("tpl-relic") != 1 {
		t.Fatalf("relic count %d", got.CountOfTemplate("tpl-relic"))
	}
}

func TestOpQuestStepInverseKeepsRewards(t *testing.T) {
	g, facade := newSessionTestGame(t)
	storeRoom(t, facade, "room-a", "Alpha")
	hero := storeHero(t, facade, "user-q", "char-q", "Quester", "room-a", false)
	quest := &quests.Quest{
		Entity:      &entities.Entity{ID: "Q-ops"},
		Name:        "Ops Quest",
		Description: "Do the thing.",
		Source:      quests.QuestSource{Type: "auto"},
		Objectives: []quests.Objective{{
			ID: "obj-1", Type: quests.ObjectiveVisit, Description: "Visit Alpha", TargetID: "room-a", Amount: 1,
		}},
		Rewards: quests.Reward{XP: 25},
	}
	stored, err := facade.QuestsService().Store(quest)
	if err != nil {
		t.Fatalf("quest: %v", err)
	}
	questID := stored.ID
	step, err := g.OpQuestStep(hero.ID, questID, "obj-1", "complete")
	if err != nil {
		t.Fatalf("step: %v", err)
	}
	if step.Detail["completionFlow"] != "progress-only" {
		t.Fatalf("flow %+v", step.Detail)
	}
	got, _ := facade.CharactersService().FindByID(hero.ID)
	if got.XP != 0 {
		t.Fatalf("progress-only granted xp %d", got.XP)
	}
	done, err := g.OpQuestStep(hero.ID, questID, "", "complete")
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if done.Detail["completionFlow"] != "rewards" {
		t.Fatalf("reward flow %+v", done.Detail)
	}
	got, _ = facade.CharactersService().FindByID(hero.ID)
	if got.XP != 25 {
		t.Fatalf("xp %d", got.XP)
	}
	if _, err := g.ApplyInverse(done.Inverse); err != nil {
		t.Fatalf("undo quest: %v", err)
	}
	progress, _ := facade.QuestsService().GetProgress(hero.ID, questID)
	if progress == nil || progress.Status == quests.QuestStatusCompleted {
		t.Fatalf("progress %+v", progress)
	}
	got, _ = facade.CharactersService().FindByID(hero.ID)
	if got.XP != 25 {
		t.Fatalf("undo clawed xp back to %d", got.XP)
	}
}

func TestOpEndCombatDoesNotHealOrReward(t *testing.T) {
	g, facade := newSessionTestGame(t)
	storeRoom(t, facade, "room-a", "Alpha")
	hero := storeHero(t, facade, "user-c", "char-c", "Fighter", "room-a", false)
	hero.CurrentHitPoints = 12
	hero.XP = 4
	if err := facade.CharactersService().Update(hero.ID, hero); err != nil {
		t.Fatal(err)
	}
	enemy := &npc.NPC{Entity: &entities.Entity{ID: "wolf"}, Name: "Wolf", CurrentHitPoints: 8, MaxHitPoints: 30}
	g.NPCManager.RegisterExistingNPC(enemy, "room-a")
	fresh, _ := facade.CharactersService().FindByID(hero.ID)
	inst := g.CombatController.InitiateCombat("room-a", []*characters.Character{fresh}, []*npc.NPC{g.NPCManager.GetInstance("wolf")})
	if inst == nil {
		t.Fatal("no fight")
	}
	inst.Players[0].CurrentHP = 7
	inst.Enemies[0].CurrentHP = 3
	res, err := g.OpEndCombat(inst.ID, "", "")
	if err != nil {
		t.Fatalf("end: %v", err)
	}
	if res.Undoable {
		t.Fatal("end combat should not be undoable")
	}
	got, _ := facade.CharactersService().FindByID(hero.ID)
	if got.InCombat || got.CurrentHitPoints != 7 || got.XP != 4 {
		t.Fatalf("char hp %d xp %d combat %v", got.CurrentHitPoints, got.XP, got.InCombat)
	}
	wolf := g.NPCManager.GetInstance("wolf")
	if wolf == nil || wolf.InCombat || wolf.CurrentHitPoints != 3 {
		t.Fatalf("wolf %+v", wolf)
	}
	if g.CombatController.GetCombatInstance(hero.ID) != nil {
		t.Fatal("fight still registered")
	}
}

func TestOpNPCRespawnHealAndDespawn(t *testing.T) {
	g, facade := newSessionTestGame(t)
	storeRoom(t, facade, "room-a", "Alpha")
	tpl := &npc.NPC{
		Entity: &entities.Entity{ID: "tpl-rat"}, Name: "Rat", IsTemplate: true,
		MaxHitPoints: 18, CurrentHitPoints: 18, SpawnRoomID: "room-a",
	}
	if _, err := facade.NPCsService().Import(tpl); err != nil {
		t.Fatal(err)
	}
	res, err := g.OpNPCRespawn("tpl-rat", "")
	if err != nil {
		t.Fatalf("respawn: %v", err)
	}
	id, _ := res.Detail["npcInstanceId"].(string)
	rows, err := g.LiveNPCs("tpl-rat", "room-a")
	if err != nil || len(rows) != 1 || rows[0].HP != 18 {
		t.Fatalf("live %+v %v", rows, err)
	}
	g.NPCManager.UpdateInstance(id, func(n *npc.NPC) { n.CurrentHitPoints = 4 })
	healed, err := g.OpNPCHeal(id)
	if err != nil {
		t.Fatal(err)
	}
	if g.NPCManager.GetInstance(id).CurrentHitPoints != 18 {
		t.Fatal("not healed")
	}
	if _, err := g.ApplyInverse(healed.Inverse); err != nil {
		t.Fatal(err)
	}
	if g.NPCManager.GetInstance(id).CurrentHitPoints != 4 {
		t.Fatal("hp undo failed")
	}
	if _, err := g.OpNPCDespawn(id); err != nil {
		t.Fatal(err)
	}
	if g.NPCManager.GetInstance(id) != nil {
		t.Fatal("still present")
	}
	if _, err := g.ApplyInverse(res.Inverse); err == nil {
		t.Fatal("despawned npc should conflict with respawn undo")
	}
}

func TestOpInstanceCleanupRelocatesAndSkipsOccupied(t *testing.T) {
	g, facade := newSessionTestGame(t)
	storeRoom(t, facade, "R0001", "Start")
	exits := rooms.Exits{}
	clone := &rooms.Room{Entity: &entities.Entity{ID: "R0002~abc"}, Name: "Cellar Copy", Exits: &exits}
	if _, err := facade.RoomsService().Import(clone); err != nil {
		t.Fatal(err)
	}
	occupied := &rooms.Room{Entity: &entities.Entity{ID: "R0003~keep"}, Name: "Occupied Copy", Exits: &exits}
	if _, err := facade.RoomsService().Import(occupied); err != nil {
		t.Fatal(err)
	}
	hero := storeHero(t, facade, "user-i", "char-i", "Inside", "R0002~abc", false)
	other := storeHero(t, facade, "user-k", "char-k", "Keeper", "R0003~keep", false)
	res, err := g.OpInstanceCleanup("R0002~abc", false)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if res.Undoable {
		t.Fatal("cleanup should not be undoable")
	}
	got, _ := facade.CharactersService().FindByID(hero.ID)
	if got.CurrentRoomID != "R0001" {
		t.Fatalf("moved to %s", got.CurrentRoomID)
	}
	if _, err := facade.RoomsService().FindByID("R0002~abc"); err == nil {
		t.Fatal("copy still exists")
	}
	empty, err := g.OpInstanceCleanup("", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := facade.RoomsService().FindByID("R0003~keep"); err != nil {
		t.Fatalf("occupied copy removed: %v empty=%+v", err, empty.Detail)
	}
	kept, _ := facade.CharactersService().FindByID(other.ID)
	if kept.CurrentRoomID != "R0003~keep" {
		t.Fatalf("keeper moved to %s", kept.CurrentRoomID)
	}
}

func TestOpRegrantStarterKit(t *testing.T) {
	def := classkit.Lookup("warrior")
	if def == nil || def.Template == nil || len(def.Template.Items) == 0 {
		t.Skip("warrior kit is not loaded")
	}
	g, facade := newSessionTestGame(t)
	storeRoom(t, facade, "room-a", "Alpha")
	hero := storeHero(t, facade, "user-w", "char-w", "Warden", "room-a", false)
	res, err := g.OpRegrantStarterKit(hero.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Undoable {
		t.Fatalf("expected kit, got %+v", res)
	}
	got, _ := facade.CharactersService().FindByID(hero.ID)
	if len(got.EquippedItems) == 0 && got.Inventory.Count() == 0 {
		t.Fatal("kit did not land")
	}
	if _, err := g.ApplyInverse(res.Inverse); err != nil {
		t.Fatal(err)
	}
	got, _ = facade.CharactersService().FindByID(hero.ID)
	if len(got.EquippedItems) != 0 || got.Inventory.Count() != 0 {
		t.Fatalf("kit remains equipped %d bag %d", len(got.EquippedItems), got.Inventory.Count())
	}
	plain := storeHero(t, facade, "user-p", "char-p", "Plain", "room-a", false)
	plain.Class = characters.Class{ID: "no-such-class", Name: "No Such"}
	if err := facade.CharactersService().Update(plain.ID, plain); err != nil {
		t.Fatal(err)
	}
	skipped, err := g.OpRegrantStarterKit(plain.ID)
	if err != nil || skipped.Undoable {
		t.Fatalf("skip %+v %v", skipped, err)
	}
}
