package reporter

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gatewarden/agent/internal/cloudflare"
	"github.com/gatewarden/agent/internal/crowdsec"
	"github.com/gatewarden/agent/internal/docker"
	"github.com/gatewarden/agent/internal/fail2ban"
	"github.com/gatewarden/agent/internal/firewall"
	"github.com/gatewarden/agent/internal/journald"
	"github.com/gatewarden/agent/internal/ports"
	"github.com/gatewarden/agent/internal/ssh"
	"github.com/gatewarden/agent/internal/system"
	"github.com/gatewarden/agent/internal/tailscale"
	"github.com/gatewarden/agent/pkg/proto"
)

// Collect gathers all check results and builds a full AgentReport.
func Collect(hostname string) *proto.AgentReport {
	now := time.Now().UTC()

	report := &proto.AgentReport{
		Timestamp:     now.Format(time.RFC3339),
		Hostname:      hostname,
		OS:            readOSInfo(),
		Kernel:        readKernel(),
		UptimeSeconds: readUptime(),
	}

	// Infrastructure checks
	report.Ports = collectPorts()
	report.Docker = collectDocker()
	report.Firewall = collectFirewall()
	report.SSH = collectSSH()
	report.System = collectSystem()
	report.AuthLog = collectAuthLog()

	// Integration checks
	report.CrowdSec = collectCrowdSec()
	report.Fail2Ban = collectFail2Ban()
	report.Tailscale = collectTailscale()
	report.CloudflareTunnel = collectCloudflare()

	return report
}

func collectPorts() *proto.PortsStatus {
	listening, err := ports.Scan()
	if err != nil {
		log.Printf("ports scan: %v", err)
		return nil
	}
	if len(listening) == 0 {
		return nil
	}
	return &proto.PortsStatus{Listening: listening}
}

func collectDocker() *proto.DockerStatus {
	d, err := docker.Check()
	if err != nil {
		log.Printf("docker: %v", err)
		return nil
	}
	return d
}

func collectFirewall() *proto.FirewallStatus {
	return firewall.Check()
}

func collectSSH() *proto.SSHStatus {
	s, err := ssh.CheckHardening()
	if err != nil {
		log.Printf("ssh: %v", err)
		return nil
	}
	return s
}

func collectSystem() *proto.SystemHealth {
	s, err := system.Gather()
	if err != nil {
		log.Printf("system health: %v", err)
		return nil
	}
	return s
}

func collectAuthLog() *proto.AuthLogStatus {
	a, err := journald.Check()
	if err != nil {
		log.Printf("authlog: %v", err)
		return nil
	}
	return a
}

func collectCrowdSec() *proto.CrowdSecStatus {
	c, err := crowdsec.Check()
	if err != nil {
		log.Printf("crowdsec: %v", err)
		return nil
	}
	return c
}

func collectFail2Ban() *proto.Fail2BanStatus {
	f, err := fail2ban.Check()
	if err != nil {
		log.Printf("fail2ban: %v", err)
		return nil
	}
	return f
}

func collectTailscale() *proto.TailscaleStatus {
	t, err := tailscale.Check()
	if err != nil {
		log.Printf("tailscale: %v", err)
		return nil
	}
	return t
}

func collectCloudflare() *proto.CloudflareStatus {
	c, err := cloudflare.Check()
	if err != nil {
		log.Printf("cloudflare: %v", err)
		return nil
	}
	return c
}

func getOSRelease() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "linux"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
		}
	}
	return "linux"
}

func readOSInfo() string {
	return getOSRelease()
}

func readKernel() string {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readUptime() int64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0
	}
	secs, _ := strconv.ParseFloat(fields[0], 64)
	return int64(secs)
}
