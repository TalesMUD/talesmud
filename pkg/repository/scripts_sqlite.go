package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	s "github.com/talesmud/talesmud/pkg/scripts"
)

type sqliteScriptsRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteScriptsRepository creates a new SQLite scripts repository.
func NewSQLiteScriptsRepository(client *dbsqlite.Client) ScriptsRepository {
	return &sqliteScriptsRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "scripts", func() interface{} {
			return &s.Script{}
		}),
	}
}

func (repo *sqliteScriptsRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}

func (repo *sqliteScriptsRepository) FindByID(id string) (*s.Script, error) {
	if id == "" {
		log.Error("Scripts::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*s.Script), nil
	}
	return nil, err
}

func (repo *sqliteScriptsRepository) FindByName(name string) ([]*s.Script, error) {
	results := make([]*s.Script, 0)
	_ = repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*s.Script))
		})
	return results, nil
}

func (repo *sqliteScriptsRepository) FindAll() ([]*s.Script, error) {
	results := make([]*s.Script, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*s.Script))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteScriptsRepository) Update(id string, script *s.Script) error {
	return repo.sqliteGenericRepo.Update(script, id)
}

func (repo *sqliteScriptsRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteScriptsRepository) Store(script *s.Script) (*s.Script, error) {
	script.Entity = entities.NewEntity()
	return repo.Import(script)
}

func (repo *sqliteScriptsRepository) Import(script *s.Script) (*s.Script, error) {
	result, err := repo.sqliteGenericRepo.Store(script)
	if result == nil {
		return nil, err
	}
	return result.(*s.Script), nil
}
