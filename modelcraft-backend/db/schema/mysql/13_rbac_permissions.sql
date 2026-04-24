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
-- 1. end_user_permissions — 权限点
--    一行 = 一个「模型 × 动作 × 行范围」的细粒度权限
-- -------------------------------------------------------------
CREATE TABLE `end_user_permissions` (
  `id`            VARCHAR(36)  NOT NULL                    COMMENT '权限点 UUID',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织（冗余，不做 FK）',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目（冗余，不做 FK）',
  `model_id`      VARCHAR(36)  NOT NULL                    COMMENT '关联模型 ID，FK → models.id',
  `name`          VARCHAR(128) NOT NULL                    COMMENT '权限点名称，人类可读',
  `description`   TEXT         NULL                        COMMENT '权限点描述',
  `action`        ENUM(
                    'select',
                    'insert',
                    'update',
                    'delete',
                    'export'
                  )            NOT NULL                    COMMENT '操作动作',
  `column_policy` JSON         NULL                        COMMENT '列策略 JSON，结构见注释',
  -- column_policy 结构示例：
  -- {
  --   "defaultMode": "VISIBLE",          // VISIBLE | HIDDEN | MASKED
  --   "rules": [
  --     { "fieldName": "salary", "mode": "MASKED", "maskPattern": "***" },
  --     { "fieldName": "id_card", "mode": "HIDDEN" }
  --   ]
  -- }
  `row_scope`     ENUM(
                    'ALL',
                    'SELF',
                    'DEPT',
                    'DEPT_AND_CHILDREN'
                  )            NOT NULL DEFAULT 'ALL'      COMMENT '行范围：ALL全量/SELF本人/DEPT本部门/DEPT_AND_CHILDREN含子部门',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  -- 业务唯一键：同一模型下，相同动作+行范围+名称不可重复
  UNIQUE KEY `uq_permissions_model_action_scope_name`
    (`model_id`, `action`, `row_scope`, `name`),
  -- 按组织/项目快速检索（不做 FK，Project 不删除只归档）
  INDEX `idx_permissions_org_project` (`org_name`, `project_slug`),
  INDEX `idx_permissions_model_id` (`model_id`),
  -- FK → models.id（模型可删除，级联清理权限点）
  CONSTRAINT `fk_permissions_model`
    FOREIGN KEY (`model_id`) REFERENCES `models` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='权限点：每行描述对某模型某动作的行列级权限配置';

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
  `permission_id`  VARCHAR(36) NOT NULL                    COMMENT '权限点 ID，FK → end_user_permissions.id',
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
    FOREIGN KEY (`permission_id`) REFERENCES `end_user_permissions` (`id`)
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
  -- 复合 FK → end_user_users(org_name, project_slug, id)
  CONSTRAINT `fk_user_bundles_user`
    FOREIGN KEY (`org_name`, `project_slug`, `user_id`)
    REFERENCES `end_user_users` (`org_name`, `project_slug`, `id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_user_bundles_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='用户直接授权-权限包：绕过角色直接给用户授予权限包';