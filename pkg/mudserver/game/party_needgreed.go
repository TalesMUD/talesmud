package game

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

// Party Need/Greed v1.
//
// When a victory has 2+ living same-room online party members (same band as
// Party Loot & XP Share), each item drop opens a Need/Greed/Pass round instead
// of landing on the ground immediately. Gold and XP share are unchanged.
// Rounds queue one at a time. Need beats Greed beats Pass; ties are uniform
// random among the highest tier. All-pass or timeout with no Need/Greed leaves
// the item on the ground as free-for-all. Solo and non-party kills are unchanged.

// NeedGreedTTL is how long a loot roll stays open.
const NeedGreedTTL = 30 * time.Second

// lootRollChoice is one member's Need / Greed / Pass pick.
type lootRollChoice int

const (
	lootChoiceUnset lootRollChoice = iota
	lootChoicePass
	lootChoiceGreed
	lootChoiceNeed
)

func (c lootRollChoice) String() string {
	switch c {
	case lootChoiceNeed:
		return "Need"
	case lootChoiceGreed:
		return "Greed"
	case lootChoicePass:
		return "Pass"
	default:
		return ""
	}
}

func parseLootRollChoice(raw string) (lootRollChoice, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "need":
		return lootChoiceNeed, true
	case "greed":
		return lootChoiceGreed, true
	case "pass":
		return lootChoicePass, true
	default:
		return lootChoiceUnset, false
	}
}

// needGreedMember is one roller in an open round.
type needGreedMember struct {
	CharacterID string
	UserID      string
	Name        string
}

// needGreedRound is one item waiting on Need/Greed/Pass.
type needGreedRound struct {
	ID       string
	Item     *items.Item
	RoomID   string
	PartyID  string
	Members  []needGreedMember
	Choices  map[string]lootRollChoice
	OpenedAt time.Time
	Expires  time.Time
}

type needGreedBook struct {
	mu      sync.Mutex
	active  *needGreedRound
	pending []*needGreedRound
	seq     uint64
}

// resolveNeedGreedWinner picks among Need, then Greed. All-pass returns ok=false.
// rng(n) must return [0,n). When rng is nil, math/rand is used.
func resolveNeedGreedWinner(choices map[string]lootRollChoice, rng func(n int) int) (winnerID, tier string, ok bool) {
	if rng == nil {
		rng = rand.Intn
	}
	var needs, greeds []string
	for id, choice := range choices {
		if id == "" {
			continue
		}
		switch choice {
		case lootChoiceNeed:
			needs = append(needs, id)
		case lootChoiceGreed:
			greeds = append(greeds, id)
		}
	}
	pool := needs
	tier = "Need"
	if len(pool) == 0 {
		pool = greeds
		tier = "Greed"
	}
	if len(pool) == 0 {
		return "", "", false
	}
	sort.Strings(pool)
	return pool[rng(len(pool))], tier, true
}

// needGreedRollers returns party members eligible for item rolls on this victory.
// Same band as XP/gold share, restricted to a party that has 2+ recipients.
func (c *CombatController) needGreedRollers(instance *combat.CombatInstance, living []*combat.CombatantRef) ([]needGreedMember, string, bool) {
	if c == nil || c.game == nil || instance == nil || len(living) == 0 {
		return nil, "", false
	}
	shares, partySplit := c.planVictoryShares(instance, living, 0, 0)
	if !partySplit || len(shares) < 2 {
		return nil, "", false
	}

	memberParty := map[string]string{}
	for _, share := range shares {
		if share.ID == "" || c.game.Facade == nil {
			continue
		}
		party, err := c.game.Facade.PartiesService().FindByCharacterID(share.ID)
		if err != nil || party == nil {
			continue
		}
		memberParty[share.ID] = party.ID
	}

	counts := map[string]int{}
	for _, share := range shares {
		if pid := memberParty[share.ID]; pid != "" {
			counts[pid]++
		}
	}
	partyID := ""
	for pid, n := range counts {
		if n >= 2 {
			partyID = pid
			break
		}
	}
	if partyID == "" {
		return nil, "", false
	}

	onlineUser := map[string]string{}
	for _, op := range c.game.GetOnlinePlayers() {
		if op.CharacterID != "" {
			onlineUser[op.CharacterID] = op.UserID
		}
	}

	rollers := make([]needGreedMember, 0, len(shares))
	for _, share := range shares {
		if memberParty[share.ID] != partyID {
			continue
		}
		char, err := c.game.Facade.CharactersService().FindByID(share.ID)
		if err != nil || char == nil {
			continue
		}
		if char.CurrentHitPoints <= 0 {
			continue
		}
		userID := onlineUser[share.ID]
		if userID == "" {
			userID = char.BelongsUserID
		}
		if userID == "" || onlineUser[share.ID] == "" {
			continue
		}
		name := share.Name
		if name == "" {
			name = char.Name
		}
		rollers = append(rollers, needGreedMember{
			CharacterID: share.ID,
			UserID:      userID,
			Name:        name,
		})
	}
	if len(rollers) < 2 {
		return nil, "", false
	}
	return rollers, partyID, true
}

// enqueueNeedGreedItems queues one round per item and opens the first.
func (g *Game) enqueueNeedGreedItems(roomID, partyID string, members []needGreedMember, dropped []*items.Item) {
	if g == nil || roomID == "" || len(members) < 2 || len(dropped) == 0 {
		return
	}
	g.needGreed.mu.Lock()
	defer g.needGreed.mu.Unlock()

	now := time.Now()
	for _, item := range dropped {
		if item == nil || item.ID == "" {
			continue
		}
		g.needGreed.seq++
		round := &needGreedRound{
			ID:       fmt.Sprintf("ng-%d", g.needGreed.seq),
			Item:     item,
			RoomID:   roomID,
			PartyID:  partyID,
			Members:  append([]needGreedMember(nil), members...),
			Choices:  make(map[string]lootRollChoice, len(members)),
			OpenedAt: now,
			Expires:  now.Add(NeedGreedTTL),
		}
		for _, m := range members {
			round.Choices[m.CharacterID] = lootChoiceUnset
		}
		if g.needGreed.active == nil {
			g.needGreed.active = round
			g.broadcastNeedGreedLocked(round, true)
		} else {
			g.needGreed.pending = append(g.needGreed.pending, round)
		}
	}
}

func (g *Game) broadcastNeedGreedLocked(round *needGreedRound, pending bool) {
	if g == nil || round == nil || g.sendMessage == nil {
		return
	}
	name, quality, qty := needGreedItemLabel(round.Item)
	for _, m := range round.Members {
		if m.UserID == "" {
			continue
		}
		g.sendMessage <- messages.NewNeedGreedMessage(m.UserID, pending, round.ID, name, quality, qty, round.Expires)
		if pending {
			g.sendMessage <- messages.Reply(m.UserID, fmt.Sprintf(
				"[Party] Loot roll: %s — type need, greed, or pass (%.0fs).",
				name, NeedGreedTTL.Seconds()))
		}
	}
}

func needGreedItemLabel(item *items.Item) (name, quality string, qty int32) {
	if item == nil {
		return "an item", string(items.ItemQualityNormal), 1
	}
	name = strings.TrimSpace(item.Name)
	if name == "" {
		name = "an item"
	}
	quality = string(item.Quality)
	if quality == "" {
		quality = string(items.ItemQualityNormal)
	}
	qty = item.Quantity
	if qty < 1 {
		qty = 1
	}
	return name, quality, qty
}

// RecordNeedGreedChoice records need/greed/pass for the active round.
func (g *Game) RecordNeedGreedChoice(characterID, userID, raw string) bool {
	choice, ok := parseLootRollChoice(raw)
	if !ok || g == nil || characterID == "" {
		return false
	}
	g.needGreed.mu.Lock()
	defer g.needGreed.mu.Unlock()

	round := g.needGreed.active
	if round == nil {
		return false
	}
	g.pruneNeedGreedLocked(round, time.Now())
	if _, eligible := round.Choices[characterID]; !eligible {
		return false
	}
	if round.Choices[characterID] != lootChoiceUnset {
		if userID != "" {
			g.sendMessage <- messages.Reply(userID, "You already chose for this loot roll.")
		}
		return true
	}
	round.Choices[characterID] = choice
	if userID != "" {
		g.sendMessage <- messages.Reply(userID, "You selected "+choice.String()+".")
	}
	if needGreedAllChosen(round) {
		g.resolveNeedGreedLocked(round, time.Now())
	}
	return true
}

func needGreedAllChosen(round *needGreedRound) bool {
	if round == nil || len(round.Choices) == 0 {
		return true
	}
	for _, c := range round.Choices {
		if c == lootChoiceUnset {
			return false
		}
	}
	return true
}

// ExpireNeedGreedRounds resolves timed-out rolls and advances the queue.
func (g *Game) ExpireNeedGreedRounds() {
	if g == nil {
		return
	}
	g.needGreed.mu.Lock()
	defer g.needGreed.mu.Unlock()
	now := time.Now()
	if g.needGreed.active == nil {
		return
	}
	g.pruneNeedGreedLocked(g.needGreed.active, now)
	if now.Before(g.needGreed.active.Expires) && !needGreedAllChosen(g.needGreed.active) {
		return
	}
	g.resolveNeedGreedLocked(g.needGreed.active, now)
}

// NoteNeedGreedMemberLeft marks a member Pass (disconnect, leave room, die, leave party).
func (g *Game) NoteNeedGreedMemberLeft(characterID string) {
	if g == nil || characterID == "" {
		return
	}
	g.needGreed.mu.Lock()
	defer g.needGreed.mu.Unlock()
	now := time.Now()
	if g.needGreed.active != nil {
		g.autoPassNeedGreedLocked(g.needGreed.active, characterID)
		g.pruneNeedGreedLocked(g.needGreed.active, now)
		if needGreedAllChosen(g.needGreed.active) {
			g.resolveNeedGreedLocked(g.needGreed.active, now)
		}
	}
	for _, round := range g.needGreed.pending {
		g.autoPassNeedGreedLocked(round, characterID)
	}
}

func (g *Game) autoPassNeedGreedLocked(round *needGreedRound, characterID string) {
	if round == nil || characterID == "" {
		return
	}
	if _, ok := round.Choices[characterID]; ok {
		if round.Choices[characterID] == lootChoiceUnset {
			round.Choices[characterID] = lootChoicePass
		}
	}
}

func (g *Game) pruneNeedGreedLocked(round *needGreedRound, now time.Time) {
	if g == nil || round == nil {
		return
	}
	online := map[string]bool{}
	for _, op := range g.GetOnlinePlayers() {
		if op.CharacterID != "" {
			online[op.CharacterID] = true
		}
	}
	for id := range round.Choices {
		if round.Choices[id] != lootChoiceUnset {
			continue
		}
		if !online[id] {
			round.Choices[id] = lootChoicePass
			continue
		}
		if g.Facade == nil {
			continue
		}
		char, err := g.Facade.CharactersService().FindByID(id)
		if err != nil || char == nil || char.CurrentHitPoints <= 0 {
			round.Choices[id] = lootChoicePass
			continue
		}
		if round.RoomID != "" && char.CurrentRoomID != round.RoomID {
			round.Choices[id] = lootChoicePass
			continue
		}
		if round.PartyID != "" {
			party, err := g.Facade.PartiesService().FindByCharacterID(id)
			if err != nil || party == nil || party.ID != round.PartyID {
				round.Choices[id] = lootChoicePass
			}
		}
	}
	_ = now
}

func (g *Game) resolveNeedGreedLocked(round *needGreedRound, now time.Time) {
	if g == nil || round == nil {
		return
	}
	if g.needGreed.active != round {
		return
	}
	g.pruneNeedGreedLocked(round, now)
	for id, c := range round.Choices {
		if c == lootChoiceUnset {
			round.Choices[id] = lootChoicePass
		}
	}

	winnerID, tier, won := resolveNeedGreedWinner(round.Choices, nil)
	itemName, _, _ := needGreedItemLabel(round.Item)

	// Clear prompts for everyone in this round.
	g.broadcastNeedGreedLocked(round, false)

	if won {
		winnerName := winnerID
		for _, m := range round.Members {
			if m.CharacterID == winnerID {
				winnerName = m.Name
				break
			}
		}
		g.grantNeedGreedItem(winnerID, round.Item, round.RoomID)
		line := fmt.Sprintf("[Party] %s won %s on %s.", winnerName, tier, itemName)
		g.announceNeedGreed(round, line)
		log.WithFields(log.Fields{
			"round":  round.ID,
			"winner": winnerID,
			"tier":   tier,
			"item":   itemName,
		}).Info("Need/Greed resolved")
	} else {
		g.placeNeedGreedOnGround(round.Item, round.RoomID)
		line := fmt.Sprintf("[Party] Everyone passed on %s — it stays on the ground.", itemName)
		g.announceNeedGreed(round, line)
		log.WithFields(log.Fields{
			"round": round.ID,
			"item":  itemName,
		}).Info("Need/Greed all-pass / timeout → ground")
	}

	g.needGreed.active = nil
	if len(g.needGreed.pending) > 0 {
		next := g.needGreed.pending[0]
		g.needGreed.pending = g.needGreed.pending[1:]
		next.OpenedAt = now
		next.Expires = now.Add(NeedGreedTTL)
		// Drop members who already auto-passed / left before this round opens.
		g.pruneNeedGreedLocked(next, now)
		eligibleLeft := 0
		for _, c := range next.Choices {
			if c == lootChoiceUnset {
				eligibleLeft++
			}
		}
		if eligibleLeft < 2 {
			// Not enough rollers left — ground it and keep draining.
			g.needGreed.active = next
			g.resolveNeedGreedLocked(next, now)
			return
		}
		g.needGreed.active = next
		g.broadcastNeedGreedLocked(next, true)
	}
}

func (g *Game) announceNeedGreed(round *needGreedRound, line string) {
	if g == nil || g.sendMessage == nil || line == "" {
		return
	}
	seen := map[string]bool{}
	for _, m := range round.Members {
		if m.UserID == "" || seen[m.UserID] {
			continue
		}
		seen[m.UserID] = true
		g.sendMessage <- messages.Reply(m.UserID, line)
	}
	if round.RoomID != "" {
		msg := messages.NewRoomBasedMessage("SYSTEM", line)
		msg.Audience = messages.MessageAudienceRoom
		msg.AudienceID = round.RoomID
		g.sendMessage <- msg
	}
}

func (g *Game) grantNeedGreedItem(characterID string, item *items.Item, roomID string) {
	if g == nil || g.Facade == nil || item == nil || characterID == "" {
		return
	}
	char, err := g.Facade.CharactersService().FindByID(characterID)
	if err != nil || char == nil {
		g.placeNeedGreedOnGround(item, roomID)
		return
	}
	err = g.Facade.CharactersService().Modify(characterID, func(ch *characters.Character) error {
		return ch.Inventory.AddItem(item)
	})
	if err != nil {
		// Inventory full — drop at their feet (or fight room).
		dropRoom := char.CurrentRoomID
		if dropRoom == "" {
			dropRoom = roomID
		}
		g.placeNeedGreedOnGround(item, dropRoom)
		if char.BelongsUserID != "" {
			g.sendMessage <- messages.Reply(char.BelongsUserID, "Your inventory is full — "+item.Name+" drops at your feet.")
		}
		return
	}
	fresh, _ := g.Facade.CharactersService().FindByID(characterID)
	if fresh != nil {
		if g.QuestTracker != nil {
			g.QuestTracker.OnItemPickup(fresh.ID, fresh.BelongsUserID, item)
		}
		g.notifyInventory(fresh, "You receive "+item.Name+".")
	}
}

func (g *Game) placeNeedGreedOnGround(item *items.Item, roomID string) {
	if g == nil || g.Facade == nil || item == nil || roomID == "" {
		return
	}
	room, err := g.Facade.RoomsService().FindByID(roomID)
	if err != nil || room == nil {
		log.WithError(err).WithField("roomID", roomID).Warn("Need/Greed could not place item on ground")
		return
	}
	// Item was already Store'd when rolled; just attach to the room.
	if err := room.AddItem(item.ID); err != nil {
		log.WithError(err).WithField("item", item.Name).Warn("Need/Greed AddItem to room failed")
		return
	}
	if err := g.Facade.RoomsService().Update(room.ID, room); err != nil {
		log.WithError(err).WithField("roomID", room.ID).Warn("Need/Greed room update failed")
		return
	}
	g.refreshRoomGround(roomID)
}

func (g *Game) refreshRoomGround(roomID string) {
	if g == nil || g.Facade == nil || roomID == "" || g.sendMessage == nil {
		return
	}
	room, err := g.Facade.RoomsService().FindByID(roomID)
	if err != nil || room == nil {
		return
	}
	for _, op := range g.GetOnlinePlayers() {
		if op.RoomID != roomID || op.UserID == "" {
			continue
		}
		char, cerr := g.Facade.CharactersService().FindByID(op.CharacterID)
		if cerr != nil || char == nil {
			continue
		}
		user, uerr := g.Facade.UsersService().FindByID(op.UserID)
		if uerr != nil || user == nil {
			continue
		}
		roomUpdate := messages.NewRoomUpdateMessage(util.RoomWithCharacterReveals(room, char), user, g, char)
		roomUpdate.AudienceID = user.ID
		g.sendMessage <- roomUpdate
	}
}

// HasActiveNeedGreed reports whether characterID is in the open loot roll.
func (g *Game) HasActiveNeedGreed(characterID string) bool {
	if g == nil || characterID == "" {
		return false
	}
	g.needGreed.mu.Lock()
	defer g.needGreed.mu.Unlock()
	if g.needGreed.active == nil {
		return false
	}
	_, ok := g.needGreed.active.Choices[characterID]
	return ok
}