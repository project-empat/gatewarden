package journald

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/gatewarden/agent/pkg/proto"
)

var (
	reFailedPassword = regexp.MustCompile(`Failed password for (?:invalid user )?(\S+) from (\S+)`)
	reInvalidUser    = regexp.MustCompile(`Invalid user (\S+) from (\S+)`)
	reAcceptedLogin  = regexp.MustCompile(`Accepted (?:password|publickey) for (\S+) from (\S+)`)
	reRootLogin      = regexp.MustCompile(`Failed password for (?:root|invalid user root)`)
	reSUDO           = regexp.MustCompile(`sudo:.*COMMAND=.*`)
	reSUDOFailure    = regexp.MustCompile(`sudo:.*authentication failure`)
)

// TimeWindow holds counted metrics for a time range.
type TimeWindow struct {
	FailedSSH      int
	FailedRoot     int
	Successful     int
	SUDOFailures   int
	SourceIPs      map[string]int
	TargetedUsers  map[string]int
}

// Check returns auth log status by parsing journald or auth.log.
func Check() (*proto.AuthLogStatus, error) {
	// Try journald first
	if hasJournalctl() {
		window, err := parseJournald()
		if err == nil {
			return buildStatus(window, "journald"), nil
		}
	}

	// Fallback to /var/log/auth.log
	if hasAuthLog() {
		window, err := parseAuthLog()
		if err == nil {
			return buildStatus(window, "auth.log"), nil
		}
	}

	return nil, nil // No auth log source available
}

func hasJournalctl() bool {
	return exec.Command("journalctl", "--version").Run() == nil
}

func hasAuthLog() bool {
	_, err := os.Stat("/var/log/auth.log")
	return err == nil
}

func buildStatus(w *TimeWindow, source string) *proto.AuthLogStatus {
	// Build top IPs list
	topIPs := make([]proto.IPCount, 0, len(w.SourceIPs))
	for ip, count := range w.SourceIPs {
		topIPs = append(topIPs, proto.IPCount{IP: ip, Attempts: count})
	}
	// Sort by attempts descending (simple bubble sort for small lists)
	for i := 0; i < len(topIPs); i++ {
		for j := i + 1; j < len(topIPs); j++ {
			if topIPs[j].Attempts > topIPs[i].Attempts {
				topIPs[i], topIPs[j] = topIPs[j], topIPs[i]
			}
		}
	}
	if len(topIPs) > 10 {
		topIPs = topIPs[:10]
	}

	// Build targeted usernames list
	users := make([]proto.UsernameCount, 0, len(w.TargetedUsers))
	for user, count := range w.TargetedUsers {
		users = append(users, proto.UsernameCount{Username: user, Attempts: count})
	}
	for i := 0; i < len(users); i++ {
		for j := i + 1; j < len(users); j++ {
			if users[j].Attempts > users[i].Attempts {
				users[i], users[j] = users[j], users[i]
			}
		}
	}
	if len(users) > 10 {
		users = users[:10]
	}

	if topIPs == nil {
		topIPs = []proto.IPCount{}
	}
	if users == nil {
		users = []proto.UsernameCount{}
	}

	return &proto.AuthLogStatus{
		FailedSSHLastHour:  w.FailedSSH,
		FailedRootLastHour: w.FailedRoot,
		TopSourceIPs:       topIPs,
		TargetedUsernames:  users,
		SUDOUsageLastHour:  w.SUDOFailures,
		LogSource:          source,
	}
}

func parseJournald() (*TimeWindow, error) {
	// Get SSH auth entries from last hour
	since := fmt.Sprintf("-%.0f minutes", 60.0)
	cmd := exec.Command("journalctl", "-u", "sshd", "--since", since, "--no-pager", "-o", "short-iso")
	out, err := cmd.Output()
	if err != nil {
		// Try broader query without unit filter
		cmd = exec.Command("journalctl", "--since", since, "--no-pager", "-o", "short-iso")
		out, err = cmd.Output()
		if err != nil {
			return nil, err
		}
	}

	return parseAuthLines(string(out))
}

func parseAuthLog() (*TimeWindow, error) {
	data, err := os.ReadFile("/var/log/auth.log")
	if err != nil {
		return nil, err
	}

	// Only consider entries from the last hour
	lines := strings.Split(string(data), "\n")
	now := time.Now()

	var recentLines []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		// auth.log format: "May 26 11:45:00 hostname sshd[1234]: ..."
		// Try to parse timestamp
		if isRecentLine(line, now) {
			recentLines = append(recentLines, line)
		}
	}

	return parseAuthLines(strings.Join(recentLines, "\n"))
}

func parseAuthLines(text string) (*TimeWindow, error) {
	w := &TimeWindow{
		SourceIPs:     make(map[string]int),
		TargetedUsers: make(map[string]int),
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)

		// Failed password attempts
		if matches := reFailedPassword.FindStringSubmatch(line); len(matches) >= 3 {
			username := matches[1]
			ip := matches[2]
			w.FailedSSH++
			w.SourceIPs[ip]++
			w.TargetedUsers[username]++
			if strings.Contains(lower, "for root") || username == "root" {
				w.FailedRoot++
			}
		}

		// Invalid user attempts
		if matches := reInvalidUser.FindStringSubmatch(line); len(matches) >= 3 {
			username := matches[1]
			ip := matches[2]
			w.FailedSSH++
			w.SourceIPs[ip]++
			w.TargetedUsers[username]++
		}

		// Successful logins
		if matches := reAcceptedLogin.FindStringSubmatch(line); len(matches) >= 3 {
			w.Successful++
		}

		// Sudo events
		if reSUDOFailure.MatchString(line) {
			w.SUDOFailures++
		}
	}

	return w, nil
}

func isRecentLine(line string, now time.Time) bool {
	// auth.log timestamps: "May 26 11:45:00" — no year
	// Parse with current year
	layout := "Jan 2 15:04:05"
	if len(line) < 15 {
		return false
	}
	t, err := time.Parse(layout, line[:15])
	if err != nil {
		return false
	}
	t = t.AddDate(now.Year(), 0, 0)
	return now.Sub(t) < time.Hour
}
