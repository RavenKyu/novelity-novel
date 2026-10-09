package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"novelity-novel/core-api/internal/auth"
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

	first, err := s.UpsertGoogleUser(ctx, auth.Profile{Subject: "sub-1", Email: "a@example.com", Name: "A"}, t0)
	if err != nil {
		t.Fatal(err)
	}
	t1 := t0.Add(time.Hour)
	again, err := s.UpsertGoogleUser(ctx, auth.Profile{Subject: "sub-1", Email: "a2@example.com", Name: "A2"}, t1)
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

	u, err := s.UpsertGoogleUser(ctx, auth.Profile{Subject: "sub-1", Email: "a@example.com"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, auth.Session{TokenHash: "h1", UserID: u.ID, CreatedAt: now, ExpiresAt: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}

	got, err := s.SessionUser(ctx, "h1", now)
	if err != nil || got.ID != u.ID {
		t.Fatalf("SessionUser = %+v, %v; want user %s", got, err, u.ID)
	}
	if _, err := s.SessionUser(ctx, "h1", now.Add(2*time.Hour)); !errors.Is(err, auth.ErrNoSession) {
		t.Fatalf("expired session: err = %v, want ErrNoSession", err)
	}
	if _, err := s.SessionUser(ctx, "missing", now); !errors.Is(err, auth.ErrNoSession) {
		t.Fatalf("missing session: err = %v, want ErrNoSession", err)
	}

	if err := s.DeleteSession(ctx, "h1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SessionUser(ctx, "h1", now); !errors.Is(err, auth.ErrNoSession) {
		t.Fatalf("deleted session: err = %v, want ErrNoSession", err)
	}
}
