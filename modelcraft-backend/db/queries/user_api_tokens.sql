-- name: InsertAPIToken :exec
INSERT INTO user_api_tokens (id, org_name, end_user_id, name, token_hash, expires_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetAPITokenByHash :one
SELECT id, org_name, end_user_id, name, token_hash,
       expires_at, last_used_at, created_at, deleted_at, delete_token
FROM user_api_tokens
WHERE token_hash = ? AND deleted_at = 0
LIMIT 1;

-- name: ListAPITokensByUser :many
SELECT id, org_name, end_user_id, name, token_hash,
       expires_at, last_used_at, created_at, deleted_at, delete_token
FROM user_api_tokens
WHERE org_name = ? AND end_user_id = ? AND deleted_at = 0
ORDER BY created_at DESC;

-- name: SoftDeleteAPIToken :execrows
UPDATE user_api_tokens
SET deleted_at = ?, delete_token = ?
WHERE id = ? AND org_name = ? AND end_user_id = ? AND deleted_at = 0;

-- name: UpdateAPITokenLastUsed :exec
UPDATE user_api_tokens
SET last_used_at = ?
WHERE id = ?;
