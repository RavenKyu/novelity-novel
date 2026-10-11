// Package studio registers the studio HTTP surface. Domain services stay outside this package.
package studio

import (
	"github.com/labstack/echo/v5"
	"net/http"
)

// Register supplies a connectivity endpoint; domain endpoints are added by their feature epics.
func Register(group *echo.Group) {
	group.GET("/healthz", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "surface": "studio"})
	})
}
