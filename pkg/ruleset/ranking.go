package ruleset

import (
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/ranking"
)

// rankingFile is the optional ruleset block. Nil pointers keep the engine default.
type rankingFile struct {
	CacheSeconds   *int           `yaml:"cache_seconds"`
	UnknownQuality *int           `yaml:"unknown_quality"`
	Quality        map[string]int `yaml:"quality"`
	StatKeys       []string       `yaml:"stat_keys"`
}

type rankingProfile struct {
	cache   time.Duration
	weights ranking.Weights
}

func defaultRanking() rankingProfile {
	return rankingProfile{
		cache:   15 * time.Second,
		weights: ranking.DefaultWeights(),
	}
}

func applyRanking(base rankingProfile, file *rankingFile) rankingProfile {
	if file == nil {
		return base
	}
	next := base
	next.weights = base.weights.Clone()
	if file.CacheSeconds != nil {
		sec := *file.CacheSeconds
		if sec < 0 {
			sec = 0
		}
		if sec > 300 {
			sec = 300
		}
		next.cache = time.Duration(sec) * time.Second
	}
	if file.UnknownQuality != nil {
		unknown := *file.UnknownQuality
		if unknown < 0 {
			unknown = 0
		}
		next.weights.Unknown = unknown
	}
	if file.Quality != nil {
		for key, value := range file.Quality {
			key = strings.ToLower(strings.TrimSpace(key))
			if key == "" {
				continue
			}
			if value < 0 {
				value = 0
			}
			next.weights.Quality[key] = value
		}
	}
	if file.StatKeys != nil {
		seen := map[string]bool{}
		keys := make([]string, 0, len(file.StatKeys))
		for _, key := range file.StatKeys {
			key = strings.TrimSpace(key)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			keys = append(keys, key)
		}
		next.weights.StatKeys = keys
	}
	return next
}

// ItemScoreWeights is a copy of the equipped-gear formula.
func ItemScoreWeights() ranking.Weights {
	mu.RLock()
	defer mu.RUnlock()
	return current.ranking.weights.Clone()
}

// RankingCache is how long a leaderboard snapshot may be reused.
// Zero recomputes every request.
func RankingCache() time.Duration {
	mu.RLock()
	defer mu.RUnlock()
	if current.ranking.cache < 0 {
		return 0
	}
	return current.ranking.cache
}
