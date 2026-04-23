-- =============================================================================
-- End User Auth (Tenant Scoped in mc_meta)
-- 说明：
-- - 所有终端用户认证数据统一存放在 mc_meta
-- - 通过 (org_name, project_slug) 做租户隔离
-- - 表名统一使用 end_user_ 前缀，避免与平台 users/accounts 命名冲突
-- =============================================================================

CREATE TABLE IF NOT EXISTS `end_user_users` (
  `id` VARCHAR(36) NOT NULL COMMENT '终端用户 ID (UUID)',
  `org_name` VARCHAR(36) NOT NULL COMMENT '租户组织名',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '租户项目标识',
  `username` VARCHAR(64) NOT NULL COMMENT '项目内唯一用户名',
  `password` VARCHAR(255) NOT NULL COMMENT 'bcrypt 密码哈希',
  `is_forbidden` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否禁用',
  `created_by` VARCHAR(36) NULL COMMENT '创建者（平台用户 ID）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_users_scope_username` (`org_name`, `project_slug`, `username`),
  UNIQUE KEY `uk_end_user_users_scope_id` (`org_name`, `project_slug`, `id`),
  KEY `idx_end_user_users_scope` (`org_name`, `project_slug`),
  KEY `idx_end_user_users_created_by` (`created_by`),

  CONSTRAINT `fk_end_user_users_created_by`
    FOREIGN KEY (`created_by`) REFERENCES `users`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='终端用户账号表（租户隔离）';

CREATE TABLE IF NOT EXISTS `end_user_accounts` (
  `id` VARCHAR(36) NOT NULL COMMENT '会话 ID (UUID)',
  `org_name` VARCHAR(36) NOT NULL COMMENT '租户组织名',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '租户项目标识',
  `user_id` VARCHAR(36) NOT NULL COMMENT '终端用户 ID',
  `refresh_token_hash` VARCHAR(255) NOT NULL COMMENT '刷新令牌哈希',
  `expires_at` DATETIME NOT NULL COMMENT '过期时间',
  `revoked` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已撤销',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_accounts_token_hash` (`refresh_token_hash`),
  KEY `idx_end_user_accounts_scope_user` (`org_name`, `project_slug`, `user_id`),
  KEY `idx_end_user_accounts_scope` (`org_name`, `project_slug`),

  CONSTRAINT `fk_end_user_accounts_user`
    FOREIGN KEY (`org_name`, `project_slug`, `user_id`)
    REFERENCES `end_user_users`(`org_name`, `project_slug`, `id`)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='终端用户会话表（租户隔离）';

CREATE TABLE IF NOT EXISTS `end_user_roles` (
  `id` VARCHAR(36) NOT NULL COMMENT '角色 ID (UUID)',
  `org_name` VARCHAR(36) NOT NULL COMMENT '租户组织名',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '租户项目标识',
  `name` VARCHAR(64) NOT NULL COMMENT '角色名（租户内唯一）',
  `description` VARCHAR(255) NULL COMMENT '角色描述',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_roles_scope_name` (`org_name`, `project_slug`, `name`),
  UNIQUE KEY `uk_end_user_roles_scope_id` (`org_name`, `project_slug`, `id`),
  KEY `idx_end_user_roles_scope` (`org_name`, `project_slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='终端用户角色表（租户隔离）';

CREATE TABLE IF NOT EXISTS `end_user_role_users` (
  `id` VARCHAR(36) NOT NULL COMMENT '关联 ID (UUID)',
  `org_name` VARCHAR(36) NOT NULL COMMENT '租户组织名',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '租户项目标识',
  `role_id` VARCHAR(36) NOT NULL COMMENT '角色 ID',
  `user_id` VARCHAR(36) NOT NULL COMMENT '终端用户 ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_end_user_role_users_scope_role_user` (`org_name`, `project_slug`, `role_id`, `user_id`),
  KEY `idx_end_user_role_users_scope_role` (`org_name`, `project_slug`, `role_id`),
  KEY `idx_end_user_role_users_scope_user` (`org_name`, `project_slug`, `user_id`),

  CONSTRAINT `fk_end_user_role_users_role`
    FOREIGN KEY (`org_name`, `project_slug`, `role_id`)
    REFERENCES `end_user_roles`(`org_name`, `project_slug`, `id`)
    ON DELETE CASCADE,
  CONSTRAINT `fk_end_user_role_users_user`
    FOREIGN KEY (`org_name`, `project_slug`, `user_id`)
    REFERENCES `end_user_users`(`org_name`, `project_slug`, `id`)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='终端用户角色-用户关联表（租户隔离）';
