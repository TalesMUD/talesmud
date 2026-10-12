package commands_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestSelectCharacterCombatSnapshot(t *testing.T) {
	for _, scenario := range []string{"live", "stale-flag", "not-in-combat", "ended", "fled"} {
		t.Run(scenario, func(t *testing.T) {
			g, facade, user, fighter, _ := setupSwitchFight(t)
			if scenario != "stale-flag" {
				startRatFight(t, g, facade, fighter)
			}
			inst := g.GetCombatEngine().GetCombatInstance(fighter.ID)
			if inst != nil {
				inst.Round = 3
				inst.DecisionDeadline = time.Now().Add(4 * time.Second)
				inst.NextActionAt = inst.DecisionDeadline
				player := inst.GetPlayerByID(fighter.ID)
				player.AutoAttackTargetID = "rat-sw"
				player.QueuedAction = combat.CombatAction("defend")
				player.SkillCooldowns = map[string]int{"brace": 2}
				if scenario == "ended" {
					inst.State = combat.CombatStateVictory
				}
				if scenario == "fled" {
					player.HasFled = true
				}
			}
			fighter.InCombat = scenario != "not-in-combat"
			if err := facade.CharactersService().Update(fighter.ID, fighter); err != nil {
				t.Fatal(err)
			}
			before, _ := json.Marshal(inst)
			_ = drainSocialMessages(g.SendMessage())
			(&commands.SelectCharacterCommand{}).Execute(g, &messages.Message{FromUser: user, Character: fighter, Data: "sc " + fighter.Name})
			var snapshots []*messages.CombatStatusMessage
			for _, out := range drainSocialMessages(g.SendMessage()) {
				if _, ok := out.(*messages.CombatStartMessage); ok {
					t.Fatal("selection started combat again")
				}
				if msg, ok := out.(*messages.CombatStatusMessage); ok {
					snapshots = append(snapshots, msg)
				}
			}
			after, _ := json.Marshal(inst)
			if string(before) != string(after) {
				t.Fatal("selection mutated the fight")
			}
			if scenario != "live" {
				if len(snapshots) != 0 {
					t.Fatalf("unexpected snapshot in %s", scenario)
				}
				return
			}
			if len(snapshots) != 1 {
				t.Fatalf("got %d snapshots", len(snapshots))
			}
			msg := snapshots[0]
			actor := inst.GetCurrentTurnCombatant()
			if len(msg.Combatants) != 2 || msg.Round != 3 || msg.ActorID != actor.ID || msg.ActorName != actor.Name || msg.DeadlineMs != inst.DecisionDeadline.UnixMilli() || msg.TargetID != "rat-sw" {
				t.Fatalf("incomplete snapshot: %+v", msg)
			}
			if msg.SkillCooldowns["brace"] != 2 || msg.QueuedAction != "defend" || msg.GetAudienceID() != user.ID || msg.Message != "" {
				t.Fatalf("queue/audience/prose: %+v", msg)
			}
		})
	}
}
