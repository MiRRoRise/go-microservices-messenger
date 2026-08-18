package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/MiRRoRise/chat-service/internal/metrics"
	"github.com/go-chi/chi/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func (h *Handler) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)
		duration := time.Since(start).Seconds()

		path := r.URL.Path
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			if pattern := rctx.RoutePattern(); pattern != "" {
				path = pattern
			}
		}
		metrics.RecordRequest(r.Method, path, int64(ww.statusCode), duration)
	})
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid header", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		userID, err := h.tokenManager.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "error validate token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	if !ok {
		return 0, domain.ErrUnauthorized
	}
	return userID, nil
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
