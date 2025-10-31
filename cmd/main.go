package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"subservice/internal/app"
	"subservice/internal/config"
)

const defaultConfigPath = ""

func main() {
	lg := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	cfg, err := config.New(defaultConfigPath)
	if err != nil {
		lg.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app, err := app.New(ctx, cfg, lg)
	if err != nil {
		lg.Error("Failed to initialize app", slog.Any("error", err))
		os.Exit(1)
	}

	app.Run(ctx)
}
