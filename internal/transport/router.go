package transport

import (
	"log/slog"
	"net/http"

	_ "subservice/docs"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"subservice/internal/transport/middleware"
	"subservice/internal/transport/subscription"
)

func NewRouter(subscriptionService subscription.SubscriptionService, lg *slog.Logger, validator *validator.Validate) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.LoggingMiddleware(lg))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api", func(r chi.Router) {
		subscriptionHandler := subscription.NewSubscriptionHandler(subscriptionService, lg, validator)
		subscriptionHandler.Register(r)
	})

	return r
}
