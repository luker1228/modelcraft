# RLS 策略配置手册

> 本文档记录测试环境中各模型需要配置的 RLS 策略。
> 执行测试前请确认所有策略已按此文档配置完毕。
>
> 环境：luke_e6kz / project luke
> 表达式语法：`row.<field>` 读取行字段，`input.<field>` 读取写入值，`auth.userid` 为当前 EndUser ID

---

## 快速索引

| 数据库 | 模型 | 策略数 | 用途 |
|--------|------|--------|------|
| demo_ecommerce | orders | 6 | 核心 CRUD 测试基准 |
| demo_ecommerce | products | 4 | 公开读 + 角色写 |
| demo_pm | tasks | 5 | OR 表达式 + 权限不对称 |
| demo_pm | task_comments | 4 | 混合开放/限制 |
| demo_pm | members | 2 | self-read 模式 |
| demo_pm | milestones | 0 | **不配置**，用于测试 403 |

---

## demo_ecommerce

### orders

字段：`id, order_no, user_id, total_amount, paid_amount, remark, created_at`

测试覆盖：用户只能操作自己的订单，admin 角色可读全量，viewer 写入受限。

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| orders_user_read | read | * | `row.user_id == auth.userid` | |
| orders_user_create | create | * | | `input.user_id == auth.userid` |
| orders_user_update | update | * | `row.user_id == auth.userid` | `input.user_id == auth.userid` |
| orders_user_delete | delete | * | `row.user_id == auth.userid` | |
| orders_admin_read | read | admin | `true` | |
| orders_admin_create | create | admin | | `true` |

### products

字段：`id, name, description, price, stock, category_id`

测试覆盖：`usingExpr=true` 全员可读，admin 限写（T7、T8）。

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| products_all_read | read | * | `true` | |
| products_admin_create | create | admin | | `true` |
| products_admin_update | update | admin | `true` | |
| products_admin_delete | delete | admin | `true` | |

---

## demo_pm

### tasks

字段：`id, title, description, reporter_id, assignee_id, project_id, milestone_id, story_points, due_date, created_at, updated_at, completed_at`

测试覆盖：
- reporter 或 assignee 可读/更新（拆为两条独立策略代替 `||`）
- `in` 操作：只有指定 milestone 下的任务可被普通用户访问（`row.milestone_id in [...]`）
- reporter 独占删除权，assignee 无删除权
- create 要求 `reporter_id == auth.userid`
- `withCheckExpr` + `in`：create 时 `project_id` 必须在合法范围内（`input.project_id in ["allowed-proj-1", "allowed-proj-2"]`）
- admin 角色读取全量

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| tasks_reporter_read | read | * | `row.reporter_id == auth.userid` | |
| tasks_assignee_read | read | * | `row.assignee_id == auth.userid` | |
| tasks_create | create | * | | `input.reporter_id == auth.userid` |
| tasks_reporter_update | update | * | `row.reporter_id == auth.userid` | |
| tasks_assignee_update | update | * | `row.assignee_id == auth.userid` | |
| tasks_reporter_delete | delete | * | `row.reporter_id == auth.userid` | |
| tasks_admin_read | read | admin | `true` | |

**`in` 专项策略**（单独场景通过 upsert 覆盖 `tasks_create`，验证完下一个 Scenario 由 Background 重置）：

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| tasks_create | create | * | | `input.project_id in ["proj-allowed-1", "proj-allowed-2"]` |
| tasks_milestone_read | read | * | `row.milestone_id in ["ms-allowed-1", "ms-allowed-2"]` | |

### task_comments

字段：`id, task_id, author_id, content, created_at`

测试覆盖：
- 全员可读
- `in` 操作：只有特定 task 下的评论可被普通用户访问（`row.task_id in [...]`）
- 只有 author 可写/改/删
- `withCheckExpr` + `in`：create 时限定 task 范围（`input.task_id in ["task-a", "task-b"]`）

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| comments_all_read | read | * | `true` | |
| comments_author_create | create | * | | `input.author_id == auth.userid` |
| comments_author_update | update | * | `row.author_id == auth.userid` | |
| comments_author_delete | delete | * | `row.author_id == auth.userid` | |

**`in` 专项策略**（单独场景通过 upsert 覆盖 `comments_author_create`）：

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| comments_author_create | create | * | | `input.task_id in ["task-visible-1", "task-visible-2"]` |
| comments_task_read | read | * | `row.task_id in ["task-visible-1", "task-visible-2"]` | |

### members

字段：`id, name, email, avatar_url, joined_at`

测试覆盖：只能读自己的记录（T12），admin 可读全量。

| policyName | action | role | usingExpr | withCheckExpr |
|------------|--------|------|-----------|---------------|
| members_self_read | read | * | `row.id == auth.userid` | |
| members_admin_read | read | admin | `true` | |

### milestones

**不配置任何策略。**

用于测试 T13：无策略时所有 EndUser 请求返回 `PermissionDenied`（no RLS policy configured for action: read）。

---

## BDD 测试环境变量

`.env.test` 中对应 `@deterministic` 拨测的配置：

```bash
DET_ORG_NAME=luke_e6kz
DET_PROJECT_SLUG=luke
DET_DB_NAME=demo_ecommerce
DET_MODEL_NAME=orders
DET_PAT=mc_pat_dd2e173ea4def1d752e2a2e17fc5b2e37e802e72a911d700c39e27910f61cab3
DET_END_USER_ID=rls-test-user-001
DET_END_USER_NAME=rls-test-user
```

扩展测试（`@deterministic`，demo_pm 数据库）使用：

```bash
DET_PM_DB_NAME=demo_pm
DET_TASK_MODEL=tasks
DET_COMMENTS_MODEL=task_comments
DET_MEMBERS_MODEL=members
DET_REPORTER_USER_ID=rls-reporter-001
DET_ASSIGNEE_USER_ID=rls-assignee-002
# in 测试中允许访问的 milestone / project / task ID
DET_ALLOWED_MILESTONE_IDS=ms-allowed-1,ms-allowed-2
DET_ALLOWED_PROJECT_IDS=proj-allowed-1,proj-allowed-2
DET_ALLOWED_TASK_IDS=task-visible-1,task-visible-2
```
