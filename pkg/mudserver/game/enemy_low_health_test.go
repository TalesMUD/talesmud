package game

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

type summonLogHook struct {
	lines []string
}

func (h *summonLogHook) Levels() []log.Level { return log.AllLevels }

func (h *summonLogHook) Fire(e *log.Entry) error {
	h.lines = append(h.lines, e.Message)
	return nil
}

func (h *summonLogHook) has(substr string) bool {
	for _, line := range h.lines {
		if line == substr {
			return true
		}
	}
	return false
}

func captureSummonLogs(t *testing.T) *summonLogHook {
	t.Helper()
	logger := log.StandardLogger()
	prev := logger.Hooks
	hook := &summonLogHook{}
	logger.ReplaceHooks(make(log.LevelHooks))
	logger.AddHook(hook)
	t.Cleanup(func() { logger.ReplaceHooks(prev) })
	return hook
}

func startLowFight(t *testing.T, room, scriptID, script, enemyID string, startHP int32, trait *npc.EnemyTrait) (*Game, *combat.CombatInstance, *characters.Character) {
	t.Helper()
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, room, nil)
	if scriptID != "" {
		storeHookScript(t, facade, scriptID, script)
	}
	_, hero := sharePlayer(t, facade, "user-"+enemyID, "ref-"+enemyID, "char-"+enemyID, "Hero", room)
	enemy := hookEnemy(enemyID, enemyID, startHP, 100, trait)
	g.NPCManager.RegisterExistingNPC(enemy, room)
	inst := g.CombatController.InitiateCombat(room, []*characters.Character{hero}, []*npc.NPC{enemy})
	if inst == nil || inst.State != combat.CombatStateActive {
		t.Fatal("fight did not start")
	}
	_ = drainGameMessages(g.SendMessage())
	return g, inst, hero
}

func TestOnLowHealthFiresOncePerFight(t *testing.T) {
	g, inst, _ := startLowFight(t, "R-low", "SCR-LOW", `tales.game.msgToRoom(ctx.roomId, "the runes flare")`, "knight", 100, &npc.EnemyTrait{
		OnLowHealthScript: "SCR-LOW", Difficulty: "easy",
	})
	ref := inst.GetEnemyByID("knight")
	ref.CurrentHP = 20
	g.CombatController.flushEnemyLowHealthHooks(inst)
	g.CombatController.flushEnemyLowHealthHooks(inst)
	ref.CurrentHP = 10
	g.CombatController.flushEnemyLowHealthHooks(inst)
	if got := countRoomText(drainGameMessages(g.SendMessage()), "the runes flare"); got != 1 {
		t.Fatalf("onLowHealth messages = %d, want 1", got)
	}
}

func TestOnLowHealthCustomAndDefaultThreshold(t *testing.T) {
	g, inst, _ := startLowFight(t, "R-custom", "SCR-TH", `tales.game.msgToRoom(ctx.roomId, "below the line")`, "custom", 100, &npc.EnemyTrait{
		OnLowHealthScript: "SCR-TH", LowHealthThreshold: 0.5, Difficulty: "easy",
	})
	if got := inst.GetEnemyByID("custom").LowHealthThreshold; got != 0.5 {
		t.Fatalf("custom threshold = %v", got)
	}
	ref := inst.GetEnemyByID("custom")
	ref.CurrentHP = 50
	g.CombatController.flushEnemyLowHealthHooks(inst)
	if got := countRoomText(drainGameMessages(g.SendMessage()), "below the line"); got != 0 {
		t.Fatalf("HP on the 50%% line fired %d times", got)
	}
	ref.CurrentHP = 49
	g.CombatController.flushEnemyLowHealthHooks(inst)
	if got := countRoomText(drainGameMessages(g.SendMessage()), "below the line"); got != 1 {
		t.Fatalf("custom threshold messages = %d, want 1", got)
	}

	g2, inst2, _ := startLowFight(t, "R-default", "SCR-DEF", `tales.game.msgToRoom(ctx.roomId, "default line")`, "plain", 100, &npc.EnemyTrait{
		OnLowHealthScript: "SCR-DEF", Difficulty: "easy",
	})
	if got := inst2.GetEnemyByID("plain").LowHealthThreshold; got != npc.DefaultLowHealthFraction {
		t.Fatalf("default threshold = %v", got)
	}
	ref2 := inst2.GetEnemyByID("plain")
	ref2.CurrentHP = 30
	g2.CombatController.flushEnemyLowHealthHooks(inst2)
	if got := countRoomText(drainGameMessages(g2.SendMessage()), "default line"); got != 0 {
		t.Fatalf("HP on the default 30%% line fired %d times", got)
	}
	ref2.CurrentHP = 29
	g2.CombatController.flushEnemyLowHealthHooks(inst2)
	if got := countRoomText(drainGameMessages(g2.SendMessage()), "default line"); got != 1 {
		t.Fatalf("default threshold messages = %d, want 1", got)
	}
}

func TestOnLowHealthSkipsOverkillAndAlreadyLow(t *testing.T) {
	g, inst, _ := startLowFight(t, "R-over", "SCR-OVER", `tales.game.msgToRoom(ctx.roomId, "should not flare")`, "knight", 100, &npc.EnemyTrait{
		OnLowHealthScript: "SCR-OVER", Difficulty: "easy",
	})
	ref := inst.GetEnemyByID("knight")
	ref.CurrentHP = 0
	ref.IsAlive = false
	g.CombatController.flushEnemyLowHealthHooks(inst)
	if got := countRoomText(drainGameMessages(g.SendMessage()), "should not flare"); got != 0 {
		t.Fatalf("overkill fired onLowHealth %d times", got)
	}

	g2, inst2, _ := startLowFight(t, "R-already", "SCR-ALREADY", `tales.game.msgToRoom(ctx.roomId, "already low")`, "wounded", 20, &npc.EnemyTrait{
		OnLowHealthScript: "SCR-ALREADY", Difficulty: "easy",
	})
	if inst2.GetEnemyByID("wounded").CurrentHP != 20 {
		t.Fatalf("start HP = %d", inst2.GetEnemyByID("wounded").CurrentHP)
	}
	g2.CombatController.flushEnemyLowHealthHooks(inst2)
	inst2.GetEnemyByID("wounded").CurrentHP = 5
	g2.CombatController.flushEnemyLowHealthHooks(inst2)
	if got := countRoomText(drainGameMessages(g2.SendMessage()), "already low"); got != 0 {
		t.Fatalf("NPC that started below the line fired %d times", got)
	}
}

func TestOnLowHealthScriptErrorDoesNotStopCombat(t *testing.T) {
	g, inst, hero := startLowFight(t, "R-err", "SCR-ERR-LOW", `error("boom")`, "knight", 100, &npc.EnemyTrait{
		OnLowHealthScript: "SCR-ERR-LOW", Difficulty: "easy",
	})
	inst.GetEnemyByID("knight").CurrentHP = 10
	g.CombatController.flushEnemyLowHealthHooks(inst)
	if inst.State != combat.CombatStateActive {
		t.Fatalf("state = %s", inst.State)
	}
	if g.CombatController.GetCombatInstance(hero.ID) == nil {
		t.Fatal("fight ended after a script error")
	}
}

func importEnemyTemplate(t *testing.T, g *Game, id, scriptID string, enemy bool, xp int64) {
	t.Helper()
	n := &npc.NPC{
		Entity:           &entities.Entity{ID: id},
		Name:             id,
		IsTemplate:       true,
		Level:            1,
		MaxHitPoints:     8,
		CurrentHitPoints: 8,
	}
	if enemy {
		n.EnemyTrait = &npc.EnemyTrait{
			AttackPower: 1, Difficulty: "easy", XPReward: xp, OnAggroScript: scriptID,
			GuaranteedLoot: []string{"tpl-summon-fang"},
		}
	}
	if _, err := g.Facade.NPCsService().Import(n); err != nil {
		t.Fatal(err)
	}
}

func TestSummonAddsRespectCapsAndFailures(t *testing.T) {
	g, inst, _ := startLowFight(t, "R-sum", "SCR-SUM", `local n = tales.combat.summon("TPL-RAT", 2)
tales.game.msgToRoom(ctx.roomId, "spawned-" .. tostring(n))`, "knight", 100, &npc.EnemyTrait{
		OnLowHealthScript: "SCR-SUM", Difficulty: "easy",
	})
	importEnemyTemplate(t, g, "TPL-RAT", "", true, 80)
	if _, err := g.Facade.ItemsService().Import(&items.Item{
		Entity: &entities.Entity{ID: "tpl-summon-fang"}, Name: "Fang", IsTemplate: true,
	}); err != nil {
		t.Fatal(err)
	}
	inst.GetEnemyByID("knight").CurrentHP = 10
	g.CombatController.flushEnemyLowHealthHooks(inst)
	if got := countRoomText(drainGameMessages(g.SendMessage()), "spawned-2"); got != 1 {
		t.Fatalf("script summon result missing, messages drained without spawned-2")
	}
	if len(inst.Players) != 1 {
		t.Fatalf("opponents changed: %d players", len(inst.Players))
	}
	if len(inst.Enemies) != 3 || inst.SummonsUsed != 2 {
		t.Fatalf("enemies=%d summons=%d", len(inst.Enemies), inst.SummonsUsed)
	}
	adds := 0
	var addID string
	for _, e := range inst.Enemies {
		if !e.Summoned {
			continue
		}
		adds++
		addID = e.ID
		if e.MaxHP != 8 || e.CurrentHP != 8 {
			t.Fatalf("summon stats = %+v", e)
		}
		found := false
		for _, turn := range inst.TurnOrder {
			if turn.ID == e.ID {
				found = true
			}
		}
		if !found {
			t.Fatalf("summon %s missing from turn order", e.ID)
		}
	}
	if adds != 2 {
		t.Fatalf("summoned flags = %d", adds)
	}
	inRoom := g.NPCManager.GetInstancesInRoom("R-sum")
	seen := 0
	for _, n := range inRoom {
		if n.TemplateID == "TPL-RAT" {
			seen++
		}
	}
	if seen != 2 {
		t.Fatalf("room copies = %d", seen)
	}

	logs := captureSummonLogs(t)
	if got := g.CombatController.SummonCombatAllies("knight", "TPL-RAT", 5); got != 1 {
		t.Fatalf("remaining cap spawned %d, want 1", got)
	}
	if got := g.CombatController.SummonCombatAllies("knight", "TPL-RAT", 1); got != 0 || !logs.has("summon: cap reached") {
		t.Fatalf("full cap returned %d logs=%v", got, logs.lines)
	}
	if got := g.CombatController.SummonCombatAllies("knight", "NO-SUCH", 1); got != 0 || !logs.has("summon: unknown template") {
		t.Fatalf("unknown template returned %d logs=%v", got, logs.lines)
	}
	importEnemyTemplate(t, g, "TPL-CIV", "", false, 0)
	if got := g.CombatController.SummonCombatAllies("knight", "TPL-CIV", 1); got != 0 || !logs.has("summon: not an enemy") {
		t.Fatalf("non-enemy returned %d logs=%v", got, logs.lines)
	}
	if got := g.CombatController.SummonCombatAllies("nobody", "TPL-RAT", 1); got != 0 || !logs.has("summon: no active fight") {
		t.Fatalf("no fight returned %d logs=%v", got, logs.lines)
	}

	for i := range inst.Enemies {
		if inst.Enemies[i].Summoned {
			inst.Enemies[i].IsAlive = false
			inst.Enemies[i].CurrentHP = 0
		}
	}
	g.CombatController.processCombatVictory(inst)
	gotChar, err := g.Facade.CharactersService().FindByID("char-knight")
	if err != nil {
		t.Fatal(err)
	}
	if gotChar.XP != 0 {
		t.Fatalf("summoned adds granted xp %d", gotChar.XP)
	}
	room, err := g.Facade.RoomsService().FindByID("R-sum")
	if err != nil {
		t.Fatal(err)
	}
	if room.Items != nil && len(*room.Items) != 0 {
		t.Fatalf("summoned adds dropped %#v", room.Items)
	}

	g.CombatController.engine.EndCombat(inst, combat.CombatStateFled)
	g.CombatController.cleanupCombatInstance(inst, combat.CombatStateFled)
	if g.NPCManager.GetInstance(addID) != nil {
		t.Fatal("summoned add persisted after the fight")
	}
	if g.NPCManager.GetInstance("knight") == nil {
		t.Fatal("the original NPC was removed with the adds")
	}
}

func TestSummonPerCallClampAndAggroLoopStops(t *testing.T) {
	g, inst, _ := startLowFight(t, "R-cap", "", "", "knight", 100, &npc.EnemyTrait{Difficulty: "easy"})
	storeHookScript(t, g.Facade, "SCR-LOOP", `tales.combat.summon("TPL-LOOP", 3)`)
	importEnemyTemplate(t, g, "TPL-LOOP", "SCR-LOOP", true, 1)
	if got := g.CombatController.SummonCombatAllies("knight", "TPL-LOOP", 9); got != 1 {
		t.Fatalf("outer summon = %d, want 1 (the rest come from onAggro until the cap)", got)
	}
	if inst.SummonsUsed != MaxSummonsPerFight || len(inst.Enemies) != 4 {
		t.Fatalf("used=%d enemies=%d", inst.SummonsUsed, len(inst.Enemies))
	}

	g2, _, _ := startLowFight(t, "R-clamp", "", "", "boss", 100, &npc.EnemyTrait{Difficulty: "easy"})
	importEnemyTemplate(t, g2, "TPL-ONE", "", true, 1)
	if got := g2.CombatController.SummonCombatAllies("boss", "TPL-ONE", 5); got != 3 {
		t.Fatalf("per-call clamp spawned %d, want 3", got)
	}
}

func TestSummonSeparateCallsShareFightCap(t *testing.T) {
	g, _, _ := startLowFight(t, "R-split", "", "", "knight", 100, &npc.EnemyTrait{Difficulty: "easy"})
	importEnemyTemplate(t, g, "TPL-SPLIT", "", true, 1)
	if got := g.CombatController.SummonCombatAllies("knight", "TPL-SPLIT", 2); got != 2 {
		t.Fatalf("first call = %d", got)
	}
	if got := g.CombatController.SummonCombatAllies("knight", "TPL-SPLIT", 2); got != 1 {
		t.Fatalf("second call = %d, want 1", got)
	}
}
