package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gatewarden/api/internal/middleware"
	"github.com/gatewarden/api/internal/service"
)

type ActionHandler struct {
	svc *service.ActionService
}

func NewActionHandler(svc *service.ActionService) *ActionHandler {
	return &ActionHandler{svc: svc}
}

// CreateActionRequest for user-initiated actions.
type CreateActionRequest struct {
	NodeID     string          `json:"node_id"`
	ActionType string          `json:"action_type"`
	Params     json.RawMessage `json:"params"`
}

// Create handles POST /api/actions (user creates a remediation action).
func (h *ActionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.NodeID == "" || req.ActionType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "node_id and action_type required"})
		return
	}

	action, err := h.svc.Create(r.Context(), req.NodeID, req.ActionType, req.Params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, action)
}

// Poll handles GET /api/agent/actions (agent polls for pending actions).
func (h *ActionHandler) Poll(w http.ResponseWriter, r *http.Request) {
	nodeID := r.Context().Value(middleware.NodeIDKey).(string)

	actions, err := h.svc.PollPending(r.Context(), nodeID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, actions)
}

// Complete handles POST /api/agent/actions/{id}/complete (agent marks action done).
func (h *ActionHandler) Complete(w http.ResponseWriter, r *http.Request) {
	actionID := chi.URLParam(r, "id")

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Status = "completed"
	}

	if req.Status == "" {
		req.Status = "completed"
	}

	if err := h.svc.Complete(r.Context(), actionID, req.Status); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
