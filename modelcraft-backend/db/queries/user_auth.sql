-- name: GetUserByPhone :one
SELECT id, phone, password_hash, name, external_id, created_at, updated_at
FROM users
WHERE phone = ? LIMIT 1;

-- name: ExistsByPhone :one
SELECT EXISTS(SELECT 1 FROM users WHERE phone = ?) AS phone_exists;
