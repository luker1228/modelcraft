-- ============================================
-- RLS Policy V2 Queries — 多策略存储查询
-- ============================================

-- name: ListPoliciesByAction :many
SELECT * FROM model_rls_policies
WHERE org_name = ? AND project_slug = ? AND model_id = ? AND action = ?
  AND role IN (sqlc.slice('roles'));

-- name: ListPoliciesByModel :many
SELECT * FROM model_rls_policies
WHERE org_name = ? AND project_slug = ? AND model_id = ?
ORDER BY action, role;

-- name: UpsertPolicy :exec
INSERT INTO model_rls_policies (
    org_name, project_slug, model_id, policy_name, action, role, using_expr, with_check_expr
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    policy_name = VALUES(policy_name),
    using_expr = VALUES(using_expr),
    with_check_expr = VALUES(with_check_expr);

-- name: DeletePolicy :exec
DELETE FROM model_rls_policies
WHERE id = ? AND org_name = ? AND project_slug = ?;

-- name: DeletePoliciesByModel :exec
DELETE FROM model_rls_policies
WHERE org_name = ? AND project_slug = ? AND model_id = ?;

-- name: GetPolicyByRoleAction :one
SELECT * FROM model_rls_policies
WHERE org_name = ? AND project_slug = ? AND model_id = ? AND action = ? AND role = ?;

-- name: DeletePoliciesByRole :exec
DELETE FROM model_rls_policies
WHERE org_name = ? AND project_slug = ? AND model_id = ? AND role = ?;

-- name: PolicyExists :one
SELECT EXISTS(
    SELECT 1 FROM model_rls_policies
    WHERE org_name = ? AND project_slug = ? AND model_id = ? AND action = ? AND role = ?
) AS exists_flag;
