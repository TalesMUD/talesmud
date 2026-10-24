package ruleset

import (
	"testing"
	"time"
)

func TestRankingOverlayKeepsOtherQualities(t *testing.T) {
	mu.Lock()
	saved := current
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		current = saved
		mu.Unlock()
	})

	raw := []byte("ranking:\n  cache_seconds: 3\n  quality:\n    magic: 20\n  stat_keys: [damage]\n")
	if err := LoadBytes(raw); err != nil {
		t.Fatal(err)
	}
	weights := ItemScoreWeights()
	if weights.Quality["magic"] != 20 || weights.Quality["rare"] != 24 || weights.Unknown != 10 {
		t.Fatalf("%#v", weights)
	}
	if len(weights.StatKeys) != 1 || weights.StatKeys[0] != "damage" {
		t.Fatalf("keys %#v", weights.StatKeys)
	}
	if RankingCache() != 3*time.Second {
		t.Fatalf("cache %s", RankingCache())
	}
	// Editing the returned copy must not change the profile.
	weights.Quality["magic"] = 1
	if ItemScoreWeights().Quality["magic"] != 20 {
		t.Fatal("weights leaked")
	}
}
