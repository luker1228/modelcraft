-- ============================================
-- Project Auth Schema Queries
-- ============================================

-- ----------------------------------------
-- Project Auth Schema
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
