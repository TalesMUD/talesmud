package contenthealth

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func TestMovementHealthReportsInvalidRoutesAndSpawnerOverrides(t *testing.T) {
	ab := rooms.Exits{{Name: "east", Target: "b"}}
	ba := rooms.Exits{{Name: "west", Target: "a"}}
	w := World{StartRoomID: "a", Rooms: []*rooms.Room{{Entity: &entities.Entity{ID: "a"}, Exits: &ab}, {Entity: &entities.Entity{ID: "b"}, Exits: &ba}}, NPCs: []*npc.NPC{{Entity: &entities.Entity{ID: "template"}, IsTemplate: true, Movement: &npc.Movement{Mode: "route", Route: []string{"a", "b"}}}}, Spawners: []*npc.NPCSpawner{{Entity: &entities.Entity{ID: "s"}, TemplateID: "template", RoomID: "a"}}}
	report := Run(w, Options{})
	for _, rule := range report.Rules {
		if rule.ID == RuleMovement && len(rule.Hits) > 0 {
			t.Fatal(rule.Hits)
		}
	}
	w.Spawners[0].Movement = &npc.Movement{Mode: "route", Route: []string{"a", "missing"}}
	report = Run(w, Options{})
	found := false
	for _, rule := range report.Rules {
		if rule.ID == RuleMovement {
			for _, hit := range rule.Hits {
				if hit.EntityType == "spawner" && hit.EntityID == "s" && hit.Severity == "error" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("invalid spawner movement not an error")
	}
}
