package game

import (
	"sort"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/instances"
)

const liveRecentWindow = 30 * 24 * time.Hour

// LiveQuery filters the character list. Zero values leave that dimension open.
type LiveQuery struct {
	OnlineOnly   bool
	GuestOnly    bool
	InCombatOnly bool
	Zone         string
	All          bool
}

// LiveCharacter is one row in the operator character list.
type LiveCharacter struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Guest     bool      `json:"guest"`
	Level     int32     `json:"level"`
	ClassID   string    `json:"classId"`
	RoomID    string    `json:"roomId"`
	RoomName  string    `json:"roomName"`
	Zone      string    `json:"zone"`
	InCombat  bool      `json:"inCombat"`
	PartyID   string    `json:"partyId,omitempty"`
	PartyName string    `json:"partyName,omitempty"`
	Online    bool      `json:"online"`
	LastSeen  time.Time `json:"lastSeen,omitempty"`
}

// LiveCharacterDetail is the side drawer.
type LiveCharacterDetail struct {
	LiveCharacter
	Inventory     interface{} `json:"inventory"`
	EquippedItems interface{} `json:"equippedItems"`
	Quests        interface{} `json:"quests"`
	RevealedExits interface{} `json:"revealedExits"`
	Combat        interface{} `json:"combat,omitempty"`
}

// LiveNPC is one running NPC instance.
type LiveNPC struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	TemplateID string `json:"templateId"`
	RoomID     string `json:"roomId"`
	HP         int32  `json:"hp"`
	MaxHP      int32  `json:"maxHp"`
	InCombat   bool   `json:"inCombat"`
	Dead       bool   `json:"dead"`
	LastEvent  string `json:"lastEvent,omitempty"`
}

// LiveInstance is one instance copy.
type LiveInstance struct {
	ID         string    `json:"id"`
	SourceRoom string    `json:"sourceRoom"`
	HubRoomID  string    `json:"hubRoomId,omitempty"`
	Players    []string  `json:"players"`
	CloneIDs   []string  `json:"cloneIds"`
	Created    time.Time `json:"created,omitempty"`
}

// LiveCharacters lists online characters and recent offline ones.
func (g *Game) LiveCharacters(q LiveQuery) ([]LiveCharacter, error) {
	if g == nil || g.Facade == nil {
		return nil, opErr(503, "game is not running")
	}
	chars, err := g.Facade.CharactersService().FindAll()
	if err != nil {
		return nil, err
	}
	users := indexUsers(g)
	rooms := indexRooms(g)
	parties := indexParties(g)
	online := g.onlineCharacters()
	now := time.Now()
	out := make([]LiveCharacter, 0, len(chars))
	for _, ch := range chars {
		if ch == nil || ch.ID == "" {
			continue
		}
		row := g.liveRow(ch, users[ch.BelongsUserID], rooms[ch.CurrentRoomID], parties[ch.ID], online[ch.ID], now)
		if !q.All && !row.Online && !recentOffline(users[ch.BelongsUserID], now) {
			continue
		}
		if q.OnlineOnly && !row.Online {
			continue
		}
		if q.GuestOnly && !row.Guest {
			continue
		}
		if q.InCombatOnly && !row.InCombat {
			continue
		}
		if q.Zone != "" && row.Zone != q.Zone && !hasPrefix(row.Zone, q.Zone) {
			continue
		}
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Online != out[j].Online {
			return out[i].Online
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// LiveCharacter returns one character's operator detail.
func (g *Game) LiveCharacter(id string) (*LiveCharacterDetail, error) {
	if g == nil || g.Facade == nil {
		return nil, opErr(503, "game is not running")
	}
	ch, err := g.Facade.CharactersService().FindByID(id)
	if err != nil || ch == nil {
		return nil, opErr(404, "character not found")
	}
	users := indexUsers(g)
	rooms := indexRooms(g)
	parties := indexParties(g)
	row := g.liveRow(ch, users[ch.BelongsUserID], rooms[ch.CurrentRoomID], parties[ch.ID], g.onlineCharacters()[ch.ID], time.Now())
	var quests interface{}
	if log, qerr := g.Facade.QuestsService().BuildQuestLog(id); qerr == nil {
		quests = log
	}
	detail := &LiveCharacterDetail{
		LiveCharacter: row,
		Inventory:     ch.Inventory,
		EquippedItems: ch.EquippedItems,
		Quests:        quests,
		RevealedExits: ch.RevealedExits,
	}
	if g.CombatController != nil {
		if inst := g.CombatController.GetCombatInstance(id); inst != nil {
			enemies := make([]map[string]interface{}, 0, len(inst.Enemies))
			for _, enemy := range inst.Enemies {
				enemies = append(enemies, map[string]interface{}{
					"id":   enemy.ID,
					"name": enemy.Name,
					"hp":   enemy.CurrentHP,
					"max":  enemy.MaxHP,
				})
			}
			detail.Combat = map[string]interface{}{
				"id":      inst.ID,
				"roomId":  inst.OriginRoomID,
				"state":   inst.State,
				"enemies": enemies,
			}
			detail.InCombat = true
		}
	}
	return detail, nil
}

// LiveNPCs lists running NPC instances. templateId and roomId narrow the list.
func (g *Game) LiveNPCs(templateID, roomID string) ([]LiveNPC, error) {
	if g == nil || g.NPCManager == nil {
		return []LiveNPC{}, nil
	}
	out := make([]LiveNPC, 0)
	for _, inst := range g.NPCManager.GetAllInstances() {
		if inst == nil || inst.Entity == nil {
			continue
		}
		tpl := inst.TemplateID
		if tpl == "" {
			tpl = inst.ID
		}
		if templateID != "" && tpl != templateID && inst.ID != templateID {
			continue
		}
		if roomID != "" && inst.CurrentRoomID != roomID {
			continue
		}
		inCombat := inst.InCombat
		if g.CombatController != nil && g.CombatController.IsNPCInCombat(inst.ID) {
			inCombat = true
		}
		out = append(out, LiveNPC{
			ID:         inst.ID,
			Name:       inst.Name,
			TemplateID: tpl,
			RoomID:     inst.CurrentRoomID,
			HP:         inst.CurrentHitPoints,
			MaxHP:      inst.MaxHitPoints,
			InCombat:   inCombat,
			Dead:       inst.IsDead,
			LastEvent:  g.NPCManager.LastNote(inst.ID),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// LiveInstances lists in-memory copies and database copies the process still has.
func (g *Game) LiveInstances() ([]LiveInstance, error) {
	if g == nil || g.Facade == nil {
		return nil, opErr(503, "game is not running")
	}
	byID := map[string]*LiveInstance{}
	if g.RoomInstances != nil && g.RoomInstances.mgr != nil {
		for _, copyRow := range g.RoomInstances.mgr.List() {
			src := ""
			if len(copyRow.SourceIDs) > 0 {
				src = copyRow.SourceIDs[0]
			}
			byID[copyRow.ID] = &LiveInstance{
				ID:         copyRow.ID,
				SourceRoom: src,
				HubRoomID:  copyRow.HubRoomID,
				Players:    append([]string{}, copyRow.PlayerIDs...),
				CloneIDs:   append([]string{}, copyRow.CloneIDs...),
				Created:    copyRow.Created,
			}
		}
	}
	all, err := loadRoomMap(g.Facade.RoomsService())
	if err == nil {
		for id := range all {
			if !instances.IsCloneID(id) {
				continue
			}
			instID := instances.InstanceID(id)
			if instID == "" {
				continue
			}
			row := byID[instID]
			if row == nil {
				row = &LiveInstance{ID: instID, SourceRoom: instances.TemplateIDFromClone(id)}
				byID[instID] = row
			}
			if !containsString(row.CloneIDs, id) {
				row.CloneIDs = append(row.CloneIDs, id)
			}
		}
	}
	chars, _ := g.Facade.CharactersService().FindAll()
	for _, ch := range chars {
		if ch == nil || !instances.IsCloneID(ch.CurrentRoomID) {
			continue
		}
		instID := instances.InstanceID(ch.CurrentRoomID)
		row := byID[instID]
		if row == nil || containsString(row.Players, ch.ID) {
			continue
		}
		row.Players = append(row.Players, ch.ID)
	}
	out := make([]LiveInstance, 0, len(byID))
	for _, row := range byID {
		sort.Strings(row.CloneIDs)
		sort.Strings(row.Players)
		out = append(out, *row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (g *Game) liveRow(ch *characters.Character, user *entities.User, room *rooms.Room, party *entities.Party, online bool, now time.Time) LiveCharacter {
	row := LiveCharacter{
		ID:       ch.ID,
		Name:     ch.Name,
		UserID:   ch.BelongsUserID,
		Level:    ch.Level,
		ClassID:  ch.Class.ID,
		RoomID:   ch.CurrentRoomID,
		InCombat: ch.InCombat,
		Online:   online,
	}
	if g.CombatController != nil && g.CombatController.IsPlayerInCombat(ch.ID) {
		row.InCombat = true
	}
	if user != nil {
		row.UserName = userDisplay(user)
		row.Guest = user.IsGuest
		row.LastSeen = user.LastSeen
		if user.IsOnline {
			row.Online = true
		}
	}
	if online {
		row.Online = true
		if row.LastSeen.IsZero() {
			row.LastSeen = now
		}
	}
	if room != nil {
		row.RoomName = room.Name
		row.Zone = room.Area
	}
	if party != nil && party.Entity != nil {
		row.PartyID = party.ID
		row.PartyName = party.Name
	}
	return row
}

func (g *Game) onlineCharacters() map[string]bool {
	out := map[string]bool{}
	if g == nil || g.Sessions == nil {
		return out
	}
	for _, player := range g.Sessions.all() {
		if player.CharacterID != "" {
			out[player.CharacterID] = true
		}
	}
	return out
}

func recentOffline(user *entities.User, now time.Time) bool {
	if user == nil {
		return true
	}
	if user.LastSeen.IsZero() && user.Created.IsZero() {
		return true
	}
	cutoff := now.Add(-liveRecentWindow)
	if !user.LastSeen.IsZero() {
		return user.LastSeen.After(cutoff)
	}
	return user.Created.After(cutoff)
}

func indexUsers(g *Game) map[string]*entities.User {
	out := map[string]*entities.User{}
	if g.Facade.UsersService() == nil {
		return out
	}
	users, err := g.Facade.UsersService().FindAll()
	if err != nil {
		return out
	}
	for _, user := range users {
		if user != nil && user.ID != "" {
			out[user.ID] = user
		}
	}
	return out
}

func indexRooms(g *Game) map[string]*rooms.Room {
	out := map[string]*rooms.Room{}
	all, err := loadRoomMap(g.Facade.RoomsService())
	if err != nil {
		return out
	}
	return all
}

func indexParties(g *Game) map[string]*entities.Party {
	out := map[string]*entities.Party{}
	if g.Facade.PartiesService() == nil {
		return out
	}
	parties, err := g.Facade.PartiesService().FindAll()
	if err != nil {
		return out
	}
	for _, party := range parties {
		if party == nil {
			continue
		}
		for _, id := range party.Characters {
			out[id] = party
		}
	}
	return out
}

func userDisplay(user *entities.User) string {
	if user.Nickname != "" {
		return user.Nickname
	}
	if user.Name != "" {
		return user.Name
	}
	if user.Username != "" {
		return user.Username
	}
	return user.ID
}

func hasPrefix(value, prefix string) bool {
	return len(prefix) > 0 && len(value) >= len(prefix) && value[:len(prefix)] == prefix
}

func containsString(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}
