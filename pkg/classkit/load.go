package classkit

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func init() {
	UseDefaults()
	if testBinary() {
		if dir := findUp("pkg/classkit/testdata/classes"); dir != "" {
			if _, err := LoadDir(dir); err != nil {
				panic(err)
			}
		}
	}
}

func testBinary() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

// LoadDir replaces the catalog with every *.yaml file in dir.
// A missing directory leaves the catalog unchanged and returns 0.
func LoadDir(dir string) (int, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return 0, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	if len(files) == 0 {
		return 0, nil
	}
	sort.Strings(files)
	list := make([]*Def, 0, len(files))
	seen := map[string]string{}
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", path, err)
		}
		var d Def
		if err := yaml.Unmarshal(raw, &d); err != nil {
			return 0, fmt.Errorf("%s: %w", path, err)
		}
		if err := validate(&d, path); err != nil {
			return 0, err
		}
		if prev, ok := seen[norm(d.ID)]; ok {
			return 0, fmt.Errorf("%s: duplicate class id %s (also %s)", path, d.ID, prev)
		}
		seen[norm(d.ID)] = path
		list = append(list, &d)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Order != list[j].Order {
			return list[i].Order < list[j].Order
		}
		return list[i].ID < list[j].ID
	})
	install(list, true)
	return len(list), nil
}

func validate(d *Def, path string) error {
	if norm(d.ID) == "" || strings.TrimSpace(d.Name) == "" {
		return fmt.Errorf("%s: class needs id and name", path)
	}
	if d.HPMultiplier < 0 {
		return fmt.Errorf("%s: hp_multiplier", path)
	}
	for _, s := range d.Skills {
		if s.ID == "" || s.Name == "" {
			return fmt.Errorf("%s: skill needs id and name", path)
		}
		if !KnownEffects[norm(s.Effect)] {
			return fmt.Errorf("%s: skill %s has unknown effect %q", path, s.ID, s.Effect)
		}
	}
	return nil
}

// LoadWorld loads the first world-pack class directory that has files.
// It returns the count and the directory. A miss leaves the generic sample.
func LoadWorld() (int, string) {
	for _, dir := range worldCandidates() {
		n, err := LoadDir(dir)
		if err != nil {
			log.WithError(err).WithField("dir", dir).Error("class kit failed to load")
			continue
		}
		if n > 0 {
			log.WithField("dir", dir).WithField("classes", n).Info("class kit loaded from world pack")
			return n, dir
		}
	}
	log.Info("class kit using generic sample classes")
	return 0, ""
}

func worldCandidates() []string {
	var out []string
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		for _, existing := range out {
			if existing == p {
				return
			}
		}
		out = append(out, p)
	}
	add(os.Getenv("CLASSKIT_DIR"))
	add(filepath.Join("import", "mvp-rpg-1", "data", "classes"))
	add(filepath.Join("..", "talesmud-rpg-1", "data", "classes"))
	if dir := findUp(filepath.Join("..", "talesmud-rpg-1", "data", "classes")); dir != "" {
		add(dir)
	}
	if dir := findUp(filepath.Join("import", "mvp-rpg-1", "data", "classes")); dir != "" {
		add(dir)
	}
	if dir := findUp(filepath.Join("data", "classes")); dir != "" && !strings.Contains(dir, "testdata") {
		add(dir)
	}
	return out
}

func findUp(rel string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := cwd
	for i := 0; i < 8; i++ {
		cand := filepath.Join(dir, rel)
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
