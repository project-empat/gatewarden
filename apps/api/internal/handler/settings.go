package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gatewarden/api/internal/middleware"
	"github.com/gatewarden/api/internal/model"
	"github.com/gatewarden/api/internal/service"
)

type SettingsHandler struct {
	svc   *service.SettingsService
	audit *service.AuditService
}

func NewSettingsHandler(svc *service.SettingsService, audit *service.AuditService) *SettingsHandler {
	return &SettingsHandler{svc: svc, audit: audit}
}

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := h.svc.Get(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.Settings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	settings, err := h.svc.Update(r.Context(), req)
	if err != nil {
		_ = h.audit.Record(r.Context(), auditEvent(r, middleware.UserID(r.Context()), "settings.update", "settings", "", "error", "update failed"))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	_ = h.audit.Record(r.Context(), auditEvent(r, middleware.UserID(r.Context()), "settings.update", "settings", "", "success", ""))
	writeJSON(w, http.StatusOK, settings)
}
