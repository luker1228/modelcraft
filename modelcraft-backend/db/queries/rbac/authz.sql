-- name: GetBundleIDsByUserDirect :many
-- ⚡ 鉴权链 Step 1: 用户直接关联的权限包 ID 列表
SELECT bundle_id
FROM end_user_user_bundles
WHERE user_id = ?
  AND org_name = ?
  AND project_slug = ?;

-- name: GetBundleIDsByUserExplicitRoles :many
-- ⚡ 鉴权链 Step 2: 通过显式角色关联的权限包 ID 列表（单次 JOIN 查询，避免 N+1）
SELECT DISTINCT rb.bundle_id
FROM end_user_role_users ur
  JOIN end_user_role_bundles rb
    ON ur.role_id = rb.role_id
   AND ur.org_name = rb.org_name
WHERE ur.user_id = ?
  AND ur.org_name = ?
  AND rb.project_slug = ?;

-- name: GetBundleIDsByImplicitRoles :many
-- ⚡ 鉴权链 Step 3: 隐式角色关联的权限包 ID 列表（对所有认证用户执行，无需 user_id）
SELECT DISTINCT rb.bundle_id
FROM end_user_roles r
  JOIN end_user_role_bundles rb
    ON r.id = rb.role_id
   AND r.org_name = rb.org_name
WHERE r.org_name = ?
  AND rb.project_slug = ?
  AND r.is_implicit = TRUE;

-- name: GetPermissionsByBundleIDs :many
-- ⚡ 鉴权链 Step 4: 展开权限包 → 权限点（动态 IN，适用于 Step 1~3 合并后的 bundle_id 集合）
SELECT p.*
FROM end_user_permissions p
  JOIN end_user_bundle_permissions bp ON p.id = bp.permission_id
WHERE bp.bundle_id IN (sqlc.slice(bundleIDs))
  AND p.org_name = ?;
