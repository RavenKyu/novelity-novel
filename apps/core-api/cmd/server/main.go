package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"

	"novelity-novel/core-api/internal/auth"
	"novelity-novel/core-api/internal/config"
	"novelity-novel/core-api/internal/server"
	"novelity-novel/core-api/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := store.Connect(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = st.Close(ctx)
	}()

	idxCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := st.EnsureIndexes(idxCtx); err != nil {
		return err
	}

	authCfg := auth.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		Issuer:       auth.GoogleIssuer,
		PublicURL:    cfg.PublicURL,
		SessionTTL:   cfg.SessionTTL,
	}
	if !authCfg.Enabled() {
		slog.Warn("GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET not set; Google login is disabled")
	}

	sc := echo.StartConfig{Address: cfg.Addr, HideBanner: true}
	return sc.Start(ctx, server.New(server.Deps{DB: st, Auth: auth.New(authCfg, st)}))
}
