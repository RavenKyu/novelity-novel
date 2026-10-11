// Package auth implements Google sign-in (OIDC authorization code flow with
// PKCE) and cookie sessions backed by an identity.Repo.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo/v5"
	"golang.org/x/oauth2"

	"novelity-novel/core-api/internal/identity"
)

const (
	// SessionCookie holds the raw session token; the repo only ever sees its hash.
	SessionCookie      = "nn_session"
	AdminSessionCookie = "nn_admin_session"
	AdminSessionTTL    = 2 * time.Hour
	// flowCookie carries state, nonce and PKCE verifier between login and callback.
	flowCookie = "nn_oauth"
	flowPath   = "/api/auth/google"
	flowTTL    = 10 * time.Minute

	userKey = "auth.user"

	// GoogleIssuer is Google's OIDC issuer.
	GoogleIssuer = "https://accounts.google.com"
)

type Config struct {
	ClientID     string
	ClientSecret string
	// Issuer is the OIDC issuer; GoogleIssuer in production.
	Issuer string
	// Origins lists explicitly trusted browser-facing origins by surface.
	Origins       map[string]string
	Surfaces      []string
	SessionDomain string
	SessionTTL    time.Duration
	AdminEmails   []string
}

// Enabled reports whether Google credentials are configured.
func (c Config) Enabled() bool { return c.ClientID != "" && c.ClientSecret != "" }

type Service struct {
	cfg     Config
	repo    identity.Repo
	now     func() time.Time
	flowKey []byte

	mu       sync.Mutex
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func New(cfg Config, repo identity.Repo) *Service {
	if len(cfg.Surfaces) == 0 {
		cfg.Surfaces = []string{"studio", "store"}
	}
	if cfg.SessionTTL <= 0 {
		cfg.SessionTTL = 7 * 24 * time.Hour
	}
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return &Service{cfg: cfg, repo: repo, now: time.Now, flowKey: key}
}

// Register mounts the auth routes under g (expected to be /api).
func (s *Service) Register(g *echo.Group) {
	g.GET("/auth/google/login", s.login)
	g.GET("/auth/google/callback", s.callback)
	g.GET("/auth/me", s.me, s.RequireSession)
	g.POST("/auth/logout", s.logout, s.SameOrigin)
}

// provider runs OIDC discovery once, on first use, so the server starts even
// when the issuer is unreachable; a failed discovery is retried next time.
func (s *Service) provider(ctx context.Context, origin string) (*oauth2.Config, *oidc.IDTokenVerifier, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.oauth != nil {
		oc := *s.oauth
		oc.RedirectURL = origin + flowPath + "/callback"
		return &oc, s.verifier, nil
	}
	p, err := oidc.NewProvider(ctx, s.cfg.Issuer)
	if err != nil {
		return nil, nil, fmt.Errorf("oidc discovery: %w", err)
	}
	s.oauth = &oauth2.Config{
		ClientID:     s.cfg.ClientID,
		ClientSecret: s.cfg.ClientSecret,
		Endpoint:     p.Endpoint(),

		Scopes: []string{oidc.ScopeOpenID, "email", "profile"},
	}
	s.verifier = p.Verifier(&oidc.Config{ClientID: s.cfg.ClientID})
	oc := *s.oauth
	oc.RedirectURL = origin + flowPath + "/callback"
	return &oc, s.verifier, nil
}

func (s *Service) login(c *echo.Context) error {
	space, origin, err := s.requestSpace(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_origin"})
	}
	returnTo, err := s.validReturnTo(c.QueryParam("returnTo"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_return_to"})
	}
	// A live session can return across public spaces without another Google round trip.
	// The same audience and current role policy applies as on authenticated API requests.
	name, audience := sessionIdentity(space)
	if cookie, err := c.Cookie(name); err == nil && cookie.Value != "" {
		user, err := s.repo.SessionUser(c.Request().Context(), hashToken(cookie.Value), s.now(), audience)
		if err == nil {
			user = identity.EffectiveRoles(user, s.cfg.AdminEmails)
			if space == "admin" && !user.HasRole(identity.Admin) {
				return s.loginError(c, "forbidden", returnTo)
			}
			return c.Redirect(http.StatusFound, returnTo)
		}
		if !errors.Is(err, identity.ErrNoSession) {
			return err
		}
	}
	if !s.cfg.Enabled() {
		return s.loginError(c, "not_configured", returnTo)
	}
	oc, _, err := s.provider(c.Request().Context(), origin)
	if err != nil {
		slog.Error("google login", "err", err)
		return s.loginError(c, "server_error", returnTo)
	}
	state, nonce, verifier := randomToken(), randomToken(), oauth2.GenerateVerifier()
	if err := s.setFlow(c, space, loginFlow{State: state, Nonce: nonce, Verifier: verifier, ReturnTo: returnTo, Origin: origin, ExpiresAt: s.now().Add(flowTTL).Unix()}); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_return_to"})
	}
	url := oc.AuthCodeURL(state,
		oauth2.S256ChallengeOption(verifier),
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
	return c.Redirect(http.StatusFound, url)
}

func (s *Service) callback(c *echo.Context) error {
	ctx := c.Request().Context()
	space, origin, err := s.requestSpace(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_origin"})
	}
	flow, ok := s.takeFlow(c, space, origin)
	returnTo := flow.ReturnTo
	if !ok || subtle.ConstantTimeCompare([]byte(flow.State), []byte(c.QueryParam("state"))) != 1 {
		return s.loginError(c, "invalid_state", returnTo)
	}
	if c.QueryParam("error") != "" {
		return s.loginError(c, "denied", returnTo)
	}
	oc, idv, err := s.provider(ctx, origin)
	if err != nil {
		slog.Error("google callback", "err", err)
		return s.loginError(c, "server_error", returnTo)
	}
	tok, err := oc.Exchange(ctx, c.QueryParam("code"), oauth2.VerifierOption(flow.Verifier))
	if err != nil {
		slog.Warn("google code exchange", "err", err)
		return s.loginError(c, "exchange_failed", returnTo)
	}
	raw, _ := tok.Extra("id_token").(string)
	idt, err := idv.Verify(ctx, raw)
	if err != nil || subtle.ConstantTimeCompare([]byte(idt.Nonce), []byte(flow.Nonce)) != 1 {
		slog.Warn("google id_token", "err", err)
		return s.loginError(c, "invalid_token", returnTo)
	}
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := idt.Claims(&claims); err != nil {
		return s.loginError(c, "invalid_token", returnTo)
	}
	if claims.Email == "" || !claims.EmailVerified {
		return s.loginError(c, "unverified_email", returnTo)
	}

	now := s.now()
	user, err := s.repo.UpsertGoogleUser(ctx, identity.Profile{
		Subject: idt.Subject, Email: claims.Email, Name: claims.Name, Picture: claims.Picture,
	}, now)
	if err != nil {
		slog.Error("upsert user", "err", err)
		return s.loginError(c, "server_error", returnTo)
	}
	user = identity.EffectiveRoles(user, s.cfg.AdminEmails)
	if space == "admin" && !user.HasRole(identity.Admin) {
		return s.loginError(c, "forbidden", returnTo)
	}
	ttl := s.cfg.SessionTTL
	audience := "public"
	if space == "admin" {
		ttl = AdminSessionTTL
		audience = "admin"
	}
	token := randomToken()
	if err := s.repo.CreateSession(ctx, identity.Session{
		TokenHash: hashToken(token), UserID: user.ID, CreatedAt: now, ExpiresAt: now.Add(ttl), Audience: audience,
	}); err != nil {
		slog.Error("create session", "err", err)
		return s.loginError(c, "server_error", returnTo)
	}
	c.SetCookie(s.sessionCookie(space, origin, token, int(ttl.Seconds())))
	return c.Redirect(http.StatusFound, returnTo)
}

func (s *Service) me(c *echo.Context) error {
	return c.JSON(http.StatusOK, CurrentUser(c))
}

func (s *Service) logout(c *echo.Context) error {
	space, origin, err := s.requestSpace(c)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	name, audience := sessionIdentity(space)
	if ck, err := c.Cookie(name); err == nil {
		if err := s.repo.DeleteSession(c.Request().Context(), hashToken(ck.Value), audience); err != nil {
			return err
		}
	}
	c.SetCookie(s.sessionCookie(space, origin, "", -1))
	return c.NoContent(http.StatusNoContent)
}

func sessionIdentity(space string) (string, string) {
	if space == "admin" {
		return AdminSessionCookie, "admin"
	}
	return SessionCookie, "public"
}

// RequireSession rejects requests without a live session with 401 and makes
// the user available to handlers via CurrentUser.
func (s *Service) RequireSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		space, _, err := s.requestSpace(c)
		if err != nil {
			return unauthorized(c)
		}
		name, audience := sessionIdentity(space)
		ck, err := c.Cookie(name)
		if err != nil || ck.Value == "" {
			return unauthorized(c)
		}
		user, err := s.repo.SessionUser(c.Request().Context(), hashToken(ck.Value), s.now(), audience)
		if errors.Is(err, identity.ErrNoSession) {
			return unauthorized(c)
		}
		if err != nil {
			return err
		}
		user = identity.EffectiveRoles(user, s.cfg.AdminEmails)
		if space == "admin" && !user.HasRole(identity.Admin) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
		c.Set(userKey, user)
		return next(c)
	}
}

// CurrentUser returns the user RequireSession attached to c.
func CurrentUser(c *echo.Context) identity.User {
	u, _ := c.Get(userKey).(identity.User)
	return u
}

func (s *Service) sessionCookie(space, origin, value string, maxAge int) *http.Cookie {
	name, _ := sessionIdentity(space)
	domain := s.cfg.SessionDomain
	if space == "admin" {
		domain = ""
	}
	return &http.Cookie{
		Name: name, Domain: domain,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   strings.HasPrefix(origin, "https://"),
		SameSite: http.SameSiteLaxMode,
	}
}

func unauthorized(c *echo.Context) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// RequireRole is stacked after RequireSession; authentication and authorization stay distinct.
func RequireRole(role identity.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			user := CurrentUser(c)
			if user.ID == "" {
				return unauthorized(c)
			}
			if !user.HasRole(role) {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
			}
			return next(c)
		}
	}
}
