package service

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
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
	GetProgress(characterID, questID string) (*quests.QuestProgress, error)
	AcceptQuest(characterID, questID string) (*quests.QuestProgress, error)
	AbandonQuest(characterID, questID string) error
	UpdateProgress(progress *quests.QuestProgress) error
	CompleteQuest(characterID, questID string) (*quests.QuestProgress, error)
	GetAvailableQuests(characterID string) ([]*quests.Quest, error)

	// Objective progress (called by QuestTracker)
	IncrementObjective(characterID, questID, objectiveID string, amount int32) (*quests.QuestProgress, error)
	CheckObjectives(characterID, questID string) (bool, error)

	// Reward granting
	GrantQuestRewards(characterID, questID string) ([]string, error)

	// Internal setter for facade dependency
	SetFacade(facade Facade)
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

// --- Quest progress operations ---

func (s *questsService) GetQuestLog(characterID string) ([]*quests.QuestProgress, error) {
	return s.progressRepo.FindByCharacterID(characterID)
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
