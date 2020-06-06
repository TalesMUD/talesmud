package repository

import (
	"context"
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/db"
	e "github.com/talesmud/talesmud/pkg/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenericRepo ...
type GenericRepo struct {
	db         *db.Client
	collection string
	generator  func() interface{}
}

type entityGenerator func() interface{}
type elementCollector func(element interface{})

// FindByID ...
func (repo *GenericRepo) FindByID(id string) (interface{}, error) {

	log.WithField("id", id).Info("FindByID called")

	oid, _ := primitive.ObjectIDFromHex(id)
	result := repo.db.FindByID(repo.collection, oid)

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
			return errors.New("Could not decode element")
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
			return errors.New("Could not decode element")
		}
		collector(elem)
	}
	return nil
}

// Store stores a new entity
func (repo *GenericRepo) Store(entity interface{}) (interface{}, error) {

	if result, error := repo.db.InsertOne(repo.collection, entity); error != nil {
		log.WithField("Error", error).Error("error during insertion")
		return nil, error
	} else {
		if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
			if n, ok := entity.(e.Entity); ok {
				n.ID = oid
			}
		}
	}

	return entity, nil
}

// Delete an existing entity
func (repo *GenericRepo) Delete(id string) error {

	oid, _ := primitive.ObjectIDFromHex(id)

	if result, err := repo.db.DeleteByID(repo.collection, oid); err != nil {
		log.WithError(err).Error("Error during update")
		return err
	} else {
		log.WithField("Generic Update", result).Info("updated entity")
	}

	return nil
}

// Update an existing entity
func (repo *GenericRepo) Update(item interface{}, id string) error {

	oid, _ := primitive.ObjectIDFromHex(id)

	if result, err := repo.db.UpdateOneByID(repo.collection, oid, item); err != nil {
		log.WithError(err).Error("Error during update")
		return err
	} else {
		log.WithField("Generic Update", result).Info("updated entity")
	}

	return nil
}
