# RLS 后端实现计划

## 概览

### 实现目标
在 ModelCraft Runtime 层实现**应用层行级数据隔离（RLS）**：新建 Model 自动生成 `owner`（EndUserRef）字段，Runtime 仅接受 EndUser JWT，所有查询/变更自动注入 `WHERE owner = endUserId` 条件，开发者无需手动添加任何过滤。

### 分层影响范围

| 层级 | 路径 | 影响说明 |
|------|------|---------|
| **Domain** | `internal/domain/auth/` | 新增 EndUserClaims、迁移 issuer 校验 |
| **Domain** | `internal/domain/modeldesign/` | 新增 `END_USER_REF` FormatType、`isRLSEnabled()`、`RemoveFieldWarning` |
| **Domain** | `internal/domain/modelruntime/` | 新增 `RLSFilter` VO、Schema 生成按身份差异化、resolver 注入 RLS |
| **App** | `internal/app/modeldesign/` | CreateModelSync 自动添加 owner 字段；RemoveFieldSync 返回 warning |
| **App** | `internal/app/enduser/` | Login/Register/Refresh 颁发 `iss=mc-enduser` JWT；新增 AccessToken 生成 |
| **App** | `internal/app/modelruntime/` | Execute 从 context 读取 EndUser 身份，传递给 Schema 生成和 SQL 层 |
| **Infrastructure** | `internal/infrastructure/database/ddl/` | 新增 END_USER_REF 字段的外键 DDL 生成 |
| **Infrastructure** | `internal/infrastructure/database/dml/` | 各操作 Input 支持 RLS filter 注入 |
| **Interfaces** | `internal/interfaces/http/routes.go` | Runtime 路由换用专用 RLS 中间件 |
| **Interfaces** | `internal/middleware/` | 新增 Runtime 专用 JWT 中间件（仅接受 mc-enduser） |
| **Interfaces** | `internal/interfaces/graphql/project/` | field/model GraphQL resolver 处理新类型和 warning |
| **API Schema** | `api/graph/project/schema/` | `field.graphql`、`model.graphql` 新增枚举/类型 |
| **pkg** | `pkg/ctxutils/` | 新增 EndUser 身份 context 工具函数 |

### 实现顺序（依赖关系）

```
Wave 1（并行）
├── Task 1.1：JWT issuer 迁移（Developer + EndUser 双侧）
├── Task 1.2：EndUserRef 字段 Format（domain 层 + GraphQL Schema）
├── Task 1.3：ctxutils + RLS context 工具
├── Task 1.4：Policy 领域模型（domain 层 + GraphQL Schema）+ 三层纯函数（Validator/Compiler/Executor）
├── Task 1.5：validateRLSExpr API handler（依赖 1.4）
└── Task 1.6：auth_schema Repository + App Service（依赖 1.4）

Wave 2（依赖 Wave 1）
├── Task 2.1：Runtime 专用中间件（依赖 1.1 + 1.3）
├── Task 2.2：新建 Model 自动生成 owner + 默认 Policy（依赖 1.2 + 1.4）
├── Task 2.3：RemoveField warning 返回（依赖 1.2）
├── Task 2.4：Policy Repository + App Service + Resolver（依赖 1.4）
└── Task 2.5：auth_schema Resolver（依赖 1.6）

Wave 3（依赖 Wave 1 + Wave 2，核心 RLS 执行）
├── Task 3.1：App 层 RLS Filter 注入（依赖 2.1 + 1.4 + 2.4）
├── Task 3.2：SQL WHERE 注入（使用 PolicyExecutor.ToSQL()）（依赖 3.1）
└── Task 3.3：Runtime Schema 生成（EndUser 视角）（依赖 2.1 + 1.2）
```

---

## Wave 1 — 可并行实现的任务

> Wave 1 内部三个 Task 互不依赖，可分配给三个 worker 同时开始。

---

### Task 1.1：JWT issuer 迁移

**目标**：将 Developer JWT 的 `iss` 从 `"modelcraft"` 迁移到 `"mc-developer"`，EndUser JWT 的 `iss` 设为 `"mc-enduser"`；迁移期间设计态中间件兼容两个 issuer 值。

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `internal/domain/auth/modelcraft_claims.go` | 修改 |
| `internal/domain/auth/user_claims.go` | 修改 |
| `internal/domain/auth/enduser_claims.go` | 新增 |
| `internal/middleware/chi_jwt_auth.go` | 修改 |
| `internal/app/enduser/end_user_auth_service.go` | 修改 |
| `internal/app/auth/token_service.go` | 修改（如有 issuer 硬编码） |

#### 具体变更说明

**`internal/domain/auth/modelcraft_claims.go`**
- 将 `Validate()` 中 `c.Issuer != "modelcraft"` 改为同时接受 `"modelcraft"` 和 `"mc-developer"`（兼容期）
- 在文件中定义 issuer 常量：`IssuerDeveloper = "mc-developer"`、`IssuerLegacy = "modelcraft"`

**`internal/domain/auth/user_claims.go`**
- 同 modelcraft_claims.go，将 `Validate()` 中 issuer 检查改为接受 `"modelcraft"` 和 `"mc-developer"` 两个值

**`internal/domain/auth/enduser_claims.go`（新建文件）**
- 定义 `EndUserClaims` struct，嵌入 `jwt.RegisteredClaims`
- 包含字段：`EndUserID string json:"end_user_id"`、`ProjectSlug string json:"project_slug"`
- `Issuer` 固定为 `"mc-enduser"`，定义常量 `IssuerEndUser = "mc-enduser"`
- 实现 `Validate()` 方法：校验 `EndUserID` 非空、`iss == mc-enduser`、token 未过期

**`internal/middleware/chi_jwt_auth.go`**
- 修改 `validateModelCraftJWT` 函数：解析后检查 `claims.Issuer` 是否在允许列表中（设计态：`"modelcraft"` 或 `"mc-developer"`）
- 设计态中间件**不接受** `iss = mc-enduser`，收到时返回 401

**`internal/app/enduser/end_user_auth_service.go`**
- 在 `LoginEndUser`、`RegisterEndUser`、`RefreshEndUserToken` 的返回结果中，新增颁发 short-lived **AccessToken（JWT）** 的步骤
  - AccessToken payload：`{ iss: "mc-enduser", sub: userID, end_user_id: userID, project_slug: projectSlug, exp: <15min> }`
  - 使用与系统统一的 HMAC-SHA256 密钥签名（从配置读取）
  - `RegisterResult` / `LoginResult` 新增 `AccessToken string` 字段
- `Validate()` 逻辑中校验 `iss == "mc-enduser"`

**`internal/app/auth/token_service.go`**（如有 issuer 硬编码）
- 将颁发 Developer JWT 时的 `iss` 从 `"modelcraft"` 改为 `"mc-developer"`
- 过渡期：同时兼容验证两个值

#### 验收标准

- Developer 用旧 `iss: "modelcraft"` JWT 仍可访问设计态 GraphQL（兼容期）
- Developer 用新 `iss: "mc-developer"` JWT 可访问设计态 GraphQL
- EndUser 登录后响应包含 `accessToken` 字段，JWT payload `iss = "mc-enduser"`、`end_user_id` 字段存在
- EndUser accessToken 调用设计态 GraphQL → 401
- `EndUserClaims.Validate()` 单元测试覆盖过期/issuer 错误场景

---

### Task 1.2：EndUserRef 字段 Format

**目标**：在字段系统中新增 `END_USER_REF` Format，添加约束校验、DB DDL 外键支持，以及 GraphQL Schema 变更。

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `internal/domain/modeldesign/field_definition.go` | 修改 |
| `internal/domain/modeldesign/field_service.go` | 修改 |
| `internal/infrastructure/database/ddl/mysql_converter.go` | 修改 |
| `api/graph/project/schema/field.graphql` | 修改 |
| `api/graph/project/schema/model.graphql` | 修改 |
| `internal/interfaces/graphql/project/field.resolvers.go` | 修改 |
| `internal/interfaces/graphql/project/model.resolvers.go` | 修改 |
| `internal/app/modeldesign/commands.go` | 修改 |

#### 具体变更说明

**`internal/domain/modeldesign/field_definition.go`**
- 在 `FormatType` 常量区新增 `FormatEndUserRef FormatType = "END_USER_REF"`
- 在 `fieldTypeMap` 的 `init()` 中为 `FormatEndUserRef` 注册条目：`SchemaType = SchemaTypeString`，`Title = "终端用户归属"`
- 新增方法 `IsEndUserRefField() bool`：当 `fd.Type.Format == FormatEndUserRef` 时返回 true
- 在 `validate()` 中新增不变量：`isEndUserRef() == true` 时字段名必须为 `"owner"`，否则返回 bizerror
- `IsStringifiable()` switch 中新增 `FormatEndUserRef` 返回 true

**`internal/domain/modeldesign/field_service.go`**
- 新增方法 `HasEndUserRefField(fields []*FieldDefinition) bool`：遍历字段列表，判断是否存在 `END_USER_REF` 字段
- 新增方法 `ValidateAddEndUserRefField(existFields []*FieldDefinition) error`：若已存在 `END_USER_REF` 字段，返回 `bizerrors.NewError` 携带错误码 `END_USER_REF_ALREADY_EXISTS`，提示"每个 Model 只允许一个归属字段"

**`internal/infrastructure/database/ddl/mysql_converter.go`**
- 在字段转 DDL 的映射逻辑中，为 `FormatEndUserRef` 新增 case：
  - SQL 列类型为 `VARCHAR(36)`（存储 UUID）
  - 在 `ALTER TABLE ... ADD COLUMN` 的同时，附加 `ADD CONSTRAINT fk_{tableName}_owner FOREIGN KEY (owner) REFERENCES private_{projectSlug}.users(id)`
  - 删除 `END_USER_REF` 字段时，先 `DROP FOREIGN KEY` 再 `DROP COLUMN`
- `projectSlug` 从 `ModelLocator` 或调用参数中取得，传入 DDL 生成函数

**`api/graph/project/schema/field.graphql`**
- 在 `FormatType` 枚举中新增 `END_USER_REF`
- 新增错误类型 `EndUserRefAlreadyExists implements Error { message: String!, code: String! }`
- 扩展 `AddFieldsError` union，新增 `EndUserRefAlreadyExists` 成员
- 新增枚举 `RemoveFieldWarning { RLS_WILL_DISABLE }`
- 在 `RemoveFieldPayload` 类型中新增 `warning: RemoveFieldWarning` 字段

**`api/graph/project/schema/model.graphql`**
- 在 `Model` 类型中新增 `isRLSEnabled: Boolean!` 只读字段

**`internal/interfaces/graphql/project/field.resolvers.go`**
- 在 `addFields` resolver 中，调用 `app` 层前检测 `END_USER_REF` 字段，若触发唯一性错误 `END_USER_REF_ALREADY_EXISTS`，将对应 `AddFieldItemResult` 的 `error` 字段包装为 `EndUserRefAlreadyExists` GraphQL 类型返回
- 在 `removeField` resolver 中，接收 app 层返回的 `RemoveFieldWarning`，将其映射到 GraphQL `RemoveFieldPayload.warning` 字段

**`internal/interfaces/graphql/project/model.resolvers.go`**
- 在 `Model` 类型的字段 resolver 中，新增 `isRLSEnabled` 字段解析：调用 `model.HasEndUserRefField()` 或直接判断 `fields` 是否含 `END_USER_REF`

**`internal/app/modeldesign/commands.go`**
- 在 `RemoveFieldCommand` 的处理结果结构体（或 `RemoveFieldSync` 的返回值）中，新增 `Warning string`（对应 `RLS_WILL_DISABLE`）

#### 验收标准

- `addFields` 传入 `format = END_USER_REF` 且 `name = "owner"` → 成功创建，DB 含外键约束
- `addFields` 传入 `format = END_USER_REF` 且 `name != "owner"` → 返回 `ParamInvalid` 错误
- 同一 Model 第二次添加 `END_USER_REF` 字段 → `AddFieldItemResult.error.__typename = EndUserRefAlreadyExists`
- `removeField` 删除 `END_USER_REF` 字段 → `RemoveFieldPayload.warning = RLS_WILL_DISABLE`，操作正常执行
- `model.isRLSEnabled` 在有 `END_USER_REF` 字段时返回 `true`，无时返回 `false`

---

### Task 1.3：ctxutils + RLS context 工具

**目标**：扩展 `pkg/ctxutils`，支持将 EndUser 身份（ID + Issuer）注入和读取 request context；新建 `pkg/ctxutils/rls.go`（或在 `userctx.go` 中扩展）。

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `pkg/ctxutils/rls.go` | 新增 |

#### 具体变更说明

**`pkg/ctxutils/rls.go`（新建文件）**
- 定义 context key 常量：`ContextKeyEndUserID`、`ContextKeyEndUserIssuer`
- 提供 `SetEndUserID(ctx, endUserID string) context.Context`：将 endUserID 写入 context
- 提供 `GetEndUserID(ctx) (string, bool)`：从 context 读取 endUserID，不存在时 bool = false
- 提供 `SetEndUserIssuer(ctx, issuer string) context.Context`
- 提供 `GetEndUserIssuer(ctx) (string, bool)`
- 提供辅助函数 `IsEndUserRequest(ctx) bool`：当 `GetEndUserIssuer` 返回 `"mc-enduser"` 时为 true
- 所有 key 使用 package-private 的 `contextKey` 类型（已在 `userctx.go` 中定义），避免冲突

> 注意：**不**复用现有的 `ContextKeyUserID`（设计态 Developer 用户专用），EndUser 身份使用独立 key，防止越权访问。

#### 验收标准

- `SetEndUserID` + `GetEndUserID` 往返测试通过
- `IsEndUserRequest` 在 issuer = `"mc-enduser"` 时返回 true，其他情况返回 false
- 与现有 `SetUserID`/`GetUserIDFromContext` 不冲突（key 值不同）

---

### Task 1.4：Policy 领域模型

**目标**：在领域层新增 `ModelRLSPolicy` 实体（五件套字段类型改为 JSON String）、相关枚举和 Repository 接口；在 GraphQL Schema 中新增 Policy 相关类型和 `setModelRLSPolicy` mutation；新增 `PolicyValidator`、`PolicyCompiler`、`PolicyExecutor` 三层纯函数。

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `internal/domain/rls/policy.go` | 新增 |
| `internal/domain/rls/validator.go` | 新增 |
| `internal/domain/rls/compiler.go` | 新增 |
| `internal/domain/rls/executor.go` | 新增 |
| `api/graph/project/schema/model.graphql` | 修改（新增 Policy 类型和 mutation） |

#### 具体变更说明

**`internal/domain/rls/policy.go`（新建文件）**

```go
// JsonExpr 是 RLS 策略中的 Boolean 条件表达式（JSON 字符串）
type JsonExpr string

// ModelRLSPolicy 描述 Model 的 RLS 执行策略（五件套）
type ModelRLSPolicy struct {
    ModelID         string
    SelectPredicate JsonExpr  // SELECT USING（JSON）
    InsertCheck     JsonExpr  // INSERT WITH CHECK（JSON）
    UpdatePredicate JsonExpr  // UPDATE USING（JSON）
    UpdateCheck     JsonExpr  // UPDATE WITH CHECK（JSON）
    DeletePredicate JsonExpr  // DELETE USING（JSON）
}

// RLSPreset 是五件套 JsonExpr 的命名组合
type RLSPreset string

const (
    RLSPresetReadWriteOwner    RLSPreset = "READ_WRITE_OWNER"
    RLSPresetReadAllWriteOwner RLSPreset = "READ_ALL_WRITE_OWNER"
    RLSPresetReadAll           RLSPreset = "READ_ALL"
    RLSPresetReadWriteAll      RLSPreset = "READ_WRITE_ALL"
    RLSPresetNoAccess          RLSPreset = "NO_ACCESS"
)

// OwnerEqualsUserExpr 是 owner = auth.uid 的标准 JSON 表达式
const OwnerEqualsUserExpr JsonExpr = `{"owner":{"_eq":{"_auth":"uid"}}}`

// DefaultPolicy 返回新建 Model 时的默认 Policy（READ_WRITE_OWNER）
func DefaultPolicy(modelID string) *ModelRLSPolicy {
    return &ModelRLSPolicy{
        ModelID:         modelID,
        SelectPredicate: OwnerEqualsUserExpr,
        InsertCheck:     OwnerEqualsUserExpr,
        UpdatePredicate: OwnerEqualsUserExpr,
        UpdateCheck:     OwnerEqualsUserExpr,
        DeletePredicate: OwnerEqualsUserExpr,
    }
}

// GetPreset 返回当前五件套对应的 Preset；自定义组合返回 nil
func (p *ModelRLSPolicy) GetPreset() *RLSPreset { ... }

// ModelRLSPolicyRepository 接口
type ModelRLSPolicyRepository interface {
    FindByModelID(ctx context.Context, modelID string) (*ModelRLSPolicy, error)
    Save(ctx context.Context, policy *ModelRLSPolicy) error
    DeleteByModelID(ctx context.Context, modelID string) error
}
```

**`internal/domain/rls/validator.go`（新建文件）**
- 定义 `PolicyValidator` struct
- 实现 `Validate(json string, modelSchema ModelSchema, authSchema AuthSchema) []ValidationError`：
  - JSON Schema 校验结构合法性
  - 字段名白名单校验（对照 Model 字段列表）
  - `_auth` 变量白名单校验（`uid` 内置 + `auth_schema` 声明）
  - `_exists.model` 白名单校验
  - `insertCheck` / `updateCheck` 不含 `_exists` / `_ref`

**`internal/domain/rls/compiler.go`（新建文件）**
- 定义 `PolicyCompiler` struct 和 `CompiledPolicy` 类型
- 实现 `Compile(json string) (CompiledPolicy, error)`：递归解析 JSON 为参数化 SQL 片段

**`internal/domain/rls/executor.go`（新建文件）**
- 定义 `PolicyExecutor` struct 和 `AuthContext` 类型
- 实现 `ToSQL(compiled CompiledPolicy, authCtx AuthContext) (string, []any)`：绑定运行期变量（`uid` 等）生成最终参数化 SQL WHERE + args[]

**`api/graph/project/schema/model.graphql`**
- 按 api-contract.md §1.5 新增全部 Policy 相关类型、枚举、mutation（五件套字段类型为 `String!`）
- **删除** `ExprType` 枚举

#### 验收标准

- `DefaultPolicy()` 返回五件套均为 `{"owner":{"_eq":{"_auth":"uid"}}}`
- `GetPreset()` 正确识别 5 种 Preset 组合；非标准组合返回 nil
- `PolicyValidator` 对非法 JSON、未知字段、未声明 `_auth` 变量正确返回 `ValidationError`
- `PolicyCompiler.Compile()` + `PolicyExecutor.ToSQL()` 将 `{"owner":{"_eq":{"_auth":"uid"}}}` 编译为 `"owner = ?"` + args=[endUserID]
- GraphQL Schema 通过 `just generate-gql` 生成无报错

---

> 等待 Wave 1 全部完成后并行开始。Task 2.1、2.2、2.3 互不依赖。

---

### Task 1.5（新增）：validateRLSExpr API handler

**目标**：实现 `validateRLSExpr` mutation 的 resolver，供前端构建器实时校验 JSON 表达式合法性（编译期，不修改 Policy）。

**依赖**：Task 1.4（`PolicyValidator`、`PolicyCompiler`）

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `internal/interfaces/graphql/project/rls.resolvers.go` | 修改（新增 validateRLSExpr resolver） |
| `api/graph/project/schema/rls.graphql` | 新增（ValidateRLSExprInput/Payload/RLSOperation 类型） |

#### 具体变更说明

- `validateRLSExpr` resolver 流程：
  1. 查询 Model 字段列表（白名单来源）
  2. 查询 Project `auth_schema`
  3. 调用 `PolicyValidator.Validate(expression, modelSchema, authSchema)`
  4. 若 `operation` 为 CHECK 类（`INSERT_CHECK`/`UPDATE_CHECK`），额外校验不含 `_exists`/`_ref`
  5. 返回 `ValidateRLSExprPayload{ valid, errors }`

#### 验收标准

- 合法表达式 → `valid=true, errors=[]`
- 含未知字段的表达式 → `valid=false, errors=["未知字段: xxx"]`
- `insertCheck` 含 `_exists` → `valid=false, errors=["CHECK 谓词不允许 _exists"]`
- 含未声明 `_auth` 变量 → `valid=false, errors=["未声明的 auth 变量: tenant_id"]`

---

### Task 1.6（新增）：auth_schema Repository + App Service

**目标**：实现 Project 级别 `auth_schema` 的存储、查询和更新。

**依赖**：Task 1.4（`PolicyValidator` 依赖 `AuthSchema`）

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `db/schema/mysql/project_auth_schemas.sql` | 新增 |
| `internal/domain/rls/auth_schema.go` | 新增 |
| `internal/infrastructure/database/rls/auth_schema_repo.go` | 新增 |
| `internal/app/rls/auth_schema_app.go` | 新增 |
| `internal/interfaces/graphql/project/rls.resolvers.go` | 修改（新增 setProjectAuthSchema/getProjectAuthSchema resolver） |

#### 具体变更说明

**DB 表**：
```sql
CREATE TABLE IF NOT EXISTS project_auth_schemas (
    project_id VARCHAR(36) NOT NULL PRIMARY KEY,
    schema_json TEXT NOT NULL DEFAULT '{}' COMMENT 'JSON: {varName: {source, type}}',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
);
```

**domain**：定义 `AuthSchema`、`AuthVariable` 值对象；`AuthSchemaRepository` 接口
**app**：`SetProjectAuthSchema`、`GetProjectAuthSchema` 方法
**resolver**：实现 `setProjectAuthSchema` mutation 和 `getProjectAuthSchema` query

#### 验收标准

- 调用 `setProjectAuthSchema` 保存变量列表 → DB 更新，`getProjectAuthSchema` 查询返回新值
- `uid` 不出现在 `variables` 列表中（内置）
- `validateRLSExpr` 使用最新 `auth_schema` 进行 `_auth` 变量校验

---

### Task 2.1：Runtime 专用中间件

**目标**：为 Runtime endpoint 创建专用的 JWT 认证中间件，只接受 `iss = mc-enduser` 的 JWT，提取 `end_user_id` 注入 context；替换 `routes.go` 中 Runtime 路由使用的中间件。

**依赖**：Task 1.1（`EndUserClaims`）、Task 1.3（`SetEndUserID`/`SetEndUserIssuer`）

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `internal/middleware/chi_runtime_jwt_auth.go` | 新增 |
| `internal/interfaces/http/routes.go` | 修改 |

#### 具体变更说明

**`internal/middleware/chi_runtime_jwt_auth.go`（新建文件）**
- 定义 `ChiRuntimeJWTAuthMiddleware(secret []byte, skipValidation bool) func(http.Handler) http.Handler`
- 认证流程：
  1. 从 `Authorization: Bearer <token>` 提取 token，缺失时返回 401（JSON 格式：`errors[].extensions.code = RUNTIME_UNAUTHORIZED`）
  2. 解析 JWT，使用 HMAC-SHA256 + secret 验证签名
  3. 解析为 `auth.EndUserClaims`，调用 `claims.Validate()`
  4. 若 `claims.Issuer != "mc-enduser"` → 返回 401，body 含 `RUNTIME_UNAUTHORIZED` 错误码
  5. 校验通过后调用 `ctxutils.SetEndUserID(ctx, claims.EndUserID)` 和 `ctxutils.SetEndUserIssuer(ctx, "mc-enduser")`，将新 ctx 传给 next handler
- 错误响应格式严格遵循 api-contract.md §3.2 定义的 JSON 结构（`{"errors":[{"message":"...","extensions":{"code":"RUNTIME_UNAUTHORIZED"}}]}`）
- `skipValidation = true` 时直接放行（对应 `cfg.Auth.Runtime.Enabled = false` 的开发环境）

**`internal/interfaces/http/routes.go`**
- 在 `SetupRuntimeGraphQLRoutesOnChi` 函数中，将 `runtimeMW` 内部的 `middleware.ChiJWTAuthMiddleware(jwtConfig)` 替换为 `middleware.ChiRuntimeJWTAuthMiddleware([]byte(cfg.JWT.Secret), !cfg.Auth.Runtime.Enabled)`
- 移除 Runtime 路由中原有的 `jwtConfig`（使用通用 Developer JWT 配置的那一个）
- 保留 `requestIDInjectorMiddleware`、`ChiGraphQLOrgMiddleware`、`cacheMW` 等不变

#### 验收标准

- Developer JWT（`iss = mc-developer`）调用 Runtime `POST /graphql/...` → HTTP 401，body 含 `RUNTIME_UNAUTHORIZED`
- 无 JWT 调用 Runtime → HTTP 401
- EndUser JWT（`iss = mc-enduser`）调用 Runtime → 通过认证，context 中可读取 `endUserID`
- `skipValidation = true` 时所有请求放行（dev 环境不变）
- 设计态 GraphQL 路由不受影响（仍使用原有 `ChiJWTAuthMiddleware`）

---

### Task 2.2：新建 Model 自动生成 owner 字段 + 默认 Policy

**目标**：`CreateModelSync` 创建新建 Model 时自动注入 `owner`（`END_USER_REF`）字段，并同步创建默认 Policy（Preset = READ_WRITE_OWNER）；导入 Model（`ReverseEngineerAppService.ImportModel`）两者均不创建。

**依赖**：Task 1.2（`FormatEndUserRef`）、Task 1.4（`DefaultPolicy`、`ModelRLSPolicyRepository`）

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `internal/domain/modeldesign/field_service.go` | 修改 |
| `internal/app/modeldesign/model_app.go` | 修改 |

#### 具体变更说明

**`internal/domain/modeldesign/field_service.go`**
- 新增函数 `NewOwnerFieldDefinition(modelID string, modelLocator *ModelLocator) *FieldDefinition`：
  - 构建 `name = "owner"`，`format = END_USER_REF`，`nonNull = true`，`required = false`，`title = "归属用户"` 的 `FieldDefinition`
  - `displayOrder` 设置为一个排在所有普通字段之后的固定值（如 `"z"`）
  - `status = FieldStatusInit`

**`internal/app/modeldesign/model_app.go`**
- 在 `CreateModelSync` 函数中，在调用 `modeldesign.GetSystemFields()` 之后（注入系统字段 id/createdAt 等之后），再追加调用 `NewOwnerFieldDefinition` 的结果，通过 `model.AddFields` 加入
- **新增**：owner 字段创建成功后，调用 `ModelRLSPolicyRepository.Save(ctx, rls.DefaultPolicy(model.ID))` 创建默认 Policy
- 追加逻辑仅在 `CreateModelSync` 路径生效，**不**修改 `ReverseEngineerAppService.createModel` 路径——确保导入 Model 不受影响
- 如 `CreateModelFromSchema` 也属于"新建"语义，同样追加 owner 字段 + 创建默认 Policy（根据实际业务确认）

#### 验收标准

- 调用 `createModel` mutation（非导入）→ 返回的 `model.fields` 中包含 `{ name: "owner", format: "END_USER_REF" }`
- 调用 `createModel` mutation → `model.rlsPolicy.readScope = OWNER`，`model.rlsPolicy.writeScope = OWNER`
- `model.isRLSEnabled` 返回 `true`
- 调用 `importModel`（reverse engineer）→ `model.fields` 中无 `owner` 字段，`isRLSEnabled = false`，`rlsPolicy = null`
- DB 中对应表存在 `owner VARCHAR(36)` 列及外键约束
- DB 中 `model_rls_policies` 表存在对应记录

---

### Task 2.3：RemoveField warning 返回

**目标**：`RemoveFieldSync` 删除 `END_USER_REF` 字段时，在操作执行后于返回结果中携带 `RLS_WILL_DISABLE` warning；GraphQL resolver 将其暴露给前端。

**依赖**：Task 1.2（`FormatEndUserRef` 判断）

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `internal/app/modeldesign/commands.go` | 修改 |
| `internal/app/modeldesign/model_app.go` | 修改 |
| `internal/interfaces/graphql/project/field.resolvers.go` | 修改 |

#### 具体变更说明

**`internal/app/modeldesign/commands.go`**
- 新增类型 `RemoveFieldResult struct { Warning string }`，`Warning` 取值为空字符串（无警告）或 `"RLS_WILL_DISABLE"`
- `RemoveFieldSync` 函数签名变更：返回值从 `error` 改为 `(*RemoveFieldResult, error)`

**`internal/app/modeldesign/model_app.go`**
- 在 `RemoveFieldSync` 中，删除字段成功后，判断被删除字段的 `format == FormatEndUserRef`：
  - 若是 → 调用 `ModelRLSPolicyRepository.DeleteByModelID(ctx, modelID)` 级联删除 Policy；返回 `&RemoveFieldResult{Warning: "RLS_WILL_DISABLE"}, nil`
  - 否则 → 返回 `&RemoveFieldResult{}, nil`
- 所有调用 `RemoveFieldSync` 的地方（包含内部批量调用）更新接收新签名

**`internal/interfaces/graphql/project/field.resolvers.go`**
- `removeField` resolver 接收 `RemoveFieldResult`，将 `result.Warning` 映射到 `RemoveFieldPayload.warning` GraphQL 枚举值
- `warning = "RLS_WILL_DISABLE"` → GraphQL 响应中 `warning: RLS_WILL_DISABLE`；空字符串 → `warning: null`

#### 验收标准

- 调用 `removeField` 删除普通字段 → `RemoveFieldPayload.warning = null`
- 调用 `removeField` 删除 `END_USER_REF` 字段 → 字段正常删除 + `RemoveFieldPayload.warning = RLS_WILL_DISABLE`
- 删除后 `model.isRLSEnabled = false`，`model.rlsPolicy = null`
- DB 中 `model_rls_policies` 表对应记录被删除

---

### Task 2.4：Policy Repository + App Service + Resolver

**目标**：实现 Policy 的持久化层（DB 表 + Repository）、App Service（`setModelRLSPolicy`）、GraphQL Resolver；支持开发者通过 mutation 修改 Policy。

**依赖**：Task 1.4（`ModelRLSPolicyRepository` 接口定义）

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `db/schema/mysql/model_rls_policies.sql` | 新增 |
| `internal/infrastructure/database/rls/policy_repo.go` | 新增 |
| `internal/app/rls/policy_app.go` | 新增 |
| `internal/interfaces/graphql/project/rls.resolvers.go` | 新增 |
| `internal/interfaces/graphql/project/model.resolvers.go` | 修改（新增 `rlsPolicy` 字段 resolver） |

#### 具体变更说明

**`db/schema/mysql/model_rls_policies.sql`（新建文件）**

```sql
CREATE TABLE IF NOT EXISTS model_rls_policies (
    model_id         VARCHAR(36)  NOT NULL PRIMARY KEY,
    select_predicate TEXT         NOT NULL COMMENT 'SELECT USING：JSON Boolean 表达式',
    insert_check     TEXT         NOT NULL COMMENT 'INSERT WITH CHECK：JSON Boolean 表达式',
    update_predicate TEXT         NOT NULL COMMENT 'UPDATE USING：JSON Boolean 表达式',
    update_check     TEXT         NOT NULL COMMENT 'UPDATE WITH CHECK：JSON Boolean 表达式',
    delete_predicate TEXT         NOT NULL COMMENT 'DELETE USING：JSON Boolean 表达式',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
);
```

> 表存放在 `modelcraft_db`（平台元数据库），而非客户 DB。  
> 每列存储 JSON 字符串（如 `{"owner":{"_eq":{"_auth":"uid"}}}`），使用 TEXT 类型以支持任意长度表达式。

**`internal/infrastructure/database/rls/policy_repo.go`（新建文件）**
- 实现 `ModelRLSPolicyRepository` 接口（`FindByModelID`、`Save`、`DeleteByModelID`）
- 使用 sqlc 或直接 sql.DB 执行；`Save` 做 upsert（INSERT ... ON DUPLICATE KEY UPDATE）

**`internal/app/rls/policy_app.go`（新建文件）**
- 定义 `RLSPolicyAppService` struct
- 实现 `SetModelRLSPolicy(ctx, modelID, readScope, writeScope) error`：
  - 检查 Model 是否存在 owner 字段（`isRLSEnabled`），若否返回 `PolicyNotFound` bizerror
  - 调用 `repo.Save(ctx, policy)` 持久化
- 实现 `GetModelPolicy(ctx, modelID) (*ModelRLSPolicy, error)`

**`internal/interfaces/graphql/project/rls.resolvers.go`（新建文件）**
- 实现 `setModelRLSPolicy` mutation resolver：调用 `RLSPolicyAppService.SetModelRLSPolicy`，返回 `SetModelRLSPolicyPayload`
- 错误映射：`PolicyNotFound` bizerror → GraphQL `PolicyNotFound` 类型

**`internal/interfaces/graphql/project/model.resolvers.go`**
- 新增 `Model.rlsPolicy` 字段 resolver：调用 `RLSPolicyAppService.GetModelPolicy(ctx, model.ID)`，`isRLSEnabled=false` 时返回 nil

#### 验收标准

- 调用 `setModelRLSPolicy`（有效 Model）→ Policy 更新，`model.rlsPolicy` 返回新值
- 调用 `setModelRLSPolicy`（无 owner 字段的 Model）→ `SetModelRLSPolicyPayload.error.__typename = PolicyNotFound`
- 修改 Policy 后，Runtime 下一次请求按新 Policy 执行注入逻辑
- `model.rlsPolicy.preset` 在标准组合时返回对应枚举值，自定义组合时返回 null

> 依赖 Wave 1 全部完成 + Wave 2 全部完成后开始。Task 3.1 是 3.2 和 3.3 的前置依赖，3.2 和 3.3 可并行。

---

### Task 3.1：App 层 RLS Filter 注入

**目标**：在 `GraphqlAppService.Execute` 中，读取 context 的 EndUser 身份 + 查询 Model 的 Policy（readScope/writeScope），构建 `RLSFilter`，注入 `graphqlRequestContext`，供 resolver 使用。

**依赖**：Task 1.3（`GetEndUserID`/`IsEndUserRequest`）、Task 1.4（`ModelRLSPolicy`、`RLSReadScope`、`RLSWriteScope`）、Task 2.1（中间件已将 endUserID 注入 context）、Task 2.4（`RLSPolicyAppService.GetModelPolicy`）

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `internal/domain/modelruntime/rls.go` | 新增 |
| `internal/domain/modelruntime/graphql_request_context.go` | 修改 |
| `internal/app/modelruntime/graphql_app.go` | 修改 |

#### 具体变更说明

**`internal/domain/modelruntime/rls.go`（新建文件）**
- 定义 `RLSFilter` 值对象：`SelectPredicate JsonExpr`、`InsertCheck JsonExpr`、`UpdatePredicate JsonExpr`、`UpdateCheck JsonExpr`、`DeletePredicate JsonExpr`、`FieldName string`（固定 `"owner"`）、`EndUserID string`
- 定义 `ResolveRLSFilter(ctx context.Context, policy *rls.ModelRLSPolicy) *RLSFilter`：
  - `ctxutils.IsEndUserRequest(ctx) == false` → 返回 nil（不过滤，Developer 开放访问）
  - `policy == nil` → **DENY ALL**（无 Policy = Default Deny，不是开放模式）
  - 否则取 `ctxutils.GetEndUserID(ctx)` 构建并返回五件套 `&RLSFilter{...}`

**`internal/domain/modelruntime/graphql_request_context.go`**
- 在 `graphqlRequestContext` struct 中新增字段 `RLSFilter *RLSFilter`
- 修改 `newGraphqlRequestContext` 和 `WithGraphqlRequestContext` 函数签名，增加 `rlsFilter *RLSFilter` 参数
- 新增辅助方法 `HasRLS() bool` 和 `GetRLSFilter() *RLSFilter`

**`internal/app/modelruntime/graphql_app.go`**
- 在 `Execute` 函数中，`WithGraphqlRequestContext` 调用之前：
  1. 调用 `RLSPolicyAppService.GetModelPolicy(ctx, modelID)` 得到 `policy`
  2. 调用 `modelruntime.ResolveRLSFilter(ctx, policy)` 得到 `rlsFilter`
  3. 将 `rlsFilter` 传入 `modelruntime.WithGraphqlRequestContext(ctx, clientRepo, rlsFilter)`

#### 验收标准

- EndUser 请求 + Policy{READ_WRITE_OWNER} → `graphqlRequestContext.RLSFilter` 不为 nil，五件套均为 `OWNER_EQUALS_USER`
- EndUser 请求 + Policy{READ_ALL} → `RLSFilter` 不为 nil，`SelectPredicate=ALWAYS_TRUE`，其余=`ALWAYS_FALSE`
- EndUser 请求 + 无 Policy（model.rlsPolicy=nil）→ `RLSFilter` 触发 DENY ALL（不是 nil 开放）
- Developer 请求 → `RLSFilter == nil`（不受 RLS 约束）
- 单元测试：`ResolveRLSFilter` 覆盖三个分支

---

### Task 3.2：SQL WHERE 注入

**目标**：在各 DML 操作的 `Input` 结构和 `resolver` 中，当 `RLSFilter != nil` 时注入 `AND owner = endUserId` 条件；CreateOne 强制覆盖 `owner` 字段值。

**依赖**：Task 3.1（`RLSFilter` 已挂载到 `graphqlRequestContext`）

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `internal/domain/modelruntime/model_resolver.go` | 修改 |
| `internal/domain/modelruntime/graphql_input.go` | 修改（如 Input 结构定义在此） |

#### 具体变更说明

**`internal/domain/modelruntime/model_resolver.go`**

各 execute 方法从 `getGraphqlRequestContext(p.Context)` 取到 `rctx`，再通过 `rctx.GetRLSFilter()` 获取 filter，按以下规则处理：

**读操作（FindMany / FindFirst / Count）**，根据 `filter.SelectPredicate` 分支：
- `true`（常量） → 不追加任何 WHERE 条件，直接执行
- JSON 表达式 → `PolicyCompiler.Compile(predicate)` + `PolicyExecutor.ToSQL(compiled, AuthContext{uid: endUserID})` 生成 `(whereClause string, args []any)`；将 whereClause 追加到 WHERE（与用户传入条件取 AND）；行不存在 → 静默返回空集（不报错）
- `false`（常量） → 追加 `AND 1=0`，保证返回空集（静默，不报错）

**UpdateOne USING 过滤（updatePredicate）**，根据 `filter.UpdatePredicate` 分支：
- `true`（常量） → 不追加 WHERE，正常更新
- JSON 表达式 → PolicyExecutor.ToSQL() 生成参数化 WHERE；受影响行数为 0 → **静默返回（0 受影响，不报错，不返回 RECORD_NOT_FOUND）**
- `false`（常量） → 追加 `AND 1=0`，静默返回 0 受影响（不报错）

**UpdateOne WITH CHECK 校验（updateCheck）**，在 USING 过滤之后对更新结果校验：
- `true`（常量） → 允许写入
- JSON 表达式 → PolicyExecutor 应用层评估（不含 `_exists`/`_ref`），违反 → 返回 `RLS_CHECK_VIOLATION`，回滚操作
- `false`（常量） → 直接返回 `RLS_CHECK_VIOLATION`，不执行更新

**DeleteOne USING 过滤（deletePredicate）**，根据 `filter.DeletePredicate` 分支：
- `true`（常量） → 不追加 WHERE，正常删除
- JSON 表达式 → PolicyExecutor.ToSQL() 生成参数化 WHERE；受影响行数为 0 → **静默返回（0 受影响，不报错，不返回 RECORD_NOT_FOUND）**
- `false`（常量） → 追加 `AND 1=0`，静默返回 0 受影响（不报错）

**CreateOne WITH CHECK 校验（insertCheck）**：
- `true`（常量） → 允许写入
- JSON 表达式（含 owner 等于 uid 约束）→ 在 `input.Data` 中强制设置 `input.Data["owner"] = filter.EndUserID`（覆盖用户传入的任何值）；`filter.EndUserID == ""` 时返回 500；PolicyExecutor 应用层评估校验
- `false`（常量） → 直接返回 `RLS_CHECK_VIOLATION`，不执行插入

**无 Policy 时（RLSFilter == nil 且 EndUser 请求）**：
- 所有操作 → 返回 `PERMISSION_DENIED`（Default Deny）

**`internal/domain/modelruntime/graphql_input.go`**（如 Input 构建函数在此）
- 确认 `newCreateOneInput`、`newFindManyInput` 等函数返回的 `Where` 字段类型为 `map[string]any`，可直接追加 key
- 如已是此结构则无需修改，仅在 resolver 层追加即可

#### 验收标准

- selectPredicate=OWNER_EQUALS_USER：EndUser A 调用 `findMany` → SQL WHERE 含 `AND owner = 'A_ID'`，只返回 owner = A 的行
- selectPredicate=ALWAYS_TRUE：EndUser A 调用 `findMany` → 不注入 WHERE，返回全量数据
- selectPredicate=ALWAYS_FALSE：EndUser A 调用 `findMany` → SQL WHERE 含 `AND 1=0`，静默返回空集（不报错）
- selectPredicate=OWNER_EQUALS_USER：EndUser A 传 `where: { owner: "B_ID" }` → SQL WHERE 含 `owner = 'B_ID' AND owner = 'A_ID'`，结果为空（交集语义）
- updatePredicate=OWNER_EQUALS_USER：EndUser A 调用 `updateOne` 更新 owner = B 的记录 → 静默 0 行受影响（不报错，不返回 RECORD_NOT_FOUND）
- updateCheck=ALWAYS_FALSE：EndUser A 调用 `updateOne` → 返回 `RLS_CHECK_VIOLATION`，操作失败
- insertCheck=ALWAYS_FALSE：EndUser A 调用 `createOne` → 返回 `RLS_CHECK_VIOLATION`，操作失败
- insertCheck=OWNER_EQUALS_USER：EndUser A 调用 `createOne` 不传 owner → `owner = A_ID` 被自动填充
- insertCheck=OWNER_EQUALS_USER：EndUser A 调用 `createOne` 传 `owner = B_ID` → 被强制覆盖为 `A_ID`
- 无 Policy 的 Model + EndUser 请求 → 所有操作返回 `PERMISSION_DENIED`（Default Deny）
- Developer 请求（skipValidation）→ 所有操作行为与升级前完全一致

---

### Task 3.3：Runtime Schema 生成（EndUser 视角隐藏 owner input）

**目标**：当请求来自 EndUser 时，动态生成的 Runtime GraphQL Schema 中，`CreateInput` 和 `UpdateInput` 类型不包含 `owner` 字段；Query 返回类型中 `owner` 仍可见。

**依赖**：Task 1.2（`FormatEndUserRef` 标识）、Task 1.3（`IsEndUserRequest`）、Task 2.1（中间件已注入 issuer）

#### 变更文件列表

| 文件 | 变更类型 |
|------|---------|
| `internal/domain/modelruntime/graphql_input_types.go` | 修改 |
| `internal/domain/modelruntime/model_resolver.go` | 修改 |
| `internal/app/modelruntime/graphql_app.go` | 修改 |

#### 具体变更说明

**核心设计决策**：当前 Schema 是**请求无关**的（可跨请求缓存）。EndUser 视角的差异化需要**按请求身份生成不同 Schema** 或在运行时动态跳过字段。推荐方案：**Schema 按调用者类型（enduser / developer）分为两个版本**，以 `modelLocator + callerType` 为 cache key。

**`internal/app/modelruntime/graphql_app.go`**
- `GetSchema` 函数签名扩展，增加 `callerType string`（`"enduser"` 或 `"developer"`）参数
- Cache key 变为 `modelLocator + ":" + callerType`，两类调用者分别缓存独立 Schema
- 调用 `graphqlSchemaManager.NewSchemaFrom(ctx, model, callerType)` 传入身份类型

**`internal/domain/modelruntime/graphqlschema_manager.go`**
- `NewSchemaFrom` 签名增加 `callerType string` 参数，传递给 `newGraphqlModelResolver`

**`internal/domain/modelruntime/model_resolver.go`**
- `graphqlModelResolver` struct 新增字段 `callerType string`（`"enduser"` 或 `"developer"`）
- 新增方法 `isEndUserCaller() bool`
- 在 `createRootMutation` → `createCreateOneField` → `inputTypeGenerator.GenerateCreateOneArgs` 调用链路中，传入 `isEndUser bool` 标志

**`internal/domain/modelruntime/graphql_input_types.go`**
- `GenerateCreateOneArgs(model *RuntimeModel, isEndUser bool) (graphql.FieldConfigArgument, error)`：
  - `isEndUser == true` 时，在遍历 `model.Fields` 生成 input args 时，跳过 `format == END_USER_REF` 的字段
  - `isEndUser == false`（Developer）时行为不变，完整暴露所有字段
- 同理处理 `GenerateUpdateOneArgs`：EndUser 视角跳过 `END_USER_REF` 字段

#### 验收标准

- EndUser 请求生成的 Schema：`createOrdersInput` 类型中无 `owner` 字段
- EndUser 请求生成的 Schema：`updateOrdersInput` 类型中无 `owner` 字段
- EndUser 请求生成的 Schema：`OrdersObject` 返回类型中包含 `owner: String` 字段
- Developer 请求（skipValidation 或未来支持）生成的 Schema：`createOrdersInput` 和 `updateOrdersInput` 中均有 `owner` 字段
- Schema 缓存按 `modelLocator:callerType` 分别存储，互不干扰

---

## DB Schema 变更

### 变更说明

#### 新增平台元数据表：`model_rls_policies`

Policy 存储在 `modelcraft_db`（平台元数据库），与其他 Model 元数据（`models`、`fields` 等）同库。

```sql
CREATE TABLE IF NOT EXISTS model_rls_policies (
    model_id         VARCHAR(36)  NOT NULL PRIMARY KEY COMMENT '关联 Model ID',
    select_predicate TEXT         NOT NULL COMMENT 'SELECT USING：JSON Boolean 表达式',
    insert_check     TEXT         NOT NULL COMMENT 'INSERT WITH CHECK：JSON Boolean 表达式',
    update_predicate TEXT         NOT NULL COMMENT 'UPDATE USING：JSON Boolean 表达式',
    update_check     TEXT         NOT NULL COMMENT 'UPDATE WITH CHECK：JSON Boolean 表达式',
    delete_predicate TEXT         NOT NULL COMMENT 'DELETE USING：JSON Boolean 表达式',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
);
```

- 与 `models` 表通过 `model_id` 逻辑关联（不设 FK，避免跨事务约束）
- 插入时机：`CreateModelSync` 成功后同步插入；更新时机：`setModelRLSPolicy` mutation
- 删除时机：`RemoveFieldSync` 删除 `END_USER_REF` 字段时级联删除
- 字段类型为 `TEXT`，存储 JSON 字符串（如 `{"owner":{"_eq":{"_auth":"uid"}}}`）

#### 客户 DB 的 DDL 变更（动态生成，由 DDL 层执行）

**当开发者为 Model 添加 `END_USER_REF` 字段时（`addField` → `DeployModel`）**：

```sql
-- 新增 owner 列
ALTER TABLE {model_table_name}
  ADD COLUMN owner VARCHAR(36) NOT NULL;

-- 新增外键约束（确保 owner 值指向合法的 EndUser）
ALTER TABLE {model_table_name}
  ADD CONSTRAINT fk_{model_table_name}_owner
    FOREIGN KEY (owner)
    REFERENCES private_{projectSlug}.users(id)
    ON DELETE RESTRICT
    ON UPDATE CASCADE;
```

**当开发者删除 `END_USER_REF` 字段时（`removeField` → `DeployModelToRemoveFields`）**：

```sql
-- 先删除外键约束（MySQL 要求先删约束再删列）
ALTER TABLE {model_table_name}
  DROP FOREIGN KEY fk_{model_table_name}_owner;

-- 再删除列
ALTER TABLE {model_table_name}
  DROP COLUMN owner;
```

**新建 Model 时（`CreateModelSync` 自动注入 owner）**：
- `owner VARCHAR(36) NOT NULL` 列作为 `CREATE TABLE` 语句的一部分
- `FOREIGN KEY (owner) REFERENCES private_{projectSlug}.users(id)` 作为 table-level constraint

#### 注意事项
- `private_{projectSlug}` 数据库为 EndUser 身份存储库，`users` 表已由现有的 `PrivateDBManager` 管理
- `projectSlug` 必须在 DDL 生成时从 `ModelLocator.ProjectSlug` 传入 `mysql_converter.go`
- 跨库外键依赖 MySQL 同实例部署（同一集群），与现有 private DB 架构一致

---

## BDD 测试场景

> 以下 BDD 场景使用 Given/When/Then 语句描述，不含具体代码实现。

---

### Scenario 1：新建 Model 自动生成 owner 字段

```
Given 开发者已登录并拥有一个项目
When 开发者通过 createModel mutation 新建一个名为 "orders" 的 Model
Then 该 Model 的字段列表中包含一个 name="owner"、format="END_USER_REF" 的字段
And model.isRLSEnabled = true
And 客户 DB 中 orders 表含 owner VARCHAR(36) NOT NULL 列及对应外键约束
```

---

### Scenario 2：导入 Model 不自动生成 owner

```
Given 开发者已登录，且客户 DB 中已有表 "products"
When 开发者通过 importModel mutation 导入 "products" 表
Then 该 Model 的字段列表中不包含 name="owner" 的字段
And model.isRLSEnabled = false
```

---

### Scenario 3：添加第二个 EndUserRef 字段被拒绝

```
Given orders Model 已有一个 format="END_USER_REF" 的 owner 字段
When 开发者尝试通过 addFields mutation 再次添加一个 format="END_USER_REF" 的字段
Then addFields 的对应 AddFieldItemResult.success = false
And AddFieldItemResult.error.__typename = "EndUserRefAlreadyExists"
And AddFieldItemResult.error.code = "END_USER_REF_ALREADY_EXISTS"
```

---

### Scenario 4：删除 owner 字段返回 warning

```
Given orders Model 有 owner（END_USER_REF）字段，且 isRLSEnabled = true
When 开发者调用 removeField mutation 删除 "owner" 字段
Then 字段被成功删除
And RemoveFieldPayload.warning = RLS_WILL_DISABLE
And model.isRLSEnabled = false
And 客户 DB 中 orders 表的 owner 列及外键约束均被移除
```

---

### Scenario 5：Developer JWT 调用 Runtime 被拒绝

```
Given 开发者持有 iss="mc-developer" 的有效 JWT
When 开发者使用该 JWT 调用 Runtime GraphQL POST /graphql/org/{org}/project/{project}/db/{db}/model/{model}
Then HTTP 响应状态码为 401
And 响应体中 errors[0].extensions.code = "RUNTIME_UNAUTHORIZED"
```

---

### Scenario 6：EndUser 只能查到自己的数据

```
Given orders Model 已开启 RLS（selectPredicate={"owner":{"_eq":{"_auth":"uid"}}}）
And EndUser A 的 orders 表中有 3 条 owner=A 的记录
And EndUser B 的 orders 表中有 2 条 owner=B 的记录
When EndUser A 使用 iss="mc-enduser" 的 JWT 调用 Runtime findMany orders（不带任何 where 条件）
Then 响应中只返回 3 条 owner=A 的记录
And 不包含任何 owner=B 的记录
And 执行的 SQL WHERE 含参数化子句 "owner = ?" + args=[A_ID]
```

---

### Scenario 7：EndUser 显式传入其他用户 owner 返回空

```
Given orders Model 已开启 RLS，EndUser B 有数据
When EndUser A 调用 Runtime findMany orders，where 条件传入 { owner: B_ID }
Then 响应返回空列表（items: []）
And 不返回错误，也不返回 403
```

---

### Scenario 8：EndUser CreateOne 自动填充 owner

```
Given orders Model 已开启 RLS
When EndUser A 调用 Runtime createOne orders，input 中不传 owner 字段
Then 新创建的记录 owner = A 的 EndUserID
And 从 findMany 查询该记录可见 owner = A_ID
```

---

### Scenario 9：EndUser 故意传错 owner 被强制覆盖

```
Given orders Model 已开启 RLS
When EndUser A 调用 Runtime createOne orders，input 中传入 owner = B_ID
Then 新创建的记录 owner = A 的 EndUserID（B_ID 被覆盖）
And B_ID 不出现在数据库中
```

---

### Scenario 10：EndUser UpdateOne 跨用户数据静默返回 0 行

```
Given orders Model 已开启 RLS（updatePredicate = OWNER_EQUALS_USER）
And 数据库中存在一条 owner=B 的记录，id="order-123"
When EndUser A 调用 Runtime updateOne orders，where: { id: "order-123" }，尝试更新该记录
Then 响应中 affected rows = 0（静默，不报错）
And HTTP 状态码为 200
And 不返回 RECORD_NOT_FOUND，也不返回 403
And 数据库中该记录未被修改
```

---

### Scenario 10b：EndUser CreateOne CHECK 不通过返回 RLS_CHECK_VIOLATION

```
Given orders Model 已开启 RLS（insertCheck=false 即 ALWAYS_FALSE 常量简写）
When EndUser A 调用 Runtime createOne orders，传入任意数据
Then 响应中 errors[0].extensions.code = "RLS_CHECK_VIOLATION"
And HTTP 状态码为 200（GraphQL 错误走 200）
And 数据库中未插入任何记录
```

---

### Scenario 11：无 Policy 的 Model EndUser DENY ALL

```
Given announcements Model 无 owner 字段（isRLSEnabled = false，无 Policy）
And 数据库中有 5 条公告
When EndUser A 调用 Runtime findMany announcements（不带任何过滤）
Then 响应中 errors[0].extensions.code = "PERMISSION_DENIED"（Default Deny）
And 不返回任何数据行
And SQL 中没有执行（DENY ALL，不触达数据库）
```

---

### Scenario 12：迁移期间 Developer 旧 JWT 仍可访问设计态

```
Given Developer 持有 iss="modelcraft"（旧格式）的有效 JWT
When Developer 使用该 JWT 调用 设计态 GraphQL（project schema endpoint）
Then 响应正常，HTTP 200
And 操作执行成功（向后兼容）
```

---

### Scenario 13：EndUser DeleteOne 跨用户数据静默返回 0 行

```
Given orders Model 已开启 RLS（deletePredicate = OWNER_EQUALS_USER）
And 数据库中存在一条 owner=B 的记录，id="order-456"
When EndUser A 调用 Runtime deleteOne orders，where: { id: "order-456" }
Then 响应中 affected rows = 0（静默，不报错）
And 不返回 RECORD_NOT_FOUND，也不返回 403
And 数据库中该记录仍存在（未被删除）
```

---

### Scenario 14：owner 字段在 EndUser 视角 Mutation input 中不可见

```
Given orders Model 已开启 RLS
When EndUser A 查询 Runtime GraphQL Schema（introspection）
Then CreateOrdersInput 类型中不包含 owner 字段
And UpdateOrdersInput 类型中不包含 owner 字段
And OrdersObject 类型中包含 owner 字段（可读取）
```

---

### Scenario 15：JSON 表达式自定义策略生效

```
Given orders Model 已开启 RLS
When 开发者调用 setModelRLSPolicy，传入 selectPredicate={"owner":{"_eq":{"_auth":"uid"}}}
Then model.rlsPolicy.selectPredicate = {"owner":{"_eq":{"_auth":"uid"}}}
And EndUser A 调用 Runtime findMany orders → SQL 含 "owner = ?" + args=[A_ID]
And 只返回 owner = A 的记录
```

---

### Scenario 16：validateRLSExpr 校验非法 _exists 在 CHECK 中

```
Given orders Model 已开启 RLS
When 开发者调用 validateRLSExpr，operation=INSERT_CHECK，expression={"_exists":{"model":"...","where":{}}}
Then valid=false
And errors 包含 "CHECK 谓词不允许 _exists"
```

---

### Scenario 17：validateRLSExpr 校验未声明 auth 变量

```
Given Project 的 auth_schema 中没有声明 "org_id" 变量
When 开发者调用 validateRLSExpr，expression={"owner":{"_eq":{"_auth":"org_id"}}}
Then valid=false
And errors 包含 "未声明的 auth 变量: org_id"
```

---

### Scenario 18：auth_schema 声明变量后可在 Policy 中引用

```
Given 开发者调用 setProjectAuthSchema，声明 tenant_id（source: jwt.tenant_id, type: uuid）
When 开发者调用 validateRLSExpr，expression={"tenant_id":{"_eq":{"_auth":"tenant_id"}}}
Then valid=true
And errors=[]
```
