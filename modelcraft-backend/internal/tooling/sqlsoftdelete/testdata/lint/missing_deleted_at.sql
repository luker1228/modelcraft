-- name: GetModelByID :one
SELECT * FROM models WHERE id = ? LIMIT 1;
