-- name: CreateDatabaseCluster :exec
INSERT INTO database_clusters (id, org_name, project_slug, title, description, host, port, username, password, connection_timeout, status, version, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetDatabaseClusterByID :one
SELECT * FROM database_clusters
WHERE id = ? AND org_name = ? AND `database_clusters`.`deleted_at` = 0 LIMIT 1;

-- name: GetDatabaseClusterByProjectKey :one
SELECT * FROM database_clusters
WHERE org_name = ? AND project_slug = ? AND `database_clusters`.`deleted_at` = 0 LIMIT 1;

-- name: ListDatabaseClusters :many
SELECT * FROM database_clusters
WHERE org_name = ? AND project_slug = ? AND `database_clusters`.`deleted_at` = 0 ;

-- name: UpdateDatabaseClusterWithVersion :execresult
UPDATE database_clusters
SET title = ?, description = ?, host = ?, port = ?, username = ?, password = ?, connection_timeout = ?, status = ?, version = version + 1, updated_at = NOW(3)
WHERE id = ? AND org_name = ? AND project_slug = ? AND version = ?;

-- name: DeleteDatabaseCluster :exec
UPDATE database_clusters SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED) WHERE (id = ? AND org_name = ? AND project_slug = ?) AND `database_clusters`.`deleted_at` = 0;

-- name: ExistsDatabaseClusterByProjectKey :one
SELECT COUNT(*) FROM database_clusters
WHERE org_name = ? AND project_slug = ? AND `database_clusters`.`deleted_at` = 0 ;

-- name: ListDatabaseClustersUpdatedAfter :many
-- 全局扫描：不按租户过滤，供连接池同步使用
SELECT * FROM database_clusters
WHERE updated_at > ? AND `database_clusters`.`deleted_at` = 0 ;
