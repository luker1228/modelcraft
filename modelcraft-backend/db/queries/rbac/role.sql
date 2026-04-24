-- name: CreateEndUserRole :exec
INSERT INTO end_user_roles (
  id,
  org_name,
  project_slug,
  name,
  description,
  is_implicit
)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetEndUserRoleByID :one
SELECT *
FROM end_user_roles
WHERE id = ?
  AND org_name = ?;

-- name: ListEndUserRolesByProject :many
SELECT *
FROM end_user_roles
WHERE org_name = ?
  AND project_slug = ?
ORDER BY is_implicit DESC, name;

-- name: UpdateEndUserRole :execresult
-- 注意：is_implicit=TRUE 的角色由业务层阻断，不走 SQL 层约束
UPDATE end_user_roles
SET name = ?,
    description = ?,
    updated_at = NOW(3)
WHERE id = ?
  AND org_name = ?
  AND is_implicit = FALSE;

-- name: DeleteEndUserRole :execresult
-- 注意：is_implicit=TRUE 的角色由业务层阻断，不走 SQL 层约束
DELETE FROM end_user_roles
WHERE id = ?
  AND org_name = ?
  AND is_implicit = FALSE;

-- name: AssignBundleToRole :exec
INSERT INTO end_user_role_bundles (
  id,
  org_name,
  project_slug,
  role_id,
  bundle_id
)
VALUES (?, ?, ?, ?, ?);

-- name: RevokeBundleFromRole :execresult
DELETE FROM end_user_role_bundles
WHERE role_id = ?
  AND bundle_id = ?;

-- name: ListBundlesByRole :many
SELECT b.*
FROM end_user_permission_bundles b
  JOIN end_user_role_bundles rb ON b.id = rb.bundle_id
WHERE rb.role_id = ?;
