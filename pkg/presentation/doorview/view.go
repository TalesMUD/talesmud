// Package doorview is a text client over the shared engine.
// It reads rooms, exits, NPCs, resources, and combat status, and it sends
// normal engine commands. It does not keep hit points, prices, or fight results.
package doorview

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/presentation/ansi"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

const recentLimit = 40

// View paints one ANSI page from live engine state.
type View struct {
	Game  *game.Game
	Title string

	mu       sync.Mutex
	line     map[string]bool
	phase    map[string]string
	drafts   map[string]*draft
	notice   map[string]string
	recent   map[string][]string
	flash    map[string][]string
	sheet    map[string][]string
	floor    map[string]uint64
	hold     map[string]bool
	shopPage map[string]int
}

type draft struct {
	name    string
	path    int
	command string
	prompt  string
}

// Active reports whether this process is serving the text client.
func (v *View) Active() bool {
	return v != nil && gamemode.ANSI()
}

// OnConnect paints the current room, or a character prompt.
// A returning account applies the pending dawn pass before the first paint.
func (v *View) OnConnect(user *entities.User, send func(any)) {
	v.ensureReturning(user)
	if user != nil && user.LastCharacter != "" && v != nil && v.Game != nil && v.Game.GetFacade() != nil {
		if ch, err := v.Game.GetFacade().CharactersService().FindByID(user.LastCharacter); err == nil && ch != nil {
			v.Game.ApplySessionStart(ch)
			v.Game.EnsureLivingRoom(ch)
		}
	}
	v.paint(user, send)
}

// OnInput turns a key or a typed line into an engine command, then repaints.
func (v *View) OnInput(user *entities.User, text string, send func(any)) bool {
	if v == nil || v.Game == nil || user == nil {
		return false
	}
	text = strings.TrimSpace(text)
	if text == "" {
		phase := v.phaseOf(user.ID)
		if phase == "" && !v.peekLine(user.ID) && v.dropFightLog(user) {
			v.prepareKey(user)
			v.paint(user, send)
		}
		return true
	}
	phase := v.phaseOf(user.ID)
	if phase == "name" || phase == "amount" || v.takeLine(user.ID) {
		v.handleLine(user, text)
		v.paint(user, send)
		return true
	}
	if phase == "path" || phase == "sex" {
		if phase == "path" {
			v.choosePath(user, text)
		} else {
			v.chooseSex(user, text)
		}
		v.paint(user, send)
		return true
	}
	if text == ":" {
		v.prepareKey(user)
		v.settleLog(user, "")
		v.setLine(user.ID, true)
		v.paint(user, send)
		return true
	}
	if v.noCharacter(user) {
		v.setPhase(user.ID, "name")
		v.setLine(user.ID, true)
		v.paint(user, send)
		return true
	}
	v.prepareKey(user)
	cmd, handled := v.commandFor(user, text)
	if v.settleLog(user, cmd) {
		v.paint(user, send)
		return true
	}
	if handled {
		v.paint(user, send)
		return true
	}
	if sheetCommand(cmd) {
		if ch := v.character(user); ch != nil {
			v.setSheet(user.ID, statsLines(ch))
		}
		v.paint(user, send)
		return true
	}
	if cmd == "logout" {
		send(ansi.Frame{
			Type:      "door_frame",
			Logout:    true,
			InputMode: "hotkey",
			Cols:      ansi.Cols,
			Rows:      ansi.Rows,
			ANSI:      "\x1b[2J\x1b[H",
		})
		return true
	}
	if cmd == "shopnext" || cmd == "shopprev" {
		v.bumpPage(user.ID, cmd == "shopnext")
		v.paint(user, send)
		return true
	}
	v.Game.DispatchCommand(user, cmd)
	v.paint(user, send)
	return true
}

// OnNotice keeps a line of command or combat output and redraws the frame.
// kind is the message type. gen is the command generation, or zero when the
// line is not command output. A one-shot at or below the floor is dropped.
func (v *View) OnNotice(user *entities.User, text, kind string, gen uint64, send func(any)) {
	if v == nil || user == nil {
		return
	}
	text = strings.TrimSpace(text)
	if text == "" || skipNotice(text) {
		v.paint(user, send)
		return
	}
	if combatKind(kind) {
		lines := condense(text)
		if kind == "combatEnd" {
			lines = notableLines(text)
		}
		v.pushRecent(user.ID, lines)
		v.setHold(user.ID, true)
		v.paint(user, send)
		return
	}
	if v.isRoomEcho(user, text) {
		v.paint(user, send)
		return
	}
	if gen > 0 && gen <= v.floorOf(user.ID) {
		return
	}
	if v.characterInFight(user) {
		v.pushRecent(user.ID, condense(text))
		v.setHold(user.ID, true)
		v.paint(user, send)
		return
	}
	v.setFlash(user.ID, condense(text))
	v.paint(user, send)
}

// OnDisconnect drops line-mode state for the user.
func (v *View) OnDisconnect(user *entities.User) {
	if v == nil || user == nil {
		return
	}
	v.mu.Lock()
	delete(v.line, user.ID)
	delete(v.phase, user.ID)
	delete(v.drafts, user.ID)
	delete(v.notice, user.ID)
	delete(v.recent, user.ID)
	delete(v.flash, user.ID)
	delete(v.sheet, user.ID)
	delete(v.floor, user.ID)
	delete(v.hold, user.ID)
	v.mu.Unlock()
}

func (v *View) handleLine(user *entities.User, text string) {
	phase := v.phaseOf(user.ID)
	switch phase {
	case "amount":
		d := v.draftOf(user.ID)
		v.setPhase(user.ID, "")
		v.setLine(user.ID, false)
		v.clearNotice(user.ID)
		v.prepareKey(user)
		if d == nil || strings.EqualFold(text, "x") || text == "" {
			v.setNotice(user.ID, "Cancelled.")
			return
		}
		cmd := strings.TrimSpace(d.command + " " + text)
		if !v.settleLog(user, cmd) {
			v.Game.DispatchCommand(user, cmd)
		}
	case "name":
		v.handleName(user, text)
	default:
		v.setLine(user.ID, false)
		v.prepareKey(user)
		if v.settleLog(user, text) {
			return
		}
		if sheetCommand(text) {
			if ch := v.character(user); ch != nil {
				v.setSheet(user.ID, statsLines(ch))
			}
			return
		}
		v.Game.DispatchCommand(user, text)
	}
}

func (v *View) handleName(user *entities.User, name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		v.setPhase(user.ID, "name")
		v.setLine(user.ID, true)
		return
	}
	facade := v.Game.GetFacade()
	if facade == nil {
		return
	}
	if chars, err := facade.CharactersService().FindAllForUser(user.ID); err == nil {
		for _, ch := range chars {
			if ch != nil && strings.EqualFold(ch.Name, name) {
				v.clearNotice(user.ID)
				v.setPhase(user.ID, "")
				v.setLine(user.ID, false)
				v.Game.DispatchCommand(user, "selectcharacter "+ch.Name)
				return
			}
		}
	}
	if len(loadPaths()) == 0 {
		v.setNotice(user.ID, "No character template is available.")
		v.setPhase(user.ID, "name")
		v.setLine(user.ID, true)
		return
	}
	v.mu.Lock()
	if v.drafts == nil {
		v.drafts = map[string]*draft{}
	}
	v.drafts[user.ID] = &draft{name: name}
	v.mu.Unlock()
	v.clearNotice(user.ID)
	v.setPhase(user.ID, "path")
	v.setLine(user.ID, false)
}

func (v *View) choosePath(user *entities.User, text string) {
	paths := loadPaths()
	n, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil || n < 1 || n > len(paths) {
		v.setNotice(user.ID, "Press a number for a path.")
		return
	}
	v.mu.Lock()
	d := v.drafts[user.ID]
	if d == nil {
		d = &draft{}
		if v.drafts == nil {
			v.drafts = map[string]*draft{}
		}
		v.drafts[user.ID] = d
	}
	d.path = n - 1
	v.mu.Unlock()
	v.clearNotice(user.ID)
	v.setPhase(user.ID, "sex")
}

func (v *View) chooseSex(user *entities.User, text string) {
	word, ok := sexWord(text)
	if !ok {
		v.setNotice(user.ID, "Press M, F, or X.")
		return
	}
	d := v.draftOf(user.ID)
	paths := loadPaths()
	if d == nil || d.path < 0 || d.path >= len(paths) || strings.TrimSpace(d.name) == "" {
		v.setPhase(user.ID, "name")
		v.setLine(user.ID, true)
		v.setNotice(user.ID, "Type a character name.")
		return
	}
	v.createCharacter(user, d.name, paths[d.path], word)
}

func (v *View) createCharacter(user *entities.User, name string, path pathChoice, sex string) {
	facade := v.Game.GetFacade()
	if facade == nil {
		return
	}
	ch := &characters.Character{
		Name:             name,
		Race:             path.Race,
		Class:            path.Class,
		Level:            1,
		CurrentHitPoints: path.HP,
		MaxHitPoints:     path.HP,
		CurrentMana:      path.Mana,
		MaxMana:          path.Mana,
		Gold:             path.Gold,
		Attributes:       append(characters.Attributes(nil), path.Attrs...),
		BelongsUser:      *traits.BelongsToUser(user.ID),
		Flags:            map[string]interface{}{"sex": sex, "path": path.Name},
	}
	created, err := facade.CharactersService().Store(ch)
	if err != nil || created == nil {
		v.setNotice(user.ID, "That name is not available.")
		v.setPhase(user.ID, "name")
		v.setLine(user.ID, true)
		return
	}
	v.clearNotice(user.ID)
	v.setPhase(user.ID, "")
	v.setLine(user.ID, false)
	v.Game.DispatchCommand(user, "selectcharacter "+created.Name)
}

func (v *View) commandFor(user *entities.User, text string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(text))
	ch := v.character(user)
	var room *rooms.Room
	inCombat := false
	if ch != nil {
		inCombat = ch.InCombat
		if v.Game.GetFacade() != nil && ch.CurrentRoomID != "" {
			if found, err := v.Game.GetFacade().RoomsService().FindByID(ch.CurrentRoomID); err == nil {
				room = found
			}
		}
	}
	if dir, ok := directionOf(key); ok && roomHasExit(room, ch, dir) {
		return dir, false
	}
	binds := keysFor(room, inCombat)
	if b, ok := binds[key]; ok {
		if b.Prompt != "" && b.Command != "" {
			v.mu.Lock()
			if v.drafts == nil {
				v.drafts = map[string]*draft{}
			}
			v.drafts[user.ID] = &draft{command: b.Command, prompt: b.Prompt}
			v.mu.Unlock()
			v.setPhase(user.ID, "amount")
			v.setLine(user.ID, true)
			v.setNotice(user.ID, b.Prompt+"  x cancels.")
			return "", true
		}
		if b.Command == "" {
			if b.Notice != "" {
				v.setNotice(user.ID, b.Notice)
			}
			return "", true
		}
		return b.Command, false
	}
	if _, ok := directionOf(key); ok {
		v.setNotice(user.ID, "No exit that way.")
		return "", true
	}
	return mapKey(text), false
}

func (v *View) paint(user *entities.User, send func(any)) {
	if v == nil || v.Game == nil || user == nil || send == nil {
		return
	}
	title := v.Title
	if title == "" {
		title = gamemode.Current().Title
	}
	page := ansi.Page{
		Title:     title,
		InputMode: "hotkey",
		Prompt:    ">",
	}
	phase := v.phaseOf(user.ID)
	if phase == "name" || phase == "amount" || v.peekLine(user.ID) {
		page.InputMode = "line"
		page.Prompt = "Command:"
	}
	if phase == "amount" {
		if d := v.draftOf(user.ID); d != nil && d.prompt != "" {
			page.Prompt = d.prompt
		}
	}
	facade := v.Game.GetFacade()
	if facade == nil || phase == "name" || phase == "path" || phase == "sex" || v.noCharacter(user) {
		if phase != "path" && phase != "sex" {
			v.setPhase(user.ID, "name")
			v.setLine(user.ID, true)
			page.InputMode = "line"
			page.Prompt = "Character name:"
		}
		page.Location = "Characters"
		page.ScreenID = "select"
		page.Footer = ""
		page.Body = v.characterLines(user)
		send(ansi.Render(page))
		return
	}
	char, err := facade.CharactersService().FindByID(user.LastCharacter)
	if err != nil || char == nil {
		v.setPhase(user.ID, "name")
		page.Location = "Characters"
		page.ScreenID = "select"
		page.InputMode = "line"
		page.Prompt = "Character name:"
		page.Footer = ""
		page.Body = v.characterLines(user)
		send(ansi.Render(page))
		return
	}
	page.Location = char.Name
	page.ScreenID = char.CurrentRoomID
	hp, maxHP := char.CurrentHitPoints, char.MaxHitPoints
	if v.Game.CombatController != nil {
		if cur, mx, ok := v.Game.CombatController.PlayerHP(char.ID); ok {
			hp, maxHP = cur, mx
		}
	}
	pinned := []string{
		headerLine(char.Level, hp, maxHP, char.Gold, char.XP),
	}
	var chrome []string
	chrome = append(chrome, v.resourceLines(char.ID)...)
	var room *rooms.Room
	if found, ferr := facade.RoomsService().FindByID(char.CurrentRoomID); ferr == nil {
		room = found
	}
	inCombat := char.InCombat && v.Game.CombatController != nil
	var brief string
	if inCombat {
		brief = v.Game.CombatController.BriefStatus(char.ID)
	}
	selling := false
	if phase == "amount" {
		if d := v.draftOf(user.ID); d != nil && d.command == "sell" {
			selling = true
		}
	}
	packBinds := keysFor(room, inCombat || brief != "")
	packCount := len(packBinds)
	binds := applyOpenExits(packBinds, room, char)
	catalog := v.catalogLines(user.ID, char, room, selling)
	if brief != "" {
		pinned = append(pinned, brief)
	} else if len(catalog) > 0 {
		chrome = append(chrome, catalog...)
	} else if room != nil {
		if art := v.screenArtFor(room); art != "" {
			chrome = append(chrome, strings.Split(art, "\n")...)
		}
	}
	if note := v.peekNotice(user.ID); note != "" {
		lines := wrapPlain(note, 78)
		if len(lines) > 2 {
			lines = lines[:2]
		}
		chrome = append(chrome, lines...)
	}
	if room == nil {
		chrome = append(chrome, "You are nowhere.")
	} else {
		page.Location = room.Name
		if len(catalog) == 0 && room.Description != "" {
			lines := wrapPlain(room.Description, 78)
			limit := 2
			if brief != "" {
				limit = 1
			}
			if len(lines) > limit {
				lines = lines[:limit]
			}
			chrome = append(chrome, lines...)
		}
		if room.Exits != nil && len(*room.Exits) > 0 {
			names := make([]string, 0, len(*room.Exits))
			for _, ex := range *room.Exits {
				if ex.Hidden && !char.HasRevealedExit(room.ID, ex.Name) {
					continue
				}
				names = append(names, ex.Name)
			}
			if len(names) > 0 {
				chrome = append(chrome, "Exits: "+strings.Join(names, ", "))
			}
		}
		if packCount == 0 && len(catalog) == 0 && room.Actions != nil && len(*room.Actions) > 0 {
			names := make([]string, 0, len(*room.Actions))
			for _, action := range *room.Actions {
				if action.Name != "" {
					names = append(names, action.Name)
				}
			}
			if len(names) > 0 {
				chrome = append(chrome, "Actions: "+strings.Join(names, ", "))
			}
		}
		if len(catalog) == 0 && v.Game.NPCManager != nil {
			var npcs []string
			for _, n := range v.Game.NPCManager.GetInstancesInRoom(room.ID) {
				if n != nil && (n.MerchantTrait == nil || len(n.MerchantTrait.Inventory) == 0) {
					npcs = append(npcs, n.Name)
				}
			}
			if len(npcs) > 0 {
				chrome = append(chrome, "Here: "+strings.Join(npcs, ", "))
			}
		}
	}
	if len(binds) > 0 {
		page.Footer = commandFooter(room, char)
		page.Keys = map[string]string{}
		chrome = append(chrome, legendLines(binds)...)
		for key, b := range binds {
			if b.Command != "" {
				page.Keys[key] = b.Command
			}
		}
	} else {
		page.Footer = idleFooter(room, char)
	}
	if sheet := v.peekSheet(user.ID); len(sheet) > 0 {
		page.Body = fitBody(pinned[:1], nil, sheet, 19)
	} else {
		page.Body = fitBody(pinned, chrome, append(v.peekRecent(user.ID), v.peekFlash(user.ID)...), 19)
	}
	send(ansi.Render(page))
}

func (v *View) characterLines(user *entities.User) []string {
	phase := v.phaseOf(user.ID)
	if phase == "path" {
		lines := []string{"Choose a path."}
		if note := v.peekNotice(user.ID); note != "" {
			lines = append(lines, note)
		}
		for i, path := range loadPaths() {
			lines = append(lines, fmt.Sprintf("%d %s", i+1, path.Name))
			if path.Blurb != "" {
				lines = append(lines, "  "+path.Blurb)
			}
		}
		return lines
	}
	if phase == "sex" {
		lines := []string{"How should the square address you?"}
		if note := v.peekNotice(user.ID); note != "" {
			lines = append(lines, note)
		}
		lines = append(lines, "M Man", "F Woman", "X Neither word fits")
		return lines
	}
	lines := []string{"Type a character name."}
	if note := v.peekNotice(user.ID); note != "" {
		lines = append(lines, note)
	}
	facade := v.Game.GetFacade()
	if facade == nil {
		return lines
	}
	chars, err := facade.CharactersService().FindAllForUser(user.ID)
	if err != nil || len(chars) == 0 {
		lines = append(lines, "No characters yet.")
		return lines
	}
	for _, ch := range chars {
		if ch != nil {
			lines = append(lines, ch.Name)
		}
	}
	return lines
}

func legendLines(binds map[string]keyBind) []string {
	keys := make([]string, 0, len(binds))
	for key := range binds {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		label := binds[key].Label
		if label == "" {
			continue
		}
		shown := strings.ToUpper(key)
		if key == "?" {
			shown = "?"
		}
		parts = append(parts, shown+" "+label)
	}
	if len(parts) == 0 {
		return nil
	}
	return wrapPlain(strings.Join(parts, "   "), 78)
}

func screenArt(roomID string) string {
	root := strings.TrimSpace(gamemode.Current().WorldPack)
	if root == "" || roomID == "" {
		return ""
	}
	raw, err := os.ReadFile(filepath.Join(root, "screens", roomID+".ans"))
	if err != nil {
		return ""
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) > 8 {
		lines = lines[:8]
	}
	for len(lines) > 0 && strings.TrimSpace(stripANSI(lines[len(lines)-1])) == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

func wrapPlain(s string, width int) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if width < 8 {
		width = 8
	}
	var out []string
	for _, para := range strings.Split(s, "\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		words := strings.Fields(para)
		line := ""
		for _, word := range words {
			if line == "" {
				line = word
				continue
			}
			if len(line)+1+len(word) > width {
				out = append(out, line)
				line = word
				continue
			}
			line += " " + word
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func condense(text string) []string {
	lines := wrapPlain(text, 78)
	if len(lines) > 3 {
		lines = lines[:3]
	}
	return lines
}

func skipNotice(text string) bool {
	if text == "Quest Log" || text == "..." {
		return true
	}
	if strings.HasPrefix(text, "Queued:") || strings.HasPrefix(text, "Attack whom?") {
		return true
	}
	return false
}

func (v *View) isRoomEcho(user *entities.User, text string) bool {
	ch := v.character(user)
	if ch == nil || v.Game == nil || v.Game.GetFacade() == nil || ch.CurrentRoomID == "" {
		return false
	}
	room, err := v.Game.GetFacade().RoomsService().FindByID(ch.CurrentRoomID)
	if err != nil || room == nil || room.Name == "" {
		return false
	}
	return strings.Contains(text, "["+room.Name+"]")
}

func (v *View) resourceLines(charID string) []string {
	if v.Game == nil || v.Game.Resources == nil || charID == "" {
		return nil
	}
	var parts []string
	for _, allowance := range ruleset.ResourceAllowances() {
		res, ok, err := v.Game.Resources.Get(charID, allowance.Key)
		if err != nil || !ok {
			continue
		}
		parts = append(parts, resourceText(allowance.Key, res.Remaining, res.Allowance))
	}
	if len(parts) == 0 {
		return nil
	}
	var lines []string
	line := ""
	for _, part := range parts {
		if line == "" {
			line = part
			continue
		}
		if len(line)+3+len(part) > 78 {
			lines = append(lines, line)
			line = part
			continue
		}
		line += "   " + part
	}
	if line != "" {
		lines = append(lines, line)
	}
	if len(lines) > 2 {
		lines = lines[:2]
	}
	return lines
}

func resourceText(key string, remaining, allowance int) string {
	name := ruleset.ResourceLabel(key)
	if name == "" {
		name = key
	}
	if allowance <= 1 {
		if remaining > 0 {
			return name + ": ready"
		}
		return name + ": used"
	}
	return fmt.Sprintf("%s %d/%d", name, remaining, allowance)
}

const shopPageSize = 6

type rackRow struct {
	n    int
	line string
}

func (v *View) catalogLines(userID string, char *characters.Character, room *rooms.Room, selling bool) []string {
	if room == nil || v.Game == nil || v.Game.NPCManager == nil || v.Game.GetFacade() == nil {
		return nil
	}
	for _, n := range v.Game.NPCManager.GetInstancesInRoom(room.ID) {
		if n == nil || n.MerchantTrait == nil || len(n.MerchantTrait.Inventory) == 0 {
			continue
		}
		var rows []rackRow
		if selling {
			rows = sellRack(char, n)
		} else {
			rows = v.buyRack(n)
		}
		if len(rows) == 0 {
			if selling {
				return []string{n.Name + " will take nothing you carry."}
			}
			return []string{n.Name + " has nothing on the rack."}
		}
		pages := (len(rows) + shopPageSize - 1) / shopPageSize
		page := v.pageOf(userID)
		if page >= pages {
			page = pages - 1
		}
		if page < 0 {
			page = 0
		}
		start := page * shopPageSize
		end := start + shopPageSize
		if end > len(rows) {
			end = len(rows)
		}
		head := fmt.Sprintf("%s  %d-%d of %d", n.Name, rows[start].n, rows[end-1].n, len(rows))
		if pages > 1 {
			head += "   N next  P prev"
		}
		out := []string{clipWidth(head, 78)}
		for _, row := range rows[start:end] {
			out = append(out, clipWidth(row.line, 78))
		}
		return out
	}
	return nil
}

func (v *View) buyRack(n *npc.NPC) []rackRow {
	var rows []rackRow
	for i := range n.MerchantTrait.Inventory {
		inv := &n.MerchantTrait.Inventory[i]
		tpl, err := v.Game.GetFacade().ItemsService().FindByID(inv.ItemTemplateID)
		if err != nil || tpl == nil {
			continue
		}
		price := n.MerchantTrait.GetBuyPrice(inv, tpl.BasePrice)
		num := len(rows) + 1
		rows = append(rows, rackRow{n: num, line: formatRack(num, tpl.Name, statLabel(tpl.Attributes), price)})
	}
	return rows
}

func sellRack(char *characters.Character, n *npc.NPC) []rackRow {
	if char == nil || n == nil || n.MerchantTrait == nil {
		return nil
	}
	var rows []rackRow
	for _, item := range char.Inventory.Items {
		if item == nil || item.IsBound() {
			continue
		}
		if !n.MerchantTrait.CanBuyItem(string(item.Type), item.Tags) {
			continue
		}
		price := n.MerchantTrait.GetSellPrice(item.BasePrice)
		if price < 1 {
			price = 1
		}
		num := len(rows) + 1
		rows = append(rows, rackRow{n: num, line: formatRack(num, item.Name, statLabel(item.Attributes), price)})
	}
	return rows
}

func formatRack(n int, name, stat string, price int64) string {
	name = clipWidth(name, 22)
	if stat == "" {
		stat = "-"
	}
	return fmt.Sprintf("%2d %-22s %-8s %7d", n, name, clipWidth(stat, 8), price)
}

func statLabel(attrs map[string]interface{}) string {
	if d, ok := numberAttr(attrs, "damage"); ok {
		return fmt.Sprintf("dmg %d", d)
	}
	if d, ok := numberAttr(attrs, "defense"); ok {
		return fmt.Sprintf("def %d", d)
	}
	return ""
}

func numberAttr(attrs map[string]interface{}, key string) (int, bool) {
	if attrs == nil {
		return 0, false
	}
	switch v := attrs[key].(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		return n, err == nil
	default:
		return 0, false
	}
}

func (v *View) bumpPage(id string, next bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.shopPage == nil {
		v.shopPage = map[string]int{}
	}
	if next {
		if v.shopPage[id] < 8 {
			v.shopPage[id]++
		}
		return
	}
	if v.shopPage[id] > 0 {
		v.shopPage[id]--
	}
}

func (v *View) pageOf(id string) int {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.shopPage == nil {
		return 0
	}
	return v.shopPage[id]
}

func clipWidth(s string, width int) string {
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	return string(r[:width])
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
				j++
			}
			if j < len(s) {
				j++
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func (v *View) ensureReturning(user *entities.User) {
	if v == nil || user == nil || user.LastCharacter != "" || v.Game == nil || v.Game.GetFacade() == nil {
		return
	}
	chars, err := v.Game.GetFacade().CharactersService().FindAllForUser(user.ID)
	if err != nil || len(chars) == 0 {
		v.setPhase(user.ID, "name")
		v.setLine(user.ID, true)
		return
	}
	for _, ch := range chars {
		if ch != nil && ch.Name != "" {
			v.setPhase(user.ID, "")
			v.Game.DispatchCommand(user, "selectcharacter "+ch.Name)
			return
		}
	}
	v.setPhase(user.ID, "name")
	v.setLine(user.ID, true)
}

func (v *View) noCharacter(user *entities.User) bool {
	if user == nil || user.LastCharacter == "" || v == nil || v.Game == nil || v.Game.GetFacade() == nil {
		return true
	}
	ch, err := v.Game.GetFacade().CharactersService().FindByID(user.LastCharacter)
	return err != nil || ch == nil
}

func (v *View) character(user *entities.User) *characters.Character {
	if v.noCharacter(user) {
		return nil
	}
	ch, err := v.Game.GetFacade().CharactersService().FindByID(user.LastCharacter)
	if err != nil {
		return nil
	}
	return ch
}

func mapKey(text string) string {
	switch strings.ToLower(text) {
	case "n":
		return "north"
	case "s":
		return "south"
	case "e":
		return "east"
	case "w":
		return "west"
	case "u":
		return "up"
	case "d":
		return "down"
	case "o":
		return "out"
	case "l":
		return "look"
	case "a":
		return "attack"
	case "i":
		return "inventory"
	default:
		return text
	}
}

func (v *View) phaseOf(id string) string {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.phase == nil {
		return ""
	}
	return v.phase[id]
}

func (v *View) setPhase(id, phase string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.phase == nil {
		v.phase = map[string]string{}
	}
	if phase == "" {
		delete(v.phase, id)
		return
	}
	v.phase[id] = phase
}

func (v *View) draftOf(id string) *draft {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.drafts == nil {
		return nil
	}
	d := v.drafts[id]
	if d == nil {
		return nil
	}
	copy := *d
	return &copy
}

func (v *View) peekLine(id string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.line[id]
}

func (v *View) takeLine(id string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.line == nil || !v.line[id] {
		return false
	}
	delete(v.line, id)
	return true
}

func (v *View) setLine(id string, on bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.line == nil {
		v.line = map[string]bool{}
	}
	if on {
		v.line[id] = true
	} else {
		delete(v.line, id)
	}
}

func (v *View) peekNotice(id string) string {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.notice == nil {
		return ""
	}
	return v.notice[id]
}

func (v *View) setNotice(id, text string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.notice == nil {
		v.notice = map[string]string{}
	}
	v.notice[id] = text
}

func (v *View) clearNotice(id string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.notice, id)
}

// fitBody pins the header and, during a fight, the status line. Room art and
// the key list give up rows before those lines do. The log keeps its newest
// rows and never grows past the body.
func fitBody(pinned, chrome, log []string, limit int) []string {
	if limit < 1 {
		limit = 1
	}
	pin := append([]string{}, pinned...)
	if len(pin) > limit {
		return pin[:limit]
	}
	budget := limit - len(pin)
	rest := append([]string{}, chrome...)
	lines := append([]string{}, log...)
	for len(rest)+len(lines) > budget && len(rest) > 0 {
		rest = rest[1:]
	}
	if len(lines) > budget {
		lines = lines[len(lines)-budget:]
		rest = nil
	}
	out := append(pin, rest...)
	return append(out, lines...)
}

func (v *View) clearRecent(id string) {
	if v == nil || id == "" {
		return
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.recent, id)
}

func (v *View) pushRecent(id string, lines []string) {
	if len(lines) == 0 {
		return
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.recent == nil {
		v.recent = map[string][]string{}
	}
	next := append(append([]string{}, v.recent[id]...), lines...)
	if len(next) > recentLimit {
		next = next[len(next)-recentLimit:]
	}
	v.recent[id] = next
}

func (v *View) peekRecent(id string) []string {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.recent == nil {
		return nil
	}
	return append([]string{}, v.recent[id]...)
}
