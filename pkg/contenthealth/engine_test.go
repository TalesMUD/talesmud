package contenthealth

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/scripts"
)

func testBalance() *Balance {
	return &Balance{
		Tiers:     map[string]bool{"trivial": true, "easy": true, "normal": true, "hard": true, "boss": true},
		BossTiers: map[string]bool{"boss": true},
	}
}

func testRoom(id, name string, exits ...rooms.Exit) *rooms.Room {
	room := &rooms.Room{Entity: &entities.Entity{ID: id}, Name: name, Description: "A room."}
	if len(exits) > 0 {
		list := rooms.Exits(exits)
		room.Exits = &list
	}
	return room
}

func TestRunGraphRules(t *testing.T) {
	hidden := rooms.Exits{{Name: "secret", Target: "R1", Hidden: true}}
	start := testRoom("R0", "Start")
	start.Exits = &hidden
	items := rooms.Items{"MISSING"}
	start.Items = &items
	world := World{
		StartRoomID: "R0",
		Rooms: []*rooms.Room{
			start,
			testRoom("R1", "Island"),
			testRoom("R2", "Other", rooms.Exit{Name: "north", Target: "NOWHERE"}),
		},
		NPCs: []*npc.NPC{
			{Entity: &entities.Entity{ID: "BOSS"}, Name: "Boss", Level: 20, MaxHitPoints: 40, IsTemplate: false, SpawnRoomID: "R0", EnemyTrait: &npc.EnemyTrait{Difficulty: "boss", AttackPower: 2, Defense: 1}},
			{Entity: &entities.Entity{ID: "UNIQUE"}, Name: "Guard", Level: 2, MaxHitPoints: 10, IsTemplate: false, SpawnRoomID: "R0", EnemyTrait: &npc.EnemyTrait{Difficulty: "normal", AttackPower: 1, Defense: 1}},
			{Entity: &entities.Entity{ID: "ELITE"}, Name: "Leech", Level: 8, MaxHitPoints: 12, IsTemplate: true, EnemyTrait: &npc.EnemyTrait{Difficulty: "elite", AttackPower: 1, Defense: 1}},
			{Entity: &entities.Entity{ID: "SPAWNED"}, Name: "Spawned Boss", Level: 20, MaxHitPoints: 40, IsTemplate: true, EnemyTrait: &npc.EnemyTrait{Difficulty: "boss", AttackPower: 2, Defense: 1}},
			{Entity: &entities.Entity{ID: "COPY"}, Name: "Copy", TemplateID: "BOSS", Level: 20, MaxHitPoints: 40, EnemyTrait: &npc.EnemyTrait{Difficulty: "boss"}},
		},
		Spawners: []*npc.NPCSpawner{{Entity: &entities.Entity{ID: "SP1"}, TemplateID: "SPAWNED", RoomID: "R0", MaxInstances: 1, InitialCount: 1}},
		Quests: []*quests.Quest{{
			Entity: &entities.Entity{ID: "Q1"}, Name: "Visit",
			Objectives: []quests.Objective{{ID: "o1", Type: quests.ObjectiveVisit, TargetID: "R1", Amount: 1}},
		}},
		Scripts: []*scripts.Script{
			{Entity: &entities.Entity{ID: "S1"}, Name: "Unused", Language: "lua", Code: "print(1)"},
			{Entity: &entities.Entity{ID: "S2"}, Name: "Reveal", Language: "lua", Code: `tales.game.revealExit("R0", "deeper", "c1")`},
		},
	}
	world.Rooms[0].Actions = &rooms.Actions{{Name: "EXAMINE", Type: rooms.RoomActionTypeScript, ScriptId: "S2"}}
	report := Run(world, Options{Balance: testBalance()})

	unreachable := ruleByID(t, report, RuleUnreachable)
	if len(unreachable.Hits) != 2 {
		t.Fatalf("unreachable hits = %d, want 2 (%s)", len(unreachable.Hits), unreachable.Summary)
	}
	if !strings.Contains(unreachable.Hits[0].Message, "no inbound exit") && !strings.Contains(unreachable.Hits[1].Message, "hidden exit") && !strings.Contains(reportText(report), "no inbound exit") {
		t.Fatalf("reasons missing: %+v", unreachable.Hits)
	}
	if got := len(ruleByID(t, report, RuleDangling).Hits); got != 1 {
		t.Fatalf("dangling hits = %d, want 1", got)
	}
	if got := len(ruleByID(t, report, RuleMissingItem).Hits); got != 1 {
		t.Fatalf("missing item hits = %d, want 1", got)
	}
	bosses := ruleByID(t, report, RuleBoss)
	if len(bosses.Hits) != 1 || bosses.Hits[0].EntityID != "BOSS" {
		t.Fatalf("boss hits = %+v", bosses.Hits)
	}
	unknown := ruleByID(t, report, RuleUnknownTier)
	if len(unknown.Hits) != 1 || unknown.Hits[0].EntityID != "ELITE" {
		t.Fatalf("unknown hits = %+v", unknown.Hits)
	}
	quest := ruleByID(t, report, RuleQuest)
	if len(quest.Hits) != 1 || !strings.Contains(quest.Hits[0].Message, "Q1") {
		t.Fatalf("quest hits = %+v", quest.Hits)
	}
	if len(ruleByID(t, report, RuleUnreferenced).Hits) != 1 {
		t.Fatalf("unreferenced = %+v", ruleByID(t, report, RuleUnreferenced).Hits)
	}
	reveal := ruleByID(t, report, RuleRevealMissing)
	if len(reveal.Hits) != 1 || !strings.Contains(reveal.Hits[0].Message, "deeper") {
		t.Fatalf("reveal = %+v", reveal.Hits)
	}
	if len(ruleByID(t, report, RuleHiddenNoReveal).Hits) != 1 {
		t.Fatalf("hidden = %+v", ruleByID(t, report, RuleHiddenNoReveal).Hits)
	}

	muted := Run(world, Options{Balance: testBalance(), Muted: []string{RuleUnreachable, RuleDangling, RuleMissingItem, RuleQuest, RuleRevealMissing}})
	if muted.Summary.Errors != 0 {
		t.Fatalf("muted errors = %d", muted.Summary.Errors)
	}
	if muted.Summary.Muted == 0 {
		t.Fatal("muted hit count should be non-zero")
	}
	if Failed(muted, "error") {
		t.Fatal("warnings alone should not fail")
	}
	if !Failed(report, "error") {
		t.Fatal("errors should fail")
	}
}

func TestQuestRewardItemIsObtainable(t *testing.T) {
	start := testRoom("R0", "Start")
	clerk := &npc.NPC{Entity: &entities.Entity{ID: "N1"}, Name: "Clerk", SpawnRoomID: "R0"}
	item := &items.Item{Entity: &entities.Entity{ID: "ITM0059"}, Name: "Dispatch", IsTemplate: true}
	grant := &quests.Quest{
		Entity:     &entities.Entity{ID: "QST0302"},
		Name:       "Grant",
		Source:     quests.QuestSource{Type: "auto"},
		Objectives: []quests.Objective{{ID: "see", Type: quests.ObjectiveVisit, TargetID: "R0", Amount: 1}},
		Rewards:    quests.Reward{ItemTemplateIDs: []string{"ITM0059"}},
	}
	deliver := &quests.Quest{
		Entity:           &entities.Entity{ID: "QST0303"},
		Name:             "Deliver",
		Source:           quests.QuestSource{Type: "auto"},
		RequiredQuestIDs: []string{"QST0302"},
		Objectives: []quests.Objective{{
			ID: "drop", Type: quests.ObjectiveDeliver, TargetID: "ITM0059", DeliverToNPCID: "N1", Amount: 1,
		}},
	}
	world := World{
		StartRoomID: "R0",
		Rooms:       []*rooms.Room{start},
		NPCs:        []*npc.NPC{clerk},
		Items:       []*items.Item{item},
		Quests:      []*quests.Quest{deliver, grant},
	}
	report := Run(world, Options{Balance: testBalance()})
	for _, hit := range ruleHits(report, RuleQuest) {
		t.Fatalf("reward chain should complete, hit %+v", hit)
	}

	// A two-step reward chain: the second item exists only because the first quest's reward made the middle quest completable.
	nextItem := &items.Item{Entity: &entities.Entity{ID: "ITM0060"}, Name: "Seal", IsTemplate: true}
	middle := &quests.Quest{
		Entity:           &entities.Entity{ID: "QST0304"},
		Name:             "Middle",
		Source:           quests.QuestSource{Type: "auto"},
		RequiredQuestIDs: []string{"QST0302"},
		Objectives:       []quests.Objective{{ID: "take", Type: quests.ObjectiveCollect, TargetID: "ITM0059", Amount: 1}},
		Rewards:          quests.Reward{ItemTemplateIDs: []string{"ITM0060"}},
	}
	last := &quests.Quest{
		Entity:           &entities.Entity{ID: "QST0305"},
		Name:             "Last",
		Source:           quests.QuestSource{Type: "auto"},
		RequiredQuestIDs: []string{"QST0304"},
		Objectives:       []quests.Objective{{ID: "hand", Type: quests.ObjectiveDeliver, TargetID: "ITM0060", DeliverToNPCID: "N1", Amount: 1}},
	}
	world.Items = append(world.Items, nextItem)
	world.Quests = append(world.Quests, middle, last)
	report = Run(world, Options{Balance: testBalance()})
	for _, hit := range ruleHits(report, RuleQuest) {
		t.Fatalf("chained rewards should complete, hit %+v", hit)
	}
}

func TestQuestRewardCycleStaysImpossible(t *testing.T) {
	start := testRoom("R0", "Start")
	clerk := &npc.NPC{Entity: &entities.Entity{ID: "N1"}, Name: "Clerk", SpawnRoomID: "R0"}
	item := &items.Item{Entity: &entities.Entity{ID: "ITM0059"}, Name: "Dispatch", IsTemplate: true}
	grant := &quests.Quest{
		Entity:           &entities.Entity{ID: "QST0302"},
		Name:             "Grant",
		Source:           quests.QuestSource{Type: "auto"},
		RequiredQuestIDs: []string{"QST0303"},
		Objectives:       []quests.Objective{{ID: "see", Type: quests.ObjectiveVisit, TargetID: "R0", Amount: 1}},
		Rewards:          quests.Reward{ItemTemplateIDs: []string{"ITM0059"}},
	}
	deliver := &quests.Quest{
		Entity:           &entities.Entity{ID: "QST0303"},
		Name:             "Deliver",
		Source:           quests.QuestSource{Type: "auto"},
		RequiredQuestIDs: []string{"QST0302"},
		Objectives: []quests.Objective{{
			ID: "drop", Type: quests.ObjectiveDeliver, TargetID: "ITM0059", DeliverToNPCID: "N1", Amount: 1,
		}},
	}
	world := World{
		StartRoomID: "R0",
		Rooms:       []*rooms.Room{start},
		NPCs:        []*npc.NPC{clerk},
		Items:       []*items.Item{item},
		Quests:      []*quests.Quest{grant, deliver},
	}
	report := Run(world, Options{Balance: testBalance()})
	hits := ruleHits(report, RuleQuest)
	if len(hits) != 1 || hits[0].EntityID != "QST0303" || !strings.Contains(hits[0].Message, "ITM0059") {
		t.Fatalf("cycle hits = %+v", hits)
	}
}

func TestInfoDoesNotFail(t *testing.T) {
	world := World{StartRoomID: "R0", Rooms: []*rooms.Room{testRoom("R0", "Start")}}
	report := Run(world, Options{
		Balance: testBalance(),
		Rules: []PackRule{{
			ID: "note", Severity: "info", Entity: "room", Message: "named",
			Require: []Condition{{Field: "name", Op: "eq", Value: "missing"}},
		}},
	})
	if report.Summary.Info != 1 || report.Summary.Errors != 0 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if Failed(report, "warning") {
		t.Fatal("info should not fail")
	}
	found := false
	for _, rule := range report.Rules {
		if rule.ID == "note" && len(rule.Hits) == 1 {
			found = true
		}
	}
	if !found {
		t.Fatal("pack rule missing")
	}
}

func ruleHits(report Report, id string) []Hit {
	for _, rule := range report.Rules {
		if rule.ID == id {
			return rule.Hits
		}
	}
	return nil
}

func ruleByID(t *testing.T, report Report, id string) Rule {
	t.Helper()
	for _, rule := range report.Rules {
		if rule.ID == id {
			return rule
		}
	}
	t.Fatalf("missing rule %s", id)
	return Rule{}
}

func reportText(report Report) string {
	return FormatText(report, 30)
}
