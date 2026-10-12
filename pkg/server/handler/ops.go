package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/entities/audit"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/service"
)

// OpsAction runs one live op on the game loop. The body must include confirm:true.
func OpsAction(g *game.Game, svc service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		var probe struct {
			Confirm bool `json:"confirm"`
		}
		if len(raw) == 0 || json.Unmarshal(raw, &probe) != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		if !probe.Confirm {
			c.JSON(http.StatusBadRequest, gin.H{"error": "confirm:true is required"})
			return
		}
		if g == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "game is not running"})
			return
		}
		var res *game.OpResult
		var opErr error
		if err := g.Call(func() {
			res, opErr = g.RunOp(c.Param("action"), raw)
		}); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		if opErr != nil {
			writeCoded(c, opErr)
			return
		}
		if svc == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "the operation succeeded but the audit log failed"})
			return
		}
		entry, err := svc.Record(service.AuditRecord{
			Actor:      actorFrom(c),
			Action:     c.Param("action"),
			EntityType: res.EntityType,
			EntityID:   res.EntityID,
			Before:     res.Before,
			After:      res.After,
			Source:     audit.SourceOps,
			Undoable:   res.Undoable,
			Summary:    res.Summary,
			Inverse:    res.Inverse,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "the operation succeeded but the audit log failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"summary":  res.Summary,
			"auditId":  entry.ID,
			"undoable": res.Undoable,
			"result":   res.Detail,
		})
	}
}
