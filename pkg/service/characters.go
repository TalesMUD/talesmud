package service

import (
	"errors"
	"fmt"
	"sync"

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
	GetCharacterTemplates() []*characters.CharacterTemplate

	CreateNewCharacter(dto *dto.CreateCharacterDTO) (*characters.Character, error)
	// Modify loads the character, applies fn, and saves under a per-character
	// lock so concurrent script flags and pickups cannot drop inventory.
	Modify(id string, fn func(*characters.Character) error) error
}

//--- Implementations

type charactersService struct {
	r.CharactersRepository
	templatesRepo r.CharacterTemplatesRepository
	settings      ServerSettingsService
	rooms         RoomsService
	locks         sync.Map
}

//NewCharactersService creates a new item service
func NewCharactersService(charactersRepo r.CharactersRepository, templatesRepo r.CharacterTemplatesRepository, settings ServerSettingsService, rooms RoomsService) CharactersService {
	return &charactersService{
		CharactersRepository: charactersRepo,
		templatesRepo:        templatesRepo,
		settings:             settings,
		rooms:                rooms,
	}
}
func (srv *charactersService) CreateNewCharacter(dto *dto.CreateCharacterDTO) (*characters.Character, error) {

	// check if charactername already exists
	if srv.IsCharacterNameTaken(dto.Name) {
		return nil, errors.New("character name already taken")
	}

	// get template from DB
	template, err := srv.templatesRepo.FindByID(dto.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("could not find template: %v", err)
	}

	character := characterFromTemplate(template)
	character.Name = dto.Name
	character.Description = dto.Description
	character.BelongsUserID = dto.UserID
	if startID := ResolveStartRoomID(srv.settings, srv.rooms); startID != "" {
		character.CurrentRoomID = startID
		character.BoundRoomID = startID
	}

	if createdCharacter, err := srv.Store(character); err == nil {
		log.Info("Created new character based on template")
		return createdCharacter, nil
	}

	return nil, errors.New("could not create new character")
}

func characterFromTemplate(template *characters.CharacterTemplate) *characters.Character {
	ch := &characters.Character{
		Race:             template.Race,
		Class:            template.Class,
		CurrentHitPoints: template.CurrentHitPoints,
		MaxHitPoints:     template.MaxHitPoints,
		CurrentMana:      template.CurrentMana,
		MaxMana:          template.MaxMana,
		XP:               0,
		Level:            template.Level,
		Attributes:       template.Attributes,
	}
	if len(template.DefaultSkills) > 0 {
		ch.EquippedSkills = make([]string, len(template.DefaultSkills))
		copy(ch.EquippedSkills, template.DefaultSkills)
	}
	return ch
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

func (srv *charactersService) lockFor(id string) *sync.Mutex {
	muI, _ := srv.locks.LoadOrStore(id, &sync.Mutex{})
	return muI.(*sync.Mutex)
}

func (srv *charactersService) Modify(id string, fn func(*characters.Character) error) error {
	if id == "" {
		return errors.New("empty character id")
	}
	mu := srv.lockFor(id)
	mu.Lock()
	defer mu.Unlock()
	character, err := srv.CharactersRepository.FindByID(id)
	if err != nil {
		return err
	}
	if err := fn(character); err != nil {
		return err
	}
	return srv.CharactersRepository.Update(id, character)
}

func (srv *charactersService) GetCharacterTemplates() []*characters.CharacterTemplate {
	templates, err := srv.templatesRepo.FindAll()
	if err != nil {
		log.WithError(err).Error("Failed to fetch character templates from DB")
		return []*characters.CharacterTemplate{}
	}
	return templates
}
