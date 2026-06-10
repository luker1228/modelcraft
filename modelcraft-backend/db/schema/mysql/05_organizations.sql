-- =============================================================================
-- 组织管理 (Organization Management)
-- 多租户组织单元，每个组织是项目、集群、模型等资源的逻辑容器
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 组织表 (Organizations)
-- 组织名称全局唯一，通常来自 AuthProvider
-- owner_id 在用户创建后会填充，初始创建时可能为空
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `organizations` (
  `name` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '唯一标识符，随机slug',
  `display_name` VARCHAR(255) COMMENT '用于 UI 显示的名称',
  `owner_id` VARCHAR(36) COMMENT '组织创建者/所有者（引用 users.id）',
  `phone` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'Org 注册手机号，全局唯一',
  `status` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态：active、suspended、deleted',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',

  UNIQUE INDEX `uk_org_phone` (`phone`, `delete_token`) COMMENT 'Org 手机号全局唯一',
  INDEX `idx_org_owner` (`owner_id`) COMMENT '按所有者查找组织',
  INDEX `idx_org_status` (`status`) COMMENT '按状态筛选',
  INDEX `idx_org_live_status` (`deleted_at`, `status`) COMMENT '按活跃状态筛选组织'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织表（多租户容器）';
