package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// FriendCommand manages a per-character friends list.
type FriendCommand struct{}

func (command *FriendCommand) Key() CommandKey { return &StartsWithCommandKey{} }

func (command *FriendCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You must select a character first.")
		return true
	}
	if message.FromUser != nil && message.FromUser.IsGuest {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Friends are for lasting adventurers. Sign in to keep a friends list.")
		return true
	}

	args := strings.Fields(message.Data)
	if len(args) == 0 {
		return true
	}

	verb := "list"
	name := ""
	if len(args) >= 2 {
		verb = strings.ToLower(args[1])
		name = strings.TrimSpace(strings.Join(args[2:], " "))
	}

	switch verb {
	case "add":
		command.add(game, message, name)
	case "remove", "rm", "delete":
		command.remove(game, message, name)
	case "list", "ls":
		command.list(game, message)
	default:
		if strings.EqualFold(args[0], "friends") && len(args) == 1 {
			command.list(game, message)
			return true
		}
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Usage: friend add <name>, friend remove <name>, friend list")
	}
	return true
}

func (command *FriendCommand) add(game def.GameCtrl, message *messages.Message, name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Usage: friend add <name>")
		return
	}

	target, ok := findFriendTarget(game, name)
	if !ok {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "No adventurer named '"+name+"'.")
		return
	}
	if target.ID == message.Character.ID {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You can't add yourself.")
		return
	}
	if targetUser, err := game.GetFacade().UsersService().FindByID(target.BelongsUserID); err == nil && targetUser != nil && targetUser.IsGuest {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Guest adventurers don't linger long enough for a friends list.")
		return
	}
	if message.Character.HasFriend(target.ID) {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, target.Name+" is already on your friends list.")
		command.pushList(game, message)
		return
	}

	err := game.GetFacade().CharactersService().Modify(message.Character.ID, func(ch *characters.Character) error {
		ch.AddFriendID(target.ID)
		return nil
	})
	if err != nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Could not update your friends list.")
		return
	}
	message.Character.AddFriendID(target.ID)
	game.SendMessage() <- messages.Reply(message.FromUser.ID, "Added "+target.Name+" to your friends list.")
	command.pushList(game, message)
}

func (command *FriendCommand) remove(game def.GameCtrl, message *messages.Message, name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Usage: friend remove <name>")
		return
	}

	targetID, targetName, ok := resolveOwnedFriend(game, message.Character, name)
	if !ok {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "'"+name+"' is not on your friends list.")
		return
	}

	err := game.GetFacade().CharactersService().Modify(message.Character.ID, func(ch *characters.Character) error {
		ch.RemoveFriendID(targetID)
		return nil
	})
	if err != nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Could not update your friends list.")
		return
	}
	message.Character.RemoveFriendID(targetID)
	game.SendMessage() <- messages.Reply(message.FromUser.ID, "Removed "+targetName+" from your friends list.")
	command.pushList(game, message)
}

func (command *FriendCommand) list(game def.GameCtrl, message *messages.Message) {
	entries := collectFriendEntries(game, message.Character)
	if len(entries) == 0 {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You have no friends yet. Use: friend add <name>")
		command.pushList(game, message)
		return
	}

	var b strings.Builder
	b.WriteString("Friends:\n")
	for _, entry := range entries {
		status := "Offline"
		if entry.Online {
			status = "Online"
		}
		fmt.Fprintf(&b, "  %s [%s]\n", entry.Name, status)
	}
	game.SendMessage() <- messages.Reply(message.FromUser.ID, strings.TrimRight(b.String(), "\n"))
	command.pushList(game, message)
}

func (command *FriendCommand) pushList(game def.GameCtrl, message *messages.Message) {
	game.SendMessage() <- messages.NewFriendsMessage(message.FromUser.ID, collectFriendEntries(game, message.Character))
}

func findFriendTarget(game def.GameCtrl, name string) (*characters.Character, bool) {
	if player, ok := game.FindOnlinePlayerByName(name); ok {
		ch, err := game.GetFacade().CharactersService().FindByID(player.CharacterID)
		if err == nil && ch != nil {
			return ch, true
		}
	}

	chars, err := game.GetFacade().CharactersService().FindByName(name)
	if err != nil {
		return nil, false
	}
	for _, ch := range chars {
		if ch != nil && strings.EqualFold(ch.Name, name) {
			return ch, true
		}
	}
	return nil, false
}

func resolveOwnedFriend(game def.GameCtrl, owner *characters.Character, name string) (id, display string, ok bool) {
	if owner == nil {
		return "", "", false
	}
	if player, found := game.FindOnlinePlayerByName(name); found && owner.HasFriend(player.CharacterID) {
		return player.CharacterID, player.CharacterName, true
	}
	for _, id := range owner.FriendIDs {
		ch, err := game.GetFacade().CharactersService().FindByID(id)
		if err != nil || ch == nil {
			continue
		}
		if strings.EqualFold(ch.Name, name) {
			return ch.ID, ch.Name, true
		}
	}
	return "", "", false
}

func collectFriendEntries(game def.GameCtrl, owner *characters.Character) []messages.FriendEntry {
	if owner == nil || len(owner.FriendIDs) == 0 {
		return []messages.FriendEntry{}
	}
	online := map[string]bool{}
	for _, player := range game.GetOnlinePlayers() {
		if player.CharacterID != "" {
			online[player.CharacterID] = true
		}
	}
	entries := make([]messages.FriendEntry, 0, len(owner.FriendIDs))
	for _, id := range owner.FriendIDs {
		ch, err := game.GetFacade().CharactersService().FindByID(id)
		if err != nil || ch == nil {
			continue
		}
		entries = append(entries, messages.FriendEntry{
			ID:     ch.ID,
			Name:   ch.Name,
			Online: online[ch.ID],
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Online != entries[j].Online {
			return entries[i].Online
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries
}
