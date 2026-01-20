package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/db"
	"github.com/talesmud/talesmud/pkg/entities"
	c "github.com/talesmud/talesmud/pkg/entities/characters"
)

//--- Interface Definitions

//CharacterTemplatesRepository repository interface
type CharacterTemplatesRepository interface {
	FindAll() ([]*c.CharacterTemplate, error)
	FindByID(id string) (*c.CharacterTemplate, error)
	FindByName(name string) ([]*c.CharacterTemplate, error)
	Store(template *c.CharacterTemplate) (*c.CharacterTemplate, error)
	Import(template *c.CharacterTemplate) (*c.CharacterTemplate, error)
	Update(id string, template *c.CharacterTemplate) error
	Delete(id string) error
	Drop() error
	Count() (int, error)
}

//--- Implementations
type characterTemplatesRepository struct {
	*GenericRepo
}

//NewMongoDBCharacterTemplatesRepository creates a new mongodb character templates repository
func NewMongoDBCharacterTemplatesRepository(db *db.Client) CharacterTemplatesRepository {
	repo := &characterTemplatesRepository{
		GenericRepo: &GenericRepo{
			db:         db,
			collection: "charactertemplates",
			generator: func() interface{} {
				return &c.CharacterTemplate{}
			},
		},
	}
	repo.CreateIndex()
	return repo
}

// Drop ...
func (repo *characterTemplatesRepository) Drop() error {
	return repo.GenericRepo.DropCollection()
}

func (repo *characterTemplatesRepository) FindByID(id string) (*c.CharacterTemplate, error) {
	if id == "" {
		log.Error("CharacterTemplates::FindByID - id is empty")
		return nil, errors.New("empty id")
	}

	result, err := repo.GenericRepo.FindByID(id)
	if err == nil {
		return result.(*c.CharacterTemplate), nil
	}
	return nil, err
}

func (repo *characterTemplatesRepository) FindByName(name string) ([]*c.CharacterTemplate, error) {
	results := make([]*c.CharacterTemplate, 0)
	repo.GenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*c.CharacterTemplate))
		})
	return results, nil
}

func (repo *characterTemplatesRepository) FindAll() ([]*c.CharacterTemplate, error) {
	results := make([]*c.CharacterTemplate, 0)
	if err := repo.GenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*c.CharacterTemplate))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *characterTemplatesRepository) Count() (int, error) {
	templates, err := repo.FindAll()
	if err != nil {
		return 0, err
	}
	return len(templates), nil
}

func (repo *characterTemplatesRepository) Update(id string, template *c.CharacterTemplate) error {
	return repo.GenericRepo.Update(template, id)
}

func (repo *characterTemplatesRepository) Delete(id string) error {
	return repo.GenericRepo.Delete(id)
}

func (repo *characterTemplatesRepository) Store(template *c.CharacterTemplate) (*c.CharacterTemplate, error) {
	template.Entity = entities.NewEntity()
	return repo.Import(template)
}

func (repo *characterTemplatesRepository) Import(template *c.CharacterTemplate) (*c.CharacterTemplate, error) {
	result, err := repo.GenericRepo.Store(template)
	if result == nil {
		return nil, err
	}
	return result.(*c.CharacterTemplate), nil
}
