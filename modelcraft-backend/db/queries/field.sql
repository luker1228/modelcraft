-- name: CreateFieldDefinition :exec
INSERT INTO field_definitions (model_id, org_name, project_slug, model_name, database_name, name, enum_name, title, description, format, non_null, required, is_unique, is_primary, is_deprecated, status, validation, display_order, metadata, relate_fk_id, belongs_to_fk_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetFieldByModelIDAndName :one
SELECT * FROM field_definitions
WHERE model_id = ? AND name = ? AND `field_definitions`.`deleted_at` = 0 LIMIT 1;

-- name: GetFieldsByModelID :many
SELECT * FROM field_definitions
WHERE model_id = ? AND `field_definitions`.`deleted_at` = 0 ORDER BY display_order ASC;

-- name: CountFieldsByModelID :one
SELECT COUNT(*) FROM field_definitions WHERE model_id = ? AND `field_definitions`.`deleted_at` = 0 ;

-- name: ExistsFieldByName :one
SELECT COUNT(*) FROM field_definitions WHERE model_id = ? AND name = ? AND `field_definitions`.`deleted_at` = 0 ;

-- name: UpdateField :execresult
UPDATE field_definitions
SET title = ?, description = ?, non_null = ?, required = ?, is_unique = ?, is_primary = ?, is_deprecated = ?, status = ?, validation = ?, display_order = ?, metadata = ?, updated_at = NOW(3)
WHERE model_id = ? AND name = ?;

-- name: UpdateFieldDisplayOrder :exec
UPDATE field_definitions
SET display_order = ?, updated_at = NOW(3)
WHERE model_id = ? AND name = ?;

-- name: UpdateFieldsStatus :exec
UPDATE field_definitions
SET status = ?, updated_at = NOW(3)
WHERE model_id = ? AND name IN (sqlc.slice('names'));

-- name: DeleteFieldsByModelID :exec
UPDATE field_definitions SET `deleted_at` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED), `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED) WHERE (model_id = ?) AND `field_definitions`.`deleted_at` = 0;

-- name: DeleteFieldsByNames :execresult
UPDATE field_definitions
SET deleted_at = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    delete_token = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED)
WHERE model_id = ?
  AND name IN (sqlc.slice('names'))
  AND `field_definitions`.`deleted_at` = 0;

-- name: GetTailFieldDisplayOrder :one
SELECT display_order FROM field_definitions
WHERE model_id = ? AND `field_definitions`.`deleted_at` = 0 ORDER BY display_order DESC
LIMIT 1;
