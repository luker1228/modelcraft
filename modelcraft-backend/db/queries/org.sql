-- name: CreateOrganization :exec
INSERT INTO organizations (name, display_name, owner_id, status, created_at, updated_at)
VALUES (?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetOrganizationByName :one
SELECT * FROM organizations WHERE name = ? LIMIT 1;

-- name: ListOrganizationsByUser :many
SELECT o.* FROM organizations o
INNER JOIN user_organizations m ON o.name = m.org_name
WHERE m.user_id = ? AND m.status = 'active'
ORDER BY o.created_at DESC;

-- name: UpdateOrganization :exec
UPDATE organizations
SET display_name = ?, status = ?, updated_at = NOW(3)
WHERE name = ?;

-- name: ExistsOrganizationByName :one
SELECT COUNT(*) FROM organizations WHERE name = ?;

-- name: CreateUser :exec
INSERT INTO users (id, external_id, name, phone, display_name, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: GetUserByExternalID :one
SELECT * FROM users WHERE external_id = ? LIMIT 1;

-- name: ExistsUserByExternalID :one
SELECT COUNT(*) FROM users WHERE external_id = ?;

-- name: FindIDByExternalID :one
SELECT id FROM users WHERE external_id = ? LIMIT 1;

-- name: CreateMembership :exec
INSERT INTO user_organizations (id, user_id, org_name, status, invited_by, invited_at, joined_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetMembershipByID :one
SELECT * FROM user_organizations WHERE id = ? LIMIT 1;

-- name: GetMembershipByUserAndOrg :one
SELECT * FROM user_organizations
WHERE user_id = ? AND org_name = ?
LIMIT 1;

-- name: ListMembershipsByOrg :many
SELECT * FROM user_organizations
WHERE org_name = ?
ORDER BY created_at DESC;

-- name: ListMembershipsWithUserName :many
SELECT m.id, m.user_id, m.org_name, m.status, m.invited_by, m.invited_at, m.joined_at, m.created_at, m.updated_at,
       COALESCE(u.name, '') AS user_name
FROM user_organizations m
LEFT JOIN users u ON m.user_id = u.id
WHERE m.org_name = ?
ORDER BY m.created_at DESC;

-- name: ListMembershipsByUser :many
SELECT * FROM user_organizations
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: CountMembershipsByUser :one
SELECT COUNT(*) FROM user_organizations WHERE user_id = ?;

-- name: ListMembershipsWithOrgDetails :many
SELECT m.id, m.user_id, m.org_name, m.status, m.invited_by, m.invited_at, m.joined_at, m.created_at, m.updated_at,
       o.display_name AS org_display_name
FROM user_organizations m
INNER JOIN organizations o ON m.org_name = o.name
WHERE m.user_id = ? AND m.status = 'active'
ORDER BY m.joined_at DESC
LIMIT ?;

-- name: UpdateMembership :exec
UPDATE user_organizations
SET status = ?, invited_by = ?, invited_at = ?, joined_at = ?, updated_at = NOW(3)
WHERE id = ?;

-- name: DeleteMembership :exec
DELETE FROM user_organizations WHERE id = ?;
