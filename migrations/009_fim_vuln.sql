-- +goose Up
-- +goose StatementBegin

-- File Integrity Monitoring (periodic hash snapshots per node, plus change tracking).
CREATE TABLE node_fim_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    hash VARCHAR(64) NOT NULL,
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_changed TIMESTAMPTZ,
    UNIQUE (node_id, path)
);
CREATE INDEX idx_node_fim_node ON node_fim_files(node_id);

-- Installed packages per node for vulnerability matching (Debian ecosystem).
CREATE TABLE node_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    ecosystem VARCHAR(64) NOT NULL DEFAULT 'deb',
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (node_id, name, version)
);
CREATE INDEX idx_node_packages_node ON node_packages(node_id);

-- Cache of OSV lookups (per package@version) so we don't re-query the feed
-- on every report. Empty vulnerabilities = checked, no known CVEs.
CREATE TABLE vuln_cache (
    package VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    ecosystem VARCHAR(64) NOT NULL DEFAULT 'deb',
    vulnerabilities TEXT NOT NULL DEFAULT '[]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (package, version, ecosystem)
);

ALTER TABLE nodes ADD COLUMN security_updates_pending INT NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS node_fim_files;
DROP TABLE IF EXISTS node_packages;
DROP TABLE IF EXISTS vuln_cache;
ALTER TABLE nodes DROP COLUMN IF EXISTS security_updates_pending;

-- +goose StatementEnd
