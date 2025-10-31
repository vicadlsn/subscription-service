package middleware

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"subservice/internal/transport/response"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func LoggingMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			defer func() {
				entry.Info("Request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

type appHandler func(w http.ResponseWriter, r *http.Request) error

func ErrorLoggingMiddleware(next appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)
		if err == nil {
			return
		}

		resp := response.MapError(err)

		if errors.Is(err, response.ErrValidationFailed) {
			resp.Details = response.ExtractValidationDetails(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.Status)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
