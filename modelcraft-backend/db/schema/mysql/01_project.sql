-- projects 表 SQL 定义
-- 项目信息表（多租户架构，复合主键）

CREATE TABLE IF NOT EXISTS `projects` (
  -- 复合主键字段
  `org_name` VARCHAR(36) NOT NULL COMMENT '组织名称（复合主键之一）',
  `slug` VARCHAR(64) NOT NULL COMMENT '项目标识符（组织内唯一，复合主键之二）',

  -- 基本信息字段
  `title` VARCHAR(255) NOT NULL COMMENT '项目显示标题',
  `description` TEXT NULL COMMENT '项目描述信息',

  -- 状态字段
  `status` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '项目状态：active/archived（永不物理删除）',

  -- 集群关联字段
  `cluster_id` VARCHAR(36) NULL COMMENT '关联的集群ID（一对一关系）',

  -- 时间戳字段
  `created_at` DATETIME(3) NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL COMMENT '更新时间',
  `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',

  -- 复合主键约束
  PRIMARY KEY (`org_name`, `slug`, `delete_token`) COMMENT '组织+项目标识符复合主键（含软删避让位）',

  -- 普通索引
  KEY `idx_org_name` (`org_name`) COMMENT '组织查询索引',
  KEY `idx_project_slug` (`slug`) COMMENT '项目标识符查询索引',
  KEY `idx_projects_org_slug` (`org_name`, `slug`) COMMENT '活跃项目查询索引',
  KEY `idx_project_status` (`status`) COMMENT '状态查询索引',
  KEY `idx_project_cluster` (`cluster_id`) COMMENT '集群查询索引',
  KEY `idx_project_live_org` (`org_name`, `deleted_at`) COMMENT '组织活跃项目查询索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目信息表（多租户，复合主键，只归档不删除）';


-- project_auth_schemas 表 SQL 定义
-- 每个项目的认证变量配置表（每个项目只有一份）

CREATE TABLE IF NOT EXISTS `project_auth_schemas` (
  `id`           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `org_name`     VARCHAR(36) NOT NULL COMMENT '所属组织名称',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '所属项目标识符',

  -- 扩展变量配置（JSON 数组）
  `variables`    JSON NOT NULL COMMENT '认证变量列表 [{name, source, type}]',

  -- 时间戳字段
  `created_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  -- 唯一约束：每个项目只有一份 auth schema
  UNIQUE KEY `uk_auth_schemas_project` (`org_name`, `project_slug`),

  KEY `idx_auth_schemas_project` (`org_name`, `project_slug`) COMMENT '项目查询索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Project 认证变量配置';
