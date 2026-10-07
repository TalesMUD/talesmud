package game

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/scripts/runner"
	"github.com/talesmud/talesmud/pkg/service"
)

const warChantLua = `
local roomID = ctx.roomId
local line = "The chanter hits the drum once."
local allies = ctx.allies
local worst = nil
local worstRatio = 2
if allies then
  local n = #allies
  for i = 1, n do
    local ally = allies[i]
    if ally and ally.Alive and ally.MaxHP and ally.MaxHP > 0 and ally.HP < ally.MaxHP then
      local ratio = ally.HP / ally.MaxHP
      if ratio < worstRatio then
        worstRatio = ratio
        worst = ally
      end
    end
  end
end
if worst then
  local healed = tales.combat.healNpc(worst.ID, 18)
  if healed and healed > 0 then
    line = "The chanter hits the drum once. " .. worst.Name .. " stands a little straighter."
  end
end
if roomID and roomID ~= "" then
  tales.game.msgToRoom(roomID, line)
end
`

func newHookTestGame(t *testing.T) (*Game, service.Facade) {
	t.Helper()
	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	mr := runner.NewMultiRunner()
	t.Cleanup(mr.Shutdown)
	facade := service.NewFacade(repository.NewSQLiteFactory(client), mr)
	g := New(facade)
	mr.SetServices(facade, g)
	return g, facade
}

func storeHookScript(t *testing.T, facade service.Facade, id, code string) {
	t.Helper()
	if _, err := facade.ScriptsService().Import(&scripts.Script{
		Entity:   &entities.Entity{ID: id},
		Name:     id,
		Code:     code,
		Type:     scripts.ScriptTypeNPC,
		Language: scripts.ScriptLanguageLua,
	}); err != nil {
		t.Fatalf("import script %s: %v", id, err)
	}
}

func hookEnemy(id, name string, hp, max int32, trait *npc.EnemyTrait) *npc.NPC {
	return &npc.NPC{
		Entity:           &entities.Entity{ID: id},
		Name:             name,
		Level:            1,
		MaxHitPoints:     max,
		CurrentHitPoints: hp,
		EnemyTrait:       trait,
	}
}

func countRoomText(msgs []interface{}, needle string) int {
	n := 0
	for _, out := range msgs {
		msg, ok := out.(messages.MessageResponse)
		if ok && strings.Contains(msg.Message, needle) {
			n++
		}
	}
	return n
}

func countLogResult(inst *combat.CombatInstance, result string) int {
	n := 0
	for _, entry := range inst.Log {
		if entry.Result == result {
			n++
		}
	}
	return n
}

func TestEnemyOnAggroRunsOncePerFight(t *testing.T) {
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, "R-hook", nil)
	storeHookScript(t, facade, "SCR-AGGRO", `tales.game.msgToRoom(ctx.roomId, "The chanter hits the drum once.")`)
	_, hero := sharePlayer(t, facade, "user-aggro", "ref-aggro", "char-aggro", "Hero", "R-hook")
	enemy := hookEnemy("chanter", "Chanter", 40, 40, &npc.EnemyTrait{AttackPower: 1, OnAggroScript: "SCR-AGGRO", Difficulty: "easy"})
	g.NPCManager.RegisterExistingNPC(enemy, "R-hook")

	inst := g.CombatController.InitiateCombat("R-hook", []*characters.Character{hero}, []*npc.NPC{enemy})
	if inst == nil || inst.State != combat.CombatStateActive {
		t.Fatalf("fight did not start: %#v", inst)
	}
	for i := 0; i < 4; i++ {
		g.CombatController.engine.NextTurn(inst)
	}
	g.CombatController.runEnemyHooksOnEnter(inst)
	if got := countRoomText(drainGameMessages(g.SendMessage()), "The chanter hits the drum once."); got != 1 {
		t.Fatalf("onAggro messages = %d, want 1", got)
	}

	_, hero2 := sharePlayer(t, facade, "user-aggro-2", "ref-aggro-2", "char-aggro-2", "Hero2", "R-hook")
	enemy2 := hookEnemy("chanter-2", "Chanter", 40, 40, &npc.EnemyTrait{AttackPower: 1, OnAggroScript: "SCR-AGGRO", Difficulty: "easy"})
	g.NPCManager.RegisterExistingNPC(enemy2, "R-hook")
	if g.CombatController.InitiateCombat("R-hook", []*characters.Character{hero2}, []*npc.NPC{enemy2}) == nil {
		t.Fatal("second fight did not start")
	}
	if got := countRoomText(drainGameMessages(g.SendMessage()), "The chanter hits the drum once."); got != 1 {
		t.Fatalf("second fight onAggro messages = %d, want 1", got)
	}
}

func TestEnemyOnAggroNilRunnerDoesNotPanic(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R-nil", nil)
	_, hero := sharePlayer(t, facade, "user-nil", "ref-nil", "char-nil", "Hero", "R-nil")
	enemy := hookEnemy("nil-rat", "Rat", 8, 8, &npc.EnemyTrait{OnAggroScript: "SCR-MISSING", Difficulty: "easy"})
	inst := g.CombatController.InitiateCombat("R-nil", []*characters.Character{hero}, []*npc.NPC{enemy})
	if inst == nil || inst.State != combat.CombatStateActive {
		t.Fatal("nil runner should still start the fight")
	}
}

func TestEnemyOnDeathOnceDoesNotDoubleRewards(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, "R-death", nil)
	storeHookScript(t, facade, "SCR-DEATH", `tales.game.msgToRoom(ctx.roomId, "The drum stops.")`)
	if _, err := facade.ItemsService().Import(&items.Item{
		Entity: &entities.Entity{ID: "tpl-hook-fang"}, Name: "Hook Fang", IsTemplate: true,
	}); err != nil {
		t.Fatal(err)
	}
	_, hero := sharePlayer(t, facade, "user-death", "ref-death", "char-death", "Hero", "R-death")
	enemy := hookEnemy("dead-rat", "Rat", 8, 8, &npc.EnemyTrait{
		AttackPower: 0, Difficulty: "easy", XPReward: 1,
		OnDeathScript:  "SCR-DEATH",
		GuaranteedLoot: []string{"tpl-hook-fang"},
	})
	g.NPCManager.RegisterExistingNPC(enemy, "R-death")
	inst := g.CombatController.InitiateCombat("R-death", []*characters.Character{hero}, []*npc.NPC{enemy})
	ref := inst.GetEnemyByID(enemy.ID)
	ref.IsAlive = false
	ref.CurrentHP = 0
	g.CombatController.engine.UpdateCombatant(inst, ref)

	_ = drainGameMessages(g.SendMessage())
	g.CombatController.flushEnemyDeathHooks(inst)
	g.CombatController.flushEnemyDeathHooks(inst)
	if got := countRoomText(drainGameMessages(g.SendMessage()), "The drum stops."); got != 1 {
		t.Fatalf("onDeath messages = %d, want 1", got)
	}

	g.CombatController.engine.EndCombat(inst, combat.CombatStateVictory)
	g.CombatController.cleanupCombatInstance(inst, combat.CombatStateVictory)
	wantXP := balance.ScaleReward(1, balance.RewardMultiplier(balance.ThreatTier(1, 1)))
	gotChar, err := facade.CharactersService().FindByID(hero.ID)
	if err != nil {
		t.Fatal(err)
	}
	if int64(gotChar.XP) != wantXP {
		t.Fatalf("xp = %d, want one scaled grant %d", gotChar.XP, wantXP)
	}
	room, err := facade.RoomsService().FindByID("R-death")
	if err != nil || room == nil || room.Items == nil || len(*room.Items) != 1 {
		t.Fatalf("room items = %#v", room)
	}
	g.CombatController.flushEnemyDeathHooks(inst)
	again, _ := facade.CharactersService().FindByID(hero.ID)
	room, _ = facade.RoomsService().FindByID("R-death")
	if int64(again.XP) != wantXP || room.Items == nil || len(*room.Items) != 1 {
		t.Fatalf("second flush changed rewards xp=%d items=%v", again.XP, room.Items)
	}
}

func TestEnemyOnDeathErrorStillGrantsOnce(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, "R-err", nil)
	storeHookScript(t, facade, "SCR-ERR", `error("nope")`)
	if _, err := facade.ItemsService().Import(&items.Item{
		Entity: &entities.Entity{ID: "tpl-hook-err"}, Name: "Err Fang", IsTemplate: true,
	}); err != nil {
		t.Fatal(err)
	}
	_, hero := sharePlayer(t, facade, "user-err", "ref-err", "char-err", "Hero", "R-err")
	enemy := hookEnemy("err-rat", "Rat", 8, 8, &npc.EnemyTrait{
		Difficulty: "easy", XPReward: 1, OnDeathScript: "SCR-ERR",
		GuaranteedLoot: []string{"tpl-hook-err"},
	})
	g.NPCManager.RegisterExistingNPC(enemy, "R-err")
	inst := g.CombatController.InitiateCombat("R-err", []*characters.Character{hero}, []*npc.NPC{enemy})
	ref := inst.GetEnemyByID(enemy.ID)
	ref.IsAlive = false
	ref.CurrentHP = 0
	g.CombatController.engine.UpdateCombatant(inst, ref)
	g.CombatController.engine.EndCombat(inst, combat.CombatStateVictory)
	g.CombatController.cleanupCombatInstance(inst, combat.CombatStateVictory)
	if inst == nil {
		t.Fatal("error script broke the instance")
	}
	wantXP := balance.ScaleReward(1, balance.RewardMultiplier(balance.ThreatTier(1, 1)))
	gotChar, err := facade.CharactersService().FindByID(hero.ID)
	if err != nil {
		t.Fatal(err)
	}
	if int64(gotChar.XP) != wantXP {
		t.Fatalf("error script xp = %d, want %d", gotChar.XP, wantXP)
	}
	room, err := facade.RoomsService().FindByID("R-err")
	if err != nil || room == nil || room.Items == nil || len(*room.Items) != 1 {
		t.Fatalf("error script items = %#v err=%v", room, err)
	}
}

func TestEnemyHookLoopDoesNotBreakCombat(t *testing.T) {
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, "R-loop", nil)
	storeHookScript(t, facade, "SCR-LOOP", `while true do end`)
	_, hero := sharePlayer(t, facade, "user-loop", "ref-loop", "char-loop", "Hero", "R-loop")
	enemy := hookEnemy("loop-rat", "Rat", 8, 8, &npc.EnemyTrait{OnAggroScript: "SCR-LOOP", Difficulty: "easy"})
	inst := g.CombatController.InitiateCombat("R-loop", []*characters.Character{hero}, []*npc.NPC{enemy})
	if inst == nil || inst.State != combat.CombatStateActive {
		t.Fatalf("loop script broke combat: %#v", inst)
	}
}

func TestEnemyOnFleeRunsOnce(t *testing.T) {
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, "R-flee", nil)
	storeHookScript(t, facade, "SCR-FLEE", `tales.game.msgToRoom(ctx.roomId, "The coward breaks.")`)
	_, hero := sharePlayer(t, facade, "user-flee", "ref-flee", "char-flee", "Hero", "R-flee")
	enemy := hookEnemy("coward", "Coward", 20, 20, &npc.EnemyTrait{
		AttackPower: 1, FleeThreshold: 1, OnFleeScript: "SCR-FLEE", Difficulty: "easy",
	})
	g.NPCManager.RegisterExistingNPC(enemy, "R-flee")
	inst := g.CombatController.InitiateCombat("R-flee", []*characters.Character{hero}, []*npc.NPC{enemy})
	for i := range inst.TurnOrder {
		if inst.TurnOrder[i].ID == enemy.ID {
			inst.CurrentTurnIdx = i
			break
		}
	}
	_ = drainGameMessages(g.SendMessage())
	current := inst.GetCurrentTurnCombatant()
	g.CombatController.resolveNPCTurn(inst, current)
	g.CombatController.resolveNPCTurn(inst, current)
	if got := countRoomText(drainGameMessages(g.SendMessage()), "The coward breaks."); got != 1 {
		t.Fatalf("onFlee messages = %d, want 1", got)
	}
	if !inst.GetEnemyByID(enemy.ID).IsAlive {
		t.Fatal("flee is not death")
	}
}

func TestWarChantHealsHurtAllyOnce(t *testing.T) {
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, "R-chant", nil)
	storeHookScript(t, facade, "SCR0603", warChantLua)
	_, hero := sharePlayer(t, facade, "user-chant", "ref-chant", "char-chant", "Hero", "R-chant")
	chanter := hookEnemy("chanter", "Orc War-Chanter", 40, 40, &npc.EnemyTrait{OnAggroScript: "SCR0603", Difficulty: "easy"})
	ally := hookEnemy("grask", "Grask", 10, 40, &npc.EnemyTrait{Difficulty: "easy"})
	g.NPCManager.RegisterExistingNPC(chanter, "R-chant")
	g.NPCManager.RegisterExistingNPC(ally, "R-chant")
	inst := g.CombatController.InitiateCombat("R-chant", []*characters.Character{hero}, []*npc.NPC{chanter, ally})
	if got := inst.GetEnemyByID("grask").CurrentHP; got != 28 {
		t.Fatalf("ally hp = %d, want 28", got)
	}
	if got := g.NPCManager.GetInstance("grask").CurrentHitPoints; got != 28 {
		t.Fatalf("synced ally hp = %d, want 28", got)
	}
	if got := inst.GetEnemyByID("chanter").CurrentHP; got != 40 {
		t.Fatalf("chanter healed itself, hp=%d", got)
	}
	g.CombatController.runEnemyHooksOnEnter(inst)
	if got := inst.GetEnemyByID("grask").CurrentHP; got != 28 {
		t.Fatalf("second enter healed again, hp=%d", got)
	}
	msgs := drainGameMessages(g.SendMessage())
	if got := countRoomText(msgs, "Grask stands a little straighter."); got != 1 {
		t.Fatalf("heal line count = %d", got)
	}
	if countRoomText(msgs, "The chanter hits the drum once.") != 1 {
		t.Fatal("missing drum line")
	}
}

func TestWarChantDoesNotClaimAHealAtFullHP(t *testing.T) {
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, "R-full", nil)
	storeHookScript(t, facade, "SCR0603", warChantLua)
	_, hero := sharePlayer(t, facade, "user-full", "ref-full", "char-full", "Hero", "R-full")
	chanter := hookEnemy("chanter-full", "Orc War-Chanter", 10, 40, &npc.EnemyTrait{OnAggroScript: "SCR0603", Difficulty: "easy"})
	ally := hookEnemy("grask-full", "Grask", 40, 40, &npc.EnemyTrait{Difficulty: "easy"})
	g.NPCManager.RegisterExistingNPC(chanter, "R-full")
	g.NPCManager.RegisterExistingNPC(ally, "R-full")
	inst := g.CombatController.InitiateCombat("R-full", []*characters.Character{hero}, []*npc.NPC{chanter, ally})
	if inst.GetEnemyByID("chanter-full").CurrentHP != 10 || inst.GetEnemyByID("grask-full").CurrentHP != 40 {
		t.Fatal("full ally or wounded chanter HP changed")
	}
	msgs := drainGameMessages(g.SendMessage())
	if countRoomText(msgs, "stands a little straighter.") != 0 {
		t.Fatal("claimed a heal when nobody in the fight was hurt")
	}
	if countRoomText(msgs, "The chanter hits the drum once.") != 1 {
		t.Fatal("missing drum line at full HP")
	}
}

func TestApplyCombatEffectBuffOnly(t *testing.T) {
	g, facade := newHookTestGame(t)
	storeTestRoom(t, facade, "R-buff", nil)
	storeHookScript(t, facade, "SCR-BUFF", `
local bad = tales.combat.applyEffect(ctx.npc.ID, "warrior_cleave")
local good = tales.combat.applyEffect(ctx.npc.ID, "warrior_battle_cry")
tales.game.msgToRoom(ctx.roomId, (bad and "bad" or "no") .. " " .. (good and "yes" or "no"))
`)
	if _, err := facade.SkillsService().Import(&skills.Skill{
		Entity: &entities.Entity{ID: "test_zero_buff"}, Name: "Zero Buff",
		Effect: skills.EffectBuff, BuffStat: "defense", BuffPercent: 0.1, Duration: 0,
	}); err != nil {
		t.Fatal(err)
	}
	_, hero := sharePlayer(t, facade, "user-buff", "ref-buff", "char-buff", "Hero", "R-buff")
	enemy := hookEnemy("buff-orc", "Orc", 40, 40, &npc.EnemyTrait{OnAggroScript: "SCR-BUFF", Difficulty: "easy"})
	g.NPCManager.RegisterExistingNPC(enemy, "R-buff")
	inst := g.CombatController.InitiateCombat("R-buff", []*characters.Character{hero}, []*npc.NPC{enemy})
	if got := countRoomText(drainGameMessages(g.SendMessage()), "no yes"); got != 1 {
		t.Fatalf("effect script line missing, count=%d", got)
	}
	ref := inst.GetEnemyByID("buff-orc")
	if !hasSkillEffect(ref, "warrior_battle_cry") {
		t.Fatalf("battle cry missing: %+v", ref.StatusEffects)
	}
	if hasSkillEffect(ref, "warrior_cleave") {
		t.Fatal("damage skill was applied")
	}
	if !g.CombatController.ApplyCombatEffect("buff-orc", "test_zero_buff") {
		t.Fatal("duration 0 buff should apply")
	}
	ref = inst.GetEnemyByID("buff-orc")
	for _, se := range ref.StatusEffects {
		if se.SkillID == "test_zero_buff" && se.Duration != 1 {
			t.Fatalf("duration 0 should become 1 round, got %d", se.Duration)
		}
	}
	dead := inst.GetEnemyByID("buff-orc")
	dead.IsAlive = false
	g.CombatController.engine.UpdateCombatant(inst, dead)
	if g.CombatController.HealCombatNPC("buff-orc", 5) != 0 {
		t.Fatal("dead enemy healed")
	}
}

func hasSkillEffect(ref *combat.CombatantRef, skillID string) bool {
	if ref == nil {
		return false
	}
	for _, se := range ref.StatusEffects {
		if se.SkillID == skillID {
			return true
		}
	}
	return false
}

func TestAttackSpeedHoldAndBeatUnchanged(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R-pace", nil)
	_, hero := sharePlayer(t, facade, "user-speed", "ref-speed", "char-speed", "Hero", "R-pace")
	enemy := hookEnemy("slow-rat", "Rat", 80, 80, &npc.EnemyTrait{AttackPower: 1, Defense: 0, Difficulty: "easy", AttackSpeed: 0.5})
	g.NPCManager.RegisterExistingNPC(enemy, "R-pace")
	inst := seedPacedCombat(t, g, hero.ID, enemy.ID, false)
	ref := inst.GetEnemyByID(enemy.ID)
	ref.AttackSpeed = 0.5
	ref.Difficulty = "easy"
	ref.AttackPower = 1
	g.CombatController.engine.UpdateCombatant(inst, ref)
	player := inst.GetPlayerByID(hero.ID)
	player.CurrentHP = 500
	player.MaxHP = 500
	g.CombatController.engine.UpdateCombatant(inst, player)

	current := inst.GetCurrentTurnCombatant()
	g.CombatController.resolveNPCTurn(inst, current)
	if got := inst.GetEnemyByID(enemy.ID).AttackActions; got != 1 {
		t.Fatalf("first slow action should swing, actions=%d", got)
	}
	if countLogResult(inst, "hold") != 0 {
		t.Fatal("first action held")
	}
	g.CombatController.resolveNPCTurn(inst, inst.GetCurrentTurnCombatant())
	if got := inst.GetEnemyByID(enemy.ID).AttackActions; got != 2 {
		t.Fatalf("second slow action should hold, actions=%d", got)
	}
	if countLogResult(inst, "hold") != 1 {
		t.Fatal("expected one hold")
	}

	fast := seedPacedCombat(t, g, hero.ID, "fast-rat", false)
	// seed registered npc-pace ids only. Build a second enemy on this instance.
	fast.Enemies[0].ID = "fast-rat"
	fast.Enemies[0].AttackSpeed = 0
	fast.Enemies[0].Difficulty = "easy"
	fast.Enemies[0].AttackActions = 0
	g.CombatController.engine.UpdateCombatant(fast, &fast.Enemies[0])
	g.CombatController.resolveNPCTurn(fast, fast.GetCurrentTurnCombatant())
	g.CombatController.resolveNPCTurn(fast, fast.GetCurrentTurnCombatant())
	if got := fast.GetEnemyByID("fast-rat").AttackActions; got != 2 {
		t.Fatalf("speed 0 should swing both actions, actions=%d", got)
	}
	if countLogResult(fast, "hold") != 0 {
		t.Fatal("speed 0 held")
	}

	budget := g.CombatController.engine.Config.BeatBudget()
	assertBeat := func(speed float64) time.Duration {
		t.Helper()
		box := seedPacedCombat(t, g, hero.ID, enemy.ID, true)
		box.Enemies[0].AttackSpeed = speed
		start := time.Now()
		g.CombatController.finishTurnBeat(box)
		delta := box.NextActionAt.Sub(start)
		if delta < budget-50*time.Millisecond || delta > budget+50*time.Millisecond {
			t.Fatalf("speed %v beat %s, budget %s", speed, delta, budget)
		}
		return delta
	}
	d0 := assertBeat(0)
	d2 := assertBeat(2)
	if d0-d2 > 50*time.Millisecond || d2-d0 > 50*time.Millisecond {
		t.Fatalf("speed changed the beat: 0=%s 2=%s", d0, d2)
	}
}
