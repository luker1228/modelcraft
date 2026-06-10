-- name: CreateOrganization :exec
INSERT INTO organizations (name, display_name, owner_id, phone, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetOrganizationByName :one
SELECT * FROM organizations WHERE name = ? AND `organizations`.`deleted_at` = 0 LIMIT 1;

-- name: ListOrganizationsByUser :many
SELECT o.* FROM organizations o
INNER JOIN users u ON u.org_name = o.name
WHERE u.id = ? AND u.status = 'active' AND u.deleted_at = 0 AND `o`.`deleted_at` = 0 ORDER BY o.created_at DESC;

-- name: UpdateOrganization :exec
UPDATE organizations
SET display_name = ?, status = ?, updated_at = NOW(3)
WHERE name = ?;

-- name: ExistsOrganizationByName :one
SELECT COUNT(*) FROM organizations WHERE name = ? AND `organizations`.`deleted_at` = 0 ;

-- name: CreateUser :exec
INSERT INTO users (id, external_id, name, phone, password_hash, display_name, org_name, is_admin, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ? AND `users`.`deleted_at` = 0 LIMIT 1;

-- name: GetUserByExternalID :one
SELECT * FROM users WHERE external_id = ? AND `users`.`deleted_at` = 0 LIMIT 1;

-- name: ExistsUserByExternalID :one
SELECT COUNT(*) FROM users WHERE external_id = ? AND `users`.`deleted_at` = 0 ;

-- name: FindIDByExternalID :one
SELECT id FROM users WHERE external_id = ? AND `users`.`deleted_at` = 0 LIMIT 1;

-- name: GetOrganizationByPhone :one
SELECT * FROM organizations WHERE phone = ? AND `organizations`.`deleted_at` = 0 LIMIT 1;

-- name: ExistsOrganizationByPhone :one
SELECT COUNT(*) FROM organizations WHERE phone = ? AND `organizations`.`deleted_at` = 0;

-- name: UpdateUserStatus :exec
UPDATE users SET status = ?, updated_at = NOW(3)
WHERE id = ? AND deleted_at = 0;

-- name: UpdateUserAdmin :exec
UPDATE users SET is_admin = ?, updated_at = NOW(3)
WHERE id = ? AND deleted_at = 0;

-- name: ListUsersWithOrgDetails :many
SELECT u.id, u.org_name, u.is_admin, u.status, u.created_at,
       o.display_name AS org_display_name
FROM users u
INNER JOIN organizations o ON o.name = u.org_name
WHERE u.id = ? AND u.status = 'active' AND o.deleted_at = 0 AND u.deleted_at = 0
ORDER BY u.created_at DESC
LIMIT ?;

-- name: ListUsersByOrgWithName :many
SELECT id, name, org_name, is_admin, status, created_at, updated_at
FROM users
WHERE org_name = ? AND deleted_at = 0
ORDER BY created_at DESC;

-- name: CountUsersByOrg :one
SELECT COUNT(*) FROM users WHERE org_name = ? AND deleted_at = 0;
