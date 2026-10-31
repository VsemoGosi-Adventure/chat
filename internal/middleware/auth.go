package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

// UserContextMiddleware extracts X-User-ID header (injected by Traefik after authentication)
// and adds it to the request context. Returns 401 if the header is missing or invalid.
func UserContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDHeader := r.Header.Get("X-User-ID")
		if userIDHeader == "" {
			http.Error(w, "Unauthorized: missing X-User-ID header", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(userIDHeader)
		if err != nil {
			http.Error(w, "Unauthorized: invalid X-User-ID format", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}
