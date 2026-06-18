package game

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
)

type sessionRegistry struct {
	mu      sync.RWMutex
	players map[string]def.OnlinePlayer
	invites map[string]def.PartyInvite
}

func newSessionRegistry() *sessionRegistry {
	return &sessionRegistry{
		players: make(map[string]def.OnlinePlayer),
		invites: make(map[string]def.PartyInvite),
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

	invite.CreatedAt = time.Now()
	r.invites[invite.TargetCharacterID] = invite
}

func (r *sessionRegistry) getInvite(targetCharacterID string) (def.PartyInvite, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	invite, ok := r.invites[targetCharacterID]
	return invite, ok
}

func (r *sessionRegistry) clearInvite(targetCharacterID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.invites, targetCharacterID)
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
	if user, err := g.Facade.UsersService().FindByID(userID); err == nil && user != nil {
		user.IsOnline = false
		user.LastSeen = time.Now()
		if user.RefID != "" {
			_ = g.Facade.UsersService().Update(user.RefID, user)
		}
	}
	g.Sessions.disconnect(userID)
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
