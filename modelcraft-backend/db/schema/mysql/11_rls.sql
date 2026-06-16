-- ============================================
-- Migration: RLS (Row Level Security)
-- Description: 行级数据隔离功能
-- ============================================

-- ----------------------------------------
-- 3. 创建 project_auth_schemas 表
-- ----------------------------------------
CREATE TABLE IF NOT EXISTS project_auth_schemas (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    org_name        VARCHAR(36) NOT NULL COMMENT '所属组织名称',
    project_slug    VARCHAR(64) NOT NULL COMMENT '所属项目标识符',

    -- 扩展变量配置（JSON 数组）
    variables       JSON NOT NULL COMMENT '认证变量列表 [{name, source, type}]',

    -- 元数据
    created_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    -- 唯一约束：每个项目只有一份 auth schema
    UNIQUE KEY uk_auth_schemas_project (org_name, project_slug),

    -- 索引
    KEY idx_auth_schemas_project (org_name, project_slug) COMMENT '项目查询索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Project 认证变量配置';
