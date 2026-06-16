-- =============================================================================
-- End-User API Tokens (PAT)
-- 说明：
-- - 存储 EndUser 创建的 Personal Access Token（明文不落库，存 SHA-256 hash）
-- - token_hash UNIQUE 索引用于 O(1) 验证
-- - 软删除：deleted_at + delete_token 联合避让，支持撤销后同名重建
-- =============================================================================

CREATE TABLE IF NOT EXISTS `user_api_tokens` (
  `id`            VARCHAR(36)   NOT NULL COMMENT '唯一标识符 (UUID v7)',
  `org_name`      VARCHAR(255)  NOT NULL COMMENT '所属组织',
  `end_user_id`   VARCHAR(36)   NOT NULL COMMENT '创建者 EndUser ID',
  `name`          VARCHAR(255)  NOT NULL COMMENT '用户自定义名称',
  `token_hash`    VARCHAR(64)   NOT NULL COMMENT 'SHA-256(plaintext) hex，用于验证',
  `expires_at`    DATETIME      NULL     COMMENT 'NULL 表示永不过期',
  `last_used_at`  DATETIME      NULL     COMMENT '最近使用时间，异步更新',
  `created_at`    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `deleted_at`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token`  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_token_hash` (`token_hash`),
  UNIQUE KEY `uq_user_token_name` (`org_name`, `end_user_id`, `name`, `delete_token`),
  KEY `idx_user_tokens` (`org_name`, `end_user_id`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='EndUser Personal Access Token 注册表（PAT）';
