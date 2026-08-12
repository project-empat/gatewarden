package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gatewarden/api/internal/middleware"
	"github.com/gatewarden/api/internal/service"
)

type LicenseHandler struct {
	svc   *service.LicenseService
	audit *service.AuditService
}

func NewLicenseHandler(svc *service.LicenseService, audit *service.AuditService) *LicenseHandler {
	return &LicenseHandler{svc: svc, audit: audit}
}

// Get returns the current license (or the free default) plus build mode.
func (h *LicenseHandler) Get(w http.ResponseWriter, r *http.Request) {
	license, err := h.svc.Current(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"build_mode": h.svc.BuildMode(),
		"license":    license,
	})
}

// Activate validates and stores a license key.
func (h *LicenseHandler) Activate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LicenseKey string `json:"license_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	license, err := h.svc.Activate(r.Context(), req.LicenseKey)
	if err != nil {
		_ = h.audit.Record(r.Context(), auditEvent(r, middleware.UserID(r.Context()), "license.activate", "license", "", "error", err.Error()))
		if errors.Is(err, service.ErrEnterpriseOnly) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	_ = h.audit.Record(r.Context(), auditEvent(r, middleware.UserID(r.Context()), "license.activate", "license", license.ID, "success", ""))
	writeJSON(w, http.StatusOK, map[string]any{"license": license})
}

// Features returns the feature catalog with enabled state for the current
// build/license, so the UI can hide premium entries.
func (h *LicenseHandler) Features(w http.ResponseWriter, r *http.Request) {
	features, err := h.svc.Features(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"build_mode": h.svc.BuildMode(),
		"features":   features,
	})
}
