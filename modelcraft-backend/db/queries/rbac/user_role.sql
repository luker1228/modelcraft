-- name: AssignRoleToUser :exec
INSERT INTO end_user_role_users (
  id,
  user_id,
  role_id,
  org_name
)
VALUES (?, ?, ?, ?);

-- name: RevokeRoleFromUser :execresult
DELETE FROM end_user_role_users
WHERE user_id = ?
  AND role_id = ?
  AND org_name = ?;

-- name: ListRolesByUser :many
SELECT role_id
FROM end_user_role_users
WHERE user_id = ?
  AND org_name = ?;

-- name: ListProjectEndUserRoleUsersCount :one
-- 统计 Project 下有角色分配的用户总数（支持用户名搜索和角色过滤）
SELECT COUNT(1)
FROM end_user_role_users ur
JOIN end_user_roles r
  ON r.id = ur.role_id
  AND r.org_name = ur.org_name
JOIN end_user_users u
  ON u.id = ur.user_id
  AND u.org_name = ur.org_name
WHERE r.org_name = ?
  AND r.project_slug = ?
  AND (sqlc.arg(search_filter) = '' OR u.username LIKE CONCAT('%', sqlc.arg(search), '%'))
  AND (sqlc.arg(role_id_filter) = '' OR ur.role_id = sqlc.arg(role_id)) AND `r`.`deleted_at` = 0 ;

-- name: ListProjectEndUserRoleUsers :many
-- 分页查询 Project 下有角色分配的用户列表（支持用户名搜索和角色过滤）
SELECT
  ur.id,
  ur.org_name,
  ur.created_at,
  u.id        AS user_id,
  u.username,
  u.is_forbidden,
  u.created_by,
  u.created_at AS user_created_at,
  u.updated_at AS user_updated_at,
  r.id        AS role_id,
  r.org_name  AS role_org_name,
  r.project_slug,
  r.name      AS role_name,
  r.description AS role_description,
  r.is_implicit,
  r.created_at AS role_created_at,
  r.updated_at AS role_updated_at
FROM end_user_role_users ur
JOIN end_user_roles r
  ON r.id = ur.role_id
  AND r.org_name = ur.org_name
JOIN end_user_users u
  ON u.id = ur.user_id
  AND u.org_name = ur.org_name
WHERE r.org_name = ?
  AND r.project_slug = ?
  AND (sqlc.arg(search_filter) = '' OR u.username LIKE CONCAT('%', sqlc.arg(search), '%'))
  AND (sqlc.arg(role_id_filter) = '' OR ur.role_id = sqlc.arg(role_id))
  AND (sqlc.arg(after_filter) = '' OR ur.id > sqlc.arg(after)) AND `r`.`deleted_at` = 0 ORDER BY ur.id ASC
LIMIT ?;
