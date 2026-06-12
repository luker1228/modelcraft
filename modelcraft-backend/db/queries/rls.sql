-- ============================================
-- RLS (Row Level Security) Queries
-- ============================================

-- NOTE: model_rls_policies queries moved to rls_policy_v2.sql
-- Old single-policy queries (GetModelRLSPolicy, UpsertModelRLSPolicy,
-- DeleteModelRLSPolicy, ExistsModelRLSPolicy) removed.

-- ----------------------------------------
-- Project Auth Schema Queries
-- ----------------------------------------

-- name: GetProjectAuthSchema :one
SELECT * FROM project_auth_schemas
WHERE org_name = ? AND project_slug = ?;

-- name: UpsertProjectAuthSchema :exec
INSERT INTO project_auth_schemas (
    org_name, project_slug, variables
) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE
    variables = VALUES(variables);

-- name: DeleteProjectAuthSchema :exec
DELETE FROM project_auth_schemas
WHERE org_name = ? AND project_slug = ?;
