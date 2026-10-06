package characters

import "testing"

func TestCanonicalRaceIDElveAlias(t *testing.T) {
	if got := CanonicalRaceID("elve"); got != "elf" {
		t.Fatalf("elve -> %q", got)
	}
	if got := CanonicalRaceID("elves"); got != "elf" {
		t.Fatalf("elves -> %q", got)
	}
	race, ok := RaceByID("elve")
	if !ok || race.ID != "elf" || race.Name != "Elf" {
		t.Fatalf("alias race %+v ok=%v", race, ok)
	}
	if RaceElf.ID != "elf" || RaceElve.ID != "elf" || RaceConstruct.ID != "construct" {
		t.Fatalf("ids elf=%s elve=%s construct=%s", RaceElf.ID, RaceElve.ID, RaceConstruct.ID)
	}
}

func TestAllowListAndPresetDefaults(t *testing.T) {
	if RaceAllowed("warrior", "construct") || RaceAllowed("tpl-sentinel", "construct") {
		t.Fatal("construct allowed off rigger")
	}
	if RaceAllowed("wizard", "dwarf") || RaceAllowed("tpl-runecaster", "dwarf") || RaceAllowed("mage", "dwarf") {
		t.Fatal("dwarf allowed on runecaster")
	}
	for _, classID := range []string{"warrior", "tpl-sentinel", "hitch", "tpl-hitch", "rigger", "tpl-rigger"} {
		if RaceAllowed(classID, "elf") || RaceAllowed(classID, "elve") {
			t.Fatalf("elf allowed on %s", classID)
		}
	}
	if !RaceAllowed("rogue", "elf") || !RaceAllowed("tpl-cutpurse", "elve") {
		t.Fatal("cutpurse should allow elf")
	}
	if !RaceAllowed("rigger", "construct") || !RaceAllowed("tpl-rigger", "construct") {
		t.Fatal("rigger should allow construct")
	}
	if RaceAllowed("hitch", "construct") || RaceAllowed("tpl-sentinel", "elf") {
		t.Fatal("unexpected allow")
	}

	for _, preset := range SystemCharacterTemplatePresets() {
		if preset == nil || !RaceAllowed(preset.Class.ID, preset.Race.ID) {
			t.Fatalf("preset %s default race illegal", preset.Name)
		}
	}
	runecaster := PresetByID("tpl-runecaster")
	if runecaster == nil || runecaster.Race.ID != "human" {
		t.Fatalf("runecaster default %+v", runecaster)
	}
	rigger := PresetByID("tpl-rigger")
	if rigger == nil || rigger.Class.ID != "rigger" || rigger.Race.ID != "construct" || rigger.Name != "Rigger" {
		t.Fatalf("rigger preset %+v", rigger)
	}
	if rigger.Description == "" || rigger.Class.Name != "Rigger" {
		t.Fatalf("rigger blurb/name %+v", rigger.Class)
	}
	if rigger.MaxHitPoints != 25 {
		t.Fatalf("rigger template hp %d", rigger.MaxHitPoints)
	}
}

func TestStartingGoldAndCreateRace(t *testing.T) {
	if StartingGold(40, "human") != 55 || StartingGold(0, "human") != HumanStartingGoldBonus {
		t.Fatal("human gold")
	}
	for _, race := range []string{"dwarf", "elf", "elve", "construct", "orc"} {
		if StartingGold(40, race) != 40 {
			t.Fatalf("%s gold changed", race)
		}
	}
	cutpurse := PresetByID("tpl-cutpurse")
	race, err := ResolveCreateRace(cutpurse, "elve")
	if err != nil || race.ID != "elf" {
		t.Fatalf("store elve as elf: %+v %v", race, err)
	}
	if _, err := ResolveCreateRace(cutpurse, ""); err != ErrRaceRequired {
		t.Fatalf("missing race %v", err)
	}
	fen := PresetByID("tpl-sentinel")
	if _, err := ResolveCreateRace(fen, "elf"); err != ErrRaceNotAllowed {
		t.Fatalf("sentinel elf %v", err)
	}
	if _, err := ResolveCreateRace(fen, "construct"); err != ErrRaceNotAllowed {
		t.Fatalf("sentinel construct %v", err)
	}
	if _, err := ResolveCreateRace(PresetByID("tpl-runecaster"), "dwarf"); err != ErrRaceNotAllowed {
		t.Fatalf("runecaster dwarf %v", err)
	}
	if _, err := ResolveCreateRace(PresetByID("tpl-ward"), "elf"); err != ErrRaceNotAllowed {
		t.Fatalf("ward elf %v", err)
	}
	if !RaceAllowed("hitch", "human") || !RaceAllowed("tpl-hitch", "dwarf") || !RaceAllowed("ward", "dwarf") || !RaceAllowed("tpl-ward", "human") {
		t.Fatal("ward races")
	}
	gotRace, err := ResolveCreateRace(PresetByID("tpl-ward"), "dwarf")
	if err != nil || gotRace.ID != "dwarf" {
		t.Fatalf("ward dwarf %+v %v", gotRace, err)
	}
	if _, err := ResolveCreateRace(PresetByID("tpl-rigger"), "human"); err != ErrRaceNotAllowed {
		t.Fatalf("rigger human %v", err)
	}
	if _, err := ResolveCreateRace(PresetByID("tpl-rigger"), "elf"); err != ErrRaceNotAllowed {
		t.Fatalf("rigger elf %v", err)
	}
	got, err := ResolveCreateRace(PresetByID("tpl-rigger"), "construct")
	if err != nil || got.ID != "construct" {
		t.Fatalf("rigger construct %+v %v", got, err)
	}
	// A legal client race replaces the template race.
	got, err = ResolveCreateRace(PresetByID("tpl-sentinel"), "dwarf")
	if err != nil || got.ID != "dwarf" {
		t.Fatalf("sentinel dwarf %+v %v", got, err)
	}
}

func TestGuestPick(t *testing.T) {
	pick, err := ResolveGuestPick("", "")
	if err != nil || !pick.Random {
		t.Fatalf("empty pick %+v %v", pick, err)
	}
	if _, err := ResolveGuestPick("tpl-rigger", ""); err != ErrGuestPickIncomplete {
		t.Fatalf("template only %v", err)
	}
	if _, err := ResolveGuestPick("", "construct"); err != ErrGuestPickIncomplete {
		t.Fatalf("race only %v", err)
	}
	pick, err = ResolveGuestPick("tpl-rigger", "construct")
	if err != nil || pick.Random || pick.Race.ID != "construct" || pick.Template.Class.ID != "rigger" {
		t.Fatalf("legal rigger %+v %v", pick, err)
	}
	if _, err := ResolveGuestPick("tpl-rigger", "human"); err != ErrRaceNotAllowed {
		t.Fatalf("illegal rigger %v", err)
	}
	if _, err := ResolveGuestPick("tpl-sentinel", "elf"); err != ErrRaceNotAllowed {
		t.Fatalf("illegal sentinel %v", err)
	}
	if _, err := ResolveGuestPick("nope", "human"); err != ErrRaceNotAllowed {
		t.Fatalf("unknown template %v", err)
	}
	pick, err = ResolveGuestPick("tpl-cutpurse", "elve")
	if err != nil || pick.Race.ID != "elf" {
		t.Fatalf("guest elve alias %+v %v", pick.Race, err)
	}
	race := RandomAllowedRace("rigger", func(n int) int { return 0 })
	if race.ID != "construct" {
		t.Fatalf("random rigger %s", race.ID)
	}
	race = RandomAllowedRace("warrior", func(n int) int { return 0 })
	if race.ID != "human" {
		t.Fatalf("random sentinel %s", race.ID)
	}
	for _, id := range AllowedRaceIDs("wizard") {
		if id == "construct" || id == "dwarf" {
			t.Fatalf("runecaster list %v", AllowedRaceIDs("wizard"))
		}
	}
}
