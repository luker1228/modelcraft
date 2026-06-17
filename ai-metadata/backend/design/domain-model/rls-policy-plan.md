# RLS 测试策略规划

> 目标：覆盖尽可能多的 RLS 表达式类型、角色组合、操作场景，为 BDD 拨测提供完整的测试矩阵。
> 环境：luke_e6kz / project luke，两个数据库 demo_ecommerce / demo_pm。

---

## 测试场景矩阵

| # | 场景 | 模型 | 策略类型 | 预期结果 |
|---|------|------|---------|---------|
| T1 | 无角色用户只能读自己的行 | orders | `row.user_id == auth.userid` | 只返回 user_id 匹配行 |
| T2 | 创建时 withCheckExpr 校验 owner | orders | `input.user_id == auth.userid` | owner 匹配则成功，否则 RLSCheckViolation |
| T3 | 更新时 using + withCheck 双重校验 | orders | using + withCheck 同字段 | 只能改自己的行且不能换 owner |
| T4 | 删除时只能删自己的行 | orders | `row.user_id == auth.userid` | 非 owner 行返回 0 affected |
| T5 | admin 角色绕过行过滤读全量 | orders | `role=admin, usingExpr=true` | admin 可见所有行 |
| T6 | viewer 角色只读，写操作被拒 | orders | viewer 无写策略 | create 返回 RLSCheckViolation |
| T7 | `true` 表达式——全员可读 | products | `usingExpr=true` | 任意 EndUser 均可 findMany |
| T8 | 纯角色写控制（admin only write） | products | `role=admin, withCheckExpr=true` | 非 admin 写入被拒 |
| T9 | OR 表达式——双字段任一匹配 | tasks | `row.reporter_id == auth.userid \|\| row.assignee_id == auth.userid` | reporter 或 assignee 均可读 |
| T10 | 创建者和负责人权限不对称 | tasks | reporter 可删，assignee 不可删 | assignee 删除返回 0 affected |
| T11 | 评论：全员可读 + 作者限写 | task_comments | read=true, create/update/delete by author_id | 任何人能读，只有作者能改 |
| T12 | self-read：id == auth.userid | members | `row.id == auth.userid` | 只能查到自己的成员记录 |
| T13 | 无策略模型返回 403 | milestones（未配置） | — | 所有请求 403 |
| T14 | count 受 RLS 过滤 | orders | 同 T1 | count 只统计自己的行数 |
| T15 | 角色叠加：同时命中多条策略 | orders | user 策略 + admin 策略均存在 | admin 走 admin 策略，普通用户走 user 策略 |

---

## 策略配置

### orders（demo_ecommerce）— 核心 CRUD 测试基准

字段：`id, order_no, user_id, total_amount, paid_amount, remark, created_at`

| policyName | action | role | usingExpr | withCheckExpr | 覆盖测试 |
|------------|--------|------|-----------|---------------|---------|
| orders_user_read | read | | `row.user_id == auth.userid` | | T1, T14 |
| orders_user_create | create | | | `input.user_id == auth.userid` | T2 |
| orders_user_update | update | | `row.user_id == auth.userid` | `input.user_id == auth.userid` | T3 |
| orders_user_delete | delete | | `row.user_id == auth.userid` | | T4 |
| orders_admin_read | read | admin | `true` | | T5, T15 |
| orders_admin_create | create | admin | | `true` | T8（借用） |

> `viewer` 角色无额外策略 → viewer 用 orders_user_* 策略（无角色命中所有人），viewer 写入时因 create 策略只有 withCheckExpr 约束，需要 input.user_id 正确才能通过。

### products（demo_ecommerce）— 公开读 + 角色写

字段：`id, name, description, price, stock, category_id`

| policyName | action | role | usingExpr | withCheckExpr | 覆盖测试 |
|------------|--------|------|-----------|---------------|---------|
| products_all_read | read | | `true` | | T7 |
| products_admin_create | create | admin | | `true` | T8 |
| products_admin_update | update | admin | `true` | | T8 |
| products_admin_delete | delete | admin | `true` | | T8 |

### tasks（demo_pm）— OR 表达式 + 权限不对称

字段：`id, title, description, reporter_id, assignee_id, project_id, story_points`

| policyName | action | role | usingExpr | withCheckExpr | 覆盖测试 |
|------------|--------|------|-----------|---------------|---------|
| tasks_read | read | | `row.reporter_id == auth.userid \|\| row.assignee_id == auth.userid` | | T9 |
| tasks_create | create | | | `input.reporter_id == auth.userid` | T9 |
| tasks_update | update | | `row.reporter_id == auth.userid \|\| row.assignee_id == auth.userid` | | T10 |
| tasks_reporter_delete | delete | | `row.reporter_id == auth.userid` | | T10 |
| tasks_admin_read | read | admin | `true` | | T5（demo_pm 侧） |

> T10 验证：以 assignee 身份删除自己被 assign 但不是 reporter 的任务 → 返回 0 affected（usingExpr 不命中）。

### task_comments（demo_pm）— 混合开放/限制

字段：`id, task_id, author_id, content, created_at`

| policyName | action | role | usingExpr | withCheckExpr | 覆盖测试 |
|------------|--------|------|-----------|---------------|---------|
| comments_all_read | read | | `true` | | T11 |
| comments_author_create | create | | | `input.author_id == auth.userid` | T11 |
| comments_author_update | update | | `row.author_id == auth.userid` | | T11 |
| comments_author_delete | delete | | `row.author_id == auth.userid` | | T11 |

### members（demo_pm）— self-read 模式

字段：`id, name, email, avatar_url, joined_at`

| policyName | action | role | usingExpr | withCheckExpr | 覆盖测试 |
|------------|--------|------|-----------|---------------|---------|
| members_self_read | read | | `row.id == auth.userid` | | T12 |
| members_admin_read | read | admin | `true` | | T12（admin 侧） |

### milestones（demo_pm）— 不配置任何策略

> 保持无策略状态，用于验证 T13：所有 EndUser 请求均返回 403。

### addresses（demo_ecommerce）— 补充覆盖（update 无 withCheck）

字段：`id, user_id, name, phone, province, city, district, detail, is_default`

| policyName | action | role | usingExpr | withCheckExpr | 覆盖测试 |
|------------|--------|------|-----------|---------------|---------|
| addr_user_read | read | | `row.user_id == auth.userid` | | 补充 T1 |
| addr_user_create | create | | | `input.user_id == auth.userid` | 补充 T2 |
| addr_user_update | update | | `row.user_id == auth.userid` | | update 无 withCheck（对比 T3）|
| addr_user_delete | delete | | `row.user_id == auth.userid` | | 补充 T4 |

---

## BDD 测试覆盖计划

### 必测（确定性拨测 @deterministic）

```
DET_DB_NAME=demo_ecommerce
DET_MODEL_NAME=orders
```

| Scenario | 操作 | EndUser | 预期 |
|----------|------|---------|------|
| findMany 只返回自己的订单 | GET findMany | user-a | 仅返回 user_id=user-a 的行 |
| count 只统计自己的订单 | GET count | user-a | count <= 全量 |
| 创建 user_id 匹配的订单成功 | POST create(user_id=user-a) | user-a | 返回新建行 |
| 创建 user_id 不匹配被拒 | POST create(user_id=other) | user-a | RLSCheckViolation |
| 更新非自己的订单无效 | PUT update(id=other-row) | user-a | 0 affected，无 error |
| 删除非自己的订单无效 | DEL delete(id=other-row) | user-a | 0 affected，无 error |
| admin 可读所有订单 | GET findMany (role=admin) | user-a | 返回全量数据 |
| viewer 无策略写被拒 | POST create(user_id=wrong) | user-a | RLSCheckViolation |

### 扩展测试（@extended）

```
DET_DB_NAME=demo_pm
DET_MODEL_NAME=tasks
```

| Scenario | 操作 | 预期 |
|----------|------|------|
| reporter 能读自己创建的任务 | findMany as reporter-a | 返回 reporter_id=reporter-a 的行 |
| assignee 也能读被分配的任务 | findMany as assignee-b | 返回 assignee_id=assignee-b 的行 |
| assignee 无法删除非自己 reporter 的任务 | delete as assignee-b | 0 affected |
| 无策略模型返回 403 | GET milestones findMany | 403 |
| 评论全员可读 | GET task_comments findMany | 200 + 数据 |
| 只有作者能删评论 | DELETE comment as non-author | 0 affected |
| self-read：members 只能读到自己 | GET members findMany | 仅返回 id=auth.userid 的行 |

---

## `.env.test` 完整参数

```bash
# 核心拨测模型（orders）
DET_ORG_NAME=luke_e6kz
DET_PROJECT_SLUG=luke
DET_DB_NAME=demo_ecommerce
DET_MODEL_NAME=orders
DET_PAT=mc_pat_dd2e173ea4def1d752e2a2e17fc5b2e37e802e72a911d700c39e27910f61cab3
DET_END_USER_ID=rls-test-user-001
DET_END_USER_NAME=rls-test-user

# 扩展拨测（tasks / task_comments / members）
DET_PM_DB_NAME=demo_pm
DET_TASK_MODEL=tasks
DET_COMMENTS_MODEL=task_comments
DET_MEMBERS_MODEL=members
DET_REPORTER_USER_ID=rls-reporter-001
DET_ASSIGNEE_USER_ID=rls-assignee-002
```

---

## 注意事项

1. **OR 表达式**：`||` 是否受支持待实测，若不支持，tasks 的双字段策略需拆为两条。
2. **`true` vs 空策略**：`usingExpr=true` 表示全员通过；不配置策略则 403，两者语义不同。
3. **withCheckExpr 只在写操作触发**：read 时只看 usingExpr。
4. **0 affected vs error**：update/delete 不命中 usingExpr 时返回空结果而非报错，测试断言要区分。
5. **角色叠加**：同一 action 多条策略 OR 关系，任一命中即通过，admin 策略不会覆盖 user 策略。
