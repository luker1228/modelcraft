# Runtime WHERE 自动注入（RLS 核心执行）

> 依赖：`01-enduserref-field.md`、`02-model-owner-lifecycle.md`、`03-runtime-jwt-auth.md`、`05-policy-configuration.md`
> 对应主 PRD 章节：M4

---

## 背景

这是 RLS 的核心执行层。当 EndUser 调用 Runtime GraphQL 时，系统根据 **ModelRLSPolicy** 的五件套 JsonExpr 决定注入逻辑。Policy 每个字段存储一个 GraphQL JSON 表达式，由三层纯函数编译执行：`PolicyValidator → PolicyCompiler → PolicyExecutor`。

本子页依赖：
- `01-enduserref-field.md`：`EndUserRef` 字段和 `isRLSEnabled()` 逻辑
- `02-model-owner-lifecycle.md`：Model 创建/删除 owner 字段后的 RLS 状态
- `03-runtime-jwt-auth.md`：`endUserId` 从 JWT 中提取并注入 context
- `05-policy-configuration.md`：Policy 数据模型、Preset、JsonExpr 语义、auth_schema

---

## 需求

### 触发条件

RLS 仅在请求携带合法 EndUser JWT（`iss = mc-enduser`），`endUserId` 已注入 context 时触发。

- **有 Policy**（`model.getPolicy() != nil`）→ 按五件套 ExprType 执行注入
- **无 Policy**（`model.getPolicy() == nil`）→ **DENY ALL**（不是开放模式）

> ⚠️ **Default Deny**：无 Policy ≠ 全量访问。无 Policy 的 EndUser 请求一律被拒绝。

---

### USING 谓词（selectPredicate / updatePredicate / deletePredicate）

USING 语义：行"不存在"→ 静默过滤，**不报错**。

**编译与注入流程**：
1. 编译期（保存 Policy 时）：`PolicyValidator.Validate()` 校验 JSON 结构、字段白名单、`_auth` 变量白名单
2. 运行期（每次请求）：`PolicyCompiler.Compile()` 将 JSON 编译为参数化 SQL 片段；`PolicyExecutor.ToSQL(compiled, authCtx)` 绑定当前用户 ID，生成最终 `WHERE` 子句

| 表达式 | 注入规则 | 失败行为 |
|--------|---------|---------|
| `true` | 不注入 WHERE，全量操作 | — |
| `{"owner":{"_eq":{"_auth":"uid"}}}` | PolicyExecutor.ToSQL → `AND owner = ?`，args=[endUserId] | 行不存在 → 静默（0行/0受影响，不报错） |
| `false` | 追加 `AND 1=0` | 静默返回空/0行，不报错 |
| 通用 JSON 表达式 | 递归编译为参数化 SQL WHERE 子句 | 同上，静默 |

适用操作：

| 谓词 | 适用操作 |
|------|---------|
| `selectPredicate` | FindMany / FindFirst / Count |
| `updatePredicate` | UpdateOne（USING 过滤目标行） |
| `deletePredicate` | DeleteOne（USING 过滤目标行） |

**PREDICATE 允许的操作符**：比较 ✅、`_and/_or/_not` ✅、`_exists` ✅、`_ref` ✅

### WITH CHECK 谓词（insertCheck / updateCheck）

WITH CHECK 语义：校验写入数据是否符合策略，不通过 → **整个操作失败，抛 `RLS_CHECK_VIOLATION`**。

| 表达式 | 校验规则 | 失败行为 |
|--------|---------|---------|
| `true` | 允许写入 | — |
| `{"owner":{"_eq":{"_auth":"uid"}}}` | 应用层校验写入行的 owner = endUserId | 违反 → `RLS_CHECK_VIOLATION`，操作失败 |
| `false` | 始终拒绝 | 直接 `RLS_CHECK_VIOLATION`，操作失败 |
| 通用 JSON 表达式（不含 _exists/_ref） | 应用层评估表达式 | 违反 → `RLS_CHECK_VIOLATION` |

适用操作：

| 谓词 | 适用操作 |
|------|---------|
| `insertCheck` | CreateOne（写入前校验） |
| `updateCheck` | UpdateOne（写入后校验更新结果） |

**CHECK 允许的操作符**：比较 ✅、`_and/_or/_not` ✅、`_exists` ❌、`_ref` ❌

### CreateOne 特殊规则（当 insertCheck 含 `owner = _auth.uid` 约束时）

- 强制将 `owner` 设置为 JWT 中的 `endUserId`，**覆盖**用户传入的任何值
- `owner` 不允许为 `null`
- 前端 input 中无 `owner` 字段（Schema 层已屏蔽）

### WHERE 条件交集语义

selectPredicate=OWNER_EQUALS_USER 时，EndUser 显式传入 `where: { owner: <other_id> }` 时，系统注入条件与用户条件取**交集**（AND），结果为空集。这是预期行为，不报错，不返回 403。

---

## 验收标准

### AC-3：行级过滤生效

- [ ] selectPredicate=OWNER_EQUALS_USER：EndUser A 查询，只返回 `owner = A` 的行（即使不传任何 filter）
- [ ] selectPredicate=ALWAYS_TRUE：EndUser A 查询，返回全量数据
- [ ] selectPredicate=ALWAYS_FALSE：EndUser A 查询，返回空集（静默，不报错）
- [ ] selectPredicate=OWNER_EQUALS_USER：EndUser A 显式传 `where: { owner: B_id }` → 返回空，而不是 B 的数据
- [ ] 无 Policy 的 Model，EndUser → DENY ALL（不是全量访问）
- [ ] USING 失败（OWNER_EQUALS_USER/ALWAYS_FALSE）→ 静默 0 行/0 受影响，不报错，不返回 403

### AC-4：CreateOne 自动填充（insertCheck=OWNER_EQUALS_USER）

- [ ] EndUser A 调用 CreateOne 不传 `owner` → 自动填充为 A 的 ID
- [ ] EndUser A 调用 CreateOne 故意传 `owner = B` → 被**强制覆盖**为 A 的 ID

### AC-5：Mutation 保护（依赖 Policy）

- [ ] updatePredicate=OWNER_EQUALS_USER：EndUser A 调用 UpdateOne（指向 `owner = B` 的记录）→ 静默 0 行受影响（非 403，不报错）
- [ ] deletePredicate=OWNER_EQUALS_USER：EndUser A 调用 DeleteOne（指向 `owner = B` 的记录）→ 静默 0 行受影响（非 403，不报错）
- [ ] insertCheck=ALWAYS_FALSE：EndUser A 调用 CreateOne → 返回 `RLS_CHECK_VIOLATION`，操作失败
- [ ] updateCheck=OWNER_EQUALS_USER：EndUser A 调用 UpdateOne 设置 `owner = B` → 返回 `RLS_CHECK_VIOLATION`
- [ ] CHECK 失败 → 返回 `RLS_CHECK_VIOLATION`，HTTP 200，GraphQL error

### AC-6（RLS 部分）

- [ ] EndUser JWT 调用 Runtime → RLS 注入生效，行为符合上述规则

### AC-8：向后兼容

- [ ] 升级后，现有无 EndUserRef 字段的 Model，EndUser DENY ALL（无 Policy = 拒绝）
- [ ] 现有 BDD / 集成测试全部通过，无需修改

---

## 用户故事对应

**Story 1**（部分）：新建 Model，零配置开箱即用
> 默认 Policy = 读写自己（OWNER/OWNER），终端用户自动隔离，CreateOne 时 owner 字段自动填充。

**Story 2**（部分）：删除 owner 字段，关闭 RLS
> 删除 `EndUserRef` 字段后，Policy 同步删除，`model.getPolicy() == nil`；不再执行基于 Policy 的 WHERE 编译注入，但对 EndUser 访问按 Default Deny 拒绝。

**Story 3**：终端用户只能访问自己的数据
> readScope=OWNER 时，无论 EndUser 传什么查询条件，WHERE 注入保证只能访问 `owner = 自己 ID` 的行。

---

## 领域模型关键元素

```
RLSFilter (Value Object)
  + selectPredicate  : JsonExpr   ← 来自 ModelRLSPolicy（JSON String）
  + insertCheck      : JsonExpr   ← 来自 ModelRLSPolicy（JSON String）
  + updatePredicate  : JsonExpr   ← 来自 ModelRLSPolicy（JSON String）
  + updateCheck      : JsonExpr   ← 来自 ModelRLSPolicy（JSON String）
  + deletePredicate  : JsonExpr   ← 来自 ModelRLSPolicy（JSON String）
  + fieldName        : String = "owner"
  + endUserId        : EndUserID

PolicyValidator (Domain Service)
  + Validate(json, modelSchema, authSchema) → []ValidationError
    → JSON Schema 校验、字段白名单、_auth 变量校验
    → insertCheck/updateCheck 不含 _exists/_ref

PolicyCompiler (Domain Service)
  + Compile(json) → CompiledPolicy
    → 递归解析 JSON 为参数化 SQL 片段

PolicyExecutor (Domain Service)
  + ToSQL(compiled, authCtx) → (string, []any)
    → 绑定运行期 authCtx 生成参数化 WHERE + args[]

RLSResolver (Domain Service)
  + resolve(identity: EndUserIdentity, model: Model): RLSFilter?
    → identity.isEndUser() == false → null（不过滤，Developer 开放访问）
    → model.getPolicy() == nil → DENY ALL（无 Policy = Default Deny）
    → 否则 → RLSFilter { 五件套 JsonExpr, endUserId }
```

---

## 不做什么（本子页 Out of scope）

- 数据库原生 RLS（应用层实现，保持对外部 MySQL 集群的兼容性）
- 审计日志（访问记录和策略变更历史不在本期）
