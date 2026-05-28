-- Project 级角色表（原 end_user_roles）
CREATE TABLE IF NOT EXISTS `project_roles` (
  `id`           VARCHAR(36)     NOT NULL COMMENT '角色 ID (UUID)',
  `org_name`     VARCHAR(36)     NOT NULL COMMENT '所属 Org',
  `project_slug` VARCHAR(64)     NOT NULL COMMENT '所属项目',
  `name`         VARCHAR(64)     NOT NULL COMMENT 'Project 内唯一角色名',
  `description`  VARCHAR(255)    NULL     COMMENT '角色描述',
  `is_implicit`  TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '内置隐式角色标志',
  `created_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_roles_name`   (`org_name`, `project_slug`, `name`, `delete_token`),
  UNIQUE KEY `uk_project_roles_org_id` (`org_name`, `id`, `delete_token`),
  KEY `idx_project_roles_org_id_fk`    (`org_name`, `id`),
  KEY `idx_project_roles_project`      (`org_name`, `project_slug`),
  KEY `idx_project_roles_implicit`     (`is_implicit`),
  KEY `idx_project_roles_live`         (`org_name`, `project_slug`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Project 级数据角色表';

-- Project 级角色-用户关联表（原 end_user_role_users，FK 改指向 users）
CREATE TABLE IF NOT EXISTS `project_role_users` (
  `id`           VARCHAR(36) NOT NULL COMMENT '关联 ID (UUID)',
  `org_name`     VARCHAR(36) NOT NULL COMMENT '所属 Org',
  `role_id`      VARCHAR(36) NOT NULL COMMENT '角色 ID（引用 project_roles.id）',
  `user_id`      VARCHAR(36) NOT NULL COMMENT '用户 ID（引用 users.id）',
  `created_at`   DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_role_users` (`org_name`, `role_id`, `user_id`),
  KEY `idx_project_role_users_role`  (`org_name`, `role_id`),
  KEY `idx_project_role_users_user`  (`org_name`, `user_id`),
  CONSTRAINT `fk_project_role_users_role` FOREIGN KEY (`org_name`, `role_id`) REFERENCES `project_roles`(`org_name`, `id`) ON DELETE CASCADE,
  CONSTRAINT `fk_project_role_users_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Project 级角色-用户关联表';
