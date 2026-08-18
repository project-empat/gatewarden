package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/gatewarden/api/internal/integration"
)

const (
	vulnCacheTTL  = 24 * time.Hour
	vulnBatchSize = 100 // packages queried against OSV per enrichment pass
	vulnEcosystem = "deb"
)

// VulnerabilityService matches installed packages against OSV for known
// CVEs. Lookups are cached in vuln_cache (with a TTL) so the feed is not
// hammered on every report; enrichment runs in bounded background passes.
type VulnerabilityService struct {
	db  *pgxpool.Pool
	log *zap.SugaredLogger
	osv *integration.OSVClient
	mu  sync.Mutex // guards one enrichment pass at a time
}

func NewVulnerabilityService(db *pgxpool.Pool, log *zap.SugaredLogger) *VulnerabilityService {
	return &VulnerabilityService{
		db:  db,
		log: log,
		osv: integration.NewOSVClient(),
	}
}

// VulnerablePackage is a package with one or more known CVEs on a node.
type VulnerablePackage struct {
	NodeID    string `json:"node_id"`
	Hostname  string `json:"hostname"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	CveCount  int    `json:"cve_count"`
	TopCVE    string `json:"top_cve,omitempty"`
	Summary   string `json:"summary,omitempty"`
	CheckedAt string `json:"checked_at"`
}

// Enrich runs one bounded OSV enrichment pass for a node: it queries the
// packages for that node that are not yet cached (or whose cache is stale),
// up to vulnBatchSize, and records the results.
func (s *VulnerabilityService) Enrich(ctx context.Context, nodeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Pick up to vulnBatchSize uncached/stale installed packages for this node.
	rows, err := s.db.Query(ctx, `
		SELECT np.name, np.version
		FROM node_packages np
		LEFT JOIN vuln_cache vc
		  ON vc.package = np.name AND vc.version = np.version AND vc.ecosystem = np.ecosystem
		 AND vc.updated_at > NOW() - $2::interval
		WHERE np.node_id = $1 AND vc.package IS NULL
		LIMIT $3`, nodeID, vulnCacheTTL.String(), vulnBatchSize)
	if err != nil {
		return fmt.Errorf("query packages to enrich: %w", err)
	}
	defer rows.Close()

	type pkg struct{ name, version string }
	var pkgs []pkg
	for rows.Next() {
		var p pkg
		if err := rows.Scan(&p.name, &p.version); err != nil {
			continue
		}
		pkgs = append(pkgs, p)
	}
	rows.Close()

	for _, p := range pkgs {
		if err := ctx.Err(); err != nil {
			return nil
		}
		vulns, err := s.osv.QueryVulnerabilities(ctx, p.name, p.version, vulnEcosystem)
		if err != nil {
			s.log.Debugw("osv lookup failed", "package", p.name, "error", err)
			// Record a stale-ish negative so we don't spin on a down feed.
			vulns = []integration.VulnSummary{}
		}
		if err := s.storeCache(ctx, p.name, p.version, vulns); err != nil {
			s.log.Warnw("failed to store vuln cache", "error", err)
		}
	}
	return nil
}

func (s *VulnerabilityService) storeCache(ctx context.Context, name, version string, vulns []integration.VulnSummary) error {
	b, err := json.Marshal(vulns)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx,
		`INSERT INTO vuln_cache (package, version, ecosystem, vulnerabilities, updated_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (package, version, ecosystem) DO UPDATE
		 SET vulnerabilities = EXCLUDED.vulnerabilities, updated_at = NOW()`,
		name, version, vulnEcosystem, string(b))
	return err
}

// List returns every node-package with at least one known CVE, newest first.
func (s *VulnerabilityService) List(ctx context.Context) ([]VulnerablePackage, error) {
	return s.query(ctx, "")
}

// NodeList returns vulnerable packages for a single node.
func (s *VulnerabilityService) NodeList(ctx context.Context, nodeID string) ([]VulnerablePackage, error) {
	return s.query(ctx, nodeID)
}

func (s *VulnerabilityService) query(ctx context.Context, nodeID string) ([]VulnerablePackage, error) {
	sql := `
		SELECT np.node_id, n.hostname, np.name, np.version, vc.vulnerabilities, vc.updated_at
		FROM node_packages np
		JOIN vuln_cache vc
		  ON vc.package = np.name AND vc.version = np.version AND vc.ecosystem = np.ecosystem
		LEFT JOIN nodes n ON n.id = np.node_id
		WHERE vc.vulnerabilities <> '[]'`
	args := []any{}
	if nodeID != "" {
		args = append(args, nodeID)
		sql += " AND np.node_id = $1"
	}
	sql += " ORDER BY vc.updated_at DESC LIMIT 500"

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query vulnerable packages: %w", err)
	}
	defer rows.Close()

	var out []VulnerablePackage
	for rows.Next() {
		var vp VulnerablePackage
		var vulnsJSON string
		var checkedAt time.Time
		if err := rows.Scan(&vp.NodeID, &vp.Hostname, &vp.Name, &vp.Version, &vulnsJSON, &checkedAt); err != nil {
			continue
		}
		var vulns []integration.VulnSummary
		_ = json.Unmarshal([]byte(vulnsJSON), &vulns)
		vp.CveCount = len(vulns)
		if len(vulns) > 0 {
			vp.TopCVE = vulns[0].ID
			vp.Summary = vulns[0].Summary
		}
		vp.CheckedAt = checkedAt.UTC().Format(time.RFC3339)
		out = append(out, vp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vulnerable packages: %w", err)
	}
	if out == nil {
		out = []VulnerablePackage{}
	}
	return out, nil
}

// FIMFile is a monitored file plus its change state on a node.
type FIMFile struct {
	NodeID    string `json:"node_id"`
	Hostname  string `json:"hostname"`
	Path      string `json:"path"`
	ChangedAt string `json:"changed_at,omitempty"`
	FirstSeen string `json:"first_seen"`
}

// FIMChanges returns monitored files that have changed (non-null last_changed)
// across all nodes, newest first.
func (s *VulnerabilityService) FIMChanges(ctx context.Context) ([]FIMFile, error) {
	return s.queryFIM(ctx, "")
}

// NodeFIM returns all monitored files for one node, including changed ones.
func (s *VulnerabilityService) NodeFIM(ctx context.Context, nodeID string) ([]FIMFile, error) {
	return s.queryFIM(ctx, nodeID)
}

func (s *VulnerabilityService) queryFIM(ctx context.Context, nodeID string) ([]FIMFile, error) {
	sql := `
		SELECT f.node_id, n.hostname, f.path, f.last_changed, f.first_seen
		FROM node_fim_files f
		LEFT JOIN nodes n ON n.id = f.node_id
		WHERE ($1 = '' OR f.node_id = $1)
		  AND f.last_changed IS NOT NULL
		ORDER BY f.last_changed DESC LIMIT 300`
	rows, err := s.db.Query(ctx, sql, nodeID)
	if err != nil {
		return nil, fmt.Errorf("query fim: %w", err)
	}
	defer rows.Close()

	var out []FIMFile
	for rows.Next() {
		var f FIMFile
		var changedAt, firstSeen *time.Time
		if err := rows.Scan(&f.NodeID, &f.Hostname, &f.Path, &changedAt, &firstSeen); err != nil {
			continue
		}
		if changedAt != nil {
			f.ChangedAt = changedAt.UTC().Format(time.RFC3339)
		}
		if firstSeen != nil {
			f.FirstSeen = firstSeen.UTC().Format(time.RFC3339)
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fim: %w", err)
	}
	if out == nil {
		out = []FIMFile{}
	}
	return out, nil
}
