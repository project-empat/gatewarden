package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gatewarden/api/internal/service"
)

// AgentAuth validates the agent API key from the Authorization header.
// It injects the node_id into the request context.
func AgentAuth(svc *service.AgentService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			key := strings.TrimPrefix(header, "Bearer ")
			if key == header {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			nodeID, err := svc.ValidateAPIKey(r.Context(), key)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), NodeIDKey, nodeID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
