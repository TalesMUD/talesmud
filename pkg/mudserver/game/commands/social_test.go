package commands_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func newSocialTestGame(t *testing.T) (*game.Game, service.Facade) {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	return game.New(facade), facade
}

func drainSocialMessages(ch <-chan interface{}) []interface{} {
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

func socialReplyText(out interface{}) string {
	if msg, ok := out.(messages.MessageResponse); ok {
		return msg.Message
	}
	return ""
}

func storeSocialPlayer(t *testing.T, facade service.Facade, userID, refID, charID, charName, roomID string) (*entities.User, *characters.Character) {
	t.Helper()

	user := &entities.User{
		Entity:        &entities.Entity{ID: userID},
		RefID:         refID,
		Nickname:      charName + "User",
		LastCharacter: charID,
	}
	if _, err := facade.UsersService().Import(user); err != nil {
		t.Fatalf("import user %s: %v", userID, err)
	}

	character := &characters.Character{
		Entity:      &entities.Entity{ID: charID},
		Name:        charName,
		BelongsUser: *traits.BelongsToUser(userID),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: roomID},
	}
	if _, err := facade.CharactersService().Import(character); err != nil {
		t.Fatalf("import character %s: %v", charID, err)
	}

	return user, character
}

func TestWhoUsesLiveSessions(t *testing.T) {
	g, facade := newSocialTestGame(t)
	onlineUser, onlineChar := storeSocialPlayer(t, facade, "user-1", "ref-1", "char-1", "Aryn", "room-1")
	staleUser, _ := storeSocialPlayer(t, facade, "user-2", "ref-2", "char-2", "Bran", "room-1")
	staleUser.IsOnline = true
	if err := facade.UsersService().Update(staleUser.RefID, staleUser); err != nil {
		t.Fatalf("mark stale user online: %v", err)
	}

	g.ConnectUserSession(onlineUser)
	g.SetUserSessionCharacter(onlineUser, onlineChar)

	msg := &messages.Message{FromUser: onlineUser, Character: onlineChar, Data: "who"}
	if !(&commands.WhoCommand{}).Execute(g, msg) {
		t.Fatal("who command did not handle message")
	}

	out := drainSocialMessages(g.SendMessage())
	if len(out) != 1 {
		t.Fatalf("expected one reply, got %d", len(out))
	}
	text := socialReplyText(out[0])
	if !strings.Contains(text, "Aryn") {
		t.Fatalf("expected who output to include live player Aryn, got %q", text)
	}
	if strings.Contains(text, "Bran") {
		t.Fatalf("expected who output to exclude stale online flag, got %q", text)
	}
}

func TestPartyInviteAcceptAndChat(t *testing.T) {
	g, facade := newSocialTestGame(t)
	leaderUser, leaderChar := storeSocialPlayer(t, facade, "user-1", "ref-1", "char-1", "Aryn", "room-1")
	targetUser, targetChar := storeSocialPlayer(t, facade, "user-2", "ref-2", "char-2", "Bran", "room-1")

	g.ConnectUserSession(leaderUser)
	g.SetUserSessionCharacter(leaderUser, leaderChar)
	g.ConnectUserSession(targetUser)
	g.SetUserSessionCharacter(targetUser, targetChar)

	invite := &messages.Message{FromUser: leaderUser, Character: leaderChar, Data: "party invite Bran"}
	if !(&commands.PartyCommand{}).Execute(g, invite) {
		t.Fatal("party invite did not handle message")
	}

	accept := &messages.Message{FromUser: targetUser, Character: targetChar, Data: "party accept"}
	if !(&commands.PartyCommand{}).Execute(g, accept) {
		t.Fatal("party accept did not handle message")
	}

	party, err := facade.PartiesService().FindByCharacterID(leaderChar.ID)
	if err != nil {
		t.Fatalf("find leader party: %v", err)
	}
	if party == nil || len(party.Characters) != 2 {
		t.Fatalf("expected two-character party, got %#v", party)
	}

	_ = drainSocialMessages(g.SendMessage())
	chat := &messages.Message{FromUser: leaderUser, Character: leaderChar, Data: "party Ready?"}
	if !(&commands.PartyCommand{}).Execute(g, chat) {
		t.Fatal("party chat did not handle message")
	}

	out := drainSocialMessages(g.SendMessage())
	if len(out) != 2 {
		t.Fatalf("expected party chat to send to two users, got %d: %#v", len(out), out)
	}
	var sawLeader, sawTarget bool
	for _, raw := range out {
		msg, ok := raw.(messages.MessageResponse)
		if !ok {
			t.Fatalf("expected MessageResponse, got %T", raw)
		}
		if !strings.Contains(msg.Message, "[Party] Aryn: Ready?") {
			t.Fatalf("unexpected party chat text: %q", msg.Message)
		}
		if msg.AudienceID == leaderUser.ID {
			sawLeader = true
		}
		if msg.AudienceID == targetUser.ID {
			sawTarget = true
		}
	}
	if !sawLeader || !sawTarget {
		t.Fatalf("expected party chat for both users, saw leader=%v target=%v", sawLeader, sawTarget)
	}
}

func TestTakingExitUpdatesLiveRoomPresence(t *testing.T) {
	g, facade := newSocialTestGame(t)
	user, character := storeSocialPlayer(t, facade, "user-1", "ref-1", "char-1", "Aryn", "room-old")

	oldChars := rooms.Characters{character.ID}
	oldExits := rooms.Exits{{Name: "north", Target: "room-new"}}
	emptyExits := rooms.Exits{}
	oldRoom := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-old"},
		Name:       "Old Room",
		Exits:      &oldExits,
		Characters: &oldChars,
	}
	newChars := rooms.Characters{}
	newRoom := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-new"},
		Name:       "New Room",
		Exits:      &emptyExits,
		Characters: &newChars,
	}
	if _, err := facade.RoomsService().Import(oldRoom); err != nil {
		t.Fatalf("import old room: %v", err)
	}
	if _, err := facade.RoomsService().Import(newRoom); err != nil {
		t.Fatalf("import new room: %v", err)
	}

	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, character)

	msg := &messages.Message{FromUser: user, Character: character, Data: "north"}
	if !commands.TakeExit("north")(oldRoom, g, msg) {
		t.Fatal("take exit did not handle movement")
	}

	if players := g.GetRoomPlayers("room-old", character.ID); len(players) != 0 {
		t.Fatalf("expected old room to have no live players, got %#v", players)
	}
	players := g.GetRoomPlayers("room-new", character.ID)
	if len(players) != 1 || players[0].CharacterID != character.ID {
		t.Fatalf("expected new room live presence for character, got %#v", players)
	}
}
