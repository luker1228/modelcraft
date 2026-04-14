-- name: CreateFieldDefinition :exec
INSERT INTO field_definitions (model_id, org_name, project_slug, model_name, database_name, name, enum_name, enum_relation_id, title, description, format, non_null, required, is_unique, is_primary, is_deprecated, status, validation, display_order, metadata, relate_fk_id, belongs_to_fk_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetFieldByModelIDAndName :one
SELECT * FROM field_definitions
WHERE model_id = ? AND name = ?
LIMIT 1;

-- name: GetFieldsByModelID :many
SELECT * FROM field_definitions
WHERE model_id = ?
ORDER BY display_order ASC;

-- name: CountFieldsByModelID :one
SELECT COUNT(*) FROM field_definitions WHERE model_id = ?;

-- name: ExistsFieldByName :one
SELECT COUNT(*) FROM field_definitions WHERE model_id = ? AND name = ?;

-- name: UpdateField :execresult
UPDATE field_definitions
SET title = ?, description = ?, non_null = ?, required = ?, is_unique = ?, is_primary = ?, is_deprecated = ?, status = ?, validation = ?, display_order = ?, metadata = ?, enum_relation_id = ?, updated_at = NOW(3)
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
DELETE FROM field_definitions WHERE model_id = ?;

-- name: DeleteFieldsByNames :execresult
DELETE FROM field_definitions
WHERE model_id = ? AND name IN (sqlc.slice('names'));

-- name: GetTailFieldDisplayOrder :one
SELECT display_order FROM field_definitions
WHERE model_id = ?
ORDER BY display_order DESC
LIMIT 1;

-- name: CountFieldsByEnumRelationID :one
SELECT COUNT(*) FROM field_definitions
WHERE org_name = ? AND enum_relation_id = ?;
