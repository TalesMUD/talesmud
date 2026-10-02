package combat_test

import (
	"testing"
	"time"

	centity "github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/combat"
)

func activeFight(created, lastAction time.Time) *centity.CombatInstance {
	inst := centity.NewCombatInstance("room-idle")
	inst.State = centity.CombatStateActive
	inst.CreatedAt = created
	inst.LastActionAt = lastAction
	inst.Players = []centity.CombatantRef{{
		ID: "p1", Type: centity.CombatantTypePlayer, Name: "Gimli", IsAlive: true, CurrentHP: 50, MaxHP: 50,
	}}
	inst.Enemies = []centity.CombatantRef{{
		ID: "e1", Type: centity.CombatantTypeNPC, Name: "Sewer Rat", IsAlive: true, CurrentHP: 32, MaxHP: 32,
	}}
	return inst
}

func TestIdleCombatTimeoutReleasesWithoutDefeat(t *testing.T) {
	cfg := combat.DefaultConfig()
	if cfg.IdleCombatTimeoutMinutes != 5 {
		t.Fatalf("IdleCombatTimeoutMinutes default = %d want 5", cfg.IdleCombatTimeoutMinutes)
	}
	if cfg.CombatTimeoutMinutes != 15 {
		t.Fatalf("CombatTimeoutMinutes default = %d want 15", cfg.CombatTimeoutMinutes)
	}
	eng := combat.NewEngine(nil, cfg)

	inst := activeFight(time.Now().Add(-2*time.Minute), time.Now().Add(-2*time.Minute))
	if got := eng.CheckCombatEnd(inst); got != centity.CombatStateActive {
		t.Fatalf("2m idle → %s want active", got)
	}

	inst.LastActionAt = time.Now().Add(-6 * time.Minute)
	if got := eng.CheckCombatEnd(inst); got != centity.CombatStateTimeout {
		t.Fatalf("6m idle → %s want timeout", got)
	}

	// Absolute backstop even if LastActionAt is fresh
	inst.LastActionAt = time.Now()
	inst.CreatedAt = time.Now().Add(-16 * time.Minute)
	if got := eng.CheckCombatEnd(inst); got != centity.CombatStateTimeout {
		t.Fatalf("16m absolute → %s want timeout", got)
	}
}

func TestIdleTimeoutDisabledWhenZero(t *testing.T) {
	cfg := combat.DefaultConfig()
	cfg.IdleCombatTimeoutMinutes = 0
	cfg.CombatTimeoutMinutes = 60
	eng := combat.NewEngine(nil, cfg)
	inst := activeFight(time.Now().Add(-30*time.Minute), time.Now().Add(-30*time.Minute))
	if got := eng.CheckCombatEnd(inst); got != centity.CombatStateActive {
		t.Fatalf("idle disabled → %s want active", got)
	}
}
