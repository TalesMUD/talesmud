package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/conversations"
)

type sqliteConversationsRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteConversationsRepository creates a new SQLite conversations repository.
func NewSQLiteConversationsRepository(client *dbsqlite.Client) ConversationsRepository {
	return &sqliteConversationsRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "conversations", func() interface{} {
			return &conversations.Conversation{
				VisitedNodes: make(map[string]int),
				Context:      make(map[string]string),
			}
		}),
	}
}

func (repo *sqliteConversationsRepository) FindByID(id string) (*conversations.Conversation, error) {
	if id == "" {
		log.Error("Conversations::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*conversations.Conversation), nil
	}
	return nil, err
}

func (repo *sqliteConversationsRepository) FindByCharacterAndTarget(characterID, targetID string) (*conversations.Conversation, error) {
	if characterID == "" || targetID == "" {
		log.Error("Conversations::FindByCharacterAndTarget - characterID or targetID is empty")
		return nil, errors.New("empty characterID or targetID")
	}

	params := db.NewQueryParams().
		With(db.QueryParam{Key: "characterID", Value: characterID}).
		With(db.QueryParam{Key: "targetID", Value: targetID})

	var result *conversations.Conversation
	err := repo.sqliteGenericRepo.FindAllWithParam(params, func(elem interface{}) {
		if result == nil {
			result = elem.(*conversations.Conversation)
		}
	})

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("conversation not found")
	}
	return result, nil
}

func (repo *sqliteConversationsRepository) FindAllForCharacter(characterID string) ([]*conversations.Conversation, error) {
	if characterID == "" {
		log.Error("Conversations::FindAllForCharacter - characterID is empty")
		return nil, errors.New("empty characterID")
	}

	results := make([]*conversations.Conversation, 0)
	params := db.NewQueryParams().With(db.QueryParam{Key: "characterID", Value: characterID})
	if err := repo.sqliteGenericRepo.FindAllWithParam(params, func(elem interface{}) {
		results = append(results, elem.(*conversations.Conversation))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteConversationsRepository) Store(conv *conversations.Conversation) (*conversations.Conversation, error) {
	conv.Entity = entities.NewEntity()
	if _, err := repo.sqliteGenericRepo.Store(conv); err != nil {
		return nil, err
	}
	return conv, nil
}

func (repo *sqliteConversationsRepository) Update(id string, conv *conversations.Conversation) error {
	return repo.sqliteGenericRepo.Update(conv, id)
}

func (repo *sqliteConversationsRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}
