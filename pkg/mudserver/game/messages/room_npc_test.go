package messages

import (
	"encoding/json"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

func TestRoomNPCPayloadCopiesAggroOnSight(t *testing.T) {
	out := roomNPCPayload(util.RoomNPC{
		ID: "wolf", Name: "Wolf", IsEnemy: true, Level: 12, AggroOnSight: true,
	}, &characters.Character{Level: 12})
	if !out.AggroOnSight {
		t.Fatal("aggroOnSight was dropped from the room payload")
	}
	quiet := roomNPCPayload(util.RoomNPC{ID: "rat", Name: "Rat", IsEnemy: true}, nil)
	if quiet.AggroOnSight {
		t.Fatal("passive NPC payload marked aggroOnSight")
	}
	for _, tc := range []struct {
		name string
		npc  RoomNPC
		want bool
	}{{"aggressive", out, true}, {"passive", quiet, false}} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(tc.npc)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]interface{}
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			flag, present := fields["aggroOnSight"]
			if present != tc.want || (tc.want && flag != true) {
				t.Fatalf("aggroOnSight JSON field = %v, present = %v", flag, present)
			}
		})
	}
}
