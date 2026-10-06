package game

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/instances"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/worldmap"
)

// partyFollowChaseLimit is how many ordinary rooms a straggler may walk in
// one catch-up. Longer gaps wait for the next leader step or reconnect.
const partyFollowChaseLimit = 12

// PullPartyFollowers moves online party members who are following leader.
// allow is false for teleports, portals, bindstones, scripts, and private
// instance crossings. Followers standing in fromRoomID take that same step.
// Followers left behind (combat, a late follow, a reconnect) walk a path of
// ordinary exits toward destRoomID instead of being yanked across the map.
// Offline followers keep the follow flag. Followers in combat stay behind.
// Membership is checked again so a kick or leader change cannot yank someone
// who is no longer following that leader.
func (g *Game) PullPartyFollowers(leader *characters.Character, fromRoomID, destRoomID string, allow bool) {
	if g == nil || g.Sessions == nil || !allow || leader == nil || destRoomID == "" {
		return
	}
	followerIDs := g.Sessions.followersOf(leader.ID)
	if len(followerIDs) == 0 {
		return
	}
	if !g.partyFollowLeaderOK(leader) {
		return
	}
	member := g.partyFollowMembers(leader.ID)
	if member == nil {
		return
	}

	var behind []string
	for _, followerID := range followerIDs {
		if followerID == "" || followerID == leader.ID || !member[followerID] {
			g.Sessions.clearFollow(followerID)
			continue
		}
		userID := g.onlineUserID(followerID)
		if userID == "" {
			continue
		}
		char, err := g.Facade.CharactersService().FindByID(followerID)
		if err != nil || char == nil {
			g.Sessions.clearFollow(followerID)
			continue
		}
		if g.followerInCombat(char.ID) {
			g.SendMessage() <- messages.Reply(userID, "[Party] You stay behind. You are in combat.")
			continue
		}
		if char.CurrentRoomID == destRoomID {
			continue
		}
		if fromRoomID != "" && char.CurrentRoomID == fromRoomID {
			g.movePartyFollower(leader, char, userID, destRoomID)
			continue
		}
		behind = append(behind, followerID)
	}
	for _, followerID := range behind {
		g.chasePartyFollower(followerID, false)
	}
}

// CatchUpPartyFollow walks characterID toward the leader they are following,
// and walks anyone following characterID toward that character. The follow
// flag is left in place when the path is blocked, the leader is offline, or
// someone is in combat. Safe to call on login, after combat, and after follow.
func (g *Game) CatchUpPartyFollow(characterID string) {
	if g == nil || characterID == "" {
		return
	}
	g.chasePartyFollower(characterID, true)
	if g.Sessions == nil {
		return
	}
	for _, followerID := range g.Sessions.followersOf(characterID) {
		g.chasePartyFollower(followerID, true)
	}
}

func (g *Game) chasePartyFollower(followerID string, announceGap bool) {
	if g == nil || g.Sessions == nil || g.Facade == nil || followerID == "" {
		return
	}
	leaderID, ok := g.Sessions.followTarget(followerID)
	if !ok || leaderID == "" || leaderID == followerID {
		return
	}
	if !g.partyFollowPairOK(followerID, leaderID) {
		return
	}
	if g.onlineUserID(leaderID) == "" || g.onlineUserID(followerID) == "" {
		return
	}
	if g.followerInCombat(followerID) {
		return
	}
	follower, err := g.Facade.CharactersService().FindByID(followerID)
	if err != nil || follower == nil {
		return
	}
	leader, err := g.Facade.CharactersService().FindByID(leaderID)
	if err != nil || leader == nil {
		return
	}
	if follower.CurrentRoomID == "" || follower.CurrentRoomID == leader.CurrentRoomID {
		return
	}
	if g.RoomInstances != nil && g.RoomInstances.IsClone(follower.CurrentRoomID) {
		return
	}
	path := g.partyFollowPath(follower.CurrentRoomID, leader.CurrentRoomID)
	if len(path) == 0 {
		if announceGap {
			userID := g.onlineUserID(followerID)
			if userID != "" {
				g.SendMessage() <- messages.Reply(userID, "[Party] You can't find a way to "+leaderDisplayName(leader)+". You are still following.")
			}
		}
		return
	}
	for _, roomID := range path {
		if g.onlineUserID(followerID) == "" || g.onlineUserID(leaderID) == "" {
			return
		}
		currentLeader, still := g.Sessions.followTarget(followerID)
		if !still || currentLeader != leaderID || !g.partyFollowPairOK(followerID, leaderID) {
			return
		}
		if g.followerInCombat(followerID) {
			userID := g.onlineUserID(followerID)
			if userID != "" {
				g.SendMessage() <- messages.Reply(userID, "[Party] You stay behind. You are in combat.")
			}
			return
		}
		fresh, ferr := g.Facade.CharactersService().FindByID(followerID)
		if ferr != nil || fresh == nil {
			return
		}
		if fresh.CurrentRoomID == roomID {
			continue
		}
		leaderNow, lerr := g.Facade.CharactersService().FindByID(leaderID)
		if lerr != nil || leaderNow == nil {
			return
		}
		if !g.movePartyFollower(leaderNow, fresh, g.onlineUserID(followerID), roomID) {
			return
		}
	}
}

func (g *Game) movePartyFollower(leader, char *characters.Character, userID, destRoomID string) bool {
	if g == nil || char == nil || userID == "" || destRoomID == "" || char.CurrentRoomID == destRoomID {
		return false
	}
	g.InterruptRest(char)
	moved, ok := g.RelocateCharacter(char, userID, destRoomID)
	if !ok || moved == nil {
		return false
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

	leaderName := "your leader"
	leaderUserID := ""
	if leader != nil {
		leaderName = leaderDisplayName(leader)
		leaderUserID = g.onlineUserID(leader.ID)
		if leaderUserID == "" {
			leaderUserID = leader.BelongsUserID
		}
	}
	roomName := moved.Name
	if roomName == "" {
		roomName = "the next room"
	}
	g.SendMessage() <- messages.Reply(userID, fmt.Sprintf("[Party] You follow %s into %s.", leaderName, roomName))
	if leaderUserID != "" && leaderUserID != userID && char.Name != "" {
		g.SendMessage() <- messages.Reply(leaderUserID, fmt.Sprintf("[Party] %s follows you.", char.Name))
	}
	log.WithFields(log.Fields{
		"leader":   leaderIDOf(leader),
		"follower": char.ID,
		"room":     moved.ID,
	}).Info("Party follow move")
	return true
}

func (g *Game) partyFollowPath(fromID, toID string) []string {
	if fromID == "" || toID == "" || fromID == toID || g == nil || g.Facade == nil {
		return nil
	}
	type node struct {
		id   string
		path []string
	}
	queue := []node{{id: fromID}}
	seen := map[string]bool{fromID: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if len(cur.path) >= partyFollowChaseLimit {
			continue
		}
		room, err := g.Facade.RoomsService().FindByID(cur.id)
		if err != nil || room == nil || room.Exits == nil {
			continue
		}
		for _, ex := range *room.Exits {
			if !followExitWalkable(ex) {
				continue
			}
			nextID := strings.TrimSpace(ex.Target)
			if nextID == "" || seen[nextID] {
				continue
			}
			if g.RoomInstances != nil && g.RoomInstances.IsClone(nextID) {
				continue
			}
			next, nerr := g.Facade.RoomsService().FindByID(nextID)
			if nerr != nil || next == nil || next.IsInstanceTemplate() {
				continue
			}
			path := append(append([]string{}, cur.path...), nextID)
			if nextID == toID {
				return path
			}
			seen[nextID] = true
			queue = append(queue, node{id: nextID, path: path})
		}
	}
	return nil
}

func followExitWalkable(ex rooms.Exit) bool {
	switch strings.ToLower(strings.TrimSpace(string(ex.Type))) {
	case "", "normal", "direction":
	default:
		return false
	}
	if ex.Hidden || ex.Instance || instances.IsInstanceEntrance(ex) {
		return false
	}
	return strings.TrimSpace(ex.Target) != ""
}

func (g *Game) partyFollowLeaderOK(leader *characters.Character) bool {
	if g == nil || g.Facade == nil || leader == nil {
		return false
	}
	party, err := g.Facade.PartiesService().FindByCharacterID(leader.ID)
	if err != nil || party == nil {
		g.Sessions.dropFollowersOf(leader.ID)
		return false
	}
	party.EnsureLeader()
	if party.LeaderCharacterID != leader.ID {
		stopped := g.Sessions.dropFollowersOf(leader.ID)
		g.tellPartyFollowers(stopped, "[Party] You stop following. The party leader changed.")
		return false
	}
	return true
}

func (g *Game) partyFollowPairOK(followerID, leaderID string) bool {
	if g == nil || g.Facade == nil {
		return false
	}
	party, err := g.Facade.PartiesService().FindByCharacterID(followerID)
	if err != nil || party == nil || !partyHas(party.Characters, followerID) {
		g.Sessions.clearFollow(followerID)
		return false
	}
	party.EnsureLeader()
	if party.LeaderCharacterID != leaderID || !partyHas(party.Characters, leaderID) {
		g.Sessions.clearFollow(followerID)
		return false
	}
	return true
}

func partyHas(ids []string, id string) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}

func (g *Game) partyFollowMembers(leaderID string) map[string]bool {
	party, err := g.Facade.PartiesService().FindByCharacterID(leaderID)
	if err != nil || party == nil {
		g.Sessions.dropFollowersOf(leaderID)
		return nil
	}
	member := map[string]bool{}
	for _, id := range party.Characters {
		member[id] = true
	}
	return member
}

func (g *Game) followerInCombat(characterID string) bool {
	return g != nil && g.CombatController != nil && g.CombatController.IsPlayerInCombat(characterID)
}

func (g *Game) onlineUserID(characterID string) string {
	if g == nil || characterID == "" {
		return ""
	}
	for _, player := range g.GetOnlinePlayers() {
		if player.CharacterID == characterID && player.UserID != "" {
			return player.UserID
		}
	}
	return ""
}

func leaderDisplayName(leader *characters.Character) string {
	if leader == nil || leader.Name == "" {
		return "your leader"
	}
	return leader.Name
}

func leaderIDOf(leader *characters.Character) string {
	if leader == nil {
		return ""
	}
	return leader.ID
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
	roomsList, err := g.Facade.RoomsService().FindAll()
	if err != nil {
		log.WithError(err).Warn("Party follow: failed to load rooms for atlas")
		return
	}
	atlas := worldmap.Reveal(worldmap.Compile(roomsList), character)
	if npcs, nerr := g.Facade.NPCsService().FindAll(); nerr == nil {
		worldmap.AttachResidents(&atlas, worldmap.ResidentsFromNPCs(npcs))
	}
	g.SendMessage() <- messages.NewAtlasMessage(userID, atlas)
}
