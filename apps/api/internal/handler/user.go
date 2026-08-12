package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gatewarden/api/internal/middleware"
	"github.com/gatewarden/api/internal/service"
)

type UserHandler struct {
	svc   *service.UserService
	audit *service.AuditService
}

func NewUserHandler(svc *service.UserService, audit *service.AuditService) *UserHandler {
	return &UserHandler{svc: svc, audit: audit}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) SetRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if err := h.svc.SetRole(r.Context(), userID, req.Role); err != nil {
		if err == service.ErrInvalidRole || err == service.ErrUserNotFound {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	_ = h.audit.Record(r.Context(), auditEvent(r,
		middleware.UserID(r.Context()), "users.role.update", "users", userID, "success", ""))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
