package service

import (
	"github.com/talesmud/talesmud/pkg/entities/conversations"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// ConversationsService provides business logic for managing conversation state
type ConversationsService interface {
	r.ConversationsRepository

	// GetOrCreateConversation retrieves an existing conversation or creates a new one
	GetOrCreateConversation(characterID, targetID string, targetType conversations.TargetType, dialogID string) (*conversations.Conversation, error)

	// GetCurrentNode returns the current dialog node for a conversation
	GetCurrentNode(conv *conversations.Conversation, dialog *dialogs.Dialog) *dialogs.Dialog

	// GetFilteredOptions returns dialog options filtered by visit requirements
	GetFilteredOptions(conv *conversations.Conversation, node *dialogs.Dialog) []*dialogs.Dialog

	// AdvanceConversation moves the conversation to a new node
	AdvanceConversation(conv *conversations.Conversation, nodeID string) error

	// ResetConversation resets the conversation to the main node
	ResetConversation(conv *conversations.Conversation) error
}

type conversationsService struct {
	r.ConversationsRepository
}

// NewConversationsService creates a new conversations service
func NewConversationsService(convRepo r.ConversationsRepository) ConversationsService {
	return &conversationsService{
		convRepo,
	}
}

// GetOrCreateConversation retrieves an existing conversation or creates a new one
func (srv *conversationsService) GetOrCreateConversation(characterID, targetID string, targetType conversations.TargetType, dialogID string) (*conversations.Conversation, error) {
	// Try to find existing conversation
	conv, err := srv.FindByCharacterAndTarget(characterID, targetID)
	if err == nil && conv != nil {
		// Update interaction time
		conv.UpdateInteraction()
		srv.Update(conv.ID, conv)
		return conv, nil
	}

	// Create new conversation
	newConv := conversations.NewConversation(characterID, targetID, targetType, dialogID)
	return srv.Store(newConv)
}

// GetCurrentNode returns the current dialog node for a conversation
func (srv *conversationsService) GetCurrentNode(conv *conversations.Conversation, dialog *dialogs.Dialog) *dialogs.Dialog {
	if conv.CurrentNodeID == "" || conv.CurrentNodeID == "main" {
		return dialog
	}
	return dialog.FindDialog(conv.CurrentNodeID)
}

// GetFilteredOptions returns dialog options that the player is allowed to see
func (srv *conversationsService) GetFilteredOptions(conv *conversations.Conversation, node *dialogs.Dialog) []*dialogs.Dialog {
	if node.Options == nil {
		return nil
	}

	filtered := make([]*dialogs.Dialog, 0)
	for _, option := range node.Options {
		// Check ShowOnlyOnce - if set and already visited, skip
		if option.ShowOnlyOnce != nil && *option.ShowOnlyOnce {
			if conv.HasVisited(option.NodeID) {
				continue
			}
		}

		// Check RequiresVisitedDialogs - all required nodes must be visited
		if len(option.RequiresVisitedDialogs) > 0 {
			if !conv.HasVisitedAll(option.RequiresVisitedDialogs) {
				continue
			}
		}

		filtered = append(filtered, option)
	}

	return filtered
}

// AdvanceConversation moves the conversation to a new node and marks it as visited
func (srv *conversationsService) AdvanceConversation(conv *conversations.Conversation, nodeID string) error {
	conv.CurrentNodeID = nodeID
	conv.MarkVisited(nodeID)
	conv.UpdateInteraction()
	return srv.Update(conv.ID, conv)
}

// ResetConversation resets the conversation to the main node
func (srv *conversationsService) ResetConversation(conv *conversations.Conversation) error {
	conv.CurrentNodeID = "main"
	conv.UpdateInteraction()
	return srv.Update(conv.ID, conv)
}
