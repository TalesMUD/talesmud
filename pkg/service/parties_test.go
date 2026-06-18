package service

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/repository"
)

func newPartyTestFacade(t *testing.T) Facade {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return NewFacade(repository.NewSQLiteFactory(client), nil)
}

func TestPartiesServiceAddsCharacterOnlyOnce(t *testing.T) {
	facade := newPartyTestFacade(t)
	party, err := facade.PartiesService().CreateParty(&CreatePartyDTO{Name: "Delvers"})
	if err != nil {
		t.Fatalf("create party: %v", err)
	}
	character := &characters.Character{Entity: &entities.Entity{ID: "char-1"}, Name: "Aster"}

	if err := facade.PartiesService().AddCharacterToParty(party, character); err != nil {
		t.Fatalf("first add character: %v", err)
	}
	if err := facade.PartiesService().AddCharacterToParty(party, character); err != nil {
		t.Fatalf("second add character: %v", err)
	}

	stored, err := facade.PartiesService().GetPartyByID(party.ID)
	if err != nil {
		t.Fatalf("get party: %v", err)
	}
	if got := len(stored.Characters); got != 1 {
		t.Fatalf("expected one party member after duplicate add, got %d", got)
	}
	if stored.Characters[0] != "char-1" {
		t.Fatalf("expected char-1 member, got %#v", stored.Characters)
	}
}

func TestPartiesServiceFindsAndRemovesCharacterParty(t *testing.T) {
	facade := newPartyTestFacade(t)
	party, err := facade.PartiesService().CreateParty(&CreatePartyDTO{
		Name:       "Night Watch",
		Characters: []string{"char-1", "char-2"},
	})
	if err != nil {
		t.Fatalf("create party: %v", err)
	}

	found, err := facade.PartiesService().FindPartyForCharacter("char-2")
	if err != nil {
		t.Fatalf("find party for character: %v", err)
	}
	if found.ID != party.ID {
		t.Fatalf("expected party %s, got %s", party.ID, found.ID)
	}

	if err := facade.PartiesService().RemoveCharacterFromParty(party, "char-2"); err != nil {
		t.Fatalf("remove character from party: %v", err)
	}
	stored, err := facade.PartiesService().GetPartyByID(party.ID)
	if err != nil {
		t.Fatalf("get party: %v", err)
	}
	if got := len(stored.Characters); got != 1 {
		t.Fatalf("expected one party member after removal, got %d", got)
	}
	if stored.Characters[0] != "char-1" {
		t.Fatalf("expected char-1 to remain, got %#v", stored.Characters)
	}
}
