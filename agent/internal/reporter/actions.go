package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Action represents a pending action from the API.
type Action struct {
	ID         string          `json:"id"`
	NodeID     string          `json:"node_id"`
	ActionType string          `json:"action_type"`
	Params     json.RawMessage `json:"params"`
	Status     string          `json:"status"`
}

// ActionParams is the decoded params for an action.
type ActionParams struct {
	IP       string `json:"ip,omitempty"`
	Duration string `json:"duration,omitempty"`
	Jail     string `json:"jail,omitempty"`
	Port     int    `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Action   string `json:"action,omitempty"` // allow|deny for firewall
}

// ActionExecutor polls and executes remediation actions.
type ActionExecutor struct {
	serverURL string
	apiKey    string
	hostname  string
	nodeID    string
	client    *http.Client
}

// NewActionExecutor creates a new ActionExecutor.
func NewActionExecutor(serverURL, apiKey, hostname string) *ActionExecutor {
	return &ActionExecutor{
		serverURL: serverURL,
		apiKey:    apiKey,
		hostname:  hostname,
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

// PollAndExecute fetches pending actions and executes them.
func (e *ActionExecutor) PollAndExecute() error {
	if e.nodeID == "" {
		return nil // Not registered yet
	}

	actions, err := e.poll()
	if err != nil {
		return fmt.Errorf("poll actions: %w", err)
	}

	for _, a := range actions {
		log.Printf("executing action: %s (type=%s)", a.ID, a.ActionType)
		status := "completed"
		if err := e.execute(a); err != nil {
			log.Printf("action %s failed: %v", a.ID, err)
			status = "failed"
		}
		if err := e.complete(a.ID, status); err != nil {
			log.Printf("complete action %s: %v", a.ID, err)
		}
	}

	return nil
}

func (e *ActionExecutor) poll() ([]Action, error) {
	body, err := e.request("GET", "/api/agent/actions", nil)
	if err != nil {
		return nil, err
	}

	var actions []Action
	if err := json.Unmarshal(body, &actions); err != nil {
		return nil, fmt.Errorf("parse actions: %w", err)
	}
	return actions, nil
}

func (e *ActionExecutor) execute(a Action) error {
	var params ActionParams
	if a.Params != nil {
		json.Unmarshal(a.Params, &params)
	}

	switch a.ActionType {
	case "fail2ban_ban_ip":
		return execFail2BanBanIP(params)
	case "fail2ban_unban_ip":
		return execFail2BanUnbanIP(params)
	case "ufw_allow_port":
		return execUFWPort("allow", params)
	case "ufw_deny_port":
		return execUFWPort("deny", params)
	case "restart_service":
		return execRestartService(string(a.Params))
	default:
		return fmt.Errorf("unknown action type: %s", a.ActionType)
	}
}

func (e *ActionExecutor) complete(actionID, status string) error {
	payload := map[string]string{"status": status}
	body, _ := json.Marshal(payload)
	_, err := e.request("POST", fmt.Sprintf("/api/agent/actions/%s/complete", actionID), body)
	return err
}

func (e *ActionExecutor) request(method, path string, reqBody []byte) ([]byte, error) {
	req, err := http.NewRequest(method, e.serverURL+path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// On 401, re-register and retry
	if resp.StatusCode == 401 {
		log.Printf("action poll auth rejected, re-registering")
		return nil, fmt.Errorf("auth rejected")
	}

	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s returned %d: %s", path, resp.StatusCode, buf.String())
	}

	return buf.Bytes(), nil
}

// Close cleans up.
func (e *ActionExecutor) Close() {
	e.client.CloseIdleConnections()
}

// -- Action executors --

func execFail2BanBanIP(p ActionParams) error {
	if p.IP == "" {
		return fmt.Errorf("ip required")
	}
	jail := p.Jail
	if jail == "" {
		jail = "sshd"
	}
	cmd := exec.Command("fail2ban-client", "set", jail, "banip", p.IP)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fail2ban ban %s in %s: %w - %s", p.IP, jail, err, string(out))
	}
	log.Printf("banned %s in fail2ban jail %s", p.IP, jail)
	return nil
}

func execFail2BanUnbanIP(p ActionParams) error {
	if p.IP == "" {
		return fmt.Errorf("ip required")
	}
	jail := p.Jail
	if jail == "" {
		jail = "sshd"
	}
	cmd := exec.Command("fail2ban-client", "set", jail, "unbanip", p.IP)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fail2ban unban %s from %s: %w - %s", p.IP, jail, err, string(out))
	}
	log.Printf("unbanned %s from fail2ban jail %s", p.IP, jail)
	return nil
}

func execUFWPort(action string, p ActionParams) error {
	if p.Port == 0 {
		return fmt.Errorf("port required")
	}
	proto := p.Protocol
	if proto == "" {
		proto = "tcp"
	}
	args := []string{action, fmt.Sprintf("%d/%s", p.Port, proto)}
	cmd := exec.Command("ufw", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ufw %s port %d/%s: %w - %s", action, p.Port, proto, err, string(out))
	}
	log.Printf("ufw %s %d/%s", action, p.Port, proto)
	return nil
}

func execRestartService(params string) error {
	service := strings.TrimSpace(params)
	if service == "" {
		return fmt.Errorf("service name required")
	}
	cmd := exec.Command("systemctl", "restart", service)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restart %s: %w - %s", service, err, string(out))
	}
	log.Printf("restarted service %s", service)
	return nil
}
