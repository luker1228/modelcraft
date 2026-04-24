-- name: AssignRoleToUser :exec
INSERT INTO end_user_role_users (
  id,
  user_id,
  role_id,
  org_name,
  project_slug
)
VALUES (?, ?, ?, ?, ?);

-- name: RevokeRoleFromUser :execresult
DELETE FROM end_user_role_users
WHERE user_id = ?
  AND role_id = ?
  AND org_name = ?
  AND project_slug = ?;

-- name: ListRolesByUser :many
SELECT role_id
FROM end_user_role_users
WHERE user_id = ?
  AND org_name = ?
  AND project_slug = ?;
