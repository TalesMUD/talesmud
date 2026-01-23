package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

type sqliteNPCSpawnersRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteNPCSpawnersRepository creates a new SQLite NPC spawners repository.
func NewSQLiteNPCSpawnersRepository(client *dbsqlite.Client) NPCSpawnersRepository {
	return &sqliteNPCSpawnersRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "npc_spawners", func() interface{} {
			return &npc.NPCSpawner{}
		}),
	}
}

func (repo *sqliteNPCSpawnersRepository) FindAll() ([]*npc.NPCSpawner, error) {
	results := make([]*npc.NPCSpawner, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*npc.NPCSpawner))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteNPCSpawnersRepository) FindByID(id string) (*npc.NPCSpawner, error) {
	if id == "" {
		log.Error("NPCSpawners::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*npc.NPCSpawner), nil
	}
	return nil, err
}

func (repo *sqliteNPCSpawnersRepository) FindByRoom(roomID string) ([]*npc.NPCSpawner, error) {
	if roomID == "" {
		log.Error("NPCSpawners::FindByRoom - roomID is empty")
		return nil, errors.New("empty roomID")
	}
	results := make([]*npc.NPCSpawner, 0)
	params := db.NewQueryParams().With(db.QueryParam{Key: "roomId", Value: roomID})
	if err := repo.sqliteGenericRepo.FindAllWithParam(params, func(elem interface{}) {
		results = append(results, elem.(*npc.NPCSpawner))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteNPCSpawnersRepository) FindByTemplate(templateID string) ([]*npc.NPCSpawner, error) {
	if templateID == "" {
		log.Error("NPCSpawners::FindByTemplate - templateID is empty")
		return nil, errors.New("empty templateID")
	}
	results := make([]*npc.NPCSpawner, 0)
	params := db.NewQueryParams().With(db.QueryParam{Key: "templateId", Value: templateID})
	if err := repo.sqliteGenericRepo.FindAllWithParam(params, func(elem interface{}) {
		results = append(results, elem.(*npc.NPCSpawner))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteNPCSpawnersRepository) Store(s *npc.NPCSpawner) (*npc.NPCSpawner, error) {
	s.Entity = entities.NewEntity()
	return repo.Import(s)
}

func (repo *sqliteNPCSpawnersRepository) Import(s *npc.NPCSpawner) (*npc.NPCSpawner, error) {
	if _, err := repo.sqliteGenericRepo.Store(s); err != nil {
		return nil, err
	}
	return s, nil
}

func (repo *sqliteNPCSpawnersRepository) Update(id string, s *npc.NPCSpawner) error {
	return repo.sqliteGenericRepo.Update(s, id)
}

func (repo *sqliteNPCSpawnersRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteNPCSpawnersRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}
