-- name: InsertAPIKey :exec
INSERT INTO api_keys (id, user_id, name, key_hash, key_prefix, expires_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, NOW());

-- name: GetAPIKeyByHash :one
SELECT id, user_id, name, key_hash, key_prefix, last_used_at, expires_at, created_at, revoked_at
FROM api_keys
WHERE key_hash = ?
LIMIT 1;

-- name: ListAPIKeysByUserID :many
SELECT id, user_id, name, key_hash, key_prefix, last_used_at, expires_at, created_at, revoked_at
FROM api_keys
WHERE user_id = ? AND revoked_at IS NULL
ORDER BY created_at DESC;

-- name: CountActiveAPIKeysByUserID :one
SELECT COUNT(*) FROM api_keys
WHERE user_id = ? AND revoked_at IS NULL;

-- name: RevokeAPIKey :exec
UPDATE api_keys
SET revoked_at = NOW()
WHERE id = ? AND user_id = ?;

-- name: UpdateAPIKey :exec
UPDATE api_keys
SET name = ?, expires_at = ?
WHERE id = ? AND user_id = ?;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = ?;

-- name: GetAPIKeyByID :one
SELECT id, user_id, name, key_hash, key_prefix, last_used_at, expires_at, created_at, revoked_at
FROM api_keys
WHERE id = ?
LIMIT 1;

-- name: DeleteRevokedAPIKeys :exec
DELETE FROM api_keys
WHERE revoked_at IS NOT NULL AND revoked_at < DATE_SUB(NOW(), INTERVAL 90 DAY);
