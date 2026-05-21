package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gatewarden/api/internal/model"
	"github.com/gatewarden/api/internal/service"
)

type NodeHandler struct {
	svc *service.NodeService
}

func NewNodeHandler(svc *service.NodeService) *NodeHandler {
	return &NodeHandler{svc: svc}
}

func (h *NodeHandler) List(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.svc.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if nodes == nil {
		nodes = []model.Node{}
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (h *NodeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	node, err := h.svc.Get(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, node)
}

func (h *NodeHandler) DashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.DashboardStats(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
