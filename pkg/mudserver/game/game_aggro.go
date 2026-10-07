package game

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

// aggroFire is a grace timer that elapsed. gen ignores a timer that was
// replaced or cancelled after it was already queued.
type aggroFire struct {
	key string
	gen uint64
}

type aggroWatch struct {
	key      string
	gen      uint64
	playerID string
	npcID    string
	roomID   string
	due      time.Time
	timer    *time.Timer
}

// aggroBook is the in-memory sight scheduler. It is process-local.
// Timers only enqueue. Combat starts on the command loop, or in tests via FireDueAggro.
type aggroBook struct {
	mu       sync.Mutex
	pending  map[string]*aggroWatch
	cooldown map[string]time.Time
	now      func() time.Time
	manual   bool
	dropped  bool
	nextGen  uint64
	ch       chan aggroFire
}

func (b *aggroBook) nowTime() time.Time {
	if b.now != nil {
		return b.now()
	}
	return time.Now()
}

func aggroKey(playerID, npcID string) string {
	return playerID + "|" + npcID
}

// SetAggroClock installs a test clock and does not arm time.AfterFunc.
// Tests advance the clock and call FireDueAggro. Production leaves this unset.
func (g *Game) SetAggroClock(now func() time.Time) {
	if g == nil {
		return
	}
	g.aggro.mu.Lock()
	g.aggro.manual = true
	g.aggro.now = now
	g.aggro.mu.Unlock()
}

// FireDueAggro runs every watch whose grace has elapsed. Tests call this on the
// same goroutine that drives the fixture. Production uses the command loop.
func (g *Game) FireDueAggro() {
	if g == nil {
		return
	}
	g.aggro.mu.Lock()
	now := g.aggro.nowTime()
	due := make([]aggroFire, 0)
	for key, watch := range g.aggro.pending {
		if watch == nil || watch.due.After(now) {
			continue
		}
		due = append(due, aggroFire{key: key, gen: watch.gen})
	}
	g.aggro.mu.Unlock()
	sort.Slice(due, func(i, j int) bool { return due[i].key < due[j].key })
	for _, fire := range due {
		g.fireAggro(fire.key, fire.gen)
	}
}

// NotePlayerEntered schedules a grace watch for each aggressive NPC already in the room.
// A watch for a room the player left is dropped and does not start the reaggro cooldown.
func (g *Game) NotePlayerEntered(characterID, roomID string) {
	if g == nil || characterID == "" || roomID == "" {
		return
	}
	g.aggro.cancelPlayerExcept(characterID, roomID)
	if !ruleset.AggroOnSightEnabled() || g.NPCManager == nil {
		return
	}
	char := g.aggroCharacter(characterID)
	if !g.playerCanBeAggroed(char, roomID) {
		return
	}
	for _, inst := range g.NPCManager.GetInstancesInRoom(roomID) {
		g.scheduleAggro(char, inst, roomID)
	}
}

// NoteNPCAppeared schedules a grace watch for each player already standing in the room.
func (g *Game) NoteNPCAppeared(npcID, roomID string) {
	if g == nil || npcID == "" || roomID == "" || g.NPCManager == nil {
		return
	}
	g.aggro.cancelNPCExcept(npcID, roomID)
	if !ruleset.AggroOnSightEnabled() {
		return
	}
	inst := g.NPCManager.GetInstance(npcID)
	if !npcCanAggro(inst, roomID) || g.CombatController != nil && g.CombatController.IsNPCInCombat(inst.Entity.ID) {
		return
	}
	for _, player := range g.GetOnlinePlayers() {
		if player.CharacterID == "" {
			continue
		}
		char := g.aggroCharacter(player.CharacterID)
		if !g.playerCanBeAggroed(char, roomID) {
			continue
		}
		g.scheduleAggro(char, inst, roomID)
	}
}

func (g *Game) cancelAggroForPlayer(characterID string) {
	if g == nil || characterID == "" {
		return
	}
	g.aggro.cancelPlayer(characterID)
}

func (g *Game) aggroCharacter(id string) *characters.Character {
	if g.Facade == nil || id == "" {
		return nil
	}
	char, err := g.Facade.CharactersService().FindByID(id)
	if err != nil {
		return nil
	}
	return char
}

func (g *Game) playerCanBeAggroed(char *characters.Character, roomID string) bool {
	if char == nil || char.ID == "" || char.CurrentRoomID != roomID {
		return false
	}
	if char.CurrentHitPoints <= 0 || char.AwaitingReset || char.InCombat {
		return false
	}
	if g.playingUserID(char.ID) == "" {
		return false
	}
	if g.CombatController != nil && g.CombatController.IsPlayerInCombat(char.ID) {
		return false
	}
	return true
}

func npcCanAggro(inst *npc.NPC, roomID string) bool {
	if inst == nil || inst.Entity == nil || inst.Entity.ID == "" {
		return false
	}
	if inst.CurrentRoomID != roomID || inst.IsDead || !inst.IsEnemy() || inst.EnemyTrait == nil {
		return false
	}
	if !inst.EnemyTrait.AggroOnSight || inst.InCombat {
		return false
	}
	return true
}

func aggroLevelSkipped(playerLevel, enemyLevel int32) bool {
	gap := ruleset.AggroMaxLevelGap()
	if gap <= 0 {
		return false
	}
	return int(playerLevel) > int(enemyLevel)+gap
}

func (g *Game) playingUserID(characterID string) string {
	for _, player := range g.GetOnlinePlayers() {
		if player.CharacterID == characterID && player.UserID != "" {
			return player.UserID
		}
	}
	return ""
}

func (g *Game) scheduleAggro(char *characters.Character, inst *npc.NPC, roomID string) {
	if !npcCanAggro(inst, roomID) || !g.playerCanBeAggroed(char, roomID) {
		return
	}
	if g.CombatController != nil && g.CombatController.IsNPCInCombat(inst.Entity.ID) {
		return
	}
	if aggroLevelSkipped(char.Level, inst.Level) {
		return
	}
	key := aggroKey(char.ID, inst.Entity.ID)
	grace := ruleset.AggroGrace()
	g.aggro.arm(key, char.ID, inst.Entity.ID, roomID, grace)
}

func (b *aggroBook) arm(key, playerID, npcID, roomID string, grace time.Duration) {
	if b == nil || key == "" {
		return
	}
	b.mu.Lock()
	if b.pending == nil {
		b.pending = map[string]*aggroWatch{}
	}
	if until, ok := b.cooldown[key]; ok && b.nowTime().Before(until) {
		b.mu.Unlock()
		return
	}
	if old := b.pending[key]; old != nil && old.roomID == roomID {
		b.mu.Unlock()
		return
	}
	if old := b.pending[key]; old != nil && old.timer != nil {
		old.timer.Stop()
	}
	b.nextGen++
	watch := &aggroWatch{
		key:      key,
		gen:      b.nextGen,
		playerID: playerID,
		npcID:    npcID,
		roomID:   roomID,
		due:      b.nowTime().Add(grace),
	}
	b.pending[key] = watch
	manual := b.manual
	gen := watch.gen
	delay := time.Until(watch.due)
	if b.now != nil {
		delay = watch.due.Sub(b.nowTime())
	}
	b.mu.Unlock()
	if manual {
		return
	}
	if delay < 0 {
		delay = 0
	}
	timer := time.AfterFunc(delay, func() {
		b.notify(key, gen)
	})
	b.mu.Lock()
	if cur := b.pending[key]; cur != nil && cur.gen == gen {
		cur.timer = timer
	} else {
		timer.Stop()
	}
	b.mu.Unlock()
}

func (b *aggroBook) notify(key string, gen uint64) {
	if b == nil || b.ch == nil {
		return
	}
	select {
	case b.ch <- aggroFire{key: key, gen: gen}:
	default:
		b.mu.Lock()
		b.dropped = true
		b.mu.Unlock()
	}
}

func (b *aggroBook) cancelPlayerExcept(playerID, roomID string) {
	b.cancel(func(watch *aggroWatch) bool {
		return watch.playerID == playerID && watch.roomID != roomID
	})
}

func (b *aggroBook) cancelPlayer(playerID string) {
	b.cancel(func(watch *aggroWatch) bool {
		return watch.playerID == playerID
	})
}

func (b *aggroBook) cancelNPCExcept(npcID, roomID string) {
	b.cancel(func(watch *aggroWatch) bool {
		return watch.npcID == npcID && watch.roomID != roomID
	})
}

func (b *aggroBook) cancel(drop func(*aggroWatch) bool) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for key, watch := range b.pending {
		if watch == nil || !drop(watch) {
			continue
		}
		if watch.timer != nil {
			watch.timer.Stop()
		}
		delete(b.pending, key)
	}
}

func (g *Game) fireAggro(key string, gen uint64) {
	watch := g.aggro.take(key, gen)
	if watch == nil {
		return
	}
	g.tryAggro(watch.playerID, watch.npcID, watch.roomID)
}

func (b *aggroBook) take(key string, gen uint64) *aggroWatch {
	b.mu.Lock()
	defer b.mu.Unlock()
	watch := b.pending[key]
	if watch == nil || watch.gen != gen {
		return nil
	}
	if watch.timer != nil {
		watch.timer.Stop()
	}
	delete(b.pending, key)
	return watch
}

func (g *Game) tryAggro(playerID, npcID, roomID string) {
	if g == nil || !ruleset.AggroOnSightEnabled() || g.NPCManager == nil || g.CombatController == nil {
		return
	}
	char := g.aggroCharacter(playerID)
	inst := g.NPCManager.GetInstance(npcID)
	if !g.playerCanBeAggroed(char, roomID) || !npcCanAggro(inst, roomID) {
		return
	}
	if g.CombatController.IsNPCInCombat(inst.Entity.ID) || g.CombatController.IsPlayerInCombat(char.ID) {
		return
	}
	if aggroLevelSkipped(char.Level, inst.Level) {
		return
	}
	if g.aggro.cooling(aggroKey(playerID, npcID)) {
		return
	}
	userID := g.playingUserID(char.ID)
	if userID == "" {
		return
	}
	g.CombatController.BeginEngagement(roomID, char, userID, inst, true)
}

func (b *aggroBook) cooling(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	until, ok := b.cooldown[key]
	return ok && b.nowTime().Before(until)
}

func (g *Game) stampFightAggroCooldown(instance *combat.CombatInstance) {
	if g == nil || instance == nil {
		return
	}
	cd := ruleset.AggroReaggroCooldown()
	if cd <= 0 {
		return
	}
	// The NPCs that fought, plus every other aggroOnSight NPC still standing in the
	// room. A pair would otherwise hand the player to the wolf that never swung.
	ids := map[string]struct{}{}
	for _, enemy := range instance.Enemies {
		if enemy.ID == "" || enemy.Summoned {
			continue
		}
		ids[enemy.ID] = struct{}{}
	}
	if g.NPCManager != nil && instance.OriginRoomID != "" {
		for _, inst := range g.NPCManager.GetInstancesInRoom(instance.OriginRoomID) {
			if inst == nil || inst.Entity == nil || inst.Entity.ID == "" || inst.IsDead {
				continue
			}
			if !inst.IsEnemy() || inst.EnemyTrait == nil || !inst.EnemyTrait.AggroOnSight {
				continue
			}
			ids[inst.Entity.ID] = struct{}{}
		}
	}
	g.aggro.mu.Lock()
	defer g.aggro.mu.Unlock()
	if g.aggro.cooldown == nil {
		g.aggro.cooldown = map[string]time.Time{}
	}
	until := g.aggro.nowTime().Add(cd)
	for _, player := range instance.Players {
		if player.ID == "" {
			continue
		}
		for npcID := range ids {
			g.aggro.cooldown[aggroKey(player.ID, npcID)] = until
		}
	}
}

// nudgeDueAggro requeues a grace timer whose channel send was dropped.
// It does not start combat. The command loop does that.
func (g *Game) nudgeDueAggro() {
	if g == nil || g.aggroCh == nil {
		return
	}
	g.aggro.mu.Lock()
	if !g.aggro.dropped {
		g.aggro.mu.Unlock()
		return
	}
	now := g.aggro.nowTime()
	due := make([]aggroFire, 0)
	for key, watch := range g.aggro.pending {
		if watch == nil || watch.due.After(now) {
			continue
		}
		due = append(due, aggroFire{key: key, gen: watch.gen})
	}
	g.aggro.dropped = false
	g.aggro.mu.Unlock()
	for _, fire := range due {
		select {
		case g.aggroCh <- fire:
		default:
			g.aggro.mu.Lock()
			g.aggro.dropped = true
			g.aggro.mu.Unlock()
			return
		}
	}
}

// BeginEngagement starts a fight through InitiateCombat, including swarm pack,
// one onAggro per enemy, combatStart, and the party assist nudge.
// npcAggro skips the player's breath window and the overlevel warning. Those
// stay on the attack command. The sight line is not an enemy-hook stamp.
func (c *CombatController) BeginEngagement(roomID string, player *characters.Character, userID string, target *npc.NPC, npcAggro bool) *combat.CombatInstance {
	if c == nil || c.game == nil || player == nil || target == nil || target.Entity == nil || roomID == "" {
		return nil
	}
	npcManager := c.game.NPCManager
	if npcManager == nil {
		return nil
	}
	enemies := []*npc.NPC{target}
	pullSwarm := target.EnemyTrait != nil && target.EnemyTrait.CombatStyle == npc.CombatStyleSwarm
	if pullSwarm {
		for _, nearby := range npcManager.GetInstancesInRoom(roomID) {
			if nearby == nil || nearby.Entity == nil || nearby.Entity.ID == target.Entity.ID {
				continue
			}
			if !nearby.IsEnemy() || nearby.IsDead || c.IsNPCInCombat(nearby.Entity.ID) {
				continue
			}
			enemies = append(enemies, nearby)
		}
	}
	if npcAggro && userID != "" {
		c.game.sendMessage <- messages.Reply(userID, fmt.Sprintf("The %s spots you.", target.Name))
	}
	instance := c.InitiateCombat(roomID, []*characters.Character{player}, enemies)
	if instance == nil {
		return nil
	}
	player.InCombat = true
	player.CombatInstanceID = instance.ID
	_ = c.game.Facade.CharactersService().Update(player.ID, player)
	if userID != "" {
		if user, err := c.game.Facade.UsersService().FindByID(userID); err == nil && user != nil {
			c.game.SetUserSessionCharacter(user, player)
		}
	}
	for _, enemy := range enemies {
		npcManager.UpdateInstance(enemy.Entity.ID, func(n *npc.NPC) {
			n.InCombat = true
			n.CombatInstanceID = instance.ID
			n.State = "combat"
		})
	}
	names := make([]string, 0, len(enemies))
	for _, enemy := range enemies {
		names = append(names, enemy.Name)
	}
	var startBody strings.Builder
	startBody.WriteString("\n═══════════════════════════════════════════════════\n")
	startBody.WriteString("              COMBAT INITIATED!\n\n")
	if npcAggro {
		startBody.WriteString(fmt.Sprintf("The %s attacks you!\n\n", target.Name))
	} else {
		startBody.WriteString(fmt.Sprintf("You attack %s!\n\n", target.Name))
	}
	if len(enemies) > 1 {
		if pullSwarm {
			startBody.WriteString(fmt.Sprintf("A swarm joins the fight: %s\n\n", strings.Join(names[1:], ", ")))
		} else {
			startBody.WriteString(fmt.Sprintf("Enemies join the fight: %s\n\n", strings.Join(names[1:], ", ")))
		}
	}
	startBody.WriteString("Turn Order:\n")
	for i, combatant := range instance.TurnOrder {
		marker := "  "
		if i == instance.CurrentTurnIdx {
			marker = "► "
		}
		startBody.WriteString(fmt.Sprintf("%s%d. %s (Initiative: %d)\n", marker, i+1, combatant.Name, combatant.Initiative))
	}
	startBody.WriteString("\n")
	startBody.WriteString(c.GetCombatStatus(player.ID))
	startBody.WriteString("\n═══════════════════════════════════════════════════")
	if userID != "" {
		start := messages.NewCombatStartMessage(
			userID,
			startBody.String(),
			engageViews(instance.Enemies, player.Level),
			engageViews(instance.Players, player.Level),
		)
		start.TargetID = target.Entity.ID
		c.game.sendMessage <- start
		c.SetAutoAttackTarget(player.ID, target.Entity.ID)
		c.game.sendMessage <- messages.Reply(userID, "\nCombat is automatic. Commands: attack <target> (switch target) | defend | flee | status")
	}
	roomLine := fmt.Sprintf("%s engages %s in combat!", player.Name, strings.Join(names, ", "))
	if npcAggro {
		roomLine = fmt.Sprintf("The %s attacks %s!", target.Name, player.Name)
	}
	c.game.sendMessage <- messages.MessageResponse{
		Audience:   messages.MessageAudienceRoomWithoutOrigin,
		AudienceID: roomID,
		OriginID:   userID,
		Type:       messages.MessageTypeDefault,
		Message:    roomLine,
	}
	c.nudgePartyAssist(player, roomID, target.Name, names)
	return instance
}

func (c *CombatController) nudgePartyAssist(player *characters.Character, roomID, primaryTarget string, enemyNames []string) {
	if c == nil || c.game == nil || player == nil || c.game.Facade == nil {
		return
	}
	party, err := c.game.Facade.PartiesService().FindByCharacterID(player.ID)
	if err != nil || party == nil || len(party.Characters) == 0 {
		return
	}
	fightLabel := strings.Join(enemyNames, ", ")
	if fightLabel == "" {
		fightLabel = primaryTarget
	}
	attackHint := primaryTarget
	if attackHint == "" && len(enemyNames) > 0 {
		attackHint = enemyNames[0]
	}
	for _, memberID := range party.Characters {
		if memberID == player.ID || c.IsPlayerInCombat(memberID) {
			continue
		}
		for _, online := range c.game.GetOnlinePlayers() {
			if online.CharacterID != memberID || online.RoomID != roomID || online.UserID == "" {
				continue
			}
			c.game.sendMessage <- messages.Reply(online.UserID,
				fmt.Sprintf("[Party] %s engaged %s nearby! Type 'attack %s' to join the fight.",
					player.Name, fightLabel, attackHint))
		}
	}
}

func engageViews(refs []combat.CombatantRef, viewerLevel int32) []messages.CombatantView {
	out := make([]messages.CombatantView, 0, len(refs))
	for _, ref := range refs {
		view := messages.CombatantView{
			ID: ref.ID, Type: string(ref.Type), Name: ref.Name, Portrait: ref.Portrait, HP: ref.CurrentHP, MaxHP: ref.MaxHP,
			Mana: ref.CurrentMana, MaxMana: ref.MaxMana, ClassID: ref.ClassID, IsAlive: ref.IsAlive, HasFled: ref.HasFled, Level: ref.Level,
			Telegraph: ref.TelegraphAbility, Enraged: ref.Enraged,
			BossPhase: ref.BossPhase, BossPhaseLabel: ref.BossPhaseLabel, BossPhaseCount: ref.BossPhaseCount,
		}
		if ref.Type == combat.CombatantTypeNPC {
			view.Threat = balance.ThreatTier(viewerLevel, ref.Level)
		}
		out = append(out, view)
	}
	return out
}
