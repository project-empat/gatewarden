package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gatewarden/api/internal/middleware"
	"github.com/gatewarden/api/internal/model"
	"github.com/gatewarden/api/internal/service"
)

// AgentHandler handles agent registration, report ingestion, and heartbeats.
type AgentHandler struct {
	svc *service.AgentService
}

func NewAgentHandler(svc *service.AgentService) *AgentHandler {
	return &AgentHandler{svc: svc}
}

// Register handles POST /api/agent/register
func (h *AgentHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.Hostname == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "hostname is required"})
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Report handles POST /api/agent/report
func (h *AgentHandler) Report(w http.ResponseWriter, r *http.Request) {
	nodeID := r.Context().Value(middleware.NodeIDKey).(string)

	var report json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid report"})
		return
	}

	if err := h.svc.ProcessReport(r.Context(), nodeID, report); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Heartbeat handles POST /api/agent/heartbeat
func (h *AgentHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	nodeID := r.Context().Value(middleware.NodeIDKey).(string)

	if err := h.svc.Heartbeat(r.Context(), nodeID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
