package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/conversations"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// conversationTimeout defines how long a conversation stays "active" after last interaction
const conversationTimeout = 5 * time.Minute

// DialogSelectCommand handles number input during active conversations
// This is registered as a RoomCommand to intercept numeric input
func DialogSelectCommand(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		return false
	}

	// Check if input is a number
	input := strings.TrimSpace(message.Data)
	optionIndex, err := strconv.Atoi(input)
	if err != nil {
		return false // Not a number, let other handlers process it
	}

	// Find active conversation for this character
	convs, err := game.GetFacade().ConversationsService().FindAllForCharacter(message.Character.ID)
	if err != nil {
		return false
	}

	// Find most recent active conversation (within timeout)
	var activeConv *conversations.Conversation
	for _, conv := range convs {
		if time.Since(conv.LastInteracted) < conversationTimeout {
			if activeConv == nil || conv.LastInteracted.After(activeConv.LastInteracted) {
				activeConv = conv
			}
		}
	}

	if activeConv == nil {
		return false // No active conversation, let other handlers process
	}

	if isQuestOnlyConversation(activeConv) {
		return handleQuestOnlyDialogSelection(game, message, activeConv, optionIndex)
	}

	var dialog *dialogs.Dialog
	var filteredOptions []*dialogs.Dialog

	if activeConv.DialogID != "" {
		// Load the dialog
		loadedDialog, err := game.GetFacade().DialogsService().FindByID(activeConv.DialogID)
		if err != nil {
			log.WithError(err).Error("Error loading dialog for conversation")
			return false
		}
		dialog = loadedDialog

		// Get current node
		currentNode := game.GetFacade().ConversationsService().GetCurrentNode(activeConv, dialog)
		if currentNode == nil {
			return false
		}

		// Get filtered options
		filteredOptions = game.GetFacade().ConversationsService().GetFilteredOptions(activeConv, currentNode)
	}

	// Re-compute quest options for this NPC (only at root level)
	var questOptions []questDialogOption
	if activeConv.TargetID != "" && (activeConv.CurrentNodeID == "" || activeConv.CurrentNodeID == "main") {
		questOptions = questOptionsForConversation(game, message.Character.ID, activeConv)
	}

	totalOptions := len(filteredOptions) + len(questOptions)

	// Validate option index (1-based)
	if optionIndex < 1 || optionIndex > totalOptions {
		game.SendMessage() <- message.Reply("Invalid option. Please choose 1-" + strconv.Itoa(totalOptions))
		return true
	}

	// Get NPC name for responses
	npcName := activeConv.Context["NPC"]
	if npcName == "" {
		npcName = "NPC"
	}

	// Check if this is a quest option (index beyond dialog options)
	if optionIndex > len(filteredOptions) {
		questIdx := optionIndex - len(filteredOptions) - 1
		if questIdx < len(questOptions) {
			qo := questOptions[questIdx]
			handleQuestDialogOption(game, message, qo, npcName, activeConv)
		}
		return true
	}

	// Get selected dialog option
	selectedOption := filteredOptions[optionIndex-1]

	// Check if this is a quest option embedded in the dialog tree
	if selectedOption.QuestID != "" && selectedOption.Action != "" {
		handleQuestAction(game, message, selectedOption, npcName, activeConv)
		return true
	}

	// Check if this is a dialog exit
	if selectedOption.IsDialogExit != nil && *selectedOption.IsDialogExit {
		// End conversation
		game.GetFacade().ConversationsService().ResetConversation(activeConv)

		// If there's exit text, show it
		if selectedOption.Text != "" {
			dialogState := &dialogs.DialogState{
				Context: activeConv.Context,
			}
			exitText := selectedOption.Render(dialogState)
			game.SendMessage() <- messages.NewDialogEndMessage(message.FromUser.ID, npcName, exitText)
		} else {
			game.SendMessage() <- messages.NewDialogEndMessage(message.FromUser.ID, npcName, "The conversation has ended.")
		}
		return true
	}

	// Advance conversation to selected node
	targetNodeID := selectedOption.NodeID
	if targetNodeID == "" {
		// If option has no ID, use the text as identifier or generate one
		targetNodeID = "option_" + strconv.Itoa(optionIndex)
	}

	// Mark the selected option as visited
	activeConv.MarkVisited(targetNodeID)
	activeConv.UpdateInteraction()

	// Check if selected option has an Answer (auto-response)
	if selectedOption.Answer != nil {
		// Show the answer and then continue from there
		dialogState := &dialogs.DialogState{
			CurrentDialogID: activeConv.CurrentNodeID,
			DialogVisited:   activeConv.VisitedNodes,
			Context:         activeConv.Context,
		}
		answerText := selectedOption.Answer.Render(dialogState)

		// If the answer is a dialog exit
		if selectedOption.Answer.IsDialogExit != nil && *selectedOption.Answer.IsDialogExit {
			game.GetFacade().ConversationsService().ResetConversation(activeConv)
			game.SendMessage() <- messages.NewDialogEndMessage(message.FromUser.ID, npcName, answerText)
			return true
		}

		// If answer has further options, move to that node
		if len(selectedOption.Answer.Options) > 0 {
			// Set current node to the answer
			if selectedOption.Answer.NodeID != "" {
				activeConv.CurrentNodeID = selectedOption.Answer.NodeID
			}
			game.GetFacade().ConversationsService().Update(activeConv.ID, activeConv)

			// Send the answer with its options
			options := make([]messages.DialogOption, 0)
			answerOptions := game.GetFacade().ConversationsService().GetFilteredOptions(activeConv, selectedOption.Answer)
			for i, opt := range answerOptions {
				optText := opt.Text
				if optText == "" {
					optText = opt.RenderPlain(dialogState)
				}
				options = append(options, messages.DialogOption{
					Index: i + 1,
					Text:  optText,
				})
			}

			dialogMsg := messages.NewDialogMessage(
				message.FromUser.ID,
				npcName,
				answerText,
				options,
				activeConv.ID,
			)
			game.SendMessage() <- dialogMsg
		} else if selectedOption.Answer.NodeID != "" {
			// Answer has no inline options but references a node (back-reference).
			// Navigate to that node and show its options if it has any.
			activeConv.CurrentNodeID = selectedOption.Answer.NodeID
			game.GetFacade().ConversationsService().Update(activeConv.ID, activeConv)

			targetNode := game.GetFacade().ConversationsService().GetCurrentNode(activeConv, dialog)
			if targetNode != nil && len(targetNode.Options) > 0 {
				dialogState := &dialogs.DialogState{
					CurrentDialogID: activeConv.CurrentNodeID,
					DialogVisited:   activeConv.VisitedNodes,
					Context:         activeConv.Context,
				}
				nodeText := targetNode.Render(dialogState)
				options := make([]messages.DialogOption, 0)
				navOptions := game.GetFacade().ConversationsService().GetFilteredOptions(activeConv, targetNode)
				for i, opt := range navOptions {
					optText := opt.Text
					if optText == "" {
						optText = opt.RenderPlain(dialogState)
					}
					options = append(options, messages.DialogOption{
						Index: i + 1,
						Text:  optText,
					})
				}
				dialogMsg := messages.NewDialogMessage(
					message.FromUser.ID,
					npcName,
					nodeText,
					options,
					activeConv.ID,
				)
				game.SendMessage() <- dialogMsg
			} else {
				// Target node not found or has no options — end conversation
				game.SendMessage() <- message.Reply("[" + npcName + "] " + answerText)
				game.GetFacade().ConversationsService().ResetConversation(activeConv)
			}
		} else {
			// Answer has no options and no node reference - just show it and end
			game.SendMessage() <- message.Reply("[" + npcName + "] " + answerText)
			game.GetFacade().ConversationsService().ResetConversation(activeConv)
		}
	} else if len(selectedOption.Options) > 0 {
		// Selected option itself has sub-options - navigate into it
		activeConv.CurrentNodeID = targetNodeID
		game.GetFacade().ConversationsService().Update(activeConv.ID, activeConv)

		// Send the selected option's dialog
		dialogState := &dialogs.DialogState{
			CurrentDialogID: activeConv.CurrentNodeID,
			DialogVisited:   activeConv.VisitedNodes,
			Context:         activeConv.Context,
		}

		nodeText := selectedOption.Render(dialogState)
		options := make([]messages.DialogOption, 0)
		subOptions := game.GetFacade().ConversationsService().GetFilteredOptions(activeConv, selectedOption)
		for i, opt := range subOptions {
			optText := opt.Text
			if optText == "" {
				optText = opt.RenderPlain(dialogState)
			}
			options = append(options, messages.DialogOption{
				Index: i + 1,
				Text:  optText,
			})
		}

		dialogMsg := messages.NewDialogMessage(
			message.FromUser.ID,
			npcName,
			nodeText,
			options,
			activeConv.ID,
		)
		game.SendMessage() <- dialogMsg
	} else {
		// Option has no answer and no sub-options - show text and reset
		dialogState := &dialogs.DialogState{
			Context: activeConv.Context,
		}
		optionText := selectedOption.Render(dialogState)
		game.SendMessage() <- message.Reply("[" + npcName + "] " + optionText)
		game.GetFacade().ConversationsService().ResetConversation(activeConv)
	}

	return true
}

func isQuestOnlyConversation(conv *conversations.Conversation) bool {
	return conv != nil && conv.TargetType == conversations.TargetTypeNPC && (conv.DialogID == questOnlyDialogID || conv.DialogID == "")
}

func handleQuestOnlyDialogSelection(game def.GameCtrl, message *messages.Message, activeConv *conversations.Conversation, optionIndex int) bool {
	questOptions := questOptionsForConversation(game, message.Character.ID, activeConv)
	totalOptions := len(questOptions)
	if optionIndex < 1 || optionIndex > totalOptions {
		game.SendMessage() <- message.Reply("Invalid option. Please choose 1-" + strconv.Itoa(totalOptions))
		return true
	}

	npcName := activeConv.Context["NPC"]
	if npcName == "" {
		npcName = "NPC"
	}

	qo := questOptions[optionIndex-1]
	handleQuestDialogOption(game, message, qo, npcName, activeConv)
	return true
}

func questOptionsForConversation(game def.GameCtrl, characterID string, conv *conversations.Conversation) []questDialogOption {
	if conv == nil || conv.TargetID == "" {
		return nil
	}

	npcManager := game.GetNPCInstanceManager()
	if npcManager == nil {
		return nil
	}

	npcInst := npcManager.GetInstance(conv.TargetID)
	if npcInst == nil {
		return nil
	}

	templateID := npcInst.TemplateID
	if templateID == "" {
		templateID = npcInst.ID
	}
	return getQuestDialogOptions(game, characterID, templateID, npcInst.ID)
}

// handleQuestAction processes quest accept/complete/progress actions from dialog options
func handleQuestAction(game def.GameCtrl, message *messages.Message, selectedOption *dialogs.Dialog, npcName string, activeConv *conversations.Conversation) {
	char := message.Character
	questID := selectedOption.QuestID
	action := selectedOption.Action

	switch action {
	case "accept":
		// Accept the quest
		_, err := game.GetFacade().QuestsService().AcceptQuest(char.ID, questID)
		if err != nil {
			game.SendMessage() <- message.Reply(fmt.Sprintf("[%s] %s", npcName, err.Error()))
			return
		}

		// Get quest details
		quest, _ := game.GetFacade().QuestsService().FindByID(questID)
		if quest == nil {
			game.SendMessage() <- message.Reply("[System] Quest not found.")
			return
		}

		// Send quest accepted message with formatted response
		acceptMsg := fmt.Sprintf("╔══════════════════════════════════════════════════╗\n")
		acceptMsg += fmt.Sprintf("║         QUEST ACCEPTED                            ║\n")
		acceptMsg += fmt.Sprintf("╚══════════════════════════════════════════════════╝\n\n")
		acceptMsg += fmt.Sprintf("%s\n\n", quest.Name)
		acceptMsg += fmt.Sprintf("%s\n\n", quest.Description)
		acceptMsg += fmt.Sprintf("OBJECTIVES:\n")
		for i, obj := range quest.Objectives {
			acceptMsg += fmt.Sprintf("  %d. %s\n", i+1, obj.Description)
		}
		acceptMsg += fmt.Sprintf("\n══════════════════════════════════════════════════")

		game.SendMessage() <- messages.MessageResponse{
			Audience:   messages.MessageAudienceUser,
			AudienceID: char.BelongsUserID,
			Type:       messages.MessageTypeQuestAccepted,
			Message:    acceptMsg,
		}

		// Send quest log update
		questLog, _ := game.GetFacade().QuestsService().GetQuestLog(char.ID)
		questLogEntries := buildQuestLogEntries(game, questLog)
		game.SendMessage() <- messages.QuestLogMessage{
			MessageResponse: messages.MessageResponse{
				Audience:   messages.MessageAudienceUser,
				AudienceID: char.BelongsUserID,
				Type:       messages.MessageTypeQuestLog,
			},
			Quests: questLogEntries,
		}

		log.WithFields(log.Fields{
			"characterID": char.ID,
			"questID":     questID,
			"questName":   quest.Name,
		}).Info("Quest accepted via dialog")

	case "complete":
		// Complete the quest and grant rewards
		_, err := game.GetFacade().QuestsService().CompleteQuest(char.ID, questID)
		if err != nil {
			game.SendMessage() <- message.Reply(fmt.Sprintf("[%s] %s", npcName, err.Error()))
			return
		}

		// Grant rewards
		grantedItems, err := game.GetFacade().QuestsService().GrantQuestRewards(char.ID, questID)
		if err != nil {
			log.WithError(err).Error("Failed to grant quest rewards")
			game.SendMessage() <- message.Reply("[System] Quest completed but failed to grant rewards.")
			return
		}

		// Get quest for reward info
		quest, _ := game.GetFacade().QuestsService().FindByID(questID)
		if quest == nil {
			game.SendMessage() <- message.Reply("[System] Quest not found.")
			return
		}

		// Build reward message
		rewardMsg := buildQuestRewardMessage(quest, grantedItems)

		// Send quest completed message with rewards
		game.SendMessage() <- messages.MessageResponse{
			Audience:   messages.MessageAudienceUser,
			AudienceID: char.BelongsUserID,
			Type:       messages.MessageTypeQuestCompleted,
			Message:    rewardMsg,
		}

		// Get updated character for stats update
		updatedChar, _ := game.GetFacade().CharactersService().FindByID(char.ID)
		if updatedChar != nil {
			// Send character update (XP, gold changed)
			game.SendMessage() <- messages.NewCharacterUpdateMessage(char.BelongsUserID, updatedChar)
		}

		// Send inventory update (items added)
		if len(grantedItems) > 0 && updatedChar != nil {
			inventoryMsg := &messages.Message{
				FromUser:  message.FromUser,
				Character: updatedChar,
			}
			game.SendMessage() <- messages.NewInventoryUpdateMessage(inventoryMsg)
		}

		// Send updated quest log
		questLog, _ := game.GetFacade().QuestsService().GetQuestLog(char.ID)
		questLogEntries := buildQuestLogEntries(game, questLog)
		game.SendMessage() <- messages.QuestLogMessage{
			MessageResponse: messages.MessageResponse{
				Audience:   messages.MessageAudienceUser,
				AudienceID: char.BelongsUserID,
				Type:       messages.MessageTypeQuestLog,
			},
			Quests: questLogEntries,
		}

		log.WithFields(log.Fields{
			"characterID": char.ID,
			"questID":     questID,
			"questName":   quest.Name,
			"rewards":     grantedItems,
		}).Info("Quest completed via dialog")

	case "progress":
		// Just show current progress
		progress, err := game.GetFacade().QuestsService().GetProgress(char.ID, questID)
		if err != nil || progress == nil {
			game.SendMessage() <- message.Reply(fmt.Sprintf("[%s] I don't have any information about that quest.", npcName))
			return
		}

		quest, _ := game.GetFacade().QuestsService().FindByID(questID)
		if quest == nil {
			return
		}

		progressMsg := fmt.Sprintf("[%s] Let me check your progress on \"%s\":\n\n", npcName, quest.Name)
		for i, objProgress := range progress.Objectives {
			obj := quest.Objectives[i]
			if objProgress.Completed {
				progressMsg += fmt.Sprintf("  ✓ %s (Complete)\n", obj.Description)
			} else {
				progressMsg += fmt.Sprintf("  ○ %s (%d/%d)\n", obj.Description, objProgress.Current, objProgress.Required)
			}
		}

		game.SendMessage() <- message.Reply(progressMsg)
	}

	// End dialog after quest action
	game.GetFacade().ConversationsService().ResetConversation(activeConv)
}

// handleQuestDialogOption handles a quest option that was injected into the dialog by the talk command.
// It shows the NPC response text and then performs the quest action (accept/complete/progress).
func handleQuestDialogOption(game def.GameCtrl, message *messages.Message, qo questDialogOption, npcName string, activeConv *conversations.Conversation) {
	// Show NPC response text if available
	if qo.npcText != "" {
		game.SendMessage() <- message.Reply("[" + npcName + "] " + qo.npcText)
	}

	handleQuestAction(game, message, &dialogs.Dialog{QuestID: qo.questID, Action: qo.action}, npcName, activeConv)
}

// buildQuestRewardMessage formats the quest completion reward message
func buildQuestRewardMessage(quest *quests.Quest, grantedItems []string) string {
	var sb strings.Builder

	sb.WriteString("╔══════════════════════════════════════════════════╗\n")
	sb.WriteString("║         QUEST COMPLETED!                          ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════╝\n\n")

	sb.WriteString(fmt.Sprintf("%s\n\n", quest.Name))

	sb.WriteString("REWARDS:\n")
	if quest.Rewards.XP > 0 {
		sb.WriteString(fmt.Sprintf("  + %d XP\n", quest.Rewards.XP))
	}
	if quest.Rewards.Gold > 0 {
		sb.WriteString(fmt.Sprintf("  + %d Gold\n", quest.Rewards.Gold))
	}
	if len(grantedItems) > 0 {
		sb.WriteString("  Items:\n")
		for _, itemName := range grantedItems {
			sb.WriteString(fmt.Sprintf("    - %s\n", itemName))
		}
	}

	sb.WriteString("\n══════════════════════════════════════════════════")

	return sb.String()
}
