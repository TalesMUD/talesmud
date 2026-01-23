package commands

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/conversations"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// TalkCommand handles talking to NPCs
type TalkCommand struct {
}

// Key returns the command key matcher
func (command *TalkCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the talk command
func (command *TalkCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	if message.Character.CurrentRoomID == "" {
		game.SendMessage() <- message.Reply("You are not in a room.")
		return true
	}

	// Parse NPC name from command: "talk guard" or "talk to guard"
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Talk to whom? Usage: talk <npc-name>")
		return true
	}

	// Handle "talk to <name>" or "talk <name>"
	npcName := strings.Join(parts[1:], " ")
	if strings.HasPrefix(strings.ToLower(npcName), "to ") {
		npcName = strings.TrimPrefix(npcName, "to ")
		npcName = strings.TrimPrefix(npcName, "To ")
	}

	// Find NPC in current room via the NPC instance manager
	npcManager := game.GetNPCInstanceManager()
	if npcManager == nil {
		game.SendMessage() <- message.Reply("Error: NPC system not available.")
		return true
	}

	npc := npcManager.FindInstanceByNameInRoom(message.Character.CurrentRoomID, npcName)
	if npc == nil {
		game.SendMessage() <- message.Reply("There is no one named '" + npcName + "' here.")
		return true
	}

	// Check if NPC has a dialog
	if !npc.HasDialog() {
		game.SendMessage() <- message.Reply(npc.Name + " doesn't seem to want to talk.")
		return true
	}

	// Load the dialog
	dialog, err := game.GetFacade().DialogsService().FindByID(npc.DialogID)
	if err != nil {
		log.WithError(err).WithField("dialogID", npc.DialogID).Error("Error loading NPC dialog")
		game.SendMessage() <- message.Reply(npc.Name + " seems confused and doesn't respond.")
		return true
	}

	// Get or create conversation state
	conv, err := game.GetFacade().ConversationsService().GetOrCreateConversation(
		message.Character.ID,
		npc.ID,
		conversations.TargetTypeNPC,
		npc.DialogID,
	)
	if err != nil {
		log.WithError(err).Error("Error creating conversation")
		game.SendMessage() <- message.Reply("Something went wrong starting the conversation.")
		return true
	}

	// Set context for template rendering
	conv.SetContext("PLAYER", message.Character.Name)
	conv.SetContext("NPC", npc.Name)
	game.GetFacade().ConversationsService().Update(conv.ID, conv)

	// Send dialog message
	sendDialogMessage(game, message, npc.Name, dialog, conv)

	return true
}

// sendDialogMessage sends the current dialog state to the player
func sendDialogMessage(game def.GameCtrl, message *messages.Message, npcName string, dialog *dialogs.Dialog, conv *conversations.Conversation) {
	// Get current node
	currentNode := game.GetFacade().ConversationsService().GetCurrentNode(conv, dialog)
	if currentNode == nil {
		game.SendMessage() <- message.Reply(npcName + " has nothing more to say.")
		return
	}

	// Build dialog state for rendering
	dialogState := &dialogs.DialogState{
		CurrentDialogID: conv.CurrentNodeID,
		DialogVisited:   conv.VisitedNodes,
		Context:         conv.Context,
	}

	// Render the NPC text with context
	npcText := currentNode.Render(dialogState)

	// Get filtered options
	filteredOptions := game.GetFacade().ConversationsService().GetFilteredOptions(conv, currentNode)

	// Convert to DialogOption format
	options := make([]messages.DialogOption, 0)
	for i, opt := range filteredOptions {
		optText := opt.Text
		if optText == "" {
			optText = opt.RenderPlain(dialogState)
		}
		options = append(options, messages.DialogOption{
			Index: i + 1, // 1-based index
			Text:  optText,
		})
	}

	// Send dialog message
	dialogMsg := messages.NewDialogMessage(
		message.FromUser.ID,
		npcName,
		npcText,
		options,
		conv.ID,
	)

	game.SendMessage() <- dialogMsg
}
