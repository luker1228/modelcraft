-- name: ListModels :many
SELECT * FROM models
WHERE org_name = ?
ORDER BY created_at DESC;
