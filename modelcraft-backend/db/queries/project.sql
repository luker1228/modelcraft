-- name: CreateProject :exec
INSERT INTO projects (org_name, slug, title, description, cluster_id, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetProjectBySlugAndOrg :one
SELECT * FROM projects
WHERE slug = ? AND org_name = ? AND `projects`.`deleted_at` = 0 LIMIT 1;

-- name: GetProjectByClusterID :one
SELECT * FROM projects
WHERE org_name = ? AND cluster_id = ? AND `projects`.`deleted_at` = 0 LIMIT 1;

-- name: ListProjects :many
SELECT * FROM projects WHERE `projects`.`deleted_at` = 0 ORDER BY created_at DESC;

-- name: ListProjectsByOrg :many
SELECT * FROM projects
WHERE org_name = ? AND `projects`.`deleted_at` = 0 ORDER BY created_at DESC;

-- name: UpdateProject :exec
UPDATE projects
SET title = ?, description = ?, cluster_id = ?, updated_at = NOW(3)
WHERE slug = ? AND org_name = ?;

-- name: ArchiveProject :exec
UPDATE projects
SET status = 'archived', updated_at = NOW(3)
WHERE slug = ? AND org_name = ?;

-- name: ExistsProjectBySlug :one
SELECT COUNT(*) FROM projects
WHERE slug = ? AND org_name = ? AND `projects`.`deleted_at` = 0 ;
