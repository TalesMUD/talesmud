package repository

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/db"
	"github.com/talesmud/talesmud/pkg/entities"
	s "github.com/talesmud/talesmud/pkg/scripts"
)

//--- Interface Definitions

//ScriptsRepository repository interface
type ScriptsRepository interface {
	FindAll() ([]*s.Script, error)
	FindByID(id string) (*s.Script, error)
	FindByName(name string) ([]*s.Script, error)
	Store(script *s.Script) (*s.Script, error)
	Import(script *s.Script) (*s.Script, error)
	Update(id string, script *s.Script) error
	Delete(id string) error
	Drop() error
}

//--- Implementations
type scriptsRepository struct {
	*GenericRepo
}

//NewMongoScriptRepository creates a new mongodb charactersRepository
func NewMongoDBScriptRepository(db *db.Client) ScriptsRepository {

	// create index on id
	//db.members.createIndex( { "user_id": 1 }, { unique: true } )

	sr := &scriptsRepository{
		GenericRepo: &GenericRepo{
			db:         db,
			collection: "scripts",
			generator: func() interface{} {
				return &s.Script{}
			},
		},
	}

	sr.CreateIndex()

	return sr
}

// Drop ...
func (repo *scriptsRepository) Drop() error {
	return repo.GenericRepo.DropCollection()
}

func (repo *scriptsRepository) FindByID(id string) (*s.Script, error) {

	if id == "" {
		log.Error("Scripts::FindByID - id is empty")
		return nil, errors.New("Empty id")
	}

	result, err := repo.GenericRepo.FindByID(id)
	if err == nil {
		return result.(*s.Script), nil
	}
	return nil, err
}

func (repo *scriptsRepository) FindByName(name string) ([]*s.Script, error) {
	results := make([]*s.Script, 0)

	repo.GenericRepo.FindAllWithParam(
		db.NewQueryParams(db.QueryParam{Key: "name", Value: name}),
		func(elem interface{}) {
			results = append(results, elem.(*s.Script))
		})

	return results, nil
}

func (repo *scriptsRepository) FindAll() ([]*s.Script, error) {
	results := make([]*s.Script, 0)
	if err := repo.GenericRepo.FindAll(func(elem interface{}) {
		results = append(results, elem.(*s.Script))
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func (repo *scriptsRepository) Update(id string, script *s.Script) error {
	return repo.GenericRepo.Update(script, id)
}

func (repo *scriptsRepository) Delete(id string) error {
	return repo.GenericRepo.Delete(id)
}

func (repo *scriptsRepository) Store(script *s.Script) (*s.Script, error) {
	script.Entity = entities.NewEntity()
	return repo.Import(script)
}
func (repo *scriptsRepository) Import(script *s.Script) (*s.Script, error) {
	result, err := repo.GenericRepo.Store(script)

	if result == nil {
		return nil, err
	}
	return result.(*s.Script), nil
}
