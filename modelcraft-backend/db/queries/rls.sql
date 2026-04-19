-- ============================================
-- RLS (Row Level Security) Queries
-- ============================================

-- ----------------------------------------
-- Model RLS Policy Queries
-- ----------------------------------------

-- name: GetModelRLSPolicy :one
SELECT * FROM model_rls_policies
WHERE model_id = ?;

-- name: UpsertModelRLSPolicy :exec
INSERT INTO model_rls_policies (
    model_id, select_predicate, insert_check,
    update_predicate, update_check, delete_predicate
) VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    select_predicate = VALUES(select_predicate),
    insert_check = VALUES(insert_check),
    update_predicate = VALUES(update_predicate),
    update_check = VALUES(update_check),
    delete_predicate = VALUES(delete_predicate);

-- name: DeleteModelRLSPolicy :exec
DELETE FROM model_rls_policies
WHERE model_id = ?;

-- name: ExistsModelRLSPolicy :one
SELECT EXISTS(
    SELECT 1 FROM model_rls_policies
    WHERE model_id = ?
) AS exists_flag;

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
