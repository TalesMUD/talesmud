package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

func TestSearchEmptyQueryAndUnknownRefType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &handler.SearchHandler{}
	r.GET("/api/search", h.Search)
	r.GET("/api/refs/:type/:id", h.Refs)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=%20%20", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("empty search: %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Query string           `json:"query"`
		Hits  []worldindex.Hit `json:"hits"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Query != "" || body.Hits == nil || len(body.Hits) != 0 {
		t.Fatalf("body = %+v", body)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/refs/nope/R1", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown type: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCreatorWriteInvalidatesSearchIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/search", handler.AuditWrites(nil), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"hits": []any{}})
	})
	r.PUT("/api/rooms/:id", handler.AuditWrites(nil), func(c *gin.Context) {
		if c.GetHeader("X-Fail") == "1" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.PUT("/api/world/rooms-coords", handler.AuditWrites(nil), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	load := func() (worldindex.Snapshot, error) {
		return worldindex.NewSnapshot(), nil
	}
	if _, err := worldindex.Live.Index(load); err != nil {
		t.Fatal(err)
	}
	before := worldindex.Live.Loads()

	serve := func(method, path string, fail bool) int {
		req := httptest.NewRequest(method, path, nil)
		if fail {
			req.Header.Set("X-Fail", "1")
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec.Code
	}
	if serve(http.MethodGet, "/api/search?q=a", false) != http.StatusOK {
		t.Fatal("get")
	}
	if _, err := worldindex.Live.Index(load); err != nil {
		t.Fatal(err)
	}
	if worldindex.Live.Loads() != before {
		t.Fatalf("get reloaded: %d", worldindex.Live.Loads())
	}
	if serve(http.MethodPut, "/api/rooms/R1", true) != http.StatusBadRequest {
		t.Fatal("rejected write")
	}
	if _, err := worldindex.Live.Index(load); err != nil {
		t.Fatal(err)
	}
	if worldindex.Live.Loads() != before {
		t.Fatalf("rejected write reloaded: %d", worldindex.Live.Loads())
	}
	if serve(http.MethodPut, "/api/rooms/R1", false) != http.StatusOK {
		t.Fatal("put")
	}
	if _, err := worldindex.Live.Index(load); err != nil {
		t.Fatal(err)
	}
	if worldindex.Live.Loads() != before+1 {
		t.Fatalf("put loads %d want %d", worldindex.Live.Loads(), before+1)
	}
	if serve(http.MethodPut, "/api/world/rooms-coords", false) != http.StatusOK {
		t.Fatal("coords")
	}
	if _, err := worldindex.Live.Index(load); err != nil {
		t.Fatal(err)
	}
	if worldindex.Live.Loads() != before+2 {
		t.Fatalf("coords loads %d want %d", worldindex.Live.Loads(), before+2)
	}
}
