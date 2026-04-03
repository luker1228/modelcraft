-- =============================================================================
-- 用户管理 (User Management)
-- 包含：用户表、用户-组织关联表
-- 实现混合多租户认证：Casdoor 负责认证，ModelCraft 负责授权
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. 用户表 (Users)
-- 最小化设计：仅存储 ModelCraft 内部 UUID 和外部认证提供者 ID
-- 身份信息（邮箱、姓名、头像）来自 JWT Claims，避免同步复杂性
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `users` (
  `id` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '内部 UUID',
  `external_id` VARCHAR(255) NOT NULL UNIQUE COMMENT '外部认证提供者用户 ID（来自 JWT.sub，通常为 Casdoor 用户 ID）',
  `name` VARCHAR(255) NOT NULL UNIQUE COMMENT '用户姓名（来自 Casdoor）',
  `phone` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '用户手机号（来自 Casdoor）',
  `display_name` VARCHAR(255) COMMENT '用于 UI 显示的名称',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  INDEX `idx_external_id` (`external_id`) COMMENT '按外部 ID 快速查找'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- -----------------------------------------------------------------------------
-- 2. 用户-组织关联表 (User Organizations)
-- 将用户映射到组织，一个用户可属于多个组织
-- 注意：角色信息在 07_roles_permissions.sql 的 user_roles 表中管理
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `user_organizations` (
  `id` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT 'UUID',
  `user_id` VARCHAR(36) NOT NULL COMMENT '用户 ID（引用 users.id）',
  `org_name` VARCHAR(36) NOT NULL COMMENT '组织名称（引用 organizations.name）',
  `status` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态：active、suspended、invited',
  `invited_by` VARCHAR(36) COMMENT '邀请人（用户引用）',
  `invited_at` DATETIME(3) COMMENT '邀请发送时间',
  `joined_at` DATETIME(3) COMMENT '接受邀请时间',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  -- 外键约束
  CONSTRAINT `fk_user_org_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_org_org` FOREIGN KEY (`org_name`) REFERENCES `organizations`(`name`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_org_invited_by` FOREIGN KEY (`invited_by`) REFERENCES `users`(`id`) ON DELETE SET NULL,

  -- 唯一约束：每个用户在每个组织中只能有一条关联记录
  UNIQUE KEY `uk_user_org` (`user_id`, `org_name`) COMMENT '防止重复关联',

  -- 索引
  INDEX `idx_uo_user_id` (`user_id`) COMMENT '查找用户所属的所有组织',
  INDEX `idx_uo_status` (`status`) COMMENT '按成员状态筛选',
  INDEX `idx_uo_org_name` (`org_name`) COMMENT '按组织名称查找成员'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-组织关联表';

-- -----------------------------------------------------------------------------
-- 3. 更新组织表的外键约束
-- 现在 users 表已存在，可以为 organizations.owner_id 添加外键
-- -----------------------------------------------------------------------------
ALTER TABLE `organizations`
  ADD CONSTRAINT `fk_org_owner`
  FOREIGN KEY (`owner_id`) REFERENCES `users`(`id`) ON DELETE RESTRICT;
