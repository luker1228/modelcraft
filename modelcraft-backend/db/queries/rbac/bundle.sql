-- name: CreateEndUserBundle :exec
INSERT INTO end_user_permission_bundles (
  id,
  org_name,
  project_slug,
  name,
  description
)
VALUES (?, ?, ?, ?, ?);

-- name: GetEndUserBundleByID :one
SELECT *
FROM end_user_permission_bundles
WHERE id = ?
  AND org_name = ?;

-- name: ListEndUserBundlesByProject :many
SELECT *
FROM end_user_permission_bundles
WHERE org_name = ?
  AND project_slug = ?
ORDER BY name;

-- name: UpdateEndUserBundle :execresult
UPDATE end_user_permission_bundles
SET name = ?,
    description = ?,
    updated_at = NOW(3)
WHERE id = ?
  AND org_name = ?;

-- name: DeleteEndUserBundle :execresult
DELETE FROM end_user_permission_bundles
WHERE id = ?
  AND org_name = ?;

-- name: AddPermissionToBundle :exec
INSERT INTO end_user_bundle_permissions (
  id,
  bundle_id,
  permission_id,
  sort_order
)
VALUES (?, ?, ?, ?);

-- name: RemovePermissionFromBundle :execresult
DELETE FROM end_user_bundle_permissions
WHERE bundle_id = ?
  AND permission_id = ?;

-- name: ListPermissionsInBundle :many
SELECT p.*
FROM end_user_permissions p
  JOIN end_user_bundle_permissions bp ON p.id = bp.permission_id
WHERE bp.bundle_id = ?
ORDER BY bp.sort_order, bp.created_at;
