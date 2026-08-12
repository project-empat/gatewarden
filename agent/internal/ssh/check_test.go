package ssh

import (
	"strings"
	"testing"

	"github.com/gatewarden/agent/pkg/proto"
)

func TestParseSSHDConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		wantPort  int
		wantPW    bool
		wantRoot  string
		wantTries int
	}{
		{
			name: "secure defaults",
			config: `Port 2222
PasswordAuthentication no
PermitRootLogin prohibit-password
PubkeyAuthentication yes
MaxAuthTries 3
`,
			wantPort:  2222,
			wantPW:    false,
			wantRoot:  "prohibit-password",
			wantTries: 3,
		},
		{
			name: "insecure config",
			config: `Port 22
PasswordAuthentication yes
PermitRootLogin yes
PubkeyAuthentication yes
MaxAuthTries 6
`,
			wantPort:  22,
			wantPW:    true,
			wantRoot:  "yes",
			wantTries: 6,
		},
		{
			name: "with comments and empty lines",
			config: `# This is a comment

Port 2222

# Another comment
PasswordAuthentication no
PermitRootLogin no
`,
			wantPort:  2222,
			wantPW:    false,
			wantRoot:  "no",
			wantTries: 6,
		},
		{
			name: "partial config with defaults",
			config: `Port 2222
PasswordAuthentication yes
`,
			wantPort:  2222,
			wantPW:    true,
			wantRoot:  "no",
			wantTries: 6,
		},
		{
			name: "case insensitive keys",
			config: `PORT 2222
PASSWORDAUTHENTICATION yes
PermitRootLogin prohibit-password
`,
			wantPort:  2222,
			wantPW:    true,
			wantRoot:  "prohibit-password",
			wantTries: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &proto.SSHStatus{
				Port:         22,
				PubkeyAuth:   true,
				PasswordAuth: false,
				RootLogin:    "no",
				MaxAuthTries: 6,
			}

			got, err := parseSSHDConfig(s, strings.NewReader(tt.config))
			if err != nil {
				t.Fatalf("parseSSHDConfig() error: %v", err)
			}

			if got.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", got.Port, tt.wantPort)
			}
			if got.PasswordAuth != tt.wantPW {
				t.Errorf("PasswordAuth = %v, want %v", got.PasswordAuth, tt.wantPW)
			}
			if got.RootLogin != tt.wantRoot {
				t.Errorf("RootLogin = %q, want %q", got.RootLogin, tt.wantRoot)
			}
			if got.MaxAuthTries != tt.wantTries {
				t.Errorf("MaxAuthTries = %d, want %d", got.MaxAuthTries, tt.wantTries)
			}
		})
	}
}
