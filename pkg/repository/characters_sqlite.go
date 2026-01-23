package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	e "github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

type sqliteCharactersRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteCharactersRepository creates a new SQLite characters repository.
func NewSQLiteCharactersRepository(client *dbsqlite.Client) CharactersRepository {
	return &sqliteCharactersRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "characters", func() interface{} {
			return &e.Character{
				EquippedItems: make(map[items.ItemSlot]*items.Item),
			}
		}),
	}
}

func (repo *sqliteCharactersRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}

func (repo *sqliteCharactersRepository) FindByID(id string) (*e.Character, error) {
	if id == "" {
		log.Error("Characters::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*e.Character), nil
	}
	return nil, err
}

func (repo *sqliteCharactersRepository) FindAllForUser(userID string) ([]*e.Character, error) {
	if userID == "" {
		log.Error("Characters::FindAllForUser - userID is empty")
		return nil, errors.New("empty userID")
	}
	results := make([]*e.Character, 0)
	if err := repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "belongsUser", Value: userID}),
		func(elem interface{}) {
			results = append(results, elem.(*e.Character))
		}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteCharactersRepository) FindByName(name string) ([]*e.Character, error) {
	results := make([]*e.Character, 0)
	_ = repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*e.Character))
		})
	return results, nil
}

func (repo *sqliteCharactersRepository) FindAll() ([]*e.Character, error) {
	results := make([]*e.Character, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*e.Character))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteCharactersRepository) Update(id string, charachterSheet *e.Character) error {
	return repo.sqliteGenericRepo.Update(charachterSheet, id)
}

func (repo *sqliteCharactersRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteCharactersRepository) Store(character *e.Character) (*e.Character, error) {
	character.Entity = entities.NewEntity()
	return repo.Import(character)
}

func (repo *sqliteCharactersRepository) Import(character *e.Character) (*e.Character, error) {
	result, err := repo.sqliteGenericRepo.Store(character)
	if result == nil {
		return nil, err
	}
	return result.(*e.Character), nil
}
