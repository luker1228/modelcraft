-- name: ListMembershipsWithOrgDetails :many
SELECT m.id, o.display_name
FROM user_organizations m
INNER JOIN organizations o ON m.org_name = o.name
WHERE m.user_id = ?;
