package worldindex

import "testing"

func TestExtractLuaRevealUsesLocalString(t *testing.T) {
	src := `
local character = ctx.character
local charID = character.ID
local roomID = "R0215"
if tales.game.revealExit(roomID, "deeper", charID) then
  tales.game.msgToCharacter(charID, "opened")
end
`
	edges := ExtractLuaEdges("SCR1", src)
	if len(edges) != 1 {
		t.Fatalf("edges = %+v", edges)
	}
	if edges[0].ToType != KindRoom || edges[0].ToID != "R0215" || edges[0].ExitName != "deeper" || edges[0].How != "revealExit" {
		t.Fatalf("edge = %+v", edges[0])
	}
}

func TestExtractLuaLiteralCalls(t *testing.T) {
	src := `
tales.combat.summon("ENM0008", 2)
tales.game.giveItem(charID, "ITM0001")
tales.npcs.moveTo("NPC1", "R0002")
tales.characters.teleport(charID, "R0009")
tales.quests.accept(charID, "QST0001")
tales.npcs.spawnFromTemplate("ENM0002", "R0004")
`
	edges := ExtractLuaEdges("SCR2", src)
	want := map[string]bool{
		"npc:ENM0008:summon":            false,
		"item:ITM0001:giveItem":         false,
		"npc:NPC1:moveTo":               false,
		"room:R0002:moveTo":             false,
		"room:R0009:teleport":           false,
		"quest:QST0001:startQuest":      false,
		"npc:ENM0002:spawnFromTemplate": false,
		"room:R0004:spawnFromTemplate":  false,
	}
	for _, e := range edges {
		k := string(e.ToType) + ":" + e.ToID + ":" + e.How
		if _, ok := want[k]; !ok {
			t.Errorf("unexpected edge %s", k)
			continue
		}
		want[k] = true
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("missing edge %s", k)
		}
	}
}

func TestExtractLuaCurrentRoom(t *testing.T) {
	src := `
local room = ctx.room
local roomID = room.ID
tales.game.revealExit(roomID, "up", charID)
tales.game.revealExit(ctx.roomID, "north", charID)
tales.game.revealExit(ctx.room.ID, "down", charID)
`
	edges := ExtractLuaEdges("SCR3", src)
	if len(edges) != 3 {
		t.Fatalf("edges = %+v", edges)
	}
	for _, edge := range edges {
		if edge.ToID != CurrentRoom || edge.How != "revealExit" {
			t.Fatalf("edge = %+v", edge)
		}
	}
	if edges[0].ExitName != "up" || edges[1].ExitName != "north" || edges[2].ExitName != "down" {
		t.Fatalf("exit names = %+v", edges)
	}
}

func TestExtractLuaClearsRebinding(t *testing.T) {
	src := `
local roomID = "R0001"
roomID = ctx.room.ID
tales.game.revealExit(roomID, "up", charID)
`
	edges := ExtractLuaEdges("SCR4", src)
	if len(edges) != 1 || edges[0].ToID != CurrentRoom || edges[0].ExitName != "up" {
		t.Fatalf("rebound room should follow ctx.room, got %+v", edges)
	}
}
