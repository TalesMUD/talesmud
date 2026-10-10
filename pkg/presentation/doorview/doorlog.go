package doorview

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

// headerLine is the status row. Experience sits after gold so a reader that
// stops at the gold figure still matches.
func headerLine(level, hp, maxHP int32, gold int64, xp int32) string {
	lv, hpText, goldText, xpPart := statusParts(level, hp, maxHP, gold, xp)
	return fmt.Sprintf("Level %s   HP %s   Gold %s   %s", lv, hpText, goldText, xpPart)
}

// statusStrip is headerLine with dim labels and a bright hit-point figure.
// Stripping the colour codes yields headerLine.
func statusStrip(level, hp, maxHP int32, gold int64, xp int32) string {
	lv, hpText, goldText, xpPart := statusParts(level, hp, maxHP, gold, xp)
	const dim = "\x1b[0;37m"
	const hi = "\x1b[1;37m"
	const off = "\x1b[0m"
	return dim + "Level " + off + hi + lv + off +
		dim + "   HP " + off + hpColour(hp, maxHP) + hpText + off +
		dim + "   Gold " + off + hi + goldText + off +
		dim + "   " + off + hi + xpPart + off
}

func statusParts(level, hp, maxHP int32, gold int64, xp int32) (string, string, string, string) {
	xpPart := fmt.Sprintf("XP %d", xp)
	if level < ruleset.LevelCap() {
		next := leveling.GetXPRequired(level + 1)
		if next > xp {
			xpPart = fmt.Sprintf("XP %d/%d", xp, next)
		}
	}
	return fmt.Sprintf("%d", level), fmt.Sprintf("%d/%d", hp, maxHP), fmt.Sprintf("%d", gold), xpPart
}

func hpColour(hp, maxHP int32) string {
	if maxHP <= 0 {
		return "\x1b[1;31m"
	}
	pct := int64(hp) * 100 / int64(maxHP)
	switch {
	case pct >= 60:
		return "\x1b[1;32m"
	case pct >= 30:
		return "\x1b[1;33m"
	default:
		return "\x1b[1;31m"
	}
}

func combatKind(kind string) bool {
	switch kind {
	case "combatStart", "combatTurn", "combatAction", "combatEnd":
		return true
	default:
		return false
	}
}

// queuedAttack is a bare swing with no named target. After a fight, that key
// is the extra press still in the queue, and it must not start another swing.
func queuedAttack(cmd string) bool {
	fields := strings.Fields(strings.ToLower(cmd))
	if len(fields) != 1 {
		return false
	}
	switch fields[0] {
	case "attack", "a", "hit":
		return true
	default:
		return false
	}
}

func sheetCommand(cmd string) bool {
	fields := strings.Fields(strings.ToLower(cmd))
	if len(fields) != 1 {
		return false
	}
	switch fields[0] {
	case "stats", "character", "char":
		return true
	default:
		return false
	}
}

// notableLines pulls the outcome out of a victory or defeat box. The box
// rules themselves are not the line a 19-row frame can keep.
func notableLines(text string) []string {
	var keep []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(stripANSI(line))
		if line == "" || !notableLine(line) {
			continue
		}
		keep = append(keep, line)
	}
	if len(keep) == 0 {
		return condense(text)
	}
	if len(keep) > 6 {
		keep = keep[:6]
	}
	return keep
}

func notableLine(line string) bool {
	upper := strings.ToUpper(line)
	if strings.Contains(upper, "VICTORY") || strings.Contains(upper, "DEFEAT") || strings.Contains(upper, "ESCAP") || strings.Contains(upper, "RELEASED") || strings.Contains(upper, "FLED") {
		return true
	}
	return strings.Contains(line, "XP") || strings.Contains(line, "Gold") || strings.Contains(line, "Defeated")
}

func directionOf(text string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case "n", "north":
		return "north", true
	case "s", "south":
		return "south", true
	case "e", "east":
		return "east", true
	case "w", "west":
		return "west", true
	case "u", "up":
		return "up", true
	case "d", "down":
		return "down", true
	case "o", "out":
		return "out", true
	default:
		return "", false
	}
}

func roomHasExit(room *rooms.Room, ch *characters.Character, name string) bool {
	if room == nil || room.Exits == nil || name == "" {
		return false
	}
	for _, ex := range *room.Exits {
		if !strings.EqualFold(ex.Name, name) {
			continue
		}
		if ex.Hidden && (ch == nil || !ch.HasRevealedExit(room.ID, ex.Name)) {
			return false
		}
		return true
	}
	return false
}

func exitLabel(dir string) string {
	switch dir {
	case "north":
		return "North"
	case "south":
		return "South"
	case "east":
		return "East"
	case "west":
		return "West"
	case "up":
		return "Up"
	case "down":
		return "Down"
	case "out":
		return "Out"
	default:
		return dir
	}
}

// applyOpenExits lets a real exit win over a menu bind on the same letter,
// and lists an open way even when the pack never bound that letter.
func applyOpenExits(binds map[string]keyBind, room *rooms.Room, ch *characters.Character) map[string]keyBind {
	if room == nil {
		return binds
	}
	if binds == nil {
		binds = map[string]keyBind{}
	}
	for letter, dir := range map[string]string{
		"n": "north", "s": "south", "e": "east", "w": "west", "u": "up", "o": "out",
	} {
		if !roomHasExit(room, ch, dir) {
			continue
		}
		if b, ok := binds[letter]; ok && strings.EqualFold(b.Command, dir) {
			continue
		}
		binds[letter] = keyBind{Key: letter, Command: dir, Label: exitLabel(dir)}
	}
	return binds
}

// commandFooter names down only when the room has that exit. Otherwise the
// bottom row is just the command key.
func commandFooter(room *rooms.Room, ch *characters.Character) string {
	if roomHasExit(room, ch, "down") {
		return "d down   : command"
	}
	return ": command"
}

func idleFooter(room *rooms.Room, ch *characters.Character) string {
	ways := "n s e w u"
	if roomHasExit(room, ch, "down") {
		ways = "n s e w u d"
	}
	return ways + "   l look   a attack   i inventory   : command"
}

func (v *View) screenArtFor(room *rooms.Room) string {
	if room == nil {
		return ""
	}
	if art := screenArt(room.ID); art != "" {
		return art
	}
	for _, tag := range room.Tags {
		id, ok := strings.CutPrefix(tag, "screen:")
		if !ok || id == "" {
			continue
		}
		if art := screenArt(id); art != "" {
			return art
		}
	}
	return ""
}

func (v *View) characterInFight(user *entities.User) bool {
	ch := v.character(user)
	return ch != nil && ch.InCombat
}

// prepareKey drops a one-shot from the previous key and records that key's
// generation. A reply stamped at or below this floor arrives too late.
func (v *View) prepareKey(user *entities.User) {
	if user == nil {
		return
	}
	v.clearNotice(user.ID)
	v.clearFlash(user.ID)
	v.clearSheet(user.ID)
	v.setFloor(user.ID, messages.LastNoticeGen(user.ID))
}

// statsLines is the text-client character sheet: hit points, gold, worn
// gear, and the optional gems flag. A missing flag shows as zero.
func statsLines(ch *characters.Character) []string {
	if ch == nil {
		return nil
	}
	lines := []string{ch.Name}
	if path := flagString(ch.Flags, "path"); path != "" {
		lines = append(lines, path)
	} else if ch.Race.Name != "" || ch.Class.Name != "" {
		lines = append(lines, strings.TrimSpace(ch.Race.Name+" "+ch.Class.Name))
	}
	lines = append(lines, fmt.Sprintf("Level %d   HP %d/%d   Gold %d", ch.Level, ch.CurrentHitPoints, ch.MaxHitPoints, ch.Gold))
	lines = append(lines, fmt.Sprintf("Gems %d", flagAmount(ch.Flags, "gems")))
	lines = append(lines, fmt.Sprintf("Weapon %d   Armor %d", ch.GetWeaponDamage(), ch.GetArmorDefense()))
	worn := wornLines(ch)
	if len(worn) == 0 {
		return append(lines, "Gear: none")
	}
	return append(lines, worn...)
}

func wornLines(ch *characters.Character) []string {
	if ch == nil || len(ch.EquippedItems) == 0 {
		return nil
	}
	slots := []struct {
		slot  items.ItemSlot
		label string
	}{
		{items.ItemSlotMainHand, "Main hand"},
		{items.ItemSlotOffHand, "Off hand"},
		{items.ItemSlotHead, "Head"},
		{items.ItemSlotChest, "Chest"},
		{items.ItemSlotHands, "Hands"},
		{items.ItemSlotLegs, "Legs"},
		{items.ItemSlotBoots, "Boots"},
	}
	var lines []string
	for _, slot := range slots {
		item := ch.EquippedItems[slot.slot]
		if item == nil || strings.TrimSpace(item.Name) == "" {
			continue
		}
		lines = append(lines, slot.label+": "+item.Name)
	}
	return lines
}

func flagString(flags map[string]interface{}, key string) string {
	if flags == nil {
		return ""
	}
	switch v := flags[key].(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return ""
	}
}

func flagAmount(flags map[string]interface{}, key string) int64 {
	if flags == nil {
		return 0
	}
	switch n := flags[key].(type) {
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	case float32:
		return int64(n)
	default:
		return 0
	}
}

// dropFightLog clears a finished fight. A swing that is still in progress
// keeps the transcript.
func (v *View) dropFightLog(user *entities.User) bool {
	if user == nil || v.characterInFight(user) || !v.holdOf(user.ID) {
		return false
	}
	v.clearRecent(user.ID)
	v.clearHold(user.ID)
	return true
}

// settleLog clears a finished fight on any following command. It reports
// whether that command is a bare extra swing and should not be sent.
func (v *View) settleLog(user *entities.User, cmd string) bool {
	if !v.dropFightLog(user) {
		return false
	}
	return queuedAttack(cmd)
}

func (v *View) setHold(id string, on bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.hold == nil {
		v.hold = map[string]bool{}
	}
	if on {
		v.hold[id] = true
		return
	}
	delete(v.hold, id)
}

func (v *View) holdOf(id string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.hold[id]
}

func (v *View) clearHold(id string) {
	v.setHold(id, false)
}

func (v *View) setFloor(id string, gen uint64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.floor == nil {
		v.floor = map[string]uint64{}
	}
	v.floor[id] = gen
}

func (v *View) floorOf(id string) uint64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.floor[id]
}

func (v *View) setFlash(id string, lines []string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.flash == nil {
		v.flash = map[string][]string{}
	}
	v.flash[id] = append([]string{}, lines...)
}

func (v *View) clearFlash(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.flash, id)
}

func (v *View) peekFlash(id string) []string {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.flash == nil {
		return nil
	}
	return append([]string{}, v.flash[id]...)
}

func (v *View) setSheet(id string, lines []string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.sheet == nil {
		v.sheet = map[string][]string{}
	}
	v.sheet[id] = append([]string{}, lines...)
}

func (v *View) clearSheet(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.sheet, id)
}

func (v *View) peekSheet(id string) []string {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.sheet == nil {
		return nil
	}
	return append([]string{}, v.sheet[id]...)
}
