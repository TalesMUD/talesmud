package service

import (
	"errors"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// QuestsService delivers logical functions on top of the quests repositories
type QuestsService interface {
	// Quest definition CRUD
	FindAll() ([]*quests.Quest, error)
	FindByID(id string) (*quests.Quest, error)
	FindByName(name string) ([]*quests.Quest, error)
	FindBySourceNPC(npcID string) ([]*quests.Quest, error)
	Store(quest *quests.Quest) (*quests.Quest, error)
	Update(id string, quest *quests.Quest) error
	Delete(id string) error

	// Quest progress operations
	GetQuestLog(characterID string) ([]*quests.QuestProgress, error)
	BuildQuestLog(characterID string) ([]QuestLogEntry, error)
	GetProgress(characterID, questID string) (*quests.QuestProgress, error)
	AcceptQuest(characterID, questID string) (*quests.QuestProgress, error)
	AbandonQuest(characterID, questID string) error
	UpdateProgress(progress *quests.QuestProgress) error
	CompleteQuest(characterID, questID string) (*quests.QuestProgress, error)
	GetAvailableQuests(characterID string) ([]*quests.Quest, error)
	GrantAutoQuests(characterID, roomArea string) int

	// Objective progress (called by QuestTracker)
	ApplyQuestEvent(event QuestEvent) ([]QuestEventResult, error)
	IncrementObjective(characterID, questID, objectiveID string, amount int32) (*quests.QuestProgress, error)
	CompleteDeliveryObjective(characterID, questID, objectiveID string) (*quests.QuestProgress, error)
	CheckObjectives(characterID, questID string) (bool, error)

	// Reward granting
	TurnInQuest(characterID, questID, npcID string) (*QuestTurnInResult, error)
	TurnInQuestAnywhere(characterID, questID string) (*QuestTurnInResult, error)
	GrantQuestRewards(characterID, questID string) ([]string, error)

	// Definition validation
	ValidateQuest(quest *quests.Quest) []QuestValidationIssue

	// Internal setter for facade dependency
	SetFacade(facade Facade)
}

// QuestValidationIssue describes a quest authoring problem.
type QuestValidationIssue struct {
	Severity string `json:"severity"`
	Path     string `json:"path"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// QuestObjectiveProgressEntry combines objective definition text with player progress.
type QuestObjectiveProgressEntry struct {
	ObjectiveID string `json:"objectiveId"`
	Description string `json:"description"`
	Current     int32  `json:"current"`
	Required    int32  `json:"required"`
	Completed   bool   `json:"completed"`
}

// QuestRewardEntry describes rewards in player-facing quest log responses.
type QuestRewardEntry struct {
	XP              int32    `json:"xp,omitempty"`
	Gold            int64    `json:"gold,omitempty"`
	ItemTemplateIDs []string `json:"itemTemplateIds,omitempty"`
}

// QuestLogEntry is an enriched quest log entry for APIs and WebSocket messages.
type QuestLogEntry struct {
	QuestID        string                        `json:"questId"`
	QuestName      string                        `json:"questName"`
	Description    string                        `json:"description,omitempty"`
	Category       string                        `json:"category,omitempty"`
	Level          int32                         `json:"level,omitempty"`
	Status         string                        `json:"status"`
	ReadyToTurnIn  bool                          `json:"readyToTurnIn"`
	TurnInAnywhere bool                          `json:"turnInAnywhere,omitempty"`
	TurnInNpcID    string                        `json:"turnInNpcId,omitempty"`
	TurnInNpcName  string                        `json:"turnInNpcName,omitempty"`
	Objectives     []QuestObjectiveProgressEntry `json:"objectives"`
	Rewards        *QuestRewardEntry             `json:"rewards,omitempty"`
	AcceptedAt     string                        `json:"acceptedAt,omitempty"`
	CompletedAt    string                        `json:"completedAt,omitempty"`
}

// QuestEventType identifies a game event that can advance active quest objectives.
type QuestEventType string

const (
	QuestEventNPCKilled  QuestEventType = "npcKilled"
	QuestEventItemPickup QuestEventType = "itemPickup"
	QuestEventRoomEnter  QuestEventType = "roomEnter"
	QuestEventDialogNode QuestEventType = "dialogNode"
	QuestEventTalkToNPC  QuestEventType = "talkToNPC"
)

// QuestEvent describes a normalized game event for quest progress rules.
type QuestEvent struct {
	Type           QuestEventType
	CharacterID    string
	NPCID          string
	RoomID         string
	DialogID       string
	DialogNodeID   string
	Item           *items.Item
	ItemTemplateID string
	Amount         int32
}

// QuestEventResultKind identifies the user-facing result of applying an event.
type QuestEventResultKind string

const (
	QuestEventResultProgress      QuestEventResultKind = "progress"
	QuestEventResultReadyToTurnIn QuestEventResultKind = "readyToTurnIn"
)

// QuestEventResult describes a quest objective changed by an event.
type QuestEventResult struct {
	Kind        QuestEventResultKind
	QuestID     string
	QuestName   string
	ObjectiveID string
	Progress    *quests.QuestProgress
}

// QuestTurnInResult describes rewards granted by a successful quest turn-in.
type QuestTurnInResult struct {
	QuestID       string
	QuestName     string
	GrantedItems  []string
	XP            int32
	Gold          int64
	CompletedAt   time.Time
	QuestProgress *quests.QuestProgress
}

type questsService struct {
	questsRepo   r.QuestsRepository
	progressRepo r.QuestProgressRepository
	facade       Facade
}

// NewQuestsService creates a new quests service
func NewQuestsService(questsRepo r.QuestsRepository, progressRepo r.QuestProgressRepository) QuestsService {
	return &questsService{
		questsRepo:   questsRepo,
		progressRepo: progressRepo,
	}
}

// SetFacade sets the facade reference (called after all services are created)
func (s *questsService) SetFacade(facade Facade) {
	s.facade = facade
}

// --- Quest definition CRUD ---

func (s *questsService) FindAll() ([]*quests.Quest, error) {
	return s.questsRepo.FindAll()
}

func (s *questsService) FindByID(id string) (*quests.Quest, error) {
	return s.questsRepo.FindByID(id)
}

func (s *questsService) FindByName(name string) ([]*quests.Quest, error) {
	return s.questsRepo.FindByName(name)
}

func (s *questsService) FindBySourceNPC(npcID string) ([]*quests.Quest, error) {
	return s.questsRepo.FindBySourceNPC(npcID)
}

func (s *questsService) Store(quest *quests.Quest) (*quests.Quest, error) {
	if err := validateQuestDefinition(quest); err != nil {
		return nil, err
	}
	if err := validateQuestReferences(quest, s.facade); err != nil {
		return nil, err
	}
	quest.Created = time.Now()
	quest.Updated = time.Now()
	return s.questsRepo.Store(quest)
}

func (s *questsService) Update(id string, quest *quests.Quest) error {
	if quest.Entity == nil {
		quest.Entity = &entities.Entity{ID: id}
	} else if quest.ID == "" {
		quest.ID = id
	}
	if err := validateQuestDefinition(quest); err != nil {
		return err
	}
	if err := validateQuestReferences(quest, s.facade); err != nil {
		return err
	}
	quest.Updated = time.Now()
	return s.questsRepo.Update(id, quest)
}

func (s *questsService) Delete(id string) error {
	return s.questsRepo.Delete(id)
}

func (s *questsService) ValidateQuest(quest *quests.Quest) []QuestValidationIssue {
	var issues []QuestValidationIssue
	addError := func(path, code, message string) {
		issues = append(issues, QuestValidationIssue{
			Severity: "error",
			Path:     path,
			Code:     code,
			Message:  message,
		})
	}

	if quest == nil {
		addError("", "missing_quest", "Quest payload is required.")
		return issues
	}

	if strings.TrimSpace(quest.Name) == "" {
		addError("name", "missing_name", "Quest name is required.")
	}
	if strings.TrimSpace(quest.Description) == "" {
		addError("description", "missing_description", "Quest description is required.")
	}

	switch quest.Source.Type {
	case "npc":
		if strings.TrimSpace(quest.Source.NPCID) == "" {
			addError("source.npcId", "missing_source_npc", "NPC source quests require a source NPC.")
		} else if !s.npcExists(quest.Source.NPCID) {
			addError("source.npcId", "missing_source_npc", "Source NPC does not exist.")
		}
	case "item":
		if strings.TrimSpace(quest.Source.ItemID) == "" {
			addError("source.itemId", "missing_source_item", "Item source quests require a source item.")
		} else if !s.itemExists(quest.Source.ItemID) {
			addError("source.itemId", "missing_source_item", "Source item does not exist.")
		}
	case "auto", "script":
	default:
		addError("source.type", "invalid_source_type", "Quest source type must be npc, item, auto, or script.")
	}

	if len(quest.Objectives) == 0 {
		addError("objectives", "missing_objectives", "At least one objective is required.")
	}

	objectiveIDs := map[string]bool{}
	for i, obj := range quest.Objectives {
		path := "objectives"
		if obj.ID == "" {
			addError(path, "missing_objective_id", "Objective ID is required.")
		} else if objectiveIDs[obj.ID] {
			addError(path, "duplicate_objective_id", "Objective IDs must be unique within a quest.")
		}
		objectiveIDs[obj.ID] = true

		if strings.TrimSpace(obj.Description) == "" {
			addError(path, "missing_objective_description", "Objective description is required.")
		}
		if obj.Amount < 1 {
			addError(path, "invalid_objective_amount", "Objective amount must be at least 1.")
		}

		switch obj.Type {
		case quests.ObjectiveKill:
			if strings.TrimSpace(obj.TargetID) == "" || !s.npcExists(obj.TargetID) {
				addError(path, "missing_objective_npc", "Kill objectives require an existing NPC target.")
			}
		case quests.ObjectiveCollect:
			if strings.TrimSpace(obj.TargetID) == "" || !s.itemExists(obj.TargetID) {
				addError(path, "missing_objective_item", "Collect objectives require an existing item template target.")
			}
		case quests.ObjectiveDeliver:
			if strings.TrimSpace(obj.TargetID) == "" || !s.itemExists(obj.TargetID) {
				addError(path, "missing_objective_item", "Deliver objectives require an existing item template target.")
			}
			if strings.TrimSpace(obj.DeliverToNPCID) == "" || !s.npcExists(obj.DeliverToNPCID) {
				addError(path, "missing_delivery_npc", "Deliver objectives require an existing delivery NPC.")
			}
		case quests.ObjectiveVisit:
			if strings.TrimSpace(obj.TargetID) == "" || !s.roomExists(obj.TargetID) {
				addError(path, "missing_objective_room", "Visit objectives require an existing room target.")
			}
		case quests.ObjectiveTalk:
			if strings.TrimSpace(obj.TargetID) == "" || !s.npcExists(obj.TargetID) {
				addError(path, "missing_objective_npc", "Talk objectives require an existing NPC target.")
			}
		case quests.ObjectiveCustom:
			if strings.TrimSpace(obj.CheckScriptID) == "" || !s.scriptExists(obj.CheckScriptID) {
				addError(path, "missing_check_script", "Custom objectives require an existing check script.")
			}
		default:
			addError(path, "invalid_objective_type", "Objective type is invalid.")
		}

		_ = i
	}

	anywhere, turnInNPC := quest.ResolveTurnIn()
	if !anywhere && strings.TrimSpace(turnInNPC) == "" {
		addError("turnIn", "missing_turn_in_npc", "Quest needs turnIn: anywhere, an NPC source, or a deliver objective with a delivery NPC.")
	} else if strings.TrimSpace(turnInNPC) != "" && !s.npcExists(turnInNPC) {
		addError("turnIn", "missing_turn_in_npc", "Turn-in NPC does not exist.")
	}

	for _, itemID := range quest.Rewards.ItemTemplateIDs {
		if strings.TrimSpace(itemID) == "" || !s.itemExists(itemID) {
			addError("rewards.itemTemplateIds", "missing_reward_item", "Reward item template does not exist.")
		}
	}

	for _, reqID := range quest.RequiredQuestIDs {
		if reqID == quest.ID && quest.ID != "" {
			addError("requiredQuestIds", "self_prerequisite", "A quest cannot require itself.")
			continue
		}
		if strings.TrimSpace(reqID) == "" || !s.questExists(reqID) {
			addError("requiredQuestIds", "missing_required_quest", "Required quest does not exist.")
		}
	}

	return issues
}

func (s *questsService) npcExists(id string) bool {
	if s.facade == nil || s.facade.NPCsService() == nil {
		return strings.TrimSpace(id) != ""
	}
	found, err := s.facade.NPCsService().FindByID(id)
	return err == nil && found != nil
}

func (s *questsService) itemExists(id string) bool {
	if s.facade == nil || s.facade.ItemsService() == nil {
		return strings.TrimSpace(id) != ""
	}
	found, err := s.facade.ItemsService().FindByID(id)
	return err == nil && found != nil
}

func (s *questsService) roomExists(id string) bool {
	if s.facade == nil || s.facade.RoomsService() == nil {
		return strings.TrimSpace(id) != ""
	}
	found, err := s.facade.RoomsService().FindByID(id)
	return err == nil && found != nil
}

func (s *questsService) scriptExists(id string) bool {
	if s.facade == nil || s.facade.ScriptsService() == nil {
		return strings.TrimSpace(id) != ""
	}
	found, err := s.facade.ScriptsService().FindByID(id)
	return err == nil && found != nil
}

func (s *questsService) questExists(id string) bool {
	found, err := s.questsRepo.FindByID(id)
	return err == nil && found != nil
}

// --- Quest progress operations ---

func (s *questsService) GetQuestLog(characterID string) ([]*quests.QuestProgress, error) {
	return s.progressRepo.FindByCharacterID(characterID)
}

func (s *questsService) BuildQuestLog(characterID string) ([]QuestLogEntry, error) {
	progressList, err := s.GetQuestLog(characterID)
	if err != nil {
		return nil, err
	}

	entries := make([]QuestLogEntry, 0, len(progressList))
	for _, progress := range progressList {
		quest, err := s.FindByID(progress.QuestID)
		if err != nil || quest == nil {
			continue
		}

		entry := QuestLogEntry{
			QuestID:       progress.QuestID,
			QuestName:     quest.Name,
			Description:   quest.Description,
			Category:      quest.Category,
			Level:         quest.Level,
			Status:        string(progress.Status),
			ReadyToTurnIn: progress.Status == quests.QuestStatusActive && allObjectivesComplete(progress.Objectives),
			Objectives:    buildQuestObjectiveEntries(quest, progress),
			Rewards: &QuestRewardEntry{
				XP:              quest.Rewards.XP,
				Gold:            quest.Rewards.Gold,
				ItemTemplateIDs: quest.Rewards.ItemTemplateIDs,
			},
		}
		applyQuestLogTurnIn(&entry, quest, s.facade)
		if !progress.AcceptedAt.IsZero() {
			entry.AcceptedAt = progress.AcceptedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if !progress.CompletedAt.IsZero() {
			entry.CompletedAt = progress.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func allObjectivesComplete(objectives []quests.ObjectiveProgress) bool {
	if len(objectives) == 0 {
		return false
	}
	for _, obj := range objectives {
		if !obj.Completed {
			return false
		}
	}
	return true
}

func buildQuestObjectiveEntries(quest *quests.Quest, progress *quests.QuestProgress) []QuestObjectiveProgressEntry {
	entries := make([]QuestObjectiveProgressEntry, 0, len(progress.Objectives))
	for _, objProgress := range progress.Objectives {
		entry := QuestObjectiveProgressEntry{
			ObjectiveID: objProgress.ObjectiveID,
			Current:     objProgress.Current,
			Required:    objProgress.Required,
			Completed:   objProgress.Completed,
		}
		for _, objective := range quest.Objectives {
			if objective.ID == objProgress.ObjectiveID {
				entry.Description = objective.Description
				if objective.Amount > 0 {
					entry.Required = objective.Amount
				}
				break
			}
		}
		if entry.Required < 1 {
			entry.Required = 1
		}
		entries = append(entries, entry)
	}
	return entries
}

func (s *questsService) GetProgress(characterID, questID string) (*quests.QuestProgress, error) {
	return s.progressRepo.FindByCharacterAndQuest(characterID, questID)
}

func (s *questsService) AcceptQuest(characterID, questID string) (*quests.QuestProgress, error) {
	// Check if quest exists
	quest, err := s.questsRepo.FindByID(questID)
	if err != nil {
		return nil, errors.New("quest not found")
	}

	// Check if already accepted
	existing, _ := s.progressRepo.FindByCharacterAndQuest(characterID, questID)
	if existing != nil && existing.Status == quests.QuestStatusActive {
		return nil, errors.New("quest already active")
	}

	// Check if completed and not repeatable
	if existing != nil && existing.Status == quests.QuestStatusCompleted && !quest.Repeatable {
		return nil, errors.New("quest already completed")
	}

	// Check prerequisite quests are completed
	for _, reqID := range quest.RequiredQuestIDs {
		reqProgress, _ := s.progressRepo.FindByCharacterAndQuest(characterID, reqID)
		if reqProgress == nil || reqProgress.Status != quests.QuestStatusCompleted {
			return nil, errors.New("prerequisite quest not completed")
		}
	}

	// If previously completed/abandoned and repeatable, delete old progress
	if existing != nil {
		_ = s.progressRepo.Delete(existing.ID)
	}

	// Build objective progress from quest definition
	objectives := make([]quests.ObjectiveProgress, len(quest.Objectives))
	for i, obj := range quest.Objectives {
		required := obj.Amount
		if required < 1 {
			required = 1
		}
		objectives[i] = quests.ObjectiveProgress{
			ObjectiveID: obj.ID,
			Current:     0,
			Required:    required,
			Completed:   false,
		}
	}

	// Pre-fill collect objectives if items are already in inventory
	if s.facade != nil {
		char, charErr := s.facade.CharactersService().FindByID(characterID)
		if charErr == nil && char != nil {
			for i, obj := range quest.Objectives {
				if obj.Type != quests.ObjectiveCollect || obj.TargetID == "" {
					continue
				}
				// Count matching items in inventory
				var count int32
				for _, item := range char.Inventory.Items {
					tid := item.TemplateID
					if tid == "" {
						tid = item.ID
					}
					if tid == obj.TargetID {
						if item.Quantity > 0 {
							count += item.Quantity
						} else {
							count++
						}
					}
				}
				if count > 0 {
					if count >= objectives[i].Required {
						count = objectives[i].Required
						objectives[i].Completed = true
					}
					objectives[i].Current = count
				}
			}
		}
	}

	progress := &quests.QuestProgress{
		CharacterID: characterID,
		QuestID:     questID,
		Status:      quests.QuestStatusActive,
		Objectives:  objectives,
		AcceptedAt:  time.Now(),
	}

	return s.progressRepo.Store(progress)
}

func (s *questsService) AbandonQuest(characterID, questID string) error {
	progress, err := s.progressRepo.FindByCharacterAndQuest(characterID, questID)
	if err != nil || progress == nil {
		return errors.New("quest not found in quest log")
	}
	if progress.Status != quests.QuestStatusActive {
		return errors.New("quest is not active")
	}

	progress.Status = quests.QuestStatusAbandoned
	return s.progressRepo.Update(progress.ID, progress)
}

func (s *questsService) UpdateProgress(progress *quests.QuestProgress) error {
	return s.progressRepo.Update(progress.ID, progress)
}

func (s *questsService) CompleteQuest(characterID, questID string) (*quests.QuestProgress, error) {
	progress, err := s.progressRepo.FindByCharacterAndQuest(characterID, questID)
	if err != nil || progress == nil {
		return nil, errors.New("quest not found in quest log")
	}
	if progress.Status != quests.QuestStatusActive {
		return nil, errors.New("quest is not active")
	}

	// Verify all objectives are completed
	allComplete, _ := s.CheckObjectives(characterID, questID)
	if !allComplete {
		return nil, errors.New("not all objectives are completed")
	}

	progress.Status = quests.QuestStatusCompleted
	progress.CompletedAt = time.Now()

	if err := s.progressRepo.Update(progress.ID, progress); err != nil {
		return nil, err
	}

	return progress, nil
}

// GrantAutoQuests accepts every auto-source quest that belongs to the
// character's current room area (Z00 catacombs -> QST0001-QST0004).
func (s *questsService) GrantAutoQuests(characterID, roomArea string) int {
	allQuests, err := s.FindAll()
	if err != nil {
		return 0
	}
	granted := 0
	for _, q := range allQuests {
		if q == nil || !strings.EqualFold(q.Source.Type, "auto") {
			continue
		}
		if !s.autoQuestBelongsToArea(q, roomArea) {
			continue
		}
		if _, err := s.AcceptQuest(characterID, q.ID); err != nil {
			log.WithError(err).WithField("questID", q.ID).Debug("auto-grant skipped")
			continue
		}
		granted++
	}
	return granted
}

func (s *questsService) autoQuestBelongsToArea(q *quests.Quest, roomArea string) bool {
	if q == nil {
		return false
	}
	if strings.HasPrefix(roomArea, "Z00") && strings.HasPrefix(q.ID, "QST000") {
		return true
	}
	if strings.HasPrefix(roomArea, "Z01") && strings.HasPrefix(q.ID, "QST010") {
		return true
	}
	if s.facade == nil || roomArea == "" {
		return false
	}
	for _, obj := range q.Objectives {
		if obj.Type != quests.ObjectiveVisit || obj.TargetID == "" {
			continue
		}
		room, err := s.facade.RoomsService().FindByID(obj.TargetID)
		if err == nil && room != nil && room.Area == roomArea {
			return true
		}
	}
	return false
}

func (s *questsService) GetAvailableQuests(characterID string) ([]*quests.Quest, error) {
	allQuests, err := s.questsRepo.FindAll()
	if err != nil {
		return nil, err
	}

	progressList, err := s.progressRepo.FindByCharacterID(characterID)
	if err != nil {
		return nil, err
	}

	// Build a map of quest status for this character
	questStatus := make(map[string]quests.QuestStatus)
	for _, p := range progressList {
		questStatus[p.QuestID] = p.Status
	}

	available := make([]*quests.Quest, 0)
	for _, q := range allQuests {
		status, exists := questStatus[q.ID]
		if !exists {
			// Never started - check prerequisites
			if s.checkPrereqs(q, questStatus) {
				available = append(available, q)
			}
		} else if status == quests.QuestStatusAbandoned || (status == quests.QuestStatusCompleted && q.Repeatable) {
			if s.checkPrereqs(q, questStatus) {
				available = append(available, q)
			}
		}
	}

	return available, nil
}

func (s *questsService) checkPrereqs(quest *quests.Quest, questStatus map[string]quests.QuestStatus) bool {
	for _, reqID := range quest.RequiredQuestIDs {
		if questStatus[reqID] != quests.QuestStatusCompleted {
			return false
		}
	}
	return true
}

func (s *questsService) ApplyQuestEvent(event QuestEvent) ([]QuestEventResult, error) {
	if strings.TrimSpace(event.CharacterID) == "" {
		return nil, errors.New("character id is required")
	}

	progressList, err := s.GetQuestLog(event.CharacterID)
	if err != nil {
		return nil, err
	}

	results := make([]QuestEventResult, 0)
	for _, progress := range progressList {
		if progress == nil || progress.Status != quests.QuestStatusActive {
			continue
		}

		quest, err := s.FindByID(progress.QuestID)
		if err != nil || quest == nil {
			continue
		}

		for _, objective := range quest.Objectives {
			if !questEventMatchesObjective(event, objective) || questObjectiveAlreadyComplete(progress, objective.ID) {
				continue
			}

			updated, err := s.applyQuestEventObjective(event, progress, objective)
			if err != nil {
				if objective.Type == quests.ObjectiveDeliver {
					continue
				}
				return results, err
			}
			if updated == nil {
				continue
			}

			kind := QuestEventResultProgress
			if allObjectivesComplete(updated.Objectives) {
				kind = QuestEventResultReadyToTurnIn
			}
			results = append(results, QuestEventResult{
				Kind:        kind,
				QuestID:     quest.ID,
				QuestName:   quest.Name,
				ObjectiveID: objective.ID,
				Progress:    updated,
			})
		}
	}

	return results, nil
}

func (s *questsService) applyQuestEventObjective(event QuestEvent, progress *quests.QuestProgress, objective quests.Objective) (*quests.QuestProgress, error) {
	switch objective.Type {
	case quests.ObjectiveDeliver:
		return s.CompleteDeliveryObjective(event.CharacterID, progress.QuestID, objective.ID)
	default:
		return s.IncrementObjective(event.CharacterID, progress.QuestID, objective.ID, questEventAmount(event))
	}
}

func questEventMatchesObjective(event QuestEvent, objective quests.Objective) bool {
	switch event.Type {
	case QuestEventNPCKilled:
		return objective.Type == quests.ObjectiveKill && objective.TargetID == event.NPCID
	case QuestEventItemPickup:
		return objective.Type == quests.ObjectiveCollect && objective.TargetID == questEventItemTemplateID(event)
	case QuestEventRoomEnter:
		return objective.Type == quests.ObjectiveVisit && objective.TargetID == event.RoomID
	case QuestEventDialogNode:
		return objective.Type == quests.ObjectiveTalk &&
			(objective.TargetID == "" || objective.TargetID == event.NPCID) &&
			(objective.DialogNodeID == "" || objective.DialogNodeID == event.DialogNodeID)
	case QuestEventTalkToNPC:
		switch objective.Type {
		case quests.ObjectiveTalk:
			return objective.TargetID == "" || objective.TargetID == event.NPCID
		case quests.ObjectiveDeliver:
			return objective.DeliverToNPCID == event.NPCID
		}
	}
	return false
}

func questEventAmount(event QuestEvent) int32 {
	if event.Amount > 0 {
		return event.Amount
	}
	if event.Item != nil && event.Item.Quantity > 0 {
		return event.Item.Quantity
	}
	return 1
}

func questEventItemTemplateID(event QuestEvent) string {
	if event.ItemTemplateID != "" {
		return event.ItemTemplateID
	}
	if event.Item == nil {
		return ""
	}
	if event.Item.TemplateID != "" {
		return event.Item.TemplateID
	}
	return event.Item.ID
}

func questObjectiveAlreadyComplete(progress *quests.QuestProgress, objectiveID string) bool {
	for _, obj := range progress.Objectives {
		if obj.ObjectiveID == objectiveID {
			return obj.Completed
		}
	}
	return false
}

func (s *questsService) IncrementObjective(characterID, questID, objectiveID string, amount int32) (*quests.QuestProgress, error) {
	progress, err := s.progressRepo.FindByCharacterAndQuest(characterID, questID)
	if err != nil || progress == nil {
		return nil, errors.New("quest not found in quest log")
	}
	if progress.Status != quests.QuestStatusActive {
		return nil, errors.New("quest is not active")
	}

	for i, obj := range progress.Objectives {
		if obj.ObjectiveID == objectiveID && !obj.Completed {
			progress.Objectives[i].Current += amount
			if progress.Objectives[i].Current >= progress.Objectives[i].Required {
				progress.Objectives[i].Current = progress.Objectives[i].Required
				progress.Objectives[i].Completed = true
			}
			break
		}
	}

	if err := s.progressRepo.Update(progress.ID, progress); err != nil {
		log.WithError(err).Error("Failed to update quest progress")
		return nil, err
	}

	return progress, nil
}

func (s *questsService) CompleteDeliveryObjective(characterID, questID, objectiveID string) (*quests.QuestProgress, error) {
	if s.facade == nil {
		return nil, errors.New("facade not initialized")
	}

	quest, err := s.questsRepo.FindByID(questID)
	if err != nil || quest == nil {
		return nil, errors.New("quest not found")
	}

	var objective *quests.Objective
	for i := range quest.Objectives {
		if quest.Objectives[i].ID == objectiveID && quest.Objectives[i].Type == quests.ObjectiveDeliver {
			objective = &quest.Objectives[i]
			break
		}
	}
	if objective == nil {
		return nil, errors.New("delivery objective not found")
	}

	character, err := s.facade.CharactersService().FindByID(characterID)
	if err != nil || character == nil {
		return nil, errors.New("character not found")
	}

	required := objective.Amount
	if required < 1 {
		required = 1
	}
	if character.Inventory.CountMatchingTemplate(objective.TargetID) < required {
		return nil, errors.New("required delivery item is not in inventory")
	}
	if err := character.Inventory.ConsumeMatchingTemplate(objective.TargetID, required); err != nil {
		return nil, err
	}
	if err := s.facade.CharactersService().Update(character.ID, character); err != nil {
		return nil, err
	}

	return s.IncrementObjective(characterID, questID, objectiveID, required)
}

func (s *questsService) CheckObjectives(characterID, questID string) (bool, error) {
	progress, err := s.progressRepo.FindByCharacterAndQuest(characterID, questID)
	if err != nil || progress == nil {
		return false, errors.New("quest not found")
	}

	for _, obj := range progress.Objectives {
		if !obj.Completed {
			return false, nil
		}
	}
	return true, nil
}

func (s *questsService) TurnInQuest(characterID, questID, npcID string) (*QuestTurnInResult, error) {
	quest, err := s.FindByID(questID)
	if err != nil || quest == nil {
		return nil, errors.New("quest not found")
	}
	if !questCanTurnInAtNPC(quest, npcID) {
		return nil, errors.New("quest cannot be turned in to this NPC")
	}

	progress, err := s.CompleteQuest(characterID, questID)
	if err != nil {
		return nil, err
	}

	grantedItems, err := s.GrantQuestRewards(characterID, questID)
	if err != nil {
		return nil, err
	}

	return &QuestTurnInResult{
		QuestID:       quest.ID,
		QuestName:     quest.Name,
		GrantedItems:  grantedItems,
		XP:            quest.Rewards.XP,
		Gold:          quest.Rewards.Gold,
		CompletedAt:   progress.CompletedAt,
		QuestProgress: progress,
	}, nil
}

func (s *questsService) TurnInQuestAnywhere(characterID, questID string) (*QuestTurnInResult, error) {
	quest, err := s.FindByID(questID)
	if err != nil || quest == nil {
		return nil, errors.New("quest not found")
	}
	if !quest.AllowsAnywhereTurnIn() {
		npcID := quest.TurnInNPCID()
		if npcID == "" {
			return nil, errors.New("this quest must be turned in at an NPC")
		}
		name := s.npcName(npcID)
		if name == "" {
			name = npcID
		}
		return nil, errors.New("turn this quest in at " + name)
	}

	progress, err := s.CompleteQuest(characterID, questID)
	if err != nil {
		return nil, err
	}

	grantedItems, err := s.GrantQuestRewards(characterID, questID)
	if err != nil {
		return nil, err
	}

	return &QuestTurnInResult{
		QuestID:       quest.ID,
		QuestName:     quest.Name,
		GrantedItems:  grantedItems,
		XP:            quest.Rewards.XP,
		Gold:          quest.Rewards.Gold,
		CompletedAt:   progress.CompletedAt,
		QuestProgress: progress,
	}, nil
}

func questCanTurnInAtNPC(quest *quests.Quest, npcID string) bool {
	if strings.TrimSpace(npcID) == "" || quest == nil {
		return false
	}
	if quest.AllowsAnywhereTurnIn() {
		return false
	}
	return quest.TurnInNPCID() == npcID
}

func applyQuestLogTurnIn(entry *QuestLogEntry, quest *quests.Quest, facade Facade) {
	if entry == nil || quest == nil {
		return
	}
	anywhere, npcID := quest.ResolveTurnIn()
	entry.TurnInAnywhere = anywhere
	entry.TurnInNpcID = npcID
	if npcID != "" && facade != nil && facade.NPCsService() != nil {
		if npc, err := facade.NPCsService().FindByID(npcID); err == nil && npc != nil {
			entry.TurnInNpcName = npc.Name
		}
	}
}

func (s *questsService) npcName(id string) string {
	if s.facade == nil || s.facade.NPCsService() == nil {
		return ""
	}
	npc, err := s.facade.NPCsService().FindByID(id)
	if err != nil || npc == nil {
		return ""
	}
	return npc.Name
}

// GrantQuestRewards awards XP, gold, and items to character upon quest completion
// Returns list of granted item names
func (s *questsService) GrantQuestRewards(characterID, questID string) ([]string, error) {
	if s.facade == nil {
		return nil, errors.New("facade not initialized")
	}

	// 1. Get quest definition to retrieve rewards
	quest, err := s.FindByID(questID)
	if err != nil {
		return nil, err
	}

	// 2. Get character
	char, err := s.facade.CharactersService().FindByID(characterID)
	if err != nil {
		return nil, err
	}

	// 3. Award XP
	if quest.Rewards.XP > 0 {
		char.XP += quest.Rewards.XP
		log.WithFields(log.Fields{
			"characterID": characterID,
			"questID":     questID,
			"xpAwarded":   quest.Rewards.XP,
		}).Info("Awarded quest XP")
	}

	// 4. Award Gold
	if quest.Rewards.Gold > 0 {
		char.Gold += quest.Rewards.Gold
		log.WithFields(log.Fields{
			"characterID": characterID,
			"questID":     questID,
			"goldAwarded": quest.Rewards.Gold,
		}).Info("Awarded quest gold")
	}

	// 5. Award Items (create instances from templates)
	grantedItems := []string{}
	for _, templateID := range quest.Rewards.ItemTemplateIDs {
		instance, err := s.facade.ItemsService().CreateInstanceFromTemplate(templateID)
		if err != nil {
			log.WithError(err).WithField("templateID", templateID).Error("Failed to create item instance from template")
			continue
		}
		char.Inventory.Items = append(char.Inventory.Items, instance)
		grantedItems = append(grantedItems, instance.Name)
		log.WithFields(log.Fields{
			"characterID": characterID,
			"questID":     questID,
			"itemName":    instance.Name,
		}).Info("Awarded quest item")
	}

	// 6. Increment quests completed stat
	char.AllTimeStats.QuestsCompleted++

	// 7. Save updated character
	err = s.facade.CharactersService().Update(characterID, char)
	if err != nil {
		return nil, err
	}

	// 8. Execute OnCompleteScriptID if defined
	if quest.OnCompleteScriptID != "" {
		// Note: Script execution would be triggered here
		// This would typically be handled by the game layer that has access to script runner
		log.WithFields(log.Fields{
			"characterID": characterID,
			"questID":     questID,
			"scriptID":    quest.OnCompleteScriptID,
		}).Info("Quest completion script needs execution (handled by game layer)")
	}

	return grantedItems, nil
}
