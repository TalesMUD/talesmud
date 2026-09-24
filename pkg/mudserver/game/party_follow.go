package game

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/worldmap"
)

// PullPartyFollowers moves online party members who are following leader into
// destRoomID. v1 follows the party leader only, and only after a normal exit
// walk (allow=true). Teleports, portals, bindstones, scripts, and private
// instance crossings pass allow=false and do not pull anyone.
// Followers who are offline are left for a later step. Followers in combat
// stay behind. Membership is checked again so a kick or leader change cannot
// yank someone who is no longer following that leader.
func (g *Game) PullPartyFollowers(leader *characters.Character, destRoomID string, allow bool) {
	if g == nil || g.Sessions == nil || !allow || leader == nil || destRoomID == "" {
		return
	}
	followerIDs := g.Sessions.followersOf(leader.ID)
	if len(followerIDs) == 0 {
		return
	}
	if g.Facade == nil {
		return
	}

	party, err := g.Facade.PartiesService().FindByCharacterID(leader.ID)
	if err != nil || party == nil {
		g.Sessions.dropFollowersOf(leader.ID)
		return
	}
	party.EnsureLeader()
	if party.LeaderCharacterID != leader.ID {
		stopped := g.Sessions.dropFollowersOf(leader.ID)
		g.tellPartyFollowers(stopped, "[Party] You stop following. The party leader changed.")
		return
	}
	member := map[string]bool{}
	for _, id := range party.Characters {
		member[id] = true
	}

	dest, err := g.Facade.RoomsService().FindByID(destRoomID)
	if err != nil || dest == nil {
		return
	}

	onlineUser := map[string]string{}
	for _, player := range g.GetOnlinePlayers() {
		if player.CharacterID != "" && player.UserID != "" {
			onlineUser[player.CharacterID] = player.UserID
		}
	}
	leaderUserID := onlineUser[leader.ID]
	if leaderUserID == "" {
		leaderUserID = leader.BelongsUserID
	}
	leaderName := leader.Name
	if leaderName == "" {
		leaderName = "your leader"
	}

	for _, followerID := range followerIDs {
		if followerID == "" || followerID == leader.ID || !member[followerID] {
			g.Sessions.clearFollow(followerID)
			continue
		}
		userID := onlineUser[followerID]
		if userID == "" {
			continue
		}
		char, err := g.Facade.CharactersService().FindByID(followerID)
		if err != nil || char == nil {
			g.Sessions.clearFollow(followerID)
			continue
		}
		if g.CombatController != nil && g.CombatController.IsPlayerInCombat(char.ID) {
			g.SendMessage() <- messages.Reply(userID, "[Party] You stay behind. You are in combat.")
			continue
		}
		if char.CurrentRoomID == destRoomID {
			continue
		}

		g.InterruptRest(char)
		moved, ok := g.RelocateCharacter(char, userID, destRoomID)
		if !ok || moved == nil {
			continue
		}
		g.Sessions.updateRoom(char.ID, moved.ID)
		if fresh, ferr := g.Facade.CharactersService().FindByID(char.ID); ferr == nil && fresh != nil {
			char = fresh
		}
		if moved.Area != "" {
			g.Facade.QuestsService().GrantAutoQuests(char.ID, moved.Area)
		}
		if g.QuestTracker != nil {
			g.QuestTracker.OnRoomEnter(char.ID, userID, moved)
		}
		g.pushFollowerAtlas(userID, char)

		roomName := moved.Name
		if roomName == "" {
			roomName = "the next room"
		}
		g.SendMessage() <- messages.Reply(userID, fmt.Sprintf("[Party] You follow %s into %s.", leaderName, roomName))
		if leaderUserID != "" && leaderUserID != userID {
			g.SendMessage() <- messages.Reply(leaderUserID, fmt.Sprintf("[Party] %s follows you.", char.Name))
		}
		log.WithFields(log.Fields{
			"leader":   leader.ID,
			"follower": char.ID,
			"room":     moved.ID,
		}).Info("Party follow move")
	}
}

func (g *Game) tellPartyFollowers(characterIDs []string, text string) {
	if g == nil || text == "" || len(characterIDs) == 0 {
		return
	}
	onlineUser := map[string]string{}
	for _, player := range g.GetOnlinePlayers() {
		if player.CharacterID != "" && player.UserID != "" {
			onlineUser[player.CharacterID] = player.UserID
		}
	}
	sent := map[string]bool{}
	for _, id := range characterIDs {
		userID := onlineUser[id]
		if userID == "" || sent[userID] {
			continue
		}
		sent[userID] = true
		g.SendMessage() <- messages.Reply(userID, text)
	}
}

func (g *Game) pushFollowerAtlas(userID string, character *characters.Character) {
	if g == nil || g.Facade == nil || character == nil || userID == "" {
		return
	}
	rooms, err := g.Facade.RoomsService().FindAll()
	if err != nil {
		log.WithError(err).Warn("Party follow: failed to load rooms for atlas")
		return
	}
	atlas := worldmap.Reveal(worldmap.Compile(rooms), character)
	if npcs, nerr := g.Facade.NPCsService().FindAll(); nerr == nil {
		worldmap.AttachResidents(&atlas, worldmap.ResidentsFromNPCs(npcs))
	}
	g.SendMessage() <- messages.NewAtlasMessage(userID, atlas)
}
