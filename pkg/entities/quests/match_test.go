package quests

import "testing"

func TestKillTargetMatchesTemplateID(t *testing.T) {
	obj := Objective{Type: ObjectiveKill, TargetID: "ENM0008", TargetName: "Sewer Rat"}
	if !KillTargetMatches(obj, "ENM0008", "uuid-1", "Sewer Rat") {
		t.Fatal("expected template id match")
	}
}

func TestKillTargetMatchesCloneInstanceID(t *testing.T) {
	obj := Objective{Type: ObjectiveKill, TargetID: "ENM0008"}
	if !KillTargetMatches(obj, "", "ENM0008~R0215~abcd", "Sewer Rat") {
		t.Fatal("expected clone id prefix to match template")
	}
}

func TestKillTargetMatchesDisplayNameWhenIDsDiffer(t *testing.T) {
	obj := Objective{Type: ObjectiveKill, TargetID: "ENM0008", TargetName: "Sewer Rat"}
	if !KillTargetMatches(obj, "", "a1b2c3d4", "Sewer Rat") {
		t.Fatal("expected display name match against targetName")
	}
	if KillTargetMatches(obj, "", "a1b2c3d4", "Catacomb Rat") {
		t.Fatal("catacomb rat must not count for sewer rat objective")
	}
}

func TestRoomIDMatchesPrivateCellarClone(t *testing.T) {
	if !RoomIDMatches("R0215", "R0215") {
		t.Fatal("template room should match")
	}
	if !RoomIDMatches("R0215", "R0215~a1b2c3d4") {
		t.Fatal("private cellar clone should match template room target")
	}
	if RoomIDMatches("R0215", "R0211~a1b2c3d4") {
		t.Fatal("different room clone must not match")
	}
}
