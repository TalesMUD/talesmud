package repository

import (
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	e "github.com/talesmud/talesmud/pkg/entities"
)

type sqlitePartiesRepository struct {
	*sqliteGenericRepo
}

// NewSQLitePartiesRepository creates a new SQLite parties repository.
func NewSQLitePartiesRepository(client *dbsqlite.Client) PartiesRepository {
	return &sqlitePartiesRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "parties", func() interface{} {
			return &e.Party{}
		}),
	}
}

func (repo *sqlitePartiesRepository) FindAll() ([]*e.Party, error) {
	results := make([]*e.Party, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*e.Party))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqlitePartiesRepository) Store(party *e.Party) (*e.Party, error) {
	party.Entity = e.NewEntity()
	result, err := repo.sqliteGenericRepo.Store(party)
	return result.(*e.Party), err
}

func (repo *sqlitePartiesRepository) FindByID(id string) (*e.Party, error) {
	result, err := repo.sqliteGenericRepo.FindByID(id)
	return result.(*e.Party), err
}

func (repo *sqlitePartiesRepository) Update(id string, party *e.Party) error {
	return repo.sqliteGenericRepo.Update(party, id)
}

func (repo *sqlitePartiesRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}
