package auth

import (
	"errors"

	"github.com/labstack/echo/v5"
	"novelity-novel/core-api/internal/identity"
)

// OptionalSession attaches a live public user on the store surface. Missing or
// expired sessions remain anonymous; repository failures still reach the error handler.
func (s *Service) OptionalSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		cookie, err := c.Cookie(SessionCookie)
		if err != nil || cookie.Value == "" {
			return next(c)
		}
		user, err := s.repo.SessionUser(c.Request().Context(), hashToken(cookie.Value), s.now(), "public")
		if errors.Is(err, identity.ErrNoSession) {
			return next(c)
		}
		if err != nil {
			return err
		}
		c.Set(userKey, identity.EffectiveRoles(user, s.cfg.AdminEmails))
		return next(c)
	}
}
