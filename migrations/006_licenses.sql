-- +goose Up
-- +goose StatementBegin

CREATE TABLE licenses (
    id TEXT PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL UNIQUE,
    plan VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    seats INT NOT NULL DEFAULT 0,
    features TEXT NOT NULL DEFAULT '[]',
    issued_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    grace_period_end TIMESTAMPTZ,
    metadata TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_licenses_tenant ON licenses(tenant_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS licenses;

-- +goose StatementEnd
