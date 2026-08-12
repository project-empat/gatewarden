package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gatewarden/api/internal/model"
	"github.com/gatewarden/api/internal/service"
)

type AuthHandler struct {
	svc   *service.AuthService
	audit *service.AuditService
}

func NewAuthHandler(svc *service.AuthService, audit *service.AuditService) *AuthHandler {
	return &AuthHandler{svc: svc, audit: audit}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		_ = h.audit.Record(r.Context(), auditEvent(r, "", "auth.login", "users", "", "error", "invalid credentials"))
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	_ = h.audit.Record(r.Context(), auditEvent(r, resp.User.ID, "auth.login", "users", resp.User.ID, "success", ""))
	writeJSON(w, http.StatusOK, resp)
}
