package service

import (
	"errors"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
)

func TestApplyCreateRaceCopiesRaceAndGold(t *testing.T) {
	cutpurse := characters.PresetByID("tpl-cutpurse")
	cutpurse.Gold = 40
	ch := &characters.Character{}
	if err := applyCreateRace(cutpurse, ch, "elf"); err != nil {
		t.Fatal(err)
	}
	if ch.Race.ID != "elf" || ch.Gold != 40 {
		t.Fatalf("elf race=%s gold=%d", ch.Race.ID, ch.Gold)
	}

	ch = &characters.Character{}
	if err := applyCreateRace(cutpurse, ch, "human"); err != nil {
		t.Fatal(err)
	}
	if ch.Race.ID != "human" || ch.Gold != 55 {
		t.Fatalf("human race=%s gold=%d", ch.Race.ID, ch.Gold)
	}

	ch = &characters.Character{}
	if err := applyCreateRace(cutpurse, ch, "elve"); err != nil {
		t.Fatal(err)
	}
	if ch.Race.ID != "elf" || ch.Gold != 40 {
		t.Fatalf("elve stored as %s gold=%d", ch.Race.ID, ch.Gold)
	}

	rig := characters.PresetByID("tpl-rigger")
	rig.Gold = 7
	ch = &characters.Character{}
	if err := applyCreateRace(rig, ch, "construct"); err != nil {
		t.Fatal(err)
	}
	if ch.Race.ID != "construct" || ch.Gold != 7 {
		t.Fatalf("construct race=%s gold=%d", ch.Race.ID, ch.Gold)
	}

	if err := applyCreateRace(cutpurse, ch, ""); !errors.Is(err, characters.ErrRaceRequired) {
		t.Fatalf("missing race %v", err)
	}
	if err := applyCreateRace(characters.PresetByID("tpl-sentinel"), ch, "elf"); !errors.Is(err, characters.ErrRaceNotAllowed) {
		t.Fatalf("sentinel elf %v", err)
	}
	if err := applyCreateRace(characters.PresetByID("tpl-sentinel"), ch, "construct"); !errors.Is(err, characters.ErrRaceNotAllowed) {
		t.Fatalf("sentinel construct %v", err)
	}
	if err := applyCreateRace(characters.PresetByID("tpl-runecaster"), ch, "dwarf"); !errors.Is(err, characters.ErrRaceNotAllowed) {
		t.Fatalf("runecaster dwarf %v", err)
	}
	if err := applyCreateRace(characters.PresetByID("tpl-ward"), ch, "elf"); !errors.Is(err, characters.ErrRaceNotAllowed) {
		t.Fatalf("ward elf %v", err)
	}
	if err := applyCreateRace(characters.PresetByID("tpl-rigger"), ch, "human"); !errors.Is(err, characters.ErrRaceNotAllowed) {
		t.Fatalf("rigger human %v", err)
	}
	if err := applyCreateRace(characters.PresetByID("tpl-rigger"), ch, "elf"); !errors.Is(err, characters.ErrRaceNotAllowed) {
		t.Fatalf("rigger elf %v", err)
	}
}
