package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	r "github.com/talesmud/talesmud/pkg/entities/rooms"
)

//--- Interface Definitions

//RoomsRepository repository interface
type RoomsRepository interface {
	FindAll() ([]*r.Room, error)
	FindByID(id string) (*r.Room, error)
	FindByName(name string) ([]*r.Room, error)

	Store(room *r.Room) (*r.Room, error)
	Import(room *r.Room) (*r.Room, error)
	Update(id string, room *r.Room) error
	Delete(id string) error
	Drop() error
}

//--- Implementations

type roomsRepository struct {
	*GenericRepo
}

// Drop ...
func (repo *roomsRepository) Drop() error {
	return repo.GenericRepo.DropCollection()
}

//NewMongoDBRoomsRepository creates a new mongodb repoRepository
func NewMongoDBRoomsRepository(db *db.Client) RoomsRepository {

	r := &roomsRepository{
		GenericRepo: &GenericRepo{
			db:         db,
			collection: "rooms",
			generator: func() interface{} {
				return &r.Room{}
			},
		},
	}

	r.CreateIndex()
	return r
}

func (repo *roomsRepository) FindByID(id string) (*r.Room, error) {

	if id == "" {
		log.Error("Rooms::FindByID - id is empty")
		return nil, errors.New("Empty id")
	}

	result, err := repo.GenericRepo.FindByID(id)
	if err == nil {
		return result.(*r.Room), nil
	}
	return nil, err
}

func (repo *roomsRepository) FindByName(name string) ([]*r.Room, error) {
	results := make([]*r.Room, 0)

	if err := repo.GenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*r.Room))
		}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *roomsRepository) FindAll() ([]*r.Room, error) {
	results := make([]*r.Room, 0)
	if err := repo.GenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*r.Room))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *roomsRepository) Update(id string, charachterSheet *r.Room) error {
	return repo.GenericRepo.Update(charachterSheet, id)
}

func (repo *roomsRepository) Delete(id string) error {
	return repo.GenericRepo.Delete(id)
}

func (repo *roomsRepository) Store(rep *r.Room) (*r.Room, error) {
	//rep.Entity = e.NewEntity()
	return repo.Import(rep)
}

func (repo *roomsRepository) Import(rep *r.Room) (*r.Room, error) {
	if result, err := repo.GenericRepo.Store(rep); err == nil {
		return result.(*r.Room), nil
	} else {
		return result.(*r.Room), nil
	}
}
