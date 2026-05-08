-- =============================================================================
-- End User Auth v2 (Org Scoped in mc_meta)
-- 说明：
-- - EndUser 账号作用域由 (org_name, project_slug) 上移到 org_name
-- - EndUser 与 Project 的访问关系通过 end_user_role_users 反查（Role Assignment 为唯一授权通道）
-- - 表名统一使用 end_user_ 前缀，避免与平台 users/accounts 命名冲突
-- =============================================================================

CREATE TABLE IF NOT EXISTS `end_user_users` (
  `id` VARCHAR(36) NOT NULL COMMENT '终端用户 ID (UUID)',
  `org_name` VARCHAR(36) NOT NULL COMMENT '所属 Org',
  `username` VARCHAR(64) NOT NULL COMMENT 'Org 内唯一用户名',
  `password` VARCHAR(255) NOT NULL COMMENT 'bcrypt 密码哈希',
  `is_forbidden` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否禁用',
  `is_builtin` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为平台内置账号（每个 Org 唯一，不可删除/禁用）',
  `created_by` VARCHAR(36) NULL COMMENT '创建者（平台用户 ID）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_users_org_username` (`org_name`, `username`, `delete_token`),
  UNIQUE KEY `uk_end_user_users_org_id` (`org_name`, `id`, `delete_token`),
  KEY `idx_end_user_users_org_id_fk` (`org_name`, `id`),
  KEY `idx_end_user_users_org` (`org_name`),
  KEY `idx_end_user_users_created_by` (`created_by`),
  KEY `idx_end_user_users_live_org` (`org_name`, `deleted_at`),

  CONSTRAINT `fk_end_user_users_created_by`
    FOREIGN KEY (`created_by`) REFERENCES `users`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='终端用户账号表（Org 级隔离）';

CREATE TABLE IF NOT EXISTS `end_user_accounts` (
  `id` VARCHAR(36) NOT NULL COMMENT '会话 ID (UUID)',
  `org_name` VARCHAR(36) NOT NULL COMMENT '所属 Org',
  `user_id` VARCHAR(36) NOT NULL COMMENT '终端用户 ID',
  `refresh_token_hash` VARCHAR(255) NOT NULL COMMENT '刷新令牌哈希',
  `expires_at` DATETIME NOT NULL COMMENT '过期时间',
  `revoked` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已撤销',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_accounts_token_hash` (`refresh_token_hash`),
  KEY `idx_end_user_accounts_org_user` (`org_name`, `user_id`),
  KEY `idx_end_user_accounts_org` (`org_name`),

  CONSTRAINT `fk_end_user_accounts_user`
    FOREIGN KEY (`org_name`, `user_id`)
    REFERENCES `end_user_users`(`org_name`, `id`)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='终端用户会话表（Org 级隔离）';

CREATE TABLE IF NOT EXISTS `end_user_roles` (
  `id` VARCHAR(36) NOT NULL COMMENT '角色 ID (UUID)',
  `org_name` VARCHAR(36) NOT NULL COMMENT '所属 Org',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '所属项目',
  `name` VARCHAR(64) NOT NULL COMMENT 'Project 内唯一角色名',
  `description` VARCHAR(255) NULL COMMENT '角色描述',
  `is_implicit` TINYINT(1) NOT NULL DEFAULT 0
    COMMENT '内置隐式角色标志：0=显式角色（用户手动分配），1=隐式角色（系统自动注入）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_roles_project_name` (`org_name`, `project_slug`, `name`, `delete_token`),
  UNIQUE KEY `uk_end_user_roles_org_id` (`org_name`, `id`, `delete_token`),
  KEY `idx_end_user_roles_org_id_fk` (`org_name`, `id`),
  KEY `idx_end_user_roles_project` (`org_name`, `project_slug`),
  KEY `idx_end_user_roles_implicit` (`is_implicit`),
  KEY `idx_end_user_roles_live_project` (`org_name`, `project_slug`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='终端用户角色表（Project 级隔离）';

CREATE TABLE IF NOT EXISTS `end_user_role_users` (
  `id` VARCHAR(36) NOT NULL COMMENT '关联 ID (UUID)',
  `org_name` VARCHAR(36) NOT NULL COMMENT '所属 Org',
  `role_id` VARCHAR(36) NOT NULL COMMENT '角色 ID',
  `user_id` VARCHAR(36) NOT NULL COMMENT '终端用户 ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_eu_role_users_org_role_user` (`org_name`, `role_id`, `user_id`),
  KEY `idx_eu_role_users_org_role` (`org_name`, `role_id`),
  KEY `idx_eu_role_users_org_user` (`org_name`, `user_id`),

  CONSTRAINT `fk_eu_role_users_role`
    FOREIGN KEY (`org_name`, `role_id`)
    REFERENCES `end_user_roles`(`org_name`, `id`) ON DELETE CASCADE,

  CONSTRAINT `fk_eu_role_users_user`
    FOREIGN KEY (`org_name`, `user_id`)
    REFERENCES `end_user_users`(`org_name`, `id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='终端用户角色-用户关联表（Org 级隔离）';
