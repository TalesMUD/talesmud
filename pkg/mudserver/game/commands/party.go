package commands

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// PartyCommand handles minimal party social flow.
type PartyCommand struct{}

func (command *PartyCommand) Key() CommandKey { return &StartsWithCommandKey{} }

func (command *PartyCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You must select a character first.")
		return true
	}

	args := strings.Fields(message.Data)
	if len(args) == 1 {
		command.listParty(game, message)
		return true
	}

	switch strings.ToLower(args[1]) {
	case "invite":
		command.invite(game, message, strings.Join(args[2:], " "))
	case "accept":
		command.accept(game, message)
	case "decline":
		command.decline(game, message)
	case "leave":
		command.leave(game, message)
	case "list":
		command.listParty(game, message)
	default:
		command.chat(game, message, strings.TrimSpace(strings.TrimPrefix(message.Data, args[0]+" ")))
	}
	return true
}

func (command *PartyCommand) invite(game def.GameCtrl, message *messages.Message, targetName string) {
	targetName = strings.TrimSpace(targetName)
	if targetName == "" {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Usage: party invite <player>")
		return
	}

	target, ok := game.FindOnlinePlayerByName(targetName)
	if !ok {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Player '"+targetName+"' is not online.")
		return
	}
	if target.CharacterID == message.Character.ID {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You can't invite yourself.")
		return
	}

	party, err := game.GetFacade().PartiesService().FindByCharacterID(message.Character.ID)
	if err != nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Could not load your party.")
		return
	}
	if party == nil {
		party, err = game.GetFacade().PartiesService().Store(&entities.Party{
			Name:       message.Character.Name + "'s Party",
			Characters: []string{message.Character.ID},
		})
		if err != nil {
			game.SendMessage() <- messages.Reply(message.FromUser.ID, "Could not create a party.")
			return
		}
	}

	for _, memberID := range party.Characters {
		if memberID == target.CharacterID {
			game.SendMessage() <- messages.Reply(message.FromUser.ID, target.CharacterName+" is already in your party.")
			return
		}
	}

	game.SetPartyInvite(def.PartyInvite{
		PartyID:              party.ID,
		InviterUserID:        message.FromUser.ID,
		InviterCharacterID:   message.Character.ID,
		InviterCharacterName: message.Character.Name,
		TargetCharacterID:    target.CharacterID,
		TargetCharacterName:  target.CharacterName,
	})

	game.SendMessage() <- messages.Reply(message.FromUser.ID, "Party invite sent to "+target.CharacterName+".")
	game.SendMessage() <- messages.Reply(target.UserID, message.Character.Name+" invited you to a party. Type 'party accept' or 'party decline'.")
}

func (command *PartyCommand) accept(game def.GameCtrl, message *messages.Message) {
	invite, ok := game.GetPartyInvite(message.Character.ID)
	if !ok {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You have no pending party invite.")
		return
	}

	party, err := game.GetFacade().PartiesService().GetPartyByID(invite.PartyID)
	if err != nil || party == nil {
		game.ClearPartyInvite(message.Character.ID)
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "That party invite is no longer available.")
		return
	}

	if err := game.GetFacade().PartiesService().AddCharacterToParty(party, message.Character); err != nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Could not join the party.")
		return
	}
	game.ClearPartyInvite(message.Character.ID)

	command.sendToParty(game, party.Characters, fmt.Sprintf("[Party] %s joined the party.", message.Character.Name))
}

func (command *PartyCommand) decline(game def.GameCtrl, message *messages.Message) {
	invite, ok := game.GetPartyInvite(message.Character.ID)
	if !ok {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You have no pending party invite.")
		return
	}
	game.ClearPartyInvite(message.Character.ID)
	game.SendMessage() <- messages.Reply(message.FromUser.ID, "Party invite declined.")
	if invite.InviterUserID != "" {
		game.SendMessage() <- messages.Reply(invite.InviterUserID, message.Character.Name+" declined your party invite.")
	}
}

func (command *PartyCommand) leave(game def.GameCtrl, message *messages.Message) {
	party, err := game.GetFacade().PartiesService().FindByCharacterID(message.Character.ID)
	if err != nil || party == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You are not in a party.")
		return
	}

	membersBefore := append([]string{}, party.Characters...)
	if len(party.Characters) <= 2 {
		_ = game.GetFacade().PartiesService().DeletePartyByID(party.ID)
	} else {
		_ = game.GetFacade().PartiesService().RemoveCharacterFromParty(party, message.Character.ID)
	}
	command.sendToParty(game, membersBefore, fmt.Sprintf("[Party] %s left the party.", message.Character.Name))
}

func (command *PartyCommand) listParty(game def.GameCtrl, message *messages.Message) {
	party, err := game.GetFacade().PartiesService().FindByCharacterID(message.Character.ID)
	if err != nil || party == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You are not in a party. Use 'party invite <player>' to start one.")
		return
	}

	names := make([]string, 0, len(party.Characters))
	for _, memberID := range party.Characters {
		if char, err := game.GetFacade().CharactersService().FindByID(memberID); err == nil && char != nil {
			names = append(names, char.Name)
		}
	}
	game.SendMessage() <- messages.Reply(message.FromUser.ID, "Party members: "+strings.Join(names, ", "))
}

func (command *PartyCommand) chat(game def.GameCtrl, message *messages.Message, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Usage: party <message>")
		return
	}

	party, err := game.GetFacade().PartiesService().FindByCharacterID(message.Character.ID)
	if err != nil || party == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You are not in a party.")
		return
	}
	command.sendToParty(game, party.Characters, fmt.Sprintf("[Party] %s: %s", message.Character.Name, text))
}

func (command *PartyCommand) sendToParty(game def.GameCtrl, characterIDs []string, text string) {
	userByCharacter := map[string]string{}
	for _, player := range game.GetOnlinePlayers() {
		userByCharacter[player.CharacterID] = player.UserID
	}
	for _, characterID := range characterIDs {
		if userID := userByCharacter[characterID]; userID != "" {
			game.SendMessage() <- messages.Reply(userID, text)
		}
	}
}
