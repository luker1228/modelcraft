# RLS BDD 拨测 SQL 验证流程

> 适用场景：运行 `deterministic-runtime-rls.feature` 后，通过 `requestId` 定位后端日志，确认每条拨测实际执行的 SQL 是否符合预期。

---

## 概述

每条 Open Data API 调用会：
1. 生成唯一 `clientRequestId`（格式：`bdd-<timestamp36>-<random5>`）
2. 通过 `X-Client-Request-Id` header 发送给后端
3. 后端在响应 `extensions.requestId` 中返回服务端 `requestId`
4. 测试步骤通过 `setLastOpenDataResult` 将两个 ID 打印到控制台

---

## 第一步：运行拨测并收集 requestId

```bash
cd tests-bdd
npm run test:rls-det 2>&1 | grep "\[rls\]\|requestId="
```

输出示例：

```
[rls] findMany 查询受 RLS 约束 — EndUser 只能看到自己的数据
      requestId=9a2d9397-75aa-4dd3-b777-ca35ad704830
      └─ just log-cat 9a2d9397-75aa-4dd3-b777-ca35ad704830
[rls] count 查询受 RLS 约束
      requestId=178e0d29-2989-4344-9aae-96ad8716d40b
      └─ just log-cat 178e0d29-2989-4344-9aae-96ad8716d40b
```

---

## 第二步：按 requestId 查日志

后端跑在 Docker 容器中，日志在 `/app/logs/server.log`。

**方式一：log-cat（本地开发模式）**

```bash
cd modelcraft-backend
just log-cat <requestId>
```

**方式二：docker exec（Docker 部署模式）**

```bash
docker exec modelcraft-backend grep "<requestId>" /app/logs/server.log | python3 -m json.tool
```

**只看 SQL 行：**

```bash
RID="<requestId>"
docker exec modelcraft-backend grep "$RID" /app/logs/server.log \
  | python3 -c "
import sys, json
for line in sys.stdin:
    try:
        obj = json.loads(line.strip())
        msg = obj.get('msg','')
        if 'sql=' in msg and 'args=' in msg:
            print(msg)
    except: pass
"
```

---

## 第三步：对照预期 SQL

### 前 10 条场景验证记录

> 「被拦截」场景额外列出实际 GraphQL 错误响应，用于审计错误设计是否合理。

| # | 场景 | 预期行为 | 实际 SQL / 错误响应 | 符合预期 |
|---|------|---------|-------------------|---------|
| 1 | findMany 查询受 RLS 约束 — EndUser 只能看到自己的数据 | WHERE 含 RLS `user_id=?` 过滤 | `SELECT * FROM orders WHERE (user_id = ?) ORDER BY id ASC LIMIT ? args=[rls-test-user-001 20]` | ✅ |
| 2 | count 查询受 RLS 约束 | COUNT WHERE 含 RLS 过滤 | `SELECT COUNT(*) FROM orders WHERE (user_id = ?) args=[rls-test-user-001]` | ✅ |
| 3 | 创建记录时 user_id 自动等于当前 EndUser | INSERT user_id=当前用户 | `INSERT INTO orders (..., user_id) VALUES (..., rls-test-user-001)` | ✅ |
| 4 | 无法创建 user_id 为其他人的记录 | CHECK 拦截，无 SQL 执行 | 见下方错误详情 ↓ | ✅ |
| 5 | 无法更新非本人的记录（admin 创建目标记录） | INSERT user_id=other-owner-id | `INSERT INTO orders (..., user_id) VALUES (..., other-owner-id)` | ✅ |
| 6 | 无法更新非本人的记录（EndUser 尝试更新） | withCheckExpr `input.user_id == auth.userid` 失败（update 只传 remark，input.user_id 为空） | 见下方错误详情 ↓ | ✅ |
| 7 | 无法删除非本人的记录（admin 创建目标记录） | INSERT user_id=other-owner-id | `INSERT INTO orders (..., user_id) VALUES (..., other-owner-id)` | ✅ |
| 8 | 无法删除非本人的记录（EndUser 尝试删除） | USING filter 过滤，pre-SELECT 0 行，返回 REPO_NOT_FOUND | `SELECT * FROM orders WHERE ((id=?) AND ((user_id=?))) args=[..., rls-test-user-001]` → 0 行；返回 `[NOT_FOUND.RECORD] Record Resource not found` | ✅ |
| 9 | admin 角色可以读取所有数据 | WHERE 含 `(user_id=?) OR (1=1)` | `SELECT * FROM orders WHERE (user_id = ?) OR (1=1) ORDER BY id ASC LIMIT ? args=[rls-test-user-001 20]` | ✅ |
| 10 | viewer 角色可以读取且可以创建自己的记录 | WHERE 含 RLS `user_id=?` | `SELECT * FROM orders WHERE (user_id = ?) ORDER BY id ASC LIMIT ? args=[rls-test-user-001 20]` | ✅ |

---

### orders user_id 字段专项（#11–15）

| # | 场景 | 预期行为 | 实际 SQL / 错误响应 | 符合预期 |
|---|------|---------|-------------------|---------|
| 11 | user_id 字段 — EndUser 只能读取自己的 orders 记录 | WHERE 含 RLS `user_id=?` | `SELECT * FROM orders WHERE ((user_id = ?)) ORDER BY id ASC LIMIT ? args=[rls-test-user-001 20]` | ✅ |
| 12 | user_id 字段 — 创建 orders 记录时 user_id 必须等于当前用户 | INSERT user_id=当前用户 | `INSERT INTO orders (..., user_id) VALUES (..., rls-test-user-001)` | ✅ |
| 13 | user_id 字段 — 创建 orders 记录时 user_id 为他人则拒绝 | CHECK 拦截，无 SQL 执行 | 见下方错误详情 ↓ | ✅ |
| 14 | user_id 字段 — admin 角色可读取所有 orders 记录 | WHERE 含 `(user_id=?) OR (1=1)` | `SELECT * FROM orders WHERE ((user_id = ?) OR (1=1)) ORDER BY id ASC LIMIT ? args=[rls-test-user-001 20]` | ✅ |
| 15 | user_id 字段 — admin 角色可创建任意 user_id 的 orders 记录 | INSERT user_id=any-user-id（admin CHECK true 通过） | `INSERT INTO orders (..., user_id) VALUES (..., any-user-id)` | ✅ |

---

### products 策略专项（#16–22）

products 模型策略：全员可读（`usingExpr: true`）、admin 限写。

| # | 场景 | 预期行为 | 实际 SQL / 错误响应 | 符合预期 |
|---|------|---------|-------------------|---------|
| 16 | 全员可读 — 任意 EndUser 可以读取 products（findMany） | WHERE `(1=1)`，无用户过滤 | `SELECT * FROM products WHERE ((1=1)) ORDER BY id ASC LIMIT ? args=[20]` | ✅ |
| 17 | 全员可读 — count 查询无限制 | COUNT WHERE `(1=1)` | `SELECT COUNT(*) FROM products WHERE ((1=1)) args=[]` | ✅ |
| 18 | 普通用户无法创建 products 记录（非 admin） | 无 create policy，直接拒绝 | 见下方错误详情 ↓ | ✅ |
| 19 | admin 角色可读取所有 products 记录 | WHERE `(1=1)` | `SELECT * FROM products WHERE ((1=1)) ORDER BY id ASC LIMIT ? args=[20]` | ✅ |
| 20 | admin 角色可创建 products 记录 | INSERT 成功 | `INSERT INTO products (category_id, id, name, price, stock) VALUES (...)` | ✅ |
| 21 | admin 角色可更新 products 记录（非存在记录则无副作用） | UPDATE WHERE `(id=?) AND (1=1)` 执行，rowsAffected=0，返回 REPO_NOT_FOUND | `UPDATE products SET name=? WHERE ((id=?) AND ((1=1)))` → rowsAffected=0；返回 `[NOT_FOUND.RECORD] Record Resource not found` | ✅ |
| 22 | viewer 角色可读取 products 但无法创建 | findMany 成功（WHERE 1=1）；create 无 policy 拒绝 | findMany: `SELECT * FROM products WHERE ((1=1))...`；create: 见下方错误详情 ↓ | ✅ |

---

### tasks 双主体策略专项（#23–32）

tasks 模型：reporter/assignee 双 USING 过滤，reporter 可 create/delete，assignee 可 update。

| # | 场景 | 预期行为 | 实际 SQL / 错误响应 | 符合预期 |
|---|------|---------|-------------------|---------|
| 23 | reporter 执行 findMany | WHERE `(reporter_id=?) OR (assignee_id=?)` 且 userId=reporter | `SELECT * FROM tasks WHERE ((reporter_id = ?) OR (assignee_id = ?)) ORDER BY id ASC LIMIT ? args=[rls-reporter-001 rls-reporter-001 20]` | ✅ |
| 24 | assignee 执行 findMany | WHERE `(reporter_id=?) OR (assignee_id=?)` 且 userId=assignee | `SELECT * FROM tasks WHERE ((reporter_id = ?) OR (assignee_id = ?)) ORDER BY id ASC LIMIT ? args=[rls-assignee-002 rls-assignee-002 20]` | ✅ |
| 25 | reporter 创建 task（reporter_id=自己） | INSERT reporter_id=reporter，CHECK 通过 | `INSERT INTO tasks (id, project_id, reporter_id, title) VALUES (..., rls-reporter-001, ...)` | ✅ |
| 26 | reporter 创建 task（reporter_id=他人，withCheckExpr 拦截） | CHECK 失败，无 SQL | 见下方错误详情 ↓ | ✅ |
| 27 | reporter 创建 rls-update-target 记录 | INSERT 成功（为后续 update 测试准备） | `INSERT INTO tasks (id, project_id, reporter_id, title) VALUES (...)` | ✅ |
| 28 | assignee 尝试更新 reporter 创建的 task | USING filter 仅匹配 assignee_id，UPDATE rowsAffected=0，返回 REPO_NOT_FOUND | `UPDATE tasks SET title=? WHERE ((id=?) AND ((reporter_id=?) OR (assignee_id=?))) args=[..., rls-assignee-002, rls-assignee-002]` → rowsAffected=0；返回 `[NOT_FOUND.RECORD] Record Resource not found` | ✅ |
| 29 | reporter 创建 rls-delete-target 记录 | INSERT 成功 | `INSERT INTO tasks (id, project_id, reporter_id, title) VALUES (...)` | ✅ |
| 30 | assignee 尝试删除 reporter 创建的 task | pre-SELECT 按 reporter_id=assignee 查，0 行，返回 REPO_NOT_FOUND | `SELECT * FROM tasks WHERE ((id=?) AND ((reporter_id=?))) args=[..., rls-assignee-002]` → 0 行；返回 `[NOT_FOUND.RECORD] Record Resource not found` | ✅ |
| 31 | reporter 删除自己创建的 task | DELETE WHERE `(id=?) AND (reporter_id=?)` 匹配 1 行 | `SELECT * FROM tasks WHERE ((id=?) AND ((reporter_id=?))) args=[..., rls-reporter-001]`；`DELETE FROM tasks WHERE ((id=?) AND ((reporter_id=?))) LIMIT 1` | ✅ |
| 32 | admin 角色对 tasks 执行 findMany | WHERE `(reporter_id=?) OR (assignee_id=?) OR (1=1)` | `SELECT * FROM tasks WHERE ((reporter_id = ?) OR (assignee_id = ?) OR (1=1)) ORDER BY id ASC LIMIT ? args=[rls-test-user-001 rls-test-user-001 20]` | ✅ |

---

### tasks in 操作专项（#33–34）

| # | 场景 | 预期行为 | 实际 SQL / 错误响应 | 符合预期 |
|---|------|---------|-------------------|---------|
| 33 | tasks milestone_id in usingExpr — 只有指定 milestone 下的任务可读 | WHERE `milestone_id IN (?, ?)` | `SELECT * FROM tasks WHERE ((milestone_id IN (?, ?))) ORDER BY id ASC LIMIT ? args=[ms-allowed-1 ms-allowed-2 20]` | ✅ |
| 34 | tasks project_id in withCheckExpr — project_id 不在列表则拒绝 | CHECK in 表达式求值失败，无 SQL | 见下方错误详情 ↓ | ✅ |

---

### task_comments 策略专项（#35–39）

task_comments 模型：全员可读（`usingExpr: true`），author 限写/改/删。

| # | 场景 | 预期行为 | 实际 SQL / 错误响应 | 符合预期 |
|---|------|---------|-------------------|---------|
| 35 | 全员可读 — findMany | WHERE `(1=1)` | `SELECT * FROM task_comments WHERE ((1=1)) ORDER BY id ASC LIMIT ? args=[20]` | ✅ |
| 36 | author 创建自己的评论 | INSERT author_id=当前用户，CHECK 通过 | `INSERT INTO task_comments (author_id, content, id, task_id) VALUES (rls-test-user-001, ...)` | ✅ |
| 37 | 创建 author_id 为他人的评论（withCheckExpr 拦截） | CHECK 失败，无 SQL | 见下方错误详情 ↓ | ✅ |
| 38 | 更新 non-existent-id 评论（USING 过滤，0 行受影响） | UPDATE rowsAffected=0，返回 REPO_NOT_FOUND | `UPDATE task_comments SET content=? WHERE ((id=?) AND ((author_id=?)))` → rowsAffected=0；返回 `[NOT_FOUND.RECORD] Record Resource not found` | ✅ |
| 39 | 删除 non-existent-id 评论（USING 过滤，0 行受影响） | pre-SELECT 按 author_id 查，0 行，返回 REPO_NOT_FOUND | `SELECT * FROM task_comments WHERE ((id=?) AND ((author_id=?))) args=[non-existent-id, rls-test-user-001]` → 0 行；返回 `[NOT_FOUND.RECORD] Record Resource not found` | ✅ |

---

### task_comments in 操作专项（#40–41）

| # | 场景 | 预期行为 | 实际 SQL / 错误响应 | 符合预期 |
|---|------|---------|-------------------|---------|
| 40 | task_comments task_id in usingExpr — 只有指定 task 下的评论可读 | WHERE `task_id IN (?, ?)` | `SELECT * FROM task_comments WHERE ((task_id IN (?, ?))) ORDER BY id ASC LIMIT ? args=[task-visible-1 task-visible-2 20]` | ✅ |
| 41 | task_comments task_id in withCheckExpr — task_id 不在列表则拒绝 | CHECK in 表达式求值失败，无 SQL | 见下方错误详情 ↓ | ✅ |

---

### useAdmin 查询条件专项（#42–58）

useAdmin（`X-MC-Auth-Useadmin: true`）绕过 RLS，直接生成裸 SQL。

| # | 场景 | 预期 SQL 模式 | 实际 SQL | 符合预期 |
|---|------|-------------|---------|---------|
| 42 | useAdmin — findMany 无条件 | `SELECT * FROM orders ORDER BY id ASC LIMIT 20` | `SELECT * FROM orders ORDER BY id ASC LIMIT ? args=[20]` | ✅ |
| 43 | useAdmin — findMany take=5 skip=0 | `LIMIT 5` | `SELECT * FROM orders ORDER BY id ASC LIMIT ? args=[5]` | ✅ |
| 44 | useAdmin — findMany take=5 skip=5 | `LIMIT 5 OFFSET 5` | `SELECT * FROM orders ORDER BY id ASC LIMIT ? OFFSET ? args=[5 5]` | ✅ |
| 45 | useAdmin — findMany orderBy id desc | `ORDER BY id DESC` | `SELECT * FROM orders ORDER BY id DESC LIMIT ? args=[20]` | ✅ |
| 46 | useAdmin — findMany orderBy total_amount asc | `ORDER BY total_amount ASC` | `SELECT * FROM orders ORDER BY total_amount ASC LIMIT ? args=[20]` | ✅ |
| 47 | useAdmin — findMany where user_id eq | `WHERE user_id = ?` | `SELECT * FROM orders WHERE (user_id = ?) ORDER BY id ASC LIMIT ? args=[det-test-user-001 20]` | ✅ |
| 48 | useAdmin — findMany where user_id in 列表 | `WHERE user_id IN (?, ?)` | `SELECT * FROM orders WHERE (user_id IN (?, ?)) ORDER BY id ASC LIMIT ? args=[det-test-user-001 other-user-id 20]` | ✅ |
| 49 | useAdmin — findMany take=10 skip=0 orderBy id desc | `LIMIT 10 ORDER BY id DESC` | `SELECT * FROM orders ORDER BY id DESC LIMIT ? args=[10]` | ✅ |
| 50 | useAdmin — count 无条件 | `SELECT COUNT(*) FROM orders` | `SELECT COUNT(*) AS count FROM orders args=[]` | ✅ |
| 51 | useAdmin — count where user_id eq | `COUNT WHERE user_id = ?` | `SELECT COUNT(*) AS count FROM orders WHERE (user_id = ?) args=[det-test-user-001]` | ✅ |
| 52 | useAdmin — findMany where total_amount gt 0 | `WHERE total_amount > 0` | `SELECT * FROM orders WHERE (total_amount > ?) ORDER BY id ASC LIMIT ? args=[0 20]` | ✅ |
| 53 | useAdmin — findMany where total_amount lte 100 | `WHERE total_amount <= 100` | `SELECT * FROM orders WHERE (total_amount <= ?) ORDER BY id ASC LIMIT ? args=[100 20]` | ✅ |
| 54 | useAdmin — findMany where order_no startsWith "bdd-" | `WHERE order_no LIKE BINARY 'bdd-%'` | `SELECT * FROM orders WHERE (order_no LIKE BINARY ?) ORDER BY id ASC LIMIT ? args=[bdd-% 20]` | ✅ |
| 55 | useAdmin — findMany where address_id ne 指定值 | `WHERE address_id != ?` | `SELECT * FROM orders WHERE (address_id != ?) ORDER BY id ASC LIMIT ? args=[non-existent-addr 20]` | ✅ |
| 56 | useAdmin — findMany where remark not 指定值 | `WHERE remark != ?` | `SELECT * FROM orders WHERE (remark != ?) ORDER BY id ASC LIMIT ? args=[non-existent-remark 20]` | ✅ |
| 57 | useAdmin — findMany 多字段排序 total_amount asc id desc | `ORDER BY total_amount ASC, id DESC` | `SELECT * FROM orders ORDER BY total_amount ASC, id DESC LIMIT ? args=[20]` | ✅ |
| 58 | useAdmin — findMany take=1 单条限制 | `LIMIT 1` | `SELECT * FROM orders ORDER BY id ASC LIMIT ? args=[1]` | ✅ |
| 59 | useAdmin — findMany where paid_amount gte 0 lte 1000 | `WHERE paid_amount >= 0 AND paid_amount <= 1000` | `SELECT * FROM orders WHERE ((paid_amount >= ?) AND (paid_amount <= ?)) ORDER BY id ASC LIMIT ? args=[0 1000 20]` | ✅ |

---

### 被拦截场景错误响应审计

每个「被拒绝」场景的实际 GraphQL 错误响应，以及后端 RLS 快照，用于确认错误消息设计是否合理。

---

#### #4 无法创建 user_id 为其他人的记录

**场景**：EndUser `rls-test-user-001` 尝试 INSERT `user_id=other-user-id`，`withCheckExpr: input.user_id == auth.userid` 应拦截。

**RLS 快照**：
```
create=true（有 create policy，进入 CHECK 求值）
```

**GraphQL 响应**：
```json
{"data":{"create":null},"errors":[{"message":"[OPERATION_FAILED.PERMISSION] Permission denied: all CHECK expressions failed","locations":[{"line":1,"column":12}],"path":["create"]}]}
```

**审计结论**：✅ 错误码 `OPERATION_FAILED.PERMISSION` 语义清晰，`all CHECK expressions failed` 明确说明是 CHECK 阶段拦截。无泄漏敏感信息（未暴露具体表达式内容）。

---

#### #6 无法更新非本人的记录（EndUser 尝试更新）

**场景**：EndUser `rls-test-user-001` 尝试 UPDATE `user_id=other-owner-id` 的记录，`withCheckExpr: input.user_id == auth.userid`。update mutation 只传 `remark` 字段，`input.user_id` 为空，CHECK 求值失败。

**RLS 快照**：
```
update=true（有 update policy，withCheckExpr 生效；input.user_id 为空不等于 auth.userid）
```

**GraphQL 响应**：
```json
{"data":{"update":null},"errors":[{"message":"[OPERATION_FAILED.PERMISSION] Permission denied: all CHECK expressions failed","locations":[{"line":1,"column":12}],"path":["update"]}]}
```

**审计结论**：✅ withCheckExpr `input.user_id == auth.userid` 正确拦截——update 未传 user_id 字段时 input.user_id 为空，CHECK 求值失败。注意：此拦截发生在 CHECK 阶段，而非 USING filter 阶段；USING filter 未实际执行 SQL。

---

#### #8 无法删除非本人的记录（EndUser 尝试删除）

**场景**：EndUser `rls-test-user-001` 尝试 DELETE `user_id=other-owner-id` 的记录，`USING filter: row.user_id == auth.userid` 使该记录对当前用户不可见。

**RLS 快照**：
```
delete=true（有 delete policy，先 SELECT 确认记录可见性）
```

**实际 SQL**：
```sql
SELECT * FROM orders WHERE ((id = ?) AND ((user_id = ?))) args=[bdd-..., rls-test-user-001]
-- → 0 行（other-owner-id 的记录对当前用户不可见）
-- → 不执行 DELETE，直接返回 REPO_NOT_FOUND
```

**GraphQL 响应**：
```json
{"data":{"delete":null},"errors":[{"message":"[NOT_FOUND.RECORD] Record Resource not found","locations":[{"line":1,"column":12}],"path":["delete"]}]}
```

**审计结论**：✅ USING filter 正确过滤——pre-SELECT 未找到行，DELETE 不执行。错误码 `NOT_FOUND.RECORD` 对调用方表现为"记录不存在"，不泄露「被 RLS 过滤」的信息，符合安全设计。

---

#### #21 admin 更新 products 不存在记录

**GraphQL 响应**：
```json
{"data":{"update":null},"errors":[{"message":"[NOT_FOUND.RECORD] Record Resource not found","locations":[{"line":1,"column":12}],"path":["update"]}]}
```

**审计结论**：✅ UPDATE 执行后 rowsAffected=0，返回 `NOT_FOUND.RECORD`。RLS 行为正确（admin `1=1` policy 通过，SQL 执行，但记录不存在）。错误消息语义清晰，与 delete 场景一致。

---

#### #28 assignee 更新 reporter 的 task（USING 过滤，0 行）

**GraphQL 响应**：
```json
{"data":{"update":null},"errors":[{"message":"[NOT_FOUND.RECORD] Record Resource not found","locations":[{"line":1,"column":12}],"path":["update"]}]}
```

**审计结论**：✅ UPDATE 执行后 rowsAffected=0，返回 `NOT_FOUND.RECORD`。RLS 行为正确（USING filter `reporter_id OR assignee_id` 中 assignee 不匹配该 task 的 reporter，0 行受影响）。

---

#### #30 assignee 删除 reporter 的 task（USING 过滤，0 行）

**GraphQL 响应**：
```json
{"data":{"delete":null},"errors":[{"message":"[NOT_FOUND.RECORD] Record Resource not found","locations":[{"line":1,"column":12}],"path":["delete"]}]}
```

**审计结论**：✅ 同 #8，pre-SELECT 用 reporter_id=assignee 查不到行，DELETE 不执行，返回 `NOT_FOUND.RECORD`。对调用方表现为"记录不存在"，不泄露 RLS 过滤信息。

---

#### #38 更新 non-existent-id 评论（USING 过滤）

**GraphQL 响应**：
```json
{"data":{"update":null},"errors":[{"message":"[NOT_FOUND.RECORD] Record Resource not found","locations":[{"line":1,"column":12}],"path":["update"]}]}
```

**审计结论**：✅ 同 #21/#28，UPDATE 执行后 rowsAffected=0，返回 `NOT_FOUND.RECORD`。RLS 行为正确（USING filter `author_id=current-user` 过滤，non-existent-id 无论如何都不存在）。

---

#### #39 删除 non-existent-id 评论（USING 过滤）

**GraphQL 响应**：
```json
{"data":{"delete":null},"errors":[{"message":"[NOT_FOUND.RECORD] Record Resource not found","locations":[{"line":1,"column":12}],"path":["delete"]}]}
```

**审计结论**：✅ 同 #8/#30，pre-SELECT 用 author_id=current-user 查不到行，DELETE 不执行，返回 `NOT_FOUND.RECORD`。

---

#### #13 无法创建 user_id 为其他人的记录（user_id 字段专项）

**场景**：EndUser `rls-test-user-001` 尝试创建 `user_id=other-user-id` 的 orders 记录，`withCheckExpr: input.user_id == auth.userid` 应拦截。

**RLS 快照**：
```
create=true（有 create policy，进入 CHECK 求值）
```

**GraphQL 响应**：
```json
{"data":{"create":null},"errors":[{"message":"[OPERATION_FAILED.PERMISSION] Permission denied: all CHECK expressions failed","locations":[{"line":1,"column":12}],"path":["create"]}]}
```

**审计结论**：✅ 与 #4 相同的 CHECK 拦截路径，错误消息一致。

---

#### #18 普通用户无法创建 products 记录（非 admin）

**场景**：EndUser 对 products 模型发起 INSERT，products 模型只有 `products_admin_create`（role=admin），viewer 角色无 create policy。

**RLS 快照**：
```
create=false（无 viewer create policy，直接拒绝）
```

**GraphQL 响应**：
```json
{"data":{"create":null},"errors":[{"message":"[OPERATION_FAILED.PERMISSION] Permission denied: insert","locations":[{"line":1,"column":12}],"path":["create"]}]}
```

**审计结论**：✅ `Permission denied: insert` 与「无 create policy」的语义对应。与 CHECK 失败的错误消息（`all CHECK expressions failed`）不同，两种拦截路径错误消息有区分，设计合理。

---

#### #22 viewer 角色可读取 products 但无法创建

**GraphQL 响应**（无 create policy 拦截）：
```json
{"data":{"create":null},"errors":[{"message":"[OPERATION_FAILED.PERMISSION] Permission denied: insert","locations":[{"line":1,"column":12}],"path":["create"]}]}
```

**审计结论**：✅ 同 #18，`Permission denied: insert` 准确表达无 insert policy。

---

#### #26 tasks reporter_id 为他人（withCheckExpr 拦截）

**场景**：reporter 用户创建 task，`reporter_id=other-user-id`，`withCheckExpr: input.reporter_id == auth.userid` 拦截。

**RLS 快照**：
```
create=true（有 create policy，进入 CHECK 求值）
```

**GraphQL 响应**：
```json
{"data":{"create":null},"errors":[{"message":"[OPERATION_FAILED.PERMISSION] Permission denied: all CHECK expressions failed","locations":[{"line":1,"column":12}],"path":["create"]}]}
```

**审计结论**：✅ 与 #4 相同的 CHECK 拦截路径，错误消息一致。

---

#### #34 tasks project_id in withCheckExpr — project_id 不在列表则拒绝

**场景**：reporter 用户创建 task，`project_id=proj-forbidden` 不在 `["proj-allowed-1","proj-allowed-2"]` 内，`withCheckExpr: input.project_id in [...]` 拦截。

**RLS 快照**：
```
create=true（有 create policy，进入 CHECK 求值）
```

**GraphQL 响应**：
```json
{"data":{"create":null},"errors":[{"message":"[OPERATION_FAILED.PERMISSION] Permission denied: all CHECK expressions failed","locations":[{"line":1,"column":12}],"path":["create"]}]}
```

**审计结论**：✅ `in` 操作符的 withCheckExpr 拦截路径与普通字段比较拦截路径错误消息相同，一致性好。

---

#### #37 task_comments author_id 为他人（withCheckExpr 拦截）

**场景**：EndUser 创建评论，`author_id=other-user-id`，`withCheckExpr: input.author_id == auth.userid` 拦截。

**GraphQL 响应**：
```json
{"data":{"create":null},"errors":[{"message":"[OPERATION_FAILED.PERMISSION] Permission denied: all CHECK expressions failed","locations":[{"line":1,"column":12}],"path":["create"]}]}
```

**审计结论**：✅ 与 #4 相同的 CHECK 拦截路径，错误消息一致。

---

#### #41 task_comments task_id in withCheckExpr — task_id 不在列表则拒绝

**GraphQL 响应**：
```json
{"data":{"create":null},"errors":[{"message":"[OPERATION_FAILED.PERMISSION] Permission denied: all CHECK expressions failed","locations":[{"line":1,"column":12}],"path":["create"]}]}
```

**审计结论**：✅ 同 #34。

---

### 错误消息汇总

| 拦截类型 | 错误消息 | 备注 |
|---------|---------|------|
| withCheckExpr 求值失败（create/update） | `[OPERATION_FAILED.PERMISSION] Permission denied: all CHECK expressions failed` | 有 policy 但 CHECK 不通过 |
| 无 create policy（insert 被禁） | `[OPERATION_FAILED.PERMISSION] Permission denied: insert` | 无对应 action 的 policy |
| pre-SELECT 0 行（delete） | `[NOT_FOUND.RECORD] Record Resource not found` | 先 SELECT 后 DELETE，SELECT 返回 0 行则不执行 DELETE |
| UPDATE rowsAffected=0 | `[NOT_FOUND.RECORD] Record Resource not found` | UPDATE 执行后 0 行受影响，返回记录不存在 |

---

## RLS SQL 模式速查

| 场景类型 | 预期 SQL 模式 |
|---------|-------------|
| EndUser findMany（role=viewer，单字段 USING） | `WHERE (user_id = ?) ... args=[<userId>]` |
| EndUser findMany（role=viewer，双字段 USING） | `WHERE ((reporter_id = ?) OR (assignee_id = ?))` |
| EndUser findMany（`usingExpr: true`，全员可读） | `WHERE ((1=1))` |
| EndUser count | `WHERE (user_id = ?) args=[<userId>]` |
| admin findMany（role=admin，有 `usingExpr: true` policy） | `WHERE (user_id = ?) OR (1=1)` — admin policy `true` 产生 `OR (1=1)` |
| useAdmin findMany（X-MC-Auth-Useadmin: true） | 无 WHERE RLS 过滤，裸 SQL；分页/排序/where 参数直通 |
| `in` usingExpr（findMany） | `WHERE (field IN (?, ?)) ... args=[val1 val2]` |
| INSERT 被 CHECK 拦截（`withCheckExpr` 等值比较或 `in`） | 无 SQL 执行，后端在内存中 CEL 求值拒绝 |
| INSERT 无对应 action 的 policy | 无 SQL 执行，直接拒绝 |
| UPDATE/DELETE 被 USING 过滤 | SQL 执行但 affected rows=0，或 SELECT 0 行后不执行 |
| DELETE（reporter 有 delete policy） | SELECT 确认行可见，然后 DELETE LIMIT 1 |

---

## 常见问题

**`just log-cat <requestId>` 没有输出**

后端跑的是 Docker 容器模式，`just log-cat` 只查本地 `logs/server.log`。改用：

```bash
docker exec modelcraft-backend grep "<requestId>" /app/logs/server.log | jq .
```

**requestId 在测试输出里找不到**

确认测试版本包含 `setLastOpenDataResult` 的打印逻辑（`deterministic-runtime-rls.steps.ts` 中）。旧版测试不打印 requestId，需升级 step definitions。

**admin 场景的 SQL 出现 `AND (user_id = ?) OR (1=1) AND (user_id = ?) OR (1=1)`（双重 AND）**

这是 `count` 的 WHERE 子句中 RLS USING filter 与 where input 叠加的结果，属于预期行为——count 的 where 参数 + RLS filter 均参与。

**tasks DELETE 步骤为什么先 SELECT 再 DELETE？**

后端在执行 DELETE 前先 SELECT 确认记录对当前用户可见（USING filter），再执行实际 DELETE LIMIT 1。若 SELECT 返回 0 行则不执行 DELETE，返回「删除未生效」。
