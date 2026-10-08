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

	ruleset.SetAggroOnSight(true, time.Millisecond, 0, -1)
	if ruleset.AggroGrace() != 500*time.Millisecond || ruleset.AggroMaxLevelGap() != 0 || ruleset.AggroReaggroCooldown() != 0 {
		t.Fatalf("setter clamp grace=%s gap=%d cd=%s", ruleset.AggroGrace(), ruleset.AggroMaxLevelGap(), ruleset.AggroReaggroCooldown())
	}
}
