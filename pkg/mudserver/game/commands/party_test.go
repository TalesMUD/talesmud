package commands_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func newPartyCommandTestGame(t *testing.T) (*game.Game, service.Facade) {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	return game.New(facade), facade
}

func drainPartyMessages(ch <-chan interface{}) []interface{} {
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

func partyResponseMessages(messagesOut []interface{}) []messages.MessageResponse {
	responses := make([]messages.MessageResponse, 0)
	for _, out := range messagesOut {
		if rsp, ok := out.(messages.MessageResponse); ok {
			responses = append(responses, rsp)
		}
	}
	return responses
}

func TestPartyCreateStoresPartyForCurrentCharacter(t *testing.T) {
	g, facade := newPartyCommandTestGame(t)
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1", LastCharacter: "char-1", IsOnline: true}
	if _, err := facade.UsersService().Import(user); err != nil {
		t.Fatalf("import user: %v", err)
	}
	character := &characters.Character{
		Entity:      &entities.Entity{ID: "char-1"},
		Name:        "Aster",
		BelongsUser: *traits.BelongsToUser("user-1"),
	}
	if _, err := facade.CharactersService().Import(character); err != nil {
		t.Fatalf("import character: %v", err)
	}

	msg := &messages.Message{FromUser: user, Character: character, Data: "party create"}
	if !(&commands.PartyCommand{}).Execute(g, msg) {
		t.Fatal("party command did not handle create")
	}

	party, err := facade.PartiesService().FindPartyForCharacter("char-1")
	if err != nil {
		t.Fatalf("find party for character: %v", err)
	}
	if party.Name != "Aster's Party" {
		t.Fatalf("expected default party name, got %q", party.Name)
	}
	if len(party.Characters) != 1 || party.Characters[0] != "char-1" {
		t.Fatalf("expected creator as only member, got %#v", party.Characters)
	}
}

func TestPartyInviteAcceptAddsTargetCharacter(t *testing.T) {
	g, facade := newPartyCommandTestGame(t)
	leaderUser := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1", LastCharacter: "char-1", IsOnline: true}
	targetUser := &entities.User{Entity: &entities.Entity{ID: "user-2"}, RefID: "auth|2", LastCharacter: "char-2", IsOnline: true}
	if _, err := facade.UsersService().Import(leaderUser); err != nil {
		t.Fatalf("import leader user: %v", err)
	}
	if _, err := facade.UsersService().Import(targetUser); err != nil {
		t.Fatalf("import target user: %v", err)
	}
	leader := &characters.Character{Entity: &entities.Entity{ID: "char-1"}, Name: "Aster", BelongsUser: *traits.BelongsToUser("user-1")}
	target := &characters.Character{Entity: &entities.Entity{ID: "char-2"}, Name: "Bryn", BelongsUser: *traits.BelongsToUser("user-2")}
	if _, err := facade.CharactersService().Import(leader); err != nil {
		t.Fatalf("import leader: %v", err)
	}
	if _, err := facade.CharactersService().Import(target); err != nil {
		t.Fatalf("import target: %v", err)
	}

	if !(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: leaderUser, Character: leader, Data: "party create"}) {
		t.Fatal("party create not handled")
	}
	if !(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: leaderUser, Character: leader, Data: "party invite Bryn"}) {
		t.Fatal("party invite not handled")
	}
	if !(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: targetUser, Character: target, Data: "party accept"}) {
		t.Fatal("party accept not handled")
	}

	party, err := facade.PartiesService().FindPartyForCharacter("char-2")
	if err != nil {
		t.Fatalf("find target party: %v", err)
	}
	if len(party.Characters) != 2 {
		t.Fatalf("expected two members after accept, got %#v", party.Characters)
	}
}

func TestPartySaySendsOnlyToOnlinePartyMembers(t *testing.T) {
	g, facade := newPartyCommandTestGame(t)
	leaderUser := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1", LastCharacter: "char-1", IsOnline: true}
	onlineUser := &entities.User{Entity: &entities.Entity{ID: "user-2"}, RefID: "auth|2", LastCharacter: "char-2", IsOnline: true}
	offlineUser := &entities.User{Entity: &entities.Entity{ID: "user-3"}, RefID: "auth|3", LastCharacter: "char-3", IsOnline: false}
	for _, user := range []*entities.User{leaderUser, onlineUser, offlineUser} {
		if _, err := facade.UsersService().Import(user); err != nil {
			t.Fatalf("import user %s: %v", user.ID, err)
		}
	}
	leader := &characters.Character{Entity: &entities.Entity{ID: "char-1"}, Name: "Aster", BelongsUser: *traits.BelongsToUser("user-1")}
	online := &characters.Character{Entity: &entities.Entity{ID: "char-2"}, Name: "Bryn", BelongsUser: *traits.BelongsToUser("user-2")}
	offline := &characters.Character{Entity: &entities.Entity{ID: "char-3"}, Name: "Cato", BelongsUser: *traits.BelongsToUser("user-3")}
	for _, character := range []*characters.Character{leader, online, offline} {
		if _, err := facade.CharactersService().Import(character); err != nil {
			t.Fatalf("import character %s: %v", character.ID, err)
		}
	}
	if _, err := facade.PartiesService().CreateParty(&service.CreatePartyDTO{
		Name:       "Delvers",
		Characters: []string{"char-1", "char-2", "char-3"},
	}); err != nil {
		t.Fatalf("create party: %v", err)
	}

	if !(&commands.PartyCommand{}).Execute(g, &messages.Message{FromUser: leaderUser, Character: leader, Data: "party say regroup at the gate"}) {
		t.Fatal("party say not handled")
	}

	responses := partyResponseMessages(drainPartyMessages(g.SendMessage()))
	var toLeader, toOnline, toOffline bool
	for _, rsp := range responses {
		if strings.Contains(rsp.Message, "[Party] Aster: regroup at the gate") {
			switch rsp.AudienceID {
			case "user-1":
				toLeader = true
			case "user-2":
				toOnline = true
			case "user-3":
				toOffline = true
			}
		}
	}
	if !toLeader || !toOnline {
		t.Fatalf("expected party chat to leader and online member, got %#v", responses)
	}
	if toOffline {
		t.Fatalf("did not expect party chat for offline member, got %#v", responses)
	}
}
