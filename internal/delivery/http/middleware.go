package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/MiRRoRise/auth-service/internal/domain"
)

type contextKey string

const UserIDKey contextKey = "user_id"

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

		userID, err := h.userUseCase.ValidateRefreshToken(r.Context(), tokenString)
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
