package importer_test

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/exporter"
	"github.com/talesmud/talesmud/pkg/importer"
	"gopkg.in/yaml.v3"
)

func TestMovementImportExportAndLegacyFields(t *testing.T) {
	raw := []byte(`id: NPC_TEST
name: Walker
maxHitPoints: 100
spawnRoomId: a
wanderRadius: 2
patrolPath: [a, b]
movement:
  mode: route
  route: [a, b]
  loop: pingpong
  zones: [town]
  intervalSeconds: 90
  jitterSeconds: 30
  pauseOnEngageSeconds: 120
`)
	var source importer.YAMLNPC
	if err := yaml.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	n := source.ToEntity()
	if n.Movement == nil || n.Movement.Mode != "route" || n.WanderRadius != 2 || len(n.PatrolPath) != 2 {
		t.Fatal(n)
	}
	output, err := exporter.ContentNPCYAML(n)
	if err != nil {
		t.Fatal(err)
	}
	var again importer.YAMLNPC
	if err := yaml.Unmarshal(output, &again); err != nil {
		t.Fatal(err)
	}
	if again.Movement.Loop != "pingpong" || again.Movement.Route[1] != "b" || again.WanderRadius != 2 || again.PatrolPath[1] != "b" {
		t.Fatal(again)
	}
	var spawner importer.YAMLSpawner
	if err := yaml.Unmarshal([]byte("id: s\ntemplateId: NPC_TEST\nroomId: a\nmovement:\n  mode: roam\n  zones: [town]\n"), &spawner); err != nil {
		t.Fatal(err)
	}
	if spawner.ToEntity().Movement.Mode != "roam" {
		t.Fatal("spawner override lost")
	}
}
