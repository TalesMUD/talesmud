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
		grid[12] = "  Others on the door"
		leaders := h.leaderLines()
		for i, ln := range leaders {
			if 13+i >= 18 {
				break
			}
			grid[13+i] = "  " + ln
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
	case "shop", "armory", "weapons", "armor":
		row := 4
		for _, g := range h.shopGoodsFor(scr.ID) {
			if row >= 17 {
				break
			}
			name := g.name
			if len(name) > 28 {
				name = name[:28]
			}
			dots := 30 - len(name)
			if dots < 2 {
				dots = 2
			}
			grid[row] = fmt.Sprintf(" │ %2d. %s%s%12s", g.n, name, strings.Repeat(".", dots), comma(g.price))
			row++
		}
		grid[17] = " │ (B)uy   (S)ell   (T)own   (?)"
	case "news":
		row := 3
		for _, line := range h.pack.News {
			if row >= 19 {
				break
			}
			grid[row] = "  - " + line
			row++
		}
	case "board":
		grid[3] = "  #   NAME               LEVEL   EXPERIENCE"
		row := 4
		for _, ln := range h.leaderLines() {
			if row >= 18 {
				break
			}
			grid[row] = "  " + ln
			row++
		}
	case "inn":
		if ch != nil {
			grid[15] = fmt.Sprintf("  HP %d/%d — one ember draught softens the day.", ch.CurrentHitPoints, ch.MaxHitPoints)
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

func comma(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}
