package service

import (
	"errors"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// PartiesService delives logical functions on top of the charactersheets Repo
type PartiesService interface {
	GetPartyByID(id string) (*e.Party, error)
	GetParties() ([]*e.Party, error)
	CreateParty(createParty *CreatePartyDTO) (*e.Party, error)
	UpdateParty(id string, party *entities.Party) error
	DeletePartyByID(id string) error
	FindAll() ([]*e.Party, error)
	Store(party *e.Party) (*e.Party, error)

	AddCharacterToParty(party *e.Party, character *characters.Character) error
	RemoveCharacterFromParty(party *e.Party, characterID string) error
	FindPartyForCharacter(characterID string) (*e.Party, error)
}

type partiesService struct {
	repo r.PartiesRepository
}

// NewPartiesService creates a nwe item service
func NewPartiesService(partiesrepo r.PartiesRepository) PartiesService {
	return &partiesService{
		repo: partiesrepo,
	}
}

func (s *partiesService) GetPartyByID(id string) (*e.Party, error) {
	return s.repo.FindByID(id)
}
func (s *partiesService) DeletePartyByID(id string) error {
	return s.repo.Delete(id)
}

func (s *partiesService) UpdateParty(id string, party *e.Party) error {
	return s.repo.Update(id, party)
}

func (s *partiesService) CreateParty(createParty *CreatePartyDTO) (*e.Party, error) {

	var party entities.Party
	party.Name = createParty.Name
	party.Created = time.Now()
	party.Characters = createParty.Characters

	return s.repo.Store(&party)
}

func (s *partiesService) GetParties() ([]*e.Party, error) {
	return s.repo.FindAll()
}

func (s *partiesService) FindAll() ([]*e.Party, error) {
	return s.repo.FindAll()
}

func (s *partiesService) Store(party *e.Party) (*e.Party, error) {
	if party.Created.IsZero() {
		party.Created = time.Now()
	}
	return s.repo.Store(party)
}

func (s *partiesService) AddCharacterToParty(party *e.Party, character *characters.Character) error {
	if party == nil {
		return errors.New("party is nil")
	}
	if character == nil || character.ID == "" {
		return errors.New("character is nil")
	}

	for _, memberID := range party.Characters {
		if memberID == character.ID {
			return s.repo.Update(party.ID, party)
		}
	}

	party.Characters = append(party.Characters, character.ID)
	return s.repo.Update(party.ID, party)
}

func (s *partiesService) RemoveCharacterFromParty(party *e.Party, characterID string) error {
	if party == nil {
		return errors.New("party is nil")
	}
	if characterID == "" {
		return errors.New("character id is empty")
	}

	members := make([]string, 0, len(party.Characters))
	for _, memberID := range party.Characters {
		if memberID != characterID {
			members = append(members, memberID)
		}
	}
	party.Characters = members

	if len(party.Characters) == 0 {
		return s.repo.Delete(party.ID)
	}
	return s.repo.Update(party.ID, party)
}

func (s *partiesService) FindPartyForCharacter(characterID string) (*e.Party, error) {
	if characterID == "" {
		return nil, errors.New("character id is empty")
	}

	parties, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	for _, party := range parties {
		for _, memberID := range party.Characters {
			if memberID == characterID {
				return party, nil
			}
		}
	}
	return nil, errors.New("party not found")
}
