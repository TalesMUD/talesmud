package server

import (
	"html"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/talesmud/talesmud/pkg/gamemode"
)

// serveDoor sends the text client. The index page is branded from the
// game-mode config so the first response already has the configured title.
func serveDoor(c *gin.Context, dir string) {
	name := strings.TrimPrefix(c.Param("filepath"), "/")
	if name == "" || name == "index.html" {
		writeDoorIndex(c, dir)
		return
	}
	clean := filepath.Clean(name)
	if clean == "." || strings.HasPrefix(clean, "..") {
		c.Status(http.StatusNotFound)
		return
	}
	http.ServeFile(c.Writer, c.Request, filepath.Join(dir, clean))
}

func writeDoorIndex(c *gin.Context, dir string) {
	raw, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	title, subtitle, _ := gamemode.ClientPage()
	page := brandDoorPage(string(raw), title, subtitle)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

// brandDoorPage replaces the generic heading and subtitle in the shipped page.
func brandDoorPage(page, title, subtitle string) string {
	title = html.EscapeString(title)
	subtitle = html.EscapeString(subtitle)
	page = strings.Replace(page, "<title>TalesMUD Door</title>", "<title>"+title+"</title>", 1)
	page = strings.Replace(page, `<h1 id="title">TalesMUD Door</h1>`, `<h1 id="title">`+title+`</h1>`, 1)
	page = strings.Replace(page, `<p class="sub" id="sub">A text client on TalesMUD.</p>`, `<p class="sub" id="sub">`+subtitle+`</p>`, 1)
	return page
}
