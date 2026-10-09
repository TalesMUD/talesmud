package contenthealth

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/skills"
)

func TestCanonicalIgnoresVolatileFields(t *testing.T) {
	base := &npc.NPC{Entity: &entities.Entity{ID: "N"}, Name: "N", MaxHitPoints: 10, CurrentHitPoints: 10, State: "idle"}
	hurt := &npc.NPC{Entity: &entities.Entity{ID: "N"}, Name: "N", MaxHitPoints: 10, CurrentHitPoints: 1, State: "combat", InCombat: true}
	left, _, err := Hash(base)
	if err != nil {
		t.Fatal(err)
	}
	right, _, err := Hash(hurt)
	if err != nil {
		t.Fatal(err)
	}
	if left != right {
		t.Fatalf("volatile fields changed the hash\n%s\n%s", left, right)
	}
	renamed := &npc.NPC{Entity: &entities.Entity{ID: "N"}, Name: "Other", MaxHitPoints: 10}
	changed, _, err := Hash(renamed)
	if err != nil {
		t.Fatal(err)
	}
	if changed == left {
		t.Fatal("a content field change should change the hash")
	}
}

func TestDriftIgnoresEngineSkillsAndReportsContent(t *testing.T) {
	room := &rooms.Room{Entity: &entities.Entity{ID: "R0"}, Name: "Start", Description: "Old"}
	world := World{Rooms: []*rooms.Room{room}}
	base := &Baseline{Entries: baselineEntries(world)}
	room.Description = "New"
	world.Skills = []*skills.Skill{{Entity: &entities.Entity{ID: "warrior_power_strike"}, Name: "Power Strike"}}
	world.NPCs = []*npc.NPC{{Entity: &entities.Entity{ID: "ENM"}, Name: "New NPC", IsTemplate: true}}
	rows := compareDrift(base, baselineEntries(world))
	kinds := map[string]string{}
	for _, row := range rows {
		kinds[row.Type+":"+row.ID] = row.Kind
	}
	if kinds["skill:warrior_power_strike"] != "" {
		t.Fatalf("engine skill should be ignored: %+v", rows)
	}
	if kinds["npc:ENM"] != "added" {
		t.Fatalf("new npc = %q", kinds["npc:ENM"])
	}
	if kinds["room:R0"] != "changed" {
		t.Fatalf("room change = %q rows=%+v", kinds["room:R0"], rows)
	}
	var fields string
	for _, row := range rows {
		if row.ID == "R0" {
			for _, field := range row.Fields {
				fields += field.Path + " "
			}
		}
	}
	if !strings.Contains(fields, "description") {
		t.Fatalf("field diff = %s", fields)
	}
}
