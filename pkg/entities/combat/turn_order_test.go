package combat

import "testing"

func orderOf(ids ...string) []CombatantRef {
	out := make([]CombatantRef, len(ids))
	for i, id := range ids {
		out[i] = CombatantRef{ID: id, Name: id, IsAlive: true, Type: CombatantTypeNPC}
	}
	return out
}

func indexOf(ids []string, id string) int {
	for i, candidate := range ids {
		if candidate == id {
			return i
		}
	}
	return -1
}

func TestRemoveFromTurnOrderIndexInvariant(t *testing.T) {
	cases := []struct {
		name        string
		ids         []string
		current     string
		remove      string
		wantIDs     []string
		wantCurrent string
		wantIdx     int
	}{
		{
			name:        "remove before current",
			ids:         []string{"a", "b", "c", "d"},
			current:     "c",
			remove:      "a",
			wantIDs:     []string{"b", "c", "d"},
			wantCurrent: "c",
			wantIdx:     1,
		},
		{
			name:        "remove at current",
			ids:         []string{"a", "b", "c", "d"},
			current:     "b",
			remove:      "b",
			wantIDs:     []string{"a", "c", "d"},
			wantCurrent: "c",
			wantIdx:     1,
		},
		{
			name:        "remove after current",
			ids:         []string{"a", "b", "c", "d"},
			current:     "b",
			remove:      "d",
			wantIDs:     []string{"a", "b", "c"},
			wantCurrent: "b",
			wantIdx:     1,
		},
		{
			name:        "remove last slot while it is current",
			ids:         []string{"a", "b", "c"},
			current:     "c",
			remove:      "c",
			wantIDs:     []string{"a", "b"},
			wantCurrent: "a",
			wantIdx:     0,
		},
		{
			name:        "remove last slot after current",
			ids:         []string{"a", "b", "c"},
			current:     "a",
			remove:      "c",
			wantIDs:     []string{"a", "b"},
			wantCurrent: "a",
			wantIdx:     0,
		},
		{
			name:        "remove the only combatant",
			ids:         []string{"a"},
			current:     "a",
			remove:      "a",
			wantIDs:     nil,
			wantCurrent: "",
			wantIdx:     0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inst := NewCombatInstance("R")
			inst.TurnOrder = orderOf(tc.ids...)
			inst.CurrentTurnIdx = indexOf(tc.ids, tc.current)
			if inst.CurrentTurnIdx < 0 {
				t.Fatalf("current %s not in order", tc.current)
			}

			inst.RemoveFromTurnOrder(tc.remove)

			if len(inst.TurnOrder) == 0 {
				if inst.CurrentTurnIdx != 0 {
					t.Fatalf("empty order index = %d, want 0", inst.CurrentTurnIdx)
				}
			} else if inst.CurrentTurnIdx < 0 || inst.CurrentTurnIdx >= len(inst.TurnOrder) {
				t.Fatalf("index %d past order len %d", inst.CurrentTurnIdx, len(inst.TurnOrder))
			}
			if inst.CurrentTurnIdx != tc.wantIdx {
				t.Fatalf("index = %d, want %d", inst.CurrentTurnIdx, tc.wantIdx)
			}
			got := make([]string, len(inst.TurnOrder))
			for i := range inst.TurnOrder {
				got[i] = inst.TurnOrder[i].ID
			}
			if len(got) != len(tc.wantIDs) {
				t.Fatalf("order = %v, want %v", got, tc.wantIDs)
			}
			for i := range got {
				if got[i] != tc.wantIDs[i] {
					t.Fatalf("order = %v, want %v", got, tc.wantIDs)
				}
			}
			current := inst.GetCurrentTurnCombatant()
			if tc.wantCurrent == "" {
				if current != nil {
					t.Fatalf("current = %s, want none", current.ID)
				}
				return
			}
			if current == nil || current.ID != tc.wantCurrent {
				gotID := ""
				if current != nil {
					gotID = current.ID
				}
				t.Fatalf("current = %s, want %s", gotID, tc.wantCurrent)
			}
		})
	}
}
