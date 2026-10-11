package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"novelity-novel/core-api/internal/http/auth"
	"novelity-novel/core-api/internal/identity"
)

type storeRepo struct {
	sessionRepo
	err            error
	calls          int
	hash, audience string
}

func (r *storeRepo) SessionUser(_ context.Context, hash string, _ time.Time, audience string) (identity.User, error) {
	r.calls++
	r.hash, r.audience = hash, audience
	return r.user, r.err
}

func TestStoreOptionalSession(t *testing.T) {
	for _, tc := range []struct {
		name, cookie string
		err          error
		want, calls  int
		loggedIn     bool
	}{
		{"guest", "", nil, 200, 0, false},
		{"empty", "nn_session=", nil, 200, 0, false},
		{"expired-or-invalid", "nn_session=stale", identity.ErrNoSession, 200, 1, false},
		{"admin-cookie", "nn_admin_session=admin", nil, 200, 0, false},
		{"public-user", "nn_session=live", nil, 200, 1, true},
		{"repository-error", "nn_session=live", errors.New("database unavailable"), 500, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &storeRepo{sessionRepo: sessionRepo{identity.User{ID: "reader", Name: "독자"}}, err: tc.err}
			svc := auth.New(auth.Config{Origins: map[string]string{"store": "http://localhost:8001"}}, repo)
			e := New(Deps{DB: fakePinger{}, Auth: svc})
			req := httptest.NewRequest("GET", "http://localhost:8001/api/store/session", nil)
			req.Header.Set("Cookie", tc.cookie)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != tc.want || repo.calls != tc.calls {
				t.Fatalf("status=%d calls=%d, want %d/%d", rec.Code, repo.calls, tc.want, tc.calls)
			}
			if repo.calls > 0 {
				if repo.audience != "public" || repo.hash == "live" || repo.hash == "stale" {
					t.Fatalf("invalid session lookup: %s/%s", repo.audience, repo.hash)
				}
			}
			if rec.Code == 200 {
				var body struct {
					User *identity.User `json:"user"`
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if (body.User != nil) != tc.loggedIn {
					t.Fatalf("unexpected user: %s", rec.Body.String())
				}
				if tc.loggedIn {
					if body.User.ID != "reader" || !body.User.HasRole(identity.Reader) {
						t.Fatalf("missing effective reader: %+v", body.User)
					}
					sum := sha256.Sum256([]byte("live"))
					if repo.hash != hex.EncodeToString(sum[:]) {
						t.Fatal("session token not hashed")
					}
				}
				if rec.Header().Get("Cache-Control") != "private, no-store" {
					t.Fatal("session response may be cached")
				}
			}
		})
	}
}
