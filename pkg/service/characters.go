package service

import (
	"errors"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	r "github.com/talesmud/talesmud/pkg/repository"
)

//--- Interface Definitions

//CharactersService delives logical functions on top of the charactersheets Repo
type CharactersService interface {
	r.CharactersRepository

	IsCharacterNameTaken(name string) bool
	GetCharacterTemplates() []*characters.Character
}

//--- Implementations

type charactersService struct {
	r.CharactersRepository
}

//NewCharactersService creates a nwe item service
func NewCharactersService(charactersRepo r.CharactersRepository) CharactersService {
	return &charactersService{
		charactersRepo,
	}
}

//IsCharacterNameTaken ...
func (srv *charactersService) IsCharacterNameTaken(name string) bool {
	// check if charactername already exists
	if chars, err := srv.FindByName(name); err == nil {
		if len(chars) > 0 {
			return true
		}
	}
	return false
}

//Store ...
func (srv *charactersService) Store(character *characters.Character) (*characters.Character, error) {

	// check if charactername already exists
	if srv.IsCharacterNameTaken(character.Name) {
		return nil, errors.New("character name already taken")
	}
	return srv.Store(character)
}

func (srv *charactersService) GetCharacterTemplates() []characters.Character {
	return characters.CharacterTemplates
}
