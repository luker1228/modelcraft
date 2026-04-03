CREATE TABLE IF NOT EXISTS security_audit_logs (
    id         VARCHAR(36) NOT NULL,
    user_id    VARCHAR(36) NOT NULL,
    event      VARCHAR(50) NOT NULL COMMENT '如 REUSE_DETECTED',
    detail     JSON        NULL     COMMENT 'token_id, ip 等上下文',
    created_at DATETIME    NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_user_id_created (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
