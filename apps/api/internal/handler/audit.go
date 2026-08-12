package handler

import (
	"net/http"
	"strconv"

	"github.com/gatewarden/api/internal/model"
	"github.com/gatewarden/api/internal/service"
)

// auditEvent builds an audit event from a request. UserID empty means
// system-originated (e.g. login before the user is known).
func auditEvent(r *http.Request, userID, action, resource, resourceID, outcome, errMsg string) model.AuditEvent {
	return model.AuditEvent{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Severity:   "info",
		SourceIP:   clientIP(r),
		Outcome:    outcome,
		Error:      errMsg,
	}
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	return r.RemoteAddr
}

type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	q := model.AuditQuery{
		Action:   r.URL.Query().Get("action"),
		UserID:   r.URL.Query().Get("user_id"),
		Resource: r.URL.Query().Get("resource"),
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Offset = n
		}
	}

	events, err := h.svc.List(r.Context(), q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, events)
}
