package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"subservice/internal/config"
	"subservice/internal/repository"
	"subservice/internal/service"
	"subservice/internal/transport"
	"subservice/pkg/postgres"

	"github.com/go-playground/validator/v10"
)

type App struct {
	db     postgres.PgxClient
	router http.Handler
	server *http.Server
	lg     *slog.Logger
}

func New(ctx context.Context, cfg *config.Config, lg *slog.Logger) (*App, error) {
	pool, err := postgres.New(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	repo := repository.NewSubscriptionRepository(pool)
	service := service.NewSubscriptionService(repo, lg)

	v := validator.New(validator.WithRequiredStructEnabled())

	router := transport.NewRouter(service, lg, v)

	addr := cfg.HTTP.Host + ":" + cfg.HTTP.Port
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return &App{
		db:     pool,
		router: router,
		server: server,
		lg:     lg,
	}, nil
}

func (a *App) Run(ctx context.Context) {
	a.lg.Info("Starting HTTP server", slog.String("addr", a.server.Addr))
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.lg.Error("HTTP server error", slog.Any("error", err))
		}
	}()

	<-ctx.Done()
	a.lg.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.lg.Error("Failed to shutdown HTTP server", slog.Any("error", err))
	}

	a.db.Close()
}
