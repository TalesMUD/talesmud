package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	i "github.com/talesmud/talesmud/pkg/entities/items"
)

type sqliteItemsRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteItemsRepository creates a new SQLite items repository.
func NewSQLiteItemsRepository(client *dbsqlite.Client) ItemsRepository {
	return &sqliteItemsRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "items", func() interface{} {
			return &i.Item{}
		}),
	}
}

func (repo *sqliteItemsRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}

func (repo *sqliteItemsRepository) FindByID(id string) (*i.Item, error) {
	if id == "" {
		log.Error("Items::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*i.Item), nil
	}
	return nil, err
}

func (repo *sqliteItemsRepository) FindByName(name string) ([]*i.Item, error) {
	results := make([]*i.Item, 0)
	_ = repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*i.Item))
		})
	return results, nil
}

func (repo *sqliteItemsRepository) FindAll(query ItemsQuery) ([]*i.Item, error) {
	results := make([]*i.Item, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*i.Item))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteItemsRepository) Update(id string, item *i.Item) error {
	return repo.sqliteGenericRepo.Update(item, id)
}

func (repo *sqliteItemsRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteItemsRepository) Store(item *i.Item) (*i.Item, error) {
	item.Entity = entities.NewEntity()
	return repo.Import(item)
}

func (repo *sqliteItemsRepository) Import(item *i.Item) (*i.Item, error) {
	result, err := repo.sqliteGenericRepo.Store(item)
	if result == nil {
		return nil, err
	}
	return result.(*i.Item), nil
}
