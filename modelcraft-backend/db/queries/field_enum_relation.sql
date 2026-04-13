-- name: CreateFieldEnumRelation :exec
INSERT INTO field_enum_relations (
  id,
  model_id,
  label_field_name,
  source_field_name,
  org_name,
  project_slug,
  enum_name,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetFieldEnumRelationByID :one
SELECT * FROM field_enum_relations
WHERE org_name = ? AND id = ?
LIMIT 1;

-- name: GetFieldEnumRelationBySource :one
SELECT * FROM field_enum_relations
WHERE org_name = ? AND model_id = ? AND source_field_name = ?
LIMIT 1;

-- name: ListFieldEnumRelationsByModelID :many
SELECT * FROM field_enum_relations
WHERE org_name = ? AND model_id = ?
ORDER BY created_at ASC;

-- name: CountFieldEnumRelationsBySource :one
SELECT COUNT(*) FROM field_enum_relations
WHERE org_name = ? AND model_id = ? AND source_field_name = ?;

-- name: DeleteFieldEnumRelationByID :execresult
DELETE FROM field_enum_relations
WHERE org_name = ? AND id = ?;
