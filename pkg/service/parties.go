package service

import (
	"errors"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// PartiesService delivers logical functions on top of the parties repository.
type PartiesService interface {
	GetPartyByID(id string) (*entities.Party, error)
	GetParties() ([]*entities.Party, error)
	CreateParty(createParty *CreatePartyDTO) (*entities.Party, error)
	UpdateParty(id string, party *entities.Party) error
	DeletePartyByID(id string) error
	FindAll() ([]*entities.Party, error)
	Store(party *entities.Party) (*entities.Party, error)

	FindByCharacterID(characterID string) (*entities.Party, error)
	FindPartyForCharacter(characterID string) (*entities.Party, error)
	AddCharacterToParty(party *entities.Party, character *characters.Character) error
	RemoveCharacterFromParty(party *entities.Party, characterID string) error
}

type partiesService struct {
	repo r.PartiesRepository
}

// NewPartiesService creates a new parties service.
func NewPartiesService(partiesrepo r.PartiesRepository) PartiesService {
	return &partiesService{
		repo: partiesrepo,
	}
}

func (s *partiesService) GetPartyByID(id string) (*entities.Party, error) {
	return s.repo.FindByID(id)
}

func (s *partiesService) DeletePartyByID(id string) error {
	return s.repo.Delete(id)
}

func (s *partiesService) UpdateParty(id string, party *entities.Party) error {
	return s.repo.Update(id, party)
}

func (s *partiesService) CreateParty(createParty *CreatePartyDTO) (*entities.Party, error) {
	if createParty == nil {
		return nil, errors.New("party payload is nil")
	}

	party := &entities.Party{
		Name:       createParty.Name,
		Created:    time.Now(),
		Characters: createParty.Characters,
	}
	return s.repo.Store(party)
}

func (s *partiesService) GetParties() ([]*entities.Party, error) {
	return s.repo.FindAll()
}

func (s *partiesService) FindAll() ([]*entities.Party, error) {
	return s.repo.FindAll()
}

func (s *partiesService) Store(party *entities.Party) (*entities.Party, error) {
	if party == nil {
		return nil, errors.New("party is nil")
	}
	if party.Created.IsZero() {
		party.Created = time.Now()
	}
	return s.repo.Store(party)
}

func (s *partiesService) FindByCharacterID(characterID string) (*entities.Party, error) {
	return s.findPartyByCharacterID(characterID)
}

func (s *partiesService) FindPartyForCharacter(characterID string) (*entities.Party, error) {
	return s.findPartyByCharacterID(characterID)
}

func (s *partiesService) findPartyByCharacterID(characterID string) (*entities.Party, error) {
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
	return nil, nil
}

func (s *partiesService) AddCharacterToParty(party *entities.Party, character *characters.Character) error {
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

func (s *partiesService) RemoveCharacterFromParty(party *entities.Party, characterID string) error {
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
