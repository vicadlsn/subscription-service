package subscription

import (
	"subservice/internal/transport/middleware"

	"github.com/go-chi/chi/v5"
)

func (h *SubscriptionHandler) Register(r chi.Router) {
	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/total", middleware.ErrorLoggingMiddleware(h.GetTotalCost))
		r.Post("/", middleware.ErrorLoggingMiddleware(h.Create))
		r.Get("/{id}", middleware.ErrorLoggingMiddleware(h.GetByID))
		r.Get("/", middleware.ErrorLoggingMiddleware(h.List))
		r.Put("/{id}", middleware.ErrorLoggingMiddleware(h.Update))
		r.Delete("/{id}", middleware.ErrorLoggingMiddleware(h.Delete))
	})

}
