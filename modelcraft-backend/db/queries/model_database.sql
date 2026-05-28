-- name: CreateModelDatabase :exec
INSERT INTO model_database (id, org_name, project_slug, cluster_id, name, title, description, mode, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetModelDatabaseByID :one
SELECT * FROM model_database
WHERE id = ? AND org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` = 0
LIMIT 1;

-- name: GetModelDatabaseByName :one
SELECT * FROM model_database
WHERE org_name = ? AND project_slug = ? AND name = ? AND `model_database`.`deleted_at` = 0
LIMIT 1;

-- name: ListModelDatabasesByProject :many
SELECT * FROM model_database
WHERE org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` = 0
ORDER BY created_at ASC;

-- name: UpdateModelDatabase :exec
UPDATE model_database
SET title = ?, description = ?, mode = ?, updated_at = NOW(3)
WHERE id = ? AND org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` = 0;

-- name: DeleteModelDatabase :exec
UPDATE model_database
SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED)
WHERE id = ? AND org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` = 0;
