package handler

import (
	"net/http"
	"sync"

	"github.com/gatewarden/api/internal/integration"
	"github.com/gatewarden/api/internal/service"
)

type TailscaleHandler struct {
	svc  *service.SettingsService
	mu   sync.Mutex
}

func NewTailscaleHandler(svc *service.SettingsService) *TailscaleHandler {
	return &TailscaleHandler{svc: svc}
}

func (h *TailscaleHandler) client(r *http.Request) (*integration.TSClient, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	settings, err := h.svc.Get(r.Context())
	if err != nil {
		return nil, err
	}
	if settings.TailscaleAPIKey == "" {
		return nil, errTailscaleNotConfigured
	}
	tailnet := settings.TailscaleTailnet
	if tailnet == "" {
		return nil, &apiError{msg: "Tailscale tailnet name not configured", code: 400}
	}
	return integration.NewTSClient(settings.TailscaleAPIKey, tailnet), nil
}

var errTailscaleNotConfigured = &apiError{msg: "Tailscale API key not configured", code: 400}

// Devices returns all devices in the tailnet.
func (h *TailscaleHandler) Devices(w http.ResponseWriter, r *http.Request) {
	client, err := h.client(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	devices, err := client.ListDevices()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	if devices == nil {
		devices = []integration.TSDevice{}
	}

	writeJSON(w, http.StatusOK, devices)
}

// ACL returns the current ACL configuration and ETag.
func (h *TailscaleHandler) ACL(w http.ResponseWriter, r *http.Request) {
	client, err := h.client(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	acl, err := client.GetACL()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, acl)
}
