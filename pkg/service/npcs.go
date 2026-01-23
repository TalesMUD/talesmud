package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// NPCsService provides business logic for NPCs
type NPCsService interface {
	r.NPCsRepository

	// FindNPCInRoomByName finds an NPC in a room by partial name match (case-insensitive)
	FindNPCInRoomByName(roomID, name string) (*npc.NPC, error)

	// FindAllTemplates returns all NPCs marked as templates
	FindAllTemplates() ([]*npc.NPC, error)

	// FindAllNonTemplates returns all NPCs that are not templates (singletons)
	FindAllNonTemplates() ([]*npc.NPC, error)

	// SpawnFromTemplate creates a new NPC instance from a template
	// The instance is returned but NOT persisted (memory-only)
	SpawnFromTemplate(templateID, roomID string) (*npc.NPC, error)
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

// FindAllTemplates returns all NPCs marked as templates
func (srv *npcsService) FindAllTemplates() ([]*npc.NPC, error) {
	all, err := srv.FindAll()
	if err != nil {
		return nil, err
	}

	var templates []*npc.NPC
	for _, n := range all {
		if n.IsTemplate {
			templates = append(templates, n)
		}
	}
	return templates, nil
}

// FindAllNonTemplates returns all NPCs that are not templates
func (srv *npcsService) FindAllNonTemplates() ([]*npc.NPC, error) {
	all, err := srv.FindAll()
	if err != nil {
		return nil, err
	}

	var npcs []*npc.NPC
	for _, n := range all {
		if !n.IsTemplate {
			npcs = append(npcs, n)
		}
	}
	return npcs, nil
}

// SpawnFromTemplate creates a new NPC instance from a template
// The instance is returned but NOT persisted (memory-only)
// Note: For unique NPCs, the caller must check if an instance already exists before calling this
func (srv *npcsService) SpawnFromTemplate(templateID, roomID string) (*npc.NPC, error) {
	template, err := srv.FindByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}
	if template == nil {
		return nil, fmt.Errorf("template %s not found", templateID)
	}
	if !template.IsTemplate {
		return nil, fmt.Errorf("NPC %s is not a template", templateID)
	}

	// Generate unique suffix (first 8 chars of UUID)
	suffix := uuid.New().String()[:8]

	// Create instance with copied template data
	instance := &npc.NPC{
		Entity:         entities.NewEntity(),
		Name:           template.Name,
		Description:    template.Description,
		Race:           template.Race,
		Class:          template.Class,
		Level:          template.Level,
		MaxHitPoints:   template.MaxHitPoints,
		CurrentHitPoints: template.MaxHitPoints, // Full HP on spawn

		// Template reference
		IsTemplate:     false,
		TemplateID:     templateID,
		InstanceSuffix: suffix,

		// Behavior configuration (copied from template)
		SpawnRoomID:  roomID,
		RespawnTime:  template.RespawnTime,
		WanderRadius: template.WanderRadius,
		PatrolPath:   template.PatrolPath,

		// Initial state
		IsDead: false,
		State:  "idle",

		// Dialog references
		DialogID:          template.DialogID,
		IdleDialogID:      template.IdleDialogID,
		IdleDialogTimeout: template.IdleDialogTimeout,

		// Copy traits
		EnemyTrait:    template.EnemyTrait,
		MerchantTrait: template.MerchantTrait,

		Created: time.Now(),
	}

	// Set current room
	instance.CurrentRoomID = roomID

	return instance, nil
}
