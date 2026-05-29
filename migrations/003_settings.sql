-- +goose Up
-- +goose StatementBegin

CREATE TABLE settings (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    agent_auto_approve BOOLEAN NOT NULL DEFAULT true,
    heartbeat_interval INTEGER NOT NULL DEFAULT 60,
    log_retention_days INTEGER NOT NULL DEFAULT 30,
    cloudflare_api_token TEXT NOT NULL DEFAULT '',
    tailscale_api_key TEXT NOT NULL DEFAULT '',
    tailscale_tailnet TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default settings row
INSERT INTO settings (id) VALUES (1);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS settings;

-- +goose StatementEnd
