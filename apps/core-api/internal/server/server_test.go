package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
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
