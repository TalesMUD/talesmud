package service

import (
	"strings"
	"time"

	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// NPCsService provides business logic for NPCs
type NPCsService interface {
	r.NPCsRepository

	// FindNPCInRoomByName finds an NPC in a room by partial name match (case-insensitive)
	FindNPCInRoomByName(roomID, name string) (*npc.NPC, error)
}

type npcsService struct {
	r.NPCsRepository
}

// NewNPCsService creates a new NPCs service
func NewNPCsService(npcsRepo r.NPCsRepository) NPCsService {
	return &npcsService{
		npcsRepo,
	}
}

// Store overrides the repository store to add creation timestamp
func (srv *npcsService) Store(n *npc.NPC) (*npc.NPC, error) {
	n.Created = time.Now()
	return srv.NPCsRepository.Store(n)
}

// FindNPCInRoomByName finds an NPC in a specific room by partial name match
func (srv *npcsService) FindNPCInRoomByName(roomID, name string) (*npc.NPC, error) {
	npcs, err := srv.FindByRoom(roomID)
	if err != nil {
		return nil, err
	}

	nameLower := strings.ToLower(name)
	for _, n := range npcs {
		// Check if the NPC name contains the search term (case-insensitive)
		if strings.Contains(strings.ToLower(n.Name), nameLower) {
			return n, nil
		}
	}

	return nil, nil // Not found, but not an error
}
