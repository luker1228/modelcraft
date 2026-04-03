-- name: InsertRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
VALUES (?, ?, ?, ?, NOW());

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
FROM refresh_tokens
WHERE token_hash = ?
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE id = ?;

-- name: RevokeAllRefreshTokensByUserID :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = ? AND revoked_at IS NULL;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE (expires_at < DATE_SUB(NOW(), INTERVAL 30 DAY))
   OR (revoked_at IS NOT NULL AND revoked_at < DATE_SUB(NOW(), INTERVAL 30 DAY));
