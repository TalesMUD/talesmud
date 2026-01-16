package webui

import (
	"embed"
	"io/fs"
)

// distFS contains the compiled frontend assets.
//
// IMPORTANT:
// - `make build-frontend` / `make build` copies the Rollup output into `pkg/webui/dist/`.
// - We keep a minimal `dist/index.html` checked in so `go build` works even before running the frontend build.
//
//go:embed dist
var distFS embed.FS

// FS returns an fs.FS rooted at the embedded `dist/` directory.
func FS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		// This should never happen if the embed directive matches.
		panic(err)
	}
	return sub
}

// IndexFile is the SPA entrypoint within the embedded FS.
const IndexFile = "index.html"
