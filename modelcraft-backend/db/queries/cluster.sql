-- name: CreateDatabaseCluster :exec
INSERT INTO database_clusters (id, org_name, project_slug, title, description, host, port, username, password, connection_timeout, status, version, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetDatabaseClusterByID :one
SELECT * FROM database_clusters
WHERE id = ? AND org_name = ?
LIMIT 1;

-- name: GetDatabaseClusterByProjectKey :one
SELECT * FROM database_clusters
WHERE org_name = ? AND project_slug = ?
LIMIT 1;

-- name: ListDatabaseClusters :many
SELECT * FROM database_clusters
WHERE org_name = ? AND project_slug = ?
  AND (? IS NULL OR status = ?);

-- name: UpdateDatabaseClusterWithVersion :execresult
UPDATE database_clusters
SET title = ?, description = ?, host = ?, port = ?, username = ?, password = ?, connection_timeout = ?, status = ?, version = version + 1, updated_at = NOW(3)
WHERE id = ? AND org_name = ? AND project_slug = ? AND version = ?;

-- name: DeleteDatabaseCluster :exec
DELETE FROM database_clusters
WHERE id = ? AND org_name = ? AND project_slug = ?;

-- name: ExistsDatabaseClusterByProjectKey :one
SELECT COUNT(*) FROM database_clusters
WHERE org_name = ? AND project_slug = ?;

-- name: ListDatabaseClustersUpdatedAfter :many
SELECT * FROM database_clusters
WHERE updated_at > ?
  AND (? IS NULL OR org_name = ?)
  AND (? IS NULL OR project_slug = ?);
