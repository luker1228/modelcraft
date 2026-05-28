-- =============================================================
-- 13_rbac_permissions.sql
-- RBAC 行列级权限系统（item-centric 模型）
-- 依赖: 12_end_user_auth.sql（project_roles / project_role_users）
--       06_users.sql（users）
-- =============================================================

-- -------------------------------------------------------------
-- 1. end_user_data_permissions — 自定义数据权限实体（仅 CUSTOM）
--    一行 = 管理员手工定义的对某模型的行列级访问策略
-- -------------------------------------------------------------
CREATE TABLE `end_user_data_permissions` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT '权限实体 UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织（冗余，不做 FK）',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目（冗余，不做 FK）',
  `database_name` VARCHAR(128) NULL                        COMMENT '数据源名称（可空，预留按数据源授权能力）',
  `model_name`    VARCHAR(128) NULL                        COMMENT '模型名称（可空，预留按模型名授权能力）',
  `model_id`      VARCHAR(36)  NOT NULL                    COMMENT '关联模型 ID，FK → models.id',
  `name`          VARCHAR(128) NOT NULL                    COMMENT '权限名称，人类可读',
  `description`   TEXT         NULL                        COMMENT '权限描述',
  `column_policy` JSON         NULL                        COMMENT '列策略 JSON',
  -- column_policy 结构示例：
  -- {
  --   "defaultMode": "VISIBLE",
  --   "rules": [
  --     { "fieldName": "salary",  "mode": "MASKED",  "maskPattern": "***" },
  --     { "fieldName": "id_card", "mode": "HIDDEN" },
  --     { "fieldName": "status",  "mode": "VISIBLE" }
  --   ]
  -- }
  `row_policy`    JSON         NULL                        COMMENT '行策略 JSON，谓词为 GraphQL Runtime where 条件',
  -- row_policy 结构（每个 action 两个字段）：
  --   allowed : boolean
  --   scope   : "all"|"custom"
  --   predicate/check : GraphQL Runtime where JSON（scope=custom 时有效）
  -- {
  --   "select": { "allowed": true, "scope": "custom", "predicate": {...} },
  --   "insert": { "allowed": true, "scope": "custom", "check": {...} },
  --   "update": { "allowed": true, "scope": "custom", "predicate": {...}, "check_scope": "custom", "check": {...} },
  --   "delete": { "allowed": false }
  -- }
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at`    BIGINT UNSIGNED NOT NULL DEFAULT 0               COMMENT '软删除时间戳，0 表示活跃',
  `delete_token`  BIGINT UNSIGNED NOT NULL DEFAULT 0               COMMENT '唯一键避让位，0 表示活跃',

  PRIMARY KEY (`id`),
  -- 业务唯一键：同一模型下名称不可重复
  UNIQUE KEY `uq_permissions_model_name`
    (`model_id`, `name`, `delete_token`),
  -- 按组织/项目快速检索
  INDEX `idx_permissions_org_project` (`org_name`, `project_slug`),
  INDEX `idx_permissions_model_id` (`model_id`),
  INDEX `idx_permissions_live_org_project` (`org_name`, `project_slug`, `deleted_at`),
  -- FK → models.id（模型可删除，级联清理权限实体）
  CONSTRAINT `fk_permissions_model`
    FOREIGN KEY (`model_id`) REFERENCES `models` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='自定义数据权限实体：管理员手工定义的模型级行列策略（仅 CUSTOM）';

-- -------------------------------------------------------------
-- 2. end_user_permission_bundles — 权限包
--    权限点的命名集合，可跨模型聚合
-- -------------------------------------------------------------
CREATE TABLE `end_user_permission_bundles` (
  `id`            VARCHAR(36)   NOT NULL                   COMMENT '权限包 UUID',
  `slug`          VARCHAR(64)   NOT NULL                   COMMENT '用户可自定义的 URL 友好标识符，同项目内唯一，创建时设定后不可修改',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目',
  `name`          VARCHAR(128) NOT NULL                    COMMENT '权限包名称',
  `description`   TEXT         NULL                        COMMENT '权限包描述',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at`    BIGINT UNSIGNED NOT NULL DEFAULT 0      COMMENT '软删除时间戳，0 表示活跃',
  `delete_token`  BIGINT UNSIGNED NOT NULL DEFAULT 0      COMMENT '唯一键避让位，0 表示活跃',

  PRIMARY KEY (`id`),
  -- 同一项目下 slug 唯一（对外标识符）
  UNIQUE KEY `uq_bundles_org_project_slug`
    (`org_name`, `project_slug`, `slug`, `delete_token`),
  -- 同一项目下权限包名称唯一
  UNIQUE KEY `uq_bundles_org_project_name`
    (`org_name`, `project_slug`, `name`, `delete_token`),
  -- 快速检索
  INDEX `idx_bundles_org_project` (`org_name`, `project_slug`),
  INDEX `idx_bundles_live_org_project` (`org_name`, `project_slug`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='权限包：数据权限 item 的命名集合，用于角色授权或用户直接授权';

-- -------------------------------------------------------------
-- 3. end_user_bundle_data_permission_items — Bundle 数据权限 Item（核心绑定表）
--    每行 = bundle 在某模型上的唯一数据权限配置
--    grant_type=PRESET 时: preset 必填, custom_permission_id 为空
--    grant_type=CUSTOM 时: custom_permission_id 必填, preset 为空
-- -------------------------------------------------------------
CREATE TABLE `end_user_bundle_data_permission_items` (
  `id`                     VARCHAR(36)  NOT NULL                    COMMENT 'Item UUID',
  `bundle_id`              VARCHAR(36)  NOT NULL                    COMMENT '所属权限包 ID，FK → end_user_permission_bundles.id',
  `model_id`               VARCHAR(36)  NOT NULL                    COMMENT '目标模型 ID，FK → models.id',
  `grant_type`             ENUM('PRESET','CUSTOM') NOT NULL         COMMENT '授权来源类型',
  `preset`                 ENUM(
                             'READ_WRITE_ALL',
                             'READ_ALL',
                             'READ_WRITE_OWNER',
                             'READ_ALL_WRITE_OWNER'
                           )            NULL DEFAULT NULL           COMMENT 'PRESET 模板枚举值（仅 grant_type=PRESET 时有值）',
  `custom_permission_id`   VARCHAR(36)  NULL DEFAULT NULL           COMMENT '自定义权限实体 ID（仅 grant_type=CUSTOM 时有值），FK → end_user_data_permissions.id',
  `sort_order`             INT          NOT NULL DEFAULT 0          COMMENT '显示排序权重（ASC）',
  `created_at`             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  -- 核心约束：同一 bundle 下同一 model 最多一个 item
  UNIQUE KEY `uq_bundle_items_bundle_model`
    (`bundle_id`, `model_id`),
  INDEX `idx_bundle_items_custom_permission` (`custom_permission_id`),
  INDEX `idx_bundle_items_model_id` (`model_id`),
  -- FK → bundles（级联删除）
  CONSTRAINT `fk_bundle_items_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  -- FK → models（级联删除）
  CONSTRAINT `fk_bundle_items_model`
    FOREIGN KEY (`model_id`) REFERENCES `models` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  -- FK → custom permission（RESTRICT 防止悬挂引用；默认行为，不写显式 action 以兼容 MySQL 8.0 CHECK 约束限制）
  CONSTRAINT `fk_bundle_items_custom_permission`
    FOREIGN KEY (`custom_permission_id`) REFERENCES `end_user_data_permissions` (`id`),
  -- CHECK: PRESET item 必须有 preset，不能有 custom_permission_id
  CONSTRAINT `chk_bundle_items_preset`
    CHECK (grant_type != 'PRESET' OR (preset IS NOT NULL AND custom_permission_id IS NULL)),
  -- CHECK: CUSTOM item 必须有 custom_permission_id，不能有 preset
  CONSTRAINT `chk_bundle_items_custom`
    CHECK (grant_type != 'CUSTOM' OR (custom_permission_id IS NOT NULL AND preset IS NULL))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='Bundle 数据权限 Item：bundle 在某模型上的唯一数据权限配置';

-- -------------------------------------------------------------
-- 4. end_user_role_bundles — 角色-权限包 关联
-- -------------------------------------------------------------
CREATE TABLE `end_user_role_bundles` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT 'UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织（冗余，快速查询）',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目（冗余，快速查询）',
  `role_id`       VARCHAR(36)  NOT NULL                    COMMENT '角色 ID，FK → project_roles.id',
  `bundle_id`     VARCHAR(36)  NOT NULL                    COMMENT '权限包 ID，FK → end_user_permission_bundles.id',
  `granted_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',

  PRIMARY KEY (`id`),
  -- 同一角色不可重复授予同一权限包
  UNIQUE KEY `uq_role_bundles_role_bundle`
    (`role_id`, `bundle_id`),
  INDEX `idx_role_bundles_org_project` (`org_name`, `project_slug`),
  INDEX `idx_role_bundles_bundle_id` (`bundle_id`),
  CONSTRAINT `fk_role_bundles_role`
    FOREIGN KEY (`role_id`) REFERENCES `project_roles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_role_bundles_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='角色-权限包 关联：角色持有哪些权限包';

-- -------------------------------------------------------------
-- 5. end_user_user_bundles — 用户直接授权-权限包 关联
-- -------------------------------------------------------------
CREATE TABLE `end_user_user_bundles` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT 'UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目',
  `user_id`       VARCHAR(36)  NOT NULL                    COMMENT '用户 ID，FK → users.id',
  `bundle_id`     VARCHAR(36)  NOT NULL                    COMMENT '权限包 ID，FK → end_user_permission_bundles.id',
  `granted_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',

  PRIMARY KEY (`id`),
  -- 同一用户在同一项目内不可重复获得同一权限包
  UNIQUE KEY `uq_user_bundles_org_project_user_bundle`
    (`org_name`, `project_slug`, `user_id`, `bundle_id`),
  INDEX `idx_user_bundles_bundle_id` (`bundle_id`),
  -- FK → users(id)
  CONSTRAINT `fk_user_bundles_user`
    FOREIGN KEY (`user_id`)
    REFERENCES `users` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_user_bundles_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='用户直接授权-权限包：绕过角色直接给用户授予权限包';
