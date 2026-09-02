package commands

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// QuestLogCommand shows the player's active quests
type QuestLogCommand struct{}

// Key returns the command key matcher
func (cmd *QuestLogCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute handles the quests/ql command
func (cmd *QuestLogCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	progressList, err := game.GetFacade().QuestsService().GetQuestLog(message.Character.ID)
	if err != nil {
		game.SendMessage() <- message.Reply("Error loading quest log.")
		return true
	}

	// Filter to active quests
	var active []*quests.QuestProgress
	var completed []*quests.QuestProgress
	for _, p := range progressList {
		if p.Status == quests.QuestStatusActive {
			active = append(active, p)
		} else if p.Status == quests.QuestStatusCompleted {
			completed = append(completed, p)
		}
	}

	if len(active) == 0 && len(completed) == 0 {
		game.SendMessage() <- message.Reply("Your quest log is empty.")
		return true
	}

	var sb strings.Builder
	sb.WriteString("\n--- Quest Log ---\n")

	if len(active) > 0 {
		sb.WriteString("\nActive Quests:\n")
		for _, p := range active {
			quest, err := game.GetFacade().QuestsService().FindByID(p.QuestID)
			if err != nil || quest == nil {
				continue
			}
			sb.WriteString(fmt.Sprintf("  [*] %s\n", quest.Name))
			for _, obj := range quest.Objectives {
				for _, op := range p.Objectives {
					if op.ObjectiveID == obj.ID {
						checkmark := " "
						if op.Completed {
							checkmark = "x"
						}
						sb.WriteString(fmt.Sprintf("      [%s] %s (%d/%d)\n", checkmark, obj.Description, op.Current, op.Required))
					}
				}
			}
		}
	}

	if len(completed) > 0 {
		sb.WriteString(fmt.Sprintf("\nCompleted: %d quest(s)\n", len(completed)))
	}

	game.SendMessage() <- message.Reply(sb.String())
	return true
}

// QuestDetailCommand shows details of a specific quest
type QuestDetailCommand struct{}

// Key returns the command key matcher
func (cmd *QuestDetailCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the quest <name> command
func (cmd *QuestDetailCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Usage: quest <name>")
		return true
	}

	questName := strings.Join(parts[1:], " ")
	questNameLower := strings.ToLower(questName)

	progressList, err := game.GetFacade().QuestsService().GetQuestLog(message.Character.ID)
	if err != nil {
		game.SendMessage() <- message.Reply("Error loading quest log.")
		return true
	}

	for _, p := range progressList {
		quest, err := game.GetFacade().QuestsService().FindByID(p.QuestID)
		if err != nil || quest == nil {
			continue
		}

		if strings.Contains(strings.ToLower(quest.Name), questNameLower) {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("\n--- %s ---\n", quest.Name))
			sb.WriteString(fmt.Sprintf("Status: %s\n", string(p.Status)))
			if quest.Category != "" {
				sb.WriteString(fmt.Sprintf("Category: %s\n", quest.Category))
			}
			if quest.Level > 0 {
				sb.WriteString(fmt.Sprintf("Level: %d\n", quest.Level))
			}
			sb.WriteString(fmt.Sprintf("\n%s\n", quest.Description))

			sb.WriteString("\nObjectives:\n")
			for _, obj := range quest.Objectives {
				for _, op := range p.Objectives {
					if op.ObjectiveID == obj.ID {
						checkmark := " "
						if op.Completed {
							checkmark = "x"
						}
						sb.WriteString(fmt.Sprintf("  [%s] %s (%d/%d)\n", checkmark, obj.Description, op.Current, op.Required))
					}
				}
			}

			if quest.Rewards.XP > 0 || quest.Rewards.Gold > 0 {
				sb.WriteString("\nRewards:\n")
				if quest.Rewards.XP > 0 {
					sb.WriteString(fmt.Sprintf("  + %d XP\n", quest.Rewards.XP))
				}
				if quest.Rewards.Gold > 0 {
					sb.WriteString(fmt.Sprintf("  + %d Gold\n", quest.Rewards.Gold))
				}
			}

			game.SendMessage() <- message.Reply(sb.String())
			return true
		}
	}

	game.SendMessage() <- message.Reply("Quest '" + questName + "' not found in your quest log.")
	return true
}

// AbandonQuestCommand abandons an active quest
type AbandonQuestCommand struct{}

// Key returns the command key matcher
func (cmd *AbandonQuestCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the abandon <quest> command
func (cmd *AbandonQuestCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Usage: abandon <quest name>")
		return true
	}

	questName := strings.Join(parts[1:], " ")
	questNameLower := strings.ToLower(questName)

	progressList, err := game.GetFacade().QuestsService().GetQuestLog(message.Character.ID)
	if err != nil {
		game.SendMessage() <- message.Reply("Error loading quest log.")
		return true
	}

	for _, p := range progressList {
		if p.Status != quests.QuestStatusActive {
			continue
		}

		quest, err := game.GetFacade().QuestsService().FindByID(p.QuestID)
		if err != nil || quest == nil {
			continue
		}

		if strings.Contains(strings.ToLower(quest.Name), questNameLower) {
			err := game.GetFacade().QuestsService().AbandonQuest(message.Character.ID, p.QuestID)
			if err != nil {
				game.SendMessage() <- message.Reply("Failed to abandon quest: " + err.Error())
				return true
			}

			// Send typed questAbandoned message so the client refreshes the quest log
			game.SendMessage() <- messages.MessageResponse{
				Audience:   messages.MessageAudienceOrigin,
				AudienceID: message.FromUser.ID,
				Type:       messages.MessageTypeQuestAbandoned,
				Message:    "Quest abandoned: " + quest.Name,
			}
			return true
		}
	}

	game.SendMessage() <- message.Reply("Active quest '" + questName + "' not found.")
	return true
}

// CompleteQuestCommand turns in an anywhere-turn-in quest from the log.
type CompleteQuestCommand struct{}

func (cmd *CompleteQuestCommand) Key() CommandKey { return &StartsWithCommandKey{} }

func (cmd *CompleteQuestCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Usage: complete <quest name>")
		return true
	}

	questName := strings.Join(parts[1:], " ")
	questNameLower := strings.ToLower(questName)

	progressList, err := game.GetFacade().QuestsService().GetQuestLog(message.Character.ID)
	if err != nil {
		game.SendMessage() <- message.Reply("Error loading quest log.")
		return true
	}

	for _, p := range progressList {
		if p.Status != quests.QuestStatusActive {
			continue
		}
		quest, err := game.GetFacade().QuestsService().FindByID(p.QuestID)
		if err != nil || quest == nil {
			continue
		}
		if !strings.Contains(strings.ToLower(quest.Name), questNameLower) {
			continue
		}

		result, err := game.GetFacade().QuestsService().TurnInQuestAnywhere(message.Character.ID, p.QuestID)
		if err != nil {
			game.SendMessage() <- message.Reply(err.Error())
			return true
		}

		rewardMsg := buildQuestRewardMessage(quest, result.GrantedItems)
		game.SendMessage() <- messages.MessageResponse{
			Audience:   messages.MessageAudienceUser,
			AudienceID: message.Character.BelongsUserID,
			Type:       messages.MessageTypeQuestCompleted,
			Message:    rewardMsg,
		}

		updatedChar, _ := game.GetFacade().CharactersService().FindByID(message.Character.ID)
		if updatedChar != nil {
			game.SendMessage() <- messages.NewCharacterUpdateMessage(message.Character.BelongsUserID, updatedChar)
			if len(result.GrantedItems) > 0 {
				inventoryMsg := &messages.Message{
					FromUser:  message.FromUser,
					Character: updatedChar,
				}
				game.SendMessage() <- messages.NewInventoryUpdateMessage(inventoryMsg)
			}
		}

		questLog, _ := game.GetFacade().QuestsService().GetQuestLog(message.Character.ID)
		game.SendMessage() <- messages.QuestLogMessage{
			MessageResponse: messages.MessageResponse{
				Audience:   messages.MessageAudienceUser,
				AudienceID: message.Character.BelongsUserID,
				Type:       messages.MessageTypeQuestLog,
			},
			Quests: buildQuestLogEntries(game, questLog),
		}
		return true
	}

	game.SendMessage() <- message.Reply("Active quest '" + questName + "' not found.")
	return true
}
