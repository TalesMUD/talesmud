package service

import (
	"time"

	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// NPCSpawnersService provides business logic for NPC spawners
type NPCSpawnersService interface {
	r.NPCSpawnersRepository
}

type npcSpawnersService struct {
	r.NPCSpawnersRepository
}

// NewNPCSpawnersService creates a new NPC spawners service
func NewNPCSpawnersService(repo r.NPCSpawnersRepository) NPCSpawnersService {
	return &npcSpawnersService{
		repo,
	}
}

// Store overrides the repository store to add creation timestamp
func (srv *npcSpawnersService) Store(s *npc.NPCSpawner) (*npc.NPCSpawner, error) {
	s.Created = time.Now()
	return srv.NPCSpawnersRepository.Store(s)
}

// Update overrides the repository update to set updated timestamp
func (srv *npcSpawnersService) Update(id string, s *npc.NPCSpawner) error {
	s.Updated = time.Now()
	return srv.NPCSpawnersRepository.Update(id, s)
}
