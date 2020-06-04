package repository

import (
	"github.com/atla/owndnd/pkg/db"
	"github.com/atla/owndnd/pkg/entities"
	e "github.com/atla/owndnd/pkg/entities/characters"
)

//--- Interface Definitions

//CharactersRepository repository interface
type CharactersRepository interface {
	FindAll() ([]*e.Character, error)
	FindByID(id string) (*e.Character, error)
	FindAllForUser(userID string) ([]*e.Character, error)
	FindByName(name string) ([]*e.Character, error)
	Store(Character *e.Character) (*e.Character, error)
	Import(Character *e.Character) (*e.Character, error)
	Update(id string, Character *e.Character) error
	Delete(id string) error
}

//--- Implementations

type charactersRepository struct {
	*GenericRepo
}

//NewMongoDBcharactersRepository creates a new mongodb charactersRepository
func NewMongoDBcharactersRepository(db *db.Client) CharactersRepository {
	return &charactersRepository{
		GenericRepo: &GenericRepo{
			db:         db,
			collection: "characters",
			generator: func() interface{} {
				return &e.Character{}
			},
		},
	}
}

func (repo *charactersRepository) FindByID(id string) (*e.Character, error) {
	result, err := repo.GenericRepo.FindByID(id)
	if err == nil {
		return result.(*e.Character), nil
	}
	return nil, err
}
func (repo *charactersRepository) FindAllForUser(userID string) ([]*e.Character, error) {

	results := make([]*e.Character, 0)

	if err := repo.GenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "belongsUser", Value: userID}),
		func(elem interface{}) {
			results = append(results, elem.(*e.Character))
		}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *charactersRepository) FindByName(name string) ([]*e.Character, error) {
	results := make([]*e.Character, 0)

	if err := repo.GenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*e.Character))
		}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *charactersRepository) FindAll() ([]*e.Character, error) {
	results := make([]*e.Character, 0)
	if err := repo.GenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*e.Character))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *charactersRepository) Update(id string, charachterSheet *e.Character) error {
	return repo.GenericRepo.Update(charachterSheet, id)
}

func (repo *charactersRepository) Delete(id string) error {
	return repo.GenericRepo.Delete(id)
}

func (repo *charactersRepository) Store(character *e.Character) (*e.Character, error) {
	character.Entity = entities.NewEntity()
	return repo.Import(character)
}
func (repo *charactersRepository) Import(character *e.Character) (*e.Character, error) {
	result, err := repo.GenericRepo.Store(character)
	return result.(*e.Character), err
}
