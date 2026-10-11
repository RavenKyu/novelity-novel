package auth

import (
	"context"

	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"novelity-novel/core-api/internal/identity"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/labstack/echo/v5"
)

// fakeIdP is a minimal OIDC provider: discovery, JWKS and a token endpoint that
// enforces PKCE and signs an id_token carrying the nonce from the authorize request.
type fakeIdP struct {
	srv           *httptest.Server
	key           *rsa.PrivateKey
	email         string
	emailVerified bool

	mu      sync.Mutex
	pending map[string]authzRequest // code → request it answers
}

type authzRequest struct{ nonce, challenge string }

func newFakeIdP(t *testing.T) *fakeIdP {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	idp := &fakeIdP{key: key, email: "writer@example.com", emailVerified: true, pending: map[string]authzRequest{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                idp.srv.URL,
			"authorization_endpoint":                idp.srv.URL + "/authorize",
			"token_endpoint":                        idp.srv.URL + "/token",
			"jwks_uri":                              idp.srv.URL + "/jwks",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
			{Key: &key.PublicKey, KeyID: "k1", Algorithm: "RS256", Use: "sig"},
		}})
	})
	mux.HandleFunc("/token", idp.token(t))
	idp.srv = httptest.NewServer(mux)
	t.Cleanup(idp.srv.Close)
	return idp
}

func (idp *fakeIdP) token(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		idp.mu.Lock()
		req, ok := idp.pending[r.Form.Get("code")]
		idp.mu.Unlock()
		sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
		if !ok || base64.RawURLEncoding.EncodeToString(sum[:]) != req.challenge {
			http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
			return
		}
		claims, _ := json.Marshal(map[string]any{
			"iss": idp.srv.URL, "aud": "client-id", "sub": "google-sub-1",
			"iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix(),
			"nonce": req.nonce, "email": idp.email, "email_verified": idp.emailVerified,
			"name": "Writer", "picture": "https://example.com/p.png",
		})
		signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: idp.key},
			(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "k1"))
		if err != nil {
			t.Error(err)
			return
		}
		jws, _ := signer.Sign(claims)
		idToken, _ := jws.CompactSerialize()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "at", "token_type": "Bearer", "expires_in": 3600, "id_token": idToken,
		})
	}
}

// authorize plays the browser + Google consent: it reads the authorize URL the
// login handler redirected to, and returns the callback query Google would send.
func (idp *fakeIdP) authorize(t *testing.T, location string) url.Values {
	t.Helper()
	u, err := url.Parse(location)
	if err != nil || !strings.HasPrefix(location, idp.srv.URL+"/authorize") {
		t.Fatalf("login redirected to %q, want the IdP authorize endpoint", location)
	}
	q := u.Query()
	if q.Get("code_challenge_method") != "S256" {
		t.Fatalf("code_challenge_method = %q, want S256", q.Get("code_challenge_method"))
	}
	idp.mu.Lock()
	idp.pending["code-1"] = authzRequest{nonce: q.Get("nonce"), challenge: q.Get("code_challenge")}
	idp.mu.Unlock()
	return url.Values{"code": {"code-1"}, "state": {q.Get("state")}}
}

type memRepo struct {
	mu       sync.Mutex
	users    map[string]identity.User // by googleSub
	sessions map[string]identity.Session
}

func newMemRepo() *memRepo {
	return &memRepo{users: map[string]identity.User{}, sessions: map[string]identity.Session{}}
}

func (m *memRepo) UpsertGoogleUser(_ context.Context, p identity.Profile, now time.Time) (identity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[p.Subject]
	if !ok {
		u = identity.User{ID: "u" + p.Subject, CreatedAt: now}
	}
	u.Email, u.Name, u.Picture, u.LastLoginAt = p.Email, p.Name, p.Picture, now
	m.users[p.Subject] = u
	return u, nil
}

func (m *memRepo) CreateSession(_ context.Context, s identity.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.TokenHash] = s
	return nil
}

func (m *memRepo) SessionUser(_ context.Context, tokenHash string, now time.Time, audience string) (identity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[tokenHash]
	if !ok || !now.Before(s.ExpiresAt) || (s.Audience != audience && !(s.Audience == "" && audience == "public")) {
		return identity.User{}, identity.ErrNoSession
	}
	for _, u := range m.users {
		if u.ID == s.UserID {
			return u, nil
		}
	}
	return identity.User{}, identity.ErrNoSession
}

func (m *memRepo) DeleteSession(_ context.Context, tokenHash string, audience string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sess, ok := m.sessions[tokenHash]; ok && (sess.Audience == audience || (sess.Audience == "" && audience == "public")) {
		delete(m.sessions, tokenHash)
	}
	return nil
}

type harness struct {
	e    *echo.Echo
	idp  *fakeIdP
	repo *memRepo
	svc  *Service
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	idp := newFakeIdP(t)
	repo := newMemRepo()
	svc := New(Config{
		ClientID: "client-id", ClientSecret: "secret", Issuer: idp.srv.URL,
		Origins: map[string]string{"studio": "http://studio.test"}, SessionTTL: time.Hour,
	}, repo)
	e := echo.New()
	svc.Register(e.Group("/api"))
	return &harness{e: e, idp: idp, repo: repo, svc: svc}
}

func (h *harness) do(method, target string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	req.Host = "studio.test"
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.e.ServeHTTP(rec, req)
	return rec
}

func cookie(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// login runs login → IdP → callback and returns the callback response.
func (h *harness) login(t *testing.T, tamper func(url.Values)) *httptest.ResponseRecorder {
	t.Helper()
	rec := h.do(http.MethodGet, "/api/auth/google/login", nil)
	if rec.Code != http.StatusFound {
		t.Fatalf("login status = %d, want 302", rec.Code)
	}
	flow := cookie(rec, flowCookie)
	if flow == nil || !flow.HttpOnly {
		t.Fatalf("login must set an HttpOnly %s cookie", flowCookie)
	}
	q := h.idp.authorize(t, rec.Header().Get("Location"))
	if tamper != nil {
		tamper(q)
	}
	return h.do(http.MethodGet, "/api/auth/google/callback?"+q.Encode(), []*http.Cookie{flow})
}

func TestLoginCallbackMeLogout(t *testing.T) {
	h := newHarness(t)

	rec := h.login(t, nil)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/" {
		t.Fatalf("callback = %d → %q, want 302 → /", rec.Code, rec.Header().Get("Location"))
	}
	sess := cookie(rec, SessionCookie)
	if sess == nil || !sess.HttpOnly || sess.SameSite != http.SameSiteLaxMode || sess.Path != "/" {
		t.Fatalf("session cookie = %+v, want HttpOnly SameSite=Lax Path=/", sess)
	}
	for hash := range h.repo.sessions {
		if hash == sess.Value {
			t.Fatal("repo stores the raw session token; want only its hash")
		}
	}

	rec = h.do(http.MethodGet, "/api/auth/me", []*http.Cookie{sess})
	if rec.Code != http.StatusOK {
		t.Fatalf("me status = %d, want 200", rec.Code)
	}
	var me map[string]any
	json.Unmarshal(rec.Body.Bytes(), &me)
	if me["email"] != "writer@example.com" || me["name"] != "Writer" {
		t.Fatalf("me = %v", me)
	}

	rec = h.do(http.MethodPost, "/api/auth/logout", []*http.Cookie{sess})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want 204", rec.Code)
	}
	if c := cookie(rec, SessionCookie); c == nil || c.MaxAge >= 0 {
		t.Fatalf("logout must expire the session cookie, got %+v", c)
	}
	if rec = h.do(http.MethodGet, "/api/auth/me", []*http.Cookie{sess}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("me after logout = %d, want 401", rec.Code)
	}
}

func TestCallbackRejects(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*harness)
		tamper   func(url.Values)
		wantCode string
	}{
		{"forged state", nil, func(q url.Values) { q.Set("state", "forged") }, "invalid_state"},
		{"user denied consent", nil, func(q url.Values) { q.Del("code"); q.Set("error", "access_denied") }, "denied"},
		{"unverified email", func(h *harness) { h.idp.emailVerified = false }, nil, "unverified_email"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			if tt.setup != nil {
				tt.setup(h)
			}
			rec := h.login(t, tt.tamper)
			if got, want := rec.Header().Get("Location"), "/login?error="+tt.wantCode; got != want {
				t.Fatalf("redirect = %q, want %q", got, want)
			}
			if c := cookie(rec, SessionCookie); c != nil && c.MaxAge >= 0 {
				t.Fatal("rejected callback must not issue a session")
			}
			if len(h.repo.sessions) != 0 {
				t.Fatal("rejected callback must not create a session")
			}
		})
	}
}

func TestCallbackWithoutFlowCookie(t *testing.T) {
	h := newHarness(t)
	rec := h.do(http.MethodGet, "/api/auth/google/callback?code=x&state=y", nil)
	if got := rec.Header().Get("Location"); got != "/login?error=invalid_state" {
		t.Fatalf("redirect = %q, want /login?error=invalid_state", got)
	}
}

func TestMeWithoutSession(t *testing.T) {
	h := newHarness(t)
	if rec := h.do(http.MethodGet, "/api/auth/me", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	bogus := &http.Cookie{Name: SessionCookie, Value: "nope"}
	if rec := h.do(http.MethodGet, "/api/auth/me", []*http.Cookie{bogus}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("bogus session status = %d, want 401", rec.Code)
	}
}

func TestLoginNotConfigured(t *testing.T) {
	svc := New(Config{Origins: map[string]string{"studio": "http://studio.test"}, SessionTTL: time.Hour}, newMemRepo())
	e := echo.New()
	svc.Register(e.Group("/api"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://studio.test/api/auth/google/login", nil)
	e.ServeHTTP(rec, req)
	if got := rec.Header().Get("Location"); got != "/login?error=not_configured" {
		t.Fatalf("redirect = %q, want /login?error=not_configured", got)
	}
}

func TestRequireRole(t *testing.T) {
	e := echo.New()
	e.GET("/author", func(c *echo.Context) error { return c.NoContent(http.StatusOK) }, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			switch c.QueryParam("user") {
			case "reader":
				c.Set(userKey, identity.User{ID: "u", Roles: []identity.Role{identity.Reader}})
			case "author":
				c.Set(userKey, identity.User{ID: "u", Roles: []identity.Role{identity.Reader, identity.Author}})
			}
			return next(c)
		}
	}, RequireRole(identity.Author))
	for _, tc := range []struct {
		user string
		want int
	}{{"", 401}, {"reader", 403}, {"author", 200}} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest("GET", "/author?user="+tc.user, nil))
		if rec.Code != tc.want {
			t.Fatalf("%s: %d, want %d", tc.user, rec.Code, tc.want)
		}
	}
}
