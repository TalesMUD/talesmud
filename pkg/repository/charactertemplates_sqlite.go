package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	c "github.com/talesmud/talesmud/pkg/entities/characters"
)

type sqliteCharacterTemplatesRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteCharacterTemplatesRepository creates a new SQLite character templates repository.
func NewSQLiteCharacterTemplatesRepository(client *dbsqlite.Client) CharacterTemplatesRepository {
	return &sqliteCharacterTemplatesRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "charactertemplates", func() interface{} {
			return &c.CharacterTemplate{}
		}),
	}
}

func (repo *sqliteCharacterTemplatesRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}

func (repo *sqliteCharacterTemplatesRepository) FindByID(id string) (*c.CharacterTemplate, error) {
	if id == "" {
		log.Error("CharacterTemplates::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*c.CharacterTemplate), nil
	}
	return nil, err
}

func (repo *sqliteCharacterTemplatesRepository) FindByName(name string) ([]*c.CharacterTemplate, error) {
	results := make([]*c.CharacterTemplate, 0)
	_ = repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*c.CharacterTemplate))
		})
	return results, nil
}

func (repo *sqliteCharacterTemplatesRepository) FindAll() ([]*c.CharacterTemplate, error) {
	results := make([]*c.CharacterTemplate, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*c.CharacterTemplate))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteCharacterTemplatesRepository) Count() (int, error) {
	templates, err := repo.FindAll()
	if err != nil {
		return 0, err
	}
	return len(templates), nil
}

func (repo *sqliteCharacterTemplatesRepository) Update(id string, template *c.CharacterTemplate) error {
	return repo.sqliteGenericRepo.Update(template, id)
}

func (repo *sqliteCharacterTemplatesRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteCharacterTemplatesRepository) Store(template *c.CharacterTemplate) (*c.CharacterTemplate, error) {
	template.Entity = entities.NewEntity()
	return repo.Import(template)
}

func (repo *sqliteCharacterTemplatesRepository) Import(template *c.CharacterTemplate) (*c.CharacterTemplate, error) {
	result, err := repo.sqliteGenericRepo.Store(template)
	if result == nil {
		return nil, err
	}
	return result.(*c.CharacterTemplate), nil
}
