-- models 表 SQL 定义
-- 模型定义主表

CREATE TABLE IF NOT EXISTS `models` (
  -- 主键字段
  `id` VARCHAR(36) NOT NULL COMMENT '模型唯一标识符',

  -- 项目关联字段（复合引用）
  `org_name` VARCHAR(36) NOT NULL COMMENT '所属组织名称（来自projects表复合主键）',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '所属项目标识符（来自projects表复合主键）',

  -- 基本信息字段
  `name` VARCHAR(64) NOT NULL COMMENT '模型名称（唯一标识）',
  `title` VARCHAR(255) NOT NULL COMMENT '模型显示标题',
  `description` TEXT NULL COMMENT '模型描述信息',

  -- 存储配置字段
  `storage_type` VARCHAR(100) NOT NULL COMMENT '存储类型',
  `database_name` VARCHAR(64) NOT NULL COMMENT '数据库名称',

  -- 运行时配置字段
  `display_field` VARCHAR(64) NULL COMMENT '用于 runtime __label 解析的字段名（必须是模型中存在且可字符串化的字段）',

  -- 版本和状态字段
  `version` BIGINT NULL DEFAULT 1 COMMENT '数据版本号',
  `status` VARCHAR(50) NULL DEFAULT 'draft' COMMENT '模型状态：draft/published/archived',
  `group_id` VARCHAR(36) NULL COMMENT '所属分组ID，NULL表示未分组（ungrouped）',
  `deployment_status` VARCHAR(50) NULL DEFAULT 'pending' COMMENT '部署状态：pending/success/failed',

  -- 同步信息字段
  `last_sync_at` DATETIME(3) NULL COMMENT '最后同步时间',
  `sync_error` TEXT NULL COMMENT '同步错误信息',

  -- 时间戳字段
  `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  -- 主键约束
  PRIMARY KEY (`id`),

  -- 不使用外键约束（避免复杂性，Project永不删除只归档）

  -- 唯一索引
  UNIQUE KEY `idx_models_name` (`org_name`, `project_slug`, `database_name`, `name`) COMMENT '组织+项目内模型名称唯一索引',

  -- 普通索引
  KEY `idx_models_project` (`org_name`, `project_slug`) COMMENT '项目查询索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型定义主表';


-- model_groups 表 SQL 定义
-- 模型分组定义表

CREATE TABLE IF NOT EXISTS `model_groups` (
  -- 主键字段
  `id` VARCHAR(36) NOT NULL COMMENT '分组唯一标识符',

  -- 项目关联字段（复合引用）
  `org_name` VARCHAR(36) NOT NULL COMMENT '所属组织名称（来自projects表复合主键）',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '所属项目标识符（来自projects表复合主键）',

  -- 基本信息字段
  `name` VARCHAR(64) NOT NULL COMMENT '分组名称，仅小写字母、数字和下划线，字母开头',

  -- 排序字段
  `display_order` VARCHAR(255) NOT NULL COMMENT '字典序排序键，用于拖拽排序（lexicographic fractional index）',

  -- 时间戳字段
  `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  -- 主键约束
  PRIMARY KEY (`id`),

  -- 不使用外键约束（避免复杂性，Project永不删除只归档）

  -- 唯一索引
  UNIQUE KEY `idx_model_groups_name` (`org_name`, `project_slug`, `name`) COMMENT '组织+项目内分组名称唯一索引',

  -- 普通索引
  KEY `idx_model_groups_project_order` (`org_name`, `project_slug`, `display_order`) COMMENT '项目分组排序查询索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型分组定义表';





-- field_definitions 表 SQL 定义
-- 模型字段定义表

CREATE TABLE IF NOT EXISTS `field_definitions` (
  -- 复合主键字段
  `model_id` VARCHAR(36) NOT NULL COMMENT '所属模型ID',
  `name` VARCHAR(64) NOT NULL COMMENT '字段名称',
  
  -- 多租户隔离字段（与models表保持一致）
  `org_name` VARCHAR(36) NOT NULL COMMENT '所属组织名称（多租户隔离）',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '项目标识符（多租户隔离）',
  
  -- 模型和集群信息字段
  `model_name` VARCHAR(64) NOT NULL COMMENT '模型名称',
  `database_name` VARCHAR(64) NOT NULL COMMENT '数据库名称',
  
  -- 关系关联字段
  `parent_relation_id` CHAR(36) NULL COMMENT '依赖的关联的关系ID, 这个字段不为nul, 则该字段不能被删除',
  `enum_name` VARCHAR(64) NULL COMMENT '关联的枚举名称',

  -- 逻辑外键关联字段
  `belongs_to_fk_id` VARCHAR(36) NULL COMMENT '所属逻辑外键ID（FK列字段使用）',
  `relate_fk_id` VARCHAR(36) NULL COMMENT '关联的逻辑外键ID（RELATION格式字段使用）',

  -- 字段基本信息
  `title` VARCHAR(255) NOT NULL COMMENT '字段显示标题',
  `description` TEXT NULL COMMENT '字段描述信息',
  `format` VARCHAR(50) NOT NULL COMMENT '字段格式类型',
  
  -- 字段属性配置
  `non_null` TINYINT(1) NULL DEFAULT 0 COMMENT '是否可为null',
  `required` TINYINT(1) NULL DEFAULT 0 COMMENT '是否创建时必填',
  `is_unique` TINYINT(1) NULL DEFAULT 0 COMMENT '是否唯一',
  `is_primary` TINYINT(1) NULL DEFAULT 0 COMMENT '是否主键',
  `is_deprecated` TINYINT(1) NULL DEFAULT 0 COMMENT '是否已废弃',
  
  -- 状态和验证配置
  `status` VARCHAR(20) NOT NULL DEFAULT 'init' COMMENT '字段状态：init/active/inactive',
  `validation` JSON NULL COMMENT '字段验证规则配置',
  
  -- 显示和元数据配置
  `display_order` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '字典序排序键，用于拖拽排序（lexicographic fractional index）',
  `metadata` JSON NULL COMMENT '字段元数据配置',
  `enum_label_config` JSON NULL COMMENT '枚举标签虚拟字段配置（ENUM_LABEL格式使用）',

  -- 时间戳字段
  `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  
  -- 复合主键约束
  PRIMARY KEY (`model_id`, `name`),
  
  -- 索引：多租户查询优化（与其他表保持一致）
  INDEX `idx_field_definitions_project` (`org_name`, `project_slug`) COMMENT '项目查询索引',
  INDEX `idx_field_definitions_model` (`org_name`, `project_slug`, `model_name`) COMMENT '模型查询索引',
  
  -- 外键约束
  CONSTRAINT `fk_field_definitions_model` FOREIGN KEY (`model_id`) REFERENCES `models` (`id`) ON DELETE CASCADE
  
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型字段定义表';


-- logical_foreign_keys 表 SQL 定义
-- 逻辑外键定义表（成对存储：normal + reverse）

CREATE TABLE IF NOT EXISTS `logical_foreign_keys` (
  -- 主键字段
  `id` VARCHAR(36) NOT NULL COMMENT '逻辑外键唯一标识符',

  -- 配对字段（normal 和 reverse 共享同一 pair_id）
  `pair_id` VARCHAR(36) NOT NULL COMMENT '外键对ID，normal 和 reverse 行共享',

  -- 方向枚举
  `direction` ENUM('normal', 'reverse') NOT NULL COMMENT '外键方向：normal（正向）/ reverse（反向）',

  -- 模型关联字段
  `model_id` VARCHAR(36) NOT NULL COMMENT '所属模型ID',
  `model_name` VARCHAR(255) NOT NULL COMMENT '所属模型名称（冗余存储，model_name 不变）',
  `ref_model_id` VARCHAR(36) NOT NULL COMMENT '引用模型ID',
  `ref_model_name` VARCHAR(255) NOT NULL COMMENT '引用模型名称（冗余存储，model_name 不变）',

  -- 字段映射配置
  `source_fields` JSON NOT NULL COMMENT '源字段列表（JSON数组）',
  `target_fields` JSON NOT NULL COMMENT '目标字段列表（JSON数组）',

  -- 时间戳字段
  `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  -- 主键约束
  PRIMARY KEY (`id`),

  -- 索引
  KEY `idx_logical_fk_pair` (`pair_id`) COMMENT '外键对查询索引',
  KEY `idx_logical_fk_model` (`model_id`) COMMENT '模型查询索引',

  -- 外键约束
  CONSTRAINT `fk_logical_fk_model` FOREIGN KEY (`model_id`) REFERENCES `models` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_logical_fk_ref_model` FOREIGN KEY (`ref_model_id`) REFERENCES `models` (`id`) ON DELETE CASCADE

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='逻辑外键定义表';


-- model_enums 表 SQL 定义
-- 枚举定义表

CREATE TABLE IF NOT EXISTS `model_enums` (
  -- 主键字段
  `id` VARCHAR(36) NOT NULL COMMENT '枚举唯一标识符',

  -- 项目关联字段（复合引用）
  `org_name` VARCHAR(36) NOT NULL COMMENT '所属组织名称（来自projects表复合主键）',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '所属项目标识符（来自projects表复合主键）',

  -- 基本信息字段
  `name` VARCHAR(64) NOT NULL COMMENT '枚举名称（唯一标识）',
  `display_name` VARCHAR(255) NOT NULL COMMENT '枚举显示名称',
  `description` TEXT NULL COMMENT '枚举描述信息',

  -- 枚举选项配置
  `options` JSON NOT NULL COMMENT '枚举选项配置（JSON数组格式）',
  `is_multi_select` TINYINT(1) NULL DEFAULT 0 COMMENT '是否支持多选',

  -- 时间戳字段
  `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  -- 主键约束
  PRIMARY KEY (`id`),

  -- 不使用外键约束（避免复杂性，Project永不删除只归档）

  -- 唯一索引
  UNIQUE KEY `idx_model_enums_name` (`org_name`, `project_slug`, `name`) COMMENT '组织+项目内枚举名称唯一索引',

  -- 普通索引
  KEY `idx_model_enums_project` (`org_name`, `project_slug`) COMMENT '项目查询索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型枚举定义表';


-- model_field_enum_associations 表 SQL 定义
-- 模型字段与枚举关联表

CREATE TABLE IF NOT EXISTS `model_field_enum_associations` (
  -- 复合主键字段
  `model_id` VARCHAR(36) NOT NULL COMMENT '模型ID',
  `field_name` VARCHAR(64) NOT NULL COMMENT '字段名称',

  -- 项目关联字段（复合引用）
  `org_name` VARCHAR(36) NOT NULL COMMENT '所属组织名称',
  `project_slug` VARCHAR(64) NOT NULL COMMENT '所属项目标识符',

  -- 关联信息字段
  `enum_name` VARCHAR(64) NOT NULL COMMENT '关联的枚举名称',

  -- 数据库信息
  `database_name` VARCHAR(64) NOT NULL COMMENT '数据库名称',

  -- 时间戳字段
  `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  -- 复合主键约束
  PRIMARY KEY (`model_id`, `field_name`),

  -- 外键约束（仅对models表，不对projects表）
  CONSTRAINT `fk_field_enum_assoc_model` FOREIGN KEY (`model_id`) REFERENCES `models` (`id`) ON DELETE CASCADE,

  -- 查询优化索引
  KEY `idx_field_enum_assoc_enum_name` (`org_name`, `project_slug`, `enum_name`) COMMENT '组织+项目内枚举名称查询索引',
  KEY `idx_field_enum_assoc_cluster_db` (`database_name`) COMMENT '数据库查询索引',
  KEY `idx_field_enum_assoc_project` (`org_name`, `project_slug`) COMMENT '项目查询索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型字段与枚举关联表';
