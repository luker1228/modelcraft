-- name: CreateEndUserBundle :exec
INSERT INTO end_user_permission_bundles (
  id,
  slug,
  org_name,
  project_slug,
  name,
  description
)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetEndUserBundleByID :one
SELECT *
FROM end_user_permission_bundles
WHERE id = ?
  AND org_name = ?
  AND project_slug = ?;

-- name: GetEndUserBundleBySlug :one
SELECT *
FROM end_user_permission_bundles
WHERE slug = ?
  AND org_name = ?
  AND project_slug = ?;

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
  AND org_name = ?
  AND project_slug = ?;

-- name: DeleteEndUserBundle :execresult
DELETE FROM end_user_permission_bundles
WHERE id = ?
  AND org_name = ?
  AND project_slug = ?;

-- ─── Bundle Data Permission Items ───────────────────────────────

-- name: UpsertBundleDataPermissionItem :exec
-- Replace 语义：同一 bundle+model 最多一个 item
INSERT INTO end_user_bundle_data_permission_items (
  id,
  bundle_id,
  model_id,
  grant_type,
  preset,
  custom_permission_id,
  sort_order
)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  grant_type = VALUES(grant_type),
  preset = VALUES(preset),
  custom_permission_id = VALUES(custom_permission_id),
  sort_order = VALUES(sort_order),
  updated_at = NOW(3);

-- name: RemoveBundleDataPermissionItem :execresult
DELETE FROM end_user_bundle_data_permission_items
WHERE bundle_id = ?
  AND model_id = ?;

-- name: ListBundleDataPermissionItems :many
SELECT *
FROM end_user_bundle_data_permission_items
WHERE bundle_id = ?
ORDER BY sort_order, created_at;

-- name: ClearBundleDataPermissionItems :exec
DELETE FROM end_user_bundle_data_permission_items
WHERE bundle_id = ?;

-- name: GetBundleDataPermissionItemByBundleAndModel :one
SELECT *
FROM end_user_bundle_data_permission_items
WHERE bundle_id = ?
  AND model_id = ?;
