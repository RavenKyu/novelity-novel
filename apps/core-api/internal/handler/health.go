package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

// Pinger reports whether a backing service is reachable.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Health reports whether the process is serving and its database is reachable.
func Health(db Pinger) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "degraded", "db": "down"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}
