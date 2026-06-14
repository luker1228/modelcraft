-- ============================================
-- Migration: RLS Policy V2 — 多策略存储（role + action + 表达式）
-- Description: 替换旧单策略表，支持每个 model 多条 policy，按 action+role 匹配
-- ============================================

-- ----------------------------------------
-- 1. 删除旧 model_rls_policies 表
-- ----------------------------------------
DROP TABLE IF EXISTS model_rls_policies;

-- ----------------------------------------
-- 2. 创建新 model_rls_policies 表（多策略存储）
-- ----------------------------------------
CREATE TABLE model_rls_policies (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    org_name        VARCHAR(36) NOT NULL COMMENT '组织名',
    project_slug    VARCHAR(64) NOT NULL COMMENT '项目标识',
    model_id        VARCHAR(36) NOT NULL COMMENT '模型 ID',

    -- 策略元信息
    policy_name     VARCHAR(255) NOT NULL COMMENT '策略名称（model 内唯一）',
    action          ENUM('read', 'create', 'update', 'delete') NOT NULL COMMENT '操作类型',
    role            VARCHAR(255) NOT NULL DEFAULT '' COMMENT '匹配角色（空=默认策略）',

    -- 表达式（CEL / legacy JSON 文本）
    using_expr      TEXT NULL COMMENT 'USING 表达式（read/update/delete）',
    with_check_expr TEXT NULL COMMENT 'WITH CHECK 表达式（create/update）',

    created_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    -- 唯一约束：同一 model 下 policy_name 唯一
    UNIQUE KEY uk_policy_name (org_name, project_slug, model_id, policy_name),

    -- 查询索引：按 action + role 匹配
    INDEX idx_policy_match (org_name, project_slug, model_id, action, role)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='RLS 策略表（多策略存储）';
