# End-User Runtime 数据权限体系 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 modelruntime GraphQL 执行层为 end-user 请求接入 RBAC，实现 Action Gate（拒绝无权限操作）与 Row Filter（SELF scope 注入 WHERE 条件），tenant admin 请求跳过所有检查。

**Architecture:** 在 `domain/modelruntime/` 新增 `ResolvedModelPermissions` 值类型与 `EndUserPermissionService` 接口（依赖倒置，防止 domain 层循环依赖 rbac 包）；`app/modelruntime/` 实现该接口，per-request 按 model 粒度查询 RBAC；`graphqlRequestContext` 携带权限快照，各 `execute*` resolver 读取执行 gate 与 where 注入。

**Tech Stack:** Go, graphql-go, sqlc (dbgen), bizerrors, modelruntime domain, rbac domain

---

## File Map

| 文件 | 操作 | 说明 |
|------|------|------|
| `modelcraft-backend/internal/domain/modelruntime/permission.go` | **新建** | `Action`, `ActionPermission`, `ResolvedModelPermissions`, `EndUserPermissionService` |
| `modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go` | **修改** | 新增 `EndUserPerms` 字段，更新构造函数 |
| `modelcraft-backend/internal/domain/rbac/repository.go` | **修改** | 新增 `FindPermissionsByEndUserAndModel` 方法到接口 |
| `modelcraft-backend/internal/infrastructure/repository/sql_end_user_permission_repository.go` | **修改** | 实现新 repo 方法 |
| `modelcraft-backend/internal/app/modelruntime/permission_service.go` | **新建** | `endUserPermissionServiceImpl`，查 RBAC → 映射为 `ResolvedModelPermissions` |
| `modelcraft-backend/internal/app/modelruntime/graphql_app.go` | **修改** | 注入 `permService`，Execute() 先加载 model 再解析权限，构造含 `EndUserPerms` 的 reqCtx |
| `modelcraft-backend/internal/domain/modelruntime/model_resolver.go` | **修改** | 各 `execute*` 方法加 Action Gate + Row Filter |
| `modelcraft-backend/internal/interfaces/http/routes.go` | **修改** | `CreateRuntimeHandlers` 中构造并注入 `endUserPermissionServiceImpl` |

---

## Task 1: `domain/modelruntime/permission.go` — 权限类型与接口

**Files:**
- Create: `modelcraft-backend/internal/domain/modelruntime/permission.go`
- Test: `modelcraft-backend/internal/domain/modelruntime/permission_test.go`

- [ ] **Step 1: 写失败测试**

```go
// modelcraft-backend/internal/domain/modelruntime/permission_test.go
package modelruntime_test

import (
	"testing"

	"modelcraft/internal/domain/modelruntime"
)

func TestResolvedModelPermissions_CheckAction_NilSkipsAll(t *testing.T) {
	var p *modelruntime.ResolvedModelPermissions
	// nil = tenant admin, all actions allowed
	for _, action := range []modelruntime.Action{
		modelruntime.ActionSelect, modelruntime.ActionInsert,
		modelruntime.ActionUpdate, modelruntime.ActionDelete,
	} {
		if err := p.CheckAction(action); err != nil {
			t.Errorf("nil permissions should allow action %s, got error: %v", action, err)
		}
	}
}

func TestResolvedModelPermissions_CheckAction_DeniesWhenNotAllowed(t *testing.T) {
	p := &modelruntime.ResolvedModelPermissions{
		Select: modelruntime.ActionPermission{Allowed: true, IsSelf: false},
		// Insert/Update/Delete all false
	}
	if err := p.CheckAction(modelruntime.ActionSelect); err != nil {
		t.Errorf("expected select to be allowed, got: %v", err)
	}
	if err := p.CheckAction(modelruntime.ActionInsert); err == nil {
		t.Error("expected insert to be denied")
	}
	if err := p.CheckAction(modelruntime.ActionUpdate); err == nil {
		t.Error("expected update to be denied")
	}
	if err := p.CheckAction(modelruntime.ActionDelete); err == nil {
		t.Error("expected delete to be denied")
	}
}

func TestResolvedModelPermissions_Get(t *testing.T) {
	p := &modelruntime.ResolvedModelPermissions{
		Select: modelruntime.ActionPermission{Allowed: true, IsSelf: true},
		Insert: modelruntime.ActionPermission{Allowed: true, IsSelf: false},
		Update: modelruntime.ActionPermission{Allowed: false},
		Delete: modelruntime.ActionPermission{Allowed: false},
	}
	if got := p.Get(modelruntime.ActionSelect); !got.Allowed || !got.IsSelf {
		t.Errorf("Select: want {true,true}, got %+v", got)
	}
	if got := p.Get(modelruntime.ActionInsert); !got.Allowed || got.IsSelf {
		t.Errorf("Insert: want {true,false}, got %+v", got)
	}
	// unknown action
	if got := p.Get(modelruntime.Action("UNKNOWN")); got.Allowed {
		t.Error("unknown action should be denied")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/domain/modelruntime/... -run TestResolvedModelPermissions -v 2>&1 | head -20
```
Expected: 编译失败（类型未定义）

- [ ] **Step 3: 实现 `permission.go`**

```go
// modelcraft-backend/internal/domain/modelruntime/permission.go
package modelruntime

import (
	"context"

	"modelcraft/pkg/bizerrors"
)

// Action 操作类型。在 domain/modelruntime 内独立定义，避免循环依赖 domain/rbac。
// 值与 rbac.Action 对应（均为小写），app 层负责转换。
type Action string

const (
	ActionSelect Action = "select"
	ActionInsert Action = "insert"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// ActionPermission 单个操作的权限状态。
type ActionPermission struct {
	Allowed bool
	// IsSelf true 表示 rowScope=SELF：查询/更新/删除时需注入
	// WHERE <EndUserRef字段> = $currentEndUserID
	IsSelf bool
}

// ResolvedModelPermissions 单次请求的权限快照。
// 在 Execute() 入口解析一次，注入 graphqlRequestContext，resolver 只读。
// nil 表示 tenant admin 请求，跳过所有检查。
type ResolvedModelPermissions struct {
	Select ActionPermission
	Insert ActionPermission
	Update ActionPermission
	Delete ActionPermission
}

// Get 返回指定 action 的权限状态。未知 action 返回 denied。
func (p *ResolvedModelPermissions) Get(action Action) ActionPermission {
	if p == nil {
		return ActionPermission{Allowed: true}
	}
	switch action {
	case ActionSelect:
		return p.Select
	case ActionInsert:
		return p.Insert
	case ActionUpdate:
		return p.Update
	case ActionDelete:
		return p.Delete
	default:
		return ActionPermission{Allowed: false}
	}
}

// CheckAction 默认拒绝原则。nil receiver = tenant admin，直接放行。
func (p *ResolvedModelPermissions) CheckAction(action Action) error {
	if p == nil {
		return nil // tenant admin: skip all checks
	}
	if !p.Get(action).Allowed {
		return bizerrors.NewError(bizerrors.PermissionDenied, string(action))
	}
	return nil
}

// EndUserPermissionService 依赖倒置接口，app 层实现。
// domain/modelruntime 只依赖此接口，不感知 domain/rbac 包细节。
type EndUserPermissionService interface {
	// Resolve 查询并合并 end-user 在指定 model 上的有效权限。
	// endUserID 为空（tenant admin）时返回 nil, nil。
	Resolve(ctx context.Context, orgName, projectSlug, endUserID, modelID string) (*ResolvedModelPermissions, error)
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd modelcraft-backend && go test ./internal/domain/modelruntime/... -run TestResolvedModelPermissions -v
```
Expected: PASS (3 test functions)

- [ ] **Step 5: 提交**

```bash
cd modelcraft-backend && git add internal/domain/modelruntime/permission.go internal/domain/modelruntime/permission_test.go
git commit -m "feat(modelruntime): add ResolvedModelPermissions types and EndUserPermissionService interface"
```

---

## Task 2: 更新 `graphqlRequestContext` 携带权限快照

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go`
- Test: `modelcraft-backend/internal/domain/modelruntime/graphql_request_context_test.go`（新建）

**背景：** `graphqlRequestContext` 当前构造函数签名为：
```go
func newGraphqlRequestContext(clientRepo ClientDatabaseRepository, orgName, projectSlug, currentEndUserID string) *graphqlRequestContext
func WithGraphqlRequestContext(ctx context.Context, clientRepo ClientDatabaseRepository, orgName, projectSlug, currentEndUserID string) context.Context
```
需新增 `EndUserPerms *ResolvedModelPermissions` 参数。

- [ ] **Step 1: 写失败测试**

```go
// modelcraft-backend/internal/domain/modelruntime/graphql_request_context_test.go
package modelruntime_test

import (
	"context"
	"testing"

	"modelcraft/internal/domain/modelruntime"
)

func TestWithGraphqlRequestContext_EndUserPerms(t *testing.T) {
	perms := &modelruntime.ResolvedModelPermissions{
		Select: modelruntime.ActionPermission{Allowed: true, IsSelf: true},
	}
	ctx := modelruntime.WithGraphqlRequestContext(
		context.Background(),
		nil, // clientRepo
		"org1", "proj1", "user123",
		perms,
	)
	rctx, ok := modelruntime.GetGraphqlRequestContextForTest(ctx)
	if !ok {
		t.Fatal("expected request context in ctx")
	}
	if rctx.EndUserPerms == nil {
		t.Fatal("expected EndUserPerms to be set")
	}
	if !rctx.EndUserPerms.Select.IsSelf {
		t.Error("expected Select.IsSelf = true")
	}
}

func TestWithGraphqlRequestContext_NilPerms_TenantAdmin(t *testing.T) {
	ctx := modelruntime.WithGraphqlRequestContext(
		context.Background(),
		nil, "org1", "proj1", "",
		nil, // tenant admin
	)
	rctx, ok := modelruntime.GetGraphqlRequestContextForTest(ctx)
	if !ok {
		t.Fatal("expected request context in ctx")
	}
	if rctx.EndUserPerms != nil {
		t.Error("tenant admin should have nil EndUserPerms")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/domain/modelruntime/... -run TestWithGraphqlRequestContext -v 2>&1 | head -20
```
Expected: 编译失败（`WithGraphqlRequestContext` 参数数量不匹配）

- [ ] **Step 3: 修改 `graphql_request_context.go`**

在 `graphqlRequestContext` struct 末尾新增字段，更新构造函数和 `WithGraphqlRequestContext` 签名，并导出测试用 getter：

```go
// graphqlRequestContext struct 末尾新增：
EndUserPerms *ResolvedModelPermissions // nil = tenant admin，跳过所有检查

// newGraphqlRequestContext 签名变更（新增最后一个参数）：
func newGraphqlRequestContext(
	clientRepo ClientDatabaseRepository,
	orgName, projectSlug, currentEndUserID string,
	endUserPerms *ResolvedModelPermissions,
) *graphqlRequestContext {
	return &graphqlRequestContext{
		ClientRepo:       clientRepo,
		relationLoaders:  make(map[string]*dataloader.Loader[string, map[string]any]),
		OrgName:          orgName,
		ProjectSlug:      projectSlug,
		CurrentEndUserID: currentEndUserID,
		EndUserPerms:     endUserPerms,
	}
}

// WithGraphqlRequestContext 签名变更（新增最后一个参数）：
func WithGraphqlRequestContext(
	ctx context.Context,
	clientRepo ClientDatabaseRepository,
	orgName, projectSlug, currentEndUserID string,
	endUserPerms *ResolvedModelPermissions,
) context.Context {
	rctx := newGraphqlRequestContext(clientRepo, orgName, projectSlug, currentEndUserID, endUserPerms)
	return context.WithValue(ctx, graphqlRequestContextKey{}, rctx)
}

// GetGraphqlRequestContextForTest 仅供测试用，导出内部 context 访问。
// 生产代码使用 getGraphqlRequestContext（未导出）。
func GetGraphqlRequestContextForTest(ctx context.Context) (*graphqlRequestContext, bool) {
	return getGraphqlRequestContext(ctx)
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd modelcraft-backend && go test ./internal/domain/modelruntime/... -run TestWithGraphqlRequestContext -v
```
Expected: PASS

- [ ] **Step 5: 确认编译无错（其他调用方会在下一个 task 修复）**

```bash
cd modelcraft-backend && go build ./internal/domain/modelruntime/... 2>&1 | head -20
```

- [ ] **Step 6: 提交**

```bash
cd modelcraft-backend && git add internal/domain/modelruntime/graphql_request_context.go internal/domain/modelruntime/graphql_request_context_test.go
git commit -m "feat(modelruntime): extend graphqlRequestContext with EndUserPerms field"
```

---

## Task 3: 新增 RBAC repo 接口方法 + 基础设施实现

**Files:**
- Modify: `modelcraft-backend/internal/domain/rbac/repository.go`
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_end_user_permission_repository.go`

**背景：** `EndUserPermissionRepository` 接口（`domain/rbac/repository.go`）需新增方法。实现在 `SqlEndUserDataPermissionRepository`。

查询链路：`end_user_project_access(end_user_id, role_id)` → `end_user_role_bundles(role_id, bundle_id)` → `end_user_bundle_permissions(bundle_id, permission_id)` → `end_user_data_permissions(id=permission_id, model_id=?)`

需先确认表名：

```bash
grep -rn "end_user_project_access\|end_user_role_bundles\|end_user_bundle_permissions\|end_user_data_permissions" \
  modelcraft-backend/internal/infrastructure/dbgen/ --include="*.go" | grep "FROM\|JOIN\|TABLE" | head -20
```

- [ ] **Step 1: 确认表名和 schema**

```bash
cd modelcraft-backend && grep -rn "end_user_project_access\|EndUserProjectAccess" internal/infrastructure/dbgen/ --include="*.go" | head -10
grep -rn "end_user_role_bundles\|EndUserRoleBundle" internal/infrastructure/dbgen/ --include="*.go" | head -10
grep -rn "end_user_bundle_permissions\|EndUserBundlePermission" internal/infrastructure/dbgen/ --include="*.go" | head -10
```

Expected: 能看到 SQL 中的真实表名（注意：表名可能有 `private_` 前缀或 schema 前缀）

- [ ] **Step 2: 新增 repo 接口方法**

在 `modelcraft-backend/internal/domain/rbac/repository.go` 的 `EndUserPermissionRepository` 接口末尾，`GetPermissionsByBundleIDs` 方法之后，新增：

```go
// FindPermissionsByEndUserAndModel 查询指定 end-user 在某 model 上的
// 所有有效权限点（跨 role → bundle → permission 链路）。
// 仅查该 model，不全量拉取，用于 per-request 权限解析。
// endUserID 或 modelID 为空时返回空 slice（不报错）。
FindPermissionsByEndUserAndModel(
    ctx context.Context,
    orgName, projectSlug, endUserID, modelID string,
) ([]*EndUserPermission, error)
```

- [ ] **Step 3: 确认编译（接口新增方法，实现类还未加，预期编译失败）**

```bash
cd modelcraft-backend && go build ./internal/domain/rbac/... 2>&1 | head -5
```

- [ ] **Step 4: 在 `sql_end_user_permission_repository.go` 末尾实现新方法**

先用 Step 1 确认的真实表名替换下面 SQL 中的表名占位符。典型 SQL 模式：

```go
// FindPermissionsByEndUserAndModel 通过 role → bundle → permission 链路，
// 查询指定 end-user 在某 model 上的全部权限点。
func (r *SqlEndUserDataPermissionRepository) FindPermissionsByEndUserAndModel(
	ctx context.Context,
	orgName, projectSlug, endUserID, modelID string,
) ([]*rbac.EndUserPermission, error) {
	if endUserID == "" || modelID == "" {
		return nil, nil
	}

	// Raw SQL: end_user_project_access → role_bundles → bundle_permissions → permissions
	// 用真实表名（从 Step 1 确认）
	const query = `
		SELECT p.id, p.org_name, p.project_slug, p.model_id, p.database_name, p.model_name,
		       p.name, p.description, p.type, p.column_policy, p.row_policy, p.preset
		FROM end_user_project_access eupa
		JOIN end_user_role_bundles erb   ON erb.role_id   = eupa.role_id
		JOIN end_user_bundle_permissions ebp ON ebp.bundle_id = erb.bundle_id
		JOIN end_user_data_permissions p ON p.id          = ebp.permission_id
		WHERE eupa.org_name     = ?
		  AND eupa.project_slug = ?
		  AND eupa.end_user_id  = ?
		  AND p.model_id        = ?
	`

	db := r.q.DB() // 注意：若 q 不暴露 DB()，见 Step 5 说明

	rows, err := db.QueryContext(ctx, query, orgName, projectSlug, endUserID, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*rbac.EndUserPermission
	for rows.Next() {
		var (
			id           string
			orgN         string
			projSlug     string
			modelIDVal   string
			dbName       sql.NullString
			modelName    sql.NullString
			name         string
			description  sql.NullString
			pType        string
			columnPolicy []byte
			rowPolicy    []byte
			preset       sql.NullString
		)
		if err := rows.Scan(&id, &orgN, &projSlug, &modelIDVal, &dbName, &modelName,
			&name, &description, &pType, &columnPolicy, &rowPolicy, &preset); err != nil {
			return nil, err
		}
		perm := scanToPermission(id, orgN, projSlug, modelIDVal, dbName, modelName,
			name, description, pType, columnPolicy, rowPolicy, preset)
		result = append(result, perm)
	}
	return result, rows.Err()
}
```

**注意：** `r.q` 是 `dbgen.Querier`（interface）。若需要裸 `*sql.DB` 跑 raw SQL，需要查看 `SqlEndUserDataPermissionRepository` 是否持有 `*sql.DB`。若没有，改为使用已有的 sqlc 查询方法组合实现（查 `GetBundleIDsByUserExplicitRoles` + `GetPermissionsByBundleIDs` + 在 Go 层过滤 model_id）。先检查：

```bash
grep -n "type SqlEndUserDataPermissionRepository\|sql.DB\|sqlDB\|db \*" \
  modelcraft-backend/internal/infrastructure/repository/sql_end_user_permission_repository.go | head -10
```

若无裸 DB，使用**已有 sqlc 方法组合**实现：

```go
func (r *SqlEndUserDataPermissionRepository) FindPermissionsByEndUserAndModel(
	ctx context.Context,
	orgName, projectSlug, endUserID, modelID string,
) ([]*rbac.EndUserPermission, error) {
	if endUserID == "" || modelID == "" {
		return nil, nil
	}
	// Step 1: 通过显式角色链路拿到 bundleIDs
	bundleIDs, err := r.GetBundleIDsByUserExplicitRoles(ctx, endUserID, orgName, projectSlug)
	if err != nil {
		return nil, err
	}
	if len(bundleIDs) == 0 {
		return nil, nil
	}
	// Step 2: 展开 bundleIDs → permissions
	allPerms, err := r.GetPermissionsByBundleIDs(ctx, orgName, bundleIDs)
	if err != nil {
		return nil, err
	}
	// Step 3: 过滤 model_id
	var result []*rbac.EndUserPermission
	for _, p := range allPerms {
		if p.ModelID == modelID {
			result = append(result, p)
		}
	}
	return result, nil
}
```

- [ ] **Step 5: 编译确认**

```bash
cd modelcraft-backend && go build ./internal/domain/rbac/... ./internal/infrastructure/repository/... 2>&1 | head -20
```
Expected: 0 errors

- [ ] **Step 6: 提交**

```bash
cd modelcraft-backend && git add internal/domain/rbac/repository.go internal/infrastructure/repository/sql_end_user_permission_repository.go
git commit -m "feat(rbac): add FindPermissionsByEndUserAndModel repo method"
```

---

## Task 4: `app/modelruntime/permission_service.go` — 查询 RBAC → 映射权限快照

**Files:**
- Create: `modelcraft-backend/internal/app/modelruntime/permission_service.go`
- Test: `modelcraft-backend/internal/app/modelruntime/permission_service_test.go`

**背景：** `rbac.Action` 值为小写（`"select"`, `"insert"`, `"update"`, `"delete"`）。`rbac.EffectivePermissionSet.Merge()` 接受 `[]*rbac.EndUserPermission`，返回合并后的 set；`rbac.RowScopeAll` → `IsSelf=false`，`rbac.RowScopeSelf` → `IsSelf=true`。

- [ ] **Step 1: 写失败测试**

```go
// modelcraft-backend/internal/app/modelruntime/permission_service_test.go
package modelruntime_test

import (
	"context"
	"encoding/json"
	"testing"

	appmodelruntime "modelcraft/internal/app/modelruntime"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rbac"
)

// stubRBACRepo 是最小化的 stub，只实现测试需要的方法。
type stubRBACRepo struct {
	permissions []*rbac.EndUserPermission
}

func (s *stubRBACRepo) FindPermissionsByEndUserAndModel(
	_ context.Context, _, _, _, _ string,
) ([]*rbac.EndUserPermission, error) {
	return s.permissions, nil
}

// 其余接口方法：空实现（满足编译）
func (s *stubRBACRepo) CreatePermission(_ context.Context, _ *rbac.EndUserPermission) error {
	return nil
}
func (s *stubRBACRepo) GetPermissionByID(_ context.Context, _, _ string) (*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) ListPermissionsByProject(_ context.Context, _, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) ListPermissionsByModel(_ context.Context, _, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) ListPresetPermissionsByModel(_ context.Context, _, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetPermissionByModelTypeName(_ context.Context, _, _ string, _ rbac.PermissionType, _ string) (*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) UpdatePermission(_ context.Context, _ *rbac.EndUserPermission) error { return nil }
func (s *stubRBACRepo) DeletePermission(_ context.Context, _, _ string) error               { return nil }
func (s *stubRBACRepo) DeletePresetPermissionsByModel(_ context.Context, _, _ string) error { return nil }
func (s *stubRBACRepo) UpdatePresetPermission(_ context.Context, _ *rbac.EndUserPermission) error {
	return nil
}
func (s *stubRBACRepo) IsPermissionReferencedByBundle(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (s *stubRBACRepo) CreateBundle(_ context.Context, _ *rbac.EndUserPermissionBundle) error {
	return nil
}
func (s *stubRBACRepo) GetBundleByID(_ context.Context, _, _, _ string) (*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetBundleBySlug(_ context.Context, _, _, _ string) (*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}
func (s *stubRBACRepo) ListBundlesByProject(_ context.Context, _, _ string) ([]*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}
func (s *stubRBACRepo) UpdateBundle(_ context.Context, _ *rbac.EndUserPermissionBundle) error {
	return nil
}
func (s *stubRBACRepo) DeleteBundle(_ context.Context, _, _ string) error { return nil }
func (s *stubRBACRepo) AddPermissionToBundle(_ context.Context, _, _ string, _ int) error {
	return nil
}
func (s *stubRBACRepo) RemovePermissionFromBundle(_ context.Context, _, _ string) error { return nil }
func (s *stubRBACRepo) ListPermissionsInBundle(_ context.Context, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) UpsertBundleDataPermissionItem(_ context.Context, _ *rbac.EndUserBundleDataPermissionItem) error {
	return nil
}
func (s *stubRBACRepo) RemoveBundleDataPermissionItem(_ context.Context, _, _ string) error {
	return nil
}
func (s *stubRBACRepo) ListBundleDataPermissionItems(_ context.Context, _ string) ([]*rbac.EndUserBundleDataPermissionItem, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetBundleDataPermissionItemByBundleAndModel(_ context.Context, _, _ string) (*rbac.EndUserBundleDataPermissionItem, error) {
	return nil, nil
}
func (s *stubRBACRepo) SaveBundleSnapshot(_ context.Context, _ *rbac.BundleSnapshot) error { return nil }
func (s *stubRBACRepo) ListBundleSnapshots(_ context.Context, _ string) ([]rbac.BundleSnapshot, error) {
	return nil, nil
}
func (s *stubRBACRepo) DeleteOldBundleSnapshots(_ context.Context, _ string) error { return nil }
func (s *stubRBACRepo) GetBundleCurrentVersion(_ context.Context, _ string) (int, error) { return 0, nil }
func (s *stubRBACRepo) GetBundleSnapshotByVersion(_ context.Context, _ string, _ int) (*rbac.BundleSnapshot, error) {
	return nil, nil
}
func (s *stubRBACRepo) ClearBundlePermissions(_ context.Context, _ string) error { return nil }
func (s *stubRBACRepo) CreateRole(_ context.Context, _ *rbac.EndUserRole) error  { return nil }
func (s *stubRBACRepo) GetRoleByID(_ context.Context, _, _ string) (*rbac.EndUserRole, error) {
	return nil, nil
}
func (s *stubRBACRepo) ListRolesByProject(_ context.Context, _, _ string) ([]*rbac.EndUserRole, error) {
	return nil, nil
}
func (s *stubRBACRepo) UpdateRole(_ context.Context, _ *rbac.EndUserRole) error { return nil }
func (s *stubRBACRepo) DeleteRole(_ context.Context, _, _ string) error         { return nil }
func (s *stubRBACRepo) AssignBundleToRole(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubRBACRepo) RevokeBundleFromRole(_ context.Context, _, _ string) error     { return nil }
func (s *stubRBACRepo) ListBundlesByRole(_ context.Context, _ string) ([]*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}
func (s *stubRBACRepo) GrantBundleToUser(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubRBACRepo) RevokeBundleFromUser(_ context.Context, _, _, _, _ string) error {
	return nil
}
func (s *stubRBACRepo) AssignRoleToUser(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubRBACRepo) RevokeRoleFromUser(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubRBACRepo) ListProjectEndUserRoleUsers(_ context.Context, _ rbac.ListProjectEndUserRoleUsersQuery) ([]*rbac.ProjectEndUserRoleUser, int64, error) {
	return nil, 0, nil
}
func (s *stubRBACRepo) GetBundleIDsByUserDirect(_ context.Context, _, _, _ string) ([]string, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetBundleIDsByUserExplicitRoles(_ context.Context, _, _, _ string) ([]string, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetBundleIDsByImplicitRoles(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetPermissionsByBundleIDs(_ context.Context, _ string, _ []string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}

// ---- helpers ----

func makeRowPolicy(selectAllowed bool, selectScope rbac.PolicyScope,
	insertAllowed bool, insertScope rbac.PolicyScope,
	updateAllowed bool, updateScope rbac.PolicyScope,
	deleteAllowed bool, deleteScope rbac.PolicyScope,
) *rbac.RowPolicy {
	rp := &rbac.RowPolicy{
		Select: rbac.SelectPolicy{Allowed: selectAllowed, Scope: selectScope},
		Insert: rbac.InsertPolicy{Allowed: insertAllowed, Scope: insertScope},
		Update: rbac.UpdatePolicy{Allowed: updateAllowed, Scope: updateScope},
		Delete: rbac.DeletePolicy{Allowed: deleteAllowed, Scope: deleteScope},
	}
	rp.Normalize()
	return rp
}

// ---- tests ----

func TestEndUserPermissionService_Resolve_TenantAdmin(t *testing.T) {
	svc := appmodelruntime.NewEndUserPermissionService(&stubRBACRepo{})
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "", "model-id")
	if err != nil {
		t.Fatal(err)
	}
	if perms != nil {
		t.Error("tenant admin should return nil permissions")
	}
}

func TestEndUserPermissionService_Resolve_NoPermissions(t *testing.T) {
	svc := appmodelruntime.NewEndUserPermissionService(&stubRBACRepo{permissions: nil})
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "user1", "model-id")
	if err != nil {
		t.Fatal(err)
	}
	// No permissions → all denied
	if perms == nil {
		t.Fatal("expected non-nil perms (empty = all denied)")
	}
	if perms.Select.Allowed || perms.Insert.Allowed || perms.Update.Allowed || perms.Delete.Allowed {
		t.Error("no rbac permissions should result in all denied")
	}
}

func TestEndUserPermissionService_Resolve_SelectAll(t *testing.T) {
	rp := makeRowPolicy(true, rbac.ScopeAll, false, "", false, "", false, "")
	_ = json.Marshal(rp) // ensure it's well-formed
	stub := &stubRBACRepo{
		permissions: []*rbac.EndUserPermission{
			{ModelID: "model-id", RowPolicy: rp},
		},
	}
	svc := appmodelruntime.NewEndUserPermissionService(stub)
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "user1", "model-id")
	if err != nil {
		t.Fatal(err)
	}
	if !perms.Select.Allowed {
		t.Error("expected Select.Allowed = true")
	}
	if perms.Select.IsSelf {
		t.Error("scope=all should produce IsSelf=false")
	}
	if perms.Insert.Allowed {
		t.Error("expected Insert.Allowed = false")
	}
}

func TestEndUserPermissionService_Resolve_SelectSelf(t *testing.T) {
	rp := makeRowPolicy(true, rbac.ScopeCustom, true, rbac.ScopeCustom, false, "", false, "")
	stub := &stubRBACRepo{
		permissions: []*rbac.EndUserPermission{
			{ModelID: "model-id", RowPolicy: rp},
		},
	}
	svc := appmodelruntime.NewEndUserPermissionService(stub)
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "user1", "model-id")
	if err != nil {
		t.Fatal(err)
	}
	if !perms.Select.IsSelf {
		t.Error("scope=custom should produce IsSelf=true")
	}
	if !perms.Insert.Allowed {
		t.Error("expected Insert.Allowed = true")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/app/modelruntime/... -run TestEndUserPermissionService -v 2>&1 | head -20
```
Expected: 编译失败（`NewEndUserPermissionService` 未定义）

- [ ] **Step 3: 实现 `permission_service.go`**

```go
// modelcraft-backend/internal/app/modelruntime/permission_service.go
package modelruntime

import (
	"context"

	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rbac"
)

// endUserPermissionServiceImpl 实现 modelruntime.EndUserPermissionService。
// 依赖 rbac.EndUserPermissionRepository，per-request 按 model 粒度查权限。
type endUserPermissionServiceImpl struct {
	rbacRepo rbac.EndUserPermissionRepository
}

// NewEndUserPermissionService 创建 EndUserPermissionService 实例。
func NewEndUserPermissionService(rbacRepo rbac.EndUserPermissionRepository) modelruntime.EndUserPermissionService {
	return &endUserPermissionServiceImpl{rbacRepo: rbacRepo}
}

// Resolve 查询 end-user 在指定 model 上的有效权限，返回权限快照。
// endUserID 为空时（tenant admin）直接返回 nil, nil。
func (s *endUserPermissionServiceImpl) Resolve(
	ctx context.Context, orgName, projectSlug, endUserID, modelID string,
) (*modelruntime.ResolvedModelPermissions, error) {
	if endUserID == "" {
		return nil, nil
	}

	permissions, err := s.rbacRepo.FindPermissionsByEndUserAndModel(ctx, orgName, projectSlug, endUserID, modelID)
	if err != nil {
		return nil, err
	}

	// 合并所有权限点（取各 action 最宽 rowScope）
	eps := rbac.EffectivePermissionSet{}.Merge(permissions)
	return toResolvedModelPermissions(eps, modelID), nil
}

// toResolvedModelPermissions 将 EffectivePermissionSet 映射为 ResolvedModelPermissions。
// rowScope 映射规则：
//   - RowScopeAll  → IsSelf=false（不注入 WHERE）
//   - RowScopeSelf → IsSelf=true（注入 WHERE <EndUserRef> = $endUserID）
func toResolvedModelPermissions(eps rbac.EffectivePermissionSet, modelID string) *modelruntime.ResolvedModelPermissions {
	return &modelruntime.ResolvedModelPermissions{
		Select: toActionPermission(eps.GetPermission(modelID, rbac.ActionSelect)),
		Insert: toActionPermission(eps.GetPermission(modelID, rbac.ActionInsert)),
		Update: toActionPermission(eps.GetPermission(modelID, rbac.ActionUpdate)),
		Delete: toActionPermission(eps.GetPermission(modelID, rbac.ActionDelete)),
	}
}

func toActionPermission(ep *rbac.EffectivePermission) modelruntime.ActionPermission {
	if ep == nil {
		return modelruntime.ActionPermission{Allowed: false}
	}
	return modelruntime.ActionPermission{
		Allowed: true,
		IsSelf:  ep.RowScope == rbac.RowScopeSelf,
	}
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd modelcraft-backend && go test ./internal/app/modelruntime/... -run TestEndUserPermissionService -v
```
Expected: PASS (4 tests)

- [ ] **Step 5: 提交**

```bash
cd modelcraft-backend && git add internal/app/modelruntime/permission_service.go internal/app/modelruntime/permission_service_test.go
git commit -m "feat(app/modelruntime): implement EndUserPermissionService with RBAC query and mapping"
```

---

## Task 5: 修改 `graphql_app.go` — 注入 permService，Execute() 权限解析

**Files:**
- Modify: `modelcraft-backend/internal/app/modelruntime/graphql_app.go`
- Modify: `modelcraft-backend/internal/interfaces/http/routes.go`

**背景：** 当前 `Execute()` 流程是：`GetSchema()`（内部加载 model）→ 构造 clientRepo → `WithGraphqlRequestContext()`（无 EndUserPerms）→ `graphql.Do()`。

需要重构为：先加载 model（从 `GetSchema()` 中拆出，或新增 `loadModel()` 方法）→ 权限解析 → 构造含权限的 reqCtx → `graphql.Do()`。

同时 `WithGraphqlRequestContext` 签名已在 Task 2 变更，此处须补上最后一个参数。

- [ ] **Step 1: 修改 `GraphqlAppService` struct，新增 `permService` 字段**

在 `graphql_app.go` 中找到 struct 定义并新增字段：

```go
type GraphqlAppService struct {
	modelRepo            modelruntime.ModelRepository
	graphqlSchemaManager *modelruntime.GraphqlSchemaManager
	permService          modelruntime.EndUserPermissionService // 新增
}
```

- [ ] **Step 2: 修改 `NewGraphqlAppService` 构造函数，接受 permService 参数**

```go
// NewGraphqlAppService 创建graphql应用服务
func NewGraphqlAppService(
	modelRepo modelruntime.ModelRepository,
	lfkRepo modeldesign.LogicalForeignKeyRepository,
	permService modelruntime.EndUserPermissionService, // 新增
) *GraphqlAppService {
	schemaManager := modelruntime.NewGraphqlSchemaManager(modelRepo, lfkRepo)
	return &GraphqlAppService{
		modelRepo:            modelRepo,
		graphqlSchemaManager: schemaManager,
		permService:          permService, // 新增
	}
}
```

- [ ] **Step 3: 重构 `Execute()` — 先加载 model，再解析权限**

找到 `Execute()` 方法，在 `GetSchema()` 调用前加入 model 加载（`GetSchema()` 内部已做缓存，故直接利用 `modelRepo.GetByName()`），然后在 `WithGraphqlRequestContext` 前注入 `endUserPerms`：

```go
func (s *GraphqlAppService) Execute(ctx context.Context, orgName, projectSlug, name, databaseName string,
	cmd ExecuteGraphQLCommand,
) (*graphql.Result, error) {
	logger := logfacade.GetLogger(ctx)
	ctx = requestcontext.WithMetadata(ctx)

	modelLocator, err := modeldesign.NewModelLocator(orgName, projectSlug, databaseName, name)
	if err != nil {
		return nil, err
	}
	if err = s.denyManagedModelMutation(ctx, modelLocator, cmd); err != nil {
		return nil, err
	}

	// 加载 model（用于获取 model.ID 进行权限查询）
	model, err := s.modelRepo.GetByName(ctx, modelLocator)
	if err != nil {
		logger.Errorf(ctx, "get model fail: %v", err)
		return nil, fmt.Errorf("获取模型失败 %s", modelLocator.GetFullPath())
	}
	if model == nil {
		return nil, bizerrors.NewError(bizerrors.ModelNotFound, modelLocator.GetFullPath())
	}

	// 获取/构建 GraphQL Schema（内部会复用已缓存的 model 或重新加载，此处直接从 model 构建）
	gschema, err := s.graphqlSchemaManager.GetOrBuildSchema(ctx, model, modelLocator)
	if err != nil {
		return nil, err
	}

	// 创建请求级 DB 连接
	clientSqlDB, err := repository.DefaultClusterManager.GetConnectionWithDatabase(
		ctx, orgName, modelLocator.ProjectSlug, modelLocator.DatabaseName,
	)
	if err != nil {
		logger.Errorf(ctx, "get client sql db fail: %v", err)
		return nil, fmt.Errorf("获取客户端数据库失败 %s", databaseName)
	}
	clientRepo := dml.NewClientDB(clientSqlDB)

	// 提取 endUserID
	endUserID := ""
	if ctxutils.IsEndUser(ctx) {
		if uid, err := ctxutils.GetUserIDFromContext(ctx); err == nil {
			endUserID = uid
		}
	}

	// 解析 end-user 权限快照（tenant admin 时 permService.Resolve 返回 nil）
	endUserPerms, err := s.permService.Resolve(ctx, orgName, projectSlug, endUserID, model.ID)
	if err != nil {
		logger.Errorf(ctx, "resolve end-user permissions fail: %v", err)
		return nil, fmt.Errorf("解析权限失败")
	}

	reqCtx := modelruntime.WithGraphqlRequestContext(ctx, clientRepo, orgName, projectSlug, endUserID, endUserPerms)

	result := graphql.Do(graphql.Params{
		Schema:         *gschema,
		RequestString:  cmd.Query,
		VariableValues: cmd.Variables,
		Context:        reqCtx,
	})
	marshal, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	logger.Infof(ctx, "result=%+v", string(marshal))
	return result, nil
}
```

**注意：** `GetSchema()` 内部先查缓存，miss 时才加载 model。为避免重复加载，需检查 `GraphqlSchemaManager` 是否有 `GetOrBuildSchema(ctx, model, locator)` 方法，若没有则需新增或改用 `GetSchema()` 后再从缓存 model 取 ID。最简方案：先调 `GetSchema()`（会内部加载 model），再单独调 `modelRepo.GetByName()` 取 ID（接受一次额外查询，model 表数据量小）。

```go
// 简化版：先 GetSchema（可能触发 model 加载），再单独查 model.ID
gschema, err := s.GetSchema(ctx, orgName, modelLocator)
if err != nil {
    return nil, err
}
model, err := s.modelRepo.GetByName(ctx, modelLocator)
if err != nil || model == nil {
    return nil, bizerrors.NewError(bizerrors.ModelNotFound, modelLocator.GetFullPath())
}
```

- [ ] **Step 4: 修改 `routes.go` 中的 `CreateRuntimeHandlers`，注入 permService**

在 `modelcraft-backend/internal/interfaces/http/routes.go` 中找到 `CreateRuntimeHandlers`：

```go
func CreateRuntimeHandlers(loggingDB dbgen.Querier) *RuntimeHandlers {
	modelRuntimeRepo := repository.NewSqlModelRuntimeRepository(loggingDB)
	lfkRepo := repository.NewSqlLogicalForeignKeyRepository(loggingDB)
	permRepo := repository.NewSqlEndUserDataPermissionRepository(loggingDB) // 新增
	permService := appruntimeimport.NewEndUserPermissionService(permRepo)    // 新增（注意包导入名）
	graphqlAppService := modelruntime.NewGraphqlAppService(modelRuntimeRepo, lfkRepo, permService) // 新增第三个参数
	handler := runtimeHandler.NewModelRuntimeHandler(graphqlAppService)
	return &RuntimeHandlers{ModelRuntimeHandler: handler}
}
```

需在文件顶部导入 `app/modelruntime` 包。注意该包与 `domain/modelruntime` 包名相同，需用别名：

```go
import (
    // ...
    appruntimeimport "modelcraft/internal/app/modelruntime"
)
```

- [ ] **Step 5: 编译确认**

```bash
cd modelcraft-backend && go build ./... 2>&1 | head -30
```
Expected: 0 errors

- [ ] **Step 6: 提交**

```bash
cd modelcraft-backend && git add internal/app/modelruntime/graphql_app.go internal/interfaces/http/routes.go
git commit -m "feat(app/modelruntime): wire EndUserPermissionService into Execute() pipeline"
```

---

## Task 6: `model_resolver.go` — Action Gate + Row Filter

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/model_resolver.go`

**背景：** 需在 10 个 `execute*` 方法中分别加入 Action Gate（头部）和 Row Filter（WHERE 注入）。`endUserRefFieldName()` helper 遍历 `m.model.Fields` 找 `FormatEndUserRef`。

- [ ] **Step 1: 写集成测试（table-driven）**

在 `modelcraft-backend/internal/domain/modelruntime/` 目录下新建测试文件。由于 model_resolver 依赖 `graphqlRequestContext`，使用 `WithGraphqlRequestContext` 构造包含权限的 context 来验证 gate 行为。

```go
// modelcraft-backend/internal/domain/modelruntime/model_resolver_permission_test.go
package modelruntime_test

import (
	"context"
	"testing"

	"modelcraft/internal/domain/modelruntime"
)

// TestActionGate_DeniesWhenNoPermission 验证 end-user 无权限时各操作被拒绝。
// 使用 CheckAction 直接测试，不需要完整 GraphQL schema。
func TestActionGate_DeniesWhenNoPermission(t *testing.T) {
	// all denied
	perms := &modelruntime.ResolvedModelPermissions{}

	tests := []struct {
		action modelruntime.Action
	}{
		{modelruntime.ActionSelect},
		{modelruntime.ActionInsert},
		{modelruntime.ActionUpdate},
		{modelruntime.ActionDelete},
	}
	for _, tt := range tests {
		if err := perms.CheckAction(tt.action); err == nil {
			t.Errorf("action %s should be denied when no permissions", tt.action)
		}
	}
}

// TestActionGate_AllowsWhenPermitted
func TestActionGate_AllowsWhenPermitted(t *testing.T) {
	perms := &modelruntime.ResolvedModelPermissions{
		Select: modelruntime.ActionPermission{Allowed: true},
		Insert: modelruntime.ActionPermission{Allowed: true},
		Update: modelruntime.ActionPermission{Allowed: true},
		Delete: modelruntime.ActionPermission{Allowed: true},
	}
	for _, action := range []modelruntime.Action{
		modelruntime.ActionSelect, modelruntime.ActionInsert,
		modelruntime.ActionUpdate, modelruntime.ActionDelete,
	} {
		if err := perms.CheckAction(action); err != nil {
			t.Errorf("action %s should be allowed, got: %v", action, err)
		}
	}
}

// TestEndUserRefFieldName 验证 helper 正确找到 EndUserRef 字段名。
func TestEndUserRefFieldName(t *testing.T) {
	fieldName := modelruntime.FindEndUserRefFieldName(testModelWithEndUserRef())
	if fieldName != "owner_user" {
		t.Errorf("expected 'owner_user', got %q", fieldName)
	}
}

func TestEndUserRefFieldName_NoEndUserRef(t *testing.T) {
	fieldName := modelruntime.FindEndUserRefFieldName(testModelWithoutEndUserRef())
	if fieldName != "" {
		t.Errorf("expected empty string, got %q", fieldName)
	}
}

// TestRowFilter_InjectsWhere_WhenIsSelf verifies the row filter logic in isolation.
func TestRowFilter_BuildRowFilter(t *testing.T) {
	filter := modelruntime.BuildRowFilter(
		&modelruntime.ResolvedModelPermissions{
			Select: modelruntime.ActionPermission{Allowed: true, IsSelf: true},
		},
		modelruntime.ActionSelect,
		"owner_user",
		"user-abc",
	)
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}
	val, ok := filter["owner_user"]
	if !ok {
		t.Fatal("expected owner_user key in filter")
	}
	if val != "user-abc" {
		t.Errorf("expected 'user-abc', got %v", val)
	}
}

func TestRowFilter_NoFilter_WhenNotIsSelf(t *testing.T) {
	filter := modelruntime.BuildRowFilter(
		&modelruntime.ResolvedModelPermissions{
			Select: modelruntime.ActionPermission{Allowed: true, IsSelf: false},
		},
		modelruntime.ActionSelect,
		"owner_user",
		"user-abc",
	)
	if filter != nil {
		t.Error("should return nil filter when IsSelf=false")
	}
}

func TestRowFilter_NoFilter_WhenNoOwnerField(t *testing.T) {
	filter := modelruntime.BuildRowFilter(
		&modelruntime.ResolvedModelPermissions{
			Select: modelruntime.ActionPermission{Allowed: true, IsSelf: true},
		},
		modelruntime.ActionSelect,
		"", // no EndUserRef field
		"user-abc",
	)
	if filter != nil {
		t.Error("should return nil filter when no EndUserRef field")
	}
}
```

需在 `permission.go` 或新文件中导出两个 helper 供测试使用：`FindEndUserRefFieldName` 和 `BuildRowFilter`，以及测试 helper `testModelWithEndUserRef()`（需在同一 `_test.go` 包内）。

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/domain/modelruntime/... -run "TestActionGate|TestEndUserRefField|TestRowFilter" -v 2>&1 | head -30
```

- [ ] **Step 3: 在 `permission.go` 中新增导出 helper**

在 `modelcraft-backend/internal/domain/modelruntime/permission.go` 末尾追加：

```go
// FindEndUserRefFieldName 在 RuntimeModel 的字段中找到 FormatEndUserRef 类型字段的名称。
// 若模型无此类型字段，返回空字符串（降级为 ALL，不注入 WHERE）。
func FindEndUserRefFieldName(fields map[string]*RuntimeField) string {
	for _, f := range fields {
		if f.Type != nil && f.Type.Format == modeldesign.FormatEndUserRef {
			return f.Name
		}
	}
	return ""
}

// BuildRowFilter 根据权限和 action 构造 WHERE 注入 map。
// 若不需要注入（IsSelf=false 或 ownerField 为空），返回 nil。
func BuildRowFilter(
	perms *ResolvedModelPermissions,
	action Action,
	ownerField string,
	endUserID string,
) map[string]any {
	if perms == nil || ownerField == "" || endUserID == "" {
		return nil
	}
	if !perms.Get(action).IsSelf {
		return nil
	}
	return map[string]any{ownerField: endUserID}
}
```

在 `permission.go` 顶部 import 中加入：

```go
"modelcraft/internal/domain/modeldesign"
```

在测试文件中添加 helper：

```go
func testModelWithEndUserRef() map[string]*modelruntime.RuntimeField {
	return map[string]*modelruntime.RuntimeField{
		"owner_user": {
			Name: "owner_user",
			Type: &modeldesign.FieldType{Format: modeldesign.FormatEndUserRef},
		},
		"title": {
			Name: "title",
			Type: &modeldesign.FieldType{Format: modeldesign.FormatString},
		},
	}
}

func testModelWithoutEndUserRef() map[string]*modelruntime.RuntimeField {
	return map[string]*modelruntime.RuntimeField{
		"title": {Name: "title", Type: &modeldesign.FieldType{Format: modeldesign.FormatString}},
	}
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd modelcraft-backend && go test ./internal/domain/modelruntime/... -run "TestActionGate|TestEndUserRefField|TestRowFilter" -v
```
Expected: PASS

- [ ] **Step 5: 修改 `model_resolver.go` — 新增 `endUserRefFieldName()` 方法**

在 `graphqlModelResolver` 的方法中新增（找到文件中其他 helper 方法附近）：

```go
// endUserRefFieldName 返回此 model 中 FormatEndUserRef 类型字段的名称。
// 空字符串表示该 model 无 owner 字段，SELF scope 降级为 ALL。
func (m *graphqlModelResolver) endUserRefFieldName() string {
	return FindEndUserRefFieldName(m.model.Fields)
}
```

- [ ] **Step 6: 修改 `executeCreateOne` — 新增 Action Gate**

在 `executeCreateOne` 方法的 `rctx, _ := getGraphqlRequestContext(p.Context)` 行之后，`input, err := newCreateOneInput(...)` 之前插入：

```go
if err := rctx.EndUserPerms.CheckAction(ActionInsert); err != nil {
    return nil, err
}
```

- [ ] **Step 7: 修改 `executeUpdateOne` — Action Gate + Row Filter**

在 `rctx, _ := getGraphqlRequestContext(p.Context)` 之后插入：

```go
if err := rctx.EndUserPerms.CheckAction(ActionUpdate); err != nil {
    return nil, err
}
```

在 `input, err := newUpdateOneInput(...)` 之后，现有的 force-inject endUserRef 逻辑之后，追加 Row Filter：

```go
// Row filter: inject WHERE <EndUserRef> = endUserID when rowScope=SELF
if rf := BuildRowFilter(rctx.EndUserPerms, ActionUpdate, m.endUserRefFieldName(), rctx.CurrentEndUserID); rf != nil {
    for k, v := range rf {
        input.Where[k] = v
    }
}
```

- [ ] **Step 8: 修改 `executeDeleteOne` — Action Gate + Row Filter**

在 `rctx, _ := getGraphqlRequestContext(p.Context)` 之后插入：

```go
if err := rctx.EndUserPerms.CheckAction(ActionDelete); err != nil {
    return nil, err
}
```

在 `input, err := newDeleteOneInput(...)` 之后追加：

```go
if rf := BuildRowFilter(rctx.EndUserPerms, ActionDelete, m.endUserRefFieldName(), rctx.CurrentEndUserID); rf != nil {
    for k, v := range rf {
        input.Where[k] = v
    }
}
```

- [ ] **Step 9: 修改 `executeFindUnique` — Action Gate + Row Filter**

在 `rctx, _ := getGraphqlRequestContext(p.Context)` 之后插入：

```go
if err := rctx.EndUserPerms.CheckAction(ActionSelect); err != nil {
    return nil, err
}
```

在 `input, err := newFindUniqueInput(...)` 之后追加：

```go
if rf := BuildRowFilter(rctx.EndUserPerms, ActionSelect, m.endUserRefFieldName(), rctx.CurrentEndUserID); rf != nil {
    for k, v := range rf {
        input.Where[k] = v
    }
}
```

- [ ] **Step 10: 修改 `executeFindFirst` — Action Gate + Row Filter**

同 Step 9 模式，操作 `FindFirstInput`：

```go
// 在 rctx 行之后
if err := rctx.EndUserPerms.CheckAction(ActionSelect); err != nil {
    return nil, err
}
// 在 input 构造之后
if rf := BuildRowFilter(rctx.EndUserPerms, ActionSelect, m.endUserRefFieldName(), rctx.CurrentEndUserID); rf != nil {
    for k, v := range rf {
        input.Where[k] = v
    }
}
```

- [ ] **Step 11: 修改 `executeFindMany` — Action Gate + Row Filter**

同 Step 9 模式，操作 `FindManyInput`：

```go
if err := rctx.EndUserPerms.CheckAction(ActionSelect); err != nil {
    return nil, err
}
// after newFindManyInput
if rf := BuildRowFilter(rctx.EndUserPerms, ActionSelect, m.endUserRefFieldName(), rctx.CurrentEndUserID); rf != nil {
    for k, v := range rf {
        input.Where[k] = v
    }
}
```

- [ ] **Step 12: 修改 `executeAggregate` 和 `executeCount` — Action Gate**

```go
// executeAggregate: after rctx
if err := rctx.EndUserPerms.CheckAction(ActionSelect); err != nil {
    return nil, err
}
// 注意：aggregate/count 不注入 Row Filter（聚合语义不同，v1 范围外）

// executeCount: after rctx
if err := rctx.EndUserPerms.CheckAction(ActionSelect); err != nil {
    return nil, err
}
```

- [ ] **Step 13: 修改 `executeCreateMany` — Action Gate**

在 `rctx, _ := getGraphqlRequestContext(p.Context)` 之后插入：

```go
if err := rctx.EndUserPerms.CheckAction(ActionInsert); err != nil {
    return nil, err
}
```

- [ ] **Step 14: 修改 `executeUpdateMany` — Action Gate + Row Filter**

```go
if err := rctx.EndUserPerms.CheckAction(ActionUpdate); err != nil {
    return nil, err
}
// after newUpdateManyInput, after existing force-inject EndUserRef:
if rf := BuildRowFilter(rctx.EndUserPerms, ActionUpdate, m.endUserRefFieldName(), rctx.CurrentEndUserID); rf != nil {
    for k, v := range rf {
        input.Where[k] = v
    }
}
```

- [ ] **Step 15: 修改 `executeDeleteMany` — Action Gate + Row Filter**

```go
if err := rctx.EndUserPerms.CheckAction(ActionDelete); err != nil {
    return nil, err
}
// after newDeleteManyInput:
if rf := BuildRowFilter(rctx.EndUserPerms, ActionDelete, m.endUserRefFieldName(), rctx.CurrentEndUserID); rf != nil {
    for k, v := range rf {
        input.Where[k] = v
    }
}
```

- [ ] **Step 16: 编译确认**

```bash
cd modelcraft-backend && go build ./internal/domain/modelruntime/... 2>&1 | head -20
```
Expected: 0 errors

- [ ] **Step 17: 运行所有 modelruntime 测试**

```bash
cd modelcraft-backend && go test ./internal/domain/modelruntime/... -v 2>&1 | tail -20
```
Expected: PASS

- [ ] **Step 18: 提交**

```bash
cd modelcraft-backend && git add internal/domain/modelruntime/permission.go internal/domain/modelruntime/model_resolver.go internal/domain/modelruntime/model_resolver_permission_test.go
git commit -m "feat(modelruntime): add action gate and row filter to all execute* resolvers"
```

---

## Task 7: 全量编译 + 回归测试

**Files:** 无新文件，验证整体

- [ ] **Step 1: 全量编译**

```bash
cd modelcraft-backend && go build ./... 2>&1
```
Expected: 0 errors

- [ ] **Step 2: 运行全量测试**

```bash
cd modelcraft-backend && go test ./... 2>&1 | tail -30
```
Expected: 无新增 FAIL

- [ ] **Step 3: Lint 检查**

```bash
cd modelcraft-backend && just lint 2>&1 | tail -20
```
Expected: 无新增 lint error（如有，运行 `just lint-fix`）

- [ ] **Step 4: 提交 lint 修复（如有）**

```bash
cd modelcraft-backend && git add -p && git commit -m "fix(modelruntime): lint fixes for data permission implementation"
```
