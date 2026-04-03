---
name: backend-debug
description: >
  排查和修复 ModelCraft 后端错误。当用户提供 GraphQL 响应中的错误（包含 errors 数组、requestId、message 字段），
  或者描述后端报错、接口异常、服务崩溃时，使用此 skill。
  触发场景包括：
  (1) 用户粘贴了带 errors/requestId 的 JSON 响应，
  (2) 说"后端报错了"、"接口返回错误"、"帮我看看这个错误"、"定位问题"，
  (3) 说"使用 just log"、"查看日志"、"找到错误原因"后想修复，
  (4) 任何需要通过日志定位再修复代码的后端问题。
  遇到上述情形时，主动使用此 skill，即便用户没有明确说"debug"。
---

# 后端问题排查与修复

目标：**先用日志精准定位根因，再选择合适的验证手段（单测 / 数据库），最后最小化修复**。

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

## 第二步：用 requestId 查日志链

```bash
cd modelcraft-backend
just log-cat <requestId>
```

这会从 `logs/server.log` 中按 `request_id` 字段精确匹配，过滤出该请求的全部日志行，完整还原请求生命周期。

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

如果需要看实时日志流（没有 requestId 时）：
```bash
just logs
```

---

## 第三步：定位源码

从日志的 `caller` 字段（如 `app/modeldesign/field_app.go:87`）直接定位文件和行号。

如果 caller 不够直接，用错误消息搜索：

```bash
grep -r "failed to introspect table" modelcraft-backend/internal/ --include="*.go"
```

**理解代码路径**：错误从 Infrastructure 层产生，经过 Application 层包装，最终在 Interfaces 层记录 stack。找根因要从最内层（Infrastructure/Domain）开始读。

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

---

## 第五步：修复并确认

修复后按顺序验证：

```bash
cd modelcraft-backend

# 1. 确认编译通过
just build

# 2. 跑相关包的单元测试（如果 Domain 层有改动）
just test-unit-pkg ./internal/domain/<domain>/...

# 3. 重启服务，用原请求复现
just run force=true

# 4. 确认问题消失
just log-cat <requestId>   # 跑新请求后用新 requestId 查
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

- `just log-cat` 依赖 `logs/server.log` 存在。服务未运行过或日志被清理时，直接看 `just logs`
- 某些启动错误（配置缺失、DB 连不上）没有 requestId，直接读 `just logs` 即可
- 修复时遵守分层规则（避免引入新问题）：
  - **Repository 层**只返回 `shared.RepositoryError`，不返回 `*BusinessError`
  - **Application 层**把 nil 结果转成 `bizerrors.NewErrorFromContext()`，把 DB 错误转成 `bizerrors.ConvertRepositoryError()`
  - **Interfaces 层**才打 `logfacade.Stack(err)`
- 如果修复涉及 SQL（`db/queries/*.sql`），跑 `just generate-sqlc` 重新生成代码
- 如果修复涉及 GraphQL schema（`api/graph/*/schema/*.graphql`），跑 `just generate-gql`（**不要**跑 `just clean-gql`，会删除 resolver 实现）
