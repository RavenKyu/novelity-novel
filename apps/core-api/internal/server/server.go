package server

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"novelity-novel/core-api/internal/auth"
	"novelity-novel/core-api/internal/handler"
)

// Deps are the collaborators routes need.
type Deps struct {
	DB   handler.Pinger
	Auth *auth.Service // nil leaves the auth routes unmounted
}

// New builds the Echo instance with middleware and every route registered.
func New(d Deps) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())

	api := e.Group("/api")
	api.GET("/healthz", handler.Health(d.DB))
	if d.Auth != nil {
		d.Auth.Register(api)
	}

	return e
}
