package combat_test

import (
	"testing"
	"time"

	centity "github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/combat"
)

func TestCheckCombatEndFleeIsNotDefeat(t *testing.T) {
	eng := combat.NewEngine(nil, combat.DefaultConfig())
	now := time.Now()

	cases := []struct {
		name     string
		players  []centity.CombatantRef
		want     centity.CombatState
		wantDead bool
		wantFled bool
	}{
		{
			name: "solo fled",
			players: []centity.CombatantRef{
				{ID: "p1", IsAlive: true, HasFled: true, CurrentHP: 21, MaxHP: 21},
			},
			want:     centity.CombatStateFled,
			wantFled: true,
		},
		{
			name: "solo dead",
			players: []centity.CombatantRef{
				{ID: "p1", IsAlive: false, CurrentHP: 0, MaxHP: 21},
			},
			want:     centity.CombatStateDefeat,
			wantDead: true,
		},
		{
			name: "one dead one fled",
			players: []centity.CombatantRef{
				{ID: "dead", IsAlive: false, CurrentHP: 0, MaxHP: 21},
				{ID: "fled", IsAlive: true, HasFled: true, CurrentHP: 21, MaxHP: 21},
			},
			want:     centity.CombatStateDefeat,
			wantDead: true,
		},
		{
			name: "one fled one active",
			players: []centity.CombatantRef{
				{ID: "fled", IsAlive: true, HasFled: true, CurrentHP: 21, MaxHP: 21},
				{ID: "up", IsAlive: true, CurrentHP: 18, MaxHP: 21},
			},
			want: centity.CombatStateActive,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inst := centity.NewCombatInstance("nest")
			inst.State = centity.CombatStateActive
			inst.CreatedAt = now
			inst.LastActionAt = now
			inst.Players = tc.players
			inst.Enemies = []centity.CombatantRef{{
				ID: "rat", Type: centity.CombatantTypeNPC, Name: "Rat", IsAlive: true, CurrentHP: 8, MaxHP: 8,
			}}
			if got := inst.AllPlayersDead(); got != tc.wantDead {
				t.Fatalf("AllPlayersDead = %v, want %v", got, tc.wantDead)
			}
			if got := inst.AllPlayersFled(); got != tc.wantFled {
				t.Fatalf("AllPlayersFled = %v, want %v", got, tc.wantFled)
			}
			if got := eng.CheckCombatEnd(inst); got != tc.want {
				t.Fatalf("CheckCombatEnd = %s, want %s", got, tc.want)
			}
		})
	}
}
