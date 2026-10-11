// Package server wires HTTP adapters to their domain and platform collaborators.
package server

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"net/http"

	"novelity-novel/core-api/internal/handler"
	"novelity-novel/core-api/internal/http/admin"
	"novelity-novel/core-api/internal/http/auth"
	"novelity-novel/core-api/internal/http/store"
	"novelity-novel/core-api/internal/http/studio"
	"novelity-novel/core-api/internal/identity"
)

type Deps struct {
	DB   handler.Pinger
	Auth *auth.Service
	// Nil selects public surfaces. An explicitly empty selection mounts no surface groups.
	Surfaces []string
}

func New(d Deps) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())
	api := e.Group("/api")
	api.GET("/healthz", handler.Health(d.DB))
	if d.Auth != nil {
		d.Auth.Register(api)
	}
	surfaces := d.Surfaces
	if surfaces == nil {
		surfaces = []string{"studio", "store"}
	}
	require := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
	}
	if d.Auth != nil {
		require = d.Auth.RequireSession
	}
	for _, surface := range surfaces {
		switch surface {
		case "studio":
			group := api.Group("/studio", require)
			if d.Auth != nil {
				group.Use(d.Auth.SameOrigin)
			}
			studio.Register(group)
		case "store":
			store.Register(api.Group("/store"))
		case "admin":
			group := api.Group("/admin", require, auth.RequireRole(identity.Admin))
			if d.Auth != nil {
				group.Use(d.Auth.SameOrigin)
			}
			admin.Register(group)
		}
	}
	return e
}
