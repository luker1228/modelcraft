-- name: CreateLogicalForeignKey :exec
INSERT INTO logical_foreign_keys (id, pair_id, org_name, direction, model_id, model_name, ref_model_id, ref_model_name, source_fields, target_fields, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: DeleteLogicalForeignKeyByPairID :exec
DELETE FROM logical_foreign_keys
WHERE pair_id = ?
  AND org_name = ?;

-- name: FindLogicalForeignKeysByModelID :many
SELECT id, pair_id, org_name, direction, model_id, model_name, ref_model_id, ref_model_name, source_fields, target_fields, created_at, updated_at
FROM logical_foreign_keys
WHERE model_id = ?
  AND org_name = ?;

-- name: FindLogicalForeignKeysByPairID :many
SELECT id, pair_id, org_name, direction, model_id, model_name, ref_model_id, ref_model_name, source_fields, target_fields, created_at, updated_at
FROM logical_foreign_keys
WHERE pair_id = ?
  AND org_name = ?;

-- name: FindLogicalForeignKeysByRefModelID :many
SELECT id, pair_id, org_name, direction, model_id, model_name, ref_model_id, ref_model_name, source_fields, target_fields, created_at, updated_at
FROM logical_foreign_keys
WHERE ref_model_id = ?
  AND org_name = ?;

-- name: GetLogicalForeignKeyByID :one
SELECT id, pair_id, org_name, direction, model_id, model_name, ref_model_id, ref_model_name, source_fields, target_fields, created_at, updated_at
FROM logical_foreign_keys
WHERE id = ?
LIMIT 1;

-- name: FindFieldsByBelongsToFKID :many
SELECT model_id, name, org_name, project_slug, model_name, database_name, enum_name, enum_relation_id, belongs_to_fk_id, relate_fk_id, title, description, format, non_null, required, is_unique, is_primary, status, validation, display_order, metadata, created_at, updated_at
FROM field_definitions
WHERE belongs_to_fk_id = ?
  AND org_name = ?;

-- name: FindFieldsByRelateFKID :many
SELECT model_id, name, org_name, project_slug, model_name, database_name, enum_name, enum_relation_id, belongs_to_fk_id, relate_fk_id, title, description, format, non_null, required, is_unique, is_primary, status, validation, display_order, metadata, created_at, updated_at
FROM field_definitions
WHERE relate_fk_id = ?
  AND org_name = ?;

-- name: BindBelongsToFKIDToFields :exec
UPDATE field_definitions
SET belongs_to_fk_id = ?, updated_at = NOW(3)
WHERE org_name = ?
  AND model_id = ?
  AND name IN (sqlc.slice('field_names'));
