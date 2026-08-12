package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gatewarden/api/internal/middleware"
	"github.com/gatewarden/api/internal/model"
	"github.com/gatewarden/api/internal/service"
)

type PolicyHandler struct {
	svc   *service.PolicyService
	audit *service.AuditService
}

func NewPolicyHandler(svc *service.PolicyService, audit *service.AuditService) *PolicyHandler {
	return &PolicyHandler{svc: svc, audit: audit}
}

func (h *PolicyHandler) List(w http.ResponseWriter, r *http.Request) {
	policies, err := h.svc.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, policies)
}

func (h *PolicyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	policy, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (h *PolicyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p model.Policy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if p.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	created, err := h.svc.Create(r.Context(), p)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	_ = h.audit.Record(r.Context(), auditEvent(r, middleware.UserID(r.Context()), "policies.create", "policies", created.ID, "success", ""))
	writeJSON(w, http.StatusCreated, created)
}

func (h *PolicyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var p model.Policy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	updated, err := h.svc.Update(r.Context(), id, p)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	_ = h.audit.Record(r.Context(), auditEvent(r, middleware.UserID(r.Context()), "policies.update", "policies", id, "success", ""))
	writeJSON(w, http.StatusOK, updated)
}

func (h *PolicyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	_ = h.audit.Record(r.Context(), auditEvent(r, middleware.UserID(r.Context()), "policies.delete", "policies", id, "success", ""))
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *PolicyHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.svc.Toggle(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	_ = h.audit.Record(r.Context(), auditEvent(r, middleware.UserID(r.Context()), "policies.toggle", "policies", id, "success", ""))
	writeJSON(w, http.StatusOK, p)
}
