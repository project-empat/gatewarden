package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gatewarden/api/internal/service"
)

type VulnHandler struct {
	svc *service.VulnerabilityService
}

func NewVulnHandler(svc *service.VulnerabilityService) *VulnHandler {
	return &VulnHandler{svc: svc}
}

// List returns all packages with known CVEs across the fleet.
func (h *VulnHandler) List(w http.ResponseWriter, r *http.Request) {
	vulns, err := h.svc.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, vulns)
}

// NodeList returns vulnerable packages for a specific node.
func (h *VulnHandler) NodeList(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")
	vulns, err := h.svc.NodeList(r.Context(), nodeID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, vulns)
}

// FIM returns monitored files that have changed across the fleet.
func (h *VulnHandler) FIM(w http.ResponseWriter, r *http.Request) {
	changes, err := h.svc.FIMChanges(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, changes)
}

// NodeFIM returns the monitored files (baseline + changes) for a node.
func (h *VulnHandler) NodeFIM(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")
	files, err := h.svc.NodeFIM(r.Context(), nodeID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, files)
}
