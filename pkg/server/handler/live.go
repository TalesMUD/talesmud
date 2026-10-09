package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
)

// LiveCharacters lists online characters and recent offline ones.
func LiveCharacters(g *game.Game) gin.HandlerFunc {
	return func(c *gin.Context) {
		if g == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "game is not running"})
			return
		}
		q := game.LiveQuery{
			OnlineOnly:   queryOn(c, "online"),
			GuestOnly:    queryOn(c, "guest"),
			InCombatOnly: queryOn(c, "inCombat"),
			Zone:         c.Query("zone"),
			All:          queryOn(c, "all"),
		}
		var rows []game.LiveCharacter
		var err error
		if callErr := g.Call(func() {
			rows, err = g.LiveCharacters(q)
		}); callErr != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": callErr.Error()})
			return
		}
		if err != nil {
			writeCoded(c, err)
			return
		}
		if rows == nil {
			rows = []game.LiveCharacter{}
		}
		c.JSON(http.StatusOK, rows)
	}
}

// LiveCharacterDetail returns inventory, quests, and the current fight.
func LiveCharacterDetail(g *game.Game) gin.HandlerFunc {
	return func(c *gin.Context) {
		if g == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "game is not running"})
			return
		}
		var row *game.LiveCharacterDetail
		var err error
		if callErr := g.Call(func() {
			row, err = g.LiveCharacter(c.Param("id"))
		}); callErr != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": callErr.Error()})
			return
		}
		if err != nil {
			writeCoded(c, err)
			return
		}
		c.JSON(http.StatusOK, row)
	}
}

// LiveNPCs lists running NPC instances.
func LiveNPCs(g *game.Game) gin.HandlerFunc {
	return func(c *gin.Context) {
		if g == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "game is not running"})
			return
		}
		var rows []game.LiveNPC
		var err error
		if callErr := g.Call(func() {
			rows, err = g.LiveNPCs(c.Query("templateId"), c.Query("roomId"))
		}); callErr != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": callErr.Error()})
			return
		}
		if err != nil {
			writeCoded(c, err)
			return
		}
		if rows == nil {
			rows = []game.LiveNPC{}
		}
		c.JSON(http.StatusOK, rows)
	}
}

// LiveInstances lists instance copies.
func LiveInstances(g *game.Game) gin.HandlerFunc {
	return func(c *gin.Context) {
		if g == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "game is not running"})
			return
		}
		var rows []game.LiveInstance
		var err error
		if callErr := g.Call(func() {
			rows, err = g.LiveInstances()
		}); callErr != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": callErr.Error()})
			return
		}
		if err != nil {
			writeCoded(c, err)
			return
		}
		if rows == nil {
			rows = []game.LiveInstance{}
		}
		c.JSON(http.StatusOK, rows)
	}
}

func queryOn(c *gin.Context, key string) bool {
	switch c.Query(key) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}
