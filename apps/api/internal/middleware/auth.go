package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gatewarden/api/internal/service"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const RoleKey contextKey = "role"
const NodeIDKey contextKey = "node_id"

// UserID returns the authenticated user ID from the request context.
func UserID(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

// Role returns the authenticated user's role ("admin" or "viewer").
func Role(ctx context.Context) string {
	v, _ := ctx.Value(RoleKey).(string)
	if v == "" {
		return "viewer"
	}
	return v
}

func Auth(svc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			if token == header {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			userID, role, err := svc.ValidateToken(token)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Admin restricts a route to users with the admin role.
func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Role(r.Context()) != "admin" {
			http.Error(w, `{"error":"forbidden: admin role required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
