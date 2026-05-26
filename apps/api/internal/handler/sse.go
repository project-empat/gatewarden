package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gatewarden/api/internal/model"
	"github.com/gatewarden/api/internal/service"
)

type SSEHandler struct {
	svc *service.EventService
}

func NewSSEHandler(svc *service.EventService) *SSEHandler {
	return &SSEHandler{svc: svc}
}

func (h *SSEHandler) Stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Support token via query param for EventSource API (can't set custom headers)
	// The JWT auth middleware also validates the Authorization header
	if token := r.URL.Query().Get("token"); token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	id, ch := h.svc.Subscribe()
	defer h.svc.Unsubscribe(id)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			flusher.Flush()
		}
	}
}

func (h *SSEHandler) History(w http.ResponseWriter, r *http.Request) {
	events, err := h.svc.ListEvents(r.Context(), 50)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if events == nil {
		events = []model.Event{}
	}
	writeJSON(w, http.StatusOK, events)
}
