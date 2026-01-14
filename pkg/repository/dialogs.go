package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
)

// DialogsRepository provides access to dialog data
type DialogsRepository interface {
	FindAll() ([]*dialogs.Dialog, error)
	FindByID(id string) (*dialogs.Dialog, error)
	FindByName(name string) (*dialogs.Dialog, error)
	Store(dialog *dialogs.Dialog) (*dialogs.Dialog, error)
	Import(dialog *dialogs.Dialog) (*dialogs.Dialog, error)
	Update(id string, dialog *dialogs.Dialog) error
	Delete(id string) error
	Drop() error
}

type dialogsRepository struct {
	*GenericRepo
}

// NewMongoDBDialogsRepository creates a new dialogs repository
func NewMongoDBDialogsRepository(db *db.Client) DialogsRepository {
	repo := &dialogsRepository{
		GenericRepo: &GenericRepo{
			db:         db,
			collection: "dialogs",
			generator: func() interface{} {
				return &dialogs.Dialog{}
			},
		},
	}
	repo.CreateIndex()
	return repo
}

// FindAll returns all dialogs
func (repo *dialogsRepository) FindAll() ([]*dialogs.Dialog, error) {
	results := make([]*dialogs.Dialog, 0)
	if err := repo.GenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*dialogs.Dialog))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

// FindByID returns a dialog by its ID
func (repo *dialogsRepository) FindByID(id string) (*dialogs.Dialog, error) {
	if id == "" {
		log.Error("Dialogs::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.GenericRepo.FindByID(id)
	if err == nil {
		return result.(*dialogs.Dialog), nil
	}
	return nil, err
}

// FindByName returns a dialog by its name
func (repo *dialogsRepository) FindByName(name string) (*dialogs.Dialog, error) {
	if name == "" {
		log.Error("Dialogs::FindByName - name is empty")
		return nil, errors.New("empty name")
	}
	result, err := repo.GenericRepo.FindByField("name", name)
	if err == nil {
		return result.(*dialogs.Dialog), nil
	}
	return nil, err
}

// Store creates a new dialog with a new entity ID
func (repo *dialogsRepository) Store(dialog *dialogs.Dialog) (*dialogs.Dialog, error) {
	dialog.Entity = entities.NewEntity()
	return repo.Import(dialog)
}

// Import stores a dialog (used when dialog already has an ID)
func (repo *dialogsRepository) Import(dialog *dialogs.Dialog) (*dialogs.Dialog, error) {
	if _, err := repo.GenericRepo.Store(dialog); err != nil {
		return nil, err
	}
	return dialog, nil
}

// Update updates an existing dialog
func (repo *dialogsRepository) Update(id string, dialog *dialogs.Dialog) error {
	return repo.GenericRepo.Update(dialog, id)
}

// Delete removes a dialog
func (repo *dialogsRepository) Delete(id string) error {
	return repo.GenericRepo.Delete(id)
}

// Drop removes all dialogs
func (repo *dialogsRepository) Drop() error {
	return repo.GenericRepo.DropCollection()
}
