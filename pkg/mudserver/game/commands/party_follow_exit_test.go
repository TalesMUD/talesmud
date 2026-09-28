package commands

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func TestInstanceMarkedExitsDoNotPullFollowers(t *testing.T) {
	ex := rooms.Exit{Name: "north", Type: rooms.RoomExitTypeDirection, Target: "clone-1", Instance: true}
	if partyFollowAllowed(ex, "clone-1") {
		t.Fatal("an instance exit must not pull followers")
	}
	plain := rooms.Exit{Name: "north", Type: rooms.RoomExitTypeDirection, Target: "R0002"}
	if !partyFollowAllowed(plain, "R0002") {
		t.Fatal("a normal direction exit should still pull followers")
	}
}
