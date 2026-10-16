package guide

import "testing"

func catacombs() World {
	return World{
		CurrentRoomID: "hall",
		Rooms: []Room{
			{ID: "hall", Exits: []Exit{
				{Name: "north", Target: "rubble"},
				{Name: "east", Target: "nest"},
				{Name: "secret", Target: "nest", Hidden: true},
			}},
			{ID: "rubble", Exits: []Exit{{Name: "south", Target: "hall"}}},
			{ID: "nest", Exits: []Exit{{Name: "west", Target: "hall"}}},
			{ID: "gate", Exits: []Exit{{Name: "west", Target: "hall", Locked: true}}},
		},
		Places: map[string][]string{"rat": {"nest"}, "mira": {"gate"}},
		Quests: []Quest{
			{
				ID: "find-exit", Name: "Find the Exit", Level: 1, Status: "active", SourceType: "auto",
				TurnInAnywhere: true,
				Objectives:     []Objective{{Type: "visit", Description: "Reach the stair", TargetID: "gate", Order: 1}},
			},
			{
				ID: "pest", Name: "Pest Control", Level: 1, Status: "active", SourceType: "auto",
				TurnInAnywhere: true,
				Objectives: []Objective{{
					Type: "kill", Description: "Kill 3 cave rats", TargetID: "rat", TargetName: "Cave Rat",
					Amount: 3, Order: 1,
				}},
			},
			{
				ID: "scavenge", Name: "Scavenge", Level: 2, Status: "active", SourceType: "auto",
				Objectives: []Objective{{Type: "collect", Description: "Pick up a page", TargetID: "page", Amount: 1}},
			},
		},
	}
}

func TestFallbackPrefersKillAndHeadsThere(t *testing.T) {
	h := Build(catacombs())
	if h == nil {
		t.Fatal("expected a hint")
	}
	if h.QuestName != "Pest Control" || h.Step != "head" || h.Command != "east" {
		t.Fatalf("hint = %+v", h)
	}
	if h.Text != "Head east. Kill 3 cave rats" {
		t.Fatalf("text = %q", h.Text)
	}
}

func TestFightWhenTargetIsHere(t *testing.T) {
	w := catacombs()
	w.CurrentRoomID = "nest"
	w.Here = []Actor{{TemplateID: "rat", Name: "Cave Rat"}}
	h := Build(w)
	if h == nil || h.Step != "fight" || h.Command != "attack Cave Rat" || h.Label != "Fight" {
		t.Fatalf("hint = %+v", h)
	}
	if h.Text != "Tap the Cave Rat to fight." {
		t.Fatalf("text = %q", h.Text)
	}
}

func TestTurnInAfterKill(t *testing.T) {
	w := catacombs()
	w.Quests[1].Objectives[0].Completed = true
	w.Quests[1].Objectives[0].Current = 3
	h := Build(w)
	if h == nil || h.Step != "turnin" || h.Command != "complete Pest Control" {
		t.Fatalf("hint = %+v", h)
	}
}

func TestDismissedAndCompletedStayHidden(t *testing.T) {
	w := catacombs()
	w.Dismissed = true
	if Build(w) != nil {
		t.Fatal("dismissed character still got a hint")
	}
	w.Dismissed = false
	w.Quests[0].Status = "completed"
	if Build(w) != nil {
		t.Fatal("completed quest still got a hint")
	}
}

func TestGuideFlagBeatsActiveKill(t *testing.T) {
	w := catacombs()
	w.Quests = append(w.Quests, Quest{
		ID: "herbs", Name: "Herb Run", Level: 1, Guide: true, Status: "",
		SourceType: "npc", SourceNPCID: "mira", SourceNPCName: "Mira",
		Objectives: []Objective{{Type: "talk", TargetID: "mira", TargetName: "Mira", Description: "Speak with Mira"}},
	})
	h := Build(w)
	if h == nil || h.QuestName != "Herb Run" || h.Step != "accept" {
		t.Fatalf("hint = %+v", h)
	}
	if h.Command != "" {
		t.Fatalf("locked door must not be a command, got %q", h.Command)
	}
}

func TestHiddenExitSkippedUntilRevealed(t *testing.T) {
	w := catacombs()
	w.Rooms[0].Exits[1].Hidden = true
	h := Build(w)
	if h == nil {
		t.Fatal("expected text")
	}
	if h.Command != "" {
		t.Fatalf("hidden east was offered: %+v", h)
	}
	w.Revealed = map[string][]string{"hall": {"east"}}
	h = Build(w)
	if h == nil || h.Command != "east" {
		t.Fatalf("revealed east missing: %+v", h)
	}
}

func TestUnacceptedGuideTalksInRoom(t *testing.T) {
	w := World{
		CurrentRoomID: "hall",
		Here:          []Actor{{TemplateID: "mira", Name: "Mira"}},
		Quests: []Quest{{
			ID: "herbs", Name: "Herb Run", Guide: true, SourceType: "npc",
			SourceNPCID: "mira", SourceNPCName: "Mira",
		}},
	}
	h := Build(w)
	if h == nil || h.Command != "talk Mira" || h.Step != "accept" {
		t.Fatalf("hint = %+v", h)
	}
}

func TestNPCTurnInHeadsToGiver(t *testing.T) {
	w := World{
		CurrentRoomID: "hall",
		Rooms: []Room{
			{ID: "hall", Exits: []Exit{{Name: "north", Target: "shop"}}},
			{ID: "shop", Exits: []Exit{{Name: "south", Target: "hall"}}},
		},
		Places: map[string][]string{"mira": {"shop"}},
		Quests: []Quest{{
			ID: "herbs", Name: "Herb Run", Guide: true, Status: "active", Level: 1,
			SourceType: "npc", SourceNPCID: "mira", TurnInNPCID: "mira", TurnInNPCName: "Mira",
			Objectives: []Objective{{Type: "collect", Description: "Bring herbs", Completed: true, Amount: 1, Current: 1}},
		}},
	}
	h := Build(w)
	if h == nil || h.Step != "turnin" || h.Command != "north" {
		t.Fatalf("hint = %+v", h)
	}
}
