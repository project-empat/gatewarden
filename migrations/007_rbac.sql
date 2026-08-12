-- +goose Up
-- +goose StatementBegin

-- Basic RBAC: role on the user row (admin/viewer), plus roles and
-- assignments tables ready for the advanced (enterprise) policy engine.

ALTER TABLE users ADD COLUMN role VARCHAR(32) NOT NULL DEFAULT 'viewer';

CREATE TABLE roles (
    id TEXT PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    permissions TEXT NOT NULL DEFAULT '[]',
    parent_id TEXT,
    system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    scope TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE policy_rules (
    id TEXT PRIMARY KEY,
    role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    effect VARCHAR(16) NOT NULL,
    action VARCHAR(128) NOT NULL,
    resource VARCHAR(128) NOT NULL,
    conditions TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed roles; the existing seed user becomes admin.
INSERT INTO roles (id, name, description, permissions, system) VALUES
    ('role_admin', 'Admin', 'Full access', '["*"]', true),
    ('role_viewer', 'Viewer', 'Read-only access', '["nodes:read","incidents:read","policies:read","settings:read","reports:read","graph:read"]', true);

UPDATE users SET role = 'admin' WHERE email = 'admin@gatewarden.dev';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS policy_rules;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
ALTER TABLE users DROP COLUMN IF EXISTS role;

-- +goose StatementEnd
