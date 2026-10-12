package importer

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestSpawnerRespawnDelayImport(t *testing.T) {
	var y YAMLSpawner
	if err := yaml.Unmarshal([]byte("id: SPW_R0228_ENM0009\ntemplateId: ENM0009\nroomId: R0228\nmaxInstances: 1\nspawnInterval: 30m\ninitialCount: 1\nrespawnDelay: 30m\n"), &y); err != nil {
		t.Fatal(err)
	}
	if got := y.ToEntity().RespawnDelay; got != 30*time.Minute {
		t.Fatalf("RespawnDelay = %v", got)
	}
	y.RespawnDelay = ""
	if got := y.ToEntity().RespawnDelay; got != 0 {
		t.Fatalf("empty RespawnDelay must be 0, got %v", got)
	}
	y.RespawnDelay = "soon"
	if got := y.ToEntity().RespawnDelay; got != 0 {
		t.Fatalf("invalid RespawnDelay must be 0, got %v", got)
	}
}

func TestEnemyResetOnDisengageImport(t *testing.T) {
	var y YAMLEnemyTrait
	if err := yaml.Unmarshal([]byte("difficulty: hard\nresetOnDisengage: true\n"), &y); err != nil {
		t.Fatal(err)
	}
	if y.ResetOnDisengage == nil || !*y.ResetOnDisengage {
		t.Fatal("resetOnDisengage not parsed")
	}
}
