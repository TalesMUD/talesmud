package npc

import (
	"time"

	d "github.com/talesmud/talesmud/pkg/entities/dialogs"
)

//DialogTrait data
type DialogTrait struct {

	// InteractiveDialog is used for conversation between player characters and non player characters
	// this dialog is invoked by a player using the "talk to" command with the NPC name
	InteractiveDialog *d.Dialog `json:"interactiveDialog,omitempty"`

	// IdleDioalog is automatically started after a given IdleDialogTimeout is reached and will continue the last invoked idle dialog
	IdleDialog        *d.Dialog     `json:"idleDialog,omitempty"`
	IdleDialogTimeout time.Duration `json:"idleDialogTimeout,omitempty"`
}

//IsInteractiveDialog ...
func (dialogTrait *DialogTrait) HasInteractiveDialog() bool {
	return dialogTrait.InteractiveDialog != nil
}

func (dialogTrait *DialogTrait) HasIdleDialog() bool {
	return dialogTrait.IdleDialog != nil
}

// Companion??
