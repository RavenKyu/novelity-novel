// Package store registers the store HTTP surface. Domain services stay outside this package.
package store

import (
	"github.com/labstack/echo/v5"
	"net/http"

	"novelity-novel/core-api/internal/http/auth"
)

// Register supplies connectivity and optional-session state; catalog endpoints come later.
func Register(group *echo.Group) {
	group.GET("/healthz", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "surface": "store"})
	})
	// Session state is personalized even when the visitor is anonymous.
	group.GET("/session", func(c *echo.Context) error {
		c.Response().Header().Set("Cache-Control", "private, no-store")
		user := auth.CurrentUser(c)
		if user.ID == "" {
			return c.JSON(http.StatusOK, map[string]any{"user": nil})
		}
		return c.JSON(http.StatusOK, map[string]any{"user": user})
	})
}
