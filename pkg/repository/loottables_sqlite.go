package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

type sqliteLootTablesRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteLootTablesRepository creates a new SQLite loot tables repository.
func NewSQLiteLootTablesRepository(client *dbsqlite.Client) LootTablesRepository {
	return &sqliteLootTablesRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "loot_tables", func() interface{} {
			return &items.LootTable{}
		}),
	}
}

func (repo *sqliteLootTablesRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}

func (repo *sqliteLootTablesRepository) FindByID(id string) (*items.LootTable, error) {
	if id == "" {
		log.Error("LootTables::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*items.LootTable), nil
	}
	return nil, err
}

func (repo *sqliteLootTablesRepository) FindByName(name string) ([]*items.LootTable, error) {
	results := make([]*items.LootTable, 0)
	_ = repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*items.LootTable))
		})
	return results, nil
}

func (repo *sqliteLootTablesRepository) FindAll() ([]*items.LootTable, error) {
	results := make([]*items.LootTable, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*items.LootTable))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteLootTablesRepository) Update(id string, lootTable *items.LootTable) error {
	return repo.sqliteGenericRepo.Update(lootTable, id)
}

func (repo *sqliteLootTablesRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteLootTablesRepository) Store(lootTable *items.LootTable) (*items.LootTable, error) {
	lootTable.Entity = entities.NewEntity()
	return repo.Import(lootTable)
}

func (repo *sqliteLootTablesRepository) Import(lootTable *items.LootTable) (*items.LootTable, error) {
	result, err := repo.sqliteGenericRepo.Store(lootTable)
	if result == nil {
		return nil, err
	}
	return result.(*items.LootTable), nil
}
