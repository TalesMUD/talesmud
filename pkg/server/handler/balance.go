package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

type scalingFactors struct {
	HP      float64 `json:"hp"`
	Attack  float64 `json:"attack"`
	Defense float64 `json:"defense"`
}

// GetEnemyScaling returns the difficulty tiers and named overrides combat uses.
// The tables come from config. Nothing about a particular world pack is hardcoded.
func GetEnemyScaling(c *gin.Context) {
	cfg := balance.GetConfig()
	tiers := map[string]scalingFactors{}
	named := map[string]scalingFactors{}
	if cfg != nil {
		for name, m := range cfg.DifficultyMultipliers {
			tiers[name] = scalingFactors{HP: m.HP, Attack: m.Attack, Defense: m.Defense}
		}
		for name, m := range cfg.NamedOverrides {
			named[name] = scalingFactors{HP: m.HP, Attack: m.Attack, Defense: m.Defense}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"tiers":          tiers,
		"namedOverrides": named,
	})
}
