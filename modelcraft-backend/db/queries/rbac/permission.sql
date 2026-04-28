-- name: CreateEndUserPermission :exec
INSERT INTO end_user_data_permissions (
  id,
  org_name,
  project_slug,
  database_name,
  model_name,
  model_id,
  name,
  description,
  type,
  column_policy,
  row_policy,
  preset
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetEndUserPermissionByID :one
SELECT *
FROM end_user_data_permissions
WHERE id = ?
  AND org_name = ?;

-- name: ListEndUserPermissionsByProject :many
SELECT *
FROM end_user_data_permissions
WHERE org_name = ?
  AND project_slug = ?
ORDER BY created_at;

-- name: ListEndUserPermissionsByModel :many
SELECT *
FROM end_user_data_permissions
WHERE model_id = ?
  AND org_name = ?
ORDER BY created_at;

-- name: ListPresetPermissionsByModel :many
SELECT *
FROM end_user_data_permissions
WHERE model_id = ?
  AND org_name = ?
  AND type = 'PRESET'
ORDER BY created_at;

-- name: GetEndUserPermissionByModelTypeName :one
SELECT *
FROM end_user_data_permissions
WHERE model_id = ?
  AND org_name = ?
  AND type = ?
  AND name = ?
LIMIT 1;

-- name: UpdateEndUserPermission :execresult
UPDATE end_user_data_permissions
SET name = ?,
    description = ?,
    column_policy = ?,
    updated_at = NOW(3)
WHERE id = ?
  AND org_name = ?;

-- name: DeleteEndUserPermission :execresult
DELETE FROM end_user_data_permissions
WHERE id = ?
  AND org_name = ?;

-- name: DeleteEndUserPermissionsByModelAndType :execresult
DELETE FROM end_user_data_permissions
WHERE model_id = ?
  AND org_name = ?
  AND type = ?;

-- name: UpdateEndUserPresetPermission :execresult
UPDATE end_user_data_permissions
SET name = ?,
    description = ?,
    row_policy = ?,
    preset = ?,
    updated_at = NOW(3)
WHERE id = ?
  AND org_name = ?;

-- name: IsPermissionReferencedByBundle :one
SELECT COUNT(*) > 0 AS referenced
FROM end_user_bundle_permissions
WHERE permission_id = ?;
