package service

import (
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
	return s.repo.Store(party)
}

func (s *partiesService) AddCharacterToParty(party *e.Party, character *characters.Character) error {
	return nil
}
