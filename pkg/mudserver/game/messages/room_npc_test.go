package messages

import (
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
}
