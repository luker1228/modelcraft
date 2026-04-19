# RLS API Contract

> 基于 RLS PRD 的后端 GraphQL API 合约设计
> - 主 PRD: `ai-metadata/prd/rls/prd.md`
> - 领域模型: `ai-metadata/prd/rls/rls-domain.puml`

---

## 1. GraphQL Schema Extensions

### 1.1 Enums

#### FormatType (扩展)
在 `api/graph/project/schema/field.graphql` 中扩展现有 `FormatType` 枚举：

```graphql
# 新增 END_USER_REF 格式类型
enum FormatType {
  # ... 现有值保持不变

  # 基于STRING - 新增 EndUserRef 类型
  END_USER_REF
}
```

**说明**:
- `END_USER_REF` 表示该字段存储 EndUser ID，用于 RLS 数据归属
- 字段名固定为 `owner`，一个 Model 最多只能有一个该类型字段
- 数据库层生成外键约束: `REFERENCES private_{projectSlug}.users(id)`

#### RLSPreset (新增)
在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# RLS 预设策略枚举
enum RLSPreset {
  """
  默认策略：读写自己
  五件套均为 {"owner":{"_eq":{"_auth":"uid"}}}
  """
  READ_WRITE_OWNER

  """
  读取全部，写自己
  selectPredicate=true，其余为 OWNER_EQUALS_USER
  """
  READ_ALL_WRITE_OWNER

  """
  只读全部
  selectPredicate=true，其余为 false
  """
  READ_ALL

  """
  读写全部（⚠️ 高危策略）
  五件套均为 true
  """
  READ_WRITE_ALL

  """
  无访问权限
  五件套均为 false
  """
  NO_ACCESS
}
```

#### AuthVariableType (新增)
在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# Auth Schema 变量类型
enum AuthVariableType {
  UUID
  STRING
  INTEGER
}
```

---

### 1.2 Types

#### ModelRLSPolicy (新增)
在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# Model RLS 策略类型
type ModelRLSPolicy {
  """
  所属模型 ID
  """
  modelId: ID!

  """
  SELECT USING 谓词（JSON 字符串）
  """
  selectPredicate: String!

  """
  INSERT WITH CHECK 谓词（JSON 字符串）
  """
  insertCheck: String!

  """
  UPDATE USING 谓词（JSON 字符串）
  """
  updatePredicate: String!

  """
  UPDATE WITH CHECK 谓词（JSON 字符串）
  """
  updateCheck: String!

  """
  DELETE USING 谓词（JSON 字符串）
  """
  deletePredicate: String!

  """
  当前策略匹配的 Preset，自定义组合时返回 null
  """
  preset: RLSPreset

  """
  创建时间
  """
  createdAt: String!

  """
  更新时间
  """
  updatedAt: String!
}
```

#### AuthVariable (新增)
在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# Auth Schema 变量定义
type AuthVariable {
  """
  变量名（如 "tenant_id"）
  """
  name: String!

  """
  JWT 来源路径（如 "jwt.tenant_id"）
  """
  source: String!

  """
  变量类型
  """
  type: AuthVariableType!
}
```

#### ProjectAuthSchema (新增)
在 `api/graph/org/schema/project.graphql` 中定义（扩展 Project 配置）：

```graphql
# Project 认证变量配置
type ProjectAuthSchema {
  """
  认证变量列表（不含内置 uid）
  """
  variables: [AuthVariable!]!
}
```

**说明**: `ProjectAuthSchema` 属于 Org Domain，在 `project.graphql` 中添加。

#### ValidationResult (新增)
在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# RLS 表达式校验结果
type ValidationResult {
  """
  是否校验通过
  """
  valid: Boolean!

  """
  校验错误信息列表（valid=false 时返回）
  """
  errors: [ValidationError!]
}

# 校验错误详情
type ValidationError {
  """
  错误位置（如 "selectPredicate"、"insertCheck._exists"）
  """
  path: String!

  """
  错误描述
  """
  message: String!

  """
  错误码
  """
  code: String!
}
```

#### Model 扩展
在 `api/graph/project/schema/model.graphql` 中扩展 `Model` 类型：

```graphql
type Model implements Node {
  # ... 现有字段保持不变

  """
  RLS 策略配置（无 owner 字段时返回 null）
  """
  rlsPolicy: ModelRLSPolicy
}
```

#### Project 扩展
在 `api/graph/org/schema/project.graphql` 中扩展 `Project` 类型：

```graphql
type Project implements Node {
  # ... 现有字段保持不变

  """
  认证变量配置（用于 RLS 表达式中的 _auth 引用）
  """
  authSchema: ProjectAuthSchema!
}
```

---

### 1.3 Inputs

#### SetModelRLSPolicyInput (新增)
在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# 设置 Model RLS 策略输入
input SetModelRLSPolicyInput {
  """
  模型 ID
  """
  modelId: ID!

  """
  SELECT USING 谓词（JSON 字符串）
  """
  selectPredicate: String!

  """
  INSERT WITH CHECK 谓词（JSON 字符串）
  """
  insertCheck: String!

  """
  UPDATE USING 谓词（JSON 字符串）
  """
  updatePredicate: String!

  """
  UPDATE WITH CHECK 谓词（JSON 字符串）
  """
  updateCheck: String!

  """
  DELETE USING 谓词（JSON 字符串）
  """
  deletePredicate: String!
}
```

**说明**: 该 Input 支持完整的五件套 JSON 表达式设置，不限于 Preset。前端选择 Preset 后，转换为对应的 JSON 值调用。

#### ValidateRLSExprInput (新增)
在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# 校验 RLS 表达式输入
input ValidateRLSExprInput {
  """
  所属模型 ID（用于字段名白名单校验）
  """
  modelId: ID!

  """
  要校验的谓词类型
  """
  exprType: RLSExprType!

  """
  表达式 JSON 字符串
  """
  expression: String!
}

# RLS 表达式类型
enum RLSExprType {
  SELECT_PREDICATE
  INSERT_CHECK
  UPDATE_PREDICATE
  UPDATE_CHECK
  DELETE_PREDICATE
}
```

**校验规则**:
- JSON Schema 结构合法性
- 字段名白名单（对照 Model 字段列表）
- `_auth` 变量白名单（`uid` 内置 + `auth_schema` 声明）
- `insertCheck` / `updateCheck` 不含 `_exists` / `_ref`

#### AuthVariableInput (新增)
在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# Auth 变量输入
input AuthVariableInput {
  """
  变量名
  """
  name: String!

  """
  JWT 来源路径
  """
  source: String!

  """
  变量类型
  """
  type: AuthVariableType!
}
```

#### SetProjectAuthSchemaInput (新增)
在 `api/graph/org/schema/project.graphql` 中定义：

```graphql
# 设置 Project 认证变量配置输入
input SetProjectAuthSchemaInput {
  """
  项目 slug
  """
  projectSlug: String!

  """
  认证变量列表（uid 内置，无需声明）
  """
  variables: [AuthVariableInput!]!
}
```

---

### 1.4 Error Types

在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# ============================================
# RLS Error Types
# ============================================

# Model 无 owner 字段错误
type ModelHasNoOwnerField implements Error {
  message: String!
  suggestion: String
}

# EndUserRef 字段已存在错误
type EndUserRefAlreadyExists implements Error {
  message: String!
  suggestion: String
}

# RLS 表达式无效错误
type InvalidRLSExpression implements Error {
  message: String!
  suggestion: String
  path: String  # 错误位置，如 "selectPredicate._and[0].owner"
}

# 无效的 Auth 变量引用错误
type InvalidAuthVariable implements Error {
  message: String!
  suggestion: String
  variable: String  # 被引用的变量名
}

# RLS CHECK 约束违反（Runtime 返回）
type RLSCheckViolation implements Error {
  message: String!
  operation: String  # "INSERT" | "UPDATE"
}

# Policy 配置被禁止错误（高危操作）
type DangerousPolicyNotConfirmed implements Error {
  message: String!
  suggestion: String
}

# Error Unions
union SetModelRLSPolicyError = ModelNotFound | ModelHasNoOwnerField | InvalidRLSExpression | InvalidAuthVariable | ProjectNotFound
union ValidateRLSExprError = ModelNotFound | InvalidRLSExpression | InvalidAuthVariable | ProjectNotFound
union SetProjectAuthSchemaError = ProjectNotFound | InvalidInput
```

---

### 1.5 Payload Types

在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# ============================================
# RLS Payload Types
# ============================================

type SetModelRLSPolicyPayload {
  policy: ModelRLSPolicy
  error: SetModelRLSPolicyError
}

type ValidateRLSExprPayload {
  result: ValidationResult!
  error: ValidateRLSExprError
}
```

在 `api/graph/org/schema/project.graphql` 中定义：

```graphql
# ============================================
# Project AuthSchema Payload Types
# ============================================

type SetProjectAuthSchemaPayload {
  authSchema: ProjectAuthSchema
  error: SetProjectAuthSchemaError
}
```

---

### 1.6 Queries

在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# ============================================
# RLS Queries
# ============================================

extend type Query {
  """
  获取 Model RLS 策略配置
  """
  modelRLSPolicy(modelId: ID!): ModelRLSPolicy @hasPermission(action: "model:read")
}
```

**说明**: `model(id).rlsPolicy` 已在 Model 类型扩展中定义，Query 提供独立的获取方式。

---

### 1.7 Mutations

在 `api/graph/project/schema/rls.graphql` 中定义：

```graphql
# ============================================
# RLS Mutations
# ============================================

extend type Mutation {
  """
  设置 Model RLS 策略
  支持完整的五件套 JSON 表达式，不限于 Preset
  """
  setModelRLSPolicy(input: SetModelRLSPolicyInput!): SetModelRLSPolicyPayload! @hasPermission(action: "model:update")

  """
  校验 RLS 表达式合法性
  用于 Policy 配置页面的实时校验
  """
  validateRLSExpr(input: ValidateRLSExprInput!): ValidateRLSExprPayload! @hasPermission(action: "model:read")
}
```

在 `api/graph/org/schema/project.graphql` 中定义：

```graphql
# ============================================
# Project AuthSchema Mutations
# ============================================

extend type Mutation {
  """
  设置 Project 认证变量配置
  用于声明扩展 JWT 变量（如 tenant_id、role）
  """
  setProjectAuthSchema(input: SetProjectAuthSchemaInput!): SetProjectAuthSchemaPayload! @hasPermission(action: "project:update")
}
```

---

## 2. Runtime Authentication

### 2.1 JWT Issuer 区分

| JWT 类型 | `iss` 值 | 可访问端点 |
|---------|---------|----------|
| Developer JWT | `mc-developer` | 设计态 GraphQL (`/graphql/org/{orgName}/...`) |
| EndUser JWT | `mc-enduser` | Runtime GraphQL (`/graphql/org/{orgName}/project/{projectSlug}/runtime`) |

### 2.2 Runtime Endpoint 认证规则

```
Runtime Endpoint (Project Domain - runtime resolver):
├── 收到请求
├── 提取 JWT
├── 验证 iss = "mc-enduser"
│   ├── iss = "mc-developer" → 401 Unauthorized
│   ├── iss != "mc-enduser" → 401 Unauthorized
│   └── 无 JWT 或格式非法 → 401 Unauthorized
├── 提取 endUserId = jwt.user_id
├── 注入 request context
└── 后续 RLS 注入从 context 读取 endUserId
```

### 2.3 Context 传递规范

```go
// Context Key 定义
const EndUserContextKey = "end_user_identity"

// EndUserIdentity 结构
type EndUserIdentity struct {
    EndUserID string
    Issuer    string  // "mc-enduser"
}

// Middleware 注入
func (m *RuntimeAuthMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractJWT(r)
        identity, err := validateEndUserJWT(token)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), EndUserContextKey, identity)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Resolver 层读取
func (r *runtimeResolver) getEndUserID(ctx context.Context) (string, error) {
    identity, ok := ctx.Value(EndUserContextKey).(*EndUserIdentity)
    if !ok || identity == nil {
        return "", fmt.Errorf("end user identity not found in context")
    }
    return identity.EndUserID, nil
}
```

### 2.4 RLS 触发条件

```
RLS 注入仅在以下条件下触发：
1. 请求携带合法 EndUser JWT（iss = mc-enduser）
2. endUserId 已成功注入 context

身份识别优先级：
- identity.isEndUser() == false → 无 RLS 注入（Developer 访问，不会到达 Runtime）
- model.getPolicy() == nil → DENY ALL（无 Policy = Default Deny）
- 否则 → 按五件套 JsonExpr 执行注入
```

---

## 3. Error Codes

### 3.1 bizerrors 错误码映射

需要在 `pkg/bizerrors/common_errors.go` 中新增以下错误定义：

```go
// 定义 RLS 领域错误
var (
    // ModelHasNoOwnerField 模型无 owner 字段
    ModelHasNoOwnerField = ErrorDefinition{
        Code:      ErrorTypeOperationFailed + ".RLS.NO_OWNER_FIELD",
        EnMessage: "Model has no owner field",
        ZhMessage: "模型缺少 owner 字段，无法配置 RLS 策略",
    }

    // EndUserRefAlreadyExists EndUserRef 字段已存在
    EndUserRefAlreadyExists = ErrorDefinition{
        Code:      ErrorTypeConflict + ".FIELD.END_USER_REF",
        EnMessage: "EndUserRef field already exists in this model",
        ZhMessage: "每个模型只允许一个归属字段",
    }

    // InvalidRLSExpression RLS 表达式无效
    InvalidRLSExpression = ErrorDefinition{
        Code:      ErrorTypeParamInvalid + ".RLS.EXPR",
        EnMessage: "Invalid RLS expression: {0}",
        ZhMessage: "RLS 表达式无效: {0}",
    }

    // InvalidAuthVariable 无效的 Auth 变量
    InvalidAuthVariable = ErrorDefinition{
        Code:      ErrorTypeParamInvalid + ".RLS.AUTH_VAR",
        EnMessage: "Invalid auth variable reference: {0}",
        ZhMessage: "无效的认证变量引用: {0}",
    }

    // RLSCheckViolation RLS CHECK 约束违反
    RLSCheckViolation = ErrorDefinition{
        Code:      ErrorTypeOperationFailed + ".RLS.CHECK_VIOLATION",
        EnMessage: "RLS check violation: {0}",
        ZhMessage: "违反 RLS 策略约束: {0}",
    }

    // PermissionDeniedRLS RLS 权限拒绝（无 Policy 时的 DENY ALL）
    PermissionDeniedRLS = ErrorDefinition{
        Code:      ErrorTypeOperationFailed + ".RLS.PERMISSION_DENIED",
        EnMessage: "Permission denied: No RLS policy configured for this model",
        ZhMessage: "权限拒绝：该模型未配置 RLS 策略",
    }
)
```

### 3.2 GraphQL Error 与 HTTP 状态码映射

| 场景 | GraphQL Error Type | HTTP Status | Error Code |
|-----|-------------------|-------------|------------|
| Runtime 收到 Developer JWT | AuthenticationFailed | 401 | AUTHENTICATION_FAILED |
| Runtime 收到无效 JWT | AuthUnauthorized | 401 | UNAUTHORIZED |
| 模型无 owner 字段但尝试配置 Policy | ModelHasNoOwnerField | 200 | OPERATION_FAILED.RLS.NO_OWNER_FIELD |
| RLS 表达式语法错误 | InvalidRLSExpression | 200 | PARAM_INVALID.RLS.EXPR |
| 引用未声明的 _auth 变量 | InvalidAuthVariable | 200 | PARAM_INVALID.RLS.AUTH_VAR |
| insertCheck / updateCheck 含 _exists/_ref | InvalidRLSExpression | 200 | PARAM_INVALID.RLS.EXPR |
| CHECK 约束违反 | RLSCheckViolation | 200 | OPERATION_FAILED.RLS.CHECK_VIOLATION |
| 无 Policy 时 EndUser 访问 | PermissionDeniedRLS | 200 | OPERATION_FAILED.RLS.PERMISSION_DENIED |

**说明**: Runtime RLS 相关的错误（CHECK 失败、DENY ALL）在 GraphQL 层返回，HTTP 状态码保持 200，符合 GraphQL 规范。

---

## 4. 文件组织

### 4.1 新增文件

```
api/graph/project/schema/
└── rls.graphql          # 新增：RLS 相关 Schema

api/graph/org/schema/
└── project.graphql      # 扩展：添加 ProjectAuthSchema 相关定义
```

### 4.2 修改文件

```
api/graph/project/schema/
├── field.graphql        # 扩展：FormatType 枚举添加 END_USER_REF
├── model.graphql        # 扩展：Model 类型添加 rlsPolicy 字段
└── base.graphql         # 无需修改（已有 Error interface）

api/graph/org/schema/
└── project.graphql      # 扩展：Project 类型添加 authSchema 字段
                         #        添加 setProjectAuthSchema mutation

pkg/bizerrors/
└── common_errors.go     # 扩展：添加 RLS 相关错误定义
```

---

## 5. JSON 表达式格式规范

### 5.1 比较操作符

```json
{ "field": { "_eq": value } }
{ "field": { "_neq": value } }
{ "field": { "_gt": value } }
{ "field": { "_gte": value } }
{ "field": { "_lt": value } }
{ "field": { "_lte": value } }
{ "field": { "_in": [v1, v2] } }
{ "field": { "_nin": [v1, v2] } }
{ "field": { "_is_null": true } }
```

### 5.2 逻辑操作符

```json
{ "_and": [...conditions] }
{ "_or": [...conditions] }
{ "_not": condition }
```

### 5.3 特殊值

```json
{ "_auth": "uid" }              // 当前 EndUser ID（内置）
{ "_auth": "tenant_id" }        // auth_schema 声明的扩展变量
{ "_ref": "db.table.column" }   // 跨表字段引用（仅 PREDICATE 允许）
```

### 5.4 EXISTS 子查询（仅 PREDICATE 允许）

```json
{
  "_exists": {
    "model": "mc_meta.org_memberships",
    "where": {
      "org_id": { "_eq": { "_ref": "user_db.projects.org_id" } },
      "user_id": { "_eq": { "_auth": "uid" } }
    }
  }
}
```

### 5.5 常量简写

```json
true   // 等价于 {"_const": true}
false  // 等价于 {"_const": false}
```

---

## 6. 操作符允许矩阵

| 操作 | 比较操作符 | _and/_or/_not | _exists | _ref |
|------|-----------|---------------|---------|------|
| selectPredicate | ✅ | ✅ | ✅ | ✅ |
| insertCheck | ✅ | ✅ | ❌ | ❌ |
| updatePredicate | ✅ | ✅ | ✅ | ✅ |
| updateCheck | ✅ | ✅ | ❌ | ❌ |
| deletePredicate | ✅ | ✅ | ✅ | ✅ |

---

## 7. USING vs WITH CHECK 语义

| 语义类型 | 谓词 | 不通过行为 |
|---------|------|----------|
| **USING** | selectPredicate / updatePredicate / deletePredicate | 行"不存在"，静默过滤（0行/0受影响），**不报错** |
| **WITH CHECK** | insertCheck / updateCheck | 整个操作失败，抛 `RLS_CHECK_VIOLATION` |

---

## 8. 预设策略映射

| Preset | selectPredicate | insertCheck | updatePredicate | updateCheck | deletePredicate |
|--------|----------------|-------------|-----------------|-------------|-----------------|
| `READ_WRITE_OWNER` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` |
| `READ_ALL_WRITE_OWNER` | `true` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` |
| `READ_ALL` | `true` | `false` | `false` | `false` | `false` |
| `READ_WRITE_ALL` | `true` | `true` | `true` | `true` | `true` |
| `NO_ACCESS` | `false` | `false` | `false` | `false` | `false` |

---

## 附录：GraphQL 完整 Schema 片段

### rls.graphql (新增文件)

```graphql
# ============================================
# RLS Schema for ModelCraft
# File: api/graph/project/schema/rls.graphql
# ============================================

# ----------------------------------------
# Enums
# ----------------------------------------

enum RLSPreset {
  READ_WRITE_OWNER
  READ_ALL_WRITE_OWNER
  READ_ALL
  READ_WRITE_ALL
  NO_ACCESS
}

enum AuthVariableType {
  UUID
  STRING
  INTEGER
}

enum RLSExprType {
  SELECT_PREDICATE
  INSERT_CHECK
  UPDATE_PREDICATE
  UPDATE_CHECK
  DELETE_PREDICATE
}

# ----------------------------------------
# Types
# ----------------------------------------

type ModelRLSPolicy {
  modelId: ID!
  selectPredicate: String!
  insertCheck: String!
  updatePredicate: String!
  updateCheck: String!
  deletePredicate: String!
  preset: RLSPreset
  createdAt: String!
  updatedAt: String!
}

type AuthVariable {
  name: String!
  source: String!
  type: AuthVariableType!
}

type ValidationResult {
  valid: Boolean!
  errors: [ValidationError!]
}

type ValidationError {
  path: String!
  message: String!
  code: String!
}

# ----------------------------------------
# Error Types
# ----------------------------------------

type ModelHasNoOwnerField implements Error {
  message: String!
  suggestion: String
}

type EndUserRefAlreadyExists implements Error {
  message: String!
  suggestion: String
}

type InvalidRLSExpression implements Error {
  message: String!
  suggestion: String
  path: String
}

type InvalidAuthVariable implements Error {
  message: String!
  suggestion: String
  variable: String
}

type RLSCheckViolation implements Error {
  message: String!
  operation: String
}

union SetModelRLSPolicyError = ModelNotFound | ModelHasNoOwnerField | InvalidRLSExpression | InvalidAuthVariable | ProjectNotFound
union ValidateRLSExprError = ModelNotFound | InvalidRLSExpression | InvalidAuthVariable | ProjectNotFound

# ----------------------------------------
# Payload Types
# ----------------------------------------

type SetModelRLSPolicyPayload {
  policy: ModelRLSPolicy
  error: SetModelRLSPolicyError
}

type ValidateRLSExprPayload {
  result: ValidationResult!
  error: ValidateRLSExprError
}

# ----------------------------------------
# Input Types
# ----------------------------------------

input SetModelRLSPolicyInput {
  modelId: ID!
  selectPredicate: String!
  insertCheck: String!
  updatePredicate: String!
  updateCheck: String!
  deletePredicate: String!
}

input ValidateRLSExprInput {
  modelId: ID!
  exprType: RLSExprType!
  expression: String!
}

input AuthVariableInput {
  name: String!
  source: String!
  type: AuthVariableType!
}

# ----------------------------------------
# Queries & Mutations
# ----------------------------------------

extend type Query {
  modelRLSPolicy(modelId: ID!): ModelRLSPolicy @hasPermission(action: "model:read")
}

extend type Mutation {
  setModelRLSPolicy(input: SetModelRLSPolicyInput!): SetModelRLSPolicyPayload! @hasPermission(action: "model:update")
  validateRLSExpr(input: ValidateRLSExprInput!): ValidateRLSExprPayload! @hasPermission(action: "model:read")
}
```

### project.graphql (Org Domain 扩展)

```graphql
# 在 api/graph/org/schema/project.graphql 中添加：

# ----------------------------------------
# Types (add to existing file)
# ----------------------------------------

type ProjectAuthSchema {
  variables: [AuthVariable!]!
}

# 扩展 Project 类型
extend type Project {
  authSchema: ProjectAuthSchema!
}

# ----------------------------------------
# Error Types
# ----------------------------------------

union SetProjectAuthSchemaError = ProjectNotFound | InvalidInput

# ----------------------------------------
# Payload Types
# ----------------------------------------

type SetProjectAuthSchemaPayload {
  authSchema: ProjectAuthSchema
  error: SetProjectAuthSchemaError
}

# ----------------------------------------
# Input Types
# ----------------------------------------

input SetProjectAuthSchemaInput {
  projectSlug: String!
  variables: [AuthVariableInput!]!
}

# ----------------------------------------
# Mutations
# ----------------------------------------

extend type Mutation {
  setProjectAuthSchema(input: SetProjectAuthSchemaInput!): SetProjectAuthSchemaPayload! @hasPermission(action: "project:update")
}
```
