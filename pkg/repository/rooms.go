package repository

import (
	"errors"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	r "github.com/talesmud/talesmud/pkg/entities/rooms"
)

//--- Interface Definitions

//RoomsRepository repository interface
type RoomsRepository interface {
	FindAll() ([]*r.Room, error)
	FindAllWithQuery(query RoomsQuery) ([]*r.Room, error)
	FindByID(id string) (*r.Room, error)
	FindByName(name string) ([]*r.Room, error)

	Store(room *r.Room) (*r.Room, error)
	Import(room *r.Room) (*r.Room, error)
	Update(id string, room *r.Room) error
	Delete(id string) error
	Drop() error
}

//RoomsQuery ...
type RoomsQuery struct {
	Name        string `form:"name"`
	Description string `form:"description"`
	Detail      string `form:"detail"`
	RoomType    string `form:"roomType"`
	Area        string `form:"area"`
	AreaType    string `form:"areaType"`
}

/*
func (query RoomsQuery) matches(room *r.Room) bool {

	match := true

	if query.Name != nil && !strings.Contains(strings.ToLower(room.Name), strings.ToLower(*query.Name)) {
		match = false
	}
	if match && query.Description != nil && !strings.Contains(strings.ToLower(room.Description), strings.ToLower(*query.Description)) {
		match = false
	}
	if match && query.Detail != nil && !strings.Contains(strings.ToLower(room.Detail), strings.ToLower(*query.Detail)) {
		match = false
	}
	if match && query.RoomType != nil && !strings.Contains(strings.ToLower(string(room.RoomType)), strings.ToLower(string(*query.RoomType))) {
		match = false
	}
	if match && query.Area != nil && !strings.Contains(strings.ToLower(string(room.Area)), strings.ToLower(string(*query.Area))) {
		match = false
	}
	if match && query.AreaType != nil && !strings.Contains(strings.ToLower(string(room.AreaType)), strings.ToLower(string(*query.AreaType))) {
		match = false
	}

	return match
}*/

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

func (repo *roomsRepository) FindAllWithQuery(query RoomsQuery) ([]*r.Room, error) {
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

	if e := repo.GenericRepo.FindAllWithParam(db.NewQueryParams(params...), func(elem interface{}) {
		results = append(results, elem.(*r.Room))
	}); e != nil {
		return results, nil
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
