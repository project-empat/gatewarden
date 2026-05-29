package handler

import (
	"net/http"

	"github.com/gatewarden/api/internal/service"
)

// ReportHandler serves aggregated security reports.
type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// Posture returns the comprehensive security posture report.
func (h *ReportHandler) Posture(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.SecurityPosture(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to generate posture report"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

// Incidents returns the aggregated incident summary.
func (h *ReportHandler) Incidents(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.IncidentSummaryReport(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to generate incident summary"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// Health returns per-node health data.
func (h *ReportHandler) Health(w http.ResponseWriter, r *http.Request) {
	health, err := h.svc.NodeHealthReport(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to generate health overview"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, health)
}
