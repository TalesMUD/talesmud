package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

// NPCsRepository provides access to NPC data
type NPCsRepository interface {
	FindAll() ([]*npc.NPC, error)
	FindByID(id string) (*npc.NPC, error)
	FindByName(name string) ([]*npc.NPC, error)
	FindByRoom(roomID string) ([]*npc.NPC, error)
	Store(n *npc.NPC) (*npc.NPC, error)
	Import(n *npc.NPC) (*npc.NPC, error)
	Update(id string, n *npc.NPC) error
	Delete(id string) error
	Drop() error
}

type npcsRepository struct {
	*GenericRepo
}

// NewMongoDBNPCsRepository creates a new NPCs repository
func NewMongoDBNPCsRepository(db *db.Client) NPCsRepository {
	repo := &npcsRepository{
		GenericRepo: &GenericRepo{
			db:         db,
			collection: "npcs",
			generator: func() interface{} {
				return &npc.NPC{}
			},
		},
	}
	repo.CreateIndex()
	return repo
}

// FindAll returns all NPCs
func (repo *npcsRepository) FindAll() ([]*npc.NPC, error) {
	results := make([]*npc.NPC, 0)
	if err := repo.GenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*npc.NPC))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

// FindByID returns an NPC by its ID
func (repo *npcsRepository) FindByID(id string) (*npc.NPC, error) {
	if id == "" {
		log.Error("NPCs::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.GenericRepo.FindByID(id)
	if err == nil {
		return result.(*npc.NPC), nil
	}
	return nil, err
}

// FindByName returns NPCs matching a name (can be multiple)
func (repo *npcsRepository) FindByName(name string) ([]*npc.NPC, error) {
	if name == "" {
		log.Error("NPCs::FindByName - name is empty")
		return nil, errors.New("empty name")
	}
	results := make([]*npc.NPC, 0)
	params := db.NewQueryParams().With(db.QueryParam{Key: "name", Value: name})
	if err := repo.GenericRepo.FindAllWithParam(params, func(elem interface{}) {
		results = append(results, elem.(*npc.NPC))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

// FindByRoom returns all NPCs in a specific room
func (repo *npcsRepository) FindByRoom(roomID string) ([]*npc.NPC, error) {
	if roomID == "" {
		log.Error("NPCs::FindByRoom - roomID is empty")
		return nil, errors.New("empty roomID")
	}
	results := make([]*npc.NPC, 0)
	params := db.NewQueryParams().With(db.QueryParam{Key: "currentRoomID", Value: roomID})
	if err := repo.GenericRepo.FindAllWithParam(params, func(elem interface{}) {
		results = append(results, elem.(*npc.NPC))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

// Store creates a new NPC with a new entity ID
func (repo *npcsRepository) Store(n *npc.NPC) (*npc.NPC, error) {
	n.Entity = entities.NewEntity()
	return repo.Import(n)
}

// Import stores an NPC (used when NPC already has an ID)
func (repo *npcsRepository) Import(n *npc.NPC) (*npc.NPC, error) {
	if _, err := repo.GenericRepo.Store(n); err != nil {
		return nil, err
	}
	return n, nil
}

// Update updates an existing NPC
func (repo *npcsRepository) Update(id string, n *npc.NPC) error {
	return repo.GenericRepo.Update(n, id)
}

// Delete removes an NPC
func (repo *npcsRepository) Delete(id string) error {
	return repo.GenericRepo.Delete(id)
}

// Drop removes all NPCs
func (repo *npcsRepository) Drop() error {
	return repo.GenericRepo.DropCollection()
}
