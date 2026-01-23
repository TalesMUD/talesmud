package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
)

type sqliteDialogsRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteDialogsRepository creates a new SQLite dialogs repository.
func NewSQLiteDialogsRepository(client *dbsqlite.Client) DialogsRepository {
	return &sqliteDialogsRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "dialogs", func() interface{} {
			return &dialogs.Dialog{}
		}),
	}
}

func (repo *sqliteDialogsRepository) FindAll() ([]*dialogs.Dialog, error) {
	results := make([]*dialogs.Dialog, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*dialogs.Dialog))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteDialogsRepository) FindByID(id string) (*dialogs.Dialog, error) {
	if id == "" {
		log.Error("Dialogs::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*dialogs.Dialog), nil
	}
	return nil, err
}

func (repo *sqliteDialogsRepository) FindByName(name string) (*dialogs.Dialog, error) {
	if name == "" {
		log.Error("Dialogs::FindByName - name is empty")
		return nil, errors.New("empty name")
	}
	result, err := repo.sqliteGenericRepo.FindByField("name", name)
	if err == nil {
		return result.(*dialogs.Dialog), nil
	}
	return nil, err
}

func (repo *sqliteDialogsRepository) Store(dialog *dialogs.Dialog) (*dialogs.Dialog, error) {
	dialog.Entity = entities.NewEntity()
	return repo.Import(dialog)
}

func (repo *sqliteDialogsRepository) Import(dialog *dialogs.Dialog) (*dialogs.Dialog, error) {
	if _, err := repo.sqliteGenericRepo.Store(dialog); err != nil {
		return nil, err
	}
	return dialog, nil
}

func (repo *sqliteDialogsRepository) Update(id string, dialog *dialogs.Dialog) error {
	return repo.sqliteGenericRepo.Update(dialog, id)
}

func (repo *sqliteDialogsRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteDialogsRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}
