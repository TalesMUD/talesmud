package doorview

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/gamemode"
)

// keyBind is one pack-defined key. Command is an engine command.
// Prompt switches the client to a line whose text is appended to Command.
// Notice is shown instead of dispatching when Command and Prompt are empty.
type keyBind struct {
	Key     string
	Command string
	Label   string
	Prompt  string
	Notice  string
}

type keySpec struct {
	Command string `yaml:"command"`
	Label   string `yaml:"label"`
	Prompt  string `yaml:"prompt"`
	Notice  string `yaml:"notice"`
}

type keyFile struct {
	Combat map[string]keySpec            `yaml:"combat"`
	Areas  map[string]map[string]keySpec `yaml:"areas"`
	Rooms  map[string]map[string]keySpec `yaml:"rooms"`
}

type loadedKeys struct {
	path string
	mod  time.Time
	file keyFile
}

var (
	keyMu    sync.Mutex
	keyCache loadedKeys
)

func loadKeyFile() keyFile {
	root := strings.TrimSpace(gamemode.Current().WorldPack)
	if root == "" {
		return keyFile{}
	}
	path := filepath.Join(root, "keymap.yaml")
	st, err := os.Stat(path)
	if err != nil {
		return keyFile{}
	}
	keyMu.Lock()
	defer keyMu.Unlock()
	if keyCache.path == path && keyCache.mod.Equal(st.ModTime()) {
		return keyCache.file
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return keyFile{}
	}
	var file keyFile
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return keyFile{}
	}
	keyCache = loadedKeys{path: path, mod: st.ModTime(), file: file}
	return file
}

func bindFrom(key string, spec keySpec) (keyBind, bool) {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" || key == "d" {
		return keyBind{}, false
	}
	label := strings.TrimSpace(spec.Label)
	if label == "" {
		label = strings.TrimSpace(spec.Command)
	}
	if label == "" {
		label = key
	}
	return keyBind{
		Key:     key,
		Command: strings.TrimSpace(spec.Command),
		Label:   label,
		Prompt:  strings.TrimSpace(spec.Prompt),
		Notice:  strings.TrimSpace(spec.Notice),
	}, true
}

func putBind(dst map[string]keyBind, key string, spec keySpec) {
	b, ok := bindFrom(key, spec)
	if !ok {
		return
	}
	dst[b.Key] = b
}

// keysFor returns the bindings that apply in this room. Combat bindings win
// while a fight is active. Room bindings win over area bindings. d is never
// taken from the pack; it stays the down direction.
func keysFor(room *rooms.Room, inCombat bool) map[string]keyBind {
	file := loadKeyFile()
	merged := map[string]keyBind{}
	if room != nil {
		if area := file.Areas[room.Area]; area != nil {
			for key, spec := range area {
				putBind(merged, key, spec)
			}
		}
		if local := file.Rooms[room.ID]; local != nil {
			for key, spec := range local {
				putBind(merged, key, spec)
			}
		}
	}
	if inCombat {
		for key, spec := range file.Combat {
			putBind(merged, key, spec)
		}
	}
	return merged
}
