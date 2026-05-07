-- 15_end_user_builtin.sql
-- Add is_builtin flag to end_user_users table.
-- Uniqueness per org is enforced at application layer only
-- (multiple rows with is_builtin=0 are valid, so a DB UNIQUE would break).

ALTER TABLE `end_user_users`
  ADD COLUMN `is_builtin` TINYINT(1) NOT NULL DEFAULT 0
    COMMENT '是否为平台内置账号（每个 Org 唯一，不可删除/禁用）'
    AFTER `is_forbidden`;
