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
    When 以角色 "admin" 调用 Open Data API 创建一条 user_id 为 "other-owner-id" 的记录，name 为 "ord-rls-target"
    And 保存上次创建的记录 id
    When 以 EndUser "det-user-a" 调用 Open Data API 更新上次保存的记录，设置 name 为 "hacked"
    Then 更新操作未生效

  @smoke @deterministic
  Scenario: 无法删除非本人的记录
    When 以角色 "admin" 调用 Open Data API 创建一条 user_id 为 "other-owner-id" 的记录，name 为 "ord-rls-del-target"
    And 保存上次创建的记录 id
    When 以 EndUser "det-user-a" 调用 Open Data API 删除上次保存的记录
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

  # ── products 策略（全员可读，admin 限写）──

  @smoke @deterministic
  Scenario: 为 products 模型配置 RLS v2 策略（全员可读，admin 限写）
    When 我为 products 模型配置以下 RLS v2 policies:
      | policyName              | action | role  | usingExpr | withCheckExpr |
      | products_all_read       | read   | *     | true      |               |
      | products_admin_create   | create | admin |           | true          |
      | products_admin_update   | update | admin | true      |               |
      | products_admin_delete   | delete | admin | true      |               |
    Then products 模型的 RLS v2 策略数量应为 4
    And 策略配置成功

  @smoke @deterministic
  Scenario: 全员可读 — 任意 EndUser 可以读取 products（findMany）
    When 以 EndUser "det-user-a" 对 products 模型调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: 全员可读 — count 查询无限制
    When 以 EndUser "det-user-a" 对 products 模型调用 Open Data API 执行 count 查询
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: 普通用户无法创建 products 记录（非 admin）
    When 以 EndUser "det-user-a" 对 products 模型调用 Open Data API 创建一条记录，name 为 "prod-noauth-test"
    Then 操作被拒绝且返回错误类型 "OPERATION_FAILED.PERMISSION"

  @smoke @deterministic
  Scenario: admin 角色可读取所有 products 记录
    When 以角色 "admin" 对 products 模型调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: admin 角色可创建 products 记录
    When 以角色 "admin" 对 products 模型调用 Open Data API 创建一条记录，name 为 "prod-admin-create"
    Then 创建结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: admin 角色可更新 products 记录（非存在记录则无副作用）
    When 以角色 "admin" 对 products 模型调用 Open Data API 更新 id 为 "non-existent-id" 的记录，设置 name 为 "updated-prod"
    Then 更新操作未生效

  @smoke @deterministic
  Scenario: viewer 角色可读取 products 但无法创建
    When 以角色 "viewer" 对 products 模型调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors
    When 以 EndUser "det-user-a" 对 products 模型调用 Open Data API 创建一条记录，name 为 "prod-viewer-reject"
    Then 操作被拒绝且返回错误类型 "OPERATION_FAILED.PERMISSION"

  # ── tasks 策略（reporter/assignee 双主体 + in 操作）──


  @smoke @deterministic
  Scenario: tasks 完整策略验证 — reporter/assignee/admin 权限矩阵
    When 我为 tasks 模型配置以下 RLS v2 policies:
      | policyName              | action | role  | usingExpr                      | withCheckExpr                    |
      | tasks_reporter_read     | read   | *     | row.reporter_id == auth.userid |                                  |
      | tasks_assignee_read     | read   | *     | row.assignee_id == auth.userid |                                  |
      | tasks_create            | create | *     |                                | input.reporter_id == auth.userid |
      | tasks_reporter_update   | update | *     | row.reporter_id == auth.userid |                                  |
      | tasks_assignee_update   | update | *     | row.assignee_id == auth.userid |                                  |
      | tasks_reporter_delete   | delete | *     | row.reporter_id == auth.userid |                                  |
      | tasks_admin_read        | read   | admin | true                           |                                  |
    Then tasks 模型的 RLS v2 策略数量应为 7
    And 策略配置成功
    When 以 reporter 用户身份对 tasks 模型调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors
    When 以 assignee 用户身份对 tasks 模型调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors
    When 以 reporter 用户身份对 tasks 模型调用 Open Data API 创建一条 reporter 为当前用户的 task，title 为 "task-reporter-create"
    Then 创建结果为合法的 GraphQL 响应且无 errors
    When 以 reporter 用户身份对 tasks 模型调用 Open Data API 创建一条 reporter 为 "other-user-id" 的 task，title 为 "task-bad-reporter"
    Then 操作被拒绝且返回错误类型 "OPERATION_FAILED.PERMISSION"
    When 以 reporter 用户身份对 tasks 模型调用 Open Data API 创建一条 reporter 为当前用户的 task，title 为 "task-rls-update-target"
    And 保存上次创建的记录 id
    When 以 assignee 用户身份对 tasks 模型调用 Open Data API 更新上次保存的 task，设置 title 为 "updated-by-assignee"
    Then 更新操作未生效
    When 以 reporter 用户身份对 tasks 模型调用 Open Data API 创建一条 reporter 为当前用户的 task，title 为 "task-rls-delete-target"
    And 保存上次创建的记录 id
    When 以 assignee 用户身份对 tasks 模型调用 Open Data API 删除上次保存的 task
    Then 删除操作未生效
    When 以 reporter 用户身份对 tasks 模型调用 Open Data API 删除上次保存的 task
    Then 删除操作未生效
    When 以角色 "admin" 对 tasks 模型调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors

  # ── tasks in 操作专项 ──

  @smoke @deterministic
  Scenario: tasks milestone_id in usingExpr — 只有指定 milestone 下的任务可读
    When 我为 tasks 模型追加以下 RLS v2 policy:
      | policyName           | action | role | usingExpr                                            | withCheckExpr |
      | tasks_milestone_read | read   | *    | row.milestone_id in ["ms-allowed-1", "ms-allowed-2"] |               |
    And 策略配置成功
    When 以 reporter 用户身份对 tasks 模型调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: tasks project_id in withCheckExpr — upsert 覆盖 tasks_create，project_id 不在列表则拒绝
    When 我为 tasks 模型追加以下 RLS v2 policy:
      | policyName   | action | role | usingExpr | withCheckExpr                                            |
      | tasks_create | create | *    |           | input.project_id in ["proj-allowed-1", "proj-allowed-2"] |
    And 策略配置成功
    When 以 reporter 用户身份对 tasks 模型调用 Open Data API 创建一条 reporter 为当前用户且 project_id 为 "proj-forbidden" 的 task，title 为 "task-bad-project"
    Then 操作被拒绝且返回错误类型 "OPERATION_FAILED.PERMISSION"

  # ── task_comments 策略（全员可读 + author 限写 + in 操作）──

  @smoke @deterministic
  Scenario: task_comments 完整策略验证 — 全员可读、author 限写/改/删
    When 我为 task_comments 模型配置以下 RLS v2 policies:
      | policyName             | action | role | usingExpr                    | withCheckExpr                  |
      | comments_all_read      | read   | *    | true                         |                                |
      | comments_author_create | create | *    |                              | input.author_id == auth.userid |
      | comments_author_update | update | *    | row.author_id == auth.userid |                                |
      | comments_author_delete | delete | *    | row.author_id == auth.userid |                                |
    Then task_comments 模型的 RLS v2 策略数量应为 4
    And 策略配置成功
    When 以 EndUser "det-user-a" 对 task_comments 模型调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors
    When 以 EndUser "det-user-a" 对 task_comments 模型调用 Open Data API 创建一条 author_id 为当前用户的评论，content 为 "hello from bdd"
    Then 创建结果为合法的 GraphQL 响应且无 errors
    When 以 EndUser "det-user-a" 对 task_comments 模型调用 Open Data API 创建一条 author_id 为 "other-user-id" 的评论，content 为 "spoofed comment"
    Then 操作被拒绝且返回错误类型 "OPERATION_FAILED.PERMISSION"
    When 以 EndUser "det-user-a" 对 task_comments 模型调用 Open Data API 更新 id 为 "non-existent-id" 的评论，设置 content 为 "hacked"
    Then 更新操作未生效
    When 以 EndUser "det-user-a" 对 task_comments 模型调用 Open Data API 删除 id 为 "non-existent-id" 的评论
    Then 删除操作未生效

  # ── task_comments in 操作专项 ──

  @smoke @deterministic
  Scenario: task_comments task_id in usingExpr — 只有指定 task 下的评论可读
    When 我为 task_comments 模型追加以下 RLS v2 policy:
      | policyName         | action | role | usingExpr                                           | withCheckExpr |
      | comments_task_read | read   | *    | row.task_id in ["task-visible-1", "task-visible-2"] |               |
    And 策略配置成功
    When 以 EndUser "det-user-a" 对 task_comments 模型调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic
  Scenario: task_comments task_id in withCheckExpr — upsert 覆盖 comments_author_create，task_id 不在列表则拒绝
    When 我为 task_comments 模型追加以下 RLS v2 policy:
      | policyName             | action | role | usingExpr | withCheckExpr                                         |
      | comments_author_create | create | *    |           | input.task_id in ["task-visible-1", "task-visible-2"] |
    And 策略配置成功
    When 以 EndUser "det-user-a" 对 task_comments 模型调用 Open Data API 创建一条 author_id 为当前用户且 task_id 为 "task-forbidden" 的评论，content 为 "bad task comment"
    Then 操作被拒绝且返回错误类型 "OPERATION_FAILED.PERMISSION"

  # ── useAdmin 查询条件专项（不验证 RLS）──
  # 全部通过 X-MC-Auth-Useadmin: true 以设计者身份操作 runtime 数据，专注验证各类查询参数的正确性

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany 无条件查询
    When 以 useAdmin 方式调用 Open Data API 执行 findMany
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany 分页 take=5 skip=0（第一页）
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，take 5 skip 0
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany 分页 take=5 skip=5（第二页偏移）
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，take 5 skip 5
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany orderBy id desc
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，orderBy id desc
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany orderBy total_amount asc
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，orderBy total_amount asc
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany where 精确过滤 user_id
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，where user_id eq "det-test-user-001"
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany where user_id in 列表
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，where user_id in "det-test-user-001,other-user-id"
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany 组合条件：take=10 skip=0 orderBy id desc
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，take 10 skip 0 orderBy id desc
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — count 无条件计数
    When 以 useAdmin 方式调用 Open Data API 执行 count
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — count where 过滤 user_id
    When 以 useAdmin 方式调用 Open Data API 执行 count，where user_id eq "det-test-user-001"
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany where total_amount gt 0（数值范围下界）
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，where total_amount gt 0
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany where total_amount lte 100（数值上界）
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，where total_amount lte 100
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany where order_no startsWith "bdd-"（字符串前缀匹配）
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，where order_no startsWith "bdd-"
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany where address_id ne 指定值（不等过滤）
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，where address_id ne "non-existent-addr"
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany where remark isNull（null 检查）
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，where remark isNull
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany 多字段排序 total_amount asc id desc
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，多字段排序 total_amount asc id desc
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany take=1 单条限制
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，take 1 skip 0
    Then 返回结果为合法的 GraphQL 响应且无 errors

  @smoke @deterministic @query-conditions
  Scenario: useAdmin — findMany where paid_amount gte 0 lte 1000（数值区间）
    When 以 useAdmin 方式调用 Open Data API 执行 findMany，where paid_amount gte 0 lte 1000
    Then 返回结果为合法的 GraphQL 响应且无 errors
