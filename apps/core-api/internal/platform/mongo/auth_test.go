package mongo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"novelity-novel/core-api/internal/identity"
)

// newTestStore connects to MONGO_TEST_URI (skipping when unset) and uses a
// throwaway database dropped after the test.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set")
	}
	ctx := context.Background()
	s, err := Connect(uri, fmt.Sprintf("novelity_test_%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = s.db.Drop(ctx)
		_ = s.Close(ctx)
	})
	if err := s.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestUpsertGoogleUser(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	t0 := time.Now().UTC().Truncate(time.Millisecond)

	first, err := s.UpsertGoogleUser(ctx, identity.Profile{Subject: "sub-1", Email: "a@example.com", Name: "A"}, t0)
	if err != nil {
		t.Fatal(err)
	}
	t1 := t0.Add(time.Hour)
	again, err := s.UpsertGoogleUser(ctx, identity.Profile{Subject: "sub-1", Email: "a2@example.com", Name: "A2"}, t1)
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != first.ID {
		t.Fatalf("same Google subject got a new user: %s vs %s", again.ID, first.ID)
	}
	if again.Email != "a2@example.com" || again.Name != "A2" {
		t.Fatalf("profile not refreshed: %+v", again)
	}
	if !again.CreatedAt.Equal(t0) || !again.LastLoginAt.Equal(t1) {
		t.Fatalf("createdAt=%v lastLoginAt=%v, want %v / %v", again.CreatedAt, again.LastLoginAt, t0, t1)
	}
}

func TestSessions(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	u, err := s.UpsertGoogleUser(ctx, identity.Profile{Subject: "sub-1", Email: "a@example.com"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, identity.Session{TokenHash: "h1", UserID: u.ID, CreatedAt: now, ExpiresAt: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}

	got, err := s.SessionUser(ctx, "h1", now, "public")
	if err != nil || got.ID != u.ID {
		t.Fatalf("SessionUser = %+v, %v; want user %s", got, err, u.ID)
	}
	if _, err := s.SessionUser(ctx, "h1", now.Add(2*time.Hour), "public"); !errors.Is(err, identity.ErrNoSession) {
		t.Fatalf("expired session: err = %v, want ErrNoSession", err)
	}
	if _, err := s.SessionUser(ctx, "missing", now, "public"); !errors.Is(err, identity.ErrNoSession) {
		t.Fatalf("missing session: err = %v, want ErrNoSession", err)
	}

	if err := s.DeleteSession(ctx, "h1", "public"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SessionUser(ctx, "h1", now, "public"); !errors.Is(err, identity.ErrNoSession) {
		t.Fatalf("deleted session: err = %v, want ErrNoSession", err)
	}
}

func TestLegacyUserRoles(t *testing.T) {
	if got := (userDoc{}).user(); len(got.Roles) != 1 || !got.HasRole(identity.Reader) {
		t.Fatalf("legacy roles: %v", got.Roles)
	}
}

func TestLoginPreservesAuthorRoles(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	p := identity.Profile{Subject: "author", Email: "author@example.com"}
	first, err := s.UpsertGoogleUser(ctx, p, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !first.HasRole(identity.Reader) {
		t.Fatalf("new user roles: %v", first.Roles)
	}
	_, err = s.db.Collection(usersColl).UpdateOne(ctx, bson.M{"googleSub": p.Subject}, bson.M{"$set": bson.M{"roles": []identity.Role{identity.Reader, identity.Author}}})
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.UpsertGoogleUser(ctx, p, time.Now())
	if err != nil || !again.HasRole(identity.Author) {
		t.Fatalf("roles lost on login: %+v, %v", again, err)
	}
}

func TestSessionAudiences(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	now := time.Now()
	u, err := s.UpsertGoogleUser(ctx, identity.Profile{Subject: "audiences", Email: "a@example.com"}, now)
	if err != nil {
		t.Fatal(err)
	}
	for _, audience := range []string{"public", "admin", ""} {
		hash := "hash-" + audience
		if err := s.CreateSession(ctx, identity.Session{TokenHash: hash, UserID: u.ID, Audience: audience, CreatedAt: now, ExpiresAt: now.Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
		for _, requested := range []string{"public", "admin"} {
			_, err := s.SessionUser(ctx, hash, now, requested)
			allowed := audience == requested || (audience == "" && requested == "public")
			if allowed && err != nil {
				t.Fatalf("%q as %q: %v", audience, requested, err)
			}
			if !allowed && !errors.Is(err, identity.ErrNoSession) {
				t.Fatalf("cross-audience accepted: %q as %q", audience, requested)
			}
		}
		wrong := "admin"
		if audience == "admin" {
			wrong = "public"
		}
		if err := s.DeleteSession(ctx, hash, wrong); err != nil {
			t.Fatal(err)
		}
		requested := audience
		if requested == "" {
			requested = "public"
		}
		if _, err := s.SessionUser(ctx, hash, now, requested); err != nil {
			t.Fatalf("cross-audience deletion removed token: %v", err)
		}
	}
	// A document from ADR-001 has no audience key at all.
	uid, _ := bson.ObjectIDFromHex(u.ID)
	_, err = s.db.Collection(sessionsColl).InsertOne(ctx, bson.M{"_id": "legacy", "userId": uid, "createdAt": now, "expiresAt": now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SessionUser(ctx, "legacy", now, "public"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SessionUser(ctx, "legacy", now, "admin"); !errors.Is(err, identity.ErrNoSession) {
		t.Fatalf("legacy token became admin: %v", err)
	}
}
