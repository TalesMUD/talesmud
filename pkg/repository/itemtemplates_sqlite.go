package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	i "github.com/talesmud/talesmud/pkg/entities/items"
)

type sqliteItemTemplatesRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteItemTemplatesRepository creates a new SQLite item templates repository.
func NewSQLiteItemTemplatesRepository(client *dbsqlite.Client) ItemTemplatesRepository {
	return &sqliteItemTemplatesRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "itemtemplates", func() interface{} {
			return &i.ItemTemplate{}
		}),
	}
}

func (repo *sqliteItemTemplatesRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}

func (repo *sqliteItemTemplatesRepository) FindByID(id string) (*i.ItemTemplate, error) {
	if id == "" {
		log.Error("ItemTemplates::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*i.ItemTemplate), nil
	}
	return nil, err
}

func (repo *sqliteItemTemplatesRepository) FindByName(name string) ([]*i.ItemTemplate, error) {
	results := make([]*i.ItemTemplate, 0)
	_ = repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*i.ItemTemplate))
		})
	return results, nil
}

func (repo *sqliteItemTemplatesRepository) FindAll(query ItemsQuery) ([]*i.ItemTemplate, error) {
	results := make([]*i.ItemTemplate, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		item := elem.(*i.ItemTemplate)
		if query.matches(&item.Item) {
			results = append(results, item)
		}
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteItemTemplatesRepository) Update(id string, item *i.ItemTemplate) error {
	return repo.sqliteGenericRepo.Update(item, id)
}

func (repo *sqliteItemTemplatesRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteItemTemplatesRepository) Store(item *i.ItemTemplate) (*i.ItemTemplate, error) {
	item.Entity = entities.NewEntity()
	return repo.Import(item)
}

func (repo *sqliteItemTemplatesRepository) Import(item *i.ItemTemplate) (*i.ItemTemplate, error) {
	result, err := repo.sqliteGenericRepo.Store(item)
	if result == nil {
		return nil, err
	}
	return result.(*i.ItemTemplate), nil
}
