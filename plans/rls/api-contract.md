# RLS API Contract

> **范围**: GraphQL Design-time 变更、Runtime 行为约定、HTTP 认证变更、错误码规范
> **Schema 路径**: `modelcraft-backend/api/graph/project/schema/`

---

## 1. Design-time GraphQL 变更（Project Schema）

### 1.1 `FormatType` 枚举新增值

文件：`field.graphql`

```graphql
enum FormatType {
  # 基于STRING
  STRING
  UUID
  DATE
  DATETIME
  TIME

  # 基于NUMBER
  NUMBER
  INTEGER
  DECIMAL

  # 基于BOOLEAN
  BOOLEAN
  RELATION

  # 枚举类型
  ENUM

  # RLS 归属字段（新增）
  END_USER_REF
}
```

**约束**（后端不变量，不在 Schema 中表达）：
- `format = END_USER_REF` 时，字段名固定为 `"owner"`，`nonNull = true`，不可自定义。
- DB 层自动生成外键约束：`REFERENCES private_{projectSlug}.users(id)`。

---

### 1.2 `addFields` — 新增 `EndUserRefAlreadyExists` 错误类型

文件：`field.graphql`

```graphql
# 新增错误类型
type EndUserRefAlreadyExists implements Error {
  message: String!
  code: String!
}

# 扩展 AddFieldsError union（新增成员）
union AddFieldsError = InvalidInput | EndUserRefAlreadyExists
```

触发条件：同一 Model 已存在 `format = END_USER_REF` 的字段，再次调用 `addFields` 传入另一个 `END_USER_REF` 字段时返回此错误。

错误响应结构（`AddFieldItemResult.error` 填充）：

```json
{
  "name": "owner",
  "success": false,
  "error": {
    "__typename": "EndUserRefAlreadyExists",
    "message": "每个 Model 只允许一个归属字段",
    "code": "END_USER_REF_ALREADY_EXISTS"
  }
}
```

---

### 1.3 `removeField` — `RemoveFieldPayload` 新增 `warning` 字段

文件：`field.graphql`

```graphql
enum RemoveFieldWarning {
  RLS_WILL_DISABLE
}

type RemoveFieldPayload {
  model: Model
  error: RemoveFieldError
  warning: RemoveFieldWarning  # 新增：删除 END_USER_REF 字段时填充
}
```

语义：
- 删除 `format = END_USER_REF` 的字段时，`warning = RLS_WILL_DISABLE`，`error = null`，`model` 正常返回（操作已执行）。
- 删除普通字段时，`warning = null`。
- 前端根据 `warning` 决定是否在**操作执行前**弹出二次确认弹窗。

> ⚠️ **前端交互顺序**：弹窗 → 用户确认 → 调用 `removeField` → 后端执行删除并返回 `warning`（此时 warning 仅作前端信息记录，不阻塞删除）。  
> 后端不做"预检查"接口，warning 在删除后随 payload 返回。

---

### 1.4 `Model` 类型新增 `isRLSEnabled` 和 `rlsPolicy` 字段

文件：`model.graphql`

```graphql
type Model implements Node {
  id: ID!
  projectSlug: String!
  name: String!
  title: String!
  description: String!
  databaseName: String!
  storageType: String!
  displayField: String
  isRLSEnabled: Boolean!     # 新增：存在 END_USER_REF 字段则为 true
  rlsPolicy: ModelRLSPolicy  # 新增：null = 无 owner 字段，无 Policy
  fields: [Field!]!
  group: ModelGroup!
  dbTable: DbTableStatus
  jsonSchema: String
  createdAt: String!
  updatedAt: String!
}
```

`isRLSEnabled` 计算规则：`model.fields` 中存在任意 `format = END_USER_REF` 的字段 → `true`，否则 `false`。该字段为只读派生值，无对应 mutation。

`rlsPolicy` 语义：`isRLSEnabled = true` 时返回 Policy 对象；`isRLSEnabled = false` 时返回 `null`（无 owner 字段，无 Policy）。

---

### 1.5 Policy 相关类型（新增）

文件：`model.graphql`（或独立 `rls.graphql`）

```graphql
type ModelRLSPolicy {
  selectPredicate  : String!   # JSON 表达式
  insertCheck      : String!   # JSON 表达式
  updatePredicate  : String!   # JSON 表达式
  updateCheck      : String!   # JSON 表达式
  deletePredicate  : String!   # JSON 表达式
  preset           : RLSPreset # nullable：当前五件套组合对应的 Preset；自定义组合时为 null
}

enum RLSPreset {
  READ_WRITE_OWNER      # 五件套均为 {"owner":{"_eq":{"_auth":"uid"}}}（新建 Model 默认）
  READ_ALL_WRITE_OWNER  # select=true, 其余={"owner":{"_eq":{"_auth":"uid"}}}
  READ_ALL              # select=true, 其余=false
  READ_WRITE_ALL        # 五件套均为 true（⚠️ 高危）
  NO_ACCESS             # 五件套均为 false
}

input SetModelRLSPolicyInput {
  modelId          : ID!
  selectPredicate  : String!   # JSON 表达式
  insertCheck      : String!   # JSON 表达式
  updatePredicate  : String!   # JSON 表达式
  updateCheck      : String!   # JSON 表达式
  deletePredicate  : String!   # JSON 表达式
}

type SetModelRLSPolicyPayload {
  model: Model
  error: SetModelRLSPolicyError
}

union SetModelRLSPolicyError = PolicyNotFound | InvalidInput | RLSInvalidExpr

# PolicyNotFound：目标 Model 无 owner 字段，无法设置 Policy
type PolicyNotFound implements Error {
  message: String!
  code:    String!
}

# RLSInvalidExpr：Policy JSON 表达式不合法（编译期校验失败）
type RLSInvalidExpr implements Error {
  message: String!
  code:    String!
  details: [String!]   # 具体校验错误列表
}

extend type Mutation {
  setModelRLSPolicy(input: SetModelRLSPolicyInput!): SetModelRLSPolicyPayload!
  validateRLSExpr(input: ValidateRLSExprInput!): ValidateRLSExprPayload!
}

# 实时校验 RLS 表达式（编译期校验，不修改 Policy）
input ValidateRLSExprInput {
  projectId    : ID!
  modelId      : ID!
  operation    : RLSOperation!   # 校验哪种操作的谓词
  expression   : String!         # JSON 表达式字符串
}

enum RLSOperation {
  SELECT_PREDICATE
  INSERT_CHECK
  UPDATE_PREDICATE
  UPDATE_CHECK
  DELETE_PREDICATE
}

type ValidateRLSExprPayload {
  valid   : Boolean!
  errors  : [String!]   # 校验错误列表，valid=true 时为空
}
```

**注意**：`ExprType` 枚举已删除，字段类型改为 `String`（存储 JSON）。Preset 快捷方式由前端映射到对应 JSON 后调用 `setModelRLSPolicy`（不新增独立 setPreset mutation）。

---

### 1.6 auth_schema 相关 API（新增）

文件：`project.graphql`（或 `rls.graphql`）

```graphql
type AuthVariable {
  name:   String!
  source: String!   # "jwt.tenant_id" 等
  type:   String!   # "uuid" | "string" | "integer"
}

type ProjectAuthSchema {
  projectId : ID!
  variables : [AuthVariable!]!
}

input AuthVariableInput {
  name:   String!
  source: String!
  type:   String!
}

input SetProjectAuthSchemaInput {
  projectId  : ID!
  variables  : [AuthVariableInput!]!
}

extend type Query {
  getProjectAuthSchema(projectId: ID!): ProjectAuthSchema
}

extend type Mutation {
  setProjectAuthSchema(input: SetProjectAuthSchemaInput!): ProjectAuthSchema!
}
```

---

## 2. Runtime GraphQL 行为约定

> Runtime Schema 为动态生成，以下为生成规则变更，不修改静态 `.graphql` 文件。

### 2.1 `owner` 字段可见性规则

| 调用者身份 | Query 返回类型（`XxxObject`） | CreateInput / UpdateInput |
|----------|--------------------------|--------------------------|
| EndUser（`iss = mc-enduser`） | 包含 `owner: String` 字段，值固定为当前 EndUser ID | **不生成** `owner` 字段 |
| Developer（`iss = mc-developer`） | 包含 `owner: String` 字段，完整暴露 | 包含 `owner: String` 字段，无限制 |

EndUser 视角的 Schema 示例（以 `orders` Model 为例）：

```graphql
# Query 返回类型 —— owner 可见
type OrdersObject {
  id: ID!
  title: String
  owner: String   # 值为当前 EndUser 的 ID（只读，系统注入）
}

# CreateInput —— owner 不暴露
input CreateOrdersInput {
  title: String
  # owner 字段不存在，由系统自动填充
}

# UpdateInput —— owner 不暴露
input UpdateOrdersInput {
  title: String
  # owner 字段不存在
}
```

---

### 2.2 `CreateOne` 自动填充语义（writeScope=OWNER）

- EndUser 调用 `CreateOne`（writeScope=OWNER）：后端强制将 `owner` 设置为 JWT 中的 `endUserId`，**覆盖**任何来源。
- `owner` 不允许为 `null`；若 `endUserId` 缺失（理论上不可达，因认证已保证），返回 500。
- writeScope=NONE 时：`CreateOne` 直接返回 `PERMISSION_DENIED`，不执行插入。

---

### 2.3 WHERE 注入规范（基于 Policy）

**触发条件**：
- JWT `iss = mc-enduser`，`endUserId` 已注入 request context
- 目标 Model **有 Policy**（`model.rlsPolicy != null`）→ 按五件套 ExprType 执行
- 目标 Model **无 Policy**（`model.rlsPolicy == null`）→ **DENY ALL**（不是全量访问）

> ⚠️ **Default Deny**：无 Policy ≠ 全量访问。EndUser 调用无 Policy 的 Model 一律被拒绝。

**USING 谓词（selectPredicate / updatePredicate / deletePredicate）**：
- USING 不通过 → 行"不存在"，**静默过滤**，不报错，不返回 403

| 表达式 | 注入规则 | USING 失败行为 |
|--------|---------|--------------|
| `true`（常量） | 不注入 WHERE，全量操作 | — |
| JSON 表达式（如 `{"owner":{"_eq":{"_auth":"uid"}}}`） | PolicyExecutor.ToSQL() 生成参数化 WHERE 子句 | 静默（0行/0受影响，不报错） |
| `false`（常量） | SQL WHERE 追加 `AND 1=0` | 静默（空集/0行，不报错） |

| 谓词 | 适用操作 |
|------|---------|
| `selectPredicate` | FindMany / FindFirst / Count |
| `updatePredicate` | UpdateOne（USING 过滤目标行） |
| `deletePredicate` | DeleteOne（USING 过滤目标行） |

**WITH CHECK 谓词（insertCheck / updateCheck）**：
- WITH CHECK 不通过 → **整个操作失败**，返回 `RLS_CHECK_VIOLATION`

| 表达式 | 校验规则 | CHECK 失败行为 |
|--------|---------|--------------|
| `true`（常量） | 允许写入 | — |
| JSON 表达式（含 owner 等于 uid 约束等） | 应用层评估表达式（不含 _exists/_ref） | `RLS_CHECK_VIOLATION`，操作失败 |
| `false`（常量） | 始终拒绝 | `RLS_CHECK_VIOLATION`，操作失败 |

| 谓词 | 适用操作 |
|------|---------|
| `insertCheck` | CreateOne（写入前校验） |
| `updateCheck` | UpdateOne（写入后校验更新结果） |

**交集语义**（selectPredicate=OWNER_EQUALS_USER）：EndUser 在 `where` 中显式传 `owner: <other_id>`，系统注入与用户条件取 AND，结果为空集，返回空列表（不报错，不返回 403）。

---

## 3. HTTP API 变更

### 3.1 JWT Issuer 迁移

| 类型 | 旧 `iss` 值 | 新 `iss` 值 | 说明 |
|------|-----------|-----------|------|
| Developer JWT | `"modelcraft"` | `"mc-developer"` | 需同步迁移现有颁发逻辑 |
| EndUser JWT | `"modelcraft"` | `"mc-enduser"` | 原已独立，本期明确 iss 值 |

迁移期间后端**同时接受** `"modelcraft"` 和 `"mc-developer"`，待前端完成切换后移除旧值支持（具体时间节点在独立迁移计划中约定）。

---

### 3.2 Runtime Endpoint 认证规则

**Endpoint 格式**：`/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}`

| 请求类型 | 响应 |
|---------|------|
| 无 JWT 或 JWT 格式非法 | `401 Unauthorized` |
| `iss = mc-developer` | `401 Unauthorized` |
| `iss = mc-enduser`，签名合法，未过期 | `200`，正常执行，`endUserId` 注入 context |

401 响应体格式：

```json
{
  "errors": [
    {
      "message": "Unauthorized: Runtime endpoint requires EndUser JWT (iss: mc-enduser)",
      "extensions": {
        "code": "RUNTIME_UNAUTHORIZED"
      }
    }
  ]
}
```

---

### 3.3 EndUser Login 接口 JWT 变更

EndUser 登录接口颁发的 JWT Payload 变更：

```json
{
  "iss": "mc-enduser",
  "sub": "<endUserId>",
  "exp": 1234567890,
  "projectSlug": "<projectSlug>"
}
```

> `iss` 从 `"modelcraft"` 迁移到 `"mc-enduser"`。

---

### 3.4 Design-time GraphQL 认证规则（不变，明确记录）

**Endpoint 格式**：`/graphql/org/{orgName}/`、`/graphql/org/{orgName}/project/{projectSlug}/`

| 请求类型 | 响应 |
|---------|------|
| `iss = mc-developer`（含旧值 `"modelcraft"` 迁移期） | `200`，正常执行 |
| `iss = mc-enduser` | `401 Unauthorized` |
| 无 JWT 或格式非法 | `401 Unauthorized` |

---

## 4. 错误码规范

格式与现有 `bizerrors` 一致：`SCREAMING_SNAKE_CASE`，随 GraphQL 错误 `extensions.code` 返回。

| 错误码 | 触发场景 | 返回位置 |
|-------|---------|---------|
| `END_USER_REF_ALREADY_EXISTS` | 同一 Model 添加第二个 `END_USER_REF` 字段 | `AddFieldItemResult.error`（`EndUserRefAlreadyExists.code`） |
| `RUNTIME_UNAUTHORIZED` | Runtime endpoint 收到非 EndUser JWT 或无 JWT | HTTP 401，`errors[].extensions.code` |
| `RLS_CHECK_VIOLATION` | WITH CHECK 不通过（insertCheck/updateCheck 校验失败）→ 整个操作失败 | Runtime GraphQL `errors[].extensions.code`（新增） |
| `PERMISSION_DENIED` | 无 Policy 时 EndUser 调用 Runtime（Default Deny） | Runtime GraphQL `errors[].extensions.code`（新增） |
| `POLICY_NOT_FOUND` | 调用 setModelRLSPolicy 时 Model 无 owner 字段 | `SetModelRLSPolicyPayload.error`（`PolicyNotFound.code`） |
| `RLS_INVALID_EXPR` | Policy JSON 表达式编译校验失败 | `SetModelRLSPolicyPayload.error`（`RLSInvalidExpr.code`）；`ValidateRLSExprPayload.errors` |

> **注意**：USING 失败（selectPredicate / updatePredicate / deletePredicate 不通过）**不产生错误码**，为静默过滤（0行/0受影响）。只有 WITH CHECK 失败才抛 `RLS_CHECK_VIOLATION`。

---

## 5. 变更文件清单

| 文件 | 变更类型 | 内容摘要 |
|------|---------|---------|
| `api/graph/project/schema/field.graphql` | 修改 | `FormatType` 新增 `END_USER_REF`；`AddFieldsError` union 新增 `EndUserRefAlreadyExists`；新增 `EndUserRefAlreadyExists` 类型；`RemoveFieldPayload` 新增 `warning` 字段；新增 `RemoveFieldWarning` 枚举 |
| `api/graph/project/schema/model.graphql` | 修改 | `Model` 类型新增 `isRLSEnabled: Boolean!` 和 `rlsPolicy: ModelRLSPolicy`；新增 `ModelRLSPolicy` 类型（五件套字段类型为 `String!`）；新增 `RLSPreset` 枚举；新增 `SetModelRLSPolicyInput`（五件套类型为 `String!`）、`SetModelRLSPolicyPayload`、`PolicyNotFound`、`RLSInvalidExpr` 类型；新增 `setModelRLSPolicy`、`validateRLSExpr` mutation；**删除** `ExprType`、`RLSReadScope`、`RLSWriteScope` 枚举 |
| `api/graph/project/schema/rls.graphql`（新增） | 新增 | `AuthVariable`、`ProjectAuthSchema`、`SetProjectAuthSchemaInput`；`getProjectAuthSchema` query；`setProjectAuthSchema` mutation；`ValidateRLSExprInput`、`ValidateRLSExprPayload`、`RLSOperation` 枚举 |
| Runtime Schema 生成逻辑 | 行为变更 | EndUser 视角屏蔽 CreateInput/UpdateInput 中的 `owner` 字段 |
| Runtime 认证中间件 | 行为变更 | 仅接受 `iss = mc-enduser`，其余 401 |
| Runtime RLS 注入逻辑 | 行为变更 | 从 Policy JSON 读取，PolicyCompiler.Compile() + PolicyExecutor.ToSQL() 生成参数化 SQL |
| EndUser Login 颁发逻辑 | 行为变更 | JWT `iss` 改为 `"mc-enduser"` |
| Developer JWT 颁发逻辑 | 行为变更 | JWT `iss` 改为 `"mc-developer"` |
