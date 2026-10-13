package game

import (
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/service"
)

func roomItemCount(room interface{ GetItemIDs() []string }) int {
	if room == nil {
		return 0
	}
	return len(room.GetItemIDs())
}

func TestResolveNeedGreedWinnerTiers(t *testing.T) {
	fixed := func(n int) int { return 0 }

	winner, tier, ok := resolveNeedGreedWinner(map[string]lootRollChoice{
		"a": lootChoiceNeed,
		"b": lootChoiceGreed,
		"c": lootChoicePass,
	}, fixed)
	if !ok || winner != "a" || tier != "Need" {
		t.Fatalf("Need beats Greed = %s/%s ok=%v", winner, tier, ok)
	}

	winner, tier, ok = resolveNeedGreedWinner(map[string]lootRollChoice{
		"a": lootChoiceGreed,
		"b": lootChoicePass,
		"c": lootChoiceGreed,
	}, fixed)
	if !ok || winner != "a" || tier != "Greed" {
		t.Fatalf("Greed beats Pass = %s/%s ok=%v", winner, tier, ok)
	}

	_, _, ok = resolveNeedGreedWinner(map[string]lootRollChoice{
		"a": lootChoicePass,
		"b": lootChoicePass,
	}, fixed)
	if ok {
		t.Fatal("all-pass should not pick a winner")
	}

	pick := 0
	rng := func(n int) int {
		if n != 2 {
			t.Fatalf("expected pool of 2, got %d", n)
		}
		return pick
	}
	winner, tier, ok = resolveNeedGreedWinner(map[string]lootRollChoice{
		"a": lootChoiceNeed,
		"b": lootChoiceNeed,
	}, rng)
	if !ok || winner != "a" || tier != "Need" {
		t.Fatalf("tie pick 0 = %s/%s", winner, tier)
	}
	pick = 1
	winner, _, ok = resolveNeedGreedWinner(map[string]lootRollChoice{
		"a": lootChoiceNeed,
		"b": lootChoiceNeed,
	}, rng)
	if !ok || winner != "b" {
		t.Fatalf("tie pick 1 = %s", winner)
	}
}

func TestParseLootRollChoice(t *testing.T) {
	for _, raw := range []string{"need", "NEED", " Need "} {
		c, ok := parseLootRollChoice(raw)
		if !ok || c != lootChoiceNeed {
			t.Fatalf("%q → %#v", raw, c)
		}
	}
	if _, ok := parseLootRollChoice("maybe"); ok {
		t.Fatal("invalid choice should fail")
	}
}

func needGreedPlayer(t *testing.T, facade service.Facade, userID, charID, name, roomID string) (*entities.User, *characters.Character) {
	t.Helper()
	return sharePlayer(t, facade, userID, userID, charID, name, roomID)
}

func storeNeedGreedLootTemplate(t *testing.T, facade service.Facade, id, name string) {
	t.Helper()
	tpl := &items.Item{
		Entity:    &entities.Entity{ID: id},
		Name:      name,
		Quality:   items.ItemQualityMagic,
		Stackable: false,
		Quantity:  1,
	}
	if _, err := facade.ItemsService().Import(tpl); err != nil {
		t.Fatalf("import item template %s: %v", id, err)
	}
}

func runNeedGreedVictory(t *testing.T, g *Game, facade service.Facade, roomID, npcID string, players ...*characters.Character) {
	t.Helper()
	storeTestRoom(t, facade, roomID, nil)
	storeNeedGreedLootTemplate(t, facade, "tpl-tooth", "Rat Tooth")
	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: npcID},
		Name:             "Cellar Rat",
		Level:            1,
		IsDead:           true,
		MaxHitPoints:     8,
		CurrentHitPoints: 0,
		EnemyTrait: &npc.EnemyTrait{
			XPReward:       10,
			GoldDrop:       npc.Range{Min: 2, Max: 2},
			Difficulty:     "trivial",
			GuaranteedLoot: []string{"tpl-tooth"},
		},
	}
	g.NPCManager.RegisterExistingNPC(enemy, roomID)

	refs := make([]combat.CombatantRef, 0, len(players))
	for _, char := range players {
		refs = append(refs, combat.CombatantRef{
			ID: char.ID, Name: char.Name, IsAlive: true, CurrentHP: 30, MaxHP: 30,
		})
	}
	g.CombatController.processCombatVictory(&combat.CombatInstance{
		ID:           "combat-" + npcID,
		OriginRoomID: roomID,
		State:        combat.CombatStateVictory,
		Players:      refs,
		Enemies: []combat.CombatantRef{
			{ID: enemy.ID, Name: enemy.Name, IsAlive: false, CurrentHP: 0, MaxHP: 8},
		},
	})
}

func needGreedMsgs(out []interface{}) (prompts []messages.NeedGreedMessage, lines []string) {
	for _, msg := range out {
		switch m := msg.(type) {
		case *messages.NeedGreedMessage:
			prompts = append(prompts, *m)
		case messages.NeedGreedMessage:
			prompts = append(prompts, m)
		case messages.MessageResponse:
			if strings.Contains(m.Message, "[Party]") || strings.Contains(m.Message, "Loot roll") {
				lines = append(lines, m.AudienceID+"\n"+m.Message)
			}
		case *messages.MessageResponse:
			if m != nil && (strings.Contains(m.Message, "[Party]") || strings.Contains(m.Message, "Loot roll")) {
				lines = append(lines, m.AudienceID+"\n"+m.Message)
			}
		}
	}
	return prompts, lines
}

func TestNeedGreedSoloUnchangedGroundDrop(t *testing.T) {
	g, facade := newNPCTestGame(t)
	_, ember := needGreedPlayer(t, facade, "user-ember", "char-ember", "Ember", "room-solo")
	emberUser, _ := facade.UsersService().FindByID("user-ember")
	connectSharePlayer(g, emberUser, ember)

	drain := make(chan interface{}, 64)
	g.sendMessage = drain
	go func() {
		for range drain {
		}
	}()

	runNeedGreedVictory(t, g, facade, "room-solo", "rat-solo", ember)

	if g.HasActiveNeedGreed(ember.ID) {
		t.Fatal("solo should not open Need/Greed")
	}
	room, err := facade.RoomsService().FindByID("room-solo")
	if err != nil || roomItemCount(room) == 0 {
		t.Fatalf("solo loot should land on ground, items=%v", room)
	}
}

func TestNeedGreedPartyOpensPromptAndNeedWins(t *testing.T) {
	g, facade := newNPCTestGame(t)
	emberUser, ember := needGreedPlayer(t, facade, "user-ember", "char-ember", "Ember", "room-ng")
	thornUser, thorn := needGreedPlayer(t, facade, "user-thorn", "char-thorn", "Thorn", "room-ng")
	connectSharePlayer(g, emberUser, ember)
	connectSharePlayer(g, thornUser, thorn)
	shareParty(t, facade, "Hunt", ember.ID, ember.ID, thorn.ID)

	out := make(chan interface{}, 128)
	g.sendMessage = out

	runNeedGreedVictory(t, g, facade, "room-ng", "rat-party", ember, thorn)

	prompts, _ := needGreedMsgs(drainNeedGreed(out, 40*time.Millisecond))
	pending := 0
	for _, p := range prompts {
		if p.Pending && p.ItemName == "Rat Tooth" {
			pending++
		}
	}
	if pending < 2 {
		t.Fatalf("expected Need/Greed prompts for both members, got %d pending of %d msgs", pending, len(prompts))
	}
	room, _ := facade.RoomsService().FindByID("room-ng")
	if roomItemCount(room) > 0 {
		t.Fatalf("item should not be on ground while roll is open, got %v", room.GetItemIDs())
	}

	if !g.RecordNeedGreedChoice(ember.ID, emberUser.ID, "need") {
		t.Fatal("ember need not recorded")
	}
	if !g.RecordNeedGreedChoice(thorn.ID, thornUser.ID, "greed") {
		t.Fatal("thorn greed not recorded")
	}

	drainNeedGreed(out, 40*time.Millisecond)

	gotEmber, _ := facade.CharactersService().FindByID(ember.ID)
	if gotEmber == nil || gotEmber.Inventory.Count() < 1 {
		t.Fatal("Need winner should receive the item in inventory")
	}
	found := false
	for _, it := range gotEmber.Inventory.Items {
		if it != nil && it.Name == "Rat Tooth" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected Rat Tooth in Ember inventory")
	}
	room, _ = facade.RoomsService().FindByID("room-ng")
	if roomItemCount(room) > 0 {
		t.Fatalf("won item should not remain on ground, got %v", room.GetItemIDs())
	}
	if g.HasActiveNeedGreed(ember.ID) || g.HasActiveNeedGreed(thorn.ID) {
		t.Fatal("round should be closed after both chose")
	}
}

func TestNeedGreedAllPassGoesToGround(t *testing.T) {
	g, facade := newNPCTestGame(t)
	emberUser, ember := needGreedPlayer(t, facade, "user-ember", "char-ember", "Ember", "room-pass")
	thornUser, thorn := needGreedPlayer(t, facade, "user-thorn", "char-thorn", "Thorn", "room-pass")
	connectSharePlayer(g, emberUser, ember)
	connectSharePlayer(g, thornUser, thorn)
	shareParty(t, facade, "Hunt", ember.ID, ember.ID, thorn.ID)

	out := make(chan interface{}, 128)
	g.sendMessage = out
	runNeedGreedVictory(t, g, facade, "room-pass", "rat-pass", ember, thorn)
	drainNeedGreed(out, 20*time.Millisecond)

	g.RecordNeedGreedChoice(ember.ID, emberUser.ID, "pass")
	g.RecordNeedGreedChoice(thorn.ID, thornUser.ID, "pass")
	drainNeedGreed(out, 20*time.Millisecond)

	room, err := facade.RoomsService().FindByID("room-pass")
	if err != nil || roomItemCount(room) == 0 {
		t.Fatal("all-pass should leave item on the ground")
	}
}

func TestNeedGreedTimeoutGoesToGround(t *testing.T) {
	g, facade := newNPCTestGame(t)
	emberUser, ember := needGreedPlayer(t, facade, "user-ember", "char-ember", "Ember", "room-ttl")
	thornUser, thorn := needGreedPlayer(t, facade, "user-thorn", "char-thorn", "Thorn", "room-ttl")
	connectSharePlayer(g, emberUser, ember)
	connectSharePlayer(g, thornUser, thorn)
	shareParty(t, facade, "Hunt", ember.ID, ember.ID, thorn.ID)

	out := make(chan interface{}, 128)
	g.sendMessage = out
	runNeedGreedVictory(t, g, facade, "room-ttl", "rat-ttl", ember, thorn)
	drainNeedGreed(out, 20*time.Millisecond)

	g.needGreed.mu.Lock()
	if g.needGreed.active != nil {
		g.needGreed.active.Expires = time.Now().Add(-time.Second)
	}
	g.needGreed.mu.Unlock()
	g.ExpireNeedGreedRounds()
	drainNeedGreed(out, 20*time.Millisecond)

	room, err := facade.RoomsService().FindByID("room-ttl")
	if err != nil || roomItemCount(room) == 0 {
		t.Fatal("timeout with no Need/Greed should leave item on the ground")
	}
}

func TestNeedGreedDisconnectCountsAsPass(t *testing.T) {
	g, facade := newNPCTestGame(t)
	emberUser, ember := needGreedPlayer(t, facade, "user-ember", "char-ember", "Ember", "room-dc")
	thornUser, thorn := needGreedPlayer(t, facade, "user-thorn", "char-thorn", "Thorn", "room-dc")
	connectSharePlayer(g, emberUser, ember)
	connectSharePlayer(g, thornUser, thorn)
	shareParty(t, facade, "Hunt", ember.ID, ember.ID, thorn.ID)

	out := make(chan interface{}, 128)
	g.sendMessage = out
	runNeedGreedVictory(t, g, facade, "room-dc", "rat-dc", ember, thorn)
	drainNeedGreed(out, 20*time.Millisecond)

	g.RecordNeedGreedChoice(ember.ID, emberUser.ID, "greed")
	g.DisconnectUserSession(thornUser.ID)
	drainNeedGreed(out, 40*time.Millisecond)

	gotEmber, _ := facade.CharactersService().FindByID(ember.ID)
	if gotEmber == nil || gotEmber.Inventory.Count() < 1 {
		t.Fatal("disconnect=pass should let remaining Greed win")
	}
}

func TestNeedGreedSequentialMultiDrop(t *testing.T) {
	g, facade := newNPCTestGame(t)
	emberUser, ember := needGreedPlayer(t, facade, "user-ember", "char-ember", "Ember", "room-multi")
	thornUser, thorn := needGreedPlayer(t, facade, "user-thorn", "char-thorn", "Thorn", "room-multi")
	connectSharePlayer(g, emberUser, ember)
	connectSharePlayer(g, thornUser, thorn)
	shareParty(t, facade, "Hunt", ember.ID, ember.ID, thorn.ID)
	storeNeedGreedLootTemplate(t, facade, "tpl-bone", "Rat Bone")
	storeTestRoom(t, facade, "room-multi", nil)

	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: "rat-multi"},
		Name:             "Fat Rat",
		Level:            1,
		IsDead:           true,
		MaxHitPoints:     8,
		CurrentHitPoints: 0,
		EnemyTrait: &npc.EnemyTrait{
			XPReward:       10,
			GoldDrop:       npc.Range{Min: 1, Max: 1},
			Difficulty:     "trivial",
			GuaranteedLoot: []string{"tpl-tooth", "tpl-bone"},
		},
	}
	storeNeedGreedLootTemplate(t, facade, "tpl-tooth", "Rat Tooth")
	g.NPCManager.RegisterExistingNPC(enemy, "room-multi")

	out := make(chan interface{}, 256)
	g.sendMessage = out
	g.CombatController.processCombatVictory(&combat.CombatInstance{
		ID:           "combat-multi",
		OriginRoomID: "room-multi",
		State:        combat.CombatStateVictory,
		Players: []combat.CombatantRef{
			{ID: ember.ID, Name: ember.Name, IsAlive: true, CurrentHP: 30, MaxHP: 30},
			{ID: thorn.ID, Name: thorn.Name, IsAlive: true, CurrentHP: 30, MaxHP: 30},
		},
		Enemies: []combat.CombatantRef{
			{ID: enemy.ID, Name: enemy.Name, IsAlive: false, CurrentHP: 0, MaxHP: 8},
		},
	})
	drainNeedGreed(out, 30*time.Millisecond)

	g.needGreed.mu.Lock()
	activeName := ""
	pending := 0
	if g.needGreed.active != nil && g.needGreed.active.Item != nil {
		activeName = g.needGreed.active.Item.Name
	}
	pending = len(g.needGreed.pending)
	g.needGreed.mu.Unlock()
	if activeName == "" || pending != 1 {
		t.Fatalf("expected one active + one pending, active=%q pending=%d", activeName, pending)
	}

	g.RecordNeedGreedChoice(ember.ID, emberUser.ID, "pass")
	g.RecordNeedGreedChoice(thorn.ID, thornUser.ID, "pass")
	drainNeedGreed(out, 30*time.Millisecond)

	g.needGreed.mu.Lock()
	secondOpen := g.needGreed.active != nil
	g.needGreed.mu.Unlock()
	if !secondOpen {
		t.Fatal("second drop should open after first resolves")
	}

	g.RecordNeedGreedChoice(ember.ID, emberUser.ID, "need")
	g.RecordNeedGreedChoice(thorn.ID, thornUser.ID, "pass")
	drainNeedGreed(out, 30*time.Millisecond)

	room, _ := facade.RoomsService().FindByID("room-multi")
	if roomItemCount(room) != 1 {
		t.Fatalf("first pass→ground should leave one item, got %#v", room)
	}
	gotEmber, _ := facade.CharactersService().FindByID(ember.ID)
	if gotEmber == nil || gotEmber.Inventory.Count() < 1 {
		t.Fatal("second Need should grant inventory item")
	}
}

func TestNeedGreedEligibilityRequiresTwoPartyMembers(t *testing.T) {
	g, facade := newNPCTestGame(t)
	emberUser, ember := needGreedPlayer(t, facade, "user-ember", "char-ember", "Ember", "room-elig")
	pikeUser, pike := needGreedPlayer(t, facade, "user-pike", "char-pike", "Pike", "room-elig")
	connectSharePlayer(g, emberUser, ember)
	connectSharePlayer(g, pikeUser, pike)
	// Not in a party together — free-for-all ground loot.
	out := make(chan interface{}, 64)
	g.sendMessage = out
	runNeedGreedVictory(t, g, facade, "room-elig", "rat-elig", ember, pike)
	drainNeedGreed(out, 20*time.Millisecond)
	if g.HasActiveNeedGreed(ember.ID) {
		t.Fatal("non-party co-fighters should not open Need/Greed")
	}
	room, _ := facade.RoomsService().FindByID("room-elig")
	if roomItemCount(room) == 0 {
		t.Fatal("non-party loot should hit the ground")
	}
}

func drainNeedGreed(ch <-chan interface{}, wait time.Duration) []interface{} {
	deadline := time.Now().Add(wait)
	var out []interface{}
	for time.Now().Before(deadline) {
		select {
		case msg := <-ch:
			out = append(out, msg)
		case <-time.After(5 * time.Millisecond):
		}
	}
	return out
}
