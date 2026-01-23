package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

type sqliteNPCsRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteNPCsRepository creates a new SQLite NPC repository.
func NewSQLiteNPCsRepository(client *dbsqlite.Client) NPCsRepository {
	return &sqliteNPCsRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "npcs", func() interface{} {
			return &npc.NPC{}
		}),
	}
}

func (repo *sqliteNPCsRepository) FindAll() ([]*npc.NPC, error) {
	results := make([]*npc.NPC, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*npc.NPC))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteNPCsRepository) FindByID(id string) (*npc.NPC, error) {
	if id == "" {
		log.Error("NPCs::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*npc.NPC), nil
	}
	return nil, err
}

func (repo *sqliteNPCsRepository) FindByName(name string) ([]*npc.NPC, error) {
	if name == "" {
		log.Error("NPCs::FindByName - name is empty")
		return nil, errors.New("empty name")
	}
	results := make([]*npc.NPC, 0)
	params := db.NewQueryParams().With(db.QueryParam{Key: "name", Value: name})
	if err := repo.sqliteGenericRepo.FindAllWithParam(params, func(elem interface{}) {
		results = append(results, elem.(*npc.NPC))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteNPCsRepository) FindByRoom(roomID string) ([]*npc.NPC, error) {
	if roomID == "" {
		log.Error("NPCs::FindByRoom - roomID is empty")
		return nil, errors.New("empty roomID")
	}
	results := make([]*npc.NPC, 0)
	params := db.NewQueryParams().With(db.QueryParam{Key: "currentRoomID", Value: roomID})
	if err := repo.sqliteGenericRepo.FindAllWithParam(params, func(elem interface{}) {
		results = append(results, elem.(*npc.NPC))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteNPCsRepository) Store(n *npc.NPC) (*npc.NPC, error) {
	n.Entity = entities.NewEntity()
	return repo.Import(n)
}

func (repo *sqliteNPCsRepository) Import(n *npc.NPC) (*npc.NPC, error) {
	if _, err := repo.sqliteGenericRepo.Store(n); err != nil {
		return nil, err
	}
	return n, nil
}

func (repo *sqliteNPCsRepository) Update(id string, n *npc.NPC) error {
	return repo.sqliteGenericRepo.Update(n, id)
}

func (repo *sqliteNPCsRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteNPCsRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}
