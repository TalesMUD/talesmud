package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// PartyInspectCommand reads worn gear and panel stats for yourself or a member of your party.
// It is read-only: it never changes the caller or the target.
type PartyInspectCommand struct{}

// Key returns the command key matcher.
func (command *PartyInspectCommand) Key() CommandKey { return &StartsWithCommandKey{} }

const partyInspectNotFound = "You can only inspect members of your party."

// Execute handles partyinspect <id or name>.
func (command *PartyInspectCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Inspect whom? Usage: partyinspect <id or name>")
		return true
	}
	query := strings.Join(parts[1:], " ")

	target := partyInspectTarget(game, message.Character, query)
	if target == nil {
		game.SendMessage() <- message.Reply(partyInspectNotFound)
		return true
	}

	game.SendMessage() <- messages.NewPartyInspectMessage(message.FromUser.ID, target)
	return true
}

// partyInspectTarget resolves query to the caller or a fellow party member, matching id first, then name.
func partyInspectTarget(game def.GameCtrl, caller *characters.Character, query string) *characters.Character {
	if caller.ID == query || strings.EqualFold(caller.Name, query) {
		return caller
	}
	if game == nil || game.GetFacade() == nil {
		return nil
	}
	party, err := game.GetFacade().PartiesService().FindByCharacterID(caller.ID)
	if err != nil || party == nil {
		return nil
	}
	inParty := false
	for _, id := range party.Characters {
		if id == caller.ID {
			inParty = true
			break
		}
	}
	if !inParty {
		return nil
	}
	members := make([]*characters.Character, 0, len(party.Characters))
	for _, id := range party.Characters {
		if id == caller.ID {
			continue
		}
		if id == query {
			if ch, err := game.GetFacade().CharactersService().FindByID(id); err == nil && ch != nil {
				return ch
			}
			return nil
		}
		if ch, err := game.GetFacade().CharactersService().FindByID(id); err == nil && ch != nil {
			members = append(members, ch)
		}
	}
	for _, ch := range members {
		if strings.EqualFold(ch.Name, query) {
			return ch
		}
	}
	return nil
}
