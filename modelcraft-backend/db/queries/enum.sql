-- name: CreateEnumDefinition :exec
INSERT INTO model_enums (id, org_name, project_slug, name, display_name, description, options, is_multi_select, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetEnumByID :one
SELECT * FROM model_enums WHERE id = ? LIMIT 1;

-- name: GetEnumByName :one
SELECT * FROM model_enums
WHERE org_name = ? AND project_slug = ? AND name = ?
LIMIT 1;

-- name: ListEnums :many
SELECT * FROM model_enums
WHERE org_name = ? AND project_slug = ?
ORDER BY name ASC;

-- name: UpdateEnum :exec
UPDATE model_enums
SET display_name = ?, description = ?, options = ?, is_multi_select = ?, updated_at = NOW(3)
WHERE org_name = ? AND project_slug = ? AND name = ?;

-- name: DeleteEnum :exec
DELETE FROM model_enums
WHERE org_name = ? AND project_slug = ? AND name = ?;

-- name: ExistsEnumByName :one
SELECT COUNT(*) FROM model_enums
WHERE org_name = ? AND project_slug = ? AND name = ?;

-- name: GetEnumsByNames :many
SELECT * FROM model_enums
WHERE project_slug = ? AND name IN (sqlc.slice('names'));

-- name: CreateFieldEnumAssociation :exec
INSERT INTO model_field_enum_associations (model_id, field_name, org_name, project_slug, enum_name, database_name, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetFieldEnumAssociationByField :one
SELECT * FROM model_field_enum_associations
WHERE model_id = ? AND field_name = ?
LIMIT 1;

-- name: GetFieldEnumAssociationsByEnumName :many
SELECT * FROM model_field_enum_associations
WHERE project_slug = ? AND enum_name = ?;

-- name: GetFieldEnumAssociationsByModelID :many
SELECT * FROM model_field_enum_associations
WHERE model_id = ?;

-- name: DeleteFieldEnumAssociation :exec
DELETE FROM model_field_enum_associations
WHERE model_id = ? AND field_name = ?;

-- name: DeleteFieldEnumAssociationsByModelID :exec
DELETE FROM model_field_enum_associations WHERE model_id = ?;

-- name: GetEnumReferencesByName :many
SELECT fea.model_id, fea.field_name, fd.model_name
FROM model_field_enum_associations fea
INNER JOIN field_definitions fd ON fea.model_id = fd.model_id AND fea.field_name = fd.name
WHERE fea.project_slug = ? AND fea.enum_name = ?;
