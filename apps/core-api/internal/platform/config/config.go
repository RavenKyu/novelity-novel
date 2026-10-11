// Package config reads process configuration and validates deployment boundaries.
package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	Addr               string
	MongoURI           string
	MongoDB            string
	Origins            map[string]string
	Surfaces           string
	SessionDomain      string
	GoogleClientID     string
	GoogleClientSecret string
	SessionTTL         time.Duration
	AdminEmails        []string
}

func Load() Config {
	return Config{
		Addr:     net.JoinHostPort(os.Getenv("BIND_HOST"), env("PORT", "8080")),
		MongoURI: env("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:  env("MONGO_DB", "novelity"),
		Origins: map[string]string{
			"studio": env("STUDIO_URL", "http://localhost:8000"),
			"store":  env("BOOKSTORE_URL", "http://localhost:8001"),
			"admin":  env("ADMIN_URL", "http://localhost:8002"),
		},
		Surfaces:           env("API_SURFACES", "studio,store"),
		SessionDomain:      strings.TrimPrefix(os.Getenv("SESSION_DOMAIN"), "."),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		SessionTTL:         duration("SESSION_TTL", 7*24*time.Hour),
		AdminEmails:        strings.Split(os.Getenv("ADMIN_EMAILS"), ","),
	}
}

// ParseSurfaces fails closed on unknown names, empty selections and mixed public/admin instances.
func ParseSurfaces(value string) ([]string, error) {
	seen := map[string]bool{}
	var result []string
	for _, part := range strings.Split(value, ",") {
		name := strings.TrimSpace(part)
		if name != "studio" && name != "store" && name != "admin" {
			return nil, fmt.Errorf("unknown API surface %q", name)
		}
		if seen[name] {
			return nil, fmt.Errorf("duplicate API surface %q", name)
		}
		seen[name] = true
		result = append(result, name)
	}
	if seen["admin"] && len(result) != 1 {
		return nil, fmt.Errorf("admin must run in a separate instance")
	}
	return result, nil
}

func (c Config) Validate(surfaces []string) error {
	hosts := map[string]bool{}
	for _, surface := range surfaces {
		raw := c.Origins[surface]
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" {
			return fmt.Errorf("%s URL must be an http(s) origin without path, query or credentials", surface)
		}
		if hosts[u.Host] {
			return fmt.Errorf("surface origins must have distinct hosts (including ports)")
		}
		hosts[u.Host] = true
		if surface != "admin" && c.SessionDomain != "" {
			if u.Scheme != "https" || (u.Hostname() != c.SessionDomain && !strings.HasSuffix(u.Hostname(), "."+c.SessionDomain)) {
				return fmt.Errorf("SESSION_DOMAIN must contain every public HTTPS origin")
			}
		}
	}
	return nil
}

func duration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		slog.Warn("invalid duration, using default", "key", key, "default", fallback)
		return fallback
	}
	return d
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
