package server

import (
	"bytes"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SPAMiddleware serves a Single Page App (SPA) from the provided filesystem.
//
// Behavior:
// - Only handles requests matching the urlPrefix (or "/" for root)
// - If the requested file exists, serve it.
// - Otherwise, serve the SPA index file (client-side routing fallback).
//
// NOTE: We serve content directly from the embedded FS to avoid redirect loops
// from http.FileServer when used with SPA routing.
func SPAMiddleware(urlPrefix string, spaFS fs.FS, indexFile string) gin.HandlerFunc {
	return func(c *gin.Context) {
		pth := c.Request.URL.Path

		// Don't hijack API/WS routes.
		if strings.HasPrefix(pth, "/api/") || strings.HasPrefix(pth, "/admin/") || strings.HasPrefix(pth, "/ws") {
			return
		}

		// For non-root prefixes, only handle requests that match the prefix
		if urlPrefix != "/" {
			if !strings.HasPrefix(pth, urlPrefix) {
				return // Let other middleware handle this request
			}
			// Redirect /prefix to /prefix/ so relative asset paths in
			// index.html resolve correctly (e.g. "bundle.js" → /prefix/bundle.js).
			if pth == urlPrefix {
				target := urlPrefix + "/"
				// Preserve query string (critical for OAuth callback parameters like ?code=...)
				if c.Request.URL.RawQuery != "" {
					target += "?" + c.Request.URL.RawQuery
				}
				c.Redirect(http.StatusMovedPermanently, target)
				c.Abort()
				return
			}
		} else {
			// For root prefix, don't handle paths that start with other known prefixes
			// This allows /play to be handled by its own middleware
			if strings.HasPrefix(pth, "/play") {
				return
			}
		}

		// Normalize the request path to something we can stat in the FS.
		reqPath := strings.TrimPrefix(pth, urlPrefix)
		reqPath = strings.TrimPrefix(reqPath, "/")
		if reqPath == "" || reqPath == "." {
			reqPath = indexFile
		}

		// If file doesn't exist or is a directory, serve index.html.
		if info, err := fs.Stat(spaFS, path.Clean(reqPath)); err != nil || info.IsDir() {
			reqPath = indexFile
		}

		payload, err := fs.ReadFile(spaFS, path.Clean(reqPath))
		if err != nil {
			// Last-resort fallback to index.html.
			payload, err = fs.ReadFile(spaFS, indexFile)
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			reqPath = indexFile
		}

		http.ServeContent(c.Writer, c.Request, reqPath, time.Time{}, bytes.NewReader(payload))
		c.Abort()
	}
}
