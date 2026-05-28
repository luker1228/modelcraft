-- =============================================================================
-- Admin Role Protection (2025)
-- 说明：
-- - project_roles 增加 is_protected 列
-- - is_protected=1 的角色：不可删除、不可改名、不可修改权限包关联
-- - 每个 Project 在创建时自动生成一个 is_protected=1 的 admin 角色，
--   并将 Org 的 builtin admin 用户分配到该角色（同一事务，强校验）
-- =============================================================================

ALTER TABLE `project_roles`
  ADD COLUMN `is_protected` TINYINT(1) NOT NULL DEFAULT 0
    COMMENT '受保护角色：不可删除、不可改名、不可修改权限包关联'
    AFTER `is_implicit`;
