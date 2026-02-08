package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/quests"
)

type sqliteQuestProgressRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteQuestProgressRepository creates a new SQLite quest progress repository.
func NewSQLiteQuestProgressRepository(client *dbsqlite.Client) QuestProgressRepository {
	return &sqliteQuestProgressRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "quest_progress", func() interface{} {
			return &quests.QuestProgress{}
		}),
	}
}

func (repo *sqliteQuestProgressRepository) FindByID(id string) (*quests.QuestProgress, error) {
	if id == "" {
		log.Error("QuestProgress::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*quests.QuestProgress), nil
	}
	return nil, err
}

func (repo *sqliteQuestProgressRepository) FindByCharacterID(characterID string) ([]*quests.QuestProgress, error) {
	results := make([]*quests.QuestProgress, 0)
	_ = repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "characterId", Value: characterID}),
		func(elem interface{}) {
			results = append(results, elem.(*quests.QuestProgress))
		})
	return results, nil
}

func (repo *sqliteQuestProgressRepository) FindByCharacterAndQuest(characterID, questID string) (*quests.QuestProgress, error) {
	all, err := repo.FindByCharacterID(characterID)
	if err != nil {
		return nil, err
	}
	for _, p := range all {
		if p.QuestID == questID {
			return p, nil
		}
	}
	return nil, nil
}

func (repo *sqliteQuestProgressRepository) FindAll() ([]*quests.QuestProgress, error) {
	results := make([]*quests.QuestProgress, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*quests.QuestProgress))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteQuestProgressRepository) Update(id string, progress *quests.QuestProgress) error {
	return repo.sqliteGenericRepo.Update(progress, id)
}

func (repo *sqliteQuestProgressRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteQuestProgressRepository) Store(progress *quests.QuestProgress) (*quests.QuestProgress, error) {
	progress.Entity = entities.NewEntity()
	return repo.Import(progress)
}

func (repo *sqliteQuestProgressRepository) Import(progress *quests.QuestProgress) (*quests.QuestProgress, error) {
	result, err := repo.sqliteGenericRepo.Store(progress)
	if result == nil {
		return nil, err
	}
	return result.(*quests.QuestProgress), nil
}
