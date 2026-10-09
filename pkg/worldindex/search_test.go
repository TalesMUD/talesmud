package worldindex

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/scripts"
)

func searchWorld() *Index {
	cellar := room("R0215", rooms.Exit{Name: "down", Target: "R0230"})
	cellar.Name = "Cellar"
	cellar.Description = "A brass gear sits on the barrel."
	snap := NewSnapshot()
	snap.Rooms["R0215"] = cellar
	snap.Items["ITM0017"] = &items.Item{
		Entity:      &entities.Entity{ID: "ITM0017"},
		Name:        "Strange Gear",
		Description: "A brass gear approximately the size of your palm.",
		IsTemplate:  true,
	}
	snap.Quests["QST0217"] = &quests.Quest{
		Entity:      &entities.Entity{ID: "QST0217"},
		Name:        "The Cooked Book",
		Description: "Return the ledger.",
		Objectives: []quests.Objective{{
			ID:          "drop",
			Type:        quests.ObjectiveDeliver,
			Description: "Deliver the gear to the cook",
			TargetID:    "ITM0017",
		}},
	}
	snap.NPCs["NPC1"] = &npc.NPC{
		Entity: &entities.Entity{ID: "NPC1"},
		Name:   "Cook",
		EnemyTrait: &npc.EnemyTrait{
			LootTableID: "LT1",
		},
	}
	snap.LootTables["LT1"] = &items.LootTable{
		Entity: &entities.Entity{ID: "LT1"},
		Name:   "Cook pots",
	}
	snap.Scripts["SCR0204"] = &scripts.Script{
		Entity: &entities.Entity{ID: "SCR0204"},
		Name:   "Hatch",
		Code: `
local roomID = "R0215"
tales.game.revealExit(roomID, "down", charID)
`,
	}
	snap.Dialogs["DLG1"] = &dialogs.Dialog{
		Entity: &entities.Entity{ID: "DLG1"},
		Name:   "Cook talk",
		Text:   "The stew needs salt.",
		Options: []*dialogs.Dialog{{
			Text: "Ask about the strange gear",
		}},
	}
	snap.Spawners["SP1"] = &npc.NPCSpawner{
		Entity:     &entities.Entity{ID: "SP1"},
		Name:       "Cellar rats",
		RoomID:     "R0215",
		TemplateID: "NPC1",
	}
	return Build(snap)
}

func TestSearchRanksIDAboveNameAboveText(t *testing.T) {
	ix := searchWorld()

	exact, ok := ParseQuery("QST0217", nil, 10)
	if !ok {
		t.Fatal("query")
	}
	hits := ix.Search(exact)
	if len(hits) != 1 || hits[0].ID != "QST0217" || hits[0].Rank != RankExactID {
		t.Fatalf("exact = %+v", hits)
	}
	if hits[0].Path != "/creator/quests?id=QST0217" {
		t.Fatalf("path %q", hits[0].Path)
	}

	prefix := ix.Search(mustQuery(t, "QST", nil, 10))
	if len(prefix) == 0 || prefix[0].ID != "QST0217" || prefix[0].Rank != RankIDPrefix {
		t.Fatalf("prefix = %+v", prefix)
	}

	fragment := ix.Search(mustQuery(t, "0217", nil, 10))
	if len(fragment) == 0 || fragment[0].ID != "QST0217" || fragment[0].Rank != RankIDContains {
		t.Fatalf("fragment = %+v", fragment)
	}

	named := ix.Search(mustQuery(t, "strange ge", nil, 10))
	if len(named) < 2 {
		t.Fatalf("name hits = %+v", named)
	}
	if named[0].ID != "ITM0017" || named[0].Rank != RankName {
		t.Fatalf("name winner = %+v", named[0])
	}
	if named[0].Path != "/creator/item-templates?id=ITM0017" {
		t.Fatalf("item path %q", named[0].Path)
	}
	var textHit *Hit
	for i := range named {
		if named[i].Rank == RankText {
			textHit = &named[i]
			break
		}
	}
	if textHit == nil || !strings.Contains(strings.ToLower(textHit.Snippet), "strange gear") {
		t.Fatalf("text hit = %+v", named)
	}

	onlyItems := ix.Search(mustQuery(t, "gear", []string{"item"}, 10))
	if len(onlyItems) != 1 || onlyItems[0].Type != KindItem {
		t.Fatalf("type filter = %+v", onlyItems)
	}

	if _, ok := ParseQuery("   ", nil, 10); ok {
		t.Fatal("blank query should not search")
	}
	if got := ix.Search(Query{}); len(got) != 0 {
		t.Fatalf("empty = %+v", got)
	}

	limited := ix.Search(mustQuery(t, "e", nil, 1))
	if len(limited) != 1 {
		t.Fatalf("limit = %+v", limited)
	}

	spawner := ix.Search(mustQuery(t, "SP1", nil, 5))
	if len(spawner) != 1 || spawner[0].Path != "/creator/rooms?id=R0215" {
		t.Fatalf("spawner = %+v", spawner)
	}
	loot := ix.Search(mustQuery(t, "LT1", nil, 5))
	if len(loot) != 1 || loot[0].Path != "/creator/npcs?id=NPC1" {
		t.Fatalf("loot = %+v", loot)
	}
}

func TestReferencesGroupsBacklinks(t *testing.T) {
	ix := searchWorld()
	view, ok := ix.References(KindItem, "ITM0017")
	if !ok {
		t.Fatal("missing item")
	}
	if view.Name != "Strange Gear" {
		t.Fatalf("name %q", view.Name)
	}
	reason := inboundReason(t, view, KindQuest, "QST0217")
	if reason != "quest QST0217 objective deliver" {
		t.Fatalf("deliver reason %q", reason)
	}

	roomView, ok := ix.References(KindRoom, "R0215")
	if !ok {
		t.Fatal("missing room")
	}
	scriptReason := inboundReason(t, roomView, KindScript, "SCR0204")
	if scriptReason != "script SCR0204 revealExit" {
		t.Fatalf("script reason %q", scriptReason)
	}
	var exitReason string
	for _, group := range roomView.Outbound {
		if group.Type != KindRoom {
			continue
		}
		for _, ref := range group.Refs {
			if ref.ID == "R0230" {
				exitReason = ref.Reason
			}
		}
	}
	if exitReason != "room R0215 exit 'down'" {
		t.Fatalf("exit reason %q", exitReason)
	}

	if _, ok := ix.References(KindQuest, "NOPE"); ok {
		t.Fatal("unknown id")
	}
	if _, ok := ix.References(KindClass, "R0215"); ok {
		t.Fatal("wrong type")
	}
}

func TestCacheReloadsAfterInvalidate(t *testing.T) {
	cache := &Cache{}
	loads := 0
	load := func() (Snapshot, error) {
		loads++
		snap := NewSnapshot()
		one := room("R1")
		one.Name = "One"
		snap.Rooms["R1"] = one
		return snap, nil
	}
	first, err := cache.Index(load)
	if err != nil || first.Name(KindRoom, "R1") != "One" {
		t.Fatalf("first %v %v", first, err)
	}
	second, err := cache.Index(load)
	if err != nil || second != first || loads != 1 {
		t.Fatalf("cached loads=%d same=%v err=%v", loads, second == first, err)
	}
	cache.Invalidate()
	third, err := cache.Index(load)
	if err != nil || loads != 2 || third == first {
		t.Fatalf("reload loads=%d same=%v err=%v", loads, third == first, err)
	}
}

func inboundReason(t *testing.T, view Refs, kind Kind, id string) string {
	t.Helper()
	for _, group := range view.Inbound {
		if group.Type != kind {
			continue
		}
		for _, ref := range group.Refs {
			if ref.ID == id {
				return ref.Reason
			}
		}
	}
	t.Fatalf("no inbound %s %s in %+v", kind, id, view.Inbound)
	return ""
}

func mustQuery(t *testing.T, text string, types []string, limit int) Query {
	t.Helper()
	q, ok := ParseQuery(text, types, limit)
	if !ok {
		t.Fatalf("query %q", text)
	}
	return q
}
