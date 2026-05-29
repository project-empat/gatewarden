package proto

// Heartbeat is sent periodically to indicate the agent is alive.
type Heartbeat struct {
	Hostname  string `json:"hostname"`
	Version   string `json:"version"`
	Uptime    int64  `json:"uptime"`
	Timestamp string `json:"timestamp"`
}

// AgentReport is the full periodic status report sent to the API.
type AgentReport struct {
	NodeID           string             `json:"node_id"`
	Timestamp        string             `json:"timestamp"`
	Hostname         string             `json:"hostname"`
	OS               string             `json:"os"`
	Kernel           string             `json:"kernel"`
	UptimeSeconds    int64              `json:"uptime_seconds"`
	Docker           *DockerStatus      `json:"docker,omitempty"`
	Firewall         *FirewallStatus    `json:"firewall,omitempty"`
	SSH              *SSHStatus         `json:"ssh,omitempty"`
	CrowdSec         *CrowdSecStatus    `json:"crowdsec,omitempty"`
	Fail2Ban         *Fail2BanStatus    `json:"fail2ban,omitempty"`
	Tailscale        *TailscaleStatus   `json:"tailscale,omitempty"`
	CloudflareTunnel *CloudflareStatus  `json:"cloudflare_tunnel,omitempty"`
	AuthLog          *AuthLogStatus     `json:"auth_log,omitempty"`
	Ports            *PortsStatus       `json:"ports,omitempty"`
	System           *SystemHealth      `json:"system,omitempty"`
}

type DockerStatus struct {
	RunningContainers []DockerContainer `json:"running_containers,omitempty"`
	TotalContainers   int               `json:"total_containers"`
	SocketExposed     bool              `json:"socket_exposed"`
}

type DockerContainer struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Image         string          `json:"image"`
	Ports         []DockerPort    `json:"ports,omitempty"`
	SocketExposed bool            `json:"socket_exposed"`
	Privileged    bool            `json:"privileged"`
	Status        string          `json:"status"`
	NetworkMode   string          `json:"network_mode"`
	User          string          `json:"user"`
}

type DockerPort struct {
	Internal int    `json:"internal"`
	External int    `json:"external"`
	Exposure string `json:"exposure"` // "0.0.0.0" | "127.0.0.1" | "none"
}

type FirewallStatus struct {
	ActiveBackend string       `json:"active_backend"` // "ufw" | "nftables" | "none"
	UFW           *UFWStatus   `json:"ufw,omitempty"`
	NFTables      *NFTStatus   `json:"nftables,omitempty"`
}

type UFWStatus struct {
	Active  bool         `json:"active"`
	Rules   []FirewallRule `json:"rules,omitempty"`
	Logging string       `json:"logging"`
}

type NFTStatus struct {
	Active bool   `json:"active"`
	Rules  string `json:"rules,omitempty"`
}

type FirewallRule struct {
	Action    string `json:"action"`    // "allow" | "deny" | "limit"
	Port      int    `json:"port"`
	Protocol  string `json:"proto"`
	Source    string `json:"from"`
	Interface string `json:"interface,omitempty"`
}

type SSHStatus struct {
	Port            int      `json:"port"`
	PasswordAuth    bool     `json:"password_auth"`
	RootLogin       string   `json:"root_login"`       // "yes" | "no" | "prohibit-password"
	PubkeyAuth      bool     `json:"pubkey_auth"`
	PubliclyExposed bool     `json:"publicly_exposed"`
	ListenAddresses []string `json:"listen_addresses,omitempty"`
	MaxAuthTries    int      `json:"max_auth_tries"`
}

type CrowdSecStatus struct {
	Installed          bool     `json:"installed"`
	Running            bool     `json:"running"`
	ActiveDecisions    int      `json:"active_decisions"`
	AlertsLastHour     int      `json:"alerts_last_hour"`
	Bouncers           []string `json:"bouncers,omitempty"`
}

type Fail2BanStatus struct {
	Installed bool          `json:"installed"`
	Running   bool          `json:"running"`
	Version   string        `json:"version"`
	Jails     []Fail2BanJail `json:"jails,omitempty"`
}

type Fail2BanJail struct {
	Name            string   `json:"name"`
	Active          bool     `json:"active"`
	CurrentlyBanned int      `json:"currently_banned"`
	TotalBanned     int      `json:"total_banned"`
	FailedCount     int      `json:"failed_count"`
	Bantime         string   `json:"bantime"`
	Maxretry        string   `json:"maxretry"`
	Findtime        string   `json:"findtime"`
}

type TailscaleStatus struct {
	Installed  bool       `json:"installed"`
	Running    bool       `json:"running"`
	NodeName   string     `json:"node_name"`
	NodeIP     string     `json:"node_ip"`
	Online     bool       `json:"online"`
	PeersCount int        `json:"peers_count"`
	Version    string     `json:"version"`
	ACLVersion string     `json:"acl_version,omitempty"`
}

type CloudflareStatus struct {
	Installed  bool              `json:"installed"`
	Running    bool              `json:"running"`
	Tunnels    []CloudflareTunnel `json:"tunnels,omitempty"`
	Version    string            `json:"version"`
}

type CloudflareTunnel struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Ingress []CloudflareIngress `json:"ingress,omitempty"`
}

type CloudflareIngress struct {
	Hostname string `json:"hostname"`
	Service  string `json:"service"`
}

type AuthLogStatus struct {
	FailedSSHLastHour  int              `json:"failed_ssh_last_hour"`
	FailedRootLastHour int              `json:"failed_root_last_hour"`
	TopSourceIPs       []IPCount        `json:"failed_ssh_top_ips,omitempty"`
	TargetedUsernames  []UsernameCount  `json:"targeted_usernames,omitempty"`
	SUDOUsageLastHour  int              `json:"sudo_usage_last_hour"`
	LogSource          string           `json:"source"` // "journald" | "auth.log" | "none"
}

type IPCount struct {
	IP       string `json:"ip"`
	Attempts int    `json:"attempts"`
}

type UsernameCount struct {
	Username string `json:"username"`
	Attempts int    `json:"attempts"`
}

type PortsStatus struct {
	Listening []ListeningPort `json:"listening,omitempty"`
}

type ListeningPort struct {
	Port     int    `json:"port"`
	Protocol string `json:"proto"`
	Process  string `json:"process"`
	Exposed  bool   `json:"exposed"`
}

type SystemHealth struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
}
