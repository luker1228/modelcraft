# RLS 策略规划

> 基于 luke_e6kz / project luke 的现有模型结构，规划 Open Data API 行级安全策略。
> 表达式语法：`row.<field>` 读取行字段，`input.<field>` 读取写入值，`auth.userid` 为当前 EndUser ID。

---

## 模型字段速查

| 数据库 | 模型 | 用户关联字段 | 说明 |
|--------|------|------------|------|
| demo_ecommerce | users | id, username, email, phone | 用户表本身 |
| demo_ecommerce | orders | user_id | 订单归属用户 |
| demo_ecommerce | addresses | user_id | 收货地址归属用户 |
| demo_ecommerce | products | — | 公开商品目录 |
| demo_ecommerce | categories | — | 公开分类目录 |
| demo_ecommerce | order_items | order_id | 通过订单间接归属 |
| demo_pm | tasks | reporter_id, assignee_id | 创建者/负责人双维度 |
| demo_pm | task_comments | author_id | 评论作者 |
| demo_pm | projects | owner_id | 项目拥有者 |
| demo_pm | project_members | member_id | 成员关联 |
| demo_pm | milestones | project_id | 通过项目间接归属 |
| demo_pm | members | id | 成员自身即标识 |

---

## demo_ecommerce 策略

### orders — 用户只能操作自己的订单

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| orders_read | read | | `row.user_id == auth.userid` | |
| orders_create | create | | | `input.user_id == auth.userid` |
| orders_update | update | | `row.user_id == auth.userid` | `input.user_id == auth.userid` |
| orders_delete | delete | | `row.user_id == auth.userid` | |
| orders_admin_read | read | admin | `true` | |
| orders_admin_write | create | admin | | `true` |

### addresses — 用户只能管理自己的收货地址

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| addr_read | read | | `row.user_id == auth.userid` | |
| addr_create | create | | | `input.user_id == auth.userid` |
| addr_update | update | | `row.user_id == auth.userid` | |
| addr_delete | delete | | `row.user_id == auth.userid` | |
| addr_admin_read | read | admin | `true` | |

### products — 公开只读，admin 可写

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| products_read | read | | `true` | |
| products_create | create | admin | | `true` |
| products_update | update | admin | `true` | |
| products_delete | delete | admin | `true` | |

### categories — 公开只读，admin 可写

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| categories_read | read | | `true` | |
| categories_create | create | admin | | `true` |
| categories_update | update | admin | `true` | |
| categories_delete | delete | admin | `true` | |

### order_items — 通过订单间接归属（暂开放只读）

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| order_items_read | read | | `true` | |
| order_items_admin | create | admin | | `true` |

> 说明：order_items 没有直接 user_id，完整隔离需要 JOIN 查询，当前 RLS 表达式不支持跨表，暂用 `true` 开放只读。

---

## demo_pm 策略

### tasks — reporter 创建，reporter+assignee 可读/更新

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| tasks_read | read | | `row.reporter_id == auth.userid \|\| row.assignee_id == auth.userid` | |
| tasks_create | create | | | `input.reporter_id == auth.userid` |
| tasks_update | update | | `row.reporter_id == auth.userid \|\| row.assignee_id == auth.userid` | |
| tasks_delete | delete | | `row.reporter_id == auth.userid` | |
| tasks_admin_read | read | admin | `true` | |
| tasks_admin_write | create | admin | | `true` |

### task_comments — 评论全员可读，作者可写/删

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| comments_read | read | | `true` | |
| comments_create | create | | | `input.author_id == auth.userid` |
| comments_update | update | | `row.author_id == auth.userid` | |
| comments_delete | delete | | `row.author_id == auth.userid` | |

### projects — owner 完全控制，全员可读列表

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| projects_read | read | | `true` | |
| projects_create | create | | | `input.owner_id == auth.userid` |
| projects_update | update | | `row.owner_id == auth.userid` | |
| projects_delete | delete | | `row.owner_id == auth.userid` | |

### members — 只读自己的成员信息

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| members_self_read | read | | `row.id == auth.userid` | |
| members_admin_read | read | admin | `true` | |

### project_members — 成员只能读自己的关联记录

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| pm_read | read | | `row.member_id == auth.userid` | |
| pm_admin_read | read | admin | `true` | |
| pm_admin_write | create | admin | | `true` |

### milestones — 项目公开，里程碑全员可读

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| milestones_read | read | | `true` | |
| milestones_admin_write | create | admin | | `true` |

---

## BDD 测试推荐模型

优先选 **`orders`（demo_ecommerce）** 作为确定性拨测模型：

- 字段 `user_id` 语义等价 `owner`，RLS 模式最标准
- 支持 read / create / update / delete 四种操作测试
- 没有多字段联合判断，测试用例简单清晰

`.env.test` 修改建议：
```
DET_DB_NAME=demo_ecommerce
DET_MODEL_NAME=orders
```

对应 Open Data API 字段调整：
- `findMany` 查询字段：`id order_no user_id total_amount`
- `create` 写入字段：`user_id`（触发 withCheckExpr 验证）
- RLS 表达式：`row.user_id == auth.userid` / `input.user_id == auth.userid`

---

## 注意事项

1. **`||` 表达式**：tasks 的双字段策略（`reporter_id || assignee_id`）需确认 RLS 引擎是否支持 `||` 运算符，若不支持需拆成两条策略。
2. **order_items 跨表**：当前 RLS 不支持 JOIN，只能开放只读或在业务层控制写入。
3. **admin 角色**：`role="admin"` 的策略只对携带 `X-MC-Auth-Roles: admin` 的请求生效，与 `role=""` 策略叠加（任一命中即通过）。
4. **无策略 = 403**：模型未配置任何策略时，所有 EndUser 请求返回 403 Permission denied。
