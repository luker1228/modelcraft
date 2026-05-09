-- name: CreateEndUserRole :exec
INSERT INTO end_user_roles (
  id,
  org_name,
  project_slug,
  name,
  description,
  is_implicit,
  is_protected
)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetEndUserRoleByID :one
SELECT *
FROM end_user_roles
WHERE id = ?
  AND org_name = ? AND `end_user_roles`.`deleted_at` = 0 ;

-- name: ListEndUserRolesByProject :many
SELECT *
FROM end_user_roles
WHERE org_name = ?
  AND (project_slug = ? OR is_implicit = TRUE) AND `end_user_roles`.`deleted_at` = 0 ORDER BY is_implicit DESC, name;

-- name: UpdateEndUserRole :execresult
-- 注意：is_implicit=TRUE 或 is_protected=TRUE 的角色由业务层阻断，不走 SQL 层约束
UPDATE end_user_roles
SET name = ?,
    description = ?,
    updated_at = NOW(3)
WHERE id = ?
  AND org_name = ?
  AND is_implicit = FALSE
  AND is_protected = FALSE;

-- name: DeleteEndUserRole :execresult
UPDATE end_user_roles SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED) WHERE (id = ?
  AND org_name = ?
  AND is_implicit = FALSE
  AND is_protected = FALSE) AND `end_user_roles`.`deleted_at` = 0;

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
WHERE rb.role_id = ? AND `b`.`deleted_at` = 0 ;
