-- =============================================================================
-- 用户管理 (User Management)
-- 包含：用户表（含 is_admin / status，已合并原 user_orgs 表）
-- 混合认证：支持手机号+密码本地注册，同时兼容外部认证提供者（AuthProvider）
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. 用户表 (Users)
-- 混合认证设计：支持手机号+密码本地注册登录，同时兼容外部认证提供者（AuthProvider）
-- external_id 可为 NULL（本地注册用户无外部 ID）
-- is_admin / status 原存于 user_orgs 表，因每用户仅属一个 Org 故合并至此
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `users` (
  `id` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '内部 UUID',
  `external_id` VARCHAR(255) NULL COMMENT '外部认证提供者用户 ID（来自 JWT.sub，AuthProvider 用户有值，本地注册用户为 NULL）',
  `name` VARCHAR(255) NOT NULL COMMENT '用户名（userName）',
  `phone` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '用户手机号',
  `password_hash` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'bcrypt 密码哈希（本地注册用户有值，AuthProvider 用户为空）',
  `display_name` VARCHAR(255) COMMENT '用于 UI 显示的名称',
  `org_name` VARCHAR(36) NOT NULL DEFAULT '' COMMENT '所属 Org，创建时绑定（引用 organizations.name）',
  `is_admin` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为管理员',
  `status` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态：active | suspended',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',

  UNIQUE INDEX `uk_org_user_phone` (`org_name`, `phone`, `delete_token`) COMMENT 'Org 内手机号唯一',
  UNIQUE INDEX `uk_org_user_name` (`org_name`, `name`, `delete_token`) COMMENT 'Org 内用户名唯一',
  INDEX `idx_external_id` (`external_id`) COMMENT '按外部 ID 快速查找',
  INDEX `idx_users_live_name` (`deleted_at`, `org_name`, `name`) COMMENT '活跃用户查询索引',
  CONSTRAINT `fk_users_org` FOREIGN KEY (`org_name`) REFERENCES `organizations`(`name`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- -----------------------------------------------------------------------------
-- 2. 用户资料表 (Profile)
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
  `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',

  CONSTRAINT `fk_profile_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
  UNIQUE KEY `uk_profile_user_id` (`user_id`, `delete_token`) COMMENT '保证一个用户仅有一个活跃 profile',
  INDEX `idx_profile_live_user` (`deleted_at`, `user_id`) COMMENT '活跃 profile 查询索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户资料表';

-- -----------------------------------------------------------------------------
-- 3. 更新组织表的外键约束
-- 现在 users 表已存在，可以为 organizations.owner_id 添加外键
-- -----------------------------------------------------------------------------
ALTER TABLE `organizations`
  ADD CONSTRAINT `fk_org_owner`
  FOREIGN KEY (`owner_id`) REFERENCES `users`(`id`) ON DELETE RESTRICT;
