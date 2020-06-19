package repository

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/db"
	"github.com/talesmud/talesmud/pkg/entities"
	i "github.com/talesmud/talesmud/pkg/entities/items"
)

//--- Interface Definitions

//ItemsRepository repository interface
type ItemsRepository interface {
	FindAll(query ItemsQuery) ([]*i.Item, error)
	FindByID(id string) (*i.Item, error)
	FindByName(name string) ([]*i.Item, error)
	Store(item *i.Item) (*i.Item, error)
	Import(item *i.Item) (*i.Item, error)
	Update(id string, item *i.Item) error
	Delete(id string) error
	Drop() error
}

//ItemsQuery ...
type ItemsQuery struct {
	Name        *string        `form:"name"`
	Description *string        `form:"description"`
	Detail      *string        `form:"detail"`
	Type        *i.ItemType    `form:"type"`
	SubType     *i.ItemSubType `form:"subType"`
	Quality     *i.ItemQuality `form:"quality"`
	Slot        *i.ItemSlot    `form:"slot"`
	Level       *int32         `form:"level"`
}

func (query ItemsQuery) matches(item *i.Item) bool {

	match := true

	if query.Name != nil && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(*query.Name)) {
		match = false
	}
	if match && query.Description != nil && !strings.Contains(strings.ToLower(item.Description), strings.ToLower(*query.Description)) {
		match = false
	}
	if match && query.Detail != nil && !strings.Contains(strings.ToLower(item.Detail), strings.ToLower(*query.Detail)) {
		match = false
	}
	if match && query.Type != nil && !strings.Contains(strings.ToLower(string(item.Type)), strings.ToLower(string(*query.Type))) {
		match = false
	}
	if match && query.Slot != nil && !strings.Contains(strings.ToLower(string(item.Slot)), strings.ToLower(string(*query.Slot))) {
		match = false
	}
	if match && query.SubType != nil && !strings.Contains(strings.ToLower(string(item.SubType)), strings.ToLower(string(*query.SubType))) {
		match = false
	}
	if match && query.Quality != nil && !strings.Contains(strings.ToLower(string(item.Quality)), strings.ToLower(string(*query.Quality))) {
		match = false
	}
	if match && query.Level != nil && item.Level != *query.Level {
		match = false
	}
	return match
}

//--- Implementations
type itemsRepository struct {
	*GenericRepo
}

//NewMongoDBItemsRepository creates a new mongodb charactersRepository
func NewMongoDBItemsRepository(db *db.Client) ItemsRepository {

	// create index on id
	//db.members.createIndex( { "user_id": 1 }, { unique: true } )

	ir := &itemsRepository{
		GenericRepo: &GenericRepo{
			db:         db,
			collection: "items",
			generator: func() interface{} {
				return &i.Item{}
			},
		},
	}

	ir.CreateIndex()

	return ir
}

// Drop ...
func (repo *itemsRepository) Drop() error {
	return repo.GenericRepo.DropCollection()
}

func (repo *itemsRepository) FindByID(id string) (*i.Item, error) {

	if id == "" {
		log.Error("Items::FindByID - id is empty")
		return nil, errors.New("Empty id")
	}

	result, err := repo.GenericRepo.FindByID(id)
	if err == nil {
		return result.(*i.Item), nil
	}
	return nil, err
}

func (repo *itemsRepository) FindByName(name string) ([]*i.Item, error) {
	results := make([]*i.Item, 0)

	repo.GenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*i.Item))
		})

	return results, nil
}

func (repo *itemsRepository) FindAll(query ItemsQuery) ([]*i.Item, error) {
	results := make([]*i.Item, 0)
	if err := repo.GenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*i.Item))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *itemsRepository) Update(id string, item *i.Item) error {
	return repo.GenericRepo.Update(item, id)
}

func (repo *itemsRepository) Delete(id string) error {
	return repo.GenericRepo.Delete(id)
}

func (repo *itemsRepository) Store(item *i.Item) (*i.Item, error) {
	item.Entity = entities.NewEntity()
	return repo.Import(item)
}
func (repo *itemsRepository) Import(item *i.Item) (*i.Item, error) {
	result, err := repo.GenericRepo.Store(item)

	if result == nil {
		return nil, err
	}
	return result.(*i.Item), nil
}
