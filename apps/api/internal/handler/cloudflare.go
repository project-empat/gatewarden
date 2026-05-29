package handler

import (
	"net/http"
	"sync"

	"github.com/gatewarden/api/internal/integration"
	"github.com/gatewarden/api/internal/service"
)

type CloudflareHandler struct {
	svc  *service.SettingsService
	mu   sync.Mutex
}

func NewCloudflareHandler(svc *service.SettingsService) *CloudflareHandler {
	return &CloudflareHandler{svc: svc}
}

// client creates a CFClient from stored settings.
func (h *CloudflareHandler) client(r *http.Request) (*integration.CFClient, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	settings, err := h.svc.Get(r.Context())
	if err != nil {
		return nil, err
	}
	if settings.CloudflareAPIToken == "" {
		return nil, errCloudflareNotConfigured
	}
	return integration.NewCFClient(settings.CloudflareAPIToken), nil
}

var errCloudflareNotConfigured = &apiError{msg: "Cloudflare API token not configured", code: 400}

type apiError struct {
	msg  string
	code int
}

func (e *apiError) Error() string { return e.msg }

// Accounts returns the Cloudflare accounts accessible with the configured token.
func (h *CloudflareHandler) Accounts(w http.ResponseWriter, r *http.Request) {
	client, err := h.client(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	accounts, err := client.GetAccounts()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, accounts)
}

// Tunnels returns all CF Tunnels for an account.
func (h *CloudflareHandler) Tunnels(w http.ResponseWriter, r *http.Request) {
	client, err := h.client(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	accountID := r.URL.Query().Get("account_id")
	if accountID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account_id query parameter required"})
		return
	}

	tunnels, err := client.ListTunnels(accountID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, tunnels)
}

// TunnelHealth returns the health status of a specific tunnel.
func (h *CloudflareHandler) TunnelHealth(w http.ResponseWriter, r *http.Request) {
	client, err := h.client(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	accountID := r.URL.Query().Get("account_id")
	tunnelID := r.URL.Query().Get("tunnel_id")
	if accountID == "" || tunnelID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account_id and tunnel_id required"})
		return
	}

	tunnel, err := client.GetTunnel(accountID, tunnelID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, tunnel)
}
