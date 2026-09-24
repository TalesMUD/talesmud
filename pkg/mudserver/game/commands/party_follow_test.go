package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/service"
)

func importFollowRooms(t *testing.T, facade service.Facade, specs ...rooms.Room) {
	t.Helper()
	for i := range specs {
		room := specs[i]
		if room.Exits == nil {
			exits := rooms.Exits{}
			room.Exits = &exits
		}
		if room.Characters == nil {
			chars := rooms.Characters{}
			room.Characters = &chars
		}
		if _, err := facade.RoomsService().Import(&room); err != nil {
			t.Fatalf("import room %s: %v", room.ID, err)
		}
	}
}

func followParty(t *testing.T, facade service.Facade, leaderID string, memberIDs ...string) {
	t.Helper()
	if _, err := facade.PartiesService().CreateParty(&service.CreatePartyDTO{
		Name:              "Lanterns",
		Characters:        memberIDs,
		LeaderCharacterID: leaderID,
	}); err != nil {
		t.Fatal(err)
	}
}

func replyBlob(out []interface{}) string {
	var b strings.Builder
	for _, raw := range out {
		switch msg := raw.(type) {
		case messages.MessageResponse:
			if msg.Message != "" {
				b.WriteString(msg.Message)
				b.WriteByte('\n')
			}
		case *messages.EnterRoomMessage:
			if msg.Message != "" {
				b.WriteString(msg.Message)
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

func TestPartyFollowMovesWithLeaderExit(t *testing.T) {
	g, facade := newSocialTestGame(t)
	emberUser, ember := storeSocialPlayer(t, facade, "user-ember", "ref-ember", "char-ember", "Ember", "room-a")
	thornUser, thorn := storeSocialPlayer(t, facade, "user-thorn", "ref-thorn", "char-thorn", "Thorn", "room-a")
	importFollowRooms(t, facade,
		rooms.Room{Entity: &entities.Entity{ID: "room-a"}, Name: "Nest", Exits: &rooms.Exits{{Name: "north", Target: "room-b", Type: rooms.RoomExitTypeDirection}}, Characters: &rooms.Characters{ember.ID, thorn.ID}},
		rooms.Room{Entity: &entities.Entity{ID: "room-b"}, Name: "Corridor", Exits: &rooms.Exits{{Name: "north", Target: "room-c"}}, Characters: &rooms.Characters{}},
		rooms.Room{Entity: &entities.Entity{ID: "room-c"}, Name: "Alcove"},
	)
	g.ConnectUserSession(emberUser)
	g.SetUserSessionCharacter(emberUser, ember)
	g.ConnectUserSession(thornUser)
	g.SetUserSessionCharacter(thornUser, thorn)
	followParty(t, facade, ember.ID, ember.ID, thorn.ID)
	_ = drainSocialMessages(g.SendMessage())

	if !(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: thornUser, Character: thorn, Data: "party follow"}) {
		t.Fatal("party follow was not handled")
	}
	if blob := replyBlob(drainSocialMessages(g.SendMessage())); !strings.Contains(blob, "You are following Ember") {
		t.Fatalf("follow reply = %q", blob)
	}

	roomA, err := facade.RoomsService().FindByID("room-a")
	if err != nil {
		t.Fatal(err)
	}
	if !commands.TakeExit("north")(roomA, g, &messages.Message{FromUser: emberUser, Character: ember, Data: "north"}) {
		t.Fatal("leader exit was not handled")
	}
	gotThorn, _ := facade.CharactersService().FindByID(thorn.ID)
	gotEmber, _ := facade.CharactersService().FindByID(ember.ID)
	if gotEmber.CurrentRoomID != "room-b" || gotThorn.CurrentRoomID != "room-b" {
		t.Fatalf("after follow move ember=%s thorn=%s, want both room-b", gotEmber.CurrentRoomID, gotThorn.CurrentRoomID)
	}
	if blob := replyBlob(drainSocialMessages(g.SendMessage())); !strings.Contains(blob, "You follow Ember into Corridor") {
		t.Fatalf("follow move text = %q", blob)
	}
	if players := g.GetRoomPlayers("room-b", ""); len(players) != 2 {
		t.Fatalf("expected both live in room-b, got %#v", players)
	}

	if !(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: thornUser, Character: gotThorn, Data: "party unfollow"}) {
		t.Fatal("party unfollow was not handled")
	}
	_ = drainSocialMessages(g.SendMessage())

	roomB, err := facade.RoomsService().FindByID("room-b")
	if err != nil {
		t.Fatal(err)
	}
	emberNow := gotEmber
	if !commands.TakeExit("north")(roomB, g, &messages.Message{FromUser: emberUser, Character: emberNow, Data: "north"}) {
		t.Fatal("second leader exit was not handled")
	}
	gotThorn, _ = facade.CharactersService().FindByID(thorn.ID)
	gotEmber, _ = facade.CharactersService().FindByID(ember.ID)
	if gotEmber.CurrentRoomID != "room-c" {
		t.Fatalf("leader should be in room-c, got %s", gotEmber.CurrentRoomID)
	}
	if gotThorn.CurrentRoomID != "room-b" {
		t.Fatalf("unfollowed thorn should stay in room-b, got %s", gotThorn.CurrentRoomID)
	}
}

func TestPartyFollowCombatBlocksMove(t *testing.T) {
	g, facade := newSocialTestGame(t)
	emberUser, ember := storeSocialPlayer(t, facade, "user-ember2", "ref-ember2", "char-ember2", "Ember", "room-a")
	thornUser, thorn := storeSocialPlayer(t, facade, "user-thorn2", "ref-thorn2", "char-thorn2", "Thorn", "room-a")
	importFollowRooms(t, facade,
		rooms.Room{Entity: &entities.Entity{ID: "room-a"}, Name: "Nest", Exits: &rooms.Exits{{Name: "north", Target: "room-b"}}, Characters: &rooms.Characters{ember.ID, thorn.ID}},
		rooms.Room{Entity: &entities.Entity{ID: "room-b"}, Name: "Corridor"},
	)
	g.ConnectUserSession(emberUser)
	g.SetUserSessionCharacter(emberUser, ember)
	g.ConnectUserSession(thornUser)
	g.SetUserSessionCharacter(thornUser, thorn)
	followParty(t, facade, ember.ID, ember.ID, thorn.ID)

	if !(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: thornUser, Character: thorn, Data: "party follow"}) {
		t.Fatal("follow was not handled")
	}
	_ = drainSocialMessages(g.SendMessage())

	enemy := &npc.NPC{Entity: &entities.Entity{ID: "rat-follow"}, Name: "Cellar Rat", CurrentHitPoints: 8, MaxHitPoints: 8}
	if g.GetCombatEngine().InitiateCombat("room-a", []*characters.Character{thorn}, []*npc.NPC{enemy}) == nil {
		t.Fatal("failed to start combat for follower")
	}
	if !(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: thornUser, Character: thorn, Data: "party follow"}) {
		t.Fatal("in-combat follow was not handled")
	}
	if blob := replyBlob(drainSocialMessages(g.SendMessage())); !strings.Contains(blob, "can't follow while in combat") {
		t.Fatalf("expected combat refusal, got %q", blob)
	}

	roomA, _ := facade.RoomsService().FindByID("room-a")
	if !commands.TakeExit("north")(roomA, g, &messages.Message{FromUser: emberUser, Character: ember, Data: "north"}) {
		t.Fatal("leader exit was not handled")
	}
	gotThorn, _ := facade.CharactersService().FindByID(thorn.ID)
	if gotThorn.CurrentRoomID != "room-a" {
		t.Fatalf("combat follower was moved to %s", gotThorn.CurrentRoomID)
	}
	if blob := replyBlob(drainSocialMessages(g.SendMessage())); !strings.Contains(blob, "You stay behind") {
		t.Fatalf("expected stay-behind line, got %q", blob)
	}
}

func TestPartyFollowRefusesOutOfParty(t *testing.T) {
	g, facade := newSocialTestGame(t)
	user, char := storeSocialPlayer(t, facade, "user-solo", "ref-solo", "char-solo", "Wisp", "room-a")
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)

	if !(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: user, Character: char, Data: "party follow"}) {
		t.Fatal("party follow was not handled")
	}
	blob := replyBlob(drainSocialMessages(g.SendMessage()))
	if !strings.Contains(blob, "not in a party") {
		t.Fatalf("expected out-of-party refusal, got %q", blob)
	}
	if _, ok := g.PartyFollowTarget(char.ID); ok {
		t.Fatal("out-of-party follow stored a target")
	}
}

func TestPartyFollowSkipsTeleport(t *testing.T) {
	g, facade := newSocialTestGame(t)
	emberUser, ember := storeSocialPlayer(t, facade, "user-ember3", "ref-ember3", "char-ember3", "Ember", "room-a")
	thornUser, thorn := storeSocialPlayer(t, facade, "user-thorn3", "ref-thorn3", "char-thorn3", "Thorn", "room-a")
	importFollowRooms(t, facade,
		rooms.Room{Entity: &entities.Entity{ID: "room-a"}, Name: "Nest", Exits: &rooms.Exits{{Name: "portal", Target: "room-b", Type: rooms.RoomExitTypeTeleport}}, Characters: &rooms.Characters{ember.ID, thorn.ID}},
		rooms.Room{Entity: &entities.Entity{ID: "room-b"}, Name: "Elsewhere"},
	)
	g.ConnectUserSession(emberUser)
	g.SetUserSessionCharacter(emberUser, ember)
	g.ConnectUserSession(thornUser)
	g.SetUserSessionCharacter(thornUser, thorn)
	followParty(t, facade, ember.ID, ember.ID, thorn.ID)
	g.SetPartyFollow(thorn.ID, ember.ID)

	roomA, _ := facade.RoomsService().FindByID("room-a")
	if !commands.TakeExit("portal")(roomA, g, &messages.Message{FromUser: emberUser, Character: ember, Data: "portal"}) {
		t.Fatal("teleport exit was not handled")
	}
	gotEmber, _ := facade.CharactersService().FindByID(ember.ID)
	gotThorn, _ := facade.CharactersService().FindByID(thorn.ID)
	if gotEmber.CurrentRoomID != "room-b" {
		t.Fatalf("leader should teleport, got %s", gotEmber.CurrentRoomID)
	}
	if gotThorn.CurrentRoomID != "room-a" {
		t.Fatalf("follower should not cross a teleport, got %s", gotThorn.CurrentRoomID)
	}
}

func TestPartyLeaderCannotFollowSelf(t *testing.T) {
	g, facade := newSocialTestGame(t)
	user, char := storeSocialPlayer(t, facade, "user-lead", "ref-lead", "char-lead", "Ember", "room-a")
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)
	followParty(t, facade, char.ID, char.ID)
	if !(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: user, Character: char, Data: "party follow"}) {
		t.Fatal("party follow was not handled")
	}
	blob := replyBlob(drainSocialMessages(g.SendMessage()))
	if !strings.Contains(blob, "You lead the party") {
		t.Fatalf("expected leader refusal, got %q", blob)
	}
}
