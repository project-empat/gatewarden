package journald

import (
	"testing"
)

func TestParseAuthLines(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantFailed int
		wantRoot   int
		wantIPs    int
		wantUsers  int
	}{
		{
			name: "single failed password attempt",
			input: `May 26 11:45:00 server sshd[1234]: Failed password for invalid user admin from 192.168.1.100 port 22 ssh2
`,
			wantFailed: 1,
			wantRoot:   0,
			wantIPs:    1,
			wantUsers:  1,
		},
		{
			name: "multiple failed attempts from same IP",
			input: `May 26 11:45:00 server sshd[1234]: Failed password for root from 10.0.0.5 port 22 ssh2
May 26 11:45:01 server sshd[1234]: Failed password for invalid user admin from 10.0.0.5 port 22 ssh2
May 26 11:45:02 server sshd[1234]: Failed password for invalid user test from 10.0.0.5 port 22 ssh2
`,
			wantFailed: 3,
			wantRoot:   1,
			wantIPs:    1,
			wantUsers:  3,
		},
		{
			name: "mix of failures and successes",
			input: `May 26 11:45:00 server sshd[1234]: Failed password for root from 10.0.0.5 port 22 ssh2
May 26 11:45:01 server sshd[1234]: Accepted publickey for deploy from 10.0.0.10 port 22 ssh2
May 26 11:45:02 server sshd[1234]: Invalid user admin from 10.0.0.5 port 22 ssh2
May 26 11:45:03 server sshd[1234]: Failed password for invalid user ubuntu from 10.0.0.5 port 22 ssh2
`,
			wantFailed: 3,
			wantRoot:   1,
			wantIPs:    1,
			wantUsers:  3,
		},
		{
			name: "sudo events",
			input: `May 26 11:45:00 server sudo: pam_unix(sudo:auth): authentication failure; logname=root uid=0 euid=0 tty=/dev/pts/0 ruser=root rhost=  user=root
May 26 11:45:01 server sshd[1234]: Failed password for root from 10.0.0.5 port 22 ssh2
`,
			wantFailed: 1,
			wantRoot:   1,
			wantIPs:    1,
			wantUsers:  1,
		},
		{
			name:       "empty input",
			input:      "",
			wantFailed: 0,
			wantRoot:   0,
			wantIPs:    0,
			wantUsers:  0,
		},
		{
			name: "no auth lines",
			input: `May 26 11:45:00 server kernel: something happened
May 26 11:45:01 server systemd[1]: Started something
`,
			wantFailed: 0,
			wantRoot:   0,
			wantIPs:    0,
			wantUsers:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAuthLines(tt.input)
			if err != nil {
				t.Fatalf("parseAuthLines() error: %v", err)
			}

			if got.FailedSSH != tt.wantFailed {
				t.Errorf("FailedSSH = %d, want %d", got.FailedSSH, tt.wantFailed)
			}
			if got.FailedRoot != tt.wantRoot {
				t.Errorf("FailedRoot = %d, want %d", got.FailedRoot, tt.wantRoot)
			}
			if len(got.SourceIPs) != tt.wantIPs {
				t.Errorf("SourceIPs count = %d, want %d", len(got.SourceIPs), tt.wantIPs)
			}
			if len(got.TargetedUsers) != tt.wantUsers {
				t.Errorf("TargetedUsers count = %d, want %d", len(got.TargetedUsers), tt.wantUsers)
			}
		})
	}
}

func TestBuildStatus(t *testing.T) {
	w := &TimeWindow{
		FailedSSH:  10,
		FailedRoot: 5,
		Successful: 3,
		SourceIPs: map[string]int{
			"10.0.0.5": 7,
			"10.0.0.6": 3,
		},
		TargetedUsers: map[string]int{
			"root":  5,
			"admin": 3,
			"test":  2,
		},
	}

	status := buildStatus(w, "journald")

	if status.FailedSSHLastHour != 10 {
		t.Errorf("FailedSSHLastHour = %d, want 10", status.FailedSSHLastHour)
	}
	if status.FailedRootLastHour != 5 {
		t.Errorf("FailedRootLastHour = %d, want 5", status.FailedRootLastHour)
	}
	if status.LogSource != "journald" {
		t.Errorf("LogSource = %q, want \"journald\"", status.LogSource)
	}
	if len(status.TopSourceIPs) != 2 {
		t.Errorf("TopSourceIPs count = %d, want 2", len(status.TopSourceIPs))
	}
	if len(status.TargetedUsernames) != 3 {
		t.Errorf("TargetedUsernames count = %d, want 3", len(status.TargetedUsernames))
	}

	// Top IP should be highest count
	if len(status.TopSourceIPs) > 0 && status.TopSourceIPs[0].IP != "10.0.0.5" {
		t.Errorf("Top IP = %q, want \"10.0.0.5\"", status.TopSourceIPs[0].IP)
	}

	// Top username should be highest count
	if len(status.TargetedUsernames) > 0 && status.TargetedUsernames[0].Username != "root" {
		t.Errorf("Top username = %q, want \"root\"", status.TargetedUsernames[0].Username)
	}
}
