package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GraphService builds a Cytoscape.js-compatible infrastructure security graph.
type GraphService struct {
	db *pgxpool.Pool
}

func NewGraphService(db *pgxpool.Pool) *GraphService {
	return &GraphService{db: db}
}

// GraphElem represents a single element in a Cytoscape.js graph.
type GraphElem struct {
	Data GraphElemData `json:"data"`
}

// GraphElemData carries the element's identity and visual attributes.
type GraphElemData struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Type     string `json:"type"`    // node, service, container, tunnel, firewall, integration, attack, incident, internet
	Status   string `json:"status,omitempty"`
	Severity string `json:"severity,omitempty"`
	Port     int    `json:"port,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Image    string `json:"image,omitempty"`
	Detail   string `json:"detail,omitempty"`
	NodeID   string `json:"node_id,omitempty"`
	// For edges
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`
	Color  string `json:"color,omitempty"`
}

// GraphResponse is the Cytoscape.js elements array.
type GraphResponse struct {
	Elements []GraphElem `json:"elements"`
}

// summaryFields captures top-level fields from an agent report.
type agentReportSummary struct {
	SSH       map[string]interface{} `json:"ssh"`
	Docker    map[string]interface{} `json:"docker"`
	Firewall  map[string]interface{} `json:"firewall"`
	CrowdSec  map[string]interface{} `json:"crowdsec"`
	Fail2Ban  map[string]interface{} `json:"fail2ban"`
	Tailscale map[string]interface{} `json:"tailscale"`
	Ports     []interface{}          `json:"ports"`
}

// GetFullGraph returns the complete infrastructure security graph.
func (s *GraphService) GetFullGraph(ctx context.Context) (*GraphResponse, error) {
	var elems []GraphElem

	// -- Internet node (always present) --
	elems = append(elems, GraphElem{Data: GraphElemData{
		ID:    "internet",
		Label: "Internet",
		Type:  "internet",
	}})

	// -- Load nodes --
	type nodeInfo struct {
		ID        string
		Hostname  string
		IP        string
		OS        string
		Status    string
		LastSeen  time.Time
		ReportRaw []byte
	}
	rows, err := s.db.Query(ctx, `
		SELECT n.id, n.hostname, n.ip, n.os, n.status, n.last_seen, ar.report
		FROM nodes n
		LEFT JOIN LATERAL (
			SELECT report FROM agent_reports
			WHERE node_id = n.id
			ORDER BY received_at DESC
			LIMIT 1
		) ar ON true
		WHERE n.status != 'deleted'
		ORDER BY n.hostname
	`)
	if err != nil {
		return nil, fmt.Errorf("query graph nodes: %w", err)
	}
	defer rows.Close()

	var nodes []nodeInfo
	for rows.Next() {
		var ni nodeInfo
		if err := rows.Scan(&ni.ID, &ni.Hostname, &ni.IP, &ni.OS, &ni.Status, &ni.LastSeen, &ni.ReportRaw); err != nil {
			continue
		}
		nodes = append(nodes, ni)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate nodes: %w", err)
	}

	// -- Load incidents --
	type incidentInfo struct {
		ID       string
		NodeID   string
		Severity string
		Title    string
		Status   string
	}
	incRows, err := s.db.Query(ctx, `
		SELECT id, node_id, severity, title, status
		FROM incidents
		WHERE status = 'open'
		ORDER BY created_at DESC
		LIMIT 50
	`)
	if err != nil {
		return nil, fmt.Errorf("query graph incidents: %w", err)
	}
	defer incRows.Close()

	incidentCount := 0
	incidentByNode := make(map[string][]incidentInfo)
	for incRows.Next() {
		var inc incidentInfo
		if err := incRows.Scan(&inc.ID, &inc.NodeID, &inc.Severity, &inc.Title, &inc.Status); err != nil {
			continue
		}
		incidentByNode[inc.NodeID] = append(incidentByNode[inc.NodeID], inc)
		incidentCount++
	}
	incRows.Close()

	// Build elements per-node
	for _, ni := range nodes {
		nodeID := ni.ID
		nodeElemID := "node-" + nodeID

		// Node entity
		elems = append(elems, GraphElem{Data: GraphElemData{
			ID:       nodeElemID,
			Label:    ni.Hostname,
			Type:     "node",
			Status:   ni.Status,
			Hostname: ni.Hostname,
			NodeID:   nodeID,
		}})

		// Edge: Internet → Node (if node has any public exposure)
		elems = append(elems, GraphElem{Data: GraphElemData{
			ID:     "edge-internet-" + nodeID,
			Source: "internet",
			Target: nodeElemID,
			Label:  "routes to",
			Color:  "#6b7280",
		}})

		// Parse latest report
		if ni.ReportRaw == nil {
			continue
		}
		var report map[string]interface{}
		if err := json.Unmarshal(ni.ReportRaw, &report); err != nil {
			continue
		}

		// -- Services from ports --
		if ports, ok := report["ports"].([]interface{}); ok {
			for idx, p := range ports {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				portFloat, _ := pm["port"].(float64)
				portInt := int(portFloat)
				proto, _ := pm["protocol"].(string)
				proc, _ := pm["process"].(string)
				exposure, _ := pm["exposure"].(string)

				svcID := fmt.Sprintf("svc-%s-%d", nodeID, portInt)
				svcLabel := fmt.Sprintf("%s:%d", proc, portInt)
				if proc == "" {
					svcLabel = fmt.Sprintf("Port %d", portInt)
				}

				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     svcID,
					Label:  svcLabel,
					Type:   "service",
					Port:   portInt,
					Status: exposure,
					Detail: fmt.Sprintf("%s/%s", proto, exposure),
					NodeID: nodeID,
				}})

				// Edge: Node → Service
				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     fmt.Sprintf("edge-%s-svc-%d", nodeID, portInt),
					Source: nodeElemID,
					Target: svcID,
					Label:  "listens on",
					Color:  "#3b82f6",
				}})

				// Edge: Internet → Service if publicly exposed
				if exposure == "public" {
					elems = append(elems, GraphElem{Data: GraphElemData{
						ID:     fmt.Sprintf("edge-internet-svc-%s-%d", nodeID, portInt),
						Source: "internet",
						Target: svcID,
						Label:  "exposes",
						Color:  "#ef4444",
					}})
				}
				_ = idx
			}
		}

		// -- Containers from Docker --
		if docker, ok := report["docker"].(map[string]interface{}); ok {
			if containers, ok := docker["containers"].([]interface{}); ok {
				for _, c := range containers {
					cm, ok := c.(map[string]interface{})
					if !ok {
						continue
					}
					containerID, _ := cm["container_id"].(string)
					containerName, _ := cm["name"].(string)
					image, _ := cm["image"].(string)
					containerStatus, _ := cm["status"].(string)
					ctrShort := containerID
					if len(ctrShort) > 12 {
						ctrShort = ctrShort[:12]
					}

					ctrElemID := fmt.Sprintf("ctr-%s-%s", nodeID, strings.ReplaceAll(containerID, "-", ""))
					ctrLabel := containerName
					if ctrLabel == "" {
						ctrLabel = ctrShort
					}

					elems = append(elems, GraphElem{Data: GraphElemData{
						ID:     ctrElemID,
						Label:  ctrLabel,
						Type:   "container",
						Status: containerStatus,
						Image:  image,
						NodeID: nodeID,
					}})

					// Edge: Node → Container
					elems = append(elems, GraphElem{Data: GraphElemData{
						ID:     fmt.Sprintf("edge-%s-ctr-%s", nodeID, strings.ReplaceAll(containerID, "-", "")),
						Source: nodeElemID,
						Target: ctrElemID,
						Label:  "runs",
						Color:  "#10b981",
					}})

					// Map container ports to services
					if publishedPorts, ok := cm["published_ports"].([]interface{}); ok {
						for _, pp := range publishedPorts {
							ppm, ok := pp.(map[string]interface{})
							if !ok {
								continue
							}
							hostPort, _ := ppm["host_port"].(float64)
							hostPortInt := int(hostPort)
							svcID := fmt.Sprintf("svc-%s-%d", nodeID, hostPortInt)

							elems = append(elems, GraphElem{Data: GraphElemData{
								ID:     fmt.Sprintf("edge-ctr-%s-svc-%d", strings.ReplaceAll(containerID, "-", ""), hostPortInt),
								Source: ctrElemID,
								Target: svcID,
								Label:  "exposes via",
								Color:  "#f59e0b",
							}})
						}
					}
				}
			}

			// Docker socket exposure
			if socketExposed, _ := docker["socket_exposed"].(bool); socketExposed {
				dockerIssueID := "docker-socket-" + nodeID
				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     dockerIssueID,
					Label:  "Docker Socket Exposed",
					Type:   "incident",
					Detail: "/var/run/docker.sock mounted in container",
					NodeID: nodeID,
				}})
				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     "edge-" + nodeID + "-docker-socket",
					Source: nodeElemID,
					Target: dockerIssueID,
					Label:  "has issue",
					Color:  "#ef4444",
				}})
			}
		}

		// -- Firewall --
		if fw, ok := report["firewall"].(map[string]interface{}); ok {
			fwActive, _ := fw["active"].(bool)
			fwStatus := "active"
			if !fwActive {
				fwStatus = "inactive"
			}
			fwElemID := "fw-" + nodeID
			fwLabel := fmt.Sprintf("UFW (%s)", fwStatus)
			fwBackend, _ := fw["backend"].(string)
			if fwBackend != "" {
				fwLabel = fmt.Sprintf("%s (%s)", fwBackend, fwStatus)
			}

			elems = append(elems, GraphElem{Data: GraphElemData{
				ID:     fwElemID,
				Label:  fwLabel,
				Type:   "firewall",
				Status: fwStatus,
				NodeID: nodeID,
			}})

			elems = append(elems, GraphElem{Data: GraphElemData{
				ID:     "edge-" + nodeID + "-fw",
				Source: nodeElemID,
				Target: fwElemID,
				Label:  "protected by",
				Color:  "#06b6d4",
			}})

			// Edge: Firewall → Services it allows
			if rules, ok := fw["rules"].([]interface{}); ok {
				for _, r := range rules {
					rm, ok := r.(map[string]interface{})
					if !ok {
						continue
					}
					rPort, _ := rm["port"].(float64)
					rPortInt := int(rPort)
					if rPortInt > 0 {
						svcID := fmt.Sprintf("svc-%s-%d", nodeID, rPortInt)
						elems = append(elems, GraphElem{Data: GraphElemData{
							ID:     fmt.Sprintf("edge-fw-%s-svc-%d", nodeID, rPortInt),
							Source: fwElemID,
							Target: svcID,
							Label:  "allows",
							Color:  "#10b981",
						}})
					}
				}
			}
		}

		// -- CrowdSec --
		if cs, ok := report["crowdsec"].(map[string]interface{}); ok {
			csInstalled, _ := cs["installed"].(bool)
			if csInstalled {
				csRunning, _ := cs["running"].(bool)
				csStatus := "active"
				if !csRunning {
					csStatus = "inactive"
				}
				csDecisions, _ := cs["active_decisions"].(float64)

				csElemID := "crowdsec-" + nodeID
				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     csElemID,
					Label:  fmt.Sprintf("CrowdSec (%.0f decisions)", csDecisions),
					Type:   "integration",
					Status: csStatus,
					NodeID: nodeID,
				}})

				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     "edge-" + nodeID + "-crowdsec",
					Source: nodeElemID,
					Target: csElemID,
					Label:  "runs",
					Color:  "#ec4899",
				}})
			}
		}

		// -- Fail2Ban --
		if f2b, ok := report["fail2ban"].(map[string]interface{}); ok {
			f2bInstalled, _ := f2b["installed"].(bool)
			if f2bInstalled {
				f2bRunning, _ := f2b["running"].(bool)
				f2bStatus := "active"
				if !f2bRunning {
					f2bStatus = "inactive"
				}
				f2bJails, _ := f2b["jails"].(float64)
				f2bBans, _ := f2b["current_bans"].(float64)

				f2bElemID := "fail2ban-" + nodeID
				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     f2bElemID,
					Label:  fmt.Sprintf("Fail2Ban (%.0f jails, %.0f bans)", f2bJails, f2bBans),
					Type:   "integration",
					Status: f2bStatus,
					NodeID: nodeID,
				}})

				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     "edge-" + nodeID + "-fail2ban",
					Source: nodeElemID,
					Target: f2bElemID,
					Label:  "runs",
					Color:  "#ec4899",
				}})
			}
		}

		// -- SSH status --
		if ssh, ok := report["ssh"].(map[string]interface{}); ok {
			sshExposed, _ := ssh["publicly_exposed"].(bool)
			pwAuth, _ := ssh["password_auth"].(bool)
			sshPort, _ := ssh["port"].(float64)
			sshPortInt := int(sshPort)
			if sshPortInt == 0 {
				sshPortInt = 22
			}

			svcID := fmt.Sprintf("svc-%s-%d", nodeID, sshPortInt)
			// If SSH already added as a service, update its attributes.
			// We add it here regardless since it's important.
			sshSvcLabel := fmt.Sprintf("SSH:%d", sshPortInt)
			if pwAuth {
				sshSvcLabel += " (password auth!)"
			}
			elems = append(elems, GraphElem{Data: GraphElemData{
				ID:     svcID,
				Label:  sshSvcLabel,
				Type:   "service",
				Port:   sshPortInt,
				Status: "public",
				Detail: fmt.Sprintf("PasswordAuth=%v RootLogin=%v", pwAuth, ssh["root_login"]),
				NodeID: nodeID,
			}})

			elems = append(elems, GraphElem{Data: GraphElemData{
				ID:     "edge-" + nodeID + "-ssh",
				Source: nodeElemID,
				Target: svcID,
				Label:  "listens on",
				Color:  "#3b82f6",
			}})

			if sshExposed {
				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     "edge-internet-ssh-" + nodeID,
					Source: "internet",
					Target: svcID,
					Label:  "exposes",
					Color:  "#ef4444",
				}})
			}

			// SSH hardening issues as incidents
			if pwAuth {
				pwIncID := "ssh-pw-" + nodeID
				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     pwIncID,
					Label:  "SSH Password Auth Enabled",
					Type:   "incident",
					Detail: "Password authentication is enabled",
					NodeID: nodeID,
				}})
				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     "edge-" + nodeID + "-ssh-pw",
					Source: nodeElemID,
					Target: pwIncID,
					Label:  "has issue",
					Color:  "#ef4444",
				}})
			}
		}

		// -- Incidents for this node --
		if incs, ok := incidentByNode[nodeID]; ok {
			for _, inc := range incs {
				incElemID := "inc-" + inc.ID
				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:       incElemID,
					Label:    inc.Title,
					Type:     "incident",
					Severity: inc.Severity,
					Status:   inc.Status,
					NodeID:   nodeID,
				}})
				elems = append(elems, GraphElem{Data: GraphElemData{
					ID:     "edge-inc-" + nodeID + "-" + inc.ID,
					Source: nodeElemID,
					Target: incElemID,
					Label:  "reported by",
					Color:  "#f97316",
				}})
			}
		}
	}

	// Deduplicate elements by ID (later wins)
	seen := make(map[string]bool)
	var deduped []GraphElem
	for _, e := range elems {
		elemID := e.Data.ID
		if e.Data.Source != "" || e.Data.Target != "" {
			// Edge — use composite key
			elemID = e.Data.Source + "→" + e.Data.Target + ":" + e.Data.Label
			if e.Data.Label == "" {
				elemID = e.Data.Source + "→" + e.Data.Target
			}
		}
		if seen[elemID] {
			continue
		}
		seen[elemID] = true
		deduped = append(deduped, e)
	}
	// Reassign sequential IDs to deduplicated elements
	for i := range deduped {
		el := &deduped[i]
		if el.Data.Source == "" && el.Data.Target == "" {
			el.Data.ID = fmt.Sprintf("e%d", i+1)
		} else {
			el.Data.ID = fmt.Sprintf("ee%d", i+1)
		}
	}

	_ = incidentCount
	return &GraphResponse{Elements: deduped}, nil
}

// GetNodeGraph returns the sub-graph for a specific node (node + 1-hop neighbors).
func (s *GraphService) GetNodeGraph(ctx context.Context, nodeID string) (*GraphResponse, error) {
	full, err := s.GetFullGraph(ctx)
	if err != nil {
		return nil, err
	}

	// Filter: keep only elements connected to the node
	nodeElemID := "node-" + nodeID
	nodeSet := map[string]bool{nodeElemID: true}

	// Collect all edges touching this node, then add their targets/sources
	for _, el := range full.Elements {
		if el.Data.Source == nodeElemID || el.Data.Target == nodeElemID {
			nodeSet[el.Data.Source] = true
			nodeSet[el.Data.Target] = true
		}
	}

	var filtered []GraphElem
	keepEdge := func(el GraphElem) bool {
		return nodeSet[el.Data.Source] && nodeSet[el.Data.Target]
	}
	for _, el := range full.Elements {
		if el.Data.Source != "" || el.Data.Target != "" {
			// Edge
			if keepEdge(el) {
				filtered = append(filtered, el)
			}
		} else {
			// Node
			if nodeSet[el.Data.ID] {
				filtered = append(filtered, el)
			}
		}
	}

	// Ensure the internet node is included
	internetIncluded := false
	for _, el := range filtered {
		if el.Data.ID == "internet" || el.Data.Label == "Internet" {
			internetIncluded = true
			break
		}
	}
	if !internetIncluded {
		filtered = append([]GraphElem{{Data: GraphElemData{
			ID: "internet", Label: "Internet", Type: "internet",
		}}}, filtered...)
	}

	return &GraphResponse{Elements: filtered}, nil
}

// GetGraphStats returns aggregate statistics about the security graph.
func (s *GraphService) GetGraphStats(ctx context.Context) (*GraphStats, error) {
	stats := &GraphStats{}

	// Count nodes by type from the full graph
	graph, err := s.GetFullGraph(ctx)
	if err != nil {
		return nil, err
	}

	typeCount := make(map[string]int)
	for _, el := range graph.Elements {
		if el.Data.Source == "" && el.Data.Target == "" {
			typeCount[el.Data.Type]++
		}
	}

	stats.Nodes = typeCount["node"]
	stats.Services = typeCount["service"]
	stats.Containers = typeCount["container"]
	stats.Integrations = typeCount["integration"]
	stats.Incidents = typeCount["incident"]
	stats.Firewalls = typeCount["firewall"]
	stats.TotalElements = len(graph.Elements)

	// Count risk paths (public exposures)
	stats.PublicExposures = typeCount["service"] // approximate

	return stats, nil
}

// GraphStats holds aggregate counts for the graph.
type GraphStats struct {
	Nodes            int `json:"nodes"`
	Services         int `json:"services"`
	Containers       int `json:"containers"`
	Integrations     int `json:"integrations"`
	Incidents        int `json:"incidents"`
	Firewalls        int `json:"firewalls"`
	TotalElements    int `json:"total_elements"`
	PublicExposures  int `json:"public_exposures"`
}
