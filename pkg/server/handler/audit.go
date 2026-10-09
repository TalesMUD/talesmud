package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/audit"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/service"
)

type auditRoute struct {
	entityType string
	action     string
}

// auditRoutes are the creator CRUD writes. Reads, previews, and validates stay out.
var auditRoutes = map[string]auditRoute{
	"POST /api/rooms":                     {entityType: "rooms", action: "create"},
	"PUT /api/rooms/:id":                  {entityType: "rooms", action: "update"},
	"DELETE /api/rooms/:id":               {entityType: "rooms", action: "delete"},
	"POST /api/items":                     {entityType: "items", action: "create"},
	"PUT /api/items/:id":                  {entityType: "items", action: "update"},
	"DELETE /api/items/:id":               {entityType: "items", action: "delete"},
	"POST /api/npcs":                      {entityType: "npcs", action: "create"},
	"PUT /api/npcs/:id":                   {entityType: "npcs", action: "update"},
	"DELETE /api/npcs/:id":                {entityType: "npcs", action: "delete"},
	"POST /api/loottables":                {entityType: "loottables", action: "create"},
	"PUT /api/loottables/:id":             {entityType: "loottables", action: "update"},
	"DELETE /api/loottables/:id":          {entityType: "loottables", action: "delete"},
	"POST /api/spawners":                  {entityType: "spawners", action: "create"},
	"PUT /api/spawners/:id":               {entityType: "spawners", action: "update"},
	"DELETE /api/spawners/:id":            {entityType: "spawners", action: "delete"},
	"POST /api/dialogs":                   {entityType: "dialogs", action: "create"},
	"PUT /api/dialogs/:id":                {entityType: "dialogs", action: "update"},
	"DELETE /api/dialogs/:id":             {entityType: "dialogs", action: "delete"},
	"POST /api/quests":                    {entityType: "quests", action: "create"},
	"PUT /api/quests/:id":                 {entityType: "quests", action: "update"},
	"DELETE /api/quests/:id":              {entityType: "quests", action: "delete"},
	"POST /api/scripts":                   {entityType: "scripts", action: "create"},
	"PUT /api/scripts/:id":                {entityType: "scripts", action: "update"},
	"DELETE /api/scripts/:id":             {entityType: "scripts", action: "delete"},
	"POST /api/skills":                    {entityType: "skills", action: "create"},
	"PUT /api/skills/:id":                 {entityType: "skills", action: "update"},
	"DELETE /api/skills/:id":              {entityType: "skills", action: "delete"},
	"POST /api/character-templates":       {entityType: "character-templates", action: "create"},
	"PUT /api/character-templates/:id":    {entityType: "character-templates", action: "update"},
	"DELETE /api/character-templates/:id": {entityType: "character-templates", action: "delete"},
	"PUT /api/settings":                   {entityType: "settings", action: "update"},
}

// AuditWrites records a successful creator CRUD write. A log failure does not change the response.
func AuditWrites(svc service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		spec, ok := auditRoutes[c.Request.Method+" "+c.FullPath()]
		if !ok || svc == nil {
			c.Next()
			return
		}
		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(reqBody))
		}
		entityID := c.Param("id")
		if spec.entityType == "settings" {
			entityID = "server-settings"
		}
		var before json.RawMessage
		if c.Request.Method != http.MethodPost {
			raw, err := svc.LoadRaw(spec.entityType, entityID)
			if err != nil {
				log.WithError(err).WithField("entityType", spec.entityType).Error("audit snapshot failed")
				c.Next()
				return
			}
			before = raw
		}
		cap := &bodyCapture{ResponseWriter: c.Writer}
		c.Writer = cap
		c.Next()
		if c.Writer.Status() >= 400 {
			return
		}
		if c.Request.Method == http.MethodPost {
			entityID = idFromJSON(cap.body.Bytes())
		}
		if entityID == "" {
			log.WithField("path", c.FullPath()).Error("audit skipped: no entity id")
			return
		}
		var after json.RawMessage
		if c.Request.Method != http.MethodDelete {
			raw, err := svc.LoadRaw(spec.entityType, entityID)
			if err != nil {
				log.WithError(err).Error("audit after-snapshot failed")
				return
			}
			after = raw
		}
		if _, err := svc.Record(service.AuditRecord{
			Actor:      actorFrom(c),
			Action:     spec.action,
			EntityType: spec.entityType,
			EntityID:   entityID,
			Before:     before,
			After:      after,
			Source:     audit.SourceCreator,
			Undoable:   true,
			Summary:    spec.action + " " + spec.entityType + " " + entityID,
		}); err != nil {
			log.WithError(err).Error("audit record failed")
		}
		_ = reqBody
	}
}

// ListAudit returns recent audit rows for creators.
func ListAudit(svc service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audit log is not available"})
			return
		}
		limit, _ := strconv.Atoi(c.Query("limit"))
		rows, err := svc.List(c.Query("entityType"), c.Query("entityId"), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if rows == nil {
			rows = []*audit.Entry{}
		}
		c.JSON(http.StatusOK, rows)
	}
}

// UndoAudit restores a creator row or runs an ops inverse. Admin only.
func UndoAudit(g *game.Game, svc service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audit log is not available"})
			return
		}
		entry, err := svc.Get(c.Param("id"))
		if err != nil || entry == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "audit entry not found"})
			return
		}
		actor := actorFrom(c)
		if entry.Source == audit.SourceOps {
			undoOps(c, g, svc, entry, actor)
			return
		}
		undo, err := svc.UndoCRUD(c.Param("id"), actor)
		if err != nil {
			writeCoded(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"summary":  undo.Summary,
			"auditId":  undo.ID,
			"undoable": undo.Undoable,
		})
	}
}

func undoOps(c *gin.Context, g *game.Game, svc service.AuditService, entry *audit.Entry, actor service.AuditActor) {
	if entry.UndoneBy != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this change was already undone"})
		return
	}
	if !entry.Undoable || entry.Inverse == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this change cannot be undone"})
		return
	}
	if g == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "game is not running"})
		return
	}
	var res *game.OpResult
	var opErr error
	if err := g.Call(func() {
		res, opErr = g.ApplyInverse(entry.Inverse)
	}); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	if opErr != nil {
		writeCoded(c, opErr)
		return
	}
	undo, err := svc.Record(service.AuditRecord{
		Actor:      actor,
		Action:     "undo",
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		Before:     res.Before,
		After:      res.After,
		Source:     audit.SourceOps,
		Undoable:   res.Undoable,
		Summary:    res.Summary,
		Inverse:    res.Inverse,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "the undo succeeded but the audit log failed"})
		return
	}
	if err := svc.MarkUndone(entry.ID, undo.ID); err != nil {
		log.WithError(err).Error("audit mark undone failed")
	}
	c.JSON(http.StatusOK, gin.H{
		"summary":  res.Summary,
		"auditId":  undo.ID,
		"undoable": res.Undoable,
		"result":   res.Detail,
	})
}

func actorFrom(c *gin.Context) service.AuditActor {
	usr, ok := c.Get("user")
	if !ok {
		return service.AuditActor{Name: "unknown"}
	}
	user, ok := usr.(*entities.User)
	if !ok || user == nil {
		return service.AuditActor{Name: "unknown"}
	}
	name := user.Nickname
	if name == "" {
		name = user.Name
	}
	if name == "" {
		name = user.Username
	}
	if name == "" {
		name = user.ID
	}
	return service.AuditActor{UserID: user.ID, Name: name}
}

func idFromJSON(body []byte) string {
	if len(bytes.TrimSpace(body)) == 0 {
		return ""
	}
	var probe map[string]interface{}
	if err := json.Unmarshal(body, &probe); err != nil {
		return ""
	}
	if id, ok := probe["id"].(string); ok {
		return id
	}
	return ""
}

type bodyCapture struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *bodyCapture) Write(b []byte) (int, error) {
	_, _ = w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type codedError interface {
	error
	StatusCode() int
}

func writeCoded(c *gin.Context, err error) {
	if err == nil {
		return
	}
	status := http.StatusInternalServerError
	var coded codedError
	if ok := asCoded(err, &coded); ok && coded.StatusCode() != 0 {
		status = coded.StatusCode()
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func asCoded(err error, dest *codedError) bool {
	switch e := err.(type) {
	case codedError:
		*dest = e
		return true
	default:
		return false
	}
}
