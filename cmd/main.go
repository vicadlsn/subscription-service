package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subservice/internal/config"
	"subservice/internal/repository"
	"subservice/internal/service"
	"subservice/internal/transport"
	"subservice/pkg/postgres"

	"github.com/go-playground/validator/v10"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := postgres.New(ctx, cfg.Database)
	if err != nil {
		lg.Error("Failed to connect to postgres", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewSubscriptionRepository(pool)
	svc := service.NewSubscriptionService(repo, lg)
	v := validator.New(validator.WithRequiredStructEnabled())

	router := transport.NewRouter(svc, lg, v)

	addr := cfg.HTTP.Host + ":" + cfg.HTTP.Port
	server := http.Server{
		Addr:    addr,
		Handler: router,
	}

	lg.Info("Starting http server", slog.String("addr", addr))
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Error("HTTP server error", slog.Any("error", err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	lg.Info("Shutdown signal received")
	cancel()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctxShutdown); err != nil {
		lg.Error("Failed to shutdown http server", slog.Any("error", err))
	}
}
