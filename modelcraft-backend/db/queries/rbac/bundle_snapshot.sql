-- name: InsertBundleSnapshot :exec
INSERT INTO end_user_permission_bundle_snapshots (
  id,
  bundle_id,
  version,
  permissions,
  created_by,
  restored_from
)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListBundleSnapshots :many
SELECT *
FROM end_user_permission_bundle_snapshots
WHERE bundle_id = ?
ORDER BY version DESC
LIMIT 5;

-- name: GetBundleCurrentVersion :one
SELECT COALESCE(MAX(version), 0) AS current_version
FROM end_user_permission_bundle_snapshots
WHERE bundle_id = ?;

-- name: DeleteOldBundleSnapshots :exec
DELETE snap
FROM end_user_permission_bundle_snapshots snap
LEFT JOIN (
  SELECT s.id AS keep_id
  FROM end_user_permission_bundle_snapshots s
  WHERE s.bundle_id = ?
  ORDER BY s.version DESC
  LIMIT 5
) AS latest ON snap.id = latest.keep_id
WHERE snap.bundle_id = ?
  AND latest.keep_id IS NULL;

-- name: GetBundleSnapshotByVersion :one
SELECT *
FROM end_user_permission_bundle_snapshots
WHERE bundle_id = ?
  AND version = ?;

-- name: ClearBundlePermissions :exec
DELETE FROM end_user_bundle_permissions
WHERE bundle_id = ?;
