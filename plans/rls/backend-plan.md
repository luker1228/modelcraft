# RLS 后端实现计划

> 基于 RLS PRD 的后端实现计划
> - 主 PRD: `ai-metadata/prd/rls/prd.md`
> - 领域模型: `ai-metadata/prd/rls/rls-domain.puml`
> - API 合约: `plans/rls/api-contract.md`

---

## 1. 领域层设计

### 1.1 目录结构

```
internal/domain/
├── modeldesign/
│   ├── field_definition.go          # 扩展：添加 END_USER_REF Format
│   ├── model.go                     # 扩展：添加 ModelCreationSource, RLSEnabled 逻辑
│   ├── model_rls_policy.go          # 新增：ModelRLSPolicy 实体
│   ├── model_rls_repository.go      # 新增：ModelRLSPolicy Repository 接口
│   └── model_creation_source.go     # 新增：ModelCreationSource 枚举
│
├── rls/                             # 新增：RLS 领域子包
│   ├── end_user_identity.go         # EndUserIdentity VO
│   ├── rls_filter.go                # RLSFilter VO
│   ├── auth_schema.go               # AuthSchema VO
│   ├── auth_variable.go             # AuthVariable VO
│   ├── rls_preset.go                # RLSPreset 枚举
│   ├── json_expr.go                 # JsonExpr 类型与操作符定义
│   ├── policy_validator.go          # PolicyValidator 领域服务接口
│   ├── policy_compiler.go           # PolicyCompiler 领域服务接口
│   ├── policy_executor.go           # PolicyExecutor 领域服务接口
│   └── rls_resolver.go              # RLSResolver 领域服务接口
│
├── auth/
│   └── issuer.go                    # 扩展：添加 Issuer 枚举 (mc-developer, mc-enduser)
│
└── project/
    ├── project_auth_schema.go       # 新增：ProjectAuthSchema 实体
    └── auth_schema_repository.go    # 新增：AuthSchemaRepository 接口
```

### 1.2 核心实体与值对象

#### 1.2.1 FieldFormat 枚举扩展

**文件**: `internal/domain/modeldesign/field_definition.go`

```go
// FormatType 字段格式枚举 - 扩展
const (
    // ... 现有值保持不变

    // RLS: EndUserRef 格式 - 指向 private_{projectSlug}.users.id
    FormatEndUserRef FormatType = "END_USER_REF"
)

// init() 函数扩展
func init() {
    fieldTypeMap = map[FormatType]*FieldType{
        // ... 现有映射

        // RLS: EndUserRef 类型
        FormatEndUserRef: {SchemaType: SchemaTypeString, Format: FormatEndUserRef, Title: "归属用户"},
    }
}

// FieldDefinition 扩展方法
func (fd *FieldDefinition) IsEndUserRef() bool {
    if fd.Type == nil {
        return false
    }
    return fd.Type.Format == FormatEndUserRef
}

// EndUserRefAlreadyExists 错误
var ErrEndUserRefAlreadyExists = bizerrors.NewErrorDefinition(
    bizerrors.ErrorTypeConflict + ".FIELD.END_USER_REF",
    "EndUserRef field already exists in this model",
    "每个模型只允许一个归属字段",
)
```

#### 1.2.2 ModelCreationSource 枚举

**文件**: `internal/domain/modeldesign/model_creation_source.go`

```go
package modeldesign

// ModelCreationSource 模型创建来源
type ModelCreationSource string

const (
    ModelCreationSourceNew      ModelCreationSource = "NEW"      // 新建模型，自动生成 owner 字段
    ModelCreationSourceImported ModelCreationSource = "IMPORTED" // 导入模型，不生成 owner 字段
)
```

#### 1.2.3 ModelRLSPolicy 实体

**文件**: `internal/domain/modeldesign/model_rls_policy.go`

```go
package modeldesign

import "time"

// ModelRLSPolicy RLS 策略实体（五件套 JsonExpr）
type ModelRLSPolicy struct {
    ModelID         string    `json:"modelId"`
    SelectPredicate JsonExpr  `json:"selectPredicate"`  // SELECT USING
    InsertCheck     JsonExpr  `json:"insertCheck"`      // INSERT WITH CHECK
    UpdatePredicate JsonExpr  `json:"updatePredicate"`  // UPDATE USING
    UpdateCheck     JsonExpr  `json:"updateCheck"`      // UPDATE WITH CHECK
    DeletePredicate JsonExpr  `json:"deletePredicate"`  // DELETE USING
    CreatedAt       time.Time `json:"createdAt"`
    UpdatedAt       time.Time `json:"updatedAt"`
}

// GetPreset 返回当前策略匹配的 Preset，自定义组合返回 nil
func (p *ModelRLSPolicy) GetPreset() *RLSPreset {
    // 五件套全为 OWNER_EQUALS_USER → READ_WRITE_OWNER
    // select=true, 其余=OWNER_EQUALS_USER → READ_ALL_WRITE_OWNER
    // select=true, 其余=false → READ_ALL
    // 五件套全 true → READ_WRITE_ALL
    // 五件套全 false → NO_ACCESS
    // 其他组合 → nil (自定义)
    return p.matchPreset()
}

// ApplyPreset 应用预设策略
func (p *ModelRLSPolicy) ApplyPreset(preset RLSPreset) {
    switch preset {
    case RLSPresetReadWriteOwner:
        ownerEqualsUser := JsonExpr(`{"owner":{"_eq":{"_auth":"uid"}}}`)
        p.SelectPredicate = ownerEqualsUser
        p.InsertCheck = ownerEqualsUser
        p.UpdatePredicate = ownerEqualsUser
        p.UpdateCheck = ownerEqualsUser
        p.DeletePredicate = ownerEqualsUser
    case RLSPresetReadAllWriteOwner:
        ownerEqualsUser := JsonExpr(`{"owner":{"_eq":{"_auth":"uid"}}}`)
        p.SelectPredicate = JsonExpr(`true`)
        p.InsertCheck = ownerEqualsUser
        p.UpdatePredicate = ownerEqualsUser
        p.UpdateCheck = ownerEqualsUser
        p.DeletePredicate = ownerEqualsUser
    case RLSPresetReadAll:
        p.SelectPredicate = JsonExpr(`true`)
        p.InsertCheck = JsonExpr(`false`)
        p.UpdatePredicate = JsonExpr(`false`)
        p.UpdateCheck = JsonExpr(`false`)
        p.DeletePredicate = JsonExpr(`false`)
    case RLSPresetReadWriteAll:
        allTrue := JsonExpr(`true`)
        p.SelectPredicate = allTrue
        p.InsertCheck = allTrue
        p.UpdatePredicate = allTrue
        p.UpdateCheck = allTrue
        p.DeletePredicate = allTrue
    case RLSPresetNoAccess:
        allFalse := JsonExpr(`false`)
        p.SelectPredicate = allFalse
        p.InsertCheck = allFalse
        p.UpdatePredicate = allFalse
        p.UpdateCheck = allFalse
        p.DeletePredicate = allFalse
    }
}

// IsUsingTrue 判断 USING 谓词是否为 true（全量访问）
func (p *ModelRLSPolicy) IsUsingTrue() bool {
    return p.SelectPredicate.IsTrue() && 
           p.UpdatePredicate.IsTrue() && 
           p.DeletePredicate.IsTrue()
}

// IsDenyAll 判断是否为 DENY ALL 策略
func (p *ModelRLSPolicy) IsDenyAll() bool {
    return p.SelectPredicate.IsFalse() &&
           p.InsertCheck.IsFalse() &&
           p.UpdatePredicate.IsFalse() &&
           p.UpdateCheck.IsFalse() &&
           p.DeletePredicate.IsFalse()
}
```

#### 1.2.4 RLSPreset 枚举

**文件**: `internal/domain/rls/rls_preset.go`

```go
package rls

// RLSPreset RLS 预设策略
type RLSPreset string

const (
    RLSPresetReadWriteOwner    RLSPreset = "READ_WRITE_OWNER"     // 默认：读写自己
    RLSPresetReadAllWriteOwner RLSPreset = "READ_ALL_WRITE_OWNER" // 读取全部，写自己
    RLSPresetReadAll           RLSPreset = "READ_ALL"             // 只读全部
    RLSPresetReadWriteAll      RLSPreset = "READ_WRITE_ALL"       // 读写全部（高危）
    RLSPresetNoAccess          RLSPreset = "NO_ACCESS"            // 无访问
)

// IsDangerous 判断是否为高危策略
func (p RLSPreset) IsDangerous() bool {
    return p == RLSPresetReadWriteAll
}

// String 返回显示名称
func (p RLSPreset) String() string {
    return string(p)
}
```

#### 1.2.5 JsonExpr 类型

**文件**: `internal/domain/rls/json_expr.go`

```go
package rls

import "encoding/json"

// JsonExpr RLS 表达式（JSON 字符串）
type JsonExpr string

// IsTrue 判断是否为 true 常量
func (e JsonExpr) IsTrue() bool {
    var v interface{}
    if err := json.Unmarshal([]byte(e), &v); err != nil {
        return false
    }
    if b, ok := v.(bool); ok {
        return b
    }
    // 检查 {"_const": true}
    var obj map[string]interface{}
    if err := json.Unmarshal([]byte(e), &obj); err == nil {
        if v, ok := obj["_const"]; ok {
            if b, ok := v.(bool); ok {
                return b
            }
        }
    }
    return false
}

// IsFalse 判断是否为 false 常量
func (e JsonExpr) IsFalse() bool {
    var v interface{}
    if err := json.Unmarshal([]byte(e), &v); err != nil {
        return false
    }
    if b, ok := v.(bool); ok {
        return !b
    }
    // 检查 {"_const": false}
    var obj map[string]interface{}
    if err := json.Unmarshal([]byte(e), &obj); err == nil {
        if v, ok := obj["_const"]; ok {
            if b, ok := v.(bool); ok {
                return !b
            }
        }
    }
    return false
}

// IsOwnerEqualsUser 判断是否为 {"owner":{"_eq":{"_auth":"uid"}}}
func (e JsonExpr) IsOwnerEqualsUser() bool {
    var obj map[string]interface{}
    if err := json.Unmarshal([]byte(e), &obj); err != nil {
        return false
    }
    owner, ok := obj["owner"].(map[string]interface{})
    if !ok {
        return false
    }
    eq, ok := owner["_eq"].(map[string]interface{})
    if !ok {
        return false
    }
    auth, ok := eq["_auth"].(string)
    return ok && auth == "uid"
}

// ExprType 表达式类型（用于校验时区分 PREDICATE vs CHECK）
type ExprType string

const (
    ExprTypeSelectPredicate ExprType = "SELECT_PREDICATE"
    ExprTypeInsertCheck     ExprType = "INSERT_CHECK"
    ExprTypeUpdatePredicate ExprType = "UPDATE_PREDICATE"
    ExprTypeUpdateCheck     ExprType = "UPDATE_CHECK"
    ExprTypeDeletePredicate ExprType = "DELETE_PREDICATE"
)

// IsPredicate 判断是否为 PREDICATE 类型（允许 _exists, _ref）
func (t ExprType) IsPredicate() bool {
    return t == ExprTypeSelectPredicate || 
           t == ExprTypeUpdatePredicate || 
           t == ExprTypeDeletePredicate
}

// IsCheck 判断是否为 CHECK 类型（不允许 _exists, _ref）
func (t ExprType) IsCheck() bool {
    return t == ExprTypeInsertCheck || t == ExprTypeUpdateCheck
}
```

#### 1.2.6 EndUserIdentity 值对象

**文件**: `internal/domain/rls/end_user_identity.go`

```go
package rls

// EndUserIdentity 终端用户身份信息
type EndUserIdentity struct {
    EndUserID string `json:"endUserId"`
    Issuer    string `json:"issuer"` // mc-developer 或 mc-enduser
}

// IsEndUser 判断是否为合法的 Runtime 调用者（EndUser JWT）
func (e *EndUserIdentity) IsEndUser() bool {
    return e.Issuer == "mc-enduser"
}

// IsDeveloper 判断是否为 Developer JWT
func (e *EndUserIdentity) IsDeveloper() bool {
    return e.Issuer == "mc-developer"
}
```

#### 1.2.7 RLSFilter 值对象

**文件**: `internal/domain/rls/rls_filter.go`

```go
package rls

// RLSFilter RLS 运行时过滤器
type RLSFilter struct {
    SelectPredicate  JsonExpr `json:"selectPredicate"`
    InsertCheck      JsonExpr `json:"insertCheck"`
    UpdatePredicate  JsonExpr `json:"updatePredicate"`
    UpdateCheck      JsonExpr `json:"updateCheck"`
    DeletePredicate  JsonExpr `json:"deletePredicate"`
    FieldName        string   `json:"fieldName"`  // 固定为 "owner"
    EndUserID        string   `json:"endUserId"`
}

// IsDenyAll 判断是否 DENY ALL（所有谓词都是 false）
func (f *RLSFilter) IsDenyAll() bool {
    return f.SelectPredicate.IsFalse() &&
           f.UpdatePredicate.IsFalse() &&
           f.DeletePredicate.IsFalse()
}

// ShouldInjectWhere 判断是否需要注入 WHERE 条件
func (f *RLSFilter) ShouldInjectWhere() bool {
    // 不是全量 true 且不是全量 false
    return !f.SelectPredicate.IsTrue() && !f.SelectPredicate.IsFalse()
}
```

#### 1.2.8 AuthSchema / AuthVariable 值对象

**文件**: `internal/domain/project/project_auth_schema.go`

```go
package project

// AuthSchema Project 级别认证变量配置
type AuthSchema struct {
    ProjectID string         `json:"projectId"`
    Variables []AuthVariable `json:"variables"`
}

// GetVariable 获取指定名称的变量（uid 内置，不返回）
func (a *AuthSchema) GetVariable(name string) *AuthVariable {
    if name == "uid" {
        return &AuthVariable{Name: "uid", Source: "jwt.user_id", Type: AuthVarTypeUUID}
    }
    for _, v := range a.Variables {
        if v.Name == name {
            return &v
        }
    }
    return nil
}

// IsValidRef 判断变量引用是否合法
func (a *AuthSchema) IsValidRef(name string) bool {
    return a.GetVariable(name) != nil
}

// AuthVariable 认证变量定义
type AuthVariable struct {
    Name   string          `json:"name"`   // 如 "tenant_id"
    Source string          `json:"source"` // JWT 路径，如 "jwt.tenant_id"
    Type   AuthVarType     `json:"type"`   // uuid | string | integer
}

// AuthVarType 认证变量类型
type AuthVarType string

const (
    AuthVarTypeUUID    AuthVarType = "UUID"
    AuthVarTypeString  AuthVarType = "STRING"
    AuthVarTypeInteger AuthVarType = "INTEGER"
)
```

#### 1.2.9 Issuer 枚举扩展

**文件**: `internal/domain/auth/issuer.go`

```go
package auth

// Issuer JWT 签发者
type Issuer string

const (
    IssuerDeveloper Issuer = "mc-developer"  // 开发者 JWT
    IssuerEndUser   Issuer = "mc-enduser"    // 终端用户 JWT
    IssuerLegacy    Issuer = "modelcraft"    // 兼容旧版（需迁移）
)

// IsValid 判断是否为合法的 Issuer
func (i Issuer) IsValid() bool {
    switch i {
    case IssuerDeveloper, IssuerEndUser, IssuerLegacy:
        return true
    }
    return false
}
```

### 1.3 Domain Service 接口

#### 1.3.1 PolicyValidator

**文件**: `internal/domain/rls/policy_validator.go`

```go
package rls

import (
    "context"
    "modelcraft/internal/domain/modeldesign"
    "modelcraft/internal/domain/project"
)

// PolicyValidator RLS 表达式校验器
type PolicyValidator interface {
    // Validate 校验 JSON 表达式合法性
    // - JSON Schema 结构合法性
    // - 字段名白名单（对照 Model 字段列表）
    // - _auth 变量白名单（uid 内置 + auth_schema 声明）
    // - _exists.model 白名单（已知 Model 或系统表）
    // - CHECK 类型不含 _exists / _ref
    Validate(ctx context.Context, expr JsonExpr, exprType ExprType, 
             modelSchema *modeldesign.DataModel, authSchema *project.AuthSchema) []ValidationError
}

// ValidationError 校验错误
type ValidationError struct {
    Path    string `json:"path"`    // 错误位置，如 "selectPredicate._and[0].owner"
    Message string `json:"message"` // 错误描述
    Code    string `json:"code"`    // 错误码
}
```

#### 1.3.2 PolicyCompiler

**文件**: `internal/domain/rls/policy_compiler.go`

```go
package rls

import "context"

// PolicyCompiler RLS 表达式编译器
type PolicyCompiler interface {
    // Compile 将 JSON 表达式编译为 CompiledPolicy
    // 递归解析 JSON 为参数化 SQL 片段
    // _auth.uid → ? 占位符
    // _ref → 跨表字段引用（仅 PREDICATE 允许）
    Compile(ctx context.Context, expr JsonExpr) (*CompiledPolicy, error)
}

// CompiledPolicy 编译后的策略
type CompiledPolicy struct {
    SQL    string        // 参数化 SQL 片段
    Params []interface{} // 参数占位符说明（如 {"_auth": "uid"}）
}
```

#### 1.3.3 PolicyExecutor

**文件**: `internal/domain/rls/policy_executor.go`

```go
package rls

import "context"

// PolicyExecutor RLS 策略执行器
type PolicyExecutor interface {
    // ToSQL 将编译后的策略绑定运行时 authCtx 生成最终 SQL
    // 返回参数化查询 + 绑定参数数组
    ToSQL(ctx context.Context, compiled *CompiledPolicy, authCtx *AuthContext) (string, []interface{}, error)
    
    // ValidateCheck 应用层校验 CHECK 约束（用于 insertCheck / updateCheck）
    // 返回 RLSCheckViolation 错误（如有）
    ValidateCheck(ctx context.Context, expr JsonExpr, rowData map[string]interface{}, 
                  authCtx *AuthContext) error
}

// AuthContext 运行时认证上下文
type AuthContext struct {
    EndUserID string                 `json:"endUserId"`
    Variables map[string]interface{} `json:"variables"` // auth_schema 声明的扩展变量
}
```

#### 1.3.4 RLSResolver

**文件**: `internal/domain/rls/rls_resolver.go`

```go
package rls

import (
    "context"
    "modelcraft/internal/domain/modeldesign"
)

// RLSResolver RLS 过滤器解析器
type RLSResolver interface {
    // Resolve 根据身份和 Model 策略解析 RLSFilter
    // - identity.isEndUser() == false → nil（不过滤，Developer 访问）
    // - model.getPolicy() == nil → DENY ALL（无 Policy = Default Deny）
    // - 否则 → RLSFilter { 五件套 JsonExpr, endUserId }
    Resolve(ctx context.Context, identity *EndUserIdentity, 
            model *modeldesign.DataModel) (*RLSFilter, error)
}

// DenyAllFilter 返回 DENY ALL 过滤器（全 false）
var DenyAllFilter = &RLSFilter{
    SelectPredicate: JsonExpr(`false`),
    InsertCheck:     JsonExpr(`false`),
    UpdatePredicate: JsonExpr(`false`),
    UpdateCheck:     JsonExpr(`false`),
    DeletePredicate: JsonExpr(`false`),
    FieldName:       "owner",
}
```

### 1.4 Repository 接口

#### 1.4.1 ModelRLSPolicyRepository

**文件**: `internal/domain/modeldesign/model_rls_repository.go`

```go
package modeldesign

import "context"

// ModelRLSPolicyRepository RLS 策略 Repository 接口
type ModelRLSPolicyRepository interface {
    // GetByModelID 根据 Model ID 获取 Policy
    GetByModelID(ctx context.Context, orgName, projectSlug, modelID string) (*ModelRLSPolicy, error)
    
    // Save 保存 Policy（upsert）
    Save(ctx context.Context, orgName, projectSlug string, policy *ModelRLSPolicy) error
    
    // DeleteByModelID 删除指定 Model 的 Policy
    DeleteByModelID(ctx context.Context, orgName, projectSlug, modelID string) error
    
    // ExistsByModelID 判断指定 Model 是否有 Policy
    ExistsByModelID(ctx context.Context, orgName, projectSlug, modelID string) (bool, error)
}
```

#### 1.4.2 AuthSchemaRepository

**文件**: `internal/domain/project/auth_schema_repository.go`

```go
package project

import "context"

// AuthSchemaRepository Project AuthSchema Repository 接口
type AuthSchemaRepository interface {
    // GetByProjectID 根据 Project ID 获取 AuthSchema
    GetByProjectID(ctx context.Context, orgName, projectSlug string) (*AuthSchema, error)
    
    // Save 保存 AuthSchema（upsert）
    Save(ctx context.Context, authSchema *AuthSchema) error
    
    // DeleteByProjectID 删除指定 Project 的 AuthSchema
    DeleteByProjectID(ctx context.Context, orgName, projectSlug string) error
}
```

---

## 2. Repository 层设计

### 2.1 目录结构

```
internal/infrastructure/repository/
├── sqlc/
│   ├── model_rls_policy.sql.go      # sqlc 生成：RLS Policy CRUD
│   └── project_auth_schema.sql.go   # sqlc 生成：AuthSchema CRUD
│
├── model_rls_policy_repo.go         # ModelRLSPolicyRepository 实现
└── auth_schema_repo.go              # AuthSchemaRepository 实现
```

### 2.2 SQL 查询定义

**文件**: `db/queries/model_rls_policy.sql`

```sql
-- name: GetModelRLSPolicy :one
SELECT 
    model_id,
    select_predicate,
    insert_check,
    update_predicate,
    update_check,
    delete_predicate,
    created_at,
    updated_at
FROM model_rls_policies
WHERE org_name = ? AND project_slug = ? AND model_id = ?;

-- name: UpsertModelRLSPolicy :exec
INSERT INTO model_rls_policies (
    model_id,
    org_name,
    project_slug,
    select_predicate,
    insert_check,
    update_predicate,
    update_check,
    delete_predicate
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    select_predicate = VALUES(select_predicate),
    insert_check = VALUES(insert_check),
    update_predicate = VALUES(update_predicate),
    update_check = VALUES(update_check),
    delete_predicate = VALUES(delete_predicate),
    updated_at = CURRENT_TIMESTAMP(3);

-- name: DeleteModelRLSPolicy :exec
DELETE FROM model_rls_policies
WHERE org_name = ? AND project_slug = ? AND model_id = ?;

-- name: ExistsModelRLSPolicy :one
SELECT EXISTS(
    SELECT 1 FROM model_rls_policies
    WHERE org_name = ? AND project_slug = ? AND model_id = ?
) as exists_flag;
```

**文件**: `db/queries/project_auth_schema.sql`

```sql
-- name: GetProjectAuthSchema :one
SELECT 
    org_name,
    project_slug,
    variables,
    created_at,
    updated_at
FROM project_auth_schemas
WHERE org_name = ? AND project_slug = ?;

-- name: UpsertProjectAuthSchema :exec
INSERT INTO project_auth_schemas (
    org_name,
    project_slug,
    variables
) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE
    variables = VALUES(variables),
    updated_at = CURRENT_TIMESTAMP(3);

-- name: DeleteProjectAuthSchema :exec
DELETE FROM project_auth_schemas
WHERE org_name = ? AND project_slug = ?;
```

### 2.3 Repository 实现

**文件**: `internal/infrastructure/repository/model_rls_policy_repo.go`

```go
package repository

import (
    "context"
    "modelcraft/internal/domain/modeldesign"
    "modelcraft/internal/infrastructure/persistence/sqlc"
    "modelcraft/pkg/bizerrors"
)

// SqlModelRLSPolicyRepository ModelRLSPolicyRepository 的 sqlc 实现
type SqlModelRLSPolicyRepository struct {
    q sqlc.Querier
}

var _ modeldesign.ModelRLSPolicyRepository = (*SqlModelRLSPolicyRepository)(nil)

func NewSqlModelRLSPolicyRepository(q sqlc.Querier) *SqlModelRLSPolicyRepository {
    return &SqlModelRLSPolicyRepository{q: q}
}

func (r *SqlModelRLSPolicyRepository) GetByModelID(ctx context.Context, orgName, projectSlug, modelID string) (*modeldesign.ModelRLSPolicy, error) {
    row, err := r.q.GetModelRLSPolicy(ctx, sqlc.GetModelRLSPolicyParams{
        OrgName:     orgName,
        ProjectSlug: projectSlug,
        ModelID:     modelID,
    })
    if err != nil {
        if sqlc.IsNotFoundError(err) {
            return nil, nil
        }
        return nil, bizerrors.Wrapf(err, "failed to get RLS policy for model %s", modelID)
    }
    return &modeldesign.ModelRLSPolicy{
        ModelID:         row.ModelID,
        SelectPredicate: modeldesign.JsonExpr(row.SelectPredicate),
        InsertCheck:     modeldesign.JsonExpr(row.InsertCheck),
        UpdatePredicate: modeldesign.JsonExpr(row.UpdatePredicate),
        UpdateCheck:     modeldesign.JsonExpr(row.UpdateCheck),
        DeletePredicate: modeldesign.JsonExpr(row.DeletePredicate),
        CreatedAt:       row.CreatedAt,
        UpdatedAt:       row.UpdatedAt,
    }, nil
}

func (r *SqlModelRLSPolicyRepository) Save(ctx context.Context, orgName, projectSlug string, policy *modeldesign.ModelRLSPolicy) error {
    err := r.q.UpsertModelRLSPolicy(ctx, sqlc.UpsertModelRLSPolicyParams{
        ModelID:         policy.ModelID,
        OrgName:         orgName,
        ProjectSlug:     projectSlug,
        SelectPredicate: string(policy.SelectPredicate),
        InsertCheck:     string(policy.InsertCheck),
        UpdatePredicate: string(policy.UpdatePredicate),
        UpdateCheck:     string(policy.UpdateCheck),
        DeletePredicate: string(policy.DeletePredicate),
    })
    if err != nil {
        return bizerrors.Wrapf(err, "failed to save RLS policy for model %s", policy.ModelID)
    }
    return nil
}

func (r *SqlModelRLSPolicyRepository) DeleteByModelID(ctx context.Context, orgName, projectSlug, modelID string) error {
    err := r.q.DeleteModelRLSPolicy(ctx, sqlc.DeleteModelRLSPolicyParams{
        OrgName:     orgName,
        ProjectSlug: projectSlug,
        ModelID:     modelID,
    })
    if err != nil {
        return bizerrors.Wrapf(err, "failed to delete RLS policy for model %s", modelID)
    }
    return nil
}

func (r *SqlModelRLSPolicyRepository) ExistsByModelID(ctx context.Context, orgName, projectSlug, modelID string) (bool, error) {
    result, err := r.q.ExistsModelRLSPolicy(ctx, sqlc.ExistsModelRLSPolicyParams{
        OrgName:     orgName,
        ProjectSlug: projectSlug,
        ModelID:     modelID,
    })
    if err != nil {
        return false, bizerrors.Wrapf(err, "failed to check RLS policy existence for model %s", modelID)
    }
    return result == 1, nil
}
```

---

## 3. App 层设计

### 3.1 目录结构

```
internal/app/
├── modeldesign/
│   ├── model_app_service.go         # 扩展：新建 Model 时自动创建 owner + Policy
│   └── field_app_service.go         # 扩展：添加/删除 EndUserRef 字段时检查唯一性
│
├── rls/                             # 新增：RLS App Service
│   ├── model_rls_policy_service.go  # ModelRLSPolicyAppService
│   └── policy_validator_impl.go     # PolicyValidator 实现
│
├── project/                         # 新增/扩展
│   └── auth_schema_service.go       # AuthSchemaAppService
│
└── runtime/                         # 新增/扩展
    ├── rls_resolver_impl.go         # RLSResolver 实现
    ├── policy_compiler_impl.go      # PolicyCompiler 实现
    └── policy_executor_impl.go      # PolicyExecutor 实现
```

### 3.2 ModelRLSPolicyAppService

**文件**: `internal/app/rls/model_rls_policy_service.go`

```go
package rls

import (
    "context"
    "modelcraft/internal/domain/modeldesign"
    "modelcraft/internal/domain/project"
    "modelcraft/pkg/bizerrors"
)

// SetModelRLSPolicyCommand 设置 RLS 策略命令
type SetModelRLSPolicyCommand struct {
    ModelID         string              `json:"modelId"`
    SelectPredicate modeldesign.JsonExpr `json:"selectPredicate"`
    InsertCheck     modeldesign.JsonExpr `json:"insertCheck"`
    UpdatePredicate modeldesign.JsonExpr `json:"updatePredicate"`
    UpdateCheck     modeldesign.JsonExpr `json:"updateCheck"`
    DeletePredicate modeldesign.JsonExpr `json:"deletePredicate"`
}

// ModelRLSPolicyAppService RLS 策略应用服务
type ModelRLSPolicyAppService struct {
    policyRepo  modeldesign.ModelRLSPolicyRepository
    modelRepo   modeldesign.ModelRepository
    fieldRepo   modeldesign.FieldDefinitionRepository
    authSchemaRepo project.AuthSchemaRepository
    validator   modeldesign.PolicyValidator
    txManager   sqlc.TxManager
}

// SetPolicy 设置 Model RLS 策略
func (s *ModelRLSPolicyAppService) SetPolicy(ctx context.Context, orgName, projectSlug string, 
    cmd SetModelRLSPolicyCommand) (*modeldesign.ModelRLSPolicy, error) {
    
    // 1. 检查 Model 是否存在且有 owner 字段
    model, err := s.modelRepo.GetByID(ctx, orgName, projectSlug, cmd.ModelID)
    if err != nil {
        return nil, err
    }
    if model == nil {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, cmd.ModelID)
    }
    
    // 2. 检查是否有 owner 字段
    fields, err := s.fieldRepo.ListByModelID(ctx, orgName, projectSlug, cmd.ModelID)
    if err != nil {
        return nil, err
    }
    hasOwner := false
    for _, f := range fields {
        if f.IsEndUserRef() {
            hasOwner = true
            break
        }
    }
    if !hasOwner {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelHasNoOwnerField, cmd.ModelID)
    }
    
    // 3. 获取 AuthSchema 用于校验
    authSchema, err := s.authSchemaRepo.GetByProjectID(ctx, orgName, projectSlug)
    if err != nil {
        return nil, err
    }
    if authSchema == nil {
        authSchema = &project.AuthSchema{ProjectID: projectSlug}
    }
    
    // 4. 校验五件套表达式
    dataModel := &modeldesign.DataModel{ModelMeta: *model, Fields: fields}
    exprs := map[modeldesign.ExprType]modeldesign.JsonExpr{
        modeldesign.ExprTypeSelectPredicate: cmd.SelectPredicate,
        modeldesign.ExprTypeInsertCheck:     cmd.InsertCheck,
        modeldesign.ExprTypeUpdatePredicate: cmd.UpdatePredicate,
        modeldesign.ExprTypeUpdateCheck:     cmd.UpdateCheck,
        modeldesign.ExprTypeDeletePredicate: cmd.DeletePredicate,
    }
    
    for exprType, expr := range exprs {
        if errors := s.validator.Validate(ctx, expr, exprType, dataModel, authSchema); len(errors) > 0 {
            return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.InvalidRLSExpression, 
                errors[0].Path+": "+errors[0].Message)
        }
    }
    
    // 5. 保存策略
    policy := &modeldesign.ModelRLSPolicy{
        ModelID:         cmd.ModelID,
        SelectPredicate: cmd.SelectPredicate,
        InsertCheck:     cmd.InsertCheck,
        UpdatePredicate: cmd.UpdatePredicate,
        UpdateCheck:     cmd.UpdateCheck,
        DeletePredicate: cmd.DeletePredicate,
    }
    
    if err := s.policyRepo.Save(ctx, orgName, projectSlug, policy); err != nil {
        return nil, err
    }
    
    return policy, nil
}

// GetPolicy 获取 Model RLS 策略
func (s *ModelRLSPolicyAppService) GetPolicy(ctx context.Context, orgName, projectSlug, modelID string) (*modeldesign.ModelRLSPolicy, error) {
    return s.policyRepo.GetByModelID(ctx, orgName, projectSlug, modelID)
}

// ApplyPreset 应用预设策略
func (s *ModelRLSPolicyAppService) ApplyPreset(ctx context.Context, orgName, projectSlug, modelID string, 
    preset modeldesign.RLSPreset) (*modeldesign.ModelRLSPolicy, error) {
    
    policy := &modeldesign.ModelRLSPolicy{ModelID: modelID}
    policy.ApplyPreset(preset)
    
    cmd := SetModelRLSPolicyCommand{
        ModelID:         modelID,
        SelectPredicate: policy.SelectPredicate,
        InsertCheck:     policy.InsertCheck,
        UpdatePredicate: policy.UpdatePredicate,
        UpdateCheck:     policy.UpdateCheck,
        DeletePredicate: policy.DeletePredicate,
    }
    
    return s.SetPolicy(ctx, orgName, projectSlug, cmd)
}

// ValidateExpr 校验 RLS 表达式（用于 UI 实时校验）
func (s *ModelRLSPolicyAppService) ValidateExpr(ctx context.Context, orgName, projectSlug, modelID string,
    exprType modeldesign.ExprType, expr modeldesign.JsonExpr) []modeldesign.ValidationError {
    
    // 获取 Model 和 AuthSchema
    model, err := s.modelRepo.GetByID(ctx, orgName, projectSlug, modelID)
    if err != nil || model == nil {
        return []modeldesign.ValidationError{{
            Path:    "modelId",
            Message: "Model not found",
            Code:    "MODEL_NOT_FOUND",
        }}
    }
    
    fields, err := s.fieldRepo.ListByModelID(ctx, orgName, projectSlug, modelID)
    if err != nil {
        return []modeldesign.ValidationError{{
            Path:    "",
            Message: err.Error(),
            Code:    "INTERNAL_ERROR",
        }}
    }
    
    authSchema, _ := s.authSchemaRepo.GetByProjectID(ctx, orgName, projectSlug)
    if authSchema == nil {
        authSchema = &project.AuthSchema{ProjectID: projectSlug}
    }
    
    dataModel := &modeldesign.DataModel{ModelMeta: *model, Fields: fields}
    return s.validator.Validate(ctx, expr, exprType, dataModel, authSchema)
}
```

### 3.3 AuthSchemaAppService

**文件**: `internal/app/project/auth_schema_service.go`

```go
package project

import (
    "context"
    "modelcraft/internal/domain/project"
    "modelcraft/pkg/bizerrors"
)

// SetAuthSchemaCommand 设置 AuthSchema 命令
type SetAuthSchemaCommand struct {
    ProjectSlug string               `json:"projectSlug"`
    Variables   []project.AuthVariable `json:"variables"`
}

// AuthSchemaAppService AuthSchema 应用服务
type AuthSchemaAppService struct {
    authSchemaRepo project.AuthSchemaRepository
    projectRepo    project.Repository
}

// SetAuthSchema 设置 Project AuthSchema
func (s *AuthSchemaAppService) SetAuthSchema(ctx context.Context, orgName string, 
    cmd SetAuthSchemaCommand) (*project.AuthSchema, error) {
    
    // 1. 检查 Project 是否存在
    p, err := s.projectRepo.GetBySlug(ctx, orgName, cmd.ProjectSlug)
    if err != nil {
        return nil, err
    }
    if p == nil {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ProjectNotFound, cmd.ProjectSlug)
    }
    
    // 2. 校验变量（不允许声明 uid）
    for _, v := range cmd.Variables {
        if v.Name == "uid" {
            return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, 
                "'uid' is a reserved variable and cannot be declared")
        }
    }
    
    // 3. 保存 AuthSchema
    authSchema := &project.AuthSchema{
        ProjectID: cmd.ProjectSlug,
        Variables: cmd.Variables,
    }
    
    if err := s.authSchemaRepo.Save(ctx, authSchema); err != nil {
        return nil, err
    }
    
    return authSchema, nil
}

// GetAuthSchema 获取 Project AuthSchema
func (s *AuthSchemaAppService) GetAuthSchema(ctx context.Context, orgName, projectSlug string) (*project.AuthSchema, error) {
    authSchema, err := s.authSchemaRepo.GetByProjectID(ctx, orgName, projectSlug)
    if err != nil {
        return nil, err
    }
    if authSchema == nil {
        // 返回空 AuthSchema
        return &project.AuthSchema{ProjectID: projectSlug, Variables: []project.AuthVariable{}}, nil
    }
    return authSchema, nil
}
```

### 3.4 PolicyValidator 实现

**文件**: `internal/app/rls/policy_validator_impl.go`

```go
package rls

import (
    "context"
    "encoding/json"
    "modelcraft/internal/domain/modeldesign"
    "modelcraft/internal/domain/project"
)

// PolicyValidatorImpl PolicyValidator 实现
type PolicyValidatorImpl struct{}

func NewPolicyValidator() *PolicyValidatorImpl {
    return &PolicyValidatorImpl{}
}

func (v *PolicyValidatorImpl) Validate(ctx context.Context, expr modeldesign.JsonExpr, 
    exprType modeldesign.ExprType, modelSchema *modeldesign.DataModel, 
    authSchema *project.AuthSchema) []modeldesign.ValidationError {
    
    var errors []modeldesign.ValidationError
    
    // 1. JSON 合法性校验
    var root interface{}
    if err := json.Unmarshal([]byte(expr), &root); err != nil {
        return []modeldesign.ValidationError{{
            Path:    "",
            Message: "Invalid JSON: " + err.Error(),
            Code:    "INVALID_JSON",
        }}
    }
    
    // 2. 常量简写 true/false 直接通过
    if b, ok := root.(bool); ok {
        return nil
    }
    
    // 3. 递归校验
    obj, ok := root.(map[string]interface{})
    if !ok {
        return []modeldesign.ValidationError{{
            Path:    "",
            Message: "Expression must be an object or boolean constant",
            Code:    "INVALID_STRUCTURE",
        }}
    }
    
    errors = append(errors, v.validateNode(ctx, obj, "", exprType, modelSchema, authSchema)...)
    
    return errors
}

func (v *PolicyValidatorImpl) validateNode(ctx context.Context, node map[string]interface{}, 
    path string, exprType modeldesign.ExprType, modelSchema *modeldesign.DataModel, 
    authSchema *project.AuthSchema) []modeldesign.ValidationError {
    
    var errors []modeldesign.ValidationError
    
    for key, value := range node {
        currentPath := path
        if currentPath != "" {
            currentPath += "."
        }
        currentPath += key
        
        switch key {
        case "_and", "_or":
            // 逻辑操作符：值必须是数组
            arr, ok := value.([]interface{})
            if !ok {
                errors = append(errors, modeldesign.ValidationError{
                    Path:    currentPath,
                    Message: "_and/_or value must be an array",
                    Code:    "INVALID_OPERATOR",
                })
                continue
            }
            for i, item := range arr {
                itemObj, ok := item.(map[string]interface{})
                if !ok {
                    errors = append(errors, modeldesign.ValidationError{
                        Path:    currentPath + "[" + strconv.Itoa(i) + "]",
                        Message: "Array item must be an object",
                        Code:    "INVALID_ARRAY_ITEM",
                    })
                    continue
                }
                errors = append(errors, v.validateNode(ctx, itemObj, 
                    currentPath+"["+strconv.Itoa(i)+"]", exprType, modelSchema, authSchema)...)
            }
            
        case "_not":
            // _not: 值必须是对象
            obj, ok := value.(map[string]interface{})
            if !ok {
                errors = append(errors, modeldesign.ValidationError{
                    Path:    currentPath,
                    Message: "_not value must be an object",
                    Code:    "INVALID_OPERATOR",
                })
                continue
            }
            errors = append(errors, v.validateNode(ctx, obj, currentPath, exprType, modelSchema, authSchema)...)
            
        case "_exists":
            // _exists: 仅 PREDICATE 允许，CHECK 不允许
            if exprType.IsCheck() {
                errors = append(errors, modeldesign.ValidationError{
                    Path:    currentPath,
                    Message: "_exists is not allowed in CHECK expressions (insertCheck/updateCheck)",
                    Code:    "EXISTS_IN_CHECK",
                })
                continue
            }
            // 校验 _exists 结构
            existsObj, ok := value.(map[string]interface{})
            if !ok {
                errors = append(errors, modeldesign.ValidationError{
                    Path:    currentPath,
                    Message: "_exists value must be an object with 'model' and 'where'",
                    Code:    "INVALID_EXISTS",
                })
                continue
            }
            // TODO: 校验 model 存在性，where 合法性
            
        case "_auth":
            // _auth: 校验变量名是否在白名单
            varName, ok := value.(string)
            if !ok {
                errors = append(errors, modeldesign.ValidationError{
                    Path:    currentPath,
                    Message: "_auth value must be a string",
                    Code:    "INVALID_AUTH_REF",
                })
                continue
            }
            if !authSchema.IsValidRef(varName) {
                errors = append(errors, modeldesign.ValidationError{
                    Path:    currentPath,
                    Message: fmt.Sprintf("Unknown auth variable '%s'. Declare it in project auth_schema first.", varName),
                    Code:    "UNKNOWN_AUTH_VAR",
                })
            }
            
        case "_ref":
            // _ref: 仅 PREDICATE 允许
            if exprType.IsCheck() {
                errors = append(errors, modeldesign.ValidationError{
                    Path:    currentPath,
                    Message: "_ref is not allowed in CHECK expressions (insertCheck/updateCheck)",
                    Code:    "REF_IN_CHECK",
                })
                continue
            }
            
        default:
            // 字段比较：校验字段名是否存在
            if !modelSchema.HasField(key) {
                errors = append(errors, modeldesign.ValidationError{
                    Path:    currentPath,
                    Message: fmt.Sprintf("Unknown field '%s'", key),
                    Code:    "UNKNOWN_FIELD",
                })
                continue
            }
            // 校验比较操作符
            compObj, ok := value.(map[string]interface{})
            if !ok {
                errors = append(errors, modeldesign.ValidationError{
                    Path:    currentPath,
                    Message: "Field comparison value must be an object",
                    Code:    "INVALID_COMPARISON",
                })
                continue
            }
            // TODO: 校验 _eq, _neq, _gt, _gte, _lt, _lte, _in, _nin, _is_null
        }
    }
    
    return errors
}
```

---

## 4. Runtime 层改造

### 4.1 目录结构

```
internal/
├── interfaces/
│   ├── http/
│   │   └── middleware/
│   │       └── runtime_auth_middleware.go  # Runtime JWT 认证中间件
│   │
│   └── graphql/
│       ├── resolver/
│       │   ├── runtime/                    # Runtime Resolver 扩展
│       │   │   ├── runtime_query_resolver.go   # Query: 自动注入 WHERE
│       │   │   ├── runtime_mutation_resolver.go # Mutation: 自动填充 owner
│       │   │   └── rls_resolver_impl.go         # RLSResolver 实现
│       │   └── ...
│       └── ...
│
├── app/runtime/                       # 新增：Runtime RLS 应用服务
│   ├── rls_injection_service.go       # WHERE 注入服务
│   └── owner_auto_fill_service.go     # owner 自动填充服务
│
└── domain/modelruntime/               # 扩展：Runtime 领域层
    ├── rls_filter.go                  # RLSFilter 应用
    └── query_builder.go               # Query 构建器扩展
```

### 4.2 Runtime JWT 认证中间件

**文件**: `internal/interfaces/http/middleware/runtime_auth_middleware.go`

```go
package middleware

import (
    "context"
    "net/http"
    "strings"
    
    "modelcraft/internal/domain/rls"
    "modelcraft/pkg/logfacade"
)

// RuntimeAuthMiddleware Runtime 端点认证中间件
type RuntimeAuthMiddleware struct {
    jwtValidator JWTValidator
    logger       logfacade.Logger
}

// EndUserContextKey context key 定义
const EndUserContextKey = "end_user_identity"

func (m *RuntimeAuthMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. 提取 JWT
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            m.logger.Warn("Missing Authorization header")
            http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
            return
        }
        
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            m.logger.Warn("Invalid Authorization header format")
            http.Error(w, "Unauthorized: Invalid token format", http.StatusUnauthorized)
            return
        }
        token := parts[1]
        
        // 2. 解析并验证 JWT
        claims, err := m.jwtValidator.Validate(token)
        if err != nil {
            m.logger.Warn("Invalid JWT", logfacade.Err(err))
            http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
            return
        }
        
        // 3. 校验 Issuer 必须是 mc-enduser
        issuer, ok := claims["iss"].(string)
        if !ok || issuer != "mc-enduser" {
            m.logger.Warn("Invalid JWT issuer", logfacade.Str("issuer", issuer))
            http.Error(w, "Unauthorized: Invalid issuer", http.StatusUnauthorized)
            return
        }
        
        // 4. 提取 endUserId
        endUserID, ok := claims["user_id"].(string)
        if !ok || endUserID == "" {
            m.logger.Warn("Missing user_id in JWT claims")
            http.Error(w, "Unauthorized: Invalid token claims", http.StatusUnauthorized)
            return
        }
        
        // 5. 注入 context
        identity := &rls.EndUserIdentity{
            EndUserID: endUserID,
            Issuer:    issuer,
        }
        ctx := context.WithValue(r.Context(), EndUserContextKey, identity)
        
        m.logger.Debug("EndUser authenticated", 
            logfacade.Str("endUserId", endUserID))
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 4.3 Runtime Resolver RLS 注入

**文件**: `internal/interfaces/graphql/resolver/runtime/runtime_query_resolver.go`

```go
package runtime

import (
    "context"
    "modelcraft/internal/domain/rls"
)

// FindMany 查询多条记录（自动注入 WHERE）
func (r *runtimeResolver) FindMany(ctx context.Context, modelName string, 
    filter *runtime.WhereInput, pagination *runtime.PaginationInput) (*runtime.ModelConnection, error) {
    
    // 1. 获取 EndUser 身份（来自 context）
    identity := r.getEndUserIdentity(ctx)
    if identity == nil {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.Unauthorized, 
            "EndUser identity required")
    }
    
    // 2. 获取 Model 和 Policy
    model, err := r.getModelWithFields(ctx, modelName)
    if err != nil {
        return nil, err
    }
    
    // 3. 解析 RLSFilter
    rlsFilter, err := r.rlsResolver.Resolve(ctx, identity, model)
    if err != nil {
        return nil, err
    }
    
    // 4. DENY ALL 检查
    if rlsFilter == nil || rlsFilter.IsDenyAll() {
        // 返回空结果，不报错
        return &runtime.ModelConnection{Edges: []runtime.ModelEdge{}}, nil
    }
    
    // 5. 编译 RLS WHERE 条件
    var rlsWhere *runtime.WhereInput
    if rlsFilter.ShouldInjectWhere() {
        rlsWhere, err = r.compileRLSWhere(rlsFilter.SelectPredicate, identity.EndUserID)
        if err != nil {
            return nil, err
        }
    }
    
    // 6. 合并用户 filter 和 RLS filter（AND 关系）
    mergedFilter := r.mergeFilters(filter, rlsWhere)
    
    // 7. 执行查询
    return r.executeFindMany(ctx, model, mergedFilter, pagination)
}

// FindFirst 查询单条记录（自动注入 WHERE）
func (r *runtimeResolver) FindFirst(ctx context.Context, modelName string, 
    filter *runtime.WhereInput) (*runtime.Model, error) {
    
    // 逻辑与 FindMany 类似，但返回单条
    // ...
}

// compileRLSWhere 将 JsonExpr 编译为 WhereInput
func (r *runtimeResolver) compileRLSWhere(expr modeldesign.JsonExpr, endUserID string) (*runtime.WhereInput, error) {
    // 简单情况：{"owner":{"_eq":{"_auth":"uid"}}} → where: { owner: { equals: endUserID } }
    if expr.IsOwnerEqualsUser() {
        return &runtime.WhereInput{
            Owner: &runtime.StringFilter{Equals: &endUserID},
        }, nil
    }
    
    // 复杂情况：调用 PolicyCompiler 和 PolicyExecutor
    compiled, err := r.policyCompiler.Compile(ctx, expr)
    if err != nil {
        return nil, err
    }
    
    // 转换为 WhereInput
    // ...
    return nil, nil
}

// mergeFilters 合并两个 filter（AND 关系）
func (r *runtimeResolver) mergeFilters(userFilter, rlsFilter *runtime.WhereInput) *runtime.WhereInput {
    if userFilter == nil {
        return rlsFilter
    }
    if rlsFilter == nil {
        return userFilter
    }
    
    // 使用 _and 合并
    return &runtime.WhereInput{
        And: []*runtime.WhereInput{userFilter, rlsFilter},
    }
}
```

**文件**: `internal/interfaces/graphql/resolver/runtime/runtime_mutation_resolver.go`

```go
package runtime

import (
    "context"
    "modelcraft/internal/domain/rls"
)

// CreateOne 创建单条记录（自动填充 owner）
func (r *runtimeResolver) CreateOne(ctx context.Context, modelName string, 
    input map[string]interface{}) (*runtime.Model, error) {
    
    // 1. 获取 EndUser 身份
    identity := r.getEndUserIdentity(ctx)
    if identity == nil {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.Unauthorized, 
            "EndUser identity required")
    }
    
    // 2. 获取 Model 和 Policy
    model, err := r.getModelWithFields(ctx, modelName)
    if err != nil {
        return nil, err
    }
    
    // 3. 解析 RLSFilter
    rlsFilter, err := r.rlsResolver.Resolve(ctx, identity, model)
    if err != nil {
        return nil, err
    }
    
    // 4. DENY ALL 检查
    if rlsFilter == nil || rlsFilter.IsDenyAll() {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.PermissionDeniedRLS, modelName)
    }
    
    // 5. 检查 insertCheck
    if !rlsFilter.InsertCheck.IsTrue() {
        // 应用层校验
        if rlsFilter.InsertCheck.IsFalse() {
            return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.RLSCheckViolation, "INSERT")
        }
        
        // 校验 owner = endUserId
        if rlsFilter.InsertCheck.IsOwnerEqualsUser() {
            // 强制覆盖 owner 字段
            input["owner"] = identity.EndUserID
        } else {
            // 复杂表达式校验
            if err := r.policyExecutor.ValidateCheck(ctx, rlsFilter.InsertCheck, input, 
                &rls.AuthContext{EndUserID: identity.EndUserID}); err != nil {
                return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.RLSCheckViolation, "INSERT")
            }
        }
    }
    
    // 6. 执行创建
    return r.executeCreateOne(ctx, model, input)
}

// UpdateOne 更新单条记录（自动注入 WHERE + CHECK 校验）
func (r *runtimeResolver) UpdateOne(ctx context.Context, modelName string, id string, 
    input map[string]interface{}) (*runtime.Model, error) {
    
    // 1. 获取 EndUser 身份
    identity := r.getEndUserIdentity(ctx)
    
    // 2. 获取 Model 和 Policy
    model, err := r.getModelWithFields(ctx, modelName)
    if err != nil {
        return nil, err
    }
    
    // 3. 解析 RLSFilter
    rlsFilter, err := r.rlsResolver.Resolve(ctx, identity, model)
    if err != nil {
        return nil, err
    }
    
    // 4. DENY ALL 检查
    if rlsFilter == nil || rlsFilter.IsDenyAll() {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.PermissionDeniedRLS, modelName)
    }
    
    // 5. USING 过滤：先查询记录是否存在且符合 updatePredicate
    if !rlsFilter.UpdatePredicate.IsTrue() {
        // 注入 WHERE 查询目标记录
        exists, err := r.checkRecordExistsWithRLS(ctx, model, id, rlsFilter.UpdatePredicate)
        if err != nil {
            return nil, err
        }
        if !exists {
            // 静默返回：记录不存在（符合预期，不报错）
            return nil, nil
        }
    }
    
    // 6. WITH CHECK 校验
    if !rlsFilter.UpdateCheck.IsTrue() {
        if rlsFilter.UpdateCheck.IsFalse() {
            return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.RLSCheckViolation, "UPDATE")
        }
        
        // 校验 owner = endUserId
        if rlsFilter.UpdateCheck.IsOwnerEqualsUser() {
            // 禁止更新 owner 为其他值
            if newOwner, ok := input["owner"].(string); ok && newOwner != identity.EndUserID {
                return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.RLSCheckViolation, "UPDATE: cannot change owner")
            }
        }
    }
    
    // 7. 执行更新
    return r.executeUpdateOne(ctx, model, id, input)
}

// DeleteOne 删除单条记录（自动注入 WHERE）
func (r *runtimeResolver) DeleteOne(ctx context.Context, modelName string, id string) (bool, error) {
    // 逻辑类似 UpdateOne，使用 deletePredicate
    // USING 不通过 → 静默返回 false（不报错）
    // ...
}
```

---

## 5. 数据库 Migration

### 5.1 新建 SQL 文件

**文件**: `modelcraft-backend/db/schema/mysql/11_rls.sql`

```sql
-- ============================================
-- RLS (Row Level Security) Schema
-- ============================================

-- model_rls_policies 表
-- 存储 Model 的 RLS 策略配置（五件套 JsonExpr）
CREATE TABLE IF NOT EXISTS `model_rls_policies` (
    -- 主键字段
    `model_id` VARCHAR(36) NOT NULL COMMENT '模型 ID（与 models 表 1:1）',
    
    -- 租户隔离字段
    `org_name` VARCHAR(36) NOT NULL COMMENT '所属组织名称',
    `project_slug` VARCHAR(64) NOT NULL COMMENT '所属项目标识符',
    
    -- RLS 策略字段（五件套）
    `select_predicate` JSON NOT NULL COMMENT 'SELECT USING 谓词（JSON 表达式）',
    `insert_check` JSON NOT NULL COMMENT 'INSERT WITH CHECK 谓词（JSON 表达式）',
    `update_predicate` JSON NOT NULL COMMENT 'UPDATE USING 谓词（JSON 表达式）',
    `update_check` JSON NOT NULL COMMENT 'UPDATE WITH CHECK 谓词（JSON 表达式）',
    `delete_predicate` JSON NOT NULL COMMENT 'DELETE USING 谓词（JSON 表达式）',
    
    -- 时间戳字段
    `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    
    -- 主键约束
    PRIMARY KEY (`model_id`),
    
    -- 外键约束
    CONSTRAINT `fk_rls_policy_model` FOREIGN KEY (`model_id`) REFERENCES `models` (`id`) ON DELETE CASCADE,
    
    -- 索引
    KEY `idx_rls_policies_project` (`org_name`, `project_slug`) COMMENT '项目查询索引'
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型 RLS 策略表';


-- project_auth_schemas 表
-- 存储 Project 级别的认证变量配置（用于 _auth 引用）
CREATE TABLE IF NOT EXISTS `project_auth_schemas` (
    -- 租户隔离字段（复合主键）
    `org_name` VARCHAR(36) NOT NULL COMMENT '所属组织名称',
    `project_slug` VARCHAR(64) NOT NULL COMMENT '所属项目标识符',
    
    -- 认证变量配置
    `variables` JSON NOT NULL COMMENT '认证变量列表（JSON 数组）',
    
    -- 时间戳字段
    `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    
    -- 主键约束（复合主键）
    PRIMARY KEY (`org_name`, `project_slug`),
    
    -- 索引
    KEY `idx_auth_schemas_project` (`org_name`, `project_slug`) COMMENT '项目查询索引'
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目认证变量配置表';


-- 为现有 Model 添加 created_via 字段（区分新建 vs 导入）
-- 注意：需要在 models 表添加字段
-- ALTER TABLE `models` ADD COLUMN `created_via` VARCHAR(20) NOT NULL DEFAULT 'IMPORTED' 
--     COMMENT '模型创建来源：NEW/IMPORTED' AFTER `deployment_status`;

-- 注意：END_USER_REF 字段 format 在应用层处理，DB 层 format 列已支持字符串存储
-- 外键约束由应用层在创建/更新字段时动态生成 DDL
```

### 5.2 在 models 表添加 created_via 字段

**修改**: `modelcraft-backend/db/schema/mysql/03_model_domain.sql`

```sql
-- 在 models 表添加 created_via 字段（用于区分新建 Model 和导入 Model）
ALTER TABLE `models` ADD COLUMN IF NOT EXISTS `created_via` 
    VARCHAR(20) NOT NULL DEFAULT 'IMPORTED' 
    COMMENT '模型创建来源：NEW/IMPORTED' 
    AFTER `deployment_status`;
```

---

## 6. 实现顺序

### Phase 1: 基础数据模型（Week 1）

| 优先级 | 任务 | 依赖 | 输出 |
|--------|------|------|------|
| 1 | 数据库 Migration | - | `11_rls.sql`, `03_model_domain.sql` 修改 |
| 2 | FormatType 扩展（END_USER_REF） | Migration | `field_definition.go` 修改 |
| 3 | ModelCreationSource 枚举 | - | `model_creation_source.go` |
| 4 | Model 添加 createdVia 字段 | Migration | `model.go` 扩展 |
| 5 | RLSPreset 枚举 | - | `rls_preset.go` |
| 6 | JsonExpr 类型 | - | `json_expr.go` |
| 7 | ModelRLSPolicy 实体 | JsonExpr | `model_rls_policy.go` |
| 8 | Repository 接口定义 | Policy 实体 | `model_rls_repository.go` |
| 9 | sqlc 查询 + Repository 实现 | 接口定义 | `model_rls_policy.sql`, `model_rls_policy_repo.go` |

### Phase 2: 领域服务（Week 1-2）

| 优先级 | 任务 | 依赖 | 输出 |
|--------|------|------|------|
| 10 | AuthVariable / AuthSchema VO | - | `project_auth_schema.go` |
| 11 | AuthSchemaRepository 接口 + 实现 | AuthSchema VO | `auth_schema_repository.go`, `auth_schema_repo.go` |
| 12 | EndUserIdentity VO | - | `end_user_identity.go` |
| 13 | RLSFilter VO | JsonExpr | `rls_filter.go` |
| 14 | PolicyValidator 接口 + 实现 | RLSFilter | `policy_validator.go`, `policy_validator_impl.go` |
| 15 | PolicyCompiler 接口 | - | `policy_compiler.go` |
| 16 | PolicyExecutor 接口 | - | `policy_executor.go` |
| 17 | RLSResolver 接口 | RLSFilter | `rls_resolver.go` |

### Phase 3: App 层服务（Week 2）

| 优先级 | 任务 | 依赖 | 输出 |
|--------|------|------|------|
| 18 | ModelRLSPolicyAppService | Validator, Repository | `model_rls_policy_service.go` |
| 19 | AuthSchemaAppService | Repository | `auth_schema_service.go` |
| 20 | 新建 Model 自动创建 owner + Policy | ModelCreationSource | `model_app_service.go` 扩展 |
| 21 | EndUserRef 唯一性检查 | FormatType 扩展 | `field_app_service.go` 扩展 |

### Phase 4: Runtime 层（Week 3）

| 优先级 | 任务 | 依赖 | 输出 |
|--------|------|------|------|
| 22 | Issuer 枚举扩展 | - | `issuer.go` 扩展 |
| 23 | Runtime Auth Middleware | Issuer | `runtime_auth_middleware.go` |
| 24 | RLSResolver 实现 | RLSFilter | `rls_resolver_impl.go` |
| 25 | PolicyCompiler 实现（基础） | - | `policy_compiler_impl.go` |
| 26 | PolicyExecutor 实现（基础） | - | `policy_executor_impl.go` |
| 27 | Runtime Query WHERE 注入 | Resolver, Compiler | `runtime_query_resolver.go` 扩展 |
| 28 | Runtime Mutation owner 填充 | Resolver, Executor | `runtime_mutation_resolver.go` 扩展 |

### Phase 5: GraphQL Schema + Resolver（Week 3-4）

| 优先级 | 任务 | 依赖 | 输出 |
|--------|------|------|------|
| 29 | GraphQL Schema 定义 | - | `rls.graphql`, `project.graphql` 扩展 |
| 30 | gqlgen 代码生成 | Schema | `generated/` 更新 |
| 31 | RLS Resolver 实现 | App Service | `rls_resolver.go` |
| 32 | Project Resolver 扩展 | AuthSchemaAppService | `project_resolver.go` 扩展 |
| 33 | Model Resolver 扩展 | ModelRLSPolicyAppService | `model_resolver.go` 扩展 |

### Phase 6: 错误码 + 集成测试（Week 4）

| 优先级 | 任务 | 依赖 | 输出 |
|--------|------|------|------|
| 34 | 错误码定义 | - | `common_errors.go` 扩展 |
| 35 | Error Adapter 扩展 | 错误码 | `error_adapter.go` 扩展 |
| 36 | BDD 测试场景 | 全部 | `tests-bdd/features/rls/*.feature` |
| 37 | 集成测试 | 全部 | 端到端验证 |

### 依赖关系图

```
Migration (1)
    ↓
Domain 基础类型 (2-7)
    ↓
Repository 接口 + 实现 (8-9, 11)
    ↓
Domain Services (10-17)
    ↓
App Services (18-21)
    ↓
Runtime 层 (22-28)
    ↓
GraphQL 层 (29-33)
    ↓
测试 (34-37)
```

### 关键路径

1. **最短可用路径**（MVP）：1 → 2 → 8 → 9 → 18 → 22 → 23 → 27 → 29 → 31
   - 实现基础 RLS 配置和 Runtime WHERE 注入

2. **完整功能路径**：MVP + 10 → 11 → 14 → 19 → 28 → 32
   - 添加 auth_schema、完整校验、Mutation 支持

3. **向后兼容路径**：在 20 中确保 `createdVia=IMPORTED` 的 Model 行为不变
