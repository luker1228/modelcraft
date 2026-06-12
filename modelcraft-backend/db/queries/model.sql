-- name: CreateModel :exec
INSERT INTO models (id, org_name, project_slug, name, title, description, storage_type, database_name, display_field, version, status, group_id, deployment_status, last_sync_at, sync_error, is_read_only, created_via, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetModelByID :one
SELECT * FROM models WHERE id = ? AND `models`.`deleted_at` = 0 LIMIT 1;

-- name: GetModelByName :one
SELECT * FROM models
WHERE org_name = ? AND database_name = ? AND name = ? AND project_slug = ? AND `models`.`deleted_at` = 0 LIMIT 1;

-- name: ListModels :many
SELECT * FROM models
WHERE org_name = ?
  AND project_slug = ?
  AND database_name = ?
  AND (sqlc.arg(name_filter) IS NULL OR name LIKE CONCAT('%', sqlc.arg(name_search), '%'))
  AND (sqlc.arg(title_filter) IS NULL OR title LIKE CONCAT('%', sqlc.arg(title_search), '%'))
  AND (sqlc.arg(status_filter) IS NULL OR status = sqlc.arg(status))
  AND (sqlc.arg(storage_type_filter) IS NULL OR storage_type = sqlc.arg(storage_type)) AND `models`.`deleted_at` = 0 ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountModels :one
SELECT COUNT(*) FROM models
WHERE org_name = ?
  AND project_slug = ?
  AND database_name = ?
  AND (sqlc.arg(name_filter) IS NULL OR name LIKE CONCAT('%', sqlc.arg(name_search), '%'))
  AND (sqlc.arg(title_filter) IS NULL OR title LIKE CONCAT('%', sqlc.arg(title_search), '%'))
  AND (sqlc.arg(status_filter) IS NULL OR status = sqlc.arg(status))
  AND (sqlc.arg(storage_type_filter) IS NULL OR storage_type = sqlc.arg(storage_type)) AND `models`.`deleted_at` = 0 ;

-- name: ListModelDatabases :many
SELECT DISTINCT database_name
FROM models
WHERE org_name = ?
  AND project_slug = ?
  AND (sqlc.arg(search_filter) IS NULL OR database_name LIKE CONCAT('%', sqlc.arg(search), '%')) AND `models`.`deleted_at` = 0 ORDER BY database_name ASC
LIMIT ? OFFSET ?;

-- name: CountModelDatabases :one
SELECT COUNT(DISTINCT database_name)
FROM models
WHERE org_name = ?
  AND project_slug = ?
  AND (sqlc.arg(search_filter) IS NULL OR database_name LIKE CONCAT('%', sqlc.arg(search), '%')) AND `models`.`deleted_at` = 0 ;

-- name: GetAllModels :many
SELECT * FROM models WHERE `models`.`deleted_at` = 0 ;

-- name: UpdateModel :execresult
UPDATE models
SET title = ?, description = ?, display_field = ?, status = ?, group_id = ?, deployment_status = ?, version = ?, updated_at = NOW(3)
WHERE id = ?;

-- name: UpdateModelWithVersion :execresult
UPDATE models
SET title = ?, description = ?, display_field = ?, status = ?, group_id = ?, deployment_status = ?, version = version + 1, updated_at = NOW(3)
WHERE id = ? AND version = ?;

-- name: UpdateModelDeploymentStatus :exec
UPDATE models
SET deployment_status = ?, last_sync_at = ?, sync_error = ?, updated_at = NOW(3)
WHERE id = ?;

-- name: DeleteModel :exec
UPDATE models SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED) WHERE (id = ?) AND `models`.`deleted_at` = 0;

-- name: GetModelMetaByIDs :many
SELECT * FROM models
WHERE org_name = ?
  AND project_slug = ?
  AND id IN (sqlc.slice('ids'))
  AND `models`.`deleted_at` = 0;

-- name: FindModelsByDeploymentStatus :many
SELECT * FROM models
WHERE deployment_status IN (sqlc.slice('statuses'))
  AND `models`.`deleted_at` = 0;

-- name: UpdateModelsGroupID :exec
UPDATE models
SET group_id = ?, updated_at = NOW(3)
WHERE group_id = ?;

-- name: UpdateInsertionOrderField :exec
UPDATE models
SET insertion_order_field = sqlc.narg('insertion_order_field'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
  AND delete_token = 0;
