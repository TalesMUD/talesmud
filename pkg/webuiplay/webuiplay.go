package webuiplay

import (
	"embed"
	"io/fs"
)

//go:embed dist
var distFS embed.FS

// FS returns an fs.FS rooted at the embedded `dist/` directory.
// This serves the mud-client (game client) frontend at /play
func FS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

const IndexFile = "index.html"
