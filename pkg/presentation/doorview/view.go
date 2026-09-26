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
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/presentation/ansi"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

const recentLimit = 8

// View paints one ANSI page from live engine state.
type View struct {
	Game  *game.Game
	Title string

	mu     sync.Mutex
	line   map[string]bool
	phase  map[string]string
	drafts map[string]*draft
	notice map[string]string
	recent map[string][]string
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
	if cmd, handled := v.commandFor(user, text); handled {
		v.paint(user, send)
		return true
	} else {
		v.Game.DispatchCommand(user, cmd)
	}
	v.paint(user, send)
	return true
}

// OnNotice keeps a short line of command output and redraws the frame.
func (v *View) OnNotice(user *entities.User, text string, send func(any)) {
	if v == nil || user == nil || strings.TrimSpace(text) == "" {
		return
	}
	v.pushRecent(user.ID, condense(text))
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
	v.mu.Unlock()
}

func (v *View) handleLine(user *entities.User, text string) {
	phase := v.phaseOf(user.ID)
	switch phase {
	case "amount":
		d := v.draftOf(user.ID)
		v.setPhase(user.ID, "")
		v.setLine(user.ID, false)
		if d == nil || strings.EqualFold(text, "x") || text == "" {
			v.setNotice(user.ID, "Cancelled.")
			return
		}
		cmd := strings.TrimSpace(d.command + " " + text)
		v.Game.DispatchCommand(user, cmd)
	case "name":
		v.handleName(user, text)
	default:
		v.setLine(user.ID, false)
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
		Flags:            map[string]interface{}{"sex": sex},
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
	var roomArea, roomID string
	inCombat := false
	if ch != nil {
		inCombat = ch.InCombat
		roomID = ch.CurrentRoomID
		if v.Game.GetFacade() != nil && roomID != "" {
			if room, err := v.Game.GetFacade().RoomsService().FindByID(roomID); err == nil && room != nil {
				roomArea = room.Area
				roomID = room.ID
			}
		}
	}
	binds := keysFor(&rooms.Room{Entity: &entities.Entity{ID: roomID}, Area: roomArea}, inCombat)
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
		Footer:    "n s e w u d   l look   a attack   i inventory   : command",
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
		page.Body = v.characterLines(user)
		send(ansi.Render(page))
		return
	}
	page.Location = char.Name
	page.ScreenID = char.CurrentRoomID
	body := []string{
		fmt.Sprintf("Level %d   HP %d/%d   Gold %d", char.Level, char.CurrentHitPoints, char.MaxHitPoints, char.Gold),
	}
	if v.Game.Resources != nil {
		shown := 0
		for _, allowance := range ruleset.ResourceAllowances() {
			res, ok, err := v.Game.Resources.Get(char.ID, allowance.Key)
			if err == nil && ok {
				body = append(body, fmt.Sprintf("%s %d/%d", allowance.Key, res.Remaining, res.Allowance))
				shown++
				if shown >= 4 {
					break
				}
			}
		}
	}
	var room *rooms.Room
	if found, ferr := facade.RoomsService().FindByID(char.CurrentRoomID); ferr == nil {
		room = found
	}
	inCombat := char.InCombat && v.Game.CombatController != nil
	var brief string
	if inCombat {
		brief = v.Game.CombatController.BriefStatus(char.ID)
		if brief != "" {
			body = append(body, brief)
		}
	}
	if note := v.peekNotice(user.ID); note != "" {
		body = append(body, wrapPlain(note, 78)...)
	}
	recent := v.peekRecent(user.ID)
	if len(recent) > 0 {
		body = append(body, "Recent:")
		body = append(body, recent...)
	}
	binds := keysFor(room, inCombat || brief != "")
	if len(binds) > 0 {
		page.Footer = "d down   : command"
		page.Keys = map[string]string{}
		body = append(body, legendLines(binds)...)
		for key, b := range binds {
			if b.Command != "" {
				page.Keys[key] = b.Command
			}
		}
	}
	if room == nil {
		body = append(body, "You are nowhere.")
	} else {
		page.Location = room.Name
		if len(recent) == 0 && brief == "" {
			if art := screenArt(room.ID); art != "" {
				body = append(body, strings.Split(art, "\n")...)
			}
		}
		limit := 6
		if len(recent) > 0 || brief != "" {
			limit = 3
		}
		if room.Description != "" {
			lines := wrapPlain(room.Description, 78)
			if len(lines) > limit {
				lines = lines[:limit]
			}
			body = append(body, lines...)
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
				body = append(body, "Exits: "+strings.Join(names, ", "))
			}
		}
		if len(binds) == 0 && room.Actions != nil && len(*room.Actions) > 0 {
			names := make([]string, 0, len(*room.Actions))
			for _, action := range *room.Actions {
				if action.Name != "" {
					names = append(names, action.Name)
				}
			}
			if len(names) > 0 {
				body = append(body, "Actions: "+strings.Join(names, ", "))
			}
		}
		if v.Game.NPCManager != nil {
			var npcs []string
			for _, n := range v.Game.NPCManager.GetInstancesInRoom(room.ID) {
				if n != nil {
					npcs = append(npcs, n.Name)
				}
			}
			if len(npcs) > 0 {
				body = append(body, "Here: "+strings.Join(npcs, ", "))
			}
		}
	}
	page.Body = body
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
	if len(lines) > 4 {
		lines = lines[:4]
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
		lines = append(lines[:3], "...")
	}
	return lines
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
