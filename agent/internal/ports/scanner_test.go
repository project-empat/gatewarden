package ports

import (
	"testing"
)

func TestParseSS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantPort int
		wantProc string
		wantExp  bool
	}{
		{
			name: "single SSH listener on all interfaces",
			input: `State    Recv-Q   Send-Q     Local Address:Port     Peer Address:Port    Process
LISTEN   0        128               0.0.0.0:22          0.0.0.0:*        users:(("sshd",pid=1234,fd=3))
`,
			wantLen:  1,
			wantPort: 22,
			wantProc: "sshd",
			wantExp:  true,
		},
		{
			name: "localhost only service",
			input: `State    Recv-Q   Send-Q     Local Address:Port     Peer Address:Port    Process
LISTEN   0        128             127.0.0.1:3000        0.0.0.0:*        users:(("node",pid=5678,fd=10))
`,
			wantLen:  1,
			wantPort: 3000,
			wantProc: "node",
			wantExp:  false,
		},
		{
			name: "multiple listeners",
			input: `State    Recv-Q   Send-Q     Local Address:Port     Peer Address:Port    Process
LISTEN   0        128               0.0.0.0:22          0.0.0.0:*        users:(("sshd",pid=1234,fd=3))
LISTEN   0        4096              0.0.0.0:8080        0.0.0.0:*        users:(("nginx",pid=5678,fd=6))
LISTEN   0        128             127.0.0.1:5432        0.0.0.0:*        users:(("postgres",pid=9012,fd=7))
`,
			wantLen:  3,
			wantPort: 8080,
			wantProc: "nginx",
			wantExp:  true,
		},
		{
			name: "empty output",
			input: `State    Recv-Q   Send-Q     Local Address:Port     Peer Address:Port    Process
`,
			wantLen: 0,
		},
		{
			name: "header only",
			input: `State    Recv-Q   Send-Q     Local Address:Port     Peer Address:Port    Process
LISTEN   0        128               0.0.0.0:443         0.0.0.0:*        users:(("apache2",pid=3456,fd=8))
`,
			wantLen:  1,
			wantPort: 443,
			wantProc: "apache2",
			wantExp:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSS([]byte(tt.input))
			if len(got) != tt.wantLen {
				t.Errorf("parseSS() returned %d ports, want %d", len(got), tt.wantLen)
				return
			}
			if tt.wantLen == 0 {
				return
			}
			// Find the expected port
			found := false
			for _, p := range got {
				if p.Port == tt.wantPort {
					found = true
					if p.Process != tt.wantProc {
						t.Errorf("port %d process = %q, want %q", p.Port, p.Process, tt.wantProc)
					}
					if p.Exposed != tt.wantExp {
						t.Errorf("port %d exposed = %v, want %v", p.Port, p.Exposed, tt.wantExp)
					}
					if p.Protocol != "tcp" {
						t.Errorf("port %d protocol = %q, want \"tcp\"", p.Port, p.Protocol)
					}
				}
			}
			if !found {
				t.Errorf("expected port %d not found in results", tt.wantPort)
			}
		})
	}
}

func TestExtractProcess(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		want   string
	}{
		{
			name:   "standard sshd process",
			fields: []string{"LISTEN", "0", "128", "0.0.0.0:22", "0.0.0.0:*", `users:(("sshd",pid=1234,fd=3))`},
			want:   "sshd",
		},
		{
			name:   "nginx process",
			fields: []string{"LISTEN", "0", "4096", "0.0.0.0:80", "0.0.0.0:*", `users:(("nginx",pid=5678,fd=6))`},
			want:   "nginx",
		},
		{
			name:   "no process info",
			fields: []string{"LISTEN", "0", "128", "0.0.0.0:22", "0.0.0.0:*"},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractProcess(tt.fields)
			if got != tt.want {
				t.Errorf("extractProcess() = %q, want %q", got, tt.want)
			}
		})
	}
}
