package validation

import "testing"

func TestResultAddCountsErrorsAndWarnings(t *testing.T) {
	var result Result

	result.Add(Issue{Severity: SeverityError, Code: "missing_room", EntityType: "room", EntityID: "room-a"})
	result.Add(Issue{Severity: SeverityWarning, Code: "empty_action", EntityType: "room", EntityID: "room-a"})

	if result.Valid {
		t.Fatalf("result.Valid = true, want false when errors exist")
	}
	if result.Errors != 1 {
		t.Fatalf("Errors = %d, want 1", result.Errors)
	}
	if result.Warnings != 1 {
		t.Fatalf("Warnings = %d, want 1", result.Warnings)
	}
	if len(result.Issues) != 2 {
		t.Fatalf("Issues length = %d, want 2", len(result.Issues))
	}
}

func TestSnapshotHasRoomAndScriptLookupHelpers(t *testing.T) {
	snapshot := WorldSnapshot{
		RoomIDs:   map[string]bool{"room-a": true},
		ScriptIDs: map[string]bool{"script-a": true},
	}

	if !snapshot.HasRoom("room-a") {
		t.Fatalf("HasRoom(room-a) = false, want true")
	}
	if snapshot.HasRoom("missing-room") {
		t.Fatalf("HasRoom(missing-room) = true, want false")
	}
	if !snapshot.HasScript("script-a") {
		t.Fatalf("HasScript(script-a) = false, want true")
	}
}
