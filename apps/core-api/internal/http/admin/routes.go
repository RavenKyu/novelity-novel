// Package admin registers the admin HTTP surface. Domain services stay outside this package.
package admin

import (
	"github.com/labstack/echo/v5"
	"net/http"
)

// Register supplies protected skeleton endpoints; real management belongs to feature epics.
func Register(group *echo.Group) {
	group.GET("/users", func(c *echo.Context) error {
		c.Response().Header().Set("Cache-Control", "private, no-store")
		return c.JSON(http.StatusOK, map[string]any{"users": []any{}})
	})
	group.GET("/healthz", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "surface": "admin"})
	})
}
