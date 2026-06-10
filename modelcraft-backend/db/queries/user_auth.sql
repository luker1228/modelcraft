-- name: GetUserByPhoneInOrg :one
SELECT id, phone, password_hash, name, external_id, org_name, created_at, updated_at
FROM users
WHERE org_name = ? AND phone = ? AND `users`.`deleted_at` = 0 LIMIT 1;

-- name: GetUserByNameInOrg :one
SELECT id, phone, password_hash, name, external_id, org_name, created_at, updated_at
FROM users
WHERE org_name = ? AND name = ? AND `users`.`deleted_at` = 0 LIMIT 1;

-- name: ExistsByPhoneInOrg :one
SELECT EXISTS(SELECT 1 FROM users WHERE org_name = ? AND phone = ?) AS phone_exists;

-- name: ExistsByUserNameInOrg :one
SELECT EXISTS(SELECT 1 FROM users WHERE org_name = ? AND name = ?) AS name_exists;
