-- ============================================================================
-- Authentication Configuration Schema
-- ============================================================================
-- This schema defines project-level authentication configuration for ModelCraft.
-- Each project can configure its own authentication provider (Casdoor, Keycloak, OIDC).

-- Create project_auth_configs table
CREATE TABLE IF NOT EXISTS project_auth_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Primary key',

    -- 项目关联字段（复合引用）
    org_name VARCHAR(36) NOT NULL COMMENT '所属组织名称（来自projects表复合主键）',
    project_slug VARCHAR(64) NOT NULL COMMENT '所属项目标识符（来自projects表复合主键）',

    provider VARCHAR(32) NOT NULL COMMENT 'Authentication provider type: casdoor, keycloak, oidc',
    enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT 'Whether authentication is enabled for this project',
    config JSON NOT NULL COMMENT 'Provider-specific configuration (endpoint, client_id, certificate, etc.)',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Record creation timestamp',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Record last update timestamp',

    -- 不使用外键约束（避免复杂性，Project永不删除只归档）

    -- 唯一约束：每个项目只能有一个认证配置
    UNIQUE KEY idx_project_auth_unique (org_name, project_slug),

    -- Check constraint: provider must be one of supported values
    CONSTRAINT chk_provider_type
        CHECK (provider IN ('casdoor', 'keycloak', 'oidc'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Project-level authentication configuration';

-- Create index on project for fast lookups
CREATE INDEX idx_project_auth_configs_project ON project_auth_configs (org_name, project_slug);

-- Create index on provider for analytics/reporting
CREATE INDEX idx_project_auth_configs_provider ON project_auth_configs (provider);

-- ============================================================================
-- Example Configuration JSON Structure
-- ============================================================================
-- Casdoor Configuration:
-- {
--   "endpoint": "https://casdoor.example.com",
--   "client_id": "abc123",
--   "client_secret": "secret",
--   "organization": "tenant1",
--   "application": "modelcraft",
--   "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
-- }
--
-- Keycloak Configuration (future):
-- {
--   "realm": "myrealm",
--   "auth_server_url": "https://keycloak.example.com",
--   "jwks_uri": "https://keycloak.example.com/realms/myrealm/protocol/openid-connect/certs",
--   "client_id": "modelcraft"
-- }
--
-- Generic OIDC Configuration (future):
-- {
--   "issuer": "https://sso.company.com",
--   "jwks_uri": "https://sso.company.com/.well-known/jwks.json",
--   "client_id": "modelcraft"
-- }
