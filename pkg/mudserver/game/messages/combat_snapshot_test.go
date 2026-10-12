package messages

import (
	"encoding/json"
	"testing"
)

func TestCombatStatusSnapshotJSON(t *testing.T) {
	msg := NewCombatStatusMessage("user", "", 3, CombatQueueState{SkillCooldowns: map[string]int{}})
	msg.ActorID, msg.ActorName, msg.DeadlineMs, msg.TargetID = "hero", "Hero", 1234, "rat"
	msg.Combatants = []CombatantView{{ID: "hero", Type: "player", Name: "Hero", HP: 10, MaxHP: 20, IsAlive: true}, {ID: "rat", Type: "npc", Name: "Rat", HP: 9, MaxHP: 9, IsAlive: true}}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]interface{}
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	if wire["type"] != "combatStatus" || wire["actorId"] != "hero" || wire["actorName"] != "Hero" || wire["deadlineMs"] != float64(1234) || wire["targetId"] != "rat" || wire["round"] != float64(3) || len(wire["combatants"].([]interface{})) != 2 {
		t.Fatalf("bad snapshot JSON: %s", data)
	}
	data, err = json.Marshal(NewCombatStatusMessage("user", "", 0, CombatQueueState{}))
	if err != nil {
		t.Fatal(err)
	}
	wire = map[string]interface{}{}
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"actorId", "actorName", "deadlineMs", "targetId", "combatants", "round"} {
		if _, ok := wire[key]; ok {
			t.Fatalf("empty %s was emitted: %s", key, data)
		}
	}
}
