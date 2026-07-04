-- name: CreateModelDatabase :exec
INSERT INTO model_database (id, org_name, project_slug, cluster_id, name, title, description, mode, latest_sync_job_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

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

-- name: ListModelDatabaseCatalog :many
SELECT name FROM model_database
WHERE org_name = ?
  AND project_slug = ?
  AND (sqlc.arg(search_filter) IS NULL OR name LIKE CONCAT('%', sqlc.arg(search), '%'))
  AND `model_database`.`deleted_at` = 0
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: CountModelDatabaseCatalog :one
SELECT COUNT(*) FROM model_database
WHERE org_name = ?
  AND project_slug = ?
  AND (sqlc.arg(search_filter) IS NULL OR name LIKE CONCAT('%', sqlc.arg(search), '%'))
  AND `model_database`.`deleted_at` = 0;

-- name: UpdateModelDatabase :exec
UPDATE model_database
SET title = ?, description = ?, mode = ?, latest_sync_job_id = ?, updated_at = NOW(3)
WHERE id = ? AND org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` = 0;

-- name: UpdateModelDatabaseLatestSyncJob :exec
UPDATE model_database
SET latest_sync_job_id = ?, updated_at = NOW(3)
WHERE id = ? AND org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` = 0;

-- name: DeleteModelDatabase :exec
UPDATE model_database
SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED)
WHERE id = ? AND org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` = 0;

-- name: CreateModelDatabaseSyncJob :exec
INSERT INTO model_database_sync_job (
  id,
  org_name,
  project_slug,
  database_id,
  status,
  total_tables,
  processed_tables,
  created_models,
  synced_models,
  failed_count,
  failed_tables,
  deleted_at,
  delete_token,
  started_at,
  finished_at,
  created_at,
  updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?, NOW(3), NOW(3));

-- name: GetModelDatabaseSyncJobByID :one
SELECT * FROM model_database_sync_job
WHERE id = ? AND org_name = ? AND project_slug = ? AND `model_database_sync_job`.`deleted_at` = 0
LIMIT 1;

-- name: GetActiveModelDatabaseSyncJobByDatabase :one
SELECT * FROM model_database_sync_job
WHERE org_name = ?
  AND project_slug = ?
  AND database_id = ?
  AND `model_database_sync_job`.`deleted_at` = 0
  AND status IN ('pending', 'running')
  AND updated_at > ?
ORDER BY created_at DESC
LIMIT 1;

-- name: FailStaleSyncJobs :exec
UPDATE model_database_sync_job
SET status = 'failed',
    finished_at = NOW(3),
    updated_at = NOW(3)
WHERE `model_database_sync_job`.`deleted_at` = 0
  AND status IN ('pending', 'running')
  AND updated_at <= ?;

-- name: UpdateModelDatabaseSyncJob :exec
UPDATE model_database_sync_job
SET status = ?,
    total_tables = ?,
    processed_tables = ?,
    created_models = ?,
    synced_models = ?,
    failed_count = ?,
    failed_tables = ?,
    started_at = ?,
    finished_at = ?,
    updated_at = NOW(3)
WHERE id = ? AND `model_database_sync_job`.`deleted_at` = 0;
