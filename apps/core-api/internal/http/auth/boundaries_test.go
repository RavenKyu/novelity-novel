package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"novelity-novel/core-api/internal/identity"
)

func TestReturnToValidation(t *testing.T) {
	h := newHarness(t)
	h.svc.cfg.Origins["store"] = "http://store.test"
	for _, target := range []string{"/works?tab=draft#chapter", "http://store.test/library", "http://studio.test/works"} {
		t.Run(target, func(t *testing.T) {
			rec := h.do("GET", "/api/auth/google/login?"+url.Values{"returnTo": {target}}.Encode(), nil)
			flow := cookie(rec, flowCookie)
			if flow == nil {
				t.Fatalf("login: %d %s", rec.Code, rec.Body.String())
			}
			authz, _ := url.Parse(rec.Header().Get("Location"))
			if got := authz.Query().Get("redirect_uri"); got != "http://studio.test/api/auth/google/callback" {
				t.Fatalf("callback URI changed by returnTo: %s", got)
			}
			q := h.idp.authorize(t, rec.Header().Get("Location"))
			rec = h.do("GET", "/api/auth/google/callback?"+q.Encode(), []*http.Cookie{flow})
			if rec.Header().Get("Location") != target || cookie(rec, SessionCookie) == nil {
				t.Fatalf("return: %d %s", rec.Code, rec.Header().Get("Location"))
			}
		})
	}
	for _, target := range []string{"//evil.test", "https://evil.test/", "http://studio.test.evil/", "http://studio.test@evil.test/", "http://admin.test/", "javascript:alert(1)", "/\\evil.test", "/%2f%2fevil.test", "/%5cevil.test", "/x%0d%0aLocation:evil", "works", "https://studio.test/", "http://studio.test:80/"} {
		rec := h.do("GET", "/api/auth/google/login?"+url.Values{"returnTo": {target}}.Encode(), nil)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("accepted %q: %d", target, rec.Code)
		}
	}
}

func TestFlowIntegrityExpiryAndOrigin(t *testing.T) {
	for _, mode := range []string{"tamper", "expired", "other origin"} {
		t.Run(mode, func(t *testing.T) {
			h := newHarness(t)
			h.svc.cfg.Origins["store"] = "http://store.test"
			rec := h.do("GET", "/api/auth/google/login?returnTo=/works", nil)
			flow := cookie(rec, flowCookie)
			q := h.idp.authorize(t, rec.Header().Get("Location"))
			switch mode {
			case "tamper":
				parts := strings.Split(flow.Value, ".")
				payload, _ := base64.RawURLEncoding.DecodeString(parts[0])
				var f loginFlow
				json.Unmarshal(payload, &f)
				f.ReturnTo = "https://evil.test/"
				payload, _ = json.Marshal(f)
				flow.Value = base64.RawURLEncoding.EncodeToString(payload) + "." + parts[1]
			case "expired":
				h.svc.now = func() time.Time { return time.Now().Add(flowTTL) }
			}
			request := httptest.NewRequest("GET", "http://studio.test/api/auth/google/callback?"+q.Encode(), nil)
			if mode == "other origin" {
				request.Host = "store.test"
			}
			request.AddCookie(flow)
			rec = httptest.NewRecorder()
			h.e.ServeHTTP(rec, request)
			if cookie(rec, SessionCookie) != nil || len(h.repo.sessions) != 0 || !strings.HasPrefix(rec.Header().Get("Location"), "/login?error=invalid_state") {
				t.Fatalf("flow accepted: %d %s", rec.Code, rec.Header().Get("Location"))
			}
		})
	}
}

func TestFailureKeepsReturnTo(t *testing.T) {
	h := newHarness(t)
	rec := h.do("GET", "/api/auth/google/login?returnTo=/works", nil)
	flow := cookie(rec, flowCookie)
	q := h.idp.authorize(t, rec.Header().Get("Location"))
	q.Set("error", "access_denied")
	rec = h.do("GET", "/api/auth/google/callback?"+q.Encode(), []*http.Cookie{flow})
	u, _ := url.Parse(rec.Header().Get("Location"))
	if u.Query().Get("error") != "denied" || u.Query().Get("returnTo") != "/works" {
		t.Fatalf("failure lost return: %s", u)
	}
}

func TestUnknownHostDoesNotChooseCallback(t *testing.T) {
	h := newHarness(t)
	req := httptest.NewRequest("GET", "http://evil.test/api/auth/google/login", nil)
	req.Header.Set("X-Forwarded-Host", "studio.test")
	rec := httptest.NewRecorder()
	h.e.ServeHTTP(rec, req)
	if rec.Code != 400 || cookie(rec, flowCookie) != nil {
		t.Fatalf("unknown host accepted: %d", rec.Code)
	}
}

func TestSharedPublicSession(t *testing.T) {
	h := newHarness(t)
	h.svc.cfg.Origins["store"] = "http://store.test"
	h.svc.cfg.SessionDomain = "novelity.kr"
	sess := cookie(h.login(t, nil), SessionCookie)
	if sess.Domain != "novelity.kr" {
		t.Fatalf("SSO domain: %+v", sess)
	}
	req := httptest.NewRequest("GET", "http://store.test/api/auth/me", nil)
	req.AddCookie(sess)
	rec := httptest.NewRecorder()
	h.e.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("store cannot use studio session: %d", rec.Code)
	}
}

func TestAdminSessionIsolationAndRevocation(t *testing.T) {
	h := newHarness(t)
	public := cookie(h.login(t, nil), SessionCookie)
	h.svc.cfg.Surfaces = []string{"admin"}
	h.svc.cfg.Origins = map[string]string{"admin": "http://studio.test"}
	h.svc.cfg.AdminEmails = []string{"writer@example.com"}
	h.svc.cfg.SessionDomain = "novelity.kr"
	rec := h.do("GET", "/api/auth/google/login", nil)
	flow := cookie(rec, "nn_admin_oauth")
	q := h.idp.authorize(t, rec.Header().Get("Location"))
	rec = h.do("GET", "/api/auth/google/callback?"+q.Encode(), []*http.Cookie{flow})
	admin := cookie(rec, AdminSessionCookie)
	if admin == nil || admin.Domain != "" || admin.MaxAge != 7200 || !admin.HttpOnly || admin.SameSite != http.SameSiteLaxMode {
		t.Fatalf("admin cookie: %+v", admin)
	}
	for _, sess := range h.repo.sessions {
		if sess.Audience == "admin" && sess.ExpiresAt.Sub(sess.CreatedAt) != 2*time.Hour {
			t.Fatalf("admin TTL: %+v", sess)
		}
	}
	if rec := h.do("GET", "/api/auth/me", []*http.Cookie{public}); rec.Code != 401 {
		t.Fatalf("admin accepted public cookie: %d", rec.Code)
	}
	renamed := *public
	renamed.Name = AdminSessionCookie
	if rec := h.do("GET", "/api/auth/me", []*http.Cookie{&renamed}); rec.Code != 401 {
		t.Fatalf("admin accepted public token: %d", rec.Code)
	}
	if rec := h.do("GET", "/api/auth/me", []*http.Cookie{admin}); rec.Code != 200 {
		t.Fatalf("admin rejected: %d", rec.Code)
	}
	h.svc.cfg.AdminEmails = nil
	if rec := h.do("GET", "/api/auth/me", []*http.Cookie{admin}); rec.Code != 403 {
		t.Fatalf("admin not revoked: %d", rec.Code)
	}
	h.svc.cfg.AdminEmails = []string{"writer@example.com"}
	if rec := h.do("POST", "/api/auth/logout", []*http.Cookie{public, admin}); rec.Code != 204 {
		t.Fatalf("logout: %d", rec.Code)
	}
	if _, ok := h.repo.sessions[hashToken(public.Value)]; !ok {
		t.Fatal("admin logout deleted public session")
	}
	if _, ok := h.repo.sessions[hashToken(admin.Value)]; ok {
		t.Fatal("admin logout kept admin session")
	}
}

func TestAdminLoginDeniesReader(t *testing.T) {
	h := newHarness(t)
	h.svc.cfg.Surfaces = []string{"admin"}
	h.svc.cfg.Origins = map[string]string{"admin": "http://studio.test"}
	rec := h.do("GET", "/api/auth/google/login", nil)
	q := h.idp.authorize(t, rec.Header().Get("Location"))
	rec = h.do("GET", "/api/auth/google/callback?"+q.Encode(), []*http.Cookie{cookie(rec, "nn_admin_oauth")})
	if rec.Header().Get("Location") != "/login?error=forbidden" || len(h.repo.sessions) != 0 {
		t.Fatalf("reader admin login accepted: %s", rec.Header().Get("Location"))
	}
}

func TestSecureCookiePolicies(t *testing.T) {
	s := New(Config{SessionDomain: "novelity.kr"}, newMemRepo())
	for _, space := range []string{"studio", "admin"} {
		cookie := s.sessionCookie(space, "https://studio.novelity.kr", "token", 100)
		if !cookie.Secure || !cookie.HttpOnly || cookie.Path != "/" {
			t.Fatalf("cookie flags: %+v", cookie)
		}
		if space == "admin" && cookie.Domain != "" {
			t.Fatal("admin cookie has a domain")
		}
	}
}

func TestSameOriginLogout(t *testing.T) {
	for _, origin := range []string{"http://studio.test", "http://store.test", "http://evil.test", "null"} {
		h := newHarness(t)
		sess := cookie(h.login(t, nil), SessionCookie)
		req := httptest.NewRequest("POST", "http://studio.test/api/auth/logout", nil)
		req.AddCookie(sess)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()
		h.e.ServeHTTP(rec, req)
		want := 403
		if origin == "http://studio.test" {
			want = 204
		}
		if rec.Code != want {
			t.Fatalf("%s: %d, want %d", origin, rec.Code, want)
		}
		if want == 403 {
			if _, ok := h.repo.sessions[hashToken(sess.Value)]; !ok {
				t.Fatal("cross-origin request deleted session")
			}
		}
	}
}

// Keep imports explicit for compile-time checks of middleware and domain roles.
var _ echo.MiddlewareFunc = RequireRole(identity.Author)
