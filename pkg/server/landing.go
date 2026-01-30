package server

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// LandingMiddleware serves a static landing page from the OS filesystem.
//
// If landingPath is empty or does not contain an index.html, it returns a
// no-op middleware. Otherwise, requests to "/" are served the landing page
// unless Auth0 callback query params (?code= or ?error=) are present, in
// which case the request falls through to the main SPA.
func LandingMiddleware(landingPath string) gin.HandlerFunc {
	if landingPath == "" {
		return func(c *gin.Context) {}
	}

	indexPath := filepath.Join(landingPath, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		log.WithField("path", landingPath).Info("Landing page not found, disabled")
		return func(c *gin.Context) {}
	}

	log.WithField("path", landingPath).Info("Landing page enabled")

	return func(c *gin.Context) {
		pth := c.Request.URL.Path

		// Don't intercept API, admin, WebSocket, or /play routes.
		if strings.HasPrefix(pth, "/api/") ||
			strings.HasPrefix(pth, "/admin/") ||
			strings.HasPrefix(pth, "/ws") ||
			strings.HasPrefix(pth, "/play") {
			return
		}

		// Auth0 callback passthrough: let the main SPA handle ?code= / ?error=.
		q := c.Request.URL.Query()
		if q.Get("code") != "" || q.Get("error") != "" {
			return
		}

		// Serve landing index.html for the root path.
		if pth == "/" {
			c.File(indexPath)
			c.Abort()
			return
		}

		// Serve static assets from the landing directory (e.g. images, CSS).
		clean := filepath.Clean(pth[1:]) // strip leading "/"
		assetPath := filepath.Join(landingPath, clean)

		// Prevent directory traversal.
		if !strings.HasPrefix(assetPath, filepath.Clean(landingPath)+string(os.PathSeparator)) &&
			assetPath != filepath.Clean(landingPath) {
			return
		}

		info, err := os.Stat(assetPath)
		if err != nil || info.IsDir() {
			return // file doesn't exist in landing dir, fall through
		}

		c.File(assetPath)
		c.Abort()
	}
}
