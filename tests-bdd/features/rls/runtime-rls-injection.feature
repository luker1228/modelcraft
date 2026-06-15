# Runtime RLS 注入和行级过滤 BDD 测试
# 对应 PRD: ai-metadata/prd/rls/04-runtime-rls-injection.md
# 验收标准: AC-3, AC-4, AC-5

Feature: Runtime RLS 注入和行级过滤
  作为终端用户
  我只能访问属于自己的数据
  以便确保数据隔离

  Background:
    Given 存在已配置数据库集群的组织 "testorg" 和项目 "testproject"
    And 我以管理员身份登录
    And 已创建名为 "Orders" 的模型
    And 模型已有名为 "name" 格式为 "STRING" 的字段
    And 终端用户 "userA" 已登录
    And 终端用户 "userA" 创建了一条 "Orders" 记录，name 为 "OrderA"

  @smoke
  Scenario: EndUser 只能查询到自己的数据
    Given 终端用户 "userB" 已登录
    And 终端用户 "userB" 创建了一条 "Orders" 记录，name 为 "OrderB"
    When 终端用户 "userA" 查询 "Orders"
    Then 查询结果只包含 "OrderA"
    And 查询结果不包含 "OrderB"

  Scenario: EndUser 使用 where 条件试图访问他人数据返回空
    Given 终端用户 "userB" 已登录
    And 终端用户 "userB" 的用户 ID 为 "userB_id"
    And 终端用户 "userA" 查询 "Orders"，where 条件为 owner = "userB_id"
    Then 查询结果为空

  Scenario: 配置 RLS v2 using policy 后 runtime 只返回当前用户的数据
    Given 已创建名为 "ScopedOrdersUsing" 的模型
    And 模型已有名为 "name" 格式为 "STRING" 的字段
    And 终端用户 "userA" 已登录
    And 终端用户 "userA" 创建了一条 "ScopedOrdersUsing" 记录，name 为 "ScopedOrderA"
    And 终端用户 "userB" 已登录
    And 终端用户 "userB" 创建了一条 "ScopedOrdersUsing" 记录，name 为 "ScopedOrderB"
    And 我为该模型配置以下 RLS v2 policies:
      | policyName | action | role | usingExpr                          | withCheckExpr |
      | owner_read | read   |      | {"owner":{"_eq":{"_auth":"uid"}}} |               |
    When 终端用户 "userA" 查询 "ScopedOrdersUsing"
    Then 查询结果只包含 "ScopedOrderA"
    And 查询结果不包含 "ScopedOrderB"

  Scenario: 配置 RLS v2 with check 后 runtime 拒绝错误 owner 的创建
    Given 已创建名为 "ScopedOrdersWithCheck" 的模型
    And 模型已有名为 "name" 格式为 "STRING" 的字段
    And 我为该模型配置以下 RLS v2 policies:
      | policyName   | action | role | usingExpr | withCheckExpr                      |
      | owner_create | create |      |           | {"owner":{"_eq":{"_auth":"uid"}}} |
    And 终端用户 "userA" 已登录
    And 终端用户 "userB" 的用户 ID 为 "userB_id"
    When 终端用户 "userA" 创建一条 "ScopedOrdersWithCheck" 记录，name 为 "ScopedOrderBad"，owner 为 "userB_id"
    Then 操作失败并返回错误类型 "RLSCheckViolation"

  Scenario: EndUser 创建记录时 owner 自动填充为当前用户 ID
    Given 终端用户 "userA" 已登录
    When 终端用户 "userA" 创建一条 "Orders" 记录，name 为 "AutoOwnerOrder"
    Then 记录创建成功
    And 该记录的 owner 字段值为 "userA" 的用户 ID

  Scenario: EndUser 无法通过传入 owner 参数改变数据归属
    Given 终端用户 "userA" 已登录
    And 终端用户 "userB" 的用户 ID 为 "userB_id"
    When 终端用户 "userA" 创建一条 "Orders" 记录，name 为 "TryToSetOwner"，owner 为 "userB_id"
    Then 记录创建成功
    And 该记录的 owner 字段值为 "userA" 的用户 ID（被强制覆盖）

  Scenario: updatePredicate=OWNER 时无法更新他人数据
    Given 终端用户 "userB" 已登录
    And 终端用户 "userB" 创建了一条 "Orders" 记录，name 为 "OrderB"
    And "OrderB" 记录的实际 ID 为 "orderB_id"
    When 终端用户 "userA" 尝试更新 ID 为 "orderB_id" 的 "Orders" 记录
    Then 更新操作返回 0 行受影响（静默失败，无错误）

  Scenario: deletePredicate=OWNER 时无法删除他人数据
    Given 终端用户 "userB" 已登录
    And 终端用户 "userB" 创建了一条 "Orders" 记录，name 为 "OrderB"
    And "OrderB" 记录的实际 ID 为 "orderB_id"
    When 终端用户 "userA" 尝试删除 ID 为 "orderB_id" 的 "Orders" 记录
    Then 删除操作返回 0 行受影响（静默失败，无错误）

  Scenario: updateCheck=OWNER 时无法将 owner 改为他人
    Given 终端用户 "userB" 已登录
    And 终端用户 "userB" 的用户 ID 为 "userB_id"
    And 终端用户 "userA" 创建了一条 "Orders" 记录，name 为 "OrderA"
    And "OrderA" 记录的实际 ID 为 "orderA_id"
    When 终端用户 "userA" 尝试将 "orderA_id" 的 owner 改为 "userB_id"
    Then 操作失败并返回错误类型 "RLSCheckViolation"

  Scenario: 无 Policy 的模型 EndUser 访问被 DENY ALL
    Given 已创建名为 "NoPolicyModel" 的模型
    And 该模型不存在 RLS 策略
    And 终端用户 "userA" 已登录
    When 终端用户 "userA" 查询 "NoPolicyModel"
    Then 操作失败并返回错误类型 "PermissionDeniedRLS"
