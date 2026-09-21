package door

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
)

func (h *Hub) composeArt(sess *session, ch *characters.Character, scr *Screen) view {
	name := "stranger"
	if ch != nil && ch.Name != "" {
		name = ch.Name
	} else if sess.draftName != "" {
		name = sess.draftName
	}
	prompt := scr.Prompt
	if strings.TrimSpace(prompt) == "" {
		prompt = "Your command, {{name}}? [{{mm:ss}}] :"
	}
	prompt = strings.ReplaceAll(prompt, "{{name}}", name)
	prompt = strings.ReplaceAll(prompt, "{{mm:ss}}", h.now().In(h.loc).Format("15:04"))

	grid := make([]string, Rows)
	artLines := splitArt(scr.Art)
	limit := 21
	if len(artLines) < limit {
		limit = len(artLines)
	}
	for i := 0; i < limit; i++ {
		grid[i] = artLines[i]
	}
	switch scr.ID {
	case "stats":
		lines := h.statLines(ch)
		for i, ln := range lines {
			if 5+i < 18 {
				grid[5+i] = ln
			}
		}
		grid[18] = "  (T)own       return to the square"
		grid[19] = "  (?)          help"
	case "forest":
		if sess.fight != nil {
			lines := h.fightLines(ch, sess.fight)
			row := 14
			for _, ln := range lines {
				if row >= 20 || strings.TrimSpace(stripANSI(ln)) == "" {
					if strings.TrimSpace(stripANSI(ln)) == "" {
						continue
					}
				}
				if row >= 20 {
					break
				}
				grid[row] = ln
				row++
			}
		}
	case "bank":
		hand, vault := int64(0), int64(0)
		if ch != nil {
			hand = ch.Gold
			if ch.Door != nil {
				vault = ch.Door.BankGold
			}
		}
		grid[16] = fmt.Sprintf("  On hand: %d    Vault: %d", hand, vault)
	case "healer":
		grid[16] = fmt.Sprintf("  Full mending costs %d coin.", h.pack.HealCost)
	case "shop":
		row := 14
		for _, g := range h.shopGoods() {
			if row >= 20 {
				break
			}
			grid[row] = fmt.Sprintf("  %d  %s  %d coin  %s", g.n, g.name, g.price, g.stat)
			row++
		}
	}
	if sess.notice != "" {
		grid[20] = ansiKey + sess.notice + ansiReset
	}
	grid[21] = prompt
	grid[22] = h.footer(ch)

	keys := make(map[string]string, len(scr.Hotkeys))
	for key, hk := range scr.Hotkeys {
		keys[key] = hk.Action
	}
	mode := sess.inputMode
	if mode == "" {
		mode = "hotkey"
	}
	return view{
		Prompt:    prompt,
		InputMode: mode,
		ScreenID:  scr.ID,
		Keys:      keys,
		Lines:     grid,
	}
}

func splitArt(art string) []string {
	art = strings.ReplaceAll(art, "\r\n", "\n")
	art = strings.ReplaceAll(art, "\r", "\n")
	lines := strings.Split(art, "\n")
	for len(lines) > 0 && strings.TrimSpace(stripANSI(lines[len(lines)-1])) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
