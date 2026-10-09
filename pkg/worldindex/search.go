package worldindex

import (
	"net/url"
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
)

// Rank bands. A lower number is a stronger match.
const (
	RankExactID = iota
	RankIDPrefix
	RankIDContains
	RankName
	RankText
)

// Hit is one search result.
type Hit struct {
	Type    Kind   `json:"type"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Snippet string `json:"snippet"`
	Path    string `json:"path"`
	Rank    int    `json:"rank"`
}

// Query is a creator search. Types nil searches every indexed content kind except classes.
type Query struct {
	Text  string
	Types map[Kind]bool
	Limit int
}

// ParseKind accepts the graph kind and the hyphenated URL form.
func ParseKind(raw string) (Kind, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "room":
		return KindRoom, true
	case "npc":
		return KindNPC, true
	case "item":
		return KindItem, true
	case "loottable", "loot-table", "loottables":
		return KindLootTable, true
	case "spawner":
		return KindSpawner, true
	case "dialog":
		return KindDialog, true
	case "quest":
		return KindQuest, true
	case "script":
		return KindScript, true
	case "skill":
		return KindSkill, true
	case "charactertemplate", "character-template", "character_template":
		return KindCharacterTemplate, true
	case "class":
		return KindClass, true
	default:
		return "", false
	}
}

// ParseQuery normalizes a search. An empty type list searches the content kinds.
// Limit defaults to 25 and caps at 100.
func ParseQuery(text string, types []string, limit int) (Query, bool) {
	q := Query{Text: strings.TrimSpace(text), Limit: limit}
	if q.Text == "" {
		return q, false
	}
	if q.Limit <= 0 {
		q.Limit = 25
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if len(types) == 0 {
		return q, true
	}
	q.Types = map[Kind]bool{}
	for _, raw := range types {
		for _, part := range strings.Split(raw, ",") {
			kind, ok := ParseKind(part)
			if ok && kind != KindClass {
				q.Types[kind] = true
			}
		}
	}
	return q, len(q.Types) > 0
}

// Search returns ranked hits. Exact id beats an id prefix, which beats an id
// fragment, which beats a name, which beats other text.
func (ix *Index) Search(q Query) []Hit {
	if ix == nil || strings.TrimSpace(q.Text) == "" {
		return []Hit{}
	}
	if q.Limit <= 0 {
		q.Limit = 25
	}
	needle := strings.ToLower(strings.TrimSpace(q.Text))
	var hits []Hit
	for id, room := range ix.snap.Rooms {
		if room == nil {
			continue
		}
		ix.addHit(&hits, q, KindRoom, id, room.Name, []string{room.Description, room.Detail}, needle)
	}
	for id, n := range ix.snap.NPCs {
		if n == nil {
			continue
		}
		ix.addHit(&hits, q, KindNPC, id, n.Name, []string{n.Description}, needle)
	}
	for id, item := range ix.snap.Items {
		if item == nil {
			continue
		}
		ix.addHit(&hits, q, KindItem, id, item.Name, []string{item.Description, item.Detail}, needle)
	}
	for id, table := range ix.snap.LootTables {
		if table == nil {
			continue
		}
		ix.addHit(&hits, q, KindLootTable, id, table.Name, []string{table.Description}, needle)
	}
	for id, spawner := range ix.snap.Spawners {
		if spawner == nil {
			continue
		}
		ix.addHit(&hits, q, KindSpawner, id, spawner.Name, nil, needle)
	}
	for id, dialog := range ix.snap.Dialogs {
		if dialog == nil {
			continue
		}
		ix.addHit(&hits, q, KindDialog, id, dialog.Name, dialogTexts(dialog), needle)
	}
	for id, quest := range ix.snap.Quests {
		if quest == nil {
			continue
		}
		ix.addHit(&hits, q, KindQuest, id, quest.Name, questTexts(quest), needle)
	}
	for id, script := range ix.snap.Scripts {
		if script == nil {
			continue
		}
		ix.addHit(&hits, q, KindScript, id, script.Name, []string{script.Description, script.Code}, needle)
	}
	for id, skill := range ix.snap.Skills {
		if skill == nil {
			continue
		}
		ix.addHit(&hits, q, KindSkill, id, skill.Name, []string{skill.Description}, needle)
	}
	for id, tmpl := range ix.snap.CharacterTemplates {
		if tmpl == nil {
			continue
		}
		ix.addHit(&hits, q, KindCharacterTemplate, id, tmpl.Name, []string{tmpl.Description}, needle)
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Rank != hits[j].Rank {
			return hits[i].Rank < hits[j].Rank
		}
		if hits[i].Name != hits[j].Name {
			return hits[i].Name < hits[j].Name
		}
		if hits[i].Type != hits[j].Type {
			return hits[i].Type < hits[j].Type
		}
		return hits[i].ID < hits[j].ID
	})
	if len(hits) > q.Limit {
		hits = hits[:q.Limit]
	}
	return hits
}

func (ix *Index) addHit(hits *[]Hit, q Query, kind Kind, id, name string, texts []string, needle string) {
	if q.Types != nil && !q.Types[kind] {
		return
	}
	if name == "" {
		name = id
	}
	lowerID := strings.ToLower(id)
	lowerName := strings.ToLower(name)
	switch {
	case lowerID == needle:
		*hits = append(*hits, ix.hit(kind, id, name, id, RankExactID))
	case strings.HasPrefix(lowerID, needle):
		*hits = append(*hits, ix.hit(kind, id, name, id, RankIDPrefix))
	case strings.Contains(lowerID, needle):
		*hits = append(*hits, ix.hit(kind, id, name, id, RankIDContains))
	case strings.Contains(lowerName, needle):
		*hits = append(*hits, ix.hit(kind, id, name, name, RankName))
	default:
		for _, text := range texts {
			if text == "" || !strings.Contains(strings.ToLower(text), needle) {
				continue
			}
			*hits = append(*hits, ix.hit(kind, id, name, snippetAround(text, needle), RankText))
			return
		}
	}
}

func (ix *Index) hit(kind Kind, id, name, snippet string, rank int) Hit {
	return Hit{
		Type:    kind,
		ID:      id,
		Name:    name,
		Snippet: snippet,
		Path:    ix.CreatorPath(kind, id),
		Rank:    rank,
	}
}

// CreatorPath is the Creator deep link for an entity.
func (ix *Index) CreatorPath(kind Kind, id string) string {
	if id == "" {
		return ""
	}
	switch kind {
	case KindSpawner:
		return "/creator/spawners?id=" + url.QueryEscape(id)
	case KindLootTable:
		return "/creator/loot-tables?id=" + url.QueryEscape(id)
	}
	tab := creatorTab(kind)
	if tab == "" {
		return ""
	}
	return "/creator/" + tab + "?id=" + url.QueryEscape(id)
}

func creatorTab(kind Kind) string {
	switch kind {
	case KindRoom:
		return "rooms"
	case KindNPC:
		return "npcs"
	case KindItem:
		return "item-templates"
	case KindDialog:
		return "dialogs"
	case KindQuest:
		return "quests"
	case KindScript:
		return "scripts"
	case KindSkill:
		return "skills"
	case KindCharacterTemplate:
		return "character-templates"
	default:
		return ""
	}
}

func dialogTexts(node *dialogs.Dialog) []string {
	var out []string
	var walk func(*dialogs.Dialog)
	walk = func(n *dialogs.Dialog) {
		if n == nil {
			return
		}
		if n.Text != "" {
			out = append(out, n.Text)
		}
		for _, alt := range n.AlternateTexts {
			if alt != "" {
				out = append(out, alt)
			}
		}
		for _, option := range n.Options {
			walk(option)
		}
		walk(n.Answer)
	}
	walk(node)
	return out
}

func questTexts(quest *quests.Quest) []string {
	if quest == nil {
		return nil
	}
	out := []string{}
	if quest.Description != "" {
		out = append(out, quest.Description)
	}
	for _, objective := range quest.Objectives {
		if objective.Description != "" {
			out = append(out, objective.Description)
		}
	}
	return out
}

func snippetAround(text, needle string) string {
	runes := []rune(text)
	lower := []rune(strings.ToLower(text))
	q := []rune(needle)
	idx := runeIndex(lower, q)
	if idx < 0 {
		return clipRunes(runes, 0, 80, false, len(runes) > 80)
	}
	start := idx - 24
	if start < 0 {
		start = 0
	}
	end := idx + len(q) + 40
	if end > len(runes) {
		end = len(runes)
	}
	return clipRunes(runes, start, end, start > 0, end < len(runes))
}

func clipRunes(runes []rune, start, end int, lead, trail bool) string {
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	if start > end {
		return ""
	}
	out := string(runes[start:end])
	if lead {
		out = "…" + out
	}
	if trail {
		out += "…"
	}
	return out
}

func runeIndex(text, needle []rune) int {
	if len(needle) == 0 || len(needle) > len(text) {
		return -1
	}
	for i := 0; i+len(needle) <= len(text); i++ {
		if runesEqual(text[i:i+len(needle)], needle) {
			return i
		}
	}
	return -1
}

func runesEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
