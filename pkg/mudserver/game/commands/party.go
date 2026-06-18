package commands

import (
	"strings"
	"sync"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/service"
)

type partyInvite struct {
	PartyID     string
	InviterName string
}

var partyInvites = struct {
	sync.Mutex
	byCharacterID map[string]partyInvite
}{
	byCharacterID: make(map[string]partyInvite),
}

// PartyCommand manages first-pass social party behavior.
type PartyCommand struct{}

func (command *PartyCommand) Key() CommandKey { return &StartsWithCommandKey{} }

func (command *PartyCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You must select a character first.")
		return true
	}

	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, partyUsage())
		return true
	}

	switch strings.ToLower(parts[1]) {
	case "create":
		command.createParty(game, message)
	case "invite":
		command.invite(game, message, strings.Join(parts[2:], " "))
	case "accept":
		command.accept(game, message)
	case "leave":
		command.leave(game, message)
	case "say":
		command.say(game, message, strings.Join(parts[2:], " "))
	default:
		game.SendMessage() <- messages.Reply(message.FromUser.ID, partyUsage())
	}

	return true
}

func partyUsage() string {
	return "Party commands: party create, party invite <player>, party accept, party leave, party say <message>"
}

func (command *PartyCommand) createParty(game def.GameCtrl, message *messages.Message) {
	if party, err := game.GetFacade().PartiesService().FindPartyForCharacter(message.Character.ID); err == nil && party != nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You are already in a party.")
		return
	}

	partyName := message.Character.Name + "'s Party"
	party, err := game.GetFacade().PartiesService().CreateParty(&service.CreatePartyDTO{
		Name:       partyName,
		Characters: []string{message.Character.ID},
	})
	if err != nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Could not create party.")
		return
	}

	game.SendMessage() <- messages.Reply(message.FromUser.ID, "Created party ["+party.Name+"].")
}

func (command *PartyCommand) invite(game def.GameCtrl, message *messages.Message, targetName string) {
	targetName = strings.TrimSpace(targetName)
	if targetName == "" {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Usage: party invite <player>")
		return
	}

	party, err := game.GetFacade().PartiesService().FindPartyForCharacter(message.Character.ID)
	if err != nil || party == nil {
		command.createParty(game, message)
		party, err = game.GetFacade().PartiesService().FindPartyForCharacter(message.Character.ID)
		if err != nil || party == nil {
			game.SendMessage() <- messages.Reply(message.FromUser.ID, "Could not create a party for the invite.")
			return
		}
	}

	targetUserID, targetCharName, found := findOnlinePlayer(game, targetName)
	if !found {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Player '"+targetName+"' is not online.")
		return
	}
	if targetUserID == message.FromUser.ID {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You are already in your own party.")
		return
	}

	targetUser, err := game.GetFacade().UsersService().FindByID(targetUserID)
	if err != nil || targetUser == nil || targetUser.LastCharacter == "" {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Could not invite "+targetCharName+".")
		return
	}

	partyInvites.Lock()
	partyInvites.byCharacterID[targetUser.LastCharacter] = partyInvite{
		PartyID:     party.ID,
		InviterName: message.Character.Name,
	}
	partyInvites.Unlock()

	game.SendMessage() <- messages.Reply(targetUser.ID, message.Character.Name+" invited you to join a party. Type 'party accept' to join.")
	game.SendMessage() <- messages.Reply(message.FromUser.ID, "Invited "+targetCharName+" to your party.")
}

func (command *PartyCommand) accept(game def.GameCtrl, message *messages.Message) {
	partyInvites.Lock()
	invite, ok := partyInvites.byCharacterID[message.Character.ID]
	if ok {
		delete(partyInvites.byCharacterID, message.Character.ID)
	}
	partyInvites.Unlock()

	if !ok {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You have no pending party invites.")
		return
	}

	party, err := game.GetFacade().PartiesService().GetPartyByID(invite.PartyID)
	if err != nil || party == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "That party invite is no longer valid.")
		return
	}

	if existing, err := game.GetFacade().PartiesService().FindPartyForCharacter(message.Character.ID); err == nil && existing != nil && existing.ID != party.ID {
		_ = game.GetFacade().PartiesService().RemoveCharacterFromParty(existing, message.Character.ID)
	}
	if err := game.GetFacade().PartiesService().AddCharacterToParty(party, message.Character); err != nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Could not join party.")
		return
	}

	game.SendMessage() <- messages.Reply(message.FromUser.ID, "Joined "+invite.InviterName+"'s party.")
	command.broadcastParty(game, party.ID, "[Party] "+message.Character.Name+" joined the party.")
}

func (command *PartyCommand) leave(game def.GameCtrl, message *messages.Message) {
	party, err := game.GetFacade().PartiesService().FindPartyForCharacter(message.Character.ID)
	if err != nil || party == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You are not in a party.")
		return
	}
	if err := game.GetFacade().PartiesService().RemoveCharacterFromParty(party, message.Character.ID); err != nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Could not leave party.")
		return
	}

	game.SendMessage() <- messages.Reply(message.FromUser.ID, "You left the party.")
	command.broadcastParty(game, party.ID, "[Party] "+message.Character.Name+" left the party.")
}

func (command *PartyCommand) say(game def.GameCtrl, message *messages.Message, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Usage: party say <message>")
		return
	}
	party, err := game.GetFacade().PartiesService().FindPartyForCharacter(message.Character.ID)
	if err != nil || party == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You are not in a party.")
		return
	}
	command.broadcastParty(game, party.ID, "[Party] "+message.Character.Name+": "+text)
}

func (command *PartyCommand) broadcastParty(game def.GameCtrl, partyID string, text string) {
	party, err := game.GetFacade().PartiesService().GetPartyByID(partyID)
	if err != nil || party == nil {
		return
	}

	for _, characterID := range party.Characters {
		character, err := game.GetFacade().CharactersService().FindByID(characterID)
		if err != nil || character == nil {
			continue
		}
		user, err := game.GetFacade().UsersService().FindByID(character.BelongsUserID)
		if err != nil || user == nil || !user.IsOnline || user.LastCharacter != character.ID {
			continue
		}
		game.SendMessage() <- messages.Reply(user.ID, text)
	}
}
