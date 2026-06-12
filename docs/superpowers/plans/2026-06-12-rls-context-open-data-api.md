# RLS Context + 开放数据 API 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 role-based RLS policy matching + Prisma 风格表达式编译 + Header context 注入 + Python SDK

**Architecture:** 在现有 `domain/rls/` + `app/rls/` + `interfaces/runtime/` 三层上扩展。新增多策略存储（每个 model 多条 policy，按 action+role 匹配），扩展现有 PolicyCompiler 支持 Prisma 操作符和 `{{user_id}}`/`{{user_name}}` 变量，新增 RLSContextMiddleware 从 Header 提取用户上下文。

**Tech Stack:** Go (gqlgen + sqlc), MySQL, Python (SDK)

---

## 文件结构

### 新建文件

| 文件 | 职责 |
|------|------|
| `db/schema/mysql/XX_rls_policy_v2.sql` | 新 policy 表 migration |
| `db/queries/rls_policy_v2.sql` | sqlc queries |
| `internal/domain/rls/policy.go` | 单条 policy 实体 + using/withCheck 表达式 |
| `internal/domain/rls/policy_repository.go` | Repository 接口（多策略查询） |
| `internal/domain/rls/user_context.go` | UserContext（Header 提取的上下文） |
| `internal/app/rls/policy_matching_service.go` | 策略匹配 + OR 合并 |
| `internal/interfaces/http/middleware/rls_context_middleware.go` | X-MC-User-* Header → context |
| `internal/interfaces/http/middleware/rls_context_middleware_test.go` | 测试 |
| `modelcraft-sdk-python/modelcraft/client.py` | Python SDK client |
| `modelcraft-sdk-python/modelcraft/__init__.py` | SDK 入口 |

### 修改文件

| 文件 | 改动 |
|------|------|
| `internal/app/rls/policy_compiler.go` | 新增 Prisma 操作符 + 变量支持 |
| `internal/app/rls/policy_executor.go` | 扩展 resolveAuthVar 支持 user_id/user_name |
| `internal/domain/rls/policy_compiler.go` | 更新 Compile 接口签名（允许传 UserContext） |
| `internal/interfaces/runtime/rls_resolver.go` | 替换为多策略匹配引擎，集成 PolicyCompiler 做完整表达式编译 |
| `internal/interfaces/runtime/handler.go` | 挂载 RLSContextMiddleware |
| `internal/domain/rls/end_user_identity.go` | 新增 UserContext 字段 |

### 删除/淘汰

| 文件 | 说明 |
|------|------|
| `internal/domain/modeldesign/model_rls_policy.go` | 旧五件套实体，替换为 `rls.Policy` |
| `db/schema/mysql/11_rls.sql` → 对应 migration | 旧 `model_rls_policies` 表结构，新增 migration 替换 |

---

### Task 1: DB Migration — 新策略存储

**Files:**
- Create: `modelcraft-backend/db/schema/mysql/12_rls_policy_v2.sql`
- Create: `modelcraft-backend/db/queries/rls_policy_v2.sql`

- [ ] **Step 1: 编写 migration SQL**

```sql
-- db/schema/mysql/12_rls_policy_v2.sql
-- Migration: RLS Policy V2 — 多策略存储（role + action + 表达式）

DROP TABLE IF EXISTS model_rls_policies;

CREATE TABLE model_rls_policies (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    org_name        VARCHAR(255) NOT NULL COMMENT '组织名',
    project_slug    VARCHAR(255) NOT NULL COMMENT '项目标识',
    model_id        VARCHAR(36) NOT NULL COMMENT '模型 ID',

    -- 策略元信息
    policy_name     VARCHAR(255) NOT NULL COMMENT '策略名称（model 内唯一）',
    action          ENUM('read', 'create', 'update', 'delete') NOT NULL COMMENT '操作类型',
    role            VARCHAR(255) NOT NULL DEFAULT '' COMMENT '匹配角色（空=默认策略）',

    -- 表达式（Prisma 风格 JSON）
    using_expr      JSON COMMENT 'USING 表达式（read/update/delete）',
    with_check_expr JSON COMMENT 'WITH CHECK 表达式（create/update）',

    created_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    -- 唯一约束：同一 model 下 policy_name 唯一
    UNIQUE KEY uk_policy_name (org_name, project_slug, model_id, policy_name),

    -- 查询索引：按 action + role 匹配
    INDEX idx_policy_match (org_name, project_slug, model_id, action, role)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='RLS 策略表（多策略存储）';
```

- [ ] **Step 2: 编写 sqlc queries**

```sql
-- db/queries/rls_policy_v2.sql
-- name: ListPoliciesByAction :many
SELECT * FROM model_rls_policies
WHERE org_name = ? AND project_slug = ? AND model_id = ? AND action = ?
  AND role IN (sqlc.slice('roles'));

-- name: ListPoliciesByModel :many
SELECT * FROM model_rls_policies
WHERE org_name = ? AND project_slug = ? AND model_id = ?
ORDER BY action, role;

-- name: UpsertPolicy :exec
INSERT INTO model_rls_policies (
    org_name, project_slug, model_id, policy_name, action, role, using_expr, with_check_expr
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    using_expr = VALUES(using_expr),
    with_check_expr = VALUES(with_check_expr);

-- name: DeletePolicy :exec
DELETE FROM model_rls_policies
WHERE id = ? AND org_name = ? AND project_slug = ?;

-- name: DeletePoliciesByModel :exec
DELETE FROM model_rls_policies
WHERE org_name = ? AND project_slug = ? AND model_id = ?;

-- name: PolicyExists :one
SELECT EXISTS(
    SELECT 1 FROM model_rls_policies
    WHERE org_name = ? AND project_slug = ? AND model_id = ? AND action = ? AND role = ?
) AS exists_flag;
```

- [ ] **Step 3: 运行 sqlc 代码生成**

```bash
cd modelcraft-backend && sqlc generate
```

- [ ] **Step 4: 运行 migration**

```bash
cd modelcraft-backend && atlas migrate diff --env local
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/db/schema/mysql/12_rls_policy_v2.sql \
        modelcraft-backend/db/queries/rls_policy_v2.sql \
        modelcraft-backend/internal/infrastructure/persistence/rls/
git commit -m "feat(rls): add v2 policy storage with role+action matching schema"
```

---

### Task 2: Domain Entity — Policy + UserContext

**Files:**
- Create: `modelcraft-backend/internal/domain/rls/policy.go`
- Create: `modelcraft-backend/internal/domain/rls/user_context.go`
- Modify: `modelcraft-backend/internal/domain/rls/policy_compiler.go`

- [ ] **Step 1: 定义 Policy 实体**

```go
// internal/domain/rls/policy.go
package rls

import "time"

// Action 操作类型
type Action string

const (
    ActionRead   Action = "read"
    ActionCreate Action = "create"
    ActionUpdate Action = "update"
    ActionDelete Action = "delete"
)

// Policy 单条 RLS 策略
type Policy struct {
    ID           int64       `json:"id"`
    OrgName      string      `json:"orgName"`
    ProjectSlug  string      `json:"projectSlug"`
    ModelID      string      `json:"modelId"`
    PolicyName   string      `json:"policyName"`
    Action       Action      `json:"action"`
    Role         string      `json:"role"`
    UsingExpr    JsonExpr    `json:"usingExpr"`
    WithCheckExpr JsonExpr   `json:"withCheckExpr"`
    CreatedAt    time.Time   `json:"createdAt"`
    UpdatedAt    time.Time   `json:"updatedAt"`
}
```

- [ ] **Step 2: 定义 UserContext（Header 提取的上下文）**

```go
// internal/domain/rls/user_context.go
package rls

import "strings"

// UserContext 从 Header 提取的用户上下文
type UserContext struct {
    UserID   string   `json:"userId"`
    UserName string   `json:"userName"`
    Roles    []string `json:"roles"`
}

// HasRole 判断是否包含指定 role
func (uc *UserContext) HasRole(role string) bool {
    // role="" 总是匹配（默认策略）
    if role == "" {
        return true
    }
    for _, r := range uc.Roles {
        if strings.TrimSpace(r) == role {
            return true
        }
    }
    return false
}

// ResolveVariable 解析表达式变量
func (uc *UserContext) ResolveVariable(name string) string {
    switch name {
    case "user_id":
        return uc.UserID
    case "user_name":
        return uc.UserName
    default:
        return ""
    }
}
```

- [ ] **Step 3: 更新 PolicyCompiler 接口签名**

在 `internal/domain/rls/policy_compiler.go` 中更新 `Compile` 方法签名，接收 `UserContext`：

```go
// 替换原 Compile 签名
type PolicyCompiler interface {
    // Compile 将 expression 和 UserContext 编译为参数化 SQL
    Compile(ctx context.Context, expr JsonExpr, userCtx *UserContext) (*CompiledPolicy, error)
}
```

- [ ] **Step 4: 运行现有测试确保无回归**

```bash
cd modelcraft-backend && go test ./internal/domain/rls/... -v
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/domain/rls/
git commit -m "feat(rls): add Policy entity, UserContext, and updated Compiler interface"
```

---

### Task 3: Expression Compiler — Prisma 操作符 + 变量扩展

**Files:**
- Modify: `modelcraft-backend/internal/app/rls/policy_compiler.go`
- Create: `modelcraft-backend/internal/app/rls/policy_compiler_test.go`

- [ ] **Step 1: 编写测试 — equals + 变量替换**

```go
// internal/app/rls/policy_compiler_test.go
package rls

import (
    "context"
    "modelcraft/internal/domain/rls"
    "testing"
)

func TestCompile_Equals_VariableSubstitution(t *testing.T) {
    compiler := NewPolicyCompiler()
    ctx := context.Background()

    expr := rls.JsonExpr(`{"tenant_id": {"equals": "{{user_id}}"}}`)
    userCtx := &rls.UserContext{UserID: "customer_123"}

    result, err := compiler.Compile(ctx, expr, userCtx)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    expected := "tenant_id = ?"
    if result.SQL != expected {
        t.Errorf("expected SQL %q, got %q", expected, result.SQL)
    }
    if len(result.Params) != 1 || result.Params[0] != "customer_123" {
        t.Errorf("expected params [customer_123], got %v", result.Params)
    }
}

func TestCompile_AND_With_Two_Fields(t *testing.T) {
    compiler := NewPolicyCompiler()
    ctx := context.Background()

    expr := rls.JsonExpr(`{
        "AND": [
            {"tenant_id": {"equals": "{{user_id}}"}},
            {"status": {"equals": "active"}}
        ]
    }`)
    userCtx := &rls.UserContext{UserID: "123"}

    result, err := compiler.Compile(ctx, expr, userCtx)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    expected := "(tenant_id = ?) AND (status = ?)"
    if result.SQL != expected {
        t.Errorf("expected SQL %q, got %q", expected, result.SQL)
    }
    if len(result.Params) != 2 {
        t.Errorf("expected 2 params, got %d", len(result.Params))
    }
}

func TestCompile_Contains(t *testing.T) {
    compiler := NewPolicyCompiler()
    ctx := context.Background()

    expr := rls.JsonExpr(`{"name": {"contains": "{{user_name}}"}}`)
    userCtx := &rls.UserContext{UserName: "zhangsan"}

    result, err := compiler.Compile(ctx, expr, userCtx)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if result.SQL != "name LIKE ?" {
        t.Errorf("expected 'name LIKE ?', got %q", result.SQL)
    }
    if len(result.Params) != 1 || result.Params[0] != "%zhangsan%" {
        t.Errorf("expected params [%%zhangsan%%], got %v", result.Params)
    }
}
```

- [ ] **Step 2: 运行测试确保失败**

```bash
cd modelcraft-backend && go test ./internal/app/rls/ -run "TestCompile_" -v
```

- [ ] **Step 3: 重写 PolicyCompiler，支持 Prisma 操作符**

```go
// internal/app/rls/policy_compiler.go（核心改动）
// 保持原有 _and/_or/_not/_eq 支持，新增 equals/not/in/gt/gte/lt/lte/contains/startsWith/endsWith/AND/OR/NOT

func (c *PolicyCompiler) Compile(ctx context.Context, expr rls.JsonExpr, userCtx *rls.UserContext) (*rls.CompiledPolicy, error) {
    var root interface{}
    if err := json.Unmarshal([]byte(expr), &root); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }

    if b, ok := root.(bool); ok {
        if b {
            return &rls.CompiledPolicy{SQL: "1=1", Params: nil}, nil
        }
        return &rls.CompiledPolicy{SQL: "1=0", Params: nil}, nil
    }

    obj, ok := root.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("expression must be an object or boolean constant")
    }

    return c.compileNode(ctx, obj, userCtx)
}

func (c *PolicyCompiler) compileNode(ctx context.Context, node map[string]interface{}, userCtx *rls.UserContext) (*rls.CompiledPolicy, error) {
    var conditions []string
    var params []interface{}

    for key, value := range node {
        switch c.normalizeOp(key) {
        case "AND":
            arr, ok := value.([]interface{})
            if !ok {
                return nil, fmt.Errorf("AND value must be an array")
            }
            var andConds []string
            for _, item := range arr {
                itemObj, ok := item.(map[string]interface{})
                if !ok {
                    return nil, fmt.Errorf("AND array item must be an object")
                }
                compiled, err := c.compileNode(ctx, itemObj, userCtx)
                if err != nil {
                    return nil, err
                }
                andConds = append(andConds, "("+compiled.SQL+")")
                params = append(params, compiled.Params...)
            }
            if len(andConds) > 0 {
                conditions = append(conditions, "("+strings.Join(andConds, " AND ")+")")
            }

        case "OR":
            arr, ok := value.([]interface{})
            if !ok {
                return nil, fmt.Errorf("OR value must be an array")
            }
            var orConds []string
            for _, item := range arr {
                itemObj, ok := item.(map[string]interface{})
                if !ok {
                    return nil, fmt.Errorf("OR array item must be an object")
                }
                compiled, err := c.compileNode(ctx, itemObj, userCtx)
                if err != nil {
                    return nil, err
                }
                orConds = append(orConds, "("+compiled.SQL+")")
                params = append(params, compiled.Params...)
            }
            if len(orConds) > 0 {
                conditions = append(conditions, "("+strings.Join(orConds, " OR ")+")")
            }

        case "NOT":
            obj, ok := value.(map[string]interface{})
            if !ok {
                return nil, fmt.Errorf("NOT value must be an object")
            }
            compiled, err := c.compileNode(ctx, obj, userCtx)
            if err != nil {
                return nil, err
            }
            conditions = append(conditions, "NOT ("+compiled.SQL+")")
            params = append(params, compiled.Params...)

        default:
            // 字段比较
            compObj, ok := value.(map[string]interface{})
            if !ok {
                // 尝试简单值比较（true/false 常量）
                if b, ok := value.(bool); ok {
                    if b {
                        conditions = append(conditions, "1=1")
                    } else {
                        conditions = append(conditions, "1=0")
                    }
                    continue
                }
                return nil, fmt.Errorf("field comparison must be an object for field %q", key)
            }
            fieldCond, fieldParams, err := c.compileFieldComparison(key, compObj, userCtx)
            if err != nil {
                return nil, err
            }
            conditions = append(conditions, fieldCond)
            params = append(params, fieldParams...)
        }
    }

    if len(conditions) == 0 {
        return &rls.CompiledPolicy{SQL: "1=1", Params: nil}, nil
    }

    return &rls.CompiledPolicy{
        SQL:    strings.Join(conditions, " AND "),
        Params: params,
    }, nil
}

// normalizeOp 统一操作符名（兼容旧 _eq 和新 equals）
func (c *PolicyCompiler) normalizeOp(op string) string {
    switch op {
    case "_and":
        return "AND"
    case "_or":
        return "OR"
    case "_not":
        return "NOT"
    default:
        return op
    }
}

func (c *PolicyCompiler) compileFieldComparison(fieldName string, compObj map[string]interface{}, userCtx *rls.UserContext) (string, []interface{}, error) {
    var conditions []string
    var params []interface{}

    for op, value := range compObj {
        switch c.normalizeFieldOp(op) {
        case "equals":
            resolved := c.resolveValue(value, userCtx)
            conditions = append(conditions, fmt.Sprintf("%s = ?", fieldName))
            params = append(params, resolved)
        case "not":
            resolved := c.resolveValue(value, userCtx)
            conditions = append(conditions, fmt.Sprintf("%s != ?", fieldName))
            params = append(params, resolved)
        case "gt":
            conditions = append(conditions, fmt.Sprintf("%s > ?", fieldName))
            params = append(params, c.resolveValue(value, userCtx))
        case "gte":
            conditions = append(conditions, fmt.Sprintf("%s >= ?", fieldName))
            params = append(params, c.resolveValue(value, userCtx))
        case "lt":
            conditions = append(conditions, fmt.Sprintf("%s < ?", fieldName))
            params = append(params, c.resolveValue(value, userCtx))
        case "lte":
            conditions = append(conditions, fmt.Sprintf("%s <= ?", fieldName))
            params = append(params, c.resolveValue(value, userCtx))
        case "in":
            arr, ok := value.([]interface{})
            if !ok {
                return "", nil, fmt.Errorf("in value must be an array")
            }
            placeholders := make([]string, len(arr))
            for i := range arr {
                placeholders[i] = "?"
                params = append(params, c.resolveValue(arr[i], userCtx))
            }
            conditions = append(conditions, fmt.Sprintf("%s IN (%s)", fieldName, strings.Join(placeholders, ", ")))
        case "contains":
            resolved := c.resolveValue(value, userCtx)
            conditions = append(conditions, fmt.Sprintf("%s LIKE ?", fieldName))
            params = append(params, "%"+fmt.Sprint(resolved)+"%")
        case "startsWith":
            resolved := c.resolveValue(value, userCtx)
            conditions = append(conditions, fmt.Sprintf("%s LIKE ?", fieldName))
            params = append(params, fmt.Sprint(resolved)+"%")
        case "endsWith":
            resolved := c.resolveValue(value, userCtx)
            conditions = append(conditions, fmt.Sprintf("%s LIKE ?", fieldName))
            params = append(params, "%"+fmt.Sprint(resolved))
        default:
            return "", nil, fmt.Errorf("unknown field operator: %s", op)
        }
    }

    if len(conditions) == 0 {
        return "1=1", nil, nil
    }

    return strings.Join(conditions, " AND "), params, nil
}

// normalizeFieldOp 统一字段操作符名
func (c *PolicyCompiler) normalizeFieldOp(op string) string {
    switch op {
    case "_eq":
        return "equals"
    case "_neq":
        return "not"
    case "_gt":
        return "gt"
    case "_gte":
        return "gte"
    case "_lt":
        return "lt"
    case "_lte":
        return "lte"
    case "_in":
        return "in"
    case "_is_null":
        return "isNull"
    default:
        return op
    }
}

// resolveValue 解析值中的变量占位符
func (c *PolicyCompiler) resolveValue(value interface{}, userCtx *rls.UserContext) interface{} {
    if userCtx == nil {
        return value
    }
    switch v := value.(type) {
    case string:
        // {{user_id}} / {{user_name}} 替换
        v = strings.ReplaceAll(v, "{{user_id}}", userCtx.UserID)
        v = strings.ReplaceAll(v, "{{user_name}}", userCtx.UserName)
        return v
    case map[string]interface{}:
        // 兼容旧 _auth 引用
        if authVar, ok := v["_auth"].(string); ok {
            return userCtx.ResolveVariable(authVar)
        }
        return v
    default:
        return v
    }
}
```

- [ ] **Step 4: 运行测试确保通过**

```bash
cd modelcraft-backend && go test ./internal/app/rls/ -run "TestCompile_" -v
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/app/rls/
git commit -m "feat(rls): add Prisma-style operators and variable substitution to PolicyCompiler"
```

---

### Task 4: RLS Context Middleware

**Files:**
- Create: `modelcraft-backend/internal/interfaces/http/middleware/rls_context_middleware.go`
- Create: `modelcraft-backend/internal/interfaces/http/middleware/rls_context_middleware_test.go`

- [ ] **Step 1: 编写测试**

```go
// internal/interfaces/http/middleware/rls_context_middleware_test.go
package middleware

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestRLSContextMiddleware_AllHeaders(t *testing.T) {
    var capturedCtx context.Context
    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        capturedCtx = r.Context()
        w.WriteHeader(200)
    })

    mw := NewRLSContextMiddleware()
    handler := mw.Middleware(next)

    req := httptest.NewRequest("POST", "/api/data", nil)
    req.Header.Set("X-MC-User-ID", "user_123")
    req.Header.Set("X-MC-User-Name", "zhangsan")
    req.Header.Set("X-MC-User-Roles", "admin, manager")

    handler.ServeHTTP(httptest.NewRecorder(), req)

    uc := GetUserContext(capturedCtx)
    if uc == nil {
        t.Fatal("expected UserContext in context, got nil")
    }
    if uc.UserID != "user_123" {
        t.Errorf("expected UserID 'user_123', got %q", uc.UserID)
    }
    if uc.UserName != "zhangsan" {
        t.Errorf("expected UserName 'zhangsan', got %q", uc.UserName)
    }
    if len(uc.Roles) != 2 || uc.Roles[0] != "admin" || uc.Roles[1] != "manager" {
        t.Errorf("expected Roles [admin, manager], got %v", uc.Roles)
    }
}

func TestRLSContextMiddleware_NoHeaders(t *testing.T) {
    var capturedCtx context.Context
    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        capturedCtx = r.Context()
    })

    mw := NewRLSContextMiddleware()
    handler := mw.Middleware(next)

    req := httptest.NewRequest("POST", "/api/data", nil)
    handler.ServeHTTP(httptest.NewRecorder(), req)

    uc := GetUserContext(capturedCtx)
    if uc == nil {
        t.Fatal("expected UserContext in context, got nil")
    }
    if uc.UserID != "" {
        t.Errorf("expected empty UserID, got %q", uc.UserID)
    }
    if len(uc.Roles) != 0 {
        t.Errorf("expected empty Roles, got %v", uc.Roles)
    }
}
```

- [ ] **Step 2: 实现 Middleware**

```go
// internal/interfaces/http/middleware/rls_context_middleware.go
package middleware

import (
    "context"
    "modelcraft/internal/domain/rls"
    "net/http"
    "strings"
)

type rlsContextKey struct{}

// RLSContextMiddleware 从 X-MC-User-* Header 提取 UserContext 注入 context
type RLSContextMiddleware struct{}

func NewRLSContextMiddleware() *RLSContextMiddleware {
    return &RLSContextMiddleware{}
}

func (m *RLSContextMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        uc := &rls.UserContext{
            UserID:   strings.TrimSpace(r.Header.Get("X-MC-User-ID")),
            UserName: strings.TrimSpace(r.Header.Get("X-MC-User-Name")),
        }

        rolesStr := strings.TrimSpace(r.Header.Get("X-MC-User-Roles"))
        if rolesStr != "" {
            parts := strings.Split(rolesStr, ",")
            for _, p := range parts {
                p = strings.TrimSpace(p)
                if p != "" {
                    uc.Roles = append(uc.Roles, p)
                }
            }
        }

        ctx := context.WithValue(r.Context(), rlsContextKey{}, uc)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetUserContext 从 context 获取 UserContext
func GetUserContext(ctx context.Context) *rls.UserContext {
    uc, ok := ctx.Value(rlsContextKey{}).(*rls.UserContext)
    if !ok {
        return nil
    }
    return uc
}
```

- [ ] **Step 3: 运行测试**

```bash
cd modelcraft-backend && go test ./internal/interfaces/http/middleware/ -run "TestRLSContextMiddleware" -v
```

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/internal/interfaces/http/middleware/
git commit -m "feat(rls): add RLSContextMiddleware for X-MC-User-* header extraction"
```

---

### Task 5: Policy Matching Engine

**Files:**
- Create: `modelcraft-backend/internal/app/rls/policy_matching_service.go`
- Create: `modelcraft-backend/internal/app/rls/policy_matching_service_test.go`

- [ ] **Step 1: 编写测试 — role 匹配 + OR 合并**

```go
// internal/app/rls/policy_matching_service_test.go
package rls

import (
    "context"
    "modelcraft/internal/domain/rls"
    "testing"
)

type mockPolicyRepo struct {
    policies []*rls.Policy
}

func (m *mockPolicyRepo) ListByAction(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, roles []string) ([]*rls.Policy, error) {
    var result []*rls.Policy
    for _, p := range m.policies {
        if p.OrgName != orgName || p.ProjectSlug != projectSlug || p.ModelID != modelID || p.Action != action {
            continue
        }
        userCtx := &rls.UserContext{Roles: roles}
        if userCtx.HasRole(p.Role) {
            result = append(result, p)
        }
    }
    return result, nil
}

func TestMatch_MultipleRoles_OrMerge(t *testing.T) {
    compiler := NewPolicyCompiler()
    repo := &mockPolicyRepo{policies: []*rls.Policy{
        {
            PolicyName: "admin_all", Action: rls.ActionRead, Role: "admin",
            UsingExpr: `true`,
        },
        {
            PolicyName: "user_own", Action: rls.ActionRead, Role: "user",
            UsingExpr: `{"tenant_id": {"equals": "{{user_id}}"}}`,
        },
    }}
    svc := NewPolicyMatchingService(repo, compiler)

    ctx := context.Background()
    userCtx := &rls.UserContext{UserID: "123", Roles: []string{"admin", "user"}}

    sql, params, err := svc.ResolveUsing(ctx, "my-org", "my-proj", "model-1", rls.ActionRead, userCtx)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // OR merge: (1=1) OR (tenant_id = ?)
    expected := "(1=1) OR (tenant_id = ?)"
    if sql != expected {
        t.Errorf("expected SQL %q, got %q", expected, sql)
    }
    if len(params) != 1 || params[0] != "123" {
        t.Errorf("expected params [123], got %v", params)
    }
}

func TestMatch_NoMatchingPolicies_DenyAll(t *testing.T) {
    compiler := NewPolicyCompiler()
    repo := &mockPolicyRepo{policies: []*rls.Policy{
        {
            PolicyName: "admin_only", Action: rls.ActionRead, Role: "admin",
            UsingExpr: `true`,
        },
    }}
    svc := NewPolicyMatchingService(repo, compiler)

    ctx := context.Background()
    userCtx := &rls.UserContext{Roles: []string{"user"}}

    _, _, err := svc.ResolveUsing(ctx, "my-org", "my-proj", "model-1", rls.ActionRead, userCtx)
    if err == nil {
        t.Fatal("expected deny-all error, got nil")
    }
}
```

- [ ] **Step 2: 实现 PolicyMatchingService**

```go
// internal/app/rls/policy_matching_service.go
package rls

import (
    "context"
    "fmt"
    "modelcraft/internal/domain/rls"
    "strings"
)

// PolicyRepository 策略查询接口
type PolicyRepository interface {
    ListByAction(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, roles []string) ([]*rls.Policy, error)
}

// PolicyMatchingService 策略匹配 + OR 合并引擎
type PolicyMatchingService struct {
    repo     PolicyRepository
    compiler rls.PolicyCompiler
}

func NewPolicyMatchingService(repo PolicyRepository, compiler rls.PolicyCompiler) *PolicyMatchingService {
    return &PolicyMatchingService{repo: repo, compiler: compiler}
}

// ResolveUsing 匹配策略并 OR 合并 using 表达式
func (s *PolicyMatchingService) ResolveUsing(
    ctx context.Context, orgName, projectSlug, modelID string,
    action rls.Action, userCtx *rls.UserContext,
) (string, []interface{}, error) {
    policies, err := s.repo.ListByAction(ctx, orgName, projectSlug, modelID, action, userCtx.Roles)
    if err != nil {
        return "", nil, err
    }

    if len(policies) == 0 {
        return "", nil, fmt.Errorf("RLS deny: no matching policy for action=%s, roles=%v", action, userCtx.Roles)
    }

    var orClauses []string
    var allParams []interface{}

    for _, p := range policies {
        expr := p.UsingExpr
        if action == rls.ActionCreate {
            expr = p.WithCheckExpr
        }

        compiled, err := s.compiler.Compile(ctx, expr, userCtx)
        if err != nil {
            return "", nil, fmt.Errorf("compile policy %q: %w", p.PolicyName, err)
        }
        orClauses = append(orClauses, "("+compiled.SQL+")")
        allParams = append(allParams, compiled.Params...)
    }

    return strings.Join(orClauses, " OR "), allParams, nil
}

// ResolveCheck 匹配策略并 OR 合并 withCheck 表达式
func (s *PolicyMatchingService) ResolveCheck(
    ctx context.Context, orgName, projectSlug, modelID string,
    action rls.Action, userCtx *rls.UserContext,
) (string, []interface{}, error) {
    policies, err := s.repo.ListByAction(ctx, orgName, projectSlug, modelID, action, userCtx.Roles)
    if err != nil {
        return "", nil, err
    }

    if len(policies) == 0 {
        return "", nil, fmt.Errorf("RLS deny: no matching policy for action=%s", action)
    }

    var orClauses []string
    var allParams []interface{}

    for _, p := range policies {
        expr := p.WithCheckExpr
        if expr == "" {
            continue
        }

        compiled, err := s.compiler.Compile(ctx, expr, userCtx)
        if err != nil {
            return "", nil, fmt.Errorf("compile check expression for policy %q: %w", p.PolicyName, err)
        }
        orClauses = append(orClauses, "("+compiled.SQL+")")
        allParams = append(allParams, compiled.Params...)
    }

    if len(orClauses) == 0 {
        return "1=0", nil, nil // 无 CHECK 策略 → 拒绝
    }

    return strings.Join(orClauses, " OR "), allParams, nil
}
```

- [ ] **Step 3: 运行测试**

```bash
cd modelcraft-backend && go test ./internal/app/rls/ -run "TestMatch_" -v
```

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/internal/app/rls/
git commit -m "feat(rls): add policy matching engine with role-based matching and OR merging"
```

---

### Task 6: Repository 实现

**Files:**
- Create: `modelcraft-backend/internal/infrastructure/persistence/rls/policy_repository.go`

- [ ] **Step 1: 实现 sqlc-backed Repository**

```go
// internal/infrastructure/persistence/rls/policy_repository.go
package rls

import (
    "context"
    "modelcraft/internal/domain/rls"
    "modelcraft/internal/infrastructure/persistence/rls/sqlc" // sqlc generated
)

type PolicyRepository struct {
    db *sqlc.Queries
}

func NewPolicyRepository(db *sqlc.Queries) *PolicyRepository {
    return &PolicyRepository{db: db}
}

func (r *PolicyRepository) ListByAction(
    ctx context.Context, orgName, projectSlug, modelID string,
    action rls.Action, roles []string,
) ([]*rls.Policy, error) {
    rows, err := r.db.ListPoliciesByAction(ctx, sqlc.ListPoliciesByActionParams{
        OrgName:     orgName,
        ProjectSlug: projectSlug,
        ModelID:     modelID,
        Action:      string(action),
        Roles:       append(roles, ""), // role="" 总是匹配
    })
    if err != nil {
        return nil, err
    }

    policies := make([]*rls.Policy, 0, len(rows))
    for _, row := range rows {
        policies = append(policies, &rls.Policy{
            ID:           row.ID,
            OrgName:      row.OrgName,
            ProjectSlug:  row.ProjectSlug,
            ModelID:      row.ModelID,
            PolicyName:   row.PolicyName,
            Action:       rls.Action(row.Action),
            Role:         row.Role,
            UsingExpr:    rls.JsonExpr(row.UsingExpr),
            WithCheckExpr: rls.JsonExpr(row.WithCheckExpr),
            CreatedAt:    row.CreatedAt,
            UpdatedAt:    row.UpdatedAt,
        })
    }

    return policies, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-backend/internal/infrastructure/persistence/rls/
git commit -m "feat(rls): add PolicyRepository implementation with sqlc"
```

---

### Task 7: 集成 — 替换 RLSResolver 接入 modelruntime

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/runtime/rls_resolver.go`
- Modify: `modelcraft-backend/internal/interfaces/runtime/handler.go`

- [ ] **Step 1: 重写 RLSResolver，使用 PolicyMatchingService**

核心改动：`RLSResolver` 从单策略五表达式模型改为多策略匹配引擎。

```go
// internal/interfaces/runtime/rls_resolver.go

// RLSResolver 现在使用 PolicyMatchingService 做多策略匹配
type RLSResolver struct {
    logger      logfacade.Logger
    matchingSvc MatchingService
}

// MatchingService 匹配引擎接口（app/rls.PolicyMatchingService 实现）
type MatchingService interface {
    ResolveUsing(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, userCtx *rls.UserContext) (string, []interface{}, error)
    ResolveCheck(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, userCtx *rls.UserContext) (string, []interface{}, error)
}

func NewRLSResolver(logger logfacade.Logger, matchingSvc MatchingService) *RLSResolver {
    return &RLSResolver{logger: logger, matchingSvc: matchingSvc}
}

// ResolveResult 简化：不再有 Filter 结构，直接返回 SQL
type ResolveResult struct {
    UsingSQL    string
    UsingParams []interface{}
    CheckSQL    string
    CheckParams []interface{}
    ShouldApply bool
    DenyAll     bool
}

func (r *RLSResolver) Resolve(ctx context.Context, action rls.Action, modelID string) (*ResolveResult, error) {
    identity := middleware.GetEndUserIdentity(ctx)
    userCtx := middleware.GetUserContext(ctx)

    // Developer 不受 RLS 限制
    if identity != nil && identity.IsDeveloper() {
        return &ResolveResult{ShouldApply: false}, nil
    }

    rctx, ok := getRuntimeContext(ctx)
    if !ok {
        return &ResolveResult{ShouldApply: true, DenyAll: true}, nil
    }

    if userCtx == nil {
        userCtx = &rls.UserContext{}
    }

    usingSQL, usingParams, err := r.matchingSvc.ResolveUsing(ctx, rctx.OrgName, rctx.ProjectSlug, modelID, action, userCtx)
    if err != nil {
        r.logger.Debug(ctx, "RLS policy denied", logfacade.Err(err))
        return &ResolveResult{ShouldApply: true, DenyAll: true}, nil
    }

    checkSQL, checkParams, _ := r.matchingSvc.ResolveCheck(ctx, rctx.OrgName, rctx.ProjectSlug, modelID, action, userCtx)

    return &ResolveResult{
        UsingSQL:    usingSQL,
        UsingParams: usingParams,
        CheckSQL:    checkSQL,
        CheckParams: checkParams,
        ShouldApply: true,
    }, nil
}
```

- [ ] **Step 2: 更新 handler.go — 注入依赖**

```go
// internal/interfaces/runtime/handler.go
// 在 runtime Handler 初始化时注入 RLSResolver

// 原有:
// resolver := NewRLSResolver(logger, policyRepo)
// 改为:
// resolver := NewRLSResolver(logger, matchingSvc)
```

- [ ] **Step 3: 更新 model_resolver.go — 使用 ResolveResult**

在各 `execute*` 方法中使用新的 `ResolveResult`：
- `executeFindMany/FindUnique/FindFirst`: 调用 `Resolve(ctx, ActionRead, modelID)`, 将 `UsingSQL` + `UsingParams` 注入 WHERE
- `executeCreateOne/CreateMany`: 调用 `Resolve(ctx, ActionCreate, modelID)`, 使用 `CheckSQL` 校验
- `executeUpdateOne/UpdateMany`: 调用 `Resolve(ctx, ActionUpdate, modelID)`, USING + CHECK
- `executeDeleteOne/DeleteMany`: 调用 `Resolve(ctx, ActionDelete, modelID)`, USING

- [ ] **Step 4: 在 wire/dependency injection 中注册新组件**

```go
// 在依赖注入处注册：
// policyRepo := persistence_rls.NewPolicyRepository(queries)
// compiler := app_rls.NewPolicyCompiler()
// matchingSvc := app_rls.NewPolicyMatchingService(policyRepo, compiler)
// rlsResolver := runtime.NewRLSResolver(logger, matchingSvc)
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/interfaces/runtime/ \
        modelcraft-backend/internal/domain/modelruntime/
git commit -m "feat(rls): integrate multi-policy matching engine into modelruntime"
```

---

### Task 8: Python SDK

**Files:**
- Create: `modelcraft-sdk-python/modelcraft/__init__.py`
- Create: `modelcraft-sdk-python/modelcraft/client.py`
- Create: `modelcraft-sdk-python/setup.py`

- [ ] **Step 1: 实现 Client**

```python
# modelcraft-sdk-python/modelcraft/client.py
import requests
from typing import Optional, Dict, Any, List

class Client:
    def __init__(
        self,
        endpoint: str,
        api_token: str,
        org: str,
        project: str,
    ):
        self.endpoint = endpoint.rstrip("/")
        self.api_token = api_token
        self.org = org
        self.project = project

    def _headers(
        self,
        user_id: str = "",
        user_name: str = "",
        user_roles: str = "",
    ) -> Dict[str, str]:
        return {
            "Authorization": f"Bearer {self.api_token}",
            "Content-Type": "application/json",
            "X-MC-User-ID": user_id,
            "X-MC-User-Name": user_name,
            "X-MC-User-Roles": user_roles,
        }

    def model(self, model_name: str) -> "ModelQuery":
        return ModelQuery(self, model_name)


class ModelQuery:
    def __init__(self, client: Client, model_name: str):
        self._client = client
        self._model = model_name

    def _url(self, db: str) -> str:
        return (
            f"{self._client.endpoint}/end-user/graphql"
            f"/org/{self._client.org}/project/{self._client.project}"
            f"/db/{db}/model/{self._model}"
        )

    def list(
        self,
        db: str,
        limit: int = 10,
        where: Optional[Dict] = None,
        user_id: str = "",
        user_name: str = "",
        user_roles: str = "",
    ) -> List[Dict[str, Any]]:
        query = """
        query List($limit: Int!, $where: JSON) {
            findMany(limit: $limit, where: $where) {
                items
            }
        }
        """
        resp = requests.post(
            self._url(db),
            json={
                "query": query,
                "variables": {"limit": limit, "where": where or {}},
            },
            headers=self._client._headers(user_id, user_name, user_roles),
        )
        resp.raise_for_status()
        data = resp.json()
        if "errors" in data:
            raise Exception(data["errors"])
        return data.get("data", {}).get("findMany", {}).get("items", [])

    def create(
        self,
        db: str,
        data: Dict[str, Any],
        user_id: str = "",
        user_name: str = "",
        user_roles: str = "",
    ) -> Dict[str, Any]:
        query = """
        mutation Create($data: JSON!) {
            createOne(data: $data) {
                item
            }
        }
        """
        resp = requests.post(
            self._url(db),
            json={"query": query, "variables": {"data": data}},
            headers=self._client._headers(user_id, user_name, user_roles),
        )
        resp.raise_for_status()
        result = resp.json()
        if "errors" in result:
            raise Exception(result["errors"])
        return result.get("data", {}).get("createOne", {}).get("item", {})

    def update(
        self,
        db: str,
        where: Dict[str, Any],
        data: Dict[str, Any],
        user_id: str = "",
        user_name: str = "",
        user_roles: str = "",
    ) -> Dict[str, Any]:
        query = """
        mutation Update($where: JSON!, $data: JSON!) {
            updateOne(where: $where, data: $data) {
                item
            }
        }
        """
        resp = requests.post(
            self._url(db),
            json={"query": query, "variables": {"where": where, "data": data}},
            headers=self._client._headers(user_id, user_name, user_roles),
        )
        resp.raise_for_status()
        result = resp.json()
        if "errors" in result:
            raise Exception(result["errors"])
        return result.get("data", {}).get("updateOne", {}).get("item", {})

    def delete(
        self,
        db: str,
        where: Dict[str, Any],
        user_id: str = "",
        user_name: str = "",
        user_roles: str = "",
    ) -> bool:
        query = """
        mutation Delete($where: JSON!) {
            deleteOne(where: $where) {
                item { id }
            }
        }
        """
        resp = requests.post(
            self._url(db),
            json={"query": query, "variables": {"where": where}},
            headers=self._client._headers(user_id, user_name, user_roles),
        )
        resp.raise_for_status()
        return "errors" not in resp.json()
```

- [ ] **Step 2: 编写 `__init__.py`**

```python
# modelcraft-sdk-python/modelcraft/__init__.py
from .client import Client

__all__ = ["Client"]
```

- [ ] **Step 3: 编写 setup.py**

```python
# modelcraft-sdk-python/setup.py
from setuptools import setup, find_packages

setup(
    name="modelcraft",
    version="0.1.0",
    packages=find_packages(),
    install_requires=["requests>=2.28"],
    python_requires=">=3.9",
)
```

- [ ] **Step 4: Commit**

```bash
git add modelcraft-sdk-python/
git commit -m "feat(sdk): add Python SDK with RLS context header support"
```

---

---

### Task 9: GraphQL API — Policy CRUD

**Files:**
- Create: `modelcraft-backend/api/graph/project/schema/rls_policy.graphql`
- Create: `modelcraft-backend/internal/interfaces/graphql/project/rls_policy.resolvers.go`
- Create: `modelcraft-backend/internal/app/rls/policy_crud_service.go`

- [ ] **Step 1: 定义 GraphQL Schema**

```graphql
# api/graph/project/schema/rls_policy.graphql

type RlsPolicy {
    id:           ID!
    policyName:   String!
    action:       RlsAction!
    role:         String!
    usingExpr:    JSON
    withCheckExpr: JSON
}

enum RlsAction {
    read
    create
    update
    delete
}

input RlsPolicyInput {
    policyName:   String!
    action:       RlsAction!
    role:         String!
    usingExpr:    JSON
    withCheckExpr: JSON
}

extend type Query {
    rlsPolicies(modelId: ID!): [RlsPolicy!]!
}

extend type Mutation {
    upsertRlsPolicy(modelId: ID!, input: RlsPolicyInput!): RlsPolicy!
    deleteRlsPolicy(id: ID!): Boolean!
    deleteRlsPoliciesByModel(modelId: ID!): Boolean!
}
```

- [ ] **Step 2: 运行代码生成**

```bash
cd modelcraft-backend && just generate-gql
```

- [ ] **Step 3: 实现 CRUD App Service**

```go
// internal/app/rls/policy_crud_service.go
package rls

import (
    "context"
    "modelcraft/internal/domain/rls"
)

type PolicyCRUDService struct {
    repo PolicyRepository
}

func NewPolicyCRUDService(repo PolicyRepository) *PolicyCRUDService {
    return &PolicyCRUDService{repo: repo}
}

type UpsertPolicyInput struct {
    OrgName      string
    ProjectSlug  string
    ModelID      string
    PolicyName   string
    Action       rls.Action
    Role         string
    UsingExpr    rls.JsonExpr
    WithCheckExpr rls.JsonExpr
}

func (s *PolicyCRUDService) Upsert(ctx context.Context, input UpsertPolicyInput) (*rls.Policy, error) {
    // implementation
}

func (s *PolicyCRUDService) ListByModel(ctx context.Context, orgName, projectSlug, modelID string) ([]*rls.Policy, error) {
    // implementation
}

func (s *PolicyCRUDService) Delete(ctx context.Context, orgName, projectSlug string, id int64) error {
    // implementation
}
```

- [ ] **Step 4: 实现 Resolver**

```go
// internal/interfaces/graphql/project/rls_policy.resolvers.go
// 实现 Query.rlsPolicies, Mutation.upsertRlsPolicy, Mutation.deleteRlsPolicy
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/api/graph/project/schema/rls_policy.graphql \
        modelcraft-backend/internal/interfaces/graphql/project/ \
        modelcraft-backend/internal/app/rls/policy_crud_service.go
git commit -m "feat(rls): add GraphQL API for policy CRUD"
```

---

## 实现顺序总结

1. **DB Migration** (Task 1) — 新表结构
2. **Domain Entity** (Task 2) — Policy + UserContext
3. **Expression Compiler** (Task 3) — Prisma 操作符 + 变量
4. **RLS Context Middleware** (Task 4) — Header 提取
5. **Policy Matching Engine** (Task 5) — role 匹配 + OR 合并
6. **Repository 实现** (Task 6) — sqlc CRUD
7. **Integration** (Task 7) — 接入 modelruntime 执行链路
8. **GraphQL API** (Task 9) — Policy CRUD（管理员配置接口）
9. **Python SDK** (Task 8) — 客户端封装
