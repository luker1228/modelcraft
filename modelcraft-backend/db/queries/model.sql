-- name: CreateModel :exec
INSERT INTO models (id, org_name, project_slug, name, title, description, storage_type, database_name, version, status, group_id, deployment_status, last_sync_at, sync_error, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetModelByID :one
SELECT * FROM models WHERE id = ? LIMIT 1;

-- name: GetModelByName :one
SELECT * FROM models
WHERE database_name = ? AND name = ? AND project_slug = ?
LIMIT 1;

-- name: ListModels :many
SELECT * FROM models
WHERE project_slug = ?
  AND database_name = ?
  AND (? IS NULL OR name LIKE CONCAT('%', ?, '%'))
  AND (? IS NULL OR title LIKE CONCAT('%', ?, '%'))
  AND (? IS NULL OR status = ?)
  AND (? IS NULL OR storage_type = ?)
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountModels :one
SELECT COUNT(*) FROM models
WHERE project_slug = ?
  AND database_name = ?
  AND (? IS NULL OR name LIKE CONCAT('%', ?, '%'))
  AND (? IS NULL OR title LIKE CONCAT('%', ?, '%'))
  AND (? IS NULL OR status = ?)
  AND (? IS NULL OR storage_type = ?);

-- name: GetAllModels :many
SELECT * FROM models;

-- name: UpdateModel :execresult
UPDATE models
SET title = ?, description = ?, status = ?, group_id = ?, deployment_status = ?, version = ?, updated_at = NOW(3)
WHERE id = ?;

-- name: UpdateModelWithVersion :execresult
UPDATE models
SET title = ?, description = ?, status = ?, group_id = ?, deployment_status = ?, version = version + 1, updated_at = NOW(3)
WHERE id = ? AND version = ?;

-- name: UpdateModelDeploymentStatus :exec
UPDATE models
SET deployment_status = ?, last_sync_at = ?, sync_error = ?, updated_at = NOW(3)
WHERE id = ?;

-- name: DeleteModel :exec
DELETE FROM models WHERE id = ?;

-- name: FindModelsByDeploymentStatus :many
SELECT * FROM models
WHERE deployment_status IN (sqlc.slice('statuses'));

-- name: UpdateModelsGroupID :exec
UPDATE models
SET group_id = ?, updated_at = NOW(3)
WHERE group_id = ?;
