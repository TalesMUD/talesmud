package service

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// DialogsService provides business logic for dialogs
type DialogsService interface {
	r.DialogsRepository
}

type dialogsService struct {
	r.DialogsRepository
}

// NewDialogsService creates a new dialogs service
func NewDialogsService(dialogsRepo r.DialogsRepository) DialogsService {
	return &dialogsService{
		dialogsRepo,
	}
}

// Store overrides the repository store to add metadata
func (srv *dialogsService) Store(dialog *dialogs.Dialog) (*dialogs.Dialog, error) {
	dialog.Created = time.Now()
	return srv.DialogsRepository.Store(dialog)
}
