package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatewarden/api/internal/model"
)

type IncidentHandler struct {
	db *pgxpool.Pool
}

func NewIncidentHandler(db *pgxpool.Pool) *IncidentHandler {
	return &IncidentHandler{db: db}
}

func (h *IncidentHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		`SELECT id, node_id, severity, title, message, status, created_at, resolved_at
		 FROM incidents ORDER BY created_at DESC LIMIT 100`,
	)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var incidents []model.Incident
	for rows.Next() {
		var inc model.Incident
		if err := rows.Scan(&inc.ID, &inc.NodeID, &inc.Severity, &inc.Title, &inc.Message, &inc.Status, &inc.CreatedAt, &inc.ResolvedAt); err != nil {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
		incidents = append(incidents, inc)
	}
	if incidents == nil {
		incidents = []model.Incident{}
	}
	writeJSON(w, http.StatusOK, incidents)
}

func (h *IncidentHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	now := time.Now()
	_, err := h.db.Exec(r.Context(),
		`UPDATE incidents SET status = 'resolved', resolved_at = $1 WHERE id = $2 AND status = 'open'`,
		now, id,
	)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}
