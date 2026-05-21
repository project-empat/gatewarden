package handler

import (
	"net/http"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	settings := map[string]interface{}{
		"agent_auto_approve": true,
		"heartbeat_interval": 60,
		"log_retention_days": 30,
	}
	writeJSON(w, http.StatusOK, settings)
}
