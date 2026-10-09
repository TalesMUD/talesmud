package game

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

type sessionRegistry struct {
	mu      sync.RWMutex
	players map[string]def.OnlinePlayer
	invites map[string]def.PartyInvite
	follows map[string]string // follower character ID -> leader character ID
}

func newSessionRegistry() *sessionRegistry {
	return &sessionRegistry{
		players: make(map[string]def.OnlinePlayer),
		invites: make(map[string]def.PartyInvite),
		follows: make(map[string]string),
	}
}

func (r *sessionRegistry) connect(user *entities.User) {
	if user == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	current := r.players[user.ID]
	current.UserID = user.ID
	current.LastSeen = time.Now()
	r.players[user.ID] = current
}

func (r *sessionRegistry) characterID(userID string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.players[userID].CharacterID
}

func (r *sessionRegistry) get(userID string) (def.OnlinePlayer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	player, ok := r.players[userID]
	if !ok || player.CharacterID == "" {
		return def.OnlinePlayer{}, false
	}
	return player, true
}

func (r *sessionRegistry) disconnect(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.players, userID)
	for targetID, invite := range r.invites {
		if invite.InviterUserID == userID {
			delete(r.invites, targetID)
		}
	}
}

func (r *sessionRegistry) setRoom(characterID, roomID string) {
	if r == nil || characterID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, player := range r.players {
		if player.CharacterID != characterID {
			continue
		}
		player.RoomID = roomID
		player.LastSeen = time.Now()
		r.players[id] = player
	}
}

func (r *sessionRegistry) setCharacter(user *entities.User, char *characters.Character) {
	if user == nil || char == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.players[user.ID] = def.OnlinePlayer{
		UserID:        user.ID,
		CharacterID:   char.ID,
		CharacterName: char.Name,
		RoomID:        char.CurrentRoomID,
		LastSeen:      time.Now(),
	}
}

func (r *sessionRegistry) all() []def.OnlinePlayer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]def.OnlinePlayer, 0, len(r.players))
	for _, player := range r.players {
		if player.CharacterID == "" {
			continue
		}
		players = append(players, player)
	}
	sort.Slice(players, func(i, j int) bool {
		return strings.ToLower(players[i].CharacterName) < strings.ToLower(players[j].CharacterName)
	})
	return players
}

func (r *sessionRegistry) roomPlayers(roomID, viewerCharacterID string) []def.OnlinePlayer {
	players := r.all()
	result := make([]def.OnlinePlayer, 0, len(players))
	for _, player := range players {
		if player.RoomID != roomID {
			continue
		}
		player.IsYou = player.CharacterID == viewerCharacterID
		result = append(result, player)
	}
	return result
}

func (r *sessionRegistry) findByName(name string) (def.OnlinePlayer, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return def.OnlinePlayer{}, false
	}
	for _, player := range r.all() {
		if strings.ToLower(player.CharacterName) == name {
			return player, true
		}
	}
	return def.OnlinePlayer{}, false
}

func (r *sessionRegistry) setInvite(invite def.PartyInvite) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if invite.CreatedAt.IsZero() {
		invite.CreatedAt = time.Now()
	}
	r.invites[invite.TargetCharacterID] = invite
}

func (r *sessionRegistry) getInvite(targetCharacterID string) (def.PartyInvite, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	invite, ok := r.invites[targetCharacterID]
	if !ok {
		return def.PartyInvite{}, false
	}
	if inviteExpired(invite, time.Now()) {
		delete(r.invites, targetCharacterID)
		return def.PartyInvite{}, false
	}
	return invite, true
}

func (r *sessionRegistry) clearInvite(targetCharacterID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.invites, targetCharacterID)
}

func inviteExpired(invite def.PartyInvite, now time.Time) bool {
	if invite.CreatedAt.IsZero() {
		return false
	}
	return now.Sub(invite.CreatedAt) >= def.PartyInviteTTL
}

// takeExpiredInvites removes and returns invites past PartyInviteTTL.
func (r *sessionRegistry) takeExpiredInvites(now time.Time) []def.PartyInvite {
	r.mu.Lock()
	defer r.mu.Unlock()

	expired := make([]def.PartyInvite, 0)
	for targetID, invite := range r.invites {
		if inviteExpired(invite, now) {
			expired = append(expired, invite)
			delete(r.invites, targetID)
		}
	}
	return expired
}

func (g *Game) ConnectUserSession(user *entities.User) {
	if user == nil {
		return
	}
	user.IsOnline = true
	user.LastSeen = time.Now()
	if user.RefID != "" {
		_ = g.Facade.UsersService().Update(user.RefID, user)
	}
	g.Sessions.connect(user)
}

func (g *Game) DisconnectUserSession(userID string) {
	if userID == "" {
		return
	}
	departed, _ := g.Sessions.get(userID)
	if user, err := g.Facade.UsersService().FindByID(userID); err == nil && user != nil {
		user.IsOnline = false
		user.LastSeen = time.Now()
		if user.RefID != "" {
			_ = g.Facade.UsersService().Update(user.RefID, user)
		}
	}
	if charID := g.Sessions.characterID(userID); charID != "" {
		g.cancelAggroForPlayer(charID)
		if ruleset.Disconnect() == ruleset.DisconnectRelease {
			g.ReleaseToSafety(charID)
		}
		if g.RoomInstances != nil {
			g.RoomInstances.DestroyCharacterInstance(charID)
		}
	}
	if departed.CharacterID != "" {
		name := departed.CharacterName
		if name == "" {
			name = "A party member"
		}
		if followers := g.Sessions.followersOf(departed.CharacterID); len(followers) > 0 {
			g.tellPartyFollowers(followers, "[Party] "+name+" disconnected. You are still following.")
		}
		if leaderID, ok := g.Sessions.followTarget(departed.CharacterID); ok {
			g.tellPartyFollowers([]string{leaderID}, "[Party] "+name+" disconnected. They are still following.")
		}
	}
	g.Sessions.disconnect(userID)
	if departed.CharacterID != "" {
		g.notifyFriendsOfPresence(departed.CharacterID, departed.CharacterName, false)
	}
}

func (g *Game) SetUserSessionCharacter(user *entities.User, char *characters.Character) {
	if user == nil || char == nil {
		return
	}
	user.LastCharacter = char.ID
	user.IsOnline = true
	user.LastSeen = time.Now()
	if user.RefID != "" {
		_ = g.Facade.UsersService().Update(user.RefID, user)
	}
	g.Sessions.setCharacter(user, char)
	g.notifyFriendsOfPresence(char.ID, char.Name, true)
}

func (g *Game) notifyFriendsOfPresence(characterID, characterName string, online bool) {
	if characterID == "" || characterName == "" {
		return
	}
	status := "came online"
	if !online {
		status = "went offline"
	}
	for _, player := range g.GetOnlinePlayers() {
		if player.CharacterID == "" || player.CharacterID == characterID {
			continue
		}
		watcher, err := g.Facade.CharactersService().FindByID(player.CharacterID)
		if err != nil || watcher == nil || !watcher.HasFriend(characterID) {
			continue
		}
		g.SendMessage() <- messages.Reply(player.UserID, characterName+" "+status+".")
	}
}

func (g *Game) GetOnlinePlayers() []def.OnlinePlayer {
	return g.Sessions.all()
}

func (g *Game) FindOnlinePlayerByName(name string) (def.OnlinePlayer, bool) {
	return g.Sessions.findByName(name)
}

func (g *Game) GetRoomPlayers(roomID, viewerCharacterID string) []def.OnlinePlayer {
	return g.Sessions.roomPlayers(roomID, viewerCharacterID)
}

func (g *Game) SetPartyInvite(invite def.PartyInvite) {
	g.Sessions.setInvite(invite)
}

func (g *Game) GetPartyInvite(targetCharacterID string) (def.PartyInvite, bool) {
	return g.Sessions.getInvite(targetCharacterID)
}

func (g *Game) ClearPartyInvite(targetCharacterID string) {
	g.Sessions.clearInvite(targetCharacterID)
}

// ExpirePartyInvites clears timed-out invites and notifies both sides.
func (g *Game) ExpirePartyInvites() {
	expired := g.Sessions.takeExpiredInvites(time.Now())
	for _, invite := range expired {
		targetUserID := invite.TargetUserID
		if targetUserID == "" {
			for _, player := range g.GetOnlinePlayers() {
				if player.CharacterID == invite.TargetCharacterID {
					targetUserID = player.UserID
					break
				}
			}
		}
		inviterName := invite.InviterCharacterName
		if inviterName == "" {
			inviterName = "Someone"
		}
		targetName := invite.TargetCharacterName
		if targetName == "" {
			targetName = "That player"
		}
		if targetUserID != "" {
			g.SendMessage() <- messages.Reply(targetUserID, "Party invite from "+inviterName+" expired.")
			g.SendMessage() <- messages.NewPartyInviteMessage(targetUserID, false, "", "")
		}
		if invite.InviterUserID != "" {
			g.SendMessage() <- messages.Reply(invite.InviterUserID, targetName+" did not respond to your party invite.")
		}
	}
}

func (r *sessionRegistry) setFollow(followerID, leaderID string) {
	if followerID == "" || leaderID == "" || followerID == leaderID {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.follows == nil {
		r.follows = make(map[string]string)
	}
	r.follows[followerID] = leaderID
}

func (r *sessionRegistry) clearFollow(followerID string) (string, bool) {
	if followerID == "" {
		return "", false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	leaderID, ok := r.follows[followerID]
	if ok {
		delete(r.follows, followerID)
	}
	return leaderID, ok
}

func (r *sessionRegistry) dropFollowersOf(leaderID string) []string {
	if leaderID == "" {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	var cleared []string
	for followerID, followed := range r.follows {
		if followed == leaderID {
			delete(r.follows, followerID)
			cleared = append(cleared, followerID)
		}
	}
	sort.Strings(cleared)
	return cleared
}

func (r *sessionRegistry) followTarget(followerID string) (string, bool) {
	if followerID == "" {
		return "", false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	leaderID, ok := r.follows[followerID]
	return leaderID, ok && leaderID != ""
}

func (r *sessionRegistry) followersOf(leaderID string) []string {
	if leaderID == "" {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ids []string
	for followerID, followed := range r.follows {
		if followed == leaderID {
			ids = append(ids, followerID)
		}
	}
	sort.Strings(ids)
	return ids
}

func (r *sessionRegistry) updateRoom(characterID, roomID string) {
	if characterID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for userID, player := range r.players {
		if player.CharacterID != characterID {
			continue
		}
		player.RoomID = roomID
		r.players[userID] = player
	}
}

func (g *Game) SetPartyFollow(followerID, leaderID string) {
	if g == nil || g.Sessions == nil {
		return
	}
	g.Sessions.setFollow(followerID, leaderID)
}

func (g *Game) ClearPartyFollow(followerID string) (string, bool) {
	if g == nil || g.Sessions == nil {
		return "", false
	}
	return g.Sessions.clearFollow(followerID)
}

func (g *Game) DropPartyFollowers(leaderID string) []string {
	if g == nil || g.Sessions == nil {
		return nil
	}
	return g.Sessions.dropFollowersOf(leaderID)
}

func (g *Game) PartyFollowTarget(followerID string) (string, bool) {
	if g == nil || g.Sessions == nil {
		return "", false
	}
	return g.Sessions.followTarget(followerID)
}
