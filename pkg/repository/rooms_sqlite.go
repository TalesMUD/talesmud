package repository

import (
	"errors"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	r "github.com/talesmud/talesmud/pkg/entities/rooms"
)

type sqliteRoomsRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteRoomsRepository creates a new SQLite rooms repository.
func NewSQLiteRoomsRepository(client *dbsqlite.Client) RoomsRepository {
	return &sqliteRoomsRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "rooms", func() interface{} {
			return &r.Room{}
		}),
	}
}

func (repo *sqliteRoomsRepository) Drop() error {
	return repo.sqliteGenericRepo.DropCollection()
}

func (repo *sqliteRoomsRepository) FindByID(id string) (*r.Room, error) {
	if id == "" {
		log.Error("Rooms::FindByID - id is empty")
		return nil, errors.New("empty id")
	}
	result, err := repo.sqliteGenericRepo.FindByID(id)
	if err == nil {
		return result.(*r.Room), nil
	}
	return nil, err
}

func (repo *sqliteRoomsRepository) FindByName(name string) ([]*r.Room, error) {
	results := make([]*r.Room, 0)
	if err := repo.sqliteGenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*r.Room))
		}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteRoomsRepository) FindAll() ([]*r.Room, error) {
	results := make([]*r.Room, 0)
	if err := repo.sqliteGenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*r.Room))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *sqliteRoomsRepository) FindAllWithQuery(query RoomsQuery) ([]*r.Room, error) {
	results := make([]*r.Room, 0)
	params := []db.QueryParam{}

	v := reflect.ValueOf(query)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Interface() != nil {
			p := db.QueryParam{Key: strings.ToLower(v.Type().Field(i).Name), Value: v.Field(i).Interface()}
			if p.Value != "" {
				params = append(params, p)
			}
		}
	}

	if err := repo.sqliteGenericRepo.FindAllWithParam(db.NewQueryParams(params...), func(elem interface{}) {
		results = append(results, elem.(*r.Room))
	}); err != nil {
		return results, nil
	}
	return results, nil
}

func (repo *sqliteRoomsRepository) Update(id string, room *r.Room) error {
	return repo.sqliteGenericRepo.Update(room, id)
}

func (repo *sqliteRoomsRepository) Delete(id string) error {
	return repo.sqliteGenericRepo.Delete(id)
}

func (repo *sqliteRoomsRepository) Store(rep *r.Room) (*r.Room, error) {
	return repo.Import(rep)
}

func (repo *sqliteRoomsRepository) Import(rep *r.Room) (*r.Room, error) {
	if result, err := repo.sqliteGenericRepo.Store(rep); err == nil {
		return result.(*r.Room), nil
	} else {
		return result.(*r.Room), nil
	}
}
