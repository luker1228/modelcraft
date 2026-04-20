-- ============================================
-- Migration: RLS (Row Level Security)
-- Description: 行级数据隔离功能
-- ============================================

-- ----------------------------------------
-- 1. models 表新增 created_via 字段
-- ----------------------------------------
ALTER TABLE models
    ADD COLUMN created_via ENUM('NEW', 'IMPORTED') NOT NULL DEFAULT 'NEW'
    COMMENT '模型创建来源：NEW=新建，IMPORTED=导入';

CREATE INDEX idx_models_created_via ON models(created_via);

-- ----------------------------------------
-- 2. 创建 model_rls_policies 表
-- ----------------------------------------
CREATE TABLE IF NOT EXISTS model_rls_policies (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    model_id            VARCHAR(36) NOT NULL COMMENT '模型 ID',

    -- 五件套 JSON 表达式（存储为 JSON 字符串）
    select_predicate    TEXT NOT NULL COMMENT 'SELECT USING 谓词 JSON',
    insert_check        TEXT NOT NULL COMMENT 'INSERT WITH CHECK 谓词 JSON',
    update_predicate    TEXT NOT NULL COMMENT 'UPDATE USING 谓词 JSON',
    update_check        TEXT NOT NULL COMMENT 'UPDATE WITH CHECK 谓词 JSON',
    delete_predicate    TEXT NOT NULL COMMENT 'DELETE USING 谓词 JSON',

    -- 元数据
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    -- 外键约束
    CONSTRAINT fk_mrp_model
        FOREIGN KEY (model_id) REFERENCES models(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,

    -- 唯一约束：每个模型只有一个 Policy
    UNIQUE KEY uk_model_id (model_id)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Model RLS 策略配置';

CREATE INDEX idx_mrp_model_id ON model_rls_policies(model_id);

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
