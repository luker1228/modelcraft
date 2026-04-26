-- =============================================================================
-- 用户管理 (User Management)
-- 包含：用户表、用户-组织关联表
-- 混合认证：支持手机号+密码本地注册，同时兼容外部认证提供者（AuthProvider）
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. 用户表 (Users)
-- 混合认证设计：支持手机号+密码本地注册登录，同时兼容外部认证提供者（AuthProvider）
-- external_id 可为 NULL（本地注册用户无外部 ID）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `users` (
  `id` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '内部 UUID',
  `external_id` VARCHAR(255) NULL COMMENT '外部认证提供者用户 ID（来自 JWT.sub，AuthProvider 用户有值，本地注册用户为 NULL）',
  `name` VARCHAR(255) NOT NULL COMMENT '用户名（userName）',
  `phone` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '用户手机号',
  `password_hash` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'bcrypt 密码哈希（本地注册用户有值，AuthProvider 用户为空）',
  `display_name` VARCHAR(255) COMMENT '用于 UI 显示的名称',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  UNIQUE INDEX `uk_phone` (`phone`) COMMENT '手机号唯一约束（用于本地登录）',
  UNIQUE INDEX `uk_user_name` (`name`) COMMENT '用户名（userName）唯一约束',
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
-- 3. 用户资料表 (Profile)
-- user 与 profile 为 1:1 关系：profile.user_id UNIQUE + FK(users.id)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `profile` (
  `id` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT 'profile UUID',
  `user_id` VARCHAR(36) NOT NULL COMMENT '用户 ID（引用 users.id）',
  `nickname` VARCHAR(32) NOT NULL COMMENT '昵称',
  `avatar_url` VARCHAR(512) NULL COMMENT '头像 URL（当前可为空，上传能力后续实现）',
  `bio` VARCHAR(256) NULL COMMENT '个人简介',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  CONSTRAINT `fk_profile_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
  UNIQUE KEY `uk_profile_user_id` (`user_id`) COMMENT '保证一个用户仅有一个 profile'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户资料表';

-- -----------------------------------------------------------------------------
-- 4. 更新组织表的外键约束
-- 现在 users 表已存在，可以为 organizations.owner_id 添加外键
-- -----------------------------------------------------------------------------
ALTER TABLE `organizations`
  ADD CONSTRAINT `fk_org_owner`
  FOREIGN KEY (`owner_id`) REFERENCES `users`(`id`) ON DELETE RESTRICT;
