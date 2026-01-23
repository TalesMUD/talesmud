package commands

import (
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/conversations"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
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

	// Load the dialog
	dialog, err := game.GetFacade().DialogsService().FindByID(activeConv.DialogID)
	if err != nil {
		log.WithError(err).Error("Error loading dialog for conversation")
		return false
	}

	// Get current node
	currentNode := game.GetFacade().ConversationsService().GetCurrentNode(activeConv, dialog)
	if currentNode == nil {
		return false
	}

	// Get filtered options
	filteredOptions := game.GetFacade().ConversationsService().GetFilteredOptions(activeConv, currentNode)

	// Validate option index (1-based)
	if optionIndex < 1 || optionIndex > len(filteredOptions) {
		game.SendMessage() <- message.Reply("Invalid option. Please choose 1-" + strconv.Itoa(len(filteredOptions)))
		return true
	}

	// Get selected option
	selectedOption := filteredOptions[optionIndex-1]

	// Get NPC name for responses
	npcName := activeConv.Context["NPC"]
	if npcName == "" {
		npcName = "NPC"
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
		} else {
			// Answer has no options - just show it and end or return to parent
			game.SendMessage() <- message.Reply("[" + npcName + "] " + answerText)

			// Reset to main for next conversation
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
