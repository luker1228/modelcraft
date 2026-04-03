CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          VARCHAR(36)  NOT NULL,
    user_id     VARCHAR(36)  NOT NULL,
    token_hash  VARCHAR(64)  NOT NULL COMMENT 'SHA256 hash，不存明文',
    expires_at  DATETIME     NOT NULL,
    created_at  DATETIME     NOT NULL,
    revoked_at  DATETIME     NULL     COMMENT 'NULL = 有效',
    PRIMARY KEY (id),
    INDEX idx_token_hash (token_hash),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
