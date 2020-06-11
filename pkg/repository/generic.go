package repository

import (
	"context"
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

// GenericRepo ...
type GenericRepo struct {
	db         *db.Client
	collection string
	generator  func() interface{}
}

type entityGenerator func() interface{}
type elementCollector func(element interface{})

//CreateIndex ..
func (repo *GenericRepo) CreateIndex() {
	mod := mongo.IndexModel{
		Keys: bson.M{
			"id": 1,
		},
		Options: options.Index().SetUnique(true),
	}

	name, err := repo.db.C(repo.collection).Indexes().CreateOne(context.Background(), mod)

	if err != nil {
		log.WithError(err).Info("Failed creating Index")
	} else {
		log.WithField("Index", name).Info("Created Index")
	}

}

// DropCollection ... drops the whole collection
func (repo *GenericRepo) DropCollection() error {
	return repo.db.Drop(repo.collection)
}

// FindByID ...
func (repo *GenericRepo) FindByID(id string) (interface{}, error) {

	log.WithField("id", id).Info("FindByID called")

	result := repo.db.FindByID(repo.collection, id)

	if result != nil {

		entity := repo.generator()
		if err := result.Decode(entity); err != nil {
			log.WithField("Error", err).Error("Error decoding entity")
			return nil, errors.New("entity not found")
		}
		return entity, nil
	}
	return nil, errors.New("entity not found")
}

// FindByField ...
func (repo *GenericRepo) FindByField(key string, value string) (interface{}, error) {

	log.WithField(key, value).Info("FindByField called")

	result := repo.db.FindOne(repo.collection, key, value)

	if result != nil {
		entity := repo.generator()
		if err := result.Decode(entity); err != nil {
			log.WithField("Error", err).Error("Error decoding entity")
			return nil, errors.New("entity not found")
		}
		return entity, nil
	}
	return nil, errors.New("entity not found")
}

// UpdateByField an existing entity
func (repo *GenericRepo) UpdateByField(item interface{}, key string, value string) error {

	if result, err := repo.db.UpdateOne(repo.collection, key, value, item); err != nil {
		log.WithError(err).Error("Error during update")
		return err
	} else {
		log.WithField("Generic Update", result).Info("updated entity")
	}

	return nil
}

// FindAllWithParam returns all entities
func (repo *GenericRepo) FindAllWithParam(params *db.QueryParams, collector elementCollector) error {

	cursor, err := repo.db.FindAllWithParams(repo.collection, params)

	if err != nil {
		log.WithField("collection", repo.collection).WithField("cursor", cursor).Error(err)
		return err
	}

	for cursor.Next(context.TODO()) {
		elem := repo.generator()
		err := cursor.Decode(elem)
		if err != nil {
			log.WithError(err).Error("Could not decode element")
			continue
		}
		collector(elem)
	}
	return nil
}

// FindAll returns all entities
func (repo *GenericRepo) FindAll(collector elementCollector) error {

	//var results []interface{}
	cursor, err := repo.db.FindAll(repo.collection)

	if err != nil {
		log.WithField("collection", repo.collection).WithField("cursor", cursor).Fatal(err)
	}

	for cursor.Next(context.TODO()) {
		elem := repo.generator()
		err := cursor.Decode(elem)
		if err != nil {
			log.WithError(err).Error("Could not decode element")
			continue
		}
		collector(elem)
	}
	return nil
}

// Store stores a new entity
func (repo *GenericRepo) Store(entity interface{}) (interface{}, error) {

	if _, error := repo.db.InsertOne(repo.collection, entity); error != nil {
		log.WithField("Error", error).Error("error during insertion")
		return nil, error
	}

	return entity, nil
}

// Delete an existing entity
func (repo *GenericRepo) Delete(id string) error {

	if result, err := repo.db.DeleteByID(repo.collection, id); err != nil {
		log.WithError(err).Error("Error during update")
		return err
	} else {
		log.WithField("Generic Update", result).Info("updated entity")
	}

	return nil
}

// Update an existing entity
func (repo *GenericRepo) Update(item interface{}, id string) error {

	if result, err := repo.db.UpdateOneByID(repo.collection, id, item); err != nil {
		log.WithError(err).Error("Error during update")
		return err
	} else {
		log.WithField("Generic Update", result).Info("updated entity")
	}

	return nil
}
