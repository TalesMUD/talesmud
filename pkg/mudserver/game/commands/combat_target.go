package commands

import (
	"sort"
	"strconv"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
)

// resolveInCombatTarget finds a living combat enemy by exact ID, room UI label
// (Name#N from BuildNPCDisplayNames / FindInstanceByNameInRoom), or partial name.
// Room-based #N is preferred so labels match the room/combat UI even when the
// combat pack order differs from GetInstancesInRoom order.
func resolveInCombatTarget(game def.GameCtrl, roomID string, enemies []combat.CombatantRef, query string) (id, name string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", ""
	}

	livingIDs := make(map[string]string, len(enemies))
	for i := range enemies {
		e := &enemies[i]
		if e.IsAlive {
			livingIDs[e.ID] = e.Name
		}
	}

	// Exact combatant ID
	for i := range enemies {
		e := &enemies[i]
		if e.IsAlive && strings.EqualFold(e.ID, query) {
			return e.ID, e.Name
		}
	}

	// Room instance lookup (parses Name#N like initiate combat)
	if game != nil && roomID != "" {
		if mgr := game.GetNPCInstanceManager(); mgr != nil {
			if inst := mgr.FindInstanceByNameInRoom(roomID, query); inst != nil && inst.Entity != nil {
				if n, ok := livingIDs[inst.Entity.ID]; ok {
					return inst.Entity.ID, n
				}
			}
		}
	}

	// Fallback: #N / partial among living combatants (combat pack order)
	return resolveLivingEnemyByQuery(enemies, query)
}

// resolveLivingEnemyByQuery finds a living combat enemy by exact ID, UI display
// label (Name#N), or partial name among the combatant list order.
func resolveLivingEnemyByQuery(enemies []combat.CombatantRef, query string) (id, name string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", ""
	}

	for i := range enemies {
		e := &enemies[i]
		if e.IsAlive && strings.EqualFold(e.ID, query) {
			return e.ID, e.Name
		}
	}

	queryLower := strings.ToLower(query)
	if idx := strings.LastIndex(queryLower, "#"); idx != -1 {
		baseName := strings.TrimSpace(query[:idx])
		numStr := query[idx+1:]
		if num := parsePositiveIntSuffix(numStr); num > 0 {
			count := 0
			baseLower := strings.ToLower(baseName)
			for i := range enemies {
				e := &enemies[i]
				if !e.IsAlive {
					continue
				}
				if strings.Contains(strings.ToLower(e.Name), baseLower) {
					count++
					if count == num {
						return e.ID, e.Name
					}
				}
			}
			return "", ""
		}
	}

	for i := range enemies {
		e := &enemies[i]
		if e.IsAlive && strings.Contains(strings.ToLower(e.Name), queryLower) {
			return e.ID, e.Name
		}
	}
	return "", ""
}

// livingEnemyDisplayLabels mirrors util.BuildNPCDisplayNames for combatants so
// available-target hints match the room/combat UI (Sewer Rat#2, …).
func livingEnemyDisplayLabels(enemies []combat.CombatantRef) map[string]string {
	nameCounts := make(map[string]int)
	living := make([]*combat.CombatantRef, 0, len(enemies))
	for i := range enemies {
		e := &enemies[i]
		if !e.IsAlive {
			continue
		}
		living = append(living, e)
		nameCounts[e.Name]++
	}
	sort.Slice(living, func(i, j int) bool { return living[i].ID < living[j].ID })
	nameIndex := make(map[string]int)
	result := make(map[string]string, len(living))
	for _, e := range living {
		if nameCounts[e.Name] > 1 {
			nameIndex[e.Name]++
			result[e.ID] = e.Name + "#" + strconv.Itoa(nameIndex[e.Name])
		} else {
			result[e.ID] = e.Name
		}
	}
	return result
}

func parsePositiveIntSuffix(s string) int {
	if len(s) == 0 {
		return 0
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
