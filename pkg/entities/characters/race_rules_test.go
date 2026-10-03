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
	if RaceAllowed("warrior", "construct") || RaceAllowed("tpl-fenwatch", "construct") {
		t.Fatal("construct allowed off rigger")
	}
	if RaceAllowed("wizard", "dwarf") || RaceAllowed("tpl-runehand", "dwarf") || RaceAllowed("mage", "dwarf") {
		t.Fatal("dwarf allowed on rune hand")
	}
	for _, classID := range []string{"warrior", "tpl-fenwatch", "hitch", "tpl-hitch", "rigger", "tpl-rigger"} {
		if RaceAllowed(classID, "elf") || RaceAllowed(classID, "elve") {
			t.Fatalf("elf allowed on %s", classID)
		}
	}
	if !RaceAllowed("rogue", "elf") || !RaceAllowed("tpl-alley", "elve") {
		t.Fatal("alley should allow elf")
	}
	if !RaceAllowed("rigger", "construct") || !RaceAllowed("tpl-rigger", "construct") {
		t.Fatal("rigger should allow construct")
	}
	if RaceAllowed("hitch", "construct") || RaceAllowed("tpl-fenwatch", "elf") {
		t.Fatal("unexpected allow")
	}

	for _, preset := range SystemCharacterTemplatePresets() {
		if preset == nil || !RaceAllowed(preset.Class.ID, preset.Race.ID) {
			t.Fatalf("preset %s default race illegal", preset.Name)
		}
	}
	runeHand := PresetByID("tpl-runehand")
	if runeHand == nil || runeHand.Race.ID != "human" {
		t.Fatalf("rune hand default %+v", runeHand)
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
	alley := PresetByID("tpl-alley")
	race, err := ResolveCreateRace(alley, "elve")
	if err != nil || race.ID != "elf" {
		t.Fatalf("store elve as elf: %+v %v", race, err)
	}
	if _, err := ResolveCreateRace(alley, ""); err != ErrRaceRequired {
		t.Fatalf("missing race %v", err)
	}
	fen := PresetByID("tpl-fenwatch")
	if _, err := ResolveCreateRace(fen, "elf"); err != ErrRaceNotAllowed {
		t.Fatalf("fenwatch elf %v", err)
	}
	if _, err := ResolveCreateRace(fen, "construct"); err != ErrRaceNotAllowed {
		t.Fatalf("fenwatch construct %v", err)
	}
	if _, err := ResolveCreateRace(PresetByID("tpl-runehand"), "dwarf"); err != ErrRaceNotAllowed {
		t.Fatalf("rune hand dwarf %v", err)
	}
	if _, err := ResolveCreateRace(PresetByID("tpl-hitch"), "elf"); err != ErrRaceNotAllowed {
		t.Fatalf("hitch elf %v", err)
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
	got, err = ResolveCreateRace(PresetByID("tpl-fenwatch"), "dwarf")
	if err != nil || got.ID != "dwarf" {
		t.Fatalf("fenwatch dwarf %+v %v", got, err)
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
	if _, err := ResolveGuestPick("tpl-fenwatch", "elf"); err != ErrRaceNotAllowed {
		t.Fatalf("illegal fenwatch %v", err)
	}
	if _, err := ResolveGuestPick("nope", "human"); err != ErrRaceNotAllowed {
		t.Fatalf("unknown template %v", err)
	}
	pick, err = ResolveGuestPick("tpl-alley", "elve")
	if err != nil || pick.Race.ID != "elf" {
		t.Fatalf("guest elve alias %+v %v", pick.Race, err)
	}
	race := RandomAllowedRace("rigger", func(n int) int { return 0 })
	if race.ID != "construct" {
		t.Fatalf("random rigger %s", race.ID)
	}
	race = RandomAllowedRace("warrior", func(n int) int { return 0 })
	if race.ID != "human" {
		t.Fatalf("random fenwatch %s", race.ID)
	}
	for _, id := range AllowedRaceIDs("wizard") {
		if id == "construct" || id == "dwarf" {
			t.Fatalf("rune hand list %v", AllowedRaceIDs("wizard"))
		}
	}
}
