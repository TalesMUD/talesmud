package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/conversations"
)

// ConversationsRepository provides access to conversation state data
type ConversationsRepository interface {
	FindByID(id string) (*conversations.Conversation, error)
	FindByCharacterAndTarget(characterID, targetID string) (*conversations.Conversation, error)
	FindAllForCharacter(characterID string) ([]*conversations.Conversation, error)
	Store(conv *conversations.Conversation) (*conversations.Conversation, error)
	Update(id string, conv *conversations.Conversation) error
	Delete(id string) error
}

type conversationsRepository struct {
	*GenericRepo
}

// NewMongoDBConversationsRepository creates a new conversations repository
func NewMongoDBConversationsRepository(db *db.Client) ConversationsRepository {
	repo := &conversationsRepository{
		GenericRepo: &GenericRepo{
			db:         db,
			collection: "conversations",
			generator: func() interface{} {
				return &conversations.Conversation{
					VisitedNodes: make(map[string]int),
					Context:      make(map[string]string),
				}
			},
		},
	}
	repo.CreateIndex()
	return repo
}

// FindByID returns a conversation by its ID
func (repo *conversationsRepository) FindByID(id string) (*conversations.Conversation, error) {
	if id == "" {
		log.Error("Conversations::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.GenericRepo.FindByID(id)
	if err == nil {
		return result.(*conversations.Conversation), nil
	}
	return nil, err
}

// FindByCharacterAndTarget finds a conversation between a specific character and target
func (repo *conversationsRepository) FindByCharacterAndTarget(characterID, targetID string) (*conversations.Conversation, error) {
	if characterID == "" || targetID == "" {
		log.Error("Conversations::FindByCharacterAndTarget - characterID or targetID is empty")
		return nil, errors.New("empty characterID or targetID")
	}

	params := db.NewQueryParams().
		With(db.QueryParam{Key: "characterID", Value: characterID}).
		With(db.QueryParam{Key: "targetID", Value: targetID})

	var result *conversations.Conversation
	err := repo.GenericRepo.FindAllWithParam(params, func(elem interface{}) {
		// Take the first match (should only be one)
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

// FindAllForCharacter returns all conversations for a character
func (repo *conversationsRepository) FindAllForCharacter(characterID string) ([]*conversations.Conversation, error) {
	if characterID == "" {
		log.Error("Conversations::FindAllForCharacter - characterID is empty")
		return nil, errors.New("empty characterID")
	}

	results := make([]*conversations.Conversation, 0)
	params := db.NewQueryParams().With(db.QueryParam{Key: "characterID", Value: characterID})
	if err := repo.GenericRepo.FindAllWithParam(params, func(elem interface{}) {
		results = append(results, elem.(*conversations.Conversation))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

// Store creates a new conversation with a new entity ID
func (repo *conversationsRepository) Store(conv *conversations.Conversation) (*conversations.Conversation, error) {
	conv.Entity = entities.NewEntity()
	if _, err := repo.GenericRepo.Store(conv); err != nil {
		return nil, err
	}
	return conv, nil
}

// Update updates an existing conversation
func (repo *conversationsRepository) Update(id string, conv *conversations.Conversation) error {
	return repo.GenericRepo.Update(conv, id)
}

// Delete removes a conversation
func (repo *conversationsRepository) Delete(id string) error {
	return repo.GenericRepo.Delete(id)
}
