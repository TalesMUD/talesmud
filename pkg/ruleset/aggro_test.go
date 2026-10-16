package ruleset_test

import (
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/ruleset"
)

func TestAggroOnSightDefaults(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	if !ruleset.AggroOnSightEnabled() {
		t.Fatal("omitted and shipped aggro_on_sight stays enabled")
	}
	if ruleset.AggroGrace() != 2500*time.Millisecond {
		t.Fatalf("grace = %s", ruleset.AggroGrace())
	}
	if ruleset.AggroMaxLevelGap() != 5 {
		t.Fatalf("gap = %d", ruleset.AggroMaxLevelGap())
	}
	if ruleset.AggroReaggroCooldown() != 15*time.Second {
		t.Fatalf("cooldown = %s", ruleset.AggroReaggroCooldown())
	}
	if ruleset.RespawnMin() != 30*time.Second {
		t.Fatalf("respawn floor = %s", ruleset.RespawnMin())
	}
	if ruleset.RespawnAggroGrace() != 8*time.Second || ruleset.PostCombatAggroGrace() != 10*time.Second || ruleset.SpawnAggroCooldown() != 60*time.Second {
		t.Fatalf("quiet windows respawn=%s post=%s spawn=%s", ruleset.RespawnAggroGrace(), ruleset.PostCombatAggroGrace(), ruleset.SpawnAggroCooldown())
	}
}

func TestAggroOnSightClampAndOverride(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)

	if err := ruleset.LoadBytes([]byte("combat:\n  aggro_on_sight:\n    enabled: false\n    grace_seconds: 0\n    max_level_gap: -3\n    reaggro_cooldown_seconds: 0\n")); err != nil {
		t.Fatal(err)
	}
	if ruleset.AggroOnSightEnabled() {
		t.Fatal("explicit enabled false")
	}
	if ruleset.AggroGrace() != 500*time.Millisecond {
		t.Fatalf("grace 0 clamped to %s", ruleset.AggroGrace())
	}
	if ruleset.AggroMaxLevelGap() != 0 {
		t.Fatalf("negative gap = %d", ruleset.AggroMaxLevelGap())
	}
	if ruleset.AggroReaggroCooldown() != 0 {
		t.Fatalf("explicit 0 cooldown = %s", ruleset.AggroReaggroCooldown())
	}

	if err := ruleset.LoadBytes([]byte("combat:\n  aggro_on_sight:\n    grace_seconds: 99\n")); err != nil {
		t.Fatal(err)
	}
	if !ruleset.AggroOnSightEnabled() {
		t.Fatal("omitted enabled stays on when the block is present")
	}
	if ruleset.AggroGrace() != 10*time.Second {
		t.Fatalf("grace 99 clamped to %s", ruleset.AggroGrace())
	}
	if ruleset.AggroMaxLevelGap() != 5 || ruleset.AggroReaggroCooldown() != 15*time.Second {
		t.Fatalf("omitted gap/cooldown = %d %s", ruleset.AggroMaxLevelGap(), ruleset.AggroReaggroCooldown())
	}
	if ruleset.RespawnAggroGrace() != 8*time.Second || ruleset.PostCombatAggroGrace() != 10*time.Second || ruleset.SpawnAggroCooldown() != 60*time.Second || ruleset.RespawnMin() != 30*time.Second {
		t.Fatal("omitted quiet windows and respawn floor keep their defaults")
	}

	if err := ruleset.LoadBytes([]byte("combat:\n  respawn_min_seconds: 0\n  aggro_on_sight:\n    respawn_grace_seconds: 0\n    post_combat_grace_seconds: 0\n    spawn_aggro_cooldown_seconds: 0\n")); err != nil {
		t.Fatal(err)
	}
	if ruleset.RespawnMin() != 0 || ruleset.RespawnAggroGrace() != 0 || ruleset.PostCombatAggroGrace() != 0 || ruleset.SpawnAggroCooldown() != 0 {
		t.Fatal("explicit 0 disables the floor and the quiet windows")
	}
	if !ruleset.AggroOnSightEnabled() || ruleset.AggroGrace() != 2500*time.Millisecond {
		t.Fatal("explicit 0 on the new keys must leave sight enabled")
	}

	if err := ruleset.LoadBytes([]byte("combat:\n  respawn_min_seconds: -5\n  aggro_on_sight:\n    respawn_grace_seconds: -1\n")); err != nil {
		t.Fatal(err)
	}
	if ruleset.RespawnMin() != 0 || ruleset.RespawnAggroGrace() != 0 {
		t.Fatalf("negative floor=%s respawn grace=%s", ruleset.RespawnMin(), ruleset.RespawnAggroGrace())
	}
	if ruleset.PostCombatAggroGrace() != 10*time.Second {
		t.Fatal("omitted post-combat grace stays at the default")
	}

	ruleset.SetAggroQuiet(3*time.Second, 4*time.Second, 5*time.Second)
	ruleset.SetAggroOnSight(true, time.Millisecond, 0, -1)
	if ruleset.AggroGrace() != 500*time.Millisecond || ruleset.AggroMaxLevelGap() != 0 || ruleset.AggroReaggroCooldown() != 0 {
		t.Fatalf("setter clamp grace=%s gap=%d cd=%s", ruleset.AggroGrace(), ruleset.AggroMaxLevelGap(), ruleset.AggroReaggroCooldown())
	}
	if ruleset.RespawnAggroGrace() != 3*time.Second || ruleset.PostCombatAggroGrace() != 4*time.Second || ruleset.SpawnAggroCooldown() != 5*time.Second {
		t.Fatal("SetAggroOnSight must keep the quiet windows")
	}
	ruleset.SetRespawnMin(-time.Second)
	if ruleset.RespawnMin() != 0 {
		t.Fatalf("negative floor setter = %s", ruleset.RespawnMin())
	}
}
