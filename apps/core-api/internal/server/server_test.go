package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"novelity-novel/core-api/internal/http/auth"
	"novelity-novel/core-api/internal/identity"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestHealthz(t *testing.T) {
	e := New(Deps{DB: fakePinger{}})

	rec := get(t, e, "/api/healthz")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Body.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestHealthzDBDown(t *testing.T) {
	e := New(Deps{DB: fakePinger{err: errors.New("no route")}})

	rec := get(t, e, "/api/healthz")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if got, want := rec.Body.String(), "{\"db\":\"down\",\"status\":\"degraded\"}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestOnlySelectedSurfacesAreMounted(t *testing.T) {
	for _, tc := range []struct {
		surfaces []string
		path     string
		want     int
	}{
		{nil, "/api/admin/healthz", 404},
		{nil, "/api/store/healthz", 200},
		{nil, "/api/studio/healthz", 401},
		{[]string{"store"}, "/api/studio/healthz", 404},
		{[]string{"admin"}, "/api/store/healthz", 404},
		{[]string{"admin"}, "/api/admin/healthz", 401},
		{[]string{}, "/api/store/healthz", 404},
	} {
		e := New(Deps{DB: fakePinger{}, Surfaces: tc.surfaces})
		if rec := get(t, e, tc.path); rec.Code != tc.want {
			t.Fatalf("surfaces %v %s: %d, want %d", tc.surfaces, tc.path, rec.Code, tc.want)
		}
		if rec := get(t, e, "/api/healthz"); rec.Code != 200 {
			t.Fatal("common health disappeared")
		}
	}
}

type sessionRepo struct{ user identity.User }

func (r sessionRepo) UpsertGoogleUser(context.Context, identity.Profile, time.Time) (identity.User, error) {
	return r.user, nil
}
func (r sessionRepo) CreateSession(context.Context, identity.Session) error { return nil }
func (r sessionRepo) SessionUser(_ context.Context, _ string, _ time.Time, audience string) (identity.User, error) {
	if audience != "admin" {
		return identity.User{}, identity.ErrNoSession
	}
	return r.user, nil
}
func (r sessionRepo) DeleteSession(context.Context, string, string) error { return nil }

func TestAdminSurfaceAuthorization(t *testing.T) {
	for _, tc := range []struct {
		allow []string
		want  int
	}{{nil, 403}, {[]string{"admin@example.com"}, 200}} {
		repo := sessionRepo{identity.User{ID: "u", Email: "admin@example.com", Roles: []identity.Role{identity.Reader, identity.Admin}}}
		svc := auth.New(auth.Config{Origins: map[string]string{"admin": "https://admin.novelity.kr"}, Surfaces: []string{"admin"}, AdminEmails: tc.allow}, repo)
		e := New(Deps{DB: fakePinger{}, Auth: svc, Surfaces: []string{"admin"}})
		req := httptest.NewRequest("GET", "https://admin.novelity.kr/api/admin/healthz", nil)
		req.AddCookie(&http.Cookie{Name: auth.AdminSessionCookie, Value: "admin-token"})
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("allow %v: %d, want %d", tc.allow, rec.Code, tc.want)
		}
	}
}
