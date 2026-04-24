-- name: GrantBundleToUser :exec
INSERT INTO end_user_user_bundles (
  id,
  user_id,
  org_name,
  project_slug,
  bundle_id
)
VALUES (?, ?, ?, ?, ?);

-- name: RevokeBundleFromUser :execresult
DELETE FROM end_user_user_bundles
WHERE user_id = ?
  AND bundle_id = ?
  AND org_name = ?
  AND project_slug = ?;

-- name: ListBundlesByUser :many
SELECT b.*
FROM end_user_permission_bundles b
  JOIN end_user_user_bundles ub ON b.id = ub.bundle_id
WHERE ub.user_id = ?
  AND ub.org_name = ?
  AND ub.project_slug = ?;
