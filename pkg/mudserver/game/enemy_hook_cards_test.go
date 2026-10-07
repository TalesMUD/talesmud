package game

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/scripts"
)

func roomLine(msgs []interface{}, needle string) (messages.MessageResponse, bool) {
	for _, out := range msgs {
		msg, ok := out.(messages.MessageResponse)
		if ok && strings.Contains(msg.Message, needle) {
			return msg, true
		}
	}
	return messages.MessageResponse{}, false
}

func TestHookRoomLineCarriesCombatEvent(t *testing.T) {
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, "R-card", nil)
	storeHookScript(t, facade, "SCR-CARD", `tales.game.msgToRoom(ctx.roomId, "The chanter hits the drum once.")`)
	_, hero := sharePlayer(t, facade, "user-card", "ref-card", "char-card", "Hero", "R-card")
	enemy := hookEnemy("chanter", "The Hollow Knight", 40, 40, &npc.EnemyTrait{
		AttackPower: 1, OnAggroScript: "SCR-CARD", Difficulty: "easy",
	})
	g.NPCManager.RegisterExistingNPC(enemy, "R-card")
	inst := g.CombatController.InitiateCombat("R-card", []*characters.Character{hero}, []*npc.NPC{enemy})
	if inst == nil {
		t.Fatal("fight did not start")
	}
	g.CombatController.runEnemyHooksOnEnter(inst)
	msg, ok := roomLine(drainGameMessages(g.SendMessage()), "The chanter hits the drum once.")
	if !ok {
		t.Fatal("hook line missing")
	}
	if msg.Type != messages.MessageTypeDefault || msg.Username != "SYSTEM" {
		t.Fatalf("type/username = %s %s", msg.Type, msg.Username)
	}
	if msg.Style != "combatEvent" || msg.Hook != "onAggro" || msg.Source != "The Hollow Knight" {
		t.Fatalf("event = style %q hook %q source %q", msg.Style, msg.Hook, msg.Source)
	}
}

func TestPlainScriptRoomLineOmitsCombatEvent(t *testing.T) {
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, "R-plain", nil)
	run := func(hook, source string) messages.MessageResponse {
		t.Helper()
		ctx := scripts.NewScriptContext()
		ctx.Set("roomId", "R-plain")
		if hook != "" {
			ctx.SetCombatHook(hook, source)
		}
		res := facade.Runner().RunWithResult(scripts.Script{
			Entity:   &entities.Entity{ID: "SCR-PLAIN"},
			Name:     "plain",
			Code:     `tales.game.msgToRoom(ctx.roomId, "the bricks drip")`,
			Type:     scripts.ScriptTypeRoom,
			Language: scripts.ScriptLanguageLua,
		}, ctx)
		if res == nil || !res.Success {
			t.Fatalf("script failed: %+v", res)
		}
		msg, ok := roomLine(drainGameMessages(g.SendMessage()), "the bricks drip")
		if !ok {
			t.Fatal("plain line missing")
		}
		return msg
	}
	plain := run("", "")
	if plain.Style != "" || plain.Hook != "" || plain.Source != "" {
		t.Fatalf("plain line tagged: %+v", plain)
	}
	hooked := run("onLowHealth", "Sewer Rat")
	if hooked.Style != "combatEvent" || hooked.Hook != "onLowHealth" || hooked.Source != "Sewer Rat" {
		t.Fatalf("hooked line = %+v", hooked)
	}
	again := run("", "")
	if again.Style != "" || again.Hook != "" || again.Source != "" {
		t.Fatalf("stamp leaked onto the next script: %+v", again)
	}
}

func TestSummonFlushSendsRoster(t *testing.T) {
	g, inst, _ := startLowFight(t, "R-roster", "SCR-ROSTER", `tales.combat.summon("TPL-ROSTER", 1)`, "knight", 100, &npc.EnemyTrait{
		OnLowHealthScript: "SCR-ROSTER", Difficulty: "easy",
	})
	importEnemyTemplate(t, g, "TPL-ROSTER", "", true, 0)
	inst.GetEnemyByID("knight").CurrentHP = 10
	g.CombatController.flushEnemyLowHealthHooks(inst)
	var status *messages.CombatStatusMessage
	for _, out := range drainGameMessages(g.SendMessage()) {
		msg, ok := out.(*messages.CombatStatusMessage)
		if ok {
			status = msg
		}
	}
	if status == nil {
		t.Fatal("summon did not send combatStatus")
	}
	if status.Type != messages.MessageTypeCombatStatus {
		t.Fatalf("type = %s", status.Type)
	}
	if status.Message != "" || status.Style != "" {
		t.Fatalf("roster line should be empty and unstyled: %+v", status.MessageResponse)
	}
	found := false
	for _, view := range status.Combatants {
		if view.Name == "TPL-ROSTER" && view.IsAlive {
			found = true
		}
	}
	if !found {
		t.Fatalf("roster = %+v", status.Combatants)
	}
}
