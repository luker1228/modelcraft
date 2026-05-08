-- @include_deleted
-- name: ListAllModels :many
SELECT * FROM models ORDER BY created_at DESC;
