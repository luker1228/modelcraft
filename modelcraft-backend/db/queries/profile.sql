-- name: CreateInitialProfile :exec
INSERT INTO profile (id, user_id, nickname, avatar_url, bio, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetProfileByUserID :one
SELECT p.id, p.user_id, p.nickname, p.avatar_url, p.bio, p.created_at, p.updated_at
FROM profile p
INNER JOIN user_organizations uo ON uo.user_id = p.user_id
WHERE p.user_id = ? AND uo.org_name = ? AND `p`.`deleted_at` = 0 LIMIT 1;

-- name: UpdateProfileByUserID :execresult
UPDATE profile p
INNER JOIN user_organizations uo ON uo.user_id = p.user_id
SET p.nickname = ?,
    p.avatar_url = ?,
    p.bio = ?,
    p.updated_at = NOW(3)
WHERE p.user_id = ?
  AND uo.org_name = ?;
