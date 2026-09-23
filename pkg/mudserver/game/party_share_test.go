package game

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestSplitVictoryAmount(t *testing.T) {
	full := splitVictoryAmount(11, []string{"ember"}, "ember", true)
	if full["ember"] != 11 {
		t.Fatalf("solo total = %d, want 11", full["ember"])
	}

	even := splitVictoryAmount(10, []string{"ember", "thorn"}, "ember", false)
	if even["ember"] != 5 || even["thorn"] != 5 {
		t.Fatalf("discarded remainder = %#v, want 5/5", even)
	}

	kept := splitVictoryAmount(11, []string{"ember", "thorn"}, "ember", true)
	if kept["ember"] != 6 || kept["thorn"] != 5 {
		t.Fatalf("leftover to killer = %#v, want 6/5", kept)
	}

	three := splitVictoryAmount(10, []string{"ember", "thorn", "wisp"}, "ember", true)
	if three["ember"] != 4 || three["thorn"] != 3 || three["wisp"] != 3 {
		t.Fatalf("3-way leftover = %#v, want 4/3/3", three)
	}

	missingKiller := splitVictoryAmount(5, []string{"thorn", "wisp"}, "ember", true)
	if missingKiller["thorn"] != 3 || missingKiller["wisp"] != 2 {
		t.Fatalf("leftover falls back to first recipient = %#v", missingKiller)
	}
}

func sharePlayer(t *testing.T, facade service.Facade, userID, refID, charID, name, roomID string) (*entities.User, *characters.Character) {
	t.Helper()
	user := &entities.User{
		Entity:        &entities.Entity{ID: userID},
		RefID:         refID,
		Nickname:      name + "User",
		LastCharacter: charID,
	}
	if _, err := facade.UsersService().Import(user); err != nil {
		t.Fatalf("import user %s: %v", userID, err)
	}
	character := &characters.Character{
		Entity:           &entities.Entity{ID: charID},
		Name:             name,
		BelongsUser:      *traits.BelongsToUser(userID),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: roomID},
		Level:            1,
		MaxHitPoints:     30,
		CurrentHitPoints: 30,
	}
	stored, err := facade.CharactersService().Import(character)
	if err != nil {
		t.Fatalf("import character %s: %v", name, err)
	}
	return user, stored
}

func connectSharePlayer(g *Game, user *entities.User, char *characters.Character) {
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)
}

func shareParty(t *testing.T, facade service.Facade, name, leader string, members ...string) {
	t.Helper()
	if _, err := facade.PartiesService().CreateParty(&service.CreatePartyDTO{
		Name:              name,
		Characters:        members,
		LeaderCharacterID: leader,
	}); err != nil {
		t.Fatal(err)
	}
}

func runShareVictory(t *testing.T, g *Game, facade service.Facade, roomID, npcID string, xp int64, gold int32, players ...*characters.Character) {
	t.Helper()
	storeTestRoom(t, facade, roomID, nil)
	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: npcID},
		Name:             "Cellar Rat",
		Level:            1,
		IsDead:           true,
		MaxHitPoints:     8,
		CurrentHitPoints: 0,
		EnemyTrait: &npc.EnemyTrait{
			XPReward:   xp,
			GoldDrop:   npc.Range{Min: gold, Max: gold},
			Difficulty: "trivial",
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

func shareTexts(out []interface{}) (combatEnd, partyLines []string) {
	for _, msg := range out {
		switch m := msg.(type) {
		case *messages.CombatEndMessage:
			combatEnd = append(combatEnd, m.AudienceID+"\n"+m.Message)
		case messages.MessageResponse:
			if m.Type == messages.MessageTypeDefault && strings.Contains(m.Message, "[Party]") {
				partyLines = append(partyLines, m.AudienceID+"\n"+m.Message)
			}
		}
	}
	return combatEnd, partyLines
}

func TestVictorySoloNoPartyFullAward(t *testing.T) {
	g, facade := newNPCTestGame(t)
	_, ember := sharePlayer(t, facade, "user-ember", "ref-ember", "char-ember", "Ember", "room-fight")
	// A stranger in the room is not in the party and did not fight.
	strangerUser, stranger := sharePlayer(t, facade, "user-stranger", "ref-stranger", "char-stranger", "Pike", "room-fight")
	connectSharePlayer(g, strangerUser, stranger)

	runShareVictory(t, g, facade, "room-fight", "rat-solo", 20, 8, ember)

	got, err := facade.CharactersService().FindByID(ember.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.XP != 20 || got.Gold != 8 {
		t.Fatalf("solo award = xp %d gold %d, want 20 XP and 8 gold", got.XP, got.Gold)
	}
	idle, _ := facade.CharactersService().FindByID(stranger.ID)
	if idle.XP != 0 || idle.Gold != 0 {
		t.Fatalf("same-room stranger received xp %d gold %d", idle.XP, idle.Gold)
	}

	ends, party := shareTexts(drainGameMessages(g.SendMessage()))
	blob := strings.Join(ends, "\n") + "\n" + strings.Join(party, "\n")
	if strings.Contains(blob, "PARTY SHARE") || strings.Contains(blob, "[Party]") {
		t.Fatalf("solo victory announced a party share:\n%s", blob)
	}
	if !strings.Contains(blob, "+ 20 XP") || !strings.Contains(blob, "+ 8 Gold") {
		t.Fatalf("solo victory text missing full award:\n%s", blob)
	}
}

func TestVictoryPartyTwoSameRoomEqualSplit(t *testing.T) {
	g, facade := newNPCTestGame(t)
	emberUser, ember := sharePlayer(t, facade, "user-ember2", "ref-ember2", "char-ember2", "Ember", "room-fight")
	thornUser, thorn := sharePlayer(t, facade, "user-thorn2", "ref-thorn2", "char-thorn2", "Thorn", "room-fight")
	connectSharePlayer(g, emberUser, ember)
	connectSharePlayer(g, thornUser, thorn)
	shareParty(t, facade, "Lanterns", ember.ID, ember.ID, thorn.ID)

	// Thorn is in the room and online, but did not join the fight.
	runShareVictory(t, g, facade, "room-fight", "rat-pair", 11, 10, ember)

	gotEmber, _ := facade.CharactersService().FindByID(ember.ID)
	gotThorn, _ := facade.CharactersService().FindByID(thorn.ID)
	if gotEmber.XP != 6 || gotEmber.Gold != 5 {
		t.Fatalf("killer share = xp %d gold %d, want 6 XP and 5 gold (leftover XP to killer)", gotEmber.XP, gotEmber.Gold)
	}
	if gotThorn.XP != 5 || gotThorn.Gold != 5 {
		t.Fatalf("partner share = xp %d gold %d, want 5 XP and 5 gold", gotThorn.XP, gotThorn.Gold)
	}

	ends, party := shareTexts(drainGameMessages(g.SendMessage()))
	endBlob := strings.Join(ends, "\n")
	toastBlob := strings.Join(party, "\n")
	if !strings.Contains(endBlob, "PARTY SHARE") || !strings.Contains(endBlob, "Ember: +6 XP, +5 Gold") || !strings.Contains(endBlob, "Thorn: +5 XP, +5 Gold") {
		t.Fatalf("killer combat log missing share summary:\n%s", endBlob)
	}
	if !strings.Contains(toastBlob, thornUser.ID) || !strings.Contains(toastBlob, "Ember +6 XP, +5 Gold") || !strings.Contains(toastBlob, "Thorn +5 XP, +5 Gold") {
		t.Fatalf("partner toast missing share amounts:\n%s", toastBlob)
	}
	for _, line := range ends {
		if strings.HasPrefix(line, thornUser.ID+"\n") {
			t.Fatalf("idle partner should not receive combatEnd:\n%s", line)
		}
	}
}

func TestVictoryPartyThreeWithOneOutOfRoomExcluded(t *testing.T) {
	g, facade := newNPCTestGame(t)
	emberUser, ember := sharePlayer(t, facade, "user-ember3", "ref-ember3", "char-ember3", "Ember", "room-fight")
	thornUser, thorn := sharePlayer(t, facade, "user-thorn3", "ref-thorn3", "char-thorn3", "Thorn", "room-fight")
	wispUser, wisp := sharePlayer(t, facade, "user-wisp3", "ref-wisp3", "char-wisp3", "Wisp", "room-away")
	storeTestRoom(t, facade, "room-away", nil)
	connectSharePlayer(g, emberUser, ember)
	connectSharePlayer(g, thornUser, thorn)
	connectSharePlayer(g, wispUser, wisp)
	shareParty(t, facade, "Lanterns", ember.ID, ember.ID, thorn.ID, wisp.ID)

	runShareVictory(t, g, facade, "room-fight", "rat-trio", 11, 10, ember)

	gotEmber, _ := facade.CharactersService().FindByID(ember.ID)
	gotThorn, _ := facade.CharactersService().FindByID(thorn.ID)
	gotWisp, _ := facade.CharactersService().FindByID(wisp.ID)
	if gotEmber.XP != 6 || gotEmber.Gold != 5 {
		t.Fatalf("killer share = xp %d gold %d, want 6/5", gotEmber.XP, gotEmber.Gold)
	}
	if gotThorn.XP != 5 || gotThorn.Gold != 5 {
		t.Fatalf("in-room partner share = xp %d gold %d, want 5/5", gotThorn.XP, gotThorn.Gold)
	}
	if gotWisp.XP != 0 || gotWisp.Gold != 0 {
		t.Fatalf("out-of-room member received xp %d gold %d", gotWisp.XP, gotWisp.Gold)
	}

	ends, party := shareTexts(drainGameMessages(g.SendMessage()))
	blob := strings.Join(ends, "\n") + "\n" + strings.Join(party, "\n")
	if strings.Contains(blob, "Wisp") {
		t.Fatalf("out-of-room member named in share lines:\n%s", blob)
	}
	for _, line := range append(ends, party...) {
		if strings.HasPrefix(line, wispUser.ID+"\n") {
			t.Fatalf("out-of-room member was notified:\n%s", line)
		}
	}
}

func TestVictoryPartyOfflineMemberExcluded(t *testing.T) {
	g, facade := newNPCTestGame(t)
	emberUser, ember := sharePlayer(t, facade, "user-ember4", "ref-ember4", "char-ember4", "Ember", "room-fight")
	_, thorn := sharePlayer(t, facade, "user-thorn4", "ref-thorn4", "char-thorn4", "Thorn", "room-fight")
	connectSharePlayer(g, emberUser, ember)
	// Thorn is in the room on paper, but has no live session.
	shareParty(t, facade, "Lanterns", ember.ID, ember.ID, thorn.ID)

	runShareVictory(t, g, facade, "room-fight", "rat-offline", 11, 10, ember)

	gotEmber, _ := facade.CharactersService().FindByID(ember.ID)
	gotThorn, _ := facade.CharactersService().FindByID(thorn.ID)
	if gotEmber.XP != 11 || gotEmber.Gold != 10 {
		t.Fatalf("solo-effective award = xp %d gold %d, want full 11/10", gotEmber.XP, gotEmber.Gold)
	}
	if gotThorn.XP != 0 || gotThorn.Gold != 0 {
		t.Fatalf("offline member received xp %d gold %d", gotThorn.XP, gotThorn.Gold)
	}
	ends, party := shareTexts(drainGameMessages(g.SendMessage()))
	blob := strings.Join(ends, "\n") + "\n" + strings.Join(party, "\n")
	if strings.Contains(blob, "PARTY SHARE") || strings.Contains(blob, "[Party]") {
		t.Fatalf("offline partner produced a share summary:\n%s", blob)
	}
}

func TestVictoryNonPartyJoinersSplitWithoutShareBanner(t *testing.T) {
	g, facade := newNPCTestGame(t)
	_, aryn := sharePlayer(t, facade, "user-aryn", "ref-aryn", "char-aryn", "Aryn", "room-fight")
	_, bran := sharePlayer(t, facade, "user-bran", "ref-bran", "char-bran", "Bran", "room-fight")

	runShareVictory(t, g, facade, "room-fight", "rat-joiners", 20, 10, aryn, bran)

	gotA, _ := facade.CharactersService().FindByID(aryn.ID)
	gotB, _ := facade.CharactersService().FindByID(bran.ID)
	if gotA.XP != 10 || gotB.XP != 10 || gotA.Gold != 5 || gotB.Gold != 5 {
		t.Fatalf("non-party joiner split = A %d/%d B %d/%d, want 10 XP and 5 gold each", gotA.XP, gotA.Gold, gotB.XP, gotB.Gold)
	}
	ends, party := shareTexts(drainGameMessages(g.SendMessage()))
	blob := strings.Join(ends, "\n") + "\n" + strings.Join(party, "\n")
	if strings.Contains(blob, "PARTY SHARE") || strings.Contains(blob, "[Party]") {
		t.Fatalf("non-party fight announced a party share:\n%s", blob)
	}
}
