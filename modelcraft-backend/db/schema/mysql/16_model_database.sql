-- =============================================================================
-- Model Database Registry (2026)
-- 说明：
-- - 新增 model_database 表，记录 Project 已接管的 MySQL database
-- - mode: self_hosted（可读写，支持新建/导入模型）/ managed（只读，仅同步模型）
-- - 通过 (project_slug, name, delete_token) 联合唯一索引支持软删除后重复接管
-- =============================================================================

CREATE TABLE IF NOT EXISTS `model_database` (
  `id`           VARCHAR(36)   NOT NULL COMMENT '数据库唯一标识符 (UUID)',
  `org_name`     VARCHAR(64)   NOT NULL COMMENT '所属组织名称',
  `project_slug` VARCHAR(64)   NOT NULL COMMENT '所属项目标识符',
  `cluster_id`   VARCHAR(36)   NOT NULL COMMENT '所属数据库集群 ID',
  `name`         VARCHAR(64)   NOT NULL COMMENT 'MySQL database 原始名，来自 SHOW DATABASES，注册后不可修改',
  `title`        VARCHAR(128)  NOT NULL COMMENT '用户设置的友好名称，默认等于 name',
  `description`  TEXT          NULL     COMMENT '可选描述',
  `mode`         ENUM('self_hosted','managed') NOT NULL COMMENT 'self_hosted=可读写; managed=只读',
  `deleted_at`   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',
  `created_at`   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_project_database` (`project_slug`, `name`, `delete_token`),
  KEY `idx_model_database_cluster` (`cluster_id`),
  KEY `idx_model_database_project` (`org_name`, `project_slug`),
  KEY `idx_model_database_live_project` (`org_name`, `project_slug`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Project 已接管的 MySQL database 注册表';
