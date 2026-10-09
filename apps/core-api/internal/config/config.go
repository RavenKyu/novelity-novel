package config

import (
	"log/slog"
	"os"
	"time"
)

type Config struct {
	// Addr is the listen address, e.g. ":8080".
	Addr string
	// MongoURI is the MongoDB connection string.
	MongoURI string
	// MongoDB is the database name.
	MongoDB string
	// PublicURL is the browser-facing origin; the OAuth redirect URI is derived from it.
	PublicURL          string
	GoogleClientID     string
	GoogleClientSecret string
	SessionTTL         time.Duration
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		Addr:     ":" + env("PORT", "8080"),
		MongoURI: env("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:  env("MONGO_DB", "novelity"),

		PublicURL:          env("PUBLIC_URL", "http://localhost:8000"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		SessionTTL:         duration("SESSION_TTL", 7*24*time.Hour),
	}
}

func duration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		slog.Warn("invalid duration, using default", "key", key, "value", v, "default", fallback)
		return fallback
	}
	return d
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
