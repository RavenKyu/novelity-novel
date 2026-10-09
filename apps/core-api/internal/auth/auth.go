// Package auth implements Google sign-in (OIDC authorization code flow with
// PKCE) and cookie sessions backed by a Repo.
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
)

const (
	// SessionCookie holds the raw session token; the repo only ever sees its hash.
	SessionCookie = "nn_session"
	// flowCookie carries state, nonce and PKCE verifier between login and callback.
	flowCookie = "nn_oauth"
	flowPath   = "/api/auth/google"
	flowTTL    = 10 * time.Minute

	userKey = "auth.user"

	// GoogleIssuer is Google's OIDC issuer.
	GoogleIssuer = "https://accounts.google.com"
)

// ErrNoSession means the token matches no live session.
var ErrNoSession = errors.New("auth: no session")

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Picture     string    `json:"picture"`
	CreatedAt   time.Time `json:"createdAt"`
	LastLoginAt time.Time `json:"lastLoginAt"`
}

// Profile is the identity Google vouched for in a verified id_token.
type Profile struct {
	Subject string
	Email   string
	Name    string
	Picture string
}

type Session struct {
	TokenHash string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Repo interface {
	// UpsertGoogleUser creates the user on first sign-in, otherwise refreshes its profile.
	UpsertGoogleUser(ctx context.Context, p Profile, now time.Time) (User, error)
	CreateSession(ctx context.Context, s Session) error
	// SessionUser returns the session's user, or ErrNoSession if missing or expired at now.
	SessionUser(ctx context.Context, tokenHash string, now time.Time) (User, error)
	DeleteSession(ctx context.Context, tokenHash string) error
}

type Config struct {
	ClientID     string
	ClientSecret string
	// Issuer is the OIDC issuer; GoogleIssuer in production.
	Issuer string
	// PublicURL is the browser-facing origin, e.g. http://localhost:8000.
	PublicURL  string
	SessionTTL time.Duration
}

// Enabled reports whether Google credentials are configured.
func (c Config) Enabled() bool { return c.ClientID != "" && c.ClientSecret != "" }

func (c Config) secureCookies() bool { return strings.HasPrefix(c.PublicURL, "https://") }

type Service struct {
	cfg  Config
	repo Repo
	now  func() time.Time

	mu       sync.Mutex
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func New(cfg Config, repo Repo) *Service {
	return &Service{cfg: cfg, repo: repo, now: time.Now}
}

// Register mounts the auth routes under g (expected to be /api).
func (s *Service) Register(g *echo.Group) {
	g.GET("/auth/google/login", s.login)
	g.GET("/auth/google/callback", s.callback)
	g.GET("/auth/me", s.me, s.RequireSession)
	g.POST("/auth/logout", s.logout)
}

// provider runs OIDC discovery once, on first use, so the server starts even
// when the issuer is unreachable; a failed discovery is retried next time.
func (s *Service) provider(ctx context.Context) (*oauth2.Config, *oidc.IDTokenVerifier, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.oauth != nil {
		return s.oauth, s.verifier, nil
	}
	p, err := oidc.NewProvider(ctx, s.cfg.Issuer)
	if err != nil {
		return nil, nil, fmt.Errorf("oidc discovery: %w", err)
	}
	s.oauth = &oauth2.Config{
		ClientID:     s.cfg.ClientID,
		ClientSecret: s.cfg.ClientSecret,
		Endpoint:     p.Endpoint(),
		RedirectURL:  strings.TrimRight(s.cfg.PublicURL, "/") + flowPath + "/callback",
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}
	s.verifier = p.Verifier(&oidc.Config{ClientID: s.cfg.ClientID})
	return s.oauth, s.verifier, nil
}

func (s *Service) login(c *echo.Context) error {
	if !s.cfg.Enabled() {
		return loginError(c, "not_configured")
	}
	oc, _, err := s.provider(c.Request().Context())
	if err != nil {
		slog.Error("google login", "err", err)
		return loginError(c, "server_error")
	}
	state, nonce, verifier := randomToken(), randomToken(), oauth2.GenerateVerifier()
	c.SetCookie(&http.Cookie{
		Name:     flowCookie,
		Value:    state + "." + nonce + "." + verifier,
		Path:     flowPath,
		MaxAge:   int(flowTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.cfg.secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
	url := oc.AuthCodeURL(state,
		oauth2.S256ChallengeOption(verifier),
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
	return c.Redirect(http.StatusFound, url)
}

func (s *Service) callback(c *echo.Context) error {
	ctx := c.Request().Context()
	state, nonce, verifier, ok := s.takeFlow(c)
	if !ok || subtle.ConstantTimeCompare([]byte(state), []byte(c.QueryParam("state"))) != 1 {
		return loginError(c, "invalid_state")
	}
	if c.QueryParam("error") != "" {
		return loginError(c, "denied")
	}
	oc, idv, err := s.provider(ctx)
	if err != nil {
		slog.Error("google callback", "err", err)
		return loginError(c, "server_error")
	}
	tok, err := oc.Exchange(ctx, c.QueryParam("code"), oauth2.VerifierOption(verifier))
	if err != nil {
		slog.Warn("google code exchange", "err", err)
		return loginError(c, "exchange_failed")
	}
	raw, _ := tok.Extra("id_token").(string)
	idt, err := idv.Verify(ctx, raw)
	if err != nil || subtle.ConstantTimeCompare([]byte(idt.Nonce), []byte(nonce)) != 1 {
		slog.Warn("google id_token", "err", err)
		return loginError(c, "invalid_token")
	}
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := idt.Claims(&claims); err != nil {
		return loginError(c, "invalid_token")
	}
	if claims.Email == "" || !claims.EmailVerified {
		return loginError(c, "unverified_email")
	}

	now := s.now()
	user, err := s.repo.UpsertGoogleUser(ctx, Profile{
		Subject: idt.Subject, Email: claims.Email, Name: claims.Name, Picture: claims.Picture,
	}, now)
	if err != nil {
		slog.Error("upsert user", "err", err)
		return loginError(c, "server_error")
	}
	token := randomToken()
	if err := s.repo.CreateSession(ctx, Session{
		TokenHash: hashToken(token), UserID: user.ID, CreatedAt: now, ExpiresAt: now.Add(s.cfg.SessionTTL),
	}); err != nil {
		slog.Error("create session", "err", err)
		return loginError(c, "server_error")
	}
	c.SetCookie(s.sessionCookie(token, int(s.cfg.SessionTTL.Seconds())))
	return c.Redirect(http.StatusFound, "/")
}

// takeFlow reads and clears the login flow cookie (single use).
func (s *Service) takeFlow(c *echo.Context) (state, nonce, verifier string, ok bool) {
	ck, err := c.Cookie(flowCookie)
	if err != nil {
		return "", "", "", false
	}
	c.SetCookie(&http.Cookie{Name: flowCookie, Path: flowPath, MaxAge: -1, HttpOnly: true, Secure: s.cfg.secureCookies()})
	parts := strings.Split(ck.Value, ".")
	if len(parts) != 3 {
		return "", "", "", false
	}
	return parts[0], parts[1], parts[2], true
}

func (s *Service) me(c *echo.Context) error {
	return c.JSON(http.StatusOK, CurrentUser(c))
}

func (s *Service) logout(c *echo.Context) error {
	if ck, err := c.Cookie(SessionCookie); err == nil {
		if err := s.repo.DeleteSession(c.Request().Context(), hashToken(ck.Value)); err != nil {
			slog.Error("delete session", "err", err)
		}
	}
	c.SetCookie(s.sessionCookie("", -1))
	return c.NoContent(http.StatusNoContent)
}

// RequireSession rejects requests without a live session with 401 and makes
// the user available to handlers via CurrentUser.
func (s *Service) RequireSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ck, err := c.Cookie(SessionCookie)
		if err != nil || ck.Value == "" {
			return unauthorized(c)
		}
		user, err := s.repo.SessionUser(c.Request().Context(), hashToken(ck.Value), s.now())
		if errors.Is(err, ErrNoSession) {
			return unauthorized(c)
		}
		if err != nil {
			return err
		}
		c.Set(userKey, user)
		return next(c)
	}
}

// CurrentUser returns the user RequireSession attached to c.
func CurrentUser(c *echo.Context) User {
	u, _ := c.Get(userKey).(User)
	return u
}

func (s *Service) sessionCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookie,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   s.cfg.secureCookies(),
		SameSite: http.SameSiteLaxMode,
	}
}

func unauthorized(c *echo.Context) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

func loginError(c *echo.Context, code string) error {
	return c.Redirect(http.StatusFound, "/login?error="+code)
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
