package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/contenthealth"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

// SearchHandler serves creator search and reference lookups.
type SearchHandler struct {
	Facade service.Facade
}

type searchBody struct {
	Query string           `json:"query"`
	Hits  []worldindex.Hit `json:"hits"`
}

var errSearchUnavailable = errors.New("search is unavailable")

// Search matches id, name, and text across the cached content index.
func (h *SearchHandler) Search(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	query, ok := worldindex.ParseQuery(c.Query("q"), c.QueryArray("types"), limit)
	if !ok {
		c.JSON(http.StatusOK, searchBody{
			Query: strings.TrimSpace(c.Query("q")),
			Hits:  []worldindex.Hit{},
		})
		return
	}
	ix, err := h.index()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, searchBody{Query: query.Text, Hits: ix.Search(query)})
}

// Refs lists who points at an entity, and what that entity points at.
func (h *SearchHandler) Refs(c *gin.Context) {
	kind, ok := worldindex.ParseKind(c.Param("type"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown type"})
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	ix, err := h.index()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	view, found := ix.References(kind, id)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *SearchHandler) index() (*worldindex.Index, error) {
	if h == nil || h.Facade == nil {
		return nil, errSearchUnavailable
	}
	return worldindex.Live.Index(func() (worldindex.Snapshot, error) {
		world, err := contenthealth.WorldFromFacade(h.Facade)
		if err != nil {
			return worldindex.Snapshot{}, err
		}
		return world.Snapshot(), nil
	})
}
