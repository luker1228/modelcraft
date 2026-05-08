-- =============================================================================
-- Casbin 权限系统 (Casbin RBAC Permission System)
-- 包含：角色表、用户角色绑定表、角色权限表
--
-- 设计要点：
-- 1. 系统角色 (owner/admin/editor/viewer) 权限硬编码在 Go 代码中
-- 2. 自定义角色权限存储在 role_permissions 表中
-- 3. 支持多租户：同一用户在不同 org 可以有不同角色
-- 4. 使用 org_name='__SYSTEM__' 标识系统角色（而非 NULL，避免唯一约束问题）
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Step 2: 创建角色表 (Roles)
--
-- 设计变更：
-- - org_name 使用 '__SYSTEM__' 标识系统角色（全局），其他值为租户自定义角色
-- - 移除 permissions JSON 字段（系统角色权限硬编码，自定义角色权限在 role_permissions 表）
-- - 主键改为 INT AUTO_INCREMENT（更适合 Casbin enforcer 和外键关联）
-- - is_system=TRUE 的角色不可修改删除
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `roles` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '角色 ID（主键）',
  `name` VARCHAR(64) NOT NULL COMMENT '角色名称（如 owner, admin, editor, viewer, 或自定义）',
  `description` TEXT COMMENT '角色描述',
  `is_system` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '系统角色标识（true = 不可修改删除）',
  `org_name` VARCHAR(36) NOT NULL DEFAULT '__SYSTEM__' COMMENT '所属组织名称（__SYSTEM__ = 系统角色，其他 = 租户自定义角色）',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',

  -- 唯一约束：租户内角色名称唯一（系统角色使用 __SYSTEM__ 标识，全局唯一）
  UNIQUE KEY `uk_role_name_org` (`name`, `org_name`, `delete_token`) COMMENT '租户内角色名称唯一',

  -- 索引
  INDEX `idx_org_name` (`org_name`) COMMENT '按组织查询角色',
  INDEX `idx_is_system` (`is_system`) COMMENT '筛选系统角色',
  INDEX `idx_roles_live_org` (`org_name`, `deleted_at`) COMMENT '组织活跃角色查询索引'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表（支持系统角色和租户自定义角色）';

-- -----------------------------------------------------------------------------
-- Step 3: 系统角色初始化
--
-- 注意：
-- - 系统角色的初始化由应用代码负责（启动时自动创建）
-- - 系统角色的权限在 Go 代码中硬编码（internal/infrastructure/auth/system_roles.go）
-- - org_name = '__SYSTEM__' 表示全局系统角色
-- - is_system = TRUE 防止被删除或修改
-- -----------------------------------------------------------------------------

-- 系统角色列表（由应用代码创建）:
-- - owner: 拥有者 - 完全控制所有资源和团队管理
-- - admin: 管理员 - 管理所有资源和团队成员（不含用户管理）
-- - editor: 编辑者 - 创建和编辑资源（不能删除）
-- - viewer: 查看者 - 只读访问资源

-- -----------------------------------------------------------------------------
-- Step 4: 创建用户角色绑定表 (User Roles)
--
-- 设计变更：
-- - 简化为 user_id, role_id, org_name 三字段核心模型
-- - 移除 status, invited_by 等邀请流程字段（邀请流程在 user_organizations 表处理）
-- - role_id 外键级联删除：删除角色时自动删除所有用户角色绑定
-- - 支持同一用户在不同 org 有不同角色
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `user_roles` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键 ID',
  `user_id` VARCHAR(36) NOT NULL COMMENT '用户 ID（引用 users.id）',
  `role_id` BIGINT NOT NULL COMMENT '角色 ID（引用 roles.id）',
  `org_name` VARCHAR(36) NOT NULL COMMENT '组织名称（多租户隔离）',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',

  -- 外键约束
  CONSTRAINT `fk_user_roles_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_roles_role` FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE CASCADE,

  -- 唯一约束：同一用户在同一组织不能重复绑定同一角色
  UNIQUE KEY `uk_user_role_org` (`user_id`, `role_id`, `org_name`) COMMENT '防止重复角色绑定',

  -- 索引
  INDEX `idx_user_org` (`user_id`, `org_name`) COMMENT '查询用户在某组织的所有角色',
  INDEX `idx_role_org` (`role_id`, `org_name`) COMMENT '查询某角色在某组织的所有用户',
  INDEX `idx_org_name` (`org_name`) COMMENT '按组织查询所有用户角色'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色绑定表（支持多租户）';

-- -----------------------------------------------------------------------------
-- Step 5: 创建角色权限表 (Role Permissions)
--
-- 说明：
-- - 仅存储租户自定义角色的权限（系统角色权限硬编码在 Go 代码中）
-- - obj: 资源名称（如 project, model, cluster, enum）
-- - act: 操作名称（如 create, read, update, delete, 或 * 表示所有）
-- - org_name: 冗余字段，用于优化查询（虽然可以通过 role_id JOIN 获取，但冗余提升性能）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `role_permissions` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键 ID',
  `role_id` BIGINT NOT NULL COMMENT '角色 ID（引用 roles.id）',
  `org_name` VARCHAR(36) NOT NULL COMMENT '组织名称（冗余字段，优化查询）',
  `obj` VARCHAR(64) NOT NULL COMMENT '资源对象（如 project, model, cluster, enum）',
  `act` VARCHAR(64) NOT NULL COMMENT '操作动作（如 create, read, update, delete, *）',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',

  -- 外键约束
  CONSTRAINT `fk_role_perms_role` FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE CASCADE,

  -- 唯一约束：同一角色不能有重复的权限规则
  UNIQUE KEY `uk_role_obj_act` (`role_id`, `obj`, `act`) COMMENT '防止重复权限',

  -- 索引
  INDEX `idx_role_org` (`role_id`, `org_name`) COMMENT '查询某角色在某组织的权限（优化主查询路径）',
  INDEX `idx_org_name` (`org_name`) COMMENT '按组织查询权限'

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限表（仅存储自定义角色权限）';

-- =============================================================================
-- 设计完成！
--
-- 下一步：
-- 1. 在 Go 代码中实现 Casbin enforcer 和系统角色权限硬编码
-- 2. 实现 GraphQL @hasPermission 指令
-- 3. 编写测试用例验证权限系统
--
-- 关键设计决策：
-- - 使用 '__SYSTEM__' 而非 NULL 避免 MySQL 唯一约束的 NULL 处理问题
-- - 系统角色权限硬编码提升性能且保证不可变性
-- - role_permissions 中冗余 org_name 优化查询性能
-- - 邀请流程字段保留在 user_organizations 表，角色绑定表保持简洁
-- =============================================================================
