-- projects 表 SQL 定义
-- 项目信息表（多租户架构，复合主键）

CREATE TABLE IF NOT EXISTS `projects` (
  -- 复合主键字段
  `org_name` VARCHAR(36) NOT NULL COMMENT '组织名称（复合主键之一）',
  `slug` VARCHAR(64) NOT NULL COMMENT '项目标识符（组织内唯一，复合主键之二）',

  -- 基本信息字段
  `title` VARCHAR(255) NOT NULL COMMENT '项目显示标题',
  `description` TEXT NULL COMMENT '项目描述信息',
  `login_url` VARCHAR(512) NULL COMMENT '项目登录URL地址',

  -- 状态字段
  `status` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '项目状态：active/archived（永不物理删除）',

  -- 集群关联字段
  `cluster_id` VARCHAR(36) NULL COMMENT '关联的集群ID（一对一关系）',

  -- 时间戳字段
  `created_at` DATETIME(3) NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL COMMENT '更新时间',

  -- 复合主键约束
  PRIMARY KEY (`org_name`, `slug`) COMMENT '组织+项目标识符复合主键',

  -- 普通索引
  KEY `idx_org_name` (`org_name`) COMMENT '组织查询索引',
  KEY `idx_project_slug` (`slug`) COMMENT '项目标识符查询索引',
  KEY `idx_project_status` (`status`) COMMENT '状态查询索引',
  KEY `idx_project_cluster` (`cluster_id`) COMMENT '集群查询索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目信息表（多租户，复合主键，只归档不删除）';
