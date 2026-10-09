// Package store owns the MongoDB connection and the collections built on it.
package store

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Store struct {
	client *mongo.Client
	db     *mongo.Database
}

// Connect opens a client to uri and selects database name. The driver connects
// lazily, so an unreachable server surfaces on first use (and in Ping), not here.
func Connect(uri, name string) (*Store, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	return &Store{client: client, db: client.Database(name)}, nil
}

func (s *Store) Ping(ctx context.Context) error {
	return s.client.Ping(ctx, readpref.Primary())
}

func (s *Store) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
