package game

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/combat"
)

// Party Loot & XP Share v1.
//
// On victory, gold and XP are split equally among the living combatants plus
// any online party members standing in the killer's room. Offline members and
// members in another room get nothing. Item drops stay on the ground.
//
// Remainder: integer division. When two or more recipients belong to the same
// party, leftover gold and XP (total % n) go to the primary victor — the first
// living combatant, who engaged the fight — so nothing is discarded. A solo
// victor, or a fight with no shared party, keeps the previous rule: each living
// combatant gets total/n and any remainder is dropped.

// plannedShare is one character's slice of a victory's gold and XP.
type plannedShare struct {
	ID      string
	Name    string
	InFight bool
	XP      int64
	Gold    int64
}

// splitVictoryAmount divides total across ids.
// When keepRemainder is set, total%n is added to bonusID (or ids[0]).
// A single recipient always receives the full total.
func splitVictoryAmount(total int64, ids []string, bonusID string, keepRemainder bool) map[string]int64 {
	out := make(map[string]int64, len(ids))
	n := int64(len(ids))
	if n == 0 {
		return out
	}
	if total < 0 {
		total = 0
	}
	if n == 1 {
		out[ids[0]] = total
		return out
	}
	base := total / n
	rem := int64(0)
	if keepRemainder {
		rem = total % n
	}
	for _, id := range ids {
		out[id] = base
	}
	if rem > 0 {
		target := bonusID
		if _, ok := out[target]; !ok {
			target = ids[0]
		}
		out[target] += rem
	}
	return out
}

// planVictoryShares decides who receives this victory's gold and XP.
// partySplit is true when at least two recipients are in the same party, which
// is when the leftover is kept and the share summary is shown.
func (c *CombatController) planVictoryShares(instance *combat.CombatInstance, living []*combat.CombatantRef, totalXP, totalGold int64) ([]plannedShare, bool) {
	if c == nil || instance == nil || len(living) == 0 {
		return nil, false
	}

	killerID := ""
	if living[0] != nil {
		killerID = living[0].ID
	}
	fightRoom := instance.OriginRoomID
	if c.game != nil && c.game.Facade != nil && killerID != "" {
		if killer, err := c.game.Facade.CharactersService().FindByID(killerID); err == nil && killer != nil && killer.CurrentRoomID != "" {
			fightRoom = killer.CurrentRoomID
		}
	}

	seen := map[string]bool{}
	memberParty := map[string]string{}
	prelim := make([]plannedShare, 0, len(living))

	noteParty := func(characterID string) {
		if c.game == nil || c.game.Facade == nil || characterID == "" {
			return
		}
		party, err := c.game.Facade.PartiesService().FindByCharacterID(characterID)
		if err != nil || party == nil {
			return
		}
		for _, id := range party.Characters {
			if _, ok := memberParty[id]; !ok {
				memberParty[id] = party.ID
			}
		}
	}

	for _, p := range living {
		if p == nil || p.ID == "" || seen[p.ID] {
			continue
		}
		seen[p.ID] = true
		name := p.Name
		if name == "" {
			name = p.ID
		}
		prelim = append(prelim, plannedShare{ID: p.ID, Name: name, InFight: true})
		noteParty(p.ID)
	}

	online := map[string]bool{}
	if c.game != nil {
		for _, op := range c.game.GetOnlinePlayers() {
			if op.CharacterID != "" {
				online[op.CharacterID] = true
			}
		}
	}

	// Same-room online party members who did not join the fight still share.
	if c.game != nil && c.game.Facade != nil {
		visitedParty := map[string]bool{}
		for _, p := range living {
			if p == nil || p.ID == "" {
				continue
			}
			party, err := c.game.Facade.PartiesService().FindByCharacterID(p.ID)
			if err != nil || party == nil || visitedParty[party.ID] {
				continue
			}
			visitedParty[party.ID] = true
			for _, memberID := range party.Characters {
				if seen[memberID] || !online[memberID] {
					continue
				}
				member, err := c.game.Facade.CharactersService().FindByID(memberID)
				if err != nil || member == nil {
					continue
				}
				if fightRoom == "" || member.CurrentRoomID != fightRoom {
					continue
				}
				seen[memberID] = true
				name := member.Name
				if name == "" {
					name = memberID
				}
				prelim = append(prelim, plannedShare{ID: member.ID, Name: name, InFight: false})
			}
		}
	}

	counts := map[string]int{}
	partySplit := false
	for _, share := range prelim {
		pid := memberParty[share.ID]
		if pid == "" {
			continue
		}
		counts[pid]++
		if counts[pid] >= 2 {
			partySplit = true
		}
	}

	ids := make([]string, len(prelim))
	for i, share := range prelim {
		ids[i] = share.ID
	}
	xpParts := splitVictoryAmount(totalXP, ids, killerID, partySplit)
	goldParts := splitVictoryAmount(totalGold, ids, killerID, partySplit)
	for i := range prelim {
		prelim[i].XP = xpParts[prelim[i].ID]
		prelim[i].Gold = goldParts[prelim[i].ID]
	}
	return prelim, partySplit
}

func partyShareSummary(shares []plannedShare, killerName string) string {
	if killerName == "" {
		killerName = "the victor"
	}
	var sb strings.Builder
	sb.WriteString("PARTY SHARE (equal split; leftover to ")
	sb.WriteString(killerName)
	sb.WriteString("):\n")
	for _, share := range shares {
		fmt.Fprintf(&sb, "  %s: +%d XP, +%d Gold\n", share.Name, share.XP, share.Gold)
	}
	return sb.String()
}

func partyShareToast(shares []plannedShare, killerName string) string {
	if killerName == "" {
		killerName = "the victor"
	}
	parts := make([]string, 0, len(shares))
	for _, share := range shares {
		parts = append(parts, fmt.Sprintf("%s +%d XP, +%d Gold", share.Name, share.XP, share.Gold))
	}
	return fmt.Sprintf("[Party] Equal split; leftover to %s: %s.", killerName, strings.Join(parts, "; "))
}

func formatCombatVictoryText(enemyNames, lootItems []string, xp, gold int64, shareBlock, breakdown string) string {
	var sb strings.Builder
	sb.WriteString("\n═══════════════════════════════════════════════════\n")
	sb.WriteString("              VICTORY!\n")
	sb.WriteString("═══════════════════════════════════════════════════\n\n")
	for _, name := range enemyNames {
		sb.WriteString(fmt.Sprintf("Defeated: %s\n", name))
	}
	sb.WriteString("\nREWARDS:\n")
	if breakdown != "" {
		sb.WriteString(breakdown)
		if !strings.HasSuffix(breakdown, "\n") {
			sb.WriteString("\n")
		}
	}
	if xp > 0 {
		sb.WriteString(fmt.Sprintf("  + %d XP\n", xp))
	}
	if gold > 0 {
		sb.WriteString(fmt.Sprintf("  + %d Gold\n", gold))
	}
	if len(lootItems) > 0 {
		sb.WriteString("\nLOOT DROPPED:\n")
		for _, itemName := range lootItems {
			sb.WriteString(fmt.Sprintf("  - %s\n", itemName))
		}
	}
	if xp == 0 && gold == 0 && len(lootItems) == 0 {
		sb.WriteString("  (none)\n")
	}
	if shareBlock != "" {
		sb.WriteString("\n")
		sb.WriteString(shareBlock)
	}
	sb.WriteString("\n═══════════════════════════════════════════════════")
	return sb.String()
}
