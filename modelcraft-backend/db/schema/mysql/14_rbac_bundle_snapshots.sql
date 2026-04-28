-- =============================================================
-- 14_rbac_bundle_snapshots.sql
-- 权限包版本快照
-- 依赖: 13_rbac_permissions.sql（end_user_permission_bundles）
-- =============================================================

-- -------------------------------------------------------------
-- 1. end_user_permission_bundle_snapshots — 权限包历史快照
--    每次权限列表变更时自动保存只读快照，最多保留最近 5 个版本
-- -------------------------------------------------------------
CREATE TABLE `end_user_permission_bundle_snapshots` (
  `id`              VARCHAR(36)  NOT NULL                    COMMENT '快照 UUID',
  `bundle_id`       VARCHAR(36)  NOT NULL                    COMMENT '所属权限包 ID，FK → end_user_permission_bundles.id',
  `version`         INT          NOT NULL                    COMMENT '版本号，从 1 开始，每个权限包独立计数',
  `permissions`     JSON         NOT NULL                    COMMENT '快照时刻的权限点 ID 数组，格式：[{"permissionId":"uuid","sortOrder":0}]',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '快照创建时间',
  `created_by`      VARCHAR(128) NULL                        COMMENT '操作人标识（用户 ID 或 system）',
  `restored_from`   INT          NULL DEFAULT NULL           COMMENT '若为回滚操作，指向来源版本号；否则为 NULL',

  PRIMARY KEY (`id`),
  -- 同一权限包内版本号不可重复
  UNIQUE KEY `uq_bundle_snapshots_bundle_version`
    (`bundle_id`, `version`),
  -- 按权限包 + 版本倒序快速查询（获取最近 N 个版本）
  INDEX `idx_bundle_snapshots_bundle_version_desc` (`bundle_id`, `version` DESC),
  -- FK → end_user_permission_bundles.id（权限包删除时级联删除快照）
  CONSTRAINT `fk_bundle_snapshots_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='权限包历史快照：记录每次权限列表变更，支持回滚';
