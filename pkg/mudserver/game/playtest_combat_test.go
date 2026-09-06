package game

import (
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/worldmap"
)

func TestSwarmCombatStylePullsRoomHostiles(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R-rats", nil)

	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity:           &entities.Entity{ID: "ENM-swarm"},
		Name:             "Sewer Rat",
		IsTemplate:       true,
		MaxHitPoints:     10,
		CurrentHitPoints: 10,
		Level:            1,
		EnemyTrait: &npc.EnemyTrait{
			CombatStyle: npc.CombatStyleSwarm,
			AttackPower: 1,
		},
	}); err != nil {
		t.Fatalf("import template: %v", err)
	}

	if _, err := g.NPCManager.SpawnInstanceDirect("ENM-swarm", "R-rats"); err != nil {
		t.Fatalf("spawn 1: %v", err)
	}
	if _, err := g.NPCManager.SpawnInstanceDirect("ENM-swarm", "R-rats"); err != nil {
		t.Fatalf("spawn 2: %v", err)
	}

	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}}
	character, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-1"},
		Name:             "Wanderer",
		BelongsUser:      *traits.BelongsToUser("user-1"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-rats"},
		MaxHitPoints:     50,
		CurrentHitPoints: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	msg := &messages.Message{
		FromUser:  user,
		Character: character,
		Data:      "attack Sewer Rat",
	}
	if !(&commands.AttackCommand{}).Execute(g, msg) {
		t.Fatal("attack failed")
	}
	combatInst := g.CombatController.GetCombatInstance(character.ID)
	if combatInst == nil {
		t.Fatal("expected combat instance")
	}
	if len(combatInst.Enemies) < 2 {
		t.Fatalf("swarm should pull both rats, got %d", len(combatInst.Enemies))
	}
}

func TestCombatGraceBlocksImmediateReengage(t *testing.T) {
	g, _ := newNPCTestGame(t)
	g.CombatController.markCombatGrace("char-grace")
	if !g.CombatController.CombatGraceActive("char-grace") {
		t.Fatal("expected grace active immediately after mark")
	}
	g.CombatController.graceUntil["char-grace"] = time.Now().Add(-time.Second)
	if g.CombatController.CombatGraceActive("char-grace") {
		t.Fatal("expired grace should clear")
	}
}

func TestAtlasMessageHasEmptyPlayerVisibleText(t *testing.T) {
	msg := messages.NewAtlasMessage("user-1", worldmap.PlayerMap{})
	if msg.Message != "" {
		t.Fatalf("atlas frame must not emit player-visible text, got %q", msg.Message)
	}
	if msg.Type != messages.MessageTypeAtlas {
		t.Fatalf("unexpected type %q", msg.Type)
	}
}

func TestEnterRoomMessageStripsHiddenExits(t *testing.T) {
	g, _ := newNPCTestGame(t)
	exits := rooms.Exits{
		{Name: "north", Target: "a"},
		{Name: "secret", Target: "b", Hidden: true},
	}
	chars := rooms.Characters{}
	items := rooms.Items{}
	room := &rooms.Room{
		Entity:     &entities.Entity{ID: "r-hidden"},
		Name:       "Hall",
		Exits:      &exits,
		Characters: &chars,
		Items:      &items,
	}
	out := messages.NewEnterRoomMessage(room, &entities.User{Entity: &entities.Entity{ID: "u"}}, g, nil)
	if out.Room.Exits == nil {
		t.Fatal("expected exits")
	}
	for _, e := range *out.Room.Exits {
		if e.Hidden || e.Name == "secret" {
			t.Fatalf("hidden exit leaked: %#v", e)
		}
	}
}
