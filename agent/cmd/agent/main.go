package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gatewarden/agent/internal/reporter"
)

func main() {
	serverURL := flag.String("server", "", "Gatewarden API URL (required)")
	apiKey := flag.String("api-key", "", "API key for authentication")
	interval := flag.Duration("interval", 60*time.Second, "Report interval")
	heartbeatInterval := flag.Duration("heartbeat-interval", 30*time.Second, "Heartbeat interval")
	flag.Parse()

	// Env var fallbacks
	if *serverURL == "" {
		*serverURL = os.Getenv("GATEWARDEN_SERVER")
	}
	if *apiKey == "" {
		*apiKey = os.Getenv("GATEWARDEN_API_KEY")
	}

	if *serverURL == "" {
		fmt.Fprintf(os.Stderr, "error: --server or GATEWARDEN_SERVER is required\n")
		os.Exit(1)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	rep := reporter.New(*serverURL, *apiKey, hostname)

	// Register with the API to get a node ID
	if err := rep.Register(); err != nil {
		log.Printf("registration failed (will retry): %v", err)
	}

	// Action executor for remediation commands
	exec := reporter.NewActionExecutor(*serverURL, *apiKey, hostname)
	exec.SetNodeID(rep.GetNodeID())

	// Signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("gatewarden agent starting — server=%s hostname=%s interval=%s", *serverURL, hostname, *interval)

	reportTicker := time.NewTicker(*interval)
	heartbeatTicker := time.NewTicker(*heartbeatInterval)
	actionTicker := time.NewTicker(30 * time.Second)
	defer reportTicker.Stop()
	defer heartbeatTicker.Stop()
	defer actionTicker.Stop()

	// Run initial collection immediately
	runReport(rep, hostname)

	for {
		select {
		case <-reportTicker.C:
			runReport(rep, hostname)
		case <-heartbeatTicker.C:
			if err := rep.SendHeartbeat(); err != nil {
				log.Printf("heartbeat failed: %v", err)
			}
		case <-actionTicker.C:
			// Sync node ID in case re-registration happened
			exec.SetNodeID(rep.GetNodeID())
			if err := exec.PollAndExecute(); err != nil {
				log.Printf("action poll failed: %v", err)
			}
		case sig := <-sigCh:
			log.Printf("received signal %v, shutting down", sig)
			rep.Close()
			exec.Close()
			return
		}
	}
}

func runReport(rep *reporter.Reporter, hostname string) {
	report := reporter.Collect(hostname)

	// Log summary
	log.Printf("report: %d ports, docker=%v, firewall=%s, ssh=%v, authlog=%v",
		len(report.Ports.Listening),
		report.Docker != nil,
		report.Firewall.ActiveBackend,
		report.SSH != nil,
		report.AuthLog != nil,
	)

	// Debug: print report JSON for development
	if os.Getenv("GATEWARDEN_DEBUG") == "1" {
		data, _ := json.MarshalIndent(report, "", "  ")
		log.Printf("debug report:\n%s", string(data))
	}

	if err := rep.SendReport(report); err != nil {
		log.Printf("send report failed: %v", err)
	}
}
