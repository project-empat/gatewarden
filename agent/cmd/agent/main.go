package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gatewarden/agent/internal/cloudflare"
	"github.com/gatewarden/agent/internal/crowdsec"
	"github.com/gatewarden/agent/internal/docker"
	"github.com/gatewarden/agent/internal/fail2ban"
	"github.com/gatewarden/agent/internal/firewall"
	"github.com/gatewarden/agent/internal/journald"
	"github.com/gatewarden/agent/internal/reporter"
	"github.com/gatewarden/agent/internal/ssh"
	"github.com/gatewarden/agent/internal/tailscale"
)

type Check struct {
	Name   string
	Runner func() (map[string]interface{}, error)
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Gatewarden API URL")
	apiKey := flag.String("api-key", "", "API key for authentication")
	interval := flag.Duration("interval", 60*time.Second, "Report interval")
	flag.Parse()

	if *apiKey == "" {
		*apiKey = os.Getenv("GATEWARDEN_API_KEY")
	}

	hostname, _ := os.Hostname()
	rep := reporter.New(*serverURL, *apiKey, hostname)

	checks := []Check{
		{"docker", docker.Discover},
		{"ufw", firewall.CheckUFW},
		{"nftables", firewall.CheckNFTables},
		{"journald", journald.Check},
		{"fail2ban", fail2ban.Check},
		{"crowdsec", crowdsec.Check},
		{"ssh_hardening", ssh.CheckHardening},
		{"tailscale", tailscale.Check},
		{"cloudflare_tunnel", cloudflare.Check},
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("gatewarden agent starting", "server", *serverURL, "hostname", hostname)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	runChecks := func() {
		for _, c := range checks {
			result, err := c.Runner()
			if err != nil {
				log.Printf("check %s failed: %v", c.Name, err)
				result = map[string]interface{}{
					"name":    c.Name,
					"status":  "error",
					"message": err.Error(),
				}
			}
			data, _ := json.Marshal(result)
			_ = rep.Report("check_result", string(data))
		}
	}

	runChecks()

	for {
		select {
		case <-ticker.C:
			runChecks()
		case <-sigCh:
			log.Println("agent shutting down")
			rep.Close()
			return
		}
	}
}
