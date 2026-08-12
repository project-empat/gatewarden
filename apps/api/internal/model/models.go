package model

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Node struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Hostname  string   `json:"hostname"`
	IP        string   `json:"ip"`
	OS        string   `json:"os"`
	Status    string   `json:"status"`
	Labels    []string `json:"labels"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
}

type Incident struct {
	ID         string     `json:"id"`
	NodeID     string     `json:"node_id"`
	Severity   string     `json:"severity"`
	Title      string     `json:"title"`
	Message    string     `json:"message"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

type Agent struct {
	ID            string    `json:"id"`
	NodeID        string    `json:"node_id"`
	Version       string    `json:"version"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time `json:"created_at"`
}

type AgentReport struct {
	ID         int64     `json:"id"`
	AgentID    string    `json:"agent_id"`
	NodeID     string    `json:"node_id"`
	Report     string    `json:"report"`
	ReceivedAt time.Time `json:"received_at"`
}

type Event struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"node_id"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type Policy struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	Severity    string    `json:"severity"`
	Triggers    string    `json:"triggers"`
	Actions     string    `json:"actions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DashboardStats struct {
	TotalNodes     int `json:"total_nodes"`
	OnlineNodes    int `json:"online_nodes"`
	TotalIncidents int `json:"total_incidents"`
	OpenIncidents  int `json:"open_incidents"`
}

type Settings struct {
	AgentAutoApprove   bool      `json:"agent_auto_approve"`
	HeartbeatInterval  int       `json:"heartbeat_interval"`
	LogRetentionDays   int       `json:"log_retention_days"`
	CloudflareAPIToken string    `json:"cloudflare_api_token"`
	TailscaleAPIKey    string    `json:"tailscale_api_key"`
	TailscaleTailnet   string    `json:"tailscale_tailnet"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type RegisterRequest struct {
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
}

type RegisterResponse struct {
	NodeID string `json:"node_id"`
	APIKey string `json:"api_key"`
}

type AgentAction struct {
	ID          string     `json:"id"`
	NodeID      string     `json:"node_id"`
	ActionType  string     `json:"action_type"`
	Params      string     `json:"params"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// AuditEvent is a single entry in the audit trail.
type AuditEvent struct {
	ID         string         `json:"id"`
	UserID     string         `json:"user_id"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	ResourceID string         `json:"resource_id,omitempty"`
	Severity   string         `json:"severity"`
	SourceIP   string         `json:"source_ip,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Outcome    string         `json:"outcome"`
	Error      string         `json:"error,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

// AuditQuery filters the audit trail.
type AuditQuery struct {
	Action   string `json:"action"`
	UserID   string `json:"user_id"`
	Resource string `json:"resource"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}
