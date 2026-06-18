# 确定性 RLS 运行时拨测
# 使用预置的 org / project / db / model，不创建模型，仅聚焦 RLS 策略配置与 Open Data API 运行时验证
# 参考: debug/enduser_runtime_rls_check.sh
#
# 环境变量 (.env.test):
#   DET_ORG_NAME      — 已有 Org 名称（默认取 TEST_ORG_NAME）
#   DET_PROJECT_SLUG  — 已有 Project slug（默认取 TEST_PROJECT_SLUG）
#   DET_DB_NAME       — 已有数据库名称
#   DET_MODEL_NAME    — 已有模型名称（必须含 owner 字段）
#   DET_PAT           — PAT token，用于 Open Data API 认证
#   DET_END_USER_ID   — 测试终端用户 ID（传给 X-MC-Auth-Userid-Str）
#   DET_END_USER_NAME — 测试终端用户名（传给 X-MC-Auth-Username）

Feature: 确定性 RLS 运行时拨测
  作为运维 / QA
  我希望对已存在的模型进行 RLS 策略配置与 Open Data API 运行时拨测
  以便快速验证 RLS 行级安全在端用户侧的生效情况

  Background:
    Given 我以管理员身份登录
    And 确定性拨测环境已就绪

  # ── 策略配置 ──

  @smoke @deterministic
  Scenario: 为确定性拨测模型配置 RLS v2 策略
    When 我为确定性拨测模型配置以下 RLS v2 policies:
      | policyName    | action | role | usingExpr                     | withCheckExpr                 |
      | orders_user_read    | read   | *    | row.user_id == auth.userid    |                               |
      | orders_user_create  | create | *    |                               | input.user_id == auth.userid  |
      | orders_user_update  | update | *    | row.user_id == auth.userid    | input.user_id == auth.userid  |
      | orders_user_delete  | delete | *    | row.user_id == auth.userid    |                               |
      | orders_admin_read   | read   | admin| true                          |                               |
      | orders_admin_create | create | admin|                               | true                          |
    Then 确定性拨测模型的 RLS v2 策略数量应为 6
    And 策略配置成功

  # ── 运行时读 ──

  @smoke @deterministic
  Scenario: findMany 查询受 RLS 约束 — EndUser 只能看到自己的数据
    When 以 EndUser "det-user-a" 调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: count 查询受 RLS 约束
    When 以 EndUser "det-user-a" 调用 Open Data API 执行 count 查询
    Then 返回结果为合法的 GraphQL 响应且无 errors

  # ── 运行时写 ──

  @smoke @deterministic
  Scenario: 创建记录时 user_id 自动等于当前 EndUser
    When 以 EndUser "det-user-a" 调用 Open Data API 创建一条 user_id 为当前用户的记录，name 为 "ord-own-test"
    Then 创建结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: 无法创建 user_id 为其他人的记录
    When 以 EndUser "det-user-a" 调用 Open Data API 创建一条 user_id 为 "other-user-id" 的记录，name 为 "ord-bad-owner"
    Then 操作被拒绝且返回错误类型 "OPERATION_FAILED.PERMISSION"

  @smoke @deterministic
  Scenario: 无法更新非本人的记录
    When 以 EndUser "det-user-a" 调用 Open Data API 更新 id 为 "non-existent-id" 的记录，设置 name 为 "hacked"
    Then 更新操作未生效

  @smoke @deterministic
  Scenario: 无法删除非本人的记录
    When 以 EndUser "det-user-a" 调用 Open Data API 删除 id 为 "non-existent-id" 的记录
    Then 删除操作未生效

  # ── 角色测试 ──

  @smoke @deterministic
  Scenario: admin 角色可以读取所有数据
    When 以角色 "admin" 调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: viewer 角色可以读取且可以创建自己的记录
    When 以角色 "viewer" 调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors
    When 以角色 "viewer" 调用 Open Data API 创建一条 user_id 为当前用户的记录，name 为 "ord-viewer-test"
    Then 创建结果为合法的 GraphQL 响应且无 errors

  # ── orders 策略（user_id 字段）──

  @smoke @deterministic
  Scenario: 验证 orders 表 RLS 策略已就绪（user_id 字段）
    Then 确定性拨测模型的 RLS v2 策略总数（det_ 和 orders_ 前缀）应为 6

  @smoke @deterministic
  Scenario: user_id 字段 — EndUser 只能读取自己的 orders 记录
    When 以 EndUser "det-user-a" 调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: user_id 字段 — 创建 orders 记录时 user_id 必须等于当前用户
    When 以 EndUser "det-user-a" 调用 Open Data API 创建一条 user_id 为当前用户的记录，name 为 "orders-own-test"
    Then 创建结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: user_id 字段 — 创建 orders 记录时 user_id 为他人则拒绝
    When 以 EndUser "det-user-a" 调用 Open Data API 创建一条 user_id 为 "other-user-id" 的记录，name 为 "orders-bad-owner"
    Then 操作被拒绝且返回错误类型 "OPERATION_FAILED.PERMISSION"

  @smoke @deterministic
  Scenario: user_id 字段 — admin 角色可读取所有 orders 记录
    When 以角色 "admin" 调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: user_id 字段 — admin 角色可创建任意 user_id 的 orders 记录
    When 以角色 "admin" 调用 Open Data API 创建一条 user_id 为 "any-user-id" 的记录，name 为 "orders-admin-create"
    Then 创建结果为合法的 GraphQL 响应且无 errors
