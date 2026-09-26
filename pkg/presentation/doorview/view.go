// Package doorview is a text client over the shared engine.
// It reads rooms, exits, NPCs, resources, and combat status, and it sends
// normal engine commands. It does not keep hit points, prices, or fight results.
package doorview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/presentation/ansi"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

// View paints one ANSI page from live engine state.
type View struct {
	Game  *game.Game
	Title string

	mu   sync.Mutex
	line map[string]bool
}

// Active reports whether this process is serving the text client.
func (v *View) Active() bool {
	return v != nil && gamemode.ANSI()
}

// OnConnect paints the current room, or a character prompt.
func (v *View) OnConnect(user *entities.User, send func(any)) {
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
	if v.takeLine(user.ID) {
		v.Game.DispatchCommand(user, text)
		v.paint(user, send)
		return true
	}
	if text == ":" {
		v.setLine(user.ID, true)
		v.paint(user, send)
		return true
	}
	v.Game.DispatchCommand(user, mapKey(text))
	v.paint(user, send)
	return true
}

// OnDisconnect drops line-mode state for the user.
func (v *View) OnDisconnect(user *entities.User) {
	if v == nil || user == nil {
		return
	}
	v.setLine(user.ID, false)
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
	if v.peekLine(user.ID) {
		page.InputMode = "line"
		page.Prompt = "Command:"
	}
	facade := v.Game.GetFacade()
	if facade == nil || user.LastCharacter == "" {
		page.Location = "Characters"
		page.ScreenID = "select"
		page.InputMode = "line"
		page.Prompt = "Character name:"
		page.Body = v.characterLines(user)
		send(ansi.Render(page))
		return
	}
	char, err := facade.CharactersService().FindByID(user.LastCharacter)
	if err != nil || char == nil {
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
		for _, allowance := range ruleset.ResourceAllowances() {
			res, ok, err := v.Game.Resources.Get(char.ID, allowance.Key)
			if err == nil && ok {
				body = append(body, fmt.Sprintf("%s %d/%d", allowance.Key, res.Remaining, res.Allowance))
			}
		}
	}
	room, err := facade.RoomsService().FindByID(char.CurrentRoomID)
	if err != nil || room == nil {
		body = append(body, "You are nowhere.")
	} else {
		page.Location = room.Name
		if art := screenArt(room.ID); art != "" {
			body = append(body, "")
			body = append(body, strings.Split(art, "\n")...)
		}
		if room.Description != "" {
			body = append(body, "", room.Description)
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
				body = append(body, "", "Exits: "+strings.Join(names, ", "))
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
	if char.InCombat && v.Game.CombatController != nil {
		if status := v.Game.CombatController.GetCombatStatus(char.ID); status != "" {
			body = append(body, "", status)
		}
	}
	page.Body = body
	send(ansi.Render(page))
}

func (v *View) characterLines(user *entities.User) []string {
	lines := []string{"Type a character name."}
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
	return strings.Join(lines, "\n")
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
