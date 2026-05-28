-- name: CreateOrganization :exec
INSERT INTO organizations (name, display_name, owner_id, status, created_at, updated_at)
VALUES (?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetOrganizationByName :one
SELECT * FROM organizations WHERE name = ? AND `organizations`.`deleted_at` = 0 LIMIT 1;

-- name: ListOrganizationsByUser :many
SELECT o.* FROM organizations o
INNER JOIN user_orgs m ON o.name = m.org_name
WHERE m.user_id = ? AND m.status = 'active' AND `o`.`deleted_at` = 0 ORDER BY o.created_at DESC;

-- name: UpdateOrganization :exec
UPDATE organizations
SET display_name = ?, status = ?, updated_at = NOW(3)
WHERE name = ?;

-- name: ExistsOrganizationByName :one
SELECT COUNT(*) FROM organizations WHERE name = ? AND `organizations`.`deleted_at` = 0 ;

-- name: CreateUser :exec
INSERT INTO users (id, external_id, name, phone, password_hash, display_name, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ? AND `users`.`deleted_at` = 0 LIMIT 1;

-- name: GetUserByExternalID :one
SELECT * FROM users WHERE external_id = ? AND `users`.`deleted_at` = 0 LIMIT 1;

-- name: ExistsUserByExternalID :one
SELECT COUNT(*) FROM users WHERE external_id = ? AND `users`.`deleted_at` = 0 ;

-- name: FindIDByExternalID :one
SELECT id FROM users WHERE external_id = ? AND `users`.`deleted_at` = 0 LIMIT 1;

-- name: CreateMembership :exec
INSERT INTO user_orgs (id, user_id, org_name, status, created_at, updated_at)
VALUES (?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetMembershipByID :one
SELECT * FROM user_orgs WHERE id = ? AND `user_orgs`.`deleted_at` = 0 LIMIT 1;

-- name: GetMembershipByUserAndOrg :one
SELECT * FROM user_orgs
WHERE user_id = ? AND org_name = ? AND `user_orgs`.`deleted_at` = 0
LIMIT 1;

-- name: ListMembershipsByOrg :many
SELECT * FROM user_orgs
WHERE org_name = ? AND `user_orgs`.`deleted_at` = 0
ORDER BY created_at DESC;

-- name: ListMembershipsWithUserName :many
SELECT m.id, m.user_id, m.org_name, m.status, m.created_at, m.updated_at,
       COALESCE(u.name, '') AS user_name
FROM user_orgs m
LEFT JOIN users u ON m.user_id = u.id
WHERE m.org_name = ? AND `m`.`deleted_at` = 0 AND `u`.`deleted_at` = 0 ORDER BY m.created_at DESC;

-- name: ListMembershipsByUser :many
SELECT * FROM user_orgs
WHERE user_id = ? AND `user_orgs`.`deleted_at` = 0
ORDER BY created_at DESC;

-- name: CountMembershipsByUser :one
SELECT COUNT(*) FROM user_orgs WHERE user_id = ? AND `user_orgs`.`deleted_at` = 0;

-- name: ListMembershipsWithOrgDetails :many
SELECT m.id, m.user_id, m.org_name, m.is_admin, m.status, m.created_at, m.updated_at,
       o.display_name AS org_display_name
FROM user_orgs m
INNER JOIN organizations o ON m.org_name = o.name
WHERE m.user_id = ? AND m.status = 'active' AND `o`.`deleted_at` = 0 AND `m`.`deleted_at` = 0 ORDER BY m.created_at DESC
LIMIT ?;

-- name: UpdateMembership :exec
UPDATE user_orgs
SET status = ?, updated_at = NOW(3)
WHERE id = ?;

-- name: DeleteMembership :exec
UPDATE user_orgs SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED) WHERE id = ? AND `user_orgs`.`deleted_at` = 0;

-- name: GetUserOrgByUserID :one
SELECT id, user_id, org_name, is_admin, status, deleted_at, delete_token, created_at, updated_at
FROM user_orgs
WHERE user_id = ? AND deleted_at = 0
LIMIT 1;

-- name: CreateUserOrg :exec
INSERT INTO user_orgs (id, user_id, org_name, is_admin, status, deleted_at, delete_token, created_at, updated_at)
VALUES (?, ?, ?, ?, 'active', 0, 0, NOW(3), NOW(3));

-- name: UpdateUserOrgAdmin :exec
UPDATE user_orgs
SET is_admin = ?, updated_at = CURRENT_TIMESTAMP(3)
WHERE user_id = ? AND org_name = ? AND deleted_at = 0;
