---
name: backend-debug
description: >
  排查 ModelCraft 服务端问题、查询后端日志（backend + gateway + apisix + agent）。
  以下任一场景均应触发：
  (1) 查日志 / 看日志 / 查 requestId / grep 日志 / tail 日志 / 查 server.log / 查 backend 日志；
  (2) 用户粘贴了带 errors / requestId 的后端或 gateway 响应或日志内容；
  (3) 明确说"后端报错了""gateway 报错了""接口返回错误""帮我定位服务端问题"；
  (4) 明确要求查看 backend / gateway / apisix / agent 日志并定位根因；
  (5) 用户提供 requestId，想知道发生了什么。
  只要与后端日志或服务端排错有关，就应使用本 skill。
  不触发：纯功能开发、重构、代码讲解、部署配置修改、前端问题。
---

# 服务端问题排查与修复（Backend + Gateway）

目标：**先用日志精准定位根因，再选择合适的验证手段（单测 / 数据库），最后最小化修复**。

硬规则：

- 排查问题必须以**定位根因**为目标，不接受只看到表面报错就直接改代码。
- 如果现有日志**无法定位根因**，先在最可疑的边界补日志，再重启服务、重放请求。
- 重放后拿新的 `requestId`，继续按 `apisix -> backend -> agent` 三段链路查，直到根因收敛。

---

## 0）环境说明

排查前先判断当前运行模式，后续命令按对应环境执行。

| 环境 | 运行方式 | 日志位置 | 查日志命令 |
|------|---------|---------|-----------|
| **docker**（默认） | `docker-compose up` | backend: `/app/logs/server.log`；agent: `/app/logs/agent.log`；apisix: `/usr/local/apisix/logs/access.log` / `/usr/local/apisix/logs/error.log` | `docker exec modelcraft-backend sh -c "grep '<requestId>' /app/logs/server.log"` |
| **dev**（本地调试） | `just run` | 宿主机 `logs/server.log` | `cd modelcraft-backend && just log-cat <requestId>` |

快速确认当前状态：

```bash
docker ps --format '{{.Names}}' | grep modelcraft
```

- 如果 `modelcraft-backend` 容器在跑 → **docker 模式**
- 如果容器不在、但服务能访问 → 可能是 **dev 模式**（`just run`）

---

## 第一步：提取关键信息

从用户提供的错误中找：
- `requestId`（在 `extensions.requestId` 字段）
- `message`（错误描述，往往是错误链的末端）
- `path`（出错的 GraphQL 操作名）

```json
{
  "errors": [{"message": "failed to introspect table: connection refused", "path": ["importModel"]}],
  "extensions": {"requestId": "aa65e02c-f1fc-4de6-a7cb-5169d92e0cdf"}
}
```

如果用户只描述了现象（如"字段创建失败"），先让服务运行后复现，再拿到 requestId。

---

## 第二步：用 requestId 查日志链（99% 的问题先查 requestId）

默认顺序：

1. 先查 `apisix`，确认请求有没有进网关、状态码是什么
2. 再查 `modelcraft-backend`，确认有没有转发进 Go 服务
3. 如果链路经过 `modelcraft-agent`，再查 agent 文件日志

先串链路，再猜代码。不要上来就翻源码。

### docker 模式（默认）

```bash
# backend — 容器内 grep 文件日志
docker exec modelcraft-backend sh -c "grep '<requestId>' /app/logs/server.log"

# agent — 容器内 grep 文件日志
docker exec modelcraft-agent sh -c "grep '<requestId>' /app/logs/agent.log"

# apisix — 容器内 grep access / error 文件日志
docker exec modelcraft-apisix sh -c "grep '<requestId>' /usr/local/apisix/logs/access.log /usr/local/apisix/logs/error.log"
```

### dev 模式（本地 `just run` 调试）

```bash
cd modelcraft-backend
just log-cat <requestId>
```

这会从 `logs/server.log` 中按 `request_id` 字段精确匹配，过滤出该请求的全部日志行。

### 三段链路推荐命令

```bash
# 1. 先看 apisix
docker exec modelcraft-apisix sh -c "grep '<requestId>' /usr/local/apisix/logs/access.log /usr/local/apisix/logs/error.log"

# 2. 再看 backend
docker exec modelcraft-backend sh -c "grep '<requestId>' /app/logs/server.log"

# 3. 如果请求经过 agent，再看 agent
docker exec modelcraft-agent sh -c "grep '<requestId>' /app/logs/agent.log"
```

**错误定位原则：错误发生在最后一个有日志的服务层。**

| apisix | backend | agent | 结论 |
|--------|---------|-------|------|
| 有日志 | 无日志  | —     | 错误在 apisix（请求未到达 backend） |
| 有日志 | 有日志  | 无日志 | 错误在 backend 内部，或该链路未调用 agent |
| 有日志 | 有日志  | 有日志 | 按时间顺序找第一条 `error` / `warn` |
| 无日志 | —       | —     | 请求未进入网关（客户端问题 / 网络问题） |

> requestId 已经串联三段链路的所有日志，grep 只需要 requestId 关键字，不需要查附近行（无需 `-A` / `-B`）。

### 日志格式

日志是 **JSON 格式**，每行一条，关键字段：

```json
{
  "level": "ERROR",
  "timestamp": "2026-04-02T11:15:48.188+0800",
  "caller": "app/modeldesign/field_app.go:87",
  "msg": "failed to create field",
  "request_id": "aa65e02c-f1fc-4de6-a7cb-5169d92e0cdf",
  "error": "CONFLICT.FIELD: field name 'id' already exists",
  "stack": "goroutine 47 [running]:\n..."
}
```

### 读日志的重点

- **找最早的 error/panic**——这是根本原因，后续的 error 通常是它的传播
- **看 `stack` 字段**——只有 Interfaces 层才打 stack，能定位到精确的代码路径
- **看 SQL 日志**——`[SQLC] query ok` 行带有 `sql` 和 `sql_args` 字段，可以直接拿去数据库重现查询
- 如果日志只有“失败了”但没有说明**为什么失败**，结论不是“看不出来”，而是**日志不够，需要补日志**

### 如果当前日志不足以定位根因

按下面顺序处理：

1. 判断卡在哪一层：`apisix`、`backend`、`agent`、DB、外部依赖
2. 在该层的关键边界补日志：入参、分支判断、下游请求、返回状态、错误对象、关键 ID
3. 重启对应服务
4. 重放同一个请求
5. 拿新的 `requestId` 再查三段链路

补日志原则：

- 优先加在**分支点**和**跨服务边界**，不要无脑到处打印
- 日志要能回答“进没进到这里、拿到了什么参数、为什么走这个分支、下游返回了什么”
- 补日志是为了缩小范围；一旦根因确认，应转向修复和验证

---

## 第三步：对准——把证据和判断告诉用户，等确认再继续

查到日志后，**停下来**，向用户报告：

1. **证据**：贴出关键日志行（level、msg、caller、error 字段）
2. **定位层**：错误发生在哪一层（apisix / backend / agent）
3. **猜测根因**：你认为是什么问题，以及为什么这样判断
4. **等用户确认**：明确说"请确认后我再继续"，不要自行推进到修复

示例格式：

> **证据**
> ```
> level=ERROR caller=app/modeldesign/field_app.go:87
> msg="failed to create field"
> error="CONFLICT.FIELD: field name 'id' already exists"
> ```
>
> **定位**：错误在 backend，apisix 和 agent 均无异常日志。
>
> **猜测根因**：字段名 `id` 已存在，触发唯一键冲突。问题在 Application 层的重复校验逻辑或数据库唯一约束。
>
> 请确认是否符合你的预期，确认后我再继续定位源码。

---

## 第四步：定位源码

从日志的 `caller` 字段（如 `app/modeldesign/field_app.go:87`）直接定位文件和行号。

如果 caller 不够直接，用错误消息搜索：

```bash
grep -r "failed to introspect table" modelcraft-backend/internal/ --include="*.go"
```

**理解代码路径**：错误从 Infrastructure 层产生，经过 Application 层包装，最终在 Interfaces 层记录 stack。找根因要从最内层（Infrastructure/Domain）开始读。

### Graphify 辅助理解调用路径（可选）

如果 `graphify-out/graph.json` 存在，可在定位到文件/函数名后，用 graphify 快速理解调用链和依赖关系，无需逐层手动阅读代码：

```bash
# 从错误函数名出发，广度优先查看它被哪些上层调用
/graphify query "<函数名或模块名>"

# 追踪两个组件之间的依赖路径（如 Resolver → Repository）
/graphify path "<ResolverName>" "<RepositoryName>"

# 理解某个核心函数的全部调用关系（快速掌握上下文）
/graphify explain "<FunctionName>"
```

**典型用法**：

- 日志显示 `caller: app/modeldesign/field_app.go:87` → 运行 `/graphify explain "field_app"` 查看该文件的全部入口和依赖，再决定读哪几个文件
- 错误消息涉及多个层（如 "failed to create field: connection refused"）→ 运行 `/graphify path "FieldApp" "DBConnection"` 追踪完整调用链
- 不确定某个错误来自哪个聚合 → 运行 `/graphify query "<ErrorType>"` 找出所有引用它的模块

> 提示：graphify 使用的是 AST 静态分析图，反映的是编译时依赖关系，适合理解模块间结构。运行时的动态错误仍需结合日志 `stack` 字段来确认。

---

## 第四步：选择验证手段

根因定位后，选择合适的方式验证假设——不要急着改代码，先用最轻量的方式确认自己理解对了。

### 4a. 单元测试——验证纯逻辑

**适用**：问题在 Domain 层（业务规则、字段验证、领域计算），不涉及数据库或外部依赖。

Domain 层每个文件通常都有对应的 `_test.go`，可以直接跑：

```bash
cd modelcraft-backend

# 跑单个包（最常用）
just test-unit-pkg ./internal/domain/modeldesign/...

# 跑所有单元测试
just test-unit

# 详细输出（能看到 t.Log 的内容）
just test-unit-verbose

# 快速跑（关掉 race detector，适合频繁迭代）
just test-unit-fast
```

如果现有测试没有覆盖出错的场景，在对应的 `_test.go` 里加一个针对性的 test case（参考同文件的 table-driven 写法），用来验证修复前后的行为变化。

### 4b. 数据库——验证数据状态和 SQL

**适用**：问题与数据状态有关（记录不存在、字段值异常、关联关系错乱），或者需要确认某条 SQL 是否返回预期结果。

**登录数据库**：
```bash
cd modelcraft-backend
just db login
```

**从日志里拿 SQL 直接跑**：

日志中 `[SQLC] query ok` 行包含完整的 `sql` 和 `sql_args`，把 `?` 替换为实际参数后可以直接在 MySQL 执行：

```sql
-- 日志里的 sql: SELECT * FROM field_definitions WHERE model_id = ? AND name = ?
-- sql_args: ["abc123", "id"]
SELECT * FROM field_definitions WHERE model_id = 'abc123' AND name = 'id';
```

**常见数据层排查 SQL**：

```sql
-- 查某个模型的所有字段
SELECT * FROM field_definitions WHERE model_id = '<model_id>';

-- 检查是否有重复键冲突
SELECT name, COUNT(*) FROM field_definitions
WHERE model_id = '<model_id>' GROUP BY name HAVING COUNT(*) > 1;

-- 查外键关联
SELECT * FROM logical_foreign_keys WHERE model_id = '<model_id>';

-- 查枚举关联
SELECT * FROM field_enum_associations WHERE model_id = '<model_id>';

-- 查项目/集群/组织的基本信息
SELECT * FROM projects WHERE slug = '<slug>' AND org_name = '<org_name>';
```

如果不想污染开发数据，切换到测试数据库：
```bash
just db login .env.autotest
```

### 4c. Gateway 专项验证

**适用**：问题发生在鉴权、代理转发、请求头注入、跨服务 requestId 传递。

```bash
docker exec modelcraft-apisix sh -c "grep '<requestId>' /usr/local/apisix/logs/access.log /usr/local/apisix/logs/error.log"
docker exec modelcraft-agent sh -c "grep '<requestId>' /app/logs/agent.log"
```

重点检查：

- gateway `request_start/request_end` 是否成对，状态码是否符合预期。
- 同一 `request_id` 是否在 backend 出现（确认转发链路）。
- 同一 `request_id` 是否在 agent 出现（确认 backend -> agent 调用链路）。
- 是否存在鉴权拦截（如 missing Authorization、invalid token）。
- agent 是否有上游 GraphQL / tool / LLM 调用失败、超时、429、5xx。

---

## 第五步：修复并确认

修复后按顺序验证：

```bash
cd modelcraft-backend

# 1. 确认编译通过
just build

# 2. 跑相关包的单元测试（如果 Domain 层有改动）
just test-unit-pkg ./internal/domain/<domain>/...

# 3. 重启服务
# docker 模式：
cd ../deploy && docker-compose -f compose/docker-compose.local.yml restart modelcraft
# dev 模式：
just run force=true

# 4. 确认问题消失
# docker 模式：
docker exec modelcraft-backend sh -c "grep '<requestId>' /app/logs/server.log"
# dev 模式：
just log-cat <requestId>
```

如果前面还没有定位到根因，不要直接修。先补日志，然后按下面步骤继续：

```bash
# 1. 修改代码补日志

# 2. 重启服务
cd ../deploy && docker-compose -f compose/docker-compose.local.yml restart backend apisix modelcraft-agent

# 3. 重放请求，拿到新的 requestId

# 4. 用新的 requestId 再查三段链路
docker exec modelcraft-apisix sh -c "grep '<newRequestId>' /usr/local/apisix/logs/access.log /usr/local/apisix/logs/error.log"
docker exec modelcraft-backend sh -c "grep '<newRequestId>' /app/logs/server.log"
docker exec modelcraft-agent sh -c "grep '<newRequestId>' /app/logs/agent.log"
```

---

## 常见错误速查

| 错误特征 | 根本原因 | 验证方式 | 定位方向 |
|---------|---------|---------|---------|
| `sql: no rows in result set` | DB 返回空行，代码没用 `ErrNoRows` 处理 | 数据库查询确认记录是否存在 | 找对应的 Repository，检查 `QueryWithSQLErrorHandling` |
| `NOT_FOUND.XXX` | 查询返回 nil，App 层已转为业务错误 | 先查 DB 确认记录存在 | 看 App 层的 nil 检查和 `bizerrors.NewErrorFromContext` |
| `CONFLICT.XXX` / `Duplicate entry` | 唯一键冲突 | 数据库查重复记录 | 确认唯一键定义是否有误，或数据确实重复 |
| `nil pointer dereference` | 指针未判空就解引用 | 单测复现 panic 路径 | 看 `stack` 字段的行号，补 nil 检查 |
| `unsupported Scan, storing []uint8` | sqlc 类型映射不匹配 | — | 查结构体字段，可能需要自定义 Scanner（参考 `pkg/dbgen/` 里的 `StringSlice`） |
| `context deadline exceeded` | 超时（慢查询或死锁） | 数据库执行 `EXPLAIN` | 拿日志里的 SQL 去 DB 分析，排查索引 |
| `OPERATION_DENIED.XXX` | 业务规则拦截 | 单测验证校验逻辑 | `internal/domain/<domain>/` 的校验代码 |
| `panic: runtime error` | 运行时异常 | 单测复现场景 | 从 `stack` 字段堆栈顶层函数开始读 |

---

## 注意事项

- **docker 模式**（默认）：backend 用 `/app/logs/server.log`，agent 用 `/app/logs/agent.log`，apisix 用 `/usr/local/apisix/logs/access.log` 和 `/usr/local/apisix/logs/error.log`。
- **dev 模式**（本地 `just run`）：日志在宿主机 `modelcraft-backend/logs/server.log`，用 `just log-cat` 或 `just logs` 查看。
- **排障默认动作**：只要用户给了 `requestId`，先查 `apisix` -> `backend` -> `agent` 三段，不要先猜代码。
- 某些启动错误（配置缺失、DB 连不上）没有 requestId，这时直接 `tail -f` 对应容器内日志文件。
- 修复时遵守分层规则（避免引入新问题）：
  - **Repository 层**只返回 `shared.RepositoryError`，不返回 `*BusinessError`
  - **Application 层**把 nil 结果转成 `bizerrors.NewErrorFromContext()`，把 DB 错误转成 `bizerrors.ConvertRepositoryError()`
  - **Interfaces 层**才打 `logfacade.Stack(err)`
- 如果修复涉及 SQL（`db/queries/*.sql`），跑 `just generate-sqlc` 重新生成代码
- 如果修复涉及 GraphQL schema（`api/graph/*/schema/*.graphql`），跑 `just generate-gql`（**不要**跑 `just clean-gql`，会删除 resolver 实现）
