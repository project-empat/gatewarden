package packages

import (
	"bufio"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

// parseDpkgOutput parses the output of `dpkg-query -W -f='${Package}\t${Version}\n'`.
func parseDpkgOutput(r io.Reader) []proto.PackageInfo {
	var pkgs []proto.PackageInfo
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		name := strings.TrimSpace(parts[0])
		if name == "" {
			continue
		}
		version := ""
		if len(parts) == 2 {
			version = strings.TrimSpace(parts[1])
		}
		pkgs = append(pkgs, proto.PackageInfo{Name: name, Version: version})
	}
	return pkgs
}

// parseAptCheckHuman parses the update-notifier apt-check human-readable
// output: "N updates can be applied immediately. M of these updates are
// security updates." Returns (pending, security).
func parseAptCheckHuman(out string) (int, int) {
	pending := -1
	security := -1
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		m1 := firstInt(line, "updates can be applied immediately")
		if m1 >= 0 {
			pending = m1
		}
		m2 := firstInt(line, "of these updates are security updates")
		if m2 >= 0 {
			security = m2
		}
	}
	if pending < 0 {
		pending = 0
	}
	if security < 0 {
		security = 0
	}
	return pending, security
}

func firstInt(line, marker string) int {
	idx := strings.Index(line, marker)
	if idx < 0 {
		return -1
	}
	// look for the number immediately before the marker
	tail := strings.TrimSpace(line[:idx])
	fields := strings.Fields(tail)
	if len(fields) == 0 {
		return -1
	}
	n, err := strconv.Atoi(fields[len(fields)-1])
	if err != nil {
		return -1
	}
	return n
}

// securityUpdatesPending returns how many security updates are available via
// update-notifier's apt-check. Returns 0 if unavailable (non-Debian, no
// update-notifier, or permission issues) — treated as "unknown, degrade low".
func securityUpdatesPending() int {
	cmd := exec.Command("/usr/lib/update-notifier/apt-check", "--human-readable")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0
	}
	pending, security := parseAptCheckHuman(string(out))
	_ = pending
	return security
}

// Check returns installed packages and the count of pending security updates.
func Check() (*proto.PackageStatus, error) {
	cmd := exec.Command("dpkg-query", "-W", "-f=${Package}\t${Version}\n")
	raw, err := cmd.Output()
	if err != nil {
		return nil, err // not a dpkg-based system
	}

	installed := parseDpkgOutput(strings.NewReader(string(raw)))
	if len(installed) == 0 {
		return nil, nil
	}

	return &proto.PackageStatus{
		Installed:              installed,
		SecurityUpdatesPending: securityUpdatesPending(),
	}, nil
}
