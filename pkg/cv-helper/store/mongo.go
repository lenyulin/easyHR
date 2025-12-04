package store

import (
	"context"
	"easyHR/pkg/cv-helper/model"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	client     *mongo.Client
	db         string
	collection string
}

func NewMongoStore(uri, db, collection string) (*MongoStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoStore{
		client:     client,
		db:         db,
		collection: collection,
	}, nil
}

func (s *MongoStore) Save(ctx context.Context, cv *model.CV) error {
	coll := s.client.Database(s.db).Collection(s.collection)
	_, err := coll.InsertOne(ctx, cv)
	return err
}

func (s *MongoStore) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
