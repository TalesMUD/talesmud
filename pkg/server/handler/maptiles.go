package handler

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MapTiles serves the versioned atlas sheet from the embedded play assets.
func MapTiles(assets fs.FS) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Param("filename") != "terrain-sheet.png" {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("Cache-Control", "public, max-age=86400")
		c.FileFromFS("map-tiles/terrain-sheet.png", http.FS(assets))
	}
}
