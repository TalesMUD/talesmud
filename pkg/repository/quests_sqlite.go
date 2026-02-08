package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/quests"
)

type sqliteQuestsRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteQuestsRepository creates a new SQLite quests repository.
func NewSQLiteQuestsRepository(client *dbsqlite.Client) QuestsRepository {
	return &sqliteQuestsRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "quests", func() interface{} {
			return &quests.Quest{}
		}),
	}
}

func (repo *sqliteQuestsRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}

func (repo *sqliteQuestsRepository) FindByID(id string) (*quests.Quest, error) {
	if id == "" {
		log.Error("Quests::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*quests.Quest), nil
	}
	return nil, err
}

func (repo *sqliteQuestsRepository) FindByName(name string) ([]*quests.Quest, error) {
	results := make([]*quests.Quest, 0)
	_ = repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*quests.Quest))
		})
	return results, nil
}

func (repo *sqliteQuestsRepository) FindAll() ([]*quests.Quest, error) {
	results := make([]*quests.Quest, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*quests.Quest))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteQuestsRepository) FindBySourceNPC(npcID string) ([]*quests.Quest, error) {
	results := make([]*quests.Quest, 0)
	_ = repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "source.npcId", Value: npcID}),
		func(elem interface{}) {
			results = append(results, elem.(*quests.Quest))
		})
	return results, nil
}

func (repo *sqliteQuestsRepository) Update(id string, quest *quests.Quest) error {
	return repo.sqliteGenericRepo.Update(quest, id)
}

func (repo *sqliteQuestsRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteQuestsRepository) Store(quest *quests.Quest) (*quests.Quest, error) {
	quest.Entity = entities.NewEntity()
	return repo.Import(quest)
}

func (repo *sqliteQuestsRepository) Import(quest *quests.Quest) (*quests.Quest, error) {
	result, err := repo.sqliteGenericRepo.Store(quest)
	if result == nil {
		return nil, err
	}
	return result.(*quests.Quest), nil
}
