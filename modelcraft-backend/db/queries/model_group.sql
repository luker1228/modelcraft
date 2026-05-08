-- name: CreateModelGroup :exec
INSERT INTO model_groups (id, org_name, project_slug, name, display_order, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetModelGroupByID :one
SELECT * FROM model_groups WHERE id = ? AND `model_groups`.`deleted_at` = 0 LIMIT 1;

-- name: GetModelGroupByName :one
SELECT * FROM model_groups
WHERE org_name = ? AND project_slug = ? AND name = ? AND `model_groups`.`deleted_at` = 0 LIMIT 1;

-- name: ListModelGroupsByProject :many
SELECT * FROM model_groups
WHERE org_name = ? AND project_slug = ? AND `model_groups`.`deleted_at` = 0 ORDER BY display_order ASC;

-- name: UpdateModelGroup :exec
UPDATE model_groups
SET name = ?, display_order = ?, updated_at = NOW(3)
WHERE id = ?;

-- name: DeleteModelGroup :exec
UPDATE model_groups SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED) WHERE (id = ?) AND `model_groups`.`deleted_at` = 0;

-- name: GetTailModelGroupDisplayOrder :one
SELECT display_order FROM model_groups
WHERE org_name = ? AND project_slug = ? AND `model_groups`.`deleted_at` = 0 ORDER BY display_order DESC
LIMIT 1;
