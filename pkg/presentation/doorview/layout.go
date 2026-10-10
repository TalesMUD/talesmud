package doorview

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"

	"github.com/talesmud/talesmud/pkg/gamemode"
)

// layoutFile is screens/layout.yaml. A missing file keeps the classic frame.
type layoutFile struct {
	Style   string `yaml:"style"`
	ArtRows int    `yaml:"art_rows"`
}

type loadedLayout struct {
	path string
	mod  time.Time
	file layoutFile
}

var (
	layoutMu    sync.Mutex
	layoutCache loadedLayout
)

func loadLayout() layoutFile {
	root := strings.TrimSpace(gamemode.Current().WorldPack)
	if root == "" {
		return layoutFile{}
	}
	path := filepath.Join(root, "screens", "layout.yaml")
	st, err := os.Stat(path)
	if err != nil {
		return layoutFile{}
	}
	layoutMu.Lock()
	defer layoutMu.Unlock()
	if layoutCache.path == path && layoutCache.mod.Equal(st.ModTime()) {
		return layoutCache.file
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return layoutFile{}
	}
	var file layoutFile
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return layoutFile{}
	}
	if file.Style != "rails" {
		file.Style = ""
	}
	if file.ArtRows < 0 {
		file.ArtRows = 0
	}
	if file.ArtRows > 14 {
		file.ArtRows = 14
	}
	layoutCache = loadedLayout{path: path, mod: st.ModTime(), file: file}
	return file
}

func railsOn() bool {
	return loadLayout().Style == "rails"
}

func artRowLimit() int {
	n := loadLayout().ArtRows
	if n <= 0 {
		return 8
	}
	return n
}

func visibleWidth(s string) int {
	return utf8.RuneCountInString(stripANSI(s))
}

// wrapEntries keeps each hotkey entry whole. The gap is two spaces.
func wrapEntries(parts []string, width int) []string {
	if width < 8 {
		width = 8
	}
	var out []string
	line := ""
	vis := 0
	for _, part := range parts {
		w := visibleWidth(part)
		if line != "" && vis+2+w > width {
			out = append(out, line)
			line = part
			vis = w
			continue
		}
		if line != "" {
			line += "  "
			vis += 2
		}
		line += part
		vis += w
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}

func hasInterior(art string) bool {
	if art == "" {
		return false
	}
	for _, line := range strings.Split(art, "\n") {
		if isInterior(stripANSI(line)) {
			return true
		}
	}
	return false
}

func isInterior(plain string) bool {
	r := []rune(plain)
	if len(r) < 3 {
		return false
	}
	if (r[0] != '║' && r[0] != '│') || (r[len(r)-1] != '║' && r[len(r)-1] != '│') {
		return false
	}
	for _, c := range r[1 : len(r)-1] {
		if c != ' ' {
			return false
		}
	}
	return true
}

// insetCatalog writes catalog lines into blank bordered rows and leaves any
// row that already holds a picture.
func insetCatalog(art string, catalog []string) []string {
	lines := strings.Split(art, "\n")
	n := 0
	for i, line := range lines {
		plain := stripANSI(line)
		if !isInterior(plain) {
			continue
		}
		if n >= len(catalog) {
			break
		}
		lines[i] = fillInterior(plain, catalog[n])
		n++
	}
	if n < len(catalog) {
		lines = append(lines, catalog[n:]...)
	}
	return lines
}

func fillInterior(plain, text string) string {
	runes := []rune(plain)
	if len(runes) < 2 {
		return plain
	}
	width := len(runes) - 2
	body := text
	if utf8.RuneCountInString(body) > width {
		body = string([]rune(body)[:width])
	}
	inner := body + strings.Repeat(" ", width-utf8.RuneCountInString(body))
	return "\x1b[1;34m" + string(runes[0]) + "\x1b[0m" + inner + "\x1b[0;34m" + string(runes[len(runes)-1]) + "\x1b[0m"
}
