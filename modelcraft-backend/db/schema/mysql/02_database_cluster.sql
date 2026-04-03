-- database_clusters 表 SQL 定义
-- 数据库集群配置信息表

CREATE TABLE IF NOT EXISTS `database_clusters` (
  -- 主键字段
  `id` VARCHAR(36) NOT NULL COMMENT '集群唯一标识符',

  -- 项目关联字段（复合引用）
  `org_name` VARCHAR(36) NOT NULL COMMENT '所属组织名称（来自projects表复合主键）',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '所属项目标识符（来自projects表复合主键）',

  -- 基本信息字段
  `title` VARCHAR(255) NOT NULL COMMENT '集群显示标题',
  `description` TEXT NULL COMMENT '集群描述信息',

  -- 连接信息字段
  `host` VARCHAR(255) NOT NULL COMMENT '数据库主机地址',
  `port` BIGINT NOT NULL DEFAULT 3306 COMMENT '数据库端口号',
  `username` VARCHAR(255) NOT NULL COMMENT '数据库用户名',
  `password` TEXT NOT NULL COMMENT '数据库密码',
  `connection_timeout` INT NOT NULL DEFAULT 5 COMMENT '连接超时(秒,5-15)',
  `charset` VARCHAR(50) NULL DEFAULT 'utf8mb4' COMMENT '数据库字符集',

  -- 连接池配置字段
  `max_open_conns` BIGINT NULL DEFAULT 100 COMMENT '最大打开连接数',
  `max_idle_conns` BIGINT NULL DEFAULT 10 COMMENT '最大空闲连接数',
  `conn_max_lifetime` BIGINT NULL DEFAULT 3600 COMMENT '连接最大生命周期（秒）',

  -- 状态和版本字段
  `status` VARCHAR(20) NULL DEFAULT 'active' COMMENT '集群状态:active/inactive',
  `version` BIGINT NULL DEFAULT 1 COMMENT '数据版本号',

  -- 时间戳字段
  `created_at` DATETIME(3) NULL COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL COMMENT '更新时间',

  -- 主键约束
  PRIMARY KEY (`id`),

  -- 不使用外键约束（避免复杂性，Project永不删除只归档）

  -- 唯一索引
  UNIQUE KEY `idx_cluster_project_unique` (`org_name`, `project_slug`) COMMENT '一对一约束：一个项目只能有一个集群',

  -- 普通索引
  KEY `idx_status` (`status`) COMMENT '状态查询索引',
  KEY `idx_cluster_project` (`org_name`, `project_slug`) COMMENT '项目查询索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据库集群配置信息表';
