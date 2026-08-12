package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gatewarden/api/internal/service"
)

// GraphHandler serves the infrastructure security graph.
type GraphHandler struct {
	svc *service.GraphService
}

func NewGraphHandler(svc *service.GraphService) *GraphHandler {
	return &GraphHandler{svc: svc}
}

// Full returns a page of the infrastructure security graph. Supports
// limit/offset query parameters for large infrastructures.
func (h *GraphHandler) Full(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r, 200, 0)
	graph, err := h.svc.GetFullGraph(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, `{"error":"failed to build graph"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

// Node returns the sub-graph for a specific node.
func (h *GraphHandler) Node(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")
	graph, err := h.svc.GetNodeGraph(r.Context(), nodeID)
	if err != nil {
		http.Error(w, `{"error":"failed to build node graph"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

// Stats returns aggregate graph statistics.
func (h *GraphHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetGraphStats(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to get graph stats"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
