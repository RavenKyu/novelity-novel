package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"novelity-novel/core-api/internal/auth"
)

const (
	usersColl    = "users"
	sessionsColl = "sessions"
)

type userDoc struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	GoogleSub   string        `bson:"googleSub"`
	Email       string        `bson:"email"`
	Name        string        `bson:"name"`
	Picture     string        `bson:"picture"`
	CreatedAt   time.Time     `bson:"createdAt"`
	LastLoginAt time.Time     `bson:"lastLoginAt"`
}

func (d userDoc) user() auth.User {
	return auth.User{
		ID: d.ID.Hex(), Email: d.Email, Name: d.Name, Picture: d.Picture,
		CreatedAt: d.CreatedAt, LastLoginAt: d.LastLoginAt,
	}
}

type sessionDoc struct {
	TokenHash string        `bson:"_id"`
	UserID    bson.ObjectID `bson:"userId"`
	CreatedAt time.Time     `bson:"createdAt"`
	ExpiresAt time.Time     `bson:"expiresAt"`
}

// EnsureIndexes creates the indexes the auth collections rely on. Idempotent.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	if _, err := s.db.Collection(usersColl).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "googleSub", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("users index: %w", err)
	}
	// TTL: MongoDB deletes sessions once expiresAt passes (checked about once a minute,
	// so SessionUser also filters on expiresAt).
	if _, err := s.db.Collection(sessionsColl).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}); err != nil {
		return fmt.Errorf("sessions index: %w", err)
	}
	return nil
}

func (s *Store) UpsertGoogleUser(ctx context.Context, p auth.Profile, now time.Time) (auth.User, error) {
	var d userDoc
	err := s.db.Collection(usersColl).FindOneAndUpdate(ctx,
		bson.M{"googleSub": p.Subject},
		bson.M{
			"$set":         bson.M{"email": p.Email, "name": p.Name, "picture": p.Picture, "lastLoginAt": now},
			"$setOnInsert": bson.M{"createdAt": now},
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&d)
	if err != nil {
		return auth.User{}, fmt.Errorf("upsert user: %w", err)
	}
	return d.user(), nil
}

func (s *Store) CreateSession(ctx context.Context, sess auth.Session) error {
	uid, err := bson.ObjectIDFromHex(sess.UserID)
	if err != nil {
		return fmt.Errorf("session user id: %w", err)
	}
	_, err = s.db.Collection(sessionsColl).InsertOne(ctx, sessionDoc{
		TokenHash: sess.TokenHash, UserID: uid, CreatedAt: sess.CreatedAt, ExpiresAt: sess.ExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (s *Store) SessionUser(ctx context.Context, tokenHash string, now time.Time) (auth.User, error) {
	var sess sessionDoc
	err := s.db.Collection(sessionsColl).FindOne(ctx,
		bson.M{"_id": tokenHash, "expiresAt": bson.M{"$gt": now}},
	).Decode(&sess)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return auth.User{}, auth.ErrNoSession
	}
	if err != nil {
		return auth.User{}, fmt.Errorf("find session: %w", err)
	}
	var u userDoc
	err = s.db.Collection(usersColl).FindOne(ctx, bson.M{"_id": sess.UserID}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return auth.User{}, auth.ErrNoSession
	}
	if err != nil {
		return auth.User{}, fmt.Errorf("find session user: %w", err)
	}
	return u.user(), nil
}

func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	if _, err := s.db.Collection(sessionsColl).DeleteOne(ctx, bson.M{"_id": tokenHash}); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
