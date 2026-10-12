package commands

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

func TestLookAtNPCMarksAggressive(t *testing.T) {
	quiet := &npc.NPC{
		Entity:     &entities.Entity{ID: "rat"},
		Name:       "Rat",
		EnemyTrait: &npc.EnemyTrait{},
	}
	if strings.Contains(lookAtNPC(quiet), "Aggressive") {
		t.Fatal("passive look marked Aggressive")
	}
	wolf := &npc.NPC{
		Entity:     &entities.Entity{ID: "wolf"},
		Name:       "Wolf",
		EnemyTrait: &npc.EnemyTrait{AggroOnSight: true},
	}
	out := lookAtNPC(wolf)
	if !strings.Contains(out, "\nAggressive\n") && !strings.HasSuffix(out, "\nAggressive") {
		t.Fatalf("look = %q", out)
	}
}
