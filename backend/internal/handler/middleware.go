package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gotogether/backend/internal/service"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header"})
				return
			}

			userID, err := authService.ValidateToken(parts[1])
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getUserID(r *http.Request) uuid.UUID {
	return r.Context().Value(userIDKey).(uuid.UUID)
}
