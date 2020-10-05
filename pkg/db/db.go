package db

import (
	"context"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client database
type Client struct {
	client    *mongo.Client
	Connected bool
	Database  string
}

//QueryParam ...
type QueryParam struct {
	Key   string
	Value interface{}
}

//QueryParams ...
type QueryParams struct {
	params []QueryParam
}

//NewQueryParams ...
func NewQueryParams(queryParams ...QueryParam) *QueryParams {

	var p []QueryParam
	for _, qp := range queryParams {
		p = append(p, qp)
	}
	return &QueryParams{
		params: p,
	}
}

// With ...
func (qp *QueryParams) With(param QueryParam) *QueryParams {
	qp.params = append(qp.params, param)
	return qp
}

// AsBSON ...
func (qp QueryParams) AsBSON() bson.M {

	result := bson.M{}
	for _, p := range qp.params {

		result[p.Key] = p.Value

	}
	return result
}

// New creates a new Client object
func New(database string) *Client {

	return &Client{
		Connected: false,
		Database:  database,
	}
}

// Close connection
func (dba *Client) Close() error {
	return dba.client.Disconnect(context.Background())
}

//C opens a given collection in the xdcc database
func (dba *Client) C(coll string) *mongo.Collection {
	return dba.client.Database(dba.Database).Collection(coll)
}

//FindAll returns all entities of a given collection
func (dba *Client) FindAll(coll string) (*mongo.Cursor, error) {
	findFilter := bson.D{}
	return dba.C(coll).Find(context.Background(), findFilter)
}

//Find returns all entities of a given collection
func (dba *Client) Find(coll string, key string, value string) (*mongo.Cursor, error) {
	findFilter := bson.D{{
		Key:   key,
		Value: value,
	}}
	return dba.C(coll).Find(context.Background(), findFilter)
}

//FindAllWithParams returns all entities of a given collection
func (dba *Client) FindAllWithParams(coll string, params *QueryParams) (*mongo.Cursor, error) {
	return dba.C(coll).Find(context.Background(), params.AsBSON())
}

//DeleteAll returns all entities of a given collection
func (dba *Client) DeleteAll(coll string) (*mongo.DeleteResult, error) {
	findFilter := bson.D{}
	return dba.C(coll).DeleteMany(context.Background(), findFilter)
}

//Drop drops the collection
func (dba *Client) Drop(coll string) error {
	return dba.C(coll).Drop(context.Background())
}

//Delete returns all entities of a given collection
func (dba *Client) Delete(coll string, key string, value string) (*mongo.DeleteResult, error) {
	findFilter := bson.D{{key, value}}
	return dba.C(coll).DeleteMany(context.Background(), findFilter)
}

//FindOne returns all entities of a given collection
func (dba *Client) FindOne(coll string, key string, value string) *mongo.SingleResult {

	return dba.C(coll).FindOne(context.Background(), bson.D{{key, value}})
}

//FindByID returns all entities of a given collection
func (dba *Client) FindByID(coll string, id string) *mongo.SingleResult {
	return dba.C(coll).FindOne(context.Background(), bson.M{"id": id})
}

//DeleteByID returns all entities of a given collection
func (dba *Client) DeleteByID(coll string, id string) (*mongo.DeleteResult, error) {

	return dba.C(coll).DeleteOne(context.Background(), bson.M{"id": id})
}

//InsertOne inserts one document
func (dba *Client) InsertOne(coll string, data interface{}) (*mongo.InsertOneResult, error) {
	return dba.client.Database(dba.Database).Collection(coll).InsertOne(context.Background(), data)
}

//UpdateOne inserts one document
func (dba *Client) UpdateOne(coll string, key string, value string, data interface{}) (*mongo.UpdateResult, error) {
	filter := bson.M{key: value}
	update := bson.M{"$set": data}

	return dba.client.Database(dba.Database).Collection(coll).UpdateOne(context.Background(), filter, update)
}

//UpdateOneByID inserts one document
func (dba *Client) UpdateOneByID(coll string, id string, data interface{}) (*mongo.UpdateResult, error) {
	filter := bson.M{"id": id}
	update := bson.M{"$set": data}

	return dba.client.Database(dba.Database).Collection(coll).UpdateOne(context.Background(), filter, update)
}

//Connect Connects to the xdcc database
func (dba *Client) Connect(host string) {
	// Set client options
	clientOptions := options.Client().ApplyURI(host)

	log.WithField("host", host).Info("Connecting to MongoDB instance")

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")
	dba.client = client
	dba.Connected = true
}
