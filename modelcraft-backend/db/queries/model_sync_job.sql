-- name: CreateModelSyncJob :exec
INSERT INTO model_sync_job (
  id, batch_id, database_id, org_name, project_slug, database_name, table_names,
  status, total_tables, processed_tables, created_models, synced_models,
  failed_count, failed_tables, started_at, finished_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetModelSyncJobByID :one
SELECT * FROM model_sync_job
WHERE id = ? AND org_name = ? AND project_slug = ?
LIMIT 1;

-- name: GetModelSyncJobsByIDs :many
SELECT * FROM model_sync_job
WHERE org_name = ? AND project_slug = ? AND id IN (/*SLICE:ids*/?)
ORDER BY created_at DESC;

-- name: GetModelSyncJobsByBatchID :many
SELECT * FROM model_sync_job
WHERE org_name = ? AND project_slug = ? AND batch_id = ?
ORDER BY created_at DESC;

-- name: GetActiveModelSyncJobByDatabase :one
SELECT * FROM model_sync_job
WHERE org_name = ?
  AND project_slug = ?
  AND database_id = ?
  AND status IN ('pending', 'running')
  AND updated_at > ?
ORDER BY created_at DESC
LIMIT 1;

-- name: FailStaleModelSyncJobs :exec
UPDATE model_sync_job
SET status = 'failed',
    finished_at = NOW(3),
    updated_at = NOW(3)
WHERE status IN ('pending', 'running')
  AND updated_at <= ?;

-- name: UpdateModelSyncJob :exec
UPDATE model_sync_job
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
WHERE id = ? AND org_name = ? AND project_slug = ?;
