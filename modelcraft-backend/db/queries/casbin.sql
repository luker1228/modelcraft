-- name: CreateRole :execresult
INSERT INTO roles (name, description, is_system, org_name, created_at, updated_at)
VALUES (?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = ? AND `roles`.`deleted_at` = 0 LIMIT 1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = ? AND `roles`.`deleted_at` = 0 LIMIT 1;

-- name: GetRoleByNameAndOrg :one
SELECT * FROM roles WHERE name = ? AND org_name = ? AND `roles`.`deleted_at` = 0 LIMIT 1;

-- name: GetSystemRoleByName :one
SELECT * FROM roles WHERE name = ? AND is_system = true AND `roles`.`deleted_at` = 0 LIMIT 1;

-- name: ListRoles :many
SELECT * FROM roles WHERE `roles`.`deleted_at` = 0 ORDER BY name ASC;

-- name: ListRolesByOrg :many
SELECT * FROM roles
WHERE org_name = ? AND `roles`.`deleted_at` = 0 ORDER BY name ASC;

-- name: ListRolesByOrgIncludeSystem :many
SELECT * FROM roles
WHERE org_name = ? OR org_name = '__SYSTEM__' AND `roles`.`deleted_at` = 0 ORDER BY name ASC;

-- name: UpdateRole :exec
UPDATE roles
SET description = ?, updated_at = NOW(3)
WHERE id = ?;

-- name: DeleteRole :exec
UPDATE roles SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED) WHERE (id = ?) AND `roles`.`deleted_at` = 0;

-- name: CreatePermission :exec
INSERT INTO role_permissions (role_id, org_name, obj, act, created_at)
VALUES (?, ?, ?, ?, NOW(3));

-- name: DeletePermission :exec
DELETE FROM role_permissions
WHERE role_id = ? AND obj = ? AND act = ?;

-- name: ListPermissionsByRole :many
SELECT * FROM role_permissions WHERE role_id = ?;

-- name: ListPermissionsByRoleAndOrg :many
SELECT * FROM role_permissions
WHERE role_id = ? AND org_name = ?;

-- name: DeletePermissionsByRole :exec
DELETE FROM role_permissions WHERE role_id = ?;

-- name: CreateUserRole :execresult
INSERT INTO user_roles (user_id, role_id, org_name, created_at)
VALUES (?, ?, ?, NOW(3));

-- name: DeleteUserRole :exec
DELETE FROM user_roles
WHERE user_id = ? AND role_id = ? AND org_name = ?;

-- name: ListUserRoles :many
SELECT * FROM user_roles
WHERE user_id = ? AND org_name = ?;

-- name: ListRoleUsers :many
SELECT * FROM user_roles
WHERE role_id = ? AND org_name = ?;

-- name: GetUserRole :one
SELECT * FROM user_roles
WHERE user_id = ? AND role_id = ? AND org_name = ?
LIMIT 1;

-- name: DeleteUserRolesByRole :exec
DELETE FROM user_roles WHERE role_id = ?;
