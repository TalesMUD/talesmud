package service

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	r "github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/server/dto"
)

//--- Interface Definitions

//CharactersService delives logical functions on top of the charactersheets Repo
type CharactersService interface {
	r.CharactersRepository

	IsCharacterNameTaken(name string) bool
	GetCharacterTemplates() []characters.CharacterTemplate

	CreateNewCharacter(dto *dto.CreateCharacterDTO) (*characters.Character, error)
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
func (srv *charactersService) CreateNewCharacter(dto *dto.CreateCharacterDTO) (*characters.Character, error) {

	// check if charactername already exists
	if srv.IsCharacterNameTaken(dto.Name) {
		return nil, errors.New("character name already taken")
	}

	// get template
	if template, err := srv.GetCharacterTemplate(dto.TemplateID); err != nil {
		return nil, fmt.Errorf("Could not create new character %v", dto)
	} else {

		character := characterFromTemplate(template)
		//character.Entity = entities.NewEntity()
		character.Name = dto.Name
		character.Description = dto.Description
		character.BelongsUserID = dto.UserID

		if createdCharacter, err := srv.Store(character); err == nil {

			log.Info("Created new character based on template")

			return createdCharacter, nil
		}
	}

	return nil, errors.New("Could not create new character")
}

func (srv *charactersService) GetCharacterTemplate(templateID int32) (characters.CharacterTemplate, error) {
	templates := srv.GetCharacterTemplates()

	for _, template := range templates {
		if template.TemplateID == templateID {
			return template, nil
		}
	}

	return characters.CharacterTemplate{}, fmt.Errorf("Could not find templateID %v", templateID)
}

func characterFromTemplate(template characters.CharacterTemplate) *characters.Character {

	return &characters.Character{
		Race:             template.Race,
		Class:            template.Class,
		CurrentHitPoints: template.CurrentHitPoints,
		MaxHitPoints:     template.MaxHitPoints,
		XP:               0,
		Level:            1,
		Attributes:       template.Attributes,
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
	return srv.CharactersRepository.Store(character)
}

func (srv *charactersService) GetCharacterTemplates() []characters.CharacterTemplate {
	return characters.CharacterTemplates
}
