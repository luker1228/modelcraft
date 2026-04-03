-- =============================================================================
-- 集成测试用户初始化脚本
-- 创建测试用户并分配 owner 角色
-- =============================================================================
--
-- 用途：为集成测试自动创建测试用户并分配完整权限
-- 使用场景：
--   1. 集成测试开始前自动执行（通过 pytest fixture）
--   2. 手动执行：mysql -h localhost -u root -p modelcraft < tests/setup_test_user.sql
--
-- 前置条件：
--   - 应用已启动并初始化系统角色（owner, admin, editor, viewer）
--   - 脚本会自动创建 modelcraft 组织（如果不存在）
--
-- 测试用户信息（配置于 .env 文件）：
--   CASDOOR_TEST_USERNAME=test-integration
--   CASDOOR_ORGANIZATION=modelcraft
--
-- 特性：
--   - 幂等性：可重复执行，不会产生重复数据
--   - 自动创建：如果用户或组织不存在则创建
--   - 自动分配角色：分配 owner 角色到 modelcraft 组织
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. 创建测试用户（如果不存在）
-- -----------------------------------------------------------------------------

-- 测试用户固定 UUID（基于用户名生成，保证幂等性）
SET @test_user_id = '487101d6-92bb-459e-b4f1-426255126d27';
SET @test_external_id = 'test-integration';
SET @test_user_name = 'Test Integration User';
SET @org_name = 'modelcraft';

-- 确保 modelcraft 组织存在（测试环境必需）
INSERT IGNORE INTO `organizations` (`id`, `name`, `display_name`, `status`)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    @org_name,
    'ModelCraft',
    'active'
);

-- 插入用户记录（使用 INSERT IGNORE 避免重复）
INSERT IGNORE INTO `users` (`id`, `external_id`, `name`, `phone`)
VALUES (
    @test_user_id,
    @test_external_id,
    @test_user_name,
    ''
);

-- -----------------------------------------------------------------------------
-- 2. 查找组织和角色 ID
-- -----------------------------------------------------------------------------

-- 获取 modelcraft 组织 ID
SET @org_id = (
    SELECT id FROM `organizations`
    WHERE name = @org_name
    LIMIT 1
);

-- 获取 owner 角色 ID
SET @owner_role_id = (
    SELECT id FROM `roles`
    WHERE name = 'owner' AND org_name = '__SYSTEM__'
    LIMIT 1
);

-- 获取 admin 角色 ID
SET @admin_role_id = (
    SELECT id FROM `roles`
    WHERE name = 'admin' AND org_name = '__SYSTEM__'
    LIMIT 1
);

-- 检查必需数据是否存在
SELECT
    CASE
        WHEN @org_id IS NULL THEN '❌ 错误：组织 "modelcraft" 不存在，请先运行数据库迁移'
        WHEN @owner_role_id IS NULL THEN '❌ 错误：角色 "owner" 不存在，请先运行数据库迁移'
        WHEN @admin_role_id IS NULL THEN '❌ 错误：角色 "admin" 不存在，请先运行数据库迁移'
        ELSE '✅ 组织和角色验证通过'
    END AS status;

-- -----------------------------------------------------------------------------
-- 3. 分配测试用户到组织（不含角色）
-- -----------------------------------------------------------------------------

-- 将测试用户加入组织（使用 INSERT IGNORE 避免重复）
INSERT IGNORE INTO `user_organizations` (
    `id`,
    `user_id`,
    `org_id`,
    `status`,
    `joined_at`
)
SELECT
    CONCAT('uo-', @test_user_id) AS id,
    @test_user_id AS user_id,
    @org_id AS org_id,
    'active' AS status,
    NOW(3) AS joined_at
WHERE @org_id IS NOT NULL
  -- 只在用户尚未加入组织时插入
  AND NOT EXISTS (
      SELECT 1 FROM `user_organizations` uo
      WHERE uo.user_id = @test_user_id
        AND uo.org_id = @org_id
  );

-- -----------------------------------------------------------------------------
-- 4. 分配 owner 角色到测试用户
-- -----------------------------------------------------------------------------

-- 为测试用户分配 owner 角色（使用 INSERT IGNORE 避免重复）
INSERT IGNORE INTO `user_roles` (
    `user_id`,
    `role_id`,
    `org_name`
)
SELECT
    @test_user_id AS user_id,
    @owner_role_id AS role_id,
    @org_name AS org_name
WHERE @org_id IS NOT NULL
  AND @owner_role_id IS NOT NULL
  -- 只在用户尚未分配此角色时插入
  AND NOT EXISTS (
      SELECT 1 FROM `user_roles` ur
      WHERE ur.user_id = @test_user_id
        AND ur.role_id = @owner_role_id
        AND ur.org_name = @org_name
  );

-- 为测试用户分配 admin 角色（使用 INSERT IGNORE 避免重复）
INSERT IGNORE INTO `user_roles` (
    `user_id`,
    `role_id`,
    `org_name`
)
SELECT
    @test_user_id AS user_id,
    @admin_role_id AS role_id,
    @org_name AS org_name
WHERE @org_id IS NOT NULL
  AND @admin_role_id IS NOT NULL
  -- 只在用户尚未分配此角色时插入
  AND NOT EXISTS (
      SELECT 1 FROM `user_roles` ur
      WHERE ur.user_id = @test_user_id
        AND ur.role_id = @admin_role_id
        AND ur.org_name = @org_name
  );

-- -----------------------------------------------------------------------------
-- 5. 验证结果
-- -----------------------------------------------------------------------------

SELECT
    '✅ 测试用户设置完成' AS message,
    u.id AS user_id,
    u.external_id,
    u.name AS user_name,
    o.name AS org_name,
    r.name AS role_name,
    uo.status AS org_status,
    ur.created_at AS role_assigned_at
FROM users u
JOIN user_organizations uo ON u.id = uo.user_id
JOIN organizations o ON uo.org_id = o.id
JOIN user_roles ur ON u.id = ur.user_id AND o.name = ur.org_name
JOIN roles r ON r.id = ur.role_id
WHERE u.id = @test_user_id;

-- 如果上述查询无结果，说明设置失败
SELECT
    CASE
        WHEN (SELECT COUNT(*) FROM user_organizations WHERE user_id = @test_user_id) = 0
        THEN '⚠️  警告：测试用户未加入组织，请检查组织是否存在'
        WHEN (SELECT COUNT(*) FROM user_roles WHERE user_id = @test_user_id AND org_name = @org_name) = 0
        THEN '⚠️  警告：测试用户角色分配失败，请检查角色是否存在'
        ELSE NULL
    END AS warning;

-- -----------------------------------------------------------------------------
-- 清理脚本（仅用于开发/调试，生产环境请勿使用）
-- -----------------------------------------------------------------------------

-- 取消注释以下语句可删除测试用户和关联数据
-- 注意：这将删除测试用户创建的所有项目、模型等数据

-- -- 删除测试用户的角色绑定
-- DELETE FROM user_roles WHERE user_id = @test_user_id;
--
-- -- 删除测试用户的组织关联
-- DELETE FROM user_organizations WHERE user_id = @test_user_id;
--
-- -- 删除测试用户
-- DELETE FROM users WHERE id = @test_user_id;
--
-- SELECT '🧹 测试用户已清理' AS message;
