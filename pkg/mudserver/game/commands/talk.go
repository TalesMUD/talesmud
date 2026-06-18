package commands

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/conversations"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

const questOnlyDialogID = "__quest_options__"

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

	// Track quest progress for talking to NPC (deliver objectives)
	NotifyQuestTalkToNPC(game, message.Character.ID, message.FromUser.ID, npc)

	// Check for quest offers from this NPC
	npcTemplateID := npc.TemplateID
	if npcTemplateID == "" {
		npcTemplateID = npc.ID
	}
	questOptions := getQuestDialogOptions(game, message.Character.ID, npcTemplateID, npc.ID)

	// Check if NPC has a dialog
	if !npc.HasDialog() {
		// If NPC has quest options but no dialog, show quest dialog
		if len(questOptions) > 0 {
			conv, err := game.GetFacade().ConversationsService().GetOrCreateConversation(
				message.Character.ID,
				npc.ID,
				conversations.TargetTypeNPC,
				questOnlyDialogID,
			)
			if err != nil {
				log.WithError(err).Error("Error creating quest-only conversation")
				game.SendMessage() <- message.Reply("Something went wrong starting the conversation.")
				return true
			}

			conv.DialogID = questOnlyDialogID
			conv.CurrentNodeID = "main"
			conv.SetContext("PLAYER", message.Character.Name)
			conv.SetContext("NPC", npc.Name)
			game.GetFacade().ConversationsService().Update(conv.ID, conv)

			sendQuestOnlyDialog(game, message, npc.Name, conv, questOptions)
			return true
		}
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
	conv.DialogID = npc.DialogID
	conv.SetContext("PLAYER", message.Character.Name)
	conv.SetContext("NPC", npc.Name)
	game.GetFacade().ConversationsService().Update(conv.ID, conv)

	// Track dialog node visit for quest progress
	NotifyQuestDialogNode(game, message.Character.ID, message.FromUser.ID, npc.ID, npc.DialogID, conv.CurrentNodeID)

	// Send dialog message (with quest options injected)
	sendDialogMessage(game, message, npc.Name, dialog, conv, questOptions)

	return true
}

// questDialogOption represents a quest-related dialog option
type questDialogOption struct {
	text      string // Player-facing label shown in the option list
	npcText   string // NPC response shown after the player selects this option
	questID   string
	questName string
	action    string // "accept", "complete", "progress"
}

// getQuestDialogOptions checks for quest-related dialog options for an NPC.
// Only shows quests whose prerequisites are met.
func getQuestDialogOptions(game def.GameCtrl, characterID, npcTemplateID, npcInstanceID string) []questDialogOption {
	var options []questDialogOption

	// Find quests offered by this NPC (check both template and instance ID)
	npcQuests, err := game.GetFacade().QuestsService().FindBySourceNPC(npcTemplateID)
	if err != nil {
		return options
	}

	// Also check instance ID if different
	if npcInstanceID != npcTemplateID {
		instanceQuests, err := game.GetFacade().QuestsService().FindBySourceNPC(npcInstanceID)
		if err == nil {
			npcQuests = append(npcQuests, instanceQuests...)
		}
	}

	// Load character for level check
	char, _ := game.GetFacade().CharactersService().FindByID(characterID)

	for _, quest := range npcQuests {
		progress, _ := game.GetFacade().QuestsService().GetProgress(characterID, quest.ID)

		if progress == nil || progress.Status == quests.QuestStatusAbandoned {
			// Check prerequisites before offering
			if !questPrereqsMet(game, characterID, quest, char) {
				continue
			}

			options = append(options, questDialogOption{
				text:      fmt.Sprintf("[Quest] %s", quest.Name),
				npcText:   quest.AcceptDialogText,
				questID:   quest.ID,
				questName: quest.Name,
				action:    "accept",
			})
		} else if progress.Status == quests.QuestStatusActive {
			// Check if all objectives are complete
			allComplete := true
			for _, obj := range progress.Objectives {
				if !obj.Completed {
					allComplete = false
					break
				}
			}
			if allComplete {
				options = append(options, questDialogOption{
					text:      fmt.Sprintf("[Turn In] %s", quest.Name),
					npcText:   quest.CompleteDialogText,
					questID:   quest.ID,
					questName: quest.Name,
					action:    "complete",
				})
			} else {
				options = append(options, questDialogOption{
					text:      fmt.Sprintf("[In Progress] %s", quest.Name),
					npcText:   quest.ProgressDialogText,
					questID:   quest.ID,
					questName: quest.Name,
					action:    "progress",
				})
			}
		}
	}

	return options
}

// questPrereqsMet checks if a character meets the prerequisites for a quest
func questPrereqsMet(game def.GameCtrl, characterID string, quest *quests.Quest, char *characters.Character) bool {
	// Check level requirement
	if quest.RequiredLevel > 0 && char != nil && char.Level < quest.RequiredLevel {
		return false
	}

	// Check required quests are completed
	for _, reqID := range quest.RequiredQuestIDs {
		reqProgress, _ := game.GetFacade().QuestsService().GetProgress(characterID, reqID)
		if reqProgress == nil || reqProgress.Status != quests.QuestStatusCompleted {
			return false
		}
	}

	return true
}

// sendQuestOnlyDialog sends a dialog with only quest options (for NPCs without dialogs)
func sendQuestOnlyDialog(game def.GameCtrl, message *messages.Message, npcName string, conv *conversations.Conversation, questOptions []questDialogOption) {
	options := make([]messages.DialogOption, len(questOptions))
	for i, qo := range questOptions {
		options[i] = messages.DialogOption{
			Index: i + 1,
			Text:  qo.text,
		}
	}

	dialogMsg := messages.NewDialogMessage(
		message.FromUser.ID,
		npcName,
		npcName+" looks at you expectantly.",
		options,
		conv.ID,
	)

	game.SendMessage() <- dialogMsg
}

// sendDialogMessage sends the current dialog state to the player
func sendDialogMessage(game def.GameCtrl, message *messages.Message, npcName string, dialog *dialogs.Dialog, conv *conversations.Conversation, questOptions []questDialogOption) {
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

	// Inject quest options at the end
	for _, qo := range questOptions {
		options = append(options, messages.DialogOption{
			Index: len(options) + 1,
			Text:  qo.text,
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
