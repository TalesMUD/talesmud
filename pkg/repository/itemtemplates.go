package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/db"
	"github.com/talesmud/talesmud/pkg/entities"
	i "github.com/talesmud/talesmud/pkg/entities/items"
)

//--- Interface Definitions

//ItemTemplatesRepository repository interface
type ItemTemplatesRepository interface {
	FindAll() ([]*i.ItemTemplate, error)
	FindByID(id string) (*i.ItemTemplate, error)
	FindByName(name string) ([]*i.ItemTemplate, error)
	Store(item *i.ItemTemplate) (*i.ItemTemplate, error)
	Import(item *i.ItemTemplate) (*i.ItemTemplate, error)
	Update(id string, item *i.ItemTemplate) error
	Delete(id string) error
	Drop() error
}

//--- Implementations
type itemTemplatesRepository struct {
	*GenericRepo
}

//NewMongoDBItemTemplatesRepository creates a new mongodb charactersRepository
func NewMongoDBItemTemplatesRepository(db *db.Client) ItemTemplatesRepository {

	// create index on id
	//db.members.createIndex( { "user_id": 1 }, { unique: true } )

	ir := &itemTemplatesRepository{
		GenericRepo: &GenericRepo{
			db:         db,
			collection: "itemtemplates",
			generator: func() interface{} {
				return &i.ItemTemplate{}
			},
		},
	}

	ir.CreateIndex()

	return ir
}

// Drop ...
func (repo *itemTemplatesRepository) Drop() error {
	return repo.GenericRepo.DropCollection()
}

func (repo *itemTemplatesRepository) FindByID(id string) (*i.ItemTemplate, error) {

	if id == "" {
		log.Error("ItemTemplates::FindByID - id is empty")
		return nil, errors.New("Empty id")
	}

	result, err := repo.GenericRepo.FindByID(id)
	if err == nil {
		return result.(*i.ItemTemplate), nil
	}
	return nil, err
}

func (repo *itemTemplatesRepository) FindByName(name string) ([]*i.ItemTemplate, error) {
	results := make([]*i.ItemTemplate, 0)

	repo.GenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*i.ItemTemplate))
		})

	return results, nil
}

func (repo *itemTemplatesRepository) FindAll() ([]*i.ItemTemplate, error) {
	results := make([]*i.ItemTemplate, 0)
	if err := repo.GenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*i.ItemTemplate))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *itemTemplatesRepository) Update(id string, item *i.ItemTemplate) error {
	return repo.GenericRepo.Update(item, id)
}

func (repo *itemTemplatesRepository) Delete(id string) error {
	return repo.GenericRepo.Delete(id)
}

func (repo *itemTemplatesRepository) Store(item *i.ItemTemplate) (*i.ItemTemplate, error) {
	item.Entity = entities.NewEntity()
	return repo.Import(item)
}
func (repo *itemTemplatesRepository) Import(item *i.ItemTemplate) (*i.ItemTemplate, error) {
	result, err := repo.GenericRepo.Store(item)

	if result == nil {
		return nil, err
	}
	return result.(*i.ItemTemplate), nil
}
