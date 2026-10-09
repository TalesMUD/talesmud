package contenthealth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/skills"
)

var volatileFields = map[string]bool{
	"created":          true,
	"updated":          true,
	"createdBy":        true,
	"currentHitPoints": true,
	"isDead":           true,
	"deathTime":        true,
	"state":            true,
	"inCombat":         true,
	"combatInstanceId": true,
	"lastIdleDialog":   true,
	"characters":       true,
	"npcs":             true,
	"currentRoomID":    true,
	"lastRestock":      true,
}

// Baseline is the post-import snapshot used to detect Creator edits.
type Baseline struct {
	ContentCommit string          `json:"contentCommit"`
	ImportedAt    time.Time       `json:"importedAt"`
	Rules         []PackRule      `json:"rules"`
	Entries       []BaselineEntry `json:"entries"`
}

// BaselineEntry is one imported entity, stored as canonical JSON so field diffs work.
type BaselineEntry struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Hash      string          `json:"hash"`
	Canonical json.RawMessage `json:"canonical"`
}

// DriftRow is one added, removed, or changed entity.
type DriftRow struct {
	Type   string      `json:"type"`
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Kind   string      `json:"kind"`
	Fields []FieldDiff `json:"fields,omitempty"`
}

// FieldDiff is one path that differs from the import baseline.
type FieldDiff struct {
	Path   string      `json:"path"`
	Before interface{} `json:"before,omitempty"`
	After  interface{} `json:"after,omitempty"`
}

// Canonical JSON drops volatile fields and sorts map keys.
func Canonical(v any) ([]byte, error) {
	payload, err := json.Marshal(canonicalMap(v))
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func canonicalMap(v any) map[string]interface{} {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var decoded interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil
	}
	obj, ok := decoded.(map[string]interface{})
	if !ok {
		return nil
	}
	stripVolatile(obj)
	return obj
}

func stripVolatile(v interface{}) {
	switch node := v.(type) {
	case map[string]interface{}:
		for key, child := range node {
			if volatileFields[key] {
				delete(node, key)
				continue
			}
			stripVolatile(child)
		}
	case []interface{}:
		for _, child := range node {
			stripVolatile(child)
		}
	}
}

// Hash is the hex sha256 of canonical JSON.
func Hash(v any) (string, []byte, error) {
	canonical, err := Canonical(v)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), canonical, nil
}

// ContentCommit is git HEAD of dir, or empty when dir is not a checkout.
func ContentCommit(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return ""
	}
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func compareDrift(base *Baseline, current []BaselineEntry) []DriftRow {
	if base == nil {
		return nil
	}
	prev := map[string]BaselineEntry{}
	for _, entry := range base.Entries {
		prev[entry.Type+"\x00"+entry.ID] = entry
	}
	engineSkills := engineSkillIDs()
	var rows []DriftRow
	seen := map[string]bool{}
	for _, entry := range current {
		key := entry.Type + "\x00" + entry.ID
		seen[key] = true
		old, ok := prev[key]
		if !ok {
			if entry.Type == "skill" && engineSkills[entry.ID] {
				continue
			}
			rows = append(rows, DriftRow{Type: entry.Type, ID: entry.ID, Name: entry.Name, Kind: "added"})
			continue
		}
		if old.Hash == entry.Hash {
			continue
		}
		rows = append(rows, DriftRow{
			Type:   entry.Type,
			ID:     entry.ID,
			Name:   entry.Name,
			Kind:   "changed",
			Fields: diffCanonical(old.Canonical, entry.Canonical),
		})
	}
	for key, old := range prev {
		if seen[key] {
			continue
		}
		rows = append(rows, DriftRow{Type: old.Type, ID: old.ID, Name: old.Name, Kind: "removed"})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Type != rows[j].Type {
			return rows[i].Type < rows[j].Type
		}
		if rows[i].ID != rows[j].ID {
			return rows[i].ID < rows[j].ID
		}
		return rows[i].Kind < rows[j].Kind
	})
	return rows
}

func engineSkillIDs() map[string]bool {
	ids := map[string]bool{}
	for _, skill := range skills.SeedSkills() {
		if skill != nil && skill.Entity != nil && skill.ID != "" {
			ids[skill.ID] = true
		}
	}
	for _, skill := range skills.ClassKit() {
		if skill != nil && skill.Entity != nil && skill.ID != "" {
			ids[skill.ID] = true
		}
	}
	return ids
}

func diffCanonical(before, after json.RawMessage) []FieldDiff {
	var left, right interface{}
	_ = json.Unmarshal(before, &left)
	_ = json.Unmarshal(after, &right)
	var diffs []FieldDiff
	diffValues(left, right, "", &diffs)
	return diffs
}

func diffValues(before, after interface{}, path string, out *[]FieldDiff) {
	if len(*out) >= 40 {
		return
	}
	left, lok := before.(map[string]interface{})
	right, rok := after.(map[string]interface{})
	if lok && rok {
		keys := map[string]bool{}
		for key := range left {
			keys[key] = true
		}
		for key := range right {
			keys[key] = true
		}
		ordered := make([]string, 0, len(keys))
		for key := range keys {
			ordered = append(ordered, key)
		}
		sort.Strings(ordered)
		for _, key := range ordered {
			next := key
			if path != "" {
				next = path + "." + key
			}
			lv, hasL := left[key]
			rv, hasR := right[key]
			if !hasL {
				*out = append(*out, FieldDiff{Path: next, After: rv})
				continue
			}
			if !hasR {
				*out = append(*out, FieldDiff{Path: next, Before: lv})
				continue
			}
			diffValues(lv, rv, next, out)
		}
		return
	}
	if !jsonEqual(before, after) {
		*out = append(*out, FieldDiff{Path: path, Before: before, After: after})
	}
}

func jsonEqual(a, b interface{}) bool {
	ab, aerr := json.Marshal(a)
	bb, berr := json.Marshal(b)
	if aerr != nil || berr != nil {
		return false
	}
	return string(ab) == string(bb)
}
