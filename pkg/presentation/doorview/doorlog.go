package doorview

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

// headerLine is the status row. Experience sits after gold so a reader that
// stops at the gold figure still matches.
func headerLine(level, hp, maxHP int32, gold int64, xp int32) string {
	xpPart := fmt.Sprintf("XP %d", xp)
	if level < ruleset.LevelCap() {
		next := leveling.GetXPRequired(level + 1)
		if next > xp {
			xpPart = fmt.Sprintf("XP %d/%d", xp, next)
		}
	}
	return fmt.Sprintf("Level %d   HP %d/%d   Gold %d   %s", level, hp, maxHP, gold, xpPart)
}

func combatKind(kind string) bool {
	switch kind {
	case "combatStart", "combatTurn", "combatAction", "combatEnd":
		return true
	default:
		return false
	}
}

// keepsFight is a swing, a flee, or a combat status read. Anything else,
// once the fight is over, is the key that drops the log.
func keepsFight(cmd string) bool {
	fields := strings.Fields(strings.ToLower(cmd))
	if len(fields) == 0 {
		return false
	}
	switch fields[0] {
	case "attack", "a", "hit", "flee", "run", "escape", "defend", "cast", "focus", "status", "cs":
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
// so the legend and the key agree.
func applyOpenExits(binds map[string]keyBind, room *rooms.Room, ch *characters.Character) map[string]keyBind {
	if len(binds) == 0 || room == nil {
		return binds
	}
	for letter, dir := range map[string]string{
		"n": "north", "s": "south", "e": "east", "w": "west", "u": "up", "o": "out",
	} {
		b, ok := binds[letter]
		if !ok || strings.EqualFold(b.Command, dir) || !roomHasExit(room, ch, dir) {
			continue
		}
		binds[letter] = keyBind{Key: letter, Command: dir, Label: exitLabel(dir)}
	}
	return binds
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
	v.setFloor(user.ID, messages.LastNoticeGen(user.ID))
}

// settleLog keeps the fight transcript through every swing. The first key
// that is not itself a fight action, after combat has ended, clears it.
func (v *View) settleLog(user *entities.User, cmd string) {
	if user == nil || v.characterInFight(user) || keepsFight(cmd) {
		return
	}
	v.clearRecent(user.ID)
	v.clearHold(user.ID)
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
