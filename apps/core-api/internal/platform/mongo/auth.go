package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"novelity-novel/core-api/internal/identity"
)

const (
	usersColl    = "users"
	sessionsColl = "sessions"
)

type userDoc struct {
	ID          bson.ObjectID   `bson:"_id,omitempty"`
	GoogleSub   string          `bson:"googleSub"`
	Email       string          `bson:"email"`
	Name        string          `bson:"name"`
	Picture     string          `bson:"picture"`
	CreatedAt   time.Time       `bson:"createdAt"`
	LastLoginAt time.Time       `bson:"lastLoginAt"`
	Roles       []identity.Role `bson:"roles"`
}

func (d userDoc) user() identity.User {
	return identity.User{
		ID: d.ID.Hex(), Email: d.Email, Name: d.Name, Picture: d.Picture,
		CreatedAt: d.CreatedAt, LastLoginAt: d.LastLoginAt, Roles: defaultRoles(d.Roles),
	}
}

type sessionDoc struct {
	Audience  string        `bson:"audience"`
	TokenHash string        `bson:"_id"`
	UserID    bson.ObjectID `bson:"userId"`
	CreatedAt time.Time     `bson:"createdAt"`
	ExpiresAt time.Time     `bson:"expiresAt"`
}

// EnsureIndexes creates the indexes the auth collections rely on. Idempotent.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	if _, err := s.db.Collection(usersColl).Indexes().CreateOne(ctx, driver.IndexModel{
		Keys:    bson.D{{Key: "googleSub", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("users index: %w", err)
	}
	// TTL: MongoDB deletes sessions once expiresAt passes (checked about once a minute,
	// so SessionUser also filters on expiresAt).
	if _, err := s.db.Collection(sessionsColl).Indexes().CreateOne(ctx, driver.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}); err != nil {
		return fmt.Errorf("sessions index: %w", err)
	}
	return nil
}

func (s *Store) UpsertGoogleUser(ctx context.Context, p identity.Profile, now time.Time) (identity.User, error) {
	var d userDoc
	err := s.db.Collection(usersColl).FindOneAndUpdate(ctx,
		bson.M{"googleSub": p.Subject},
		bson.M{
			"$set":         bson.M{"email": p.Email, "name": p.Name, "picture": p.Picture, "lastLoginAt": now},
			"$setOnInsert": bson.M{"createdAt": now, "roles": []identity.Role{identity.Reader}},
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&d)
	if err != nil {
		return identity.User{}, fmt.Errorf("upsert user: %w", err)
	}
	return d.user(), nil
}

func (s *Store) CreateSession(ctx context.Context, sess identity.Session) error {
	uid, err := bson.ObjectIDFromHex(sess.UserID)
	if err != nil {
		return fmt.Errorf("session user id: %w", err)
	}
	_, err = s.db.Collection(sessionsColl).InsertOne(ctx, sessionDoc{
		Audience: sess.Audience, TokenHash: sess.TokenHash, UserID: uid, CreatedAt: sess.CreatedAt, ExpiresAt: sess.ExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (s *Store) SessionUser(ctx context.Context, tokenHash string, now time.Time, audience string) (identity.User, error) {
	var sess sessionDoc
	err := s.db.Collection(sessionsColl).FindOne(ctx,
		sessionFilter(tokenHash, audience, now),
	).Decode(&sess)
	if errors.Is(err, driver.ErrNoDocuments) {
		return identity.User{}, identity.ErrNoSession
	}
	if err != nil {
		return identity.User{}, fmt.Errorf("find session: %w", err)
	}
	var u userDoc
	err = s.db.Collection(usersColl).FindOne(ctx, bson.M{"_id": sess.UserID}).Decode(&u)
	if errors.Is(err, driver.ErrNoDocuments) {
		return identity.User{}, identity.ErrNoSession
	}
	if err != nil {
		return identity.User{}, fmt.Errorf("find session user: %w", err)
	}
	return u.user(), nil
}

func (s *Store) DeleteSession(ctx context.Context, tokenHash string, audience string) error {
	if _, err := s.db.Collection(sessionsColl).DeleteOne(ctx, sessionFilter(tokenHash, audience, time.Time{})); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// Legacy users with no roles are readers; login never replaces existing author roles.
func defaultRoles(roles []identity.Role) []identity.Role {
	if len(roles) == 0 {
		return []identity.Role{identity.Reader}
	}
	return roles
}

// Existing sessions without audience remain public, never admin.
func sessionFilter(hash, audience string, now time.Time) bson.M {
	filter := bson.M{"_id": hash}
	if audience == "public" {
		filter["$or"] = bson.A{bson.M{"audience": "public"}, bson.M{"audience": ""}, bson.M{"audience": bson.M{"$exists": false}}}
	} else {
		filter["audience"] = audience
	}
	if !now.IsZero() {
		filter["expiresAt"] = bson.M{"$gt": now}
	}
	return filter
}
