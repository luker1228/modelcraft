CREATE TABLE IF NOT EXISTS api_keys (
    id           VARCHAR(36)  NOT NULL,
    user_id      VARCHAR(36)  NOT NULL,
    name         VARCHAR(100) NOT NULL COMMENT '用户命名，如 GitHub Actions',
    key_hash     VARCHAR(64)  NOT NULL COMMENT 'SHA256 hash，不存明文',
    key_prefix   VARCHAR(10)  NOT NULL COMMENT '完整 key 前 10 位，如 mc_a1b2c3d4',
    role_ids     JSON         NULL     COMMENT '绑定的角色 ID 列表（roles.id）',
    last_used_at DATETIME     NULL     COMMENT '防抖：距上次 > 1 分钟才更新',
    expires_at   DATETIME     NULL     COMMENT 'NULL = 永不过期',
    created_at   DATETIME     NOT NULL,
    revoked_at   DATETIME     NULL     COMMENT 'NULL = 有效',
    PRIMARY KEY (id),
    INDEX idx_key_hash (key_hash),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
