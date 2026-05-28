-- name: GetBuiltinEndUserByOrg :one
-- 查询 Org 内置 admin 用户（is_builtin=true），用于 Project 创建时自动分配 admin 角色
SELECT id
FROM end_user_users
WHERE org_name = ?
  AND is_builtin = 1
  AND deleted_at = 0
LIMIT 1;

-- name: AssignRoleToUser :exec
INSERT INTO project_role_users (
  id,
  user_id,
  role_id,
  org_name
)
VALUES (?, ?, ?, ?);

-- name: RevokeRoleFromUser :execresult
DELETE FROM project_role_users
WHERE user_id = ?
  AND role_id = ?
  AND org_name = ?;

-- name: IsEndUserBuiltin :one
-- 检查指定用户是否为 Org 内置 admin（is_builtin=true）
SELECT is_builtin
FROM end_user_users
WHERE id = ?
  AND org_name = ?
  AND deleted_at = 0
LIMIT 1;

-- name: ListRolesByUser :many
SELECT role_id
FROM project_role_users
WHERE user_id = ?
  AND org_name = ?;

-- name: ListProjectEndUserRoleUsersCount :one
-- 统计 Project 下有角色分配的用户总数（支持用户名搜索和角色过滤）
SELECT COUNT(1)
FROM project_role_users ur
JOIN project_roles r
  ON r.id = ur.role_id
  AND r.org_name = ur.org_name
JOIN users u
  ON u.id = ur.user_id
WHERE r.org_name = ?
  AND r.project_slug = ?
  AND (sqlc.arg(search_filter) = '' OR u.name LIKE CONCAT('%', sqlc.arg(search), '%'))
  AND (sqlc.arg(role_id_filter) = '' OR ur.role_id = sqlc.arg(role_id)) AND `r`.`deleted_at` = 0 ;

-- name: ListProjectEndUserRoleUsers :many
-- 分页查询 Project 下有角色分配的用户列表（支持用户名搜索和角色过滤）
SELECT
  ur.id,
  ur.org_name,
  ur.created_at,
  u.id        AS user_id,
  u.name,
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
FROM project_role_users ur
JOIN project_roles r
  ON r.id = ur.role_id
  AND r.org_name = ur.org_name
JOIN users u
  ON u.id = ur.user_id
WHERE r.org_name = ?
  AND r.project_slug = ?
  AND (sqlc.arg(search_filter) = '' OR u.name LIKE CONCAT('%', sqlc.arg(search), '%'))
  AND (sqlc.arg(role_id_filter) = '' OR ur.role_id = sqlc.arg(role_id))
  AND (sqlc.arg(after_filter) = '' OR ur.id > sqlc.arg(after)) AND `r`.`deleted_at` = 0 ORDER BY ur.id ASC
LIMIT ?;
