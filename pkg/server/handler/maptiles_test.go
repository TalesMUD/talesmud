package handler

import (
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestMapTilesOnlyServesSheet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assets := fstest.MapFS{"map-tiles/terrain-sheet.png": &fstest.MapFile{Data: []byte("\x89PNG\r\n\x1a\nterrain")}}
	r := gin.New()
	r.GET("/api/map-tiles/:filename", MapTiles(assets))
	r.HEAD("/api/map-tiles/:filename", MapTiles(assets))
	for _, method := range []string{"GET", "HEAD"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(method, "/api/map-tiles/terrain-sheet.png?v=hash", nil))
		if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" || w.Header().Get("Cache-Control") == "" {
			t.Fatalf("%s: %d %+v", method, w.Code, w.Header())
		}
		if method == "HEAD" && w.Body.Len() != 0 {
			t.Fatal("HEAD returned a body")
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/map-tiles/map_terrain.json", nil))
	if w.Code != 404 {
		t.Fatalf("unexpected access %d", w.Code)
	}
}
