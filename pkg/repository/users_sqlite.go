package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	e "github.com/talesmud/talesmud/pkg/entities"
)

type sqliteUsersRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteUsersRepository creates a new SQLite users repository.
func NewSQLiteUsersRepository(client *dbsqlite.Client) UsersRepository {
	return &sqliteUsersRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "users", func() interface{} {
			return &e.User{}
		}),
	}
}

func (repo *sqliteUsersRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}

func (repo *sqliteUsersRepository) Import(user *e.User) (*e.User, error) {
	result, err := repo.sqliteGenericRepo.Store(user)
	return result.(*e.User), err
}

func (repo *sqliteUsersRepository) Create(user *e.User) (*e.User, error) {
	user.Entity = e.NewEntity()
	return repo.Import(user)
}

func (repo *sqliteUsersRepository) FindAll() ([]*e.User, error) {
	results := make([]*e.User, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*e.User))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteUsersRepository) FindAllOnline() ([]*e.User, error) {
	results := make([]*e.User, 0)
	if err := repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "isOnline", Value: true}),
		func(elem interface{}) {
			results = append(results, elem.(*e.User))
		}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteUsersRepository) Update(refID string, user *e.User) error {
	return repo.sqliteGenericRepo.UpdateByField(user, "refid", refID)
}

func (repo *sqliteUsersRepository) FindByID(id string) (*e.User, error) {
	if id == "" {
		log.Error("Users::FindByID - ID is empty")
		return nil, errors.New("empty ID")
	}

	result, err := repo.sqliteGenericRepo.FindByID(id)
	if user, ok := result.(*e.User); ok {
		return user, nil
	}
	return nil, err
}

func (repo *sqliteUsersRepository) FindByRefID(refID string) (*e.User, error) {
	if refID == "" {
		log.Error("Users::FindByRefID - refID is empty")
		return nil, errors.New("empty refID")
	}

	result, err := repo.sqliteGenericRepo.FindByField("refid", refID)
	if user, ok := result.(*e.User); ok {
		return user, nil
	}
	return nil, err
}

func (repo *sqliteUsersRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}
