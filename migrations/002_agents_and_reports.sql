-- +goose Up
-- +goose StatementBegin

CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID UNIQUE NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    api_key_hash VARCHAR(255) NOT NULL,
    version VARCHAR(64) NOT NULL DEFAULT '',
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE agent_reports (
    id BIGSERIAL PRIMARY KEY,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    report JSONB NOT NULL DEFAULT '{}',
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Full text search on reports for later analysis
CREATE INDEX idx_agent_reports_received_at ON agent_reports(received_at DESC);
CREATE INDEX idx_agent_reports_agent_id ON agent_reports(agent_id);
CREATE INDEX idx_agents_node_id ON agents(node_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS agent_reports;
DROP TABLE IF EXISTS agents;

-- +goose StatementEnd
