-- =============================================================
-- 13_rbac_permissions.sql
-- RBAC 行列级权限系统
-- 依赖: 12_end_user_auth.sql（end_user_roles / end_user_users）
-- =============================================================

-- -------------------------------------------------------------
-- 0. ALTER 现有表：end_user_roles 追加 is_implicit 列
-- -------------------------------------------------------------
ALTER TABLE `end_user_roles`
  ADD COLUMN `is_implicit` TINYINT(1) NOT NULL DEFAULT 0
    COMMENT '内置隐式角色标志：0=显式角色（用户手动分配），1=隐式角色（系统自动注入）',
  ADD INDEX `idx_end_user_roles_implicit` (`is_implicit`);

-- -------------------------------------------------------------
-- 1. end_user_data_permissions — 权限点
--    一行 = 一条对某模型的行列级访问策略
-- -------------------------------------------------------------
CREATE TABLE `end_user_data_permissions` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT '权限点 UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织（冗余，不做 FK）',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目（冗余，不做 FK）',
  `database_name` VARCHAR(128) NULL                        COMMENT '数据源名称（可空，预留按数据源授权能力）',
  `model_name`    VARCHAR(128) NULL                        COMMENT '模型名称（可空，预留按模型名授权能力）',
  `model_id`      VARCHAR(36)  NOT NULL                    COMMENT '关联模型 ID，FK → models.id',
  `name`          VARCHAR(128) NOT NULL                    COMMENT '权限点名称，人类可读',
  `description`   TEXT         NULL                        COMMENT '权限点描述',
  `type`          ENUM(
                    'PRESET',
                    'CUSTOM'
                  )            NOT NULL DEFAULT 'CUSTOM'   COMMENT '权限点来源：PRESET=预设策略生成，CUSTOM=管理员手动创建',
  `column_policy` JSON         NULL                        COMMENT '列策略 JSON，结构见注释',
  -- column_policy 结构示例：
  -- {
  --   "defaultMode": "VISIBLE",          // VISIBLE | HIDDEN | MASKED
  --   "rules": [
  --     { "fieldName": "salary",  "mode": "MASKED",  "maskPattern": "***", "writable": false },
  --     { "fieldName": "id_card", "mode": "HIDDEN",  "writable": false },
  --     { "fieldName": "status",  "mode": "VISIBLE", "writable": false }
  --   ]
  -- }
  -- writable: 该列是否允许被 UPDATE SET，默认 true；仅 mode=VISIBLE 时有意义
  `row_policy`    JSON         NULL                        COMMENT '行策略 JSON，谓词为 GraphQL Runtime where 条件（与 model_rls_policies 格式一致）',
  -- row_policy 结构（每个 action 两个字段）：
  --   allowed : boolean     是否允许该操作；false 时其余字段忽略
  --   scope   : "all"|"custom"  all=全通行(predicate/check忽略)，custom=条件生效
  --   predicate/check : GraphQL Runtime where JSON（scope=custom 时有效）
  --
  -- 隐藏规则：select.allowed=false 时，insert/update/delete 全部强制 allowed=false
  --
  -- {
  --   "select": {
  --     "allowed": true,
  --     "scope": "custom",
  --     "predicate": { "status": { "_eq": "active" } }
  --   },
  --   "insert": {
  --     "allowed": true,
  --     "scope": "custom",
  --     "check": { "owner_id": { "_eq": "$endUserId" } }
  --   },
  --   "update": {
  --     "allowed": true,
  --     "scope": "custom",
  --     "predicate": { "owner_id": { "_eq": "$endUserId" } },
  --     "check_scope": "custom",
  --     "check": { "owner_id": { "_eq": "$endUserId" } }
  --   },
  --   "delete": {
  --     "allowed": false
  --   }
  -- }
  `preset`        ENUM(
                    'READ_WRITE_ALL',
                    'READ_ALL',
                    'READ_WRITE_OWNER',
                    'READ_ALL_WRITE_OWNER'
                  )            NULL DEFAULT NULL           COMMENT '来源预设，NULL 表示手动创建的自定义权限点；仅 type=PRESET 时有值',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  -- 业务唯一键：同一模型下，相同类型+名称不可重复
  UNIQUE KEY `uq_permissions_model_type_name`
    (`model_id`, `type`, `name`),
  -- 按组织/项目快速检索（不做 FK，Project 不删除只归档）
  INDEX `idx_permissions_org_project` (`org_name`, `project_slug`),
  INDEX `idx_permissions_model_id` (`model_id`),
  INDEX `idx_permissions_model_preset` (`model_id`, `preset`),
  -- FK → models.id（模型可删除，级联清理权限点）
  CONSTRAINT `fk_permissions_model`
    FOREIGN KEY (`model_id`) REFERENCES `models` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='权限点：每行描述对某模型的行列级访问策略';

-- -------------------------------------------------------------
-- 2. end_user_permission_bundles — 权限包
--    权限点的命名集合，可跨模型聚合
-- -------------------------------------------------------------
CREATE TABLE `end_user_permission_bundles` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT '权限包 UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目',
  `name`          VARCHAR(128) NOT NULL                    COMMENT '权限包名称',
  `description`   TEXT         NULL                        COMMENT '权限包描述',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  -- 同一项目下权限包名称唯一
  UNIQUE KEY `uq_bundles_org_project_name`
    (`org_name`, `project_slug`, `name`),
  -- 快速检索（不做 FK → projects）
  INDEX `idx_bundles_org_project` (`org_name`, `project_slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='权限包：权限点的命名集合，用于角色授权或用户直接授权';

-- -------------------------------------------------------------
-- 3. end_user_bundle_permissions — 权限包-权限点 有序中间表
-- -------------------------------------------------------------
CREATE TABLE `end_user_bundle_permissions` (
  `id`             VARCHAR(36) NOT NULL                    COMMENT 'UUID',
  `bundle_id`      VARCHAR(36) NOT NULL                    COMMENT '权限包 ID，FK → end_user_permission_bundles.id',
  `permission_id`  VARCHAR(36) NOT NULL                    COMMENT '权限点 ID，FK → end_user_data_permissions.id',
  `sort_order`     INT         NOT NULL DEFAULT 0          COMMENT '显示排序权重（ASC）',
  `created_at`     DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  -- 同一权限包内权限点不重复
  UNIQUE KEY `uq_bundle_permissions_bundle_perm`
    (`bundle_id`, `permission_id`),
  INDEX `idx_bundle_permissions_permission_id` (`permission_id`),
  CONSTRAINT `fk_bundle_permissions_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_bundle_permissions_permission`
    FOREIGN KEY (`permission_id`) REFERENCES `end_user_data_permissions` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='权限包-权限点 有序中间表';

-- -------------------------------------------------------------
-- 4. end_user_role_bundles — 角色-权限包 关联
-- -------------------------------------------------------------
CREATE TABLE `end_user_role_bundles` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT 'UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织（冗余，快速查询）',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目（冗余，快速查询）',
  `role_id`       VARCHAR(36)  NOT NULL                    COMMENT '角色 ID，FK → end_user_roles.id',
  `bundle_id`     VARCHAR(36)  NOT NULL                    COMMENT '权限包 ID，FK → end_user_permission_bundles.id',
  `granted_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',

  PRIMARY KEY (`id`),
  -- 同一角色不可重复授予同一权限包
  UNIQUE KEY `uq_role_bundles_role_bundle`
    (`role_id`, `bundle_id`),
  INDEX `idx_role_bundles_org_project` (`org_name`, `project_slug`),
  INDEX `idx_role_bundles_bundle_id` (`bundle_id`),
  CONSTRAINT `fk_role_bundles_role`
    FOREIGN KEY (`role_id`) REFERENCES `end_user_roles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_role_bundles_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='角色-权限包 关联：角色持有哪些权限包';

-- -------------------------------------------------------------
-- 5. end_user_user_bundles — 用户直接授权-权限包 关联
--    用户可绕过角色直接获得权限包（最高优先级通道）
-- -------------------------------------------------------------
CREATE TABLE `end_user_user_bundles` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT 'UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目',
  `user_id`       VARCHAR(36)  NOT NULL                    COMMENT '用户 ID（复合 FK 的 id 部分）',
  `bundle_id`     VARCHAR(36)  NOT NULL                    COMMENT '权限包 ID，FK → end_user_permission_bundles.id',
  `granted_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',

  PRIMARY KEY (`id`),
  -- 同一用户在同一项目内不可重复获得同一权限包
  UNIQUE KEY `uq_user_bundles_org_project_user_bundle`
    (`org_name`, `project_slug`, `user_id`, `bundle_id`),
  INDEX `idx_user_bundles_bundle_id` (`bundle_id`),
  -- FK → end_user_users(org_name, id)
  CONSTRAINT `fk_user_bundles_user`
    FOREIGN KEY (`org_name`, `user_id`)
    REFERENCES `end_user_users` (`org_name`, `id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_user_bundles_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='用户直接授权-权限包：绕过角色直接给用户授予权限包';