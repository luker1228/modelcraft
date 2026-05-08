-- name: CreateLogicalForeignKey :exec
INSERT INTO logical_foreign_keys (
  id, pair_id, org_name, direction,
  model_id, model_name, ref_model_id, ref_model_name,
  ref_database_name, ref_table_name,
  source_fields, target_fields, is_deletable,
  created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: DeleteLogicalForeignKeyByPairID :exec
UPDATE logical_foreign_keys SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED) WHERE (pair_id = ?
  AND org_name = ?) AND `logical_foreign_keys`.`deleted_at` = 0;

-- name: FindLogicalForeignKeysByModelID :many
SELECT id, pair_id, org_name, direction, model_id, model_name, ref_model_id, ref_model_name,
       ref_database_name, ref_table_name, source_fields, target_fields, is_deletable, created_at, updated_at
FROM logical_foreign_keys
WHERE model_id = ?
  AND org_name = ? AND `logical_foreign_keys`.`deleted_at` = 0 ;

-- name: FindLogicalForeignKeysByPairID :many
SELECT id, pair_id, org_name, direction, model_id, model_name, ref_model_id, ref_model_name,
       ref_database_name, ref_table_name, source_fields, target_fields, is_deletable, created_at, updated_at
FROM logical_foreign_keys
WHERE pair_id = ?
  AND org_name = ? AND `logical_foreign_keys`.`deleted_at` = 0 ;

-- name: FindLogicalForeignKeysByRefModelID :many
SELECT id, pair_id, org_name, direction, model_id, model_name, ref_model_id, ref_model_name,
       ref_database_name, ref_table_name, source_fields, target_fields, is_deletable, created_at, updated_at
FROM logical_foreign_keys
WHERE ref_model_id = ?
  AND org_name = ? AND `logical_foreign_keys`.`deleted_at` = 0 ;

-- name: GetLogicalForeignKeyByID :one
SELECT id, pair_id, org_name, direction, model_id, model_name, ref_model_id, ref_model_name,
       ref_database_name, ref_table_name, source_fields, target_fields, is_deletable, created_at, updated_at
FROM logical_foreign_keys
WHERE id = ? AND `logical_foreign_keys`.`deleted_at` = 0 LIMIT 1;

-- name: FindFieldsByBelongsToFKID :many
SELECT model_id, name, org_name, project_slug, model_name, database_name, enum_name, belongs_to_fk_id, relate_fk_id, title, description, format, non_null, required, is_unique, is_primary, status, validation, display_order, metadata, created_at, updated_at
FROM field_definitions
WHERE belongs_to_fk_id = ?
  AND org_name = ? AND `field_definitions`.`deleted_at` = 0 ;

-- name: FindFieldsByRelateFKID :many
SELECT model_id, name, org_name, project_slug, model_name, database_name, enum_name, belongs_to_fk_id, relate_fk_id, title, description, format, non_null, required, is_unique, is_primary, status, validation, display_order, metadata, created_at, updated_at
FROM field_definitions
WHERE relate_fk_id = ?
  AND org_name = ? AND `field_definitions`.`deleted_at` = 0 ;

-- name: BindBelongsToFKIDToFields :exec
UPDATE field_definitions
SET belongs_to_fk_id = ?, updated_at = NOW(3)
WHERE org_name = ?
  AND model_id = ?
  AND name IN (sqlc.slice('field_names'));
