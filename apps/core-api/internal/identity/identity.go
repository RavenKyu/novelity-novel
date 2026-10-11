// Package identity owns user, role and session contracts, independent of HTTP and storage.
package identity

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Role string

const (
	Reader Role = "reader"
	// Author grants eligibility to publish to the bookstore; studio access requires only a session.
	Author Role = "author"
	Admin  Role = "admin"
)

// ErrNoSession means the token matches no live session.
var ErrNoSession = errors.New("identity: no session")

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Picture     string    `json:"picture"`
	CreatedAt   time.Time `json:"createdAt"`
	LastLoginAt time.Time `json:"lastLoginAt"`
	Roles       []Role    `json:"roles"`
}

// Profile is the identity Google vouched for in a verified id_token.
type Profile struct {
	Subject string
	Email   string
	Name    string
	Picture string
}

type Session struct {
	Audience  string
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
	SessionUser(ctx context.Context, tokenHash string, now time.Time, audience string) (User, error)
	DeleteSession(ctx context.Context, tokenHash string, audience string) error
}

// EffectiveRoles keeps reader as the baseline and grants admin only from the current allowlist.
// Stored admin roles never bypass the allowlist; removing an email revokes existing sessions too.
func EffectiveRoles(user User, adminEmails []string) User {
	roles := []Role{Reader}
	for _, role := range user.Roles {
		if role == Author {
			roles = append(roles, Author)
			break
		}
	}
	for _, email := range adminEmails {
		if strings.EqualFold(strings.TrimSpace(email), user.Email) && user.Email != "" {
			roles = append(roles, Admin)
			break
		}
	}
	user.Roles = roles
	return user
}

func (u User) HasRole(want Role) bool {
	for _, role := range u.Roles {
		if role == want {
			return true
		}
	}
	return false
}
