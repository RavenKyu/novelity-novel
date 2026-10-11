package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"

	"novelity-novel/core-api/internal/http/auth"
	"novelity-novel/core-api/internal/platform/config"
	store "novelity-novel/core-api/internal/platform/mongo"
	"novelity-novel/core-api/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	flags := flag.NewFlagSet("core-api", flag.ContinueOnError)
	surfacesFlag := flags.String("surfaces", cfg.Surfaces, "API groups: studio,store or admin (separate instance)")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}
	surfaces, err := config.ParseSurfaces(*surfacesFlag)
	if err != nil {
		return err
	}
	if err := cfg.Validate(surfaces); err != nil {
		return err
	}

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
		ClientID:      cfg.GoogleClientID,
		ClientSecret:  cfg.GoogleClientSecret,
		Issuer:        auth.GoogleIssuer,
		Origins:       cfg.Origins,
		Surfaces:      surfaces,
		SessionDomain: cfg.SessionDomain,
		SessionTTL:    cfg.SessionTTL,
		AdminEmails:   cfg.AdminEmails,
	}
	if !authCfg.Enabled() {
		slog.Warn("GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET not set; Google login is disabled")
	}

	sc := echo.StartConfig{Address: cfg.Addr, HideBanner: true}
	return sc.Start(ctx, server.New(server.Deps{DB: st, Auth: auth.New(authCfg, st), Surfaces: surfaces}))
}
