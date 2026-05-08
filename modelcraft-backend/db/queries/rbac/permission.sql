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
  column_policy,
  row_policy
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetEndUserPermissionByID :one
SELECT *
FROM end_user_data_permissions
WHERE id = ?
  AND org_name = ? AND `end_user_data_permissions`.`deleted_at` = 0 ;

-- name: ListEndUserPermissionsByProject :many
SELECT *
FROM end_user_data_permissions
WHERE org_name = ?
  AND project_slug = ? AND `end_user_data_permissions`.`deleted_at` = 0 ORDER BY created_at;

-- name: ListEndUserPermissionsByModel :many
SELECT *
FROM end_user_data_permissions
WHERE model_id = ?
  AND org_name = ? AND `end_user_data_permissions`.`deleted_at` = 0 ORDER BY created_at;

-- name: GetEndUserPermissionByModelAndName :one
SELECT *
FROM end_user_data_permissions
WHERE model_id = ?
  AND org_name = ?
  AND name = ? AND `end_user_data_permissions`.`deleted_at` = 0 LIMIT 1;

-- name: UpdateEndUserPermission :execresult
UPDATE end_user_data_permissions
SET name = ?,
    description = ?,
    column_policy = ?,
    row_policy = ?,
    updated_at = NOW(3)
WHERE id = ?
  AND org_name = ?;

-- name: DeleteEndUserPermission :execresult
UPDATE end_user_data_permissions SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED) WHERE (id = ?
  AND org_name = ?) AND `end_user_data_permissions`.`deleted_at` = 0;

-- name: IsPermissionReferencedByBundleItem :one
SELECT COUNT(*) > 0 AS referenced
FROM end_user_bundle_data_permission_items
WHERE custom_permission_id = ?;
