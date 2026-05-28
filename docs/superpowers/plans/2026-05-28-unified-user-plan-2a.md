# Unified User System — Plan 2a: Domain + Repository Layer

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复编译错误，将 `User` domain 加入 `IsAdmin` 字段，更新 `Membership` domain 使其与新 `user_orgs` schema 对齐，新增 `UserOrgRepository` 接口，更新 `PlatformClaims` 加 `is_admin` claim，让整个项目重新编译通过。

**Architecture:** Plan 1 已完成 DB schema 重命名（`user_organizations` → `user_orgs`，加 `is_admin`）。本 Plan 修复上层 Go 代码与新 schema 的断层：`safe_querier_gen.go` 引用已删除的 dbgen 类型 → 重新生成；`MembershipToDomain` 用了已不存在的 `dbgen.UserOrganization` → 改为 `dbgen.UserOrg`；`Membership` domain 有邀请字段而新表无 → 清理；`User` domain 无 `IsAdmin` → 加上；`PlatformClaims` 无 `is_admin` → 加上。

**Tech Stack:** Go, sqlc/gowrap, JWT

**编译错误现状（Plan 1 后）：**
- `internal/infrastructure/dbgenwrap/safe_querier_gen.go` — 11 个错误，引用已删除的 `dbgen.EndUserRole`、`dbgen.UserOrganization`、`dbgen.IsEndUserBuiltinParams`、`GetBuiltinEndUserByOrg` 等
- 根因：`safe_querier_gen.go` 是 gowrap 自动生成的，只需重跑 `just generate-safe-querier` 即可同步

---

## 文件变更地图

| 操作 | 文件 |
|------|------|
| **重新生成** | `internal/infrastructure/dbgenwrap/safe_querier_gen.go` — 运行 `just generate-safe-querier` |
| **修改** | `internal/domain/membership/membership.go` — 移除 `InvitedBy`、`InvitedAt`、`JoinedAt`，加 `IsAdmin` |
| **修改** | `internal/domain/membership/membership_test.go` — 更新测试 |
| **修改** | `internal/infrastructure/repository/sql_org_repository.go` — `MembershipToDomain` 改用 `dbgen.UserOrg`，移除邀请字段转换，加 `IsAdmin` |
| **修改** | `internal/domain/user/user.go` — `User` struct 加 `IsAdmin bool` 字段 |
| **修改** | `internal/domain/auth/platform_claims.go` — `PlatformClaims` 加 `IsAdmin bool` |
| **修改** | `internal/app/organization/create_organization_service.go` — 移除 `maybeCreateBuiltinAdmin` 相关逻辑 |
| **删除** | `internal/infrastructure/repository/sql_enduser_repository.go` — 废弃，改为存根或删除 |
| **修改** | `internal/domain/enduser/end_user_repository.go` — 标记废弃（Plan 2b 彻底删除） |

---

## Task 1: 重新生成 safe_querier_gen.go 修复编译错误

**Files:**
- Regenerate: `internal/infrastructure/dbgenwrap/safe_querier_gen.go`

- [ ] **Step 1: 运行 generate-safe-querier**

```bash
cd modelcraft-backend && just generate-safe-querier
```

预期：无错误输出，`safe_querier_gen.go` 更新。

- [ ] **Step 2: 验证编译错误减少**

```bash
cd modelcraft-backend && go build ./internal/infrastructure/dbgenwrap/... 2>&1
```

预期：编译通过（此包单独编译应无错误）。

- [ ] **Step 3: 查看剩余全局编译错误**

```bash
cd modelcraft-backend && go build ./... 2>&1 | head -60
```

记录剩余错误，预期：`safe_querier_gen.go` 的 11 个错误消失，可能还有其他层的错误。

- [ ] **Step 4: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/infrastructure/dbgenwrap/safe_querier_gen.go
git commit -m "chore: regenerate safe querier wrapper after schema rename"
```

---

## Task 2: 更新 Membership domain — 移除邀请字段，加 IsAdmin

**Files:**
- Modify: `modelcraft-backend/internal/domain/membership/membership.go`
- Modify: `modelcraft-backend/internal/domain/membership/membership_test.go`

`user_orgs` 表已移除 `invited_by`、`invited_at`、`joined_at` 字段（邀请流程废弃）。同时新增 `is_admin` 字段。Membership domain 需对应更新。

- [ ] **Step 1: 读取当前文件**

```bash
cat modelcraft-backend/internal/domain/membership/membership.go
cat modelcraft-backend/internal/domain/membership/membership_test.go
```

- [ ] **Step 2: 更新 membership.go**

将 `Membership` struct 改为：

```go
// Membership 用户-组织关联实体
// 角色信息通过 user_roles 表管理，不在此实体中存储
type Membership struct {
    ID        string           // UUID
    UserID    string           // 用户 ID
    OrgName   string           // 组织名称（主键引用）
    IsAdmin   bool             // 是否为管理员（可访问管理后台）
    Status    MembershipStatus // 成员状态：active | suspended
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

更新 `Validate()`：移除 `MembershipStatusInvited` 检查：

```go
func (m *Membership) Validate() error {
    if m.ID == "" {
        return fmt.Errorf("membership ID is required")
    }
    if m.UserID == "" {
        return fmt.Errorf("user ID is required")
    }
    if m.OrgName == "" {
        return fmt.Errorf("organization name is required")
    }
    if m.Status != MembershipStatusActive &&
        m.Status != MembershipStatusSuspended {
        return fmt.Errorf("membership status must be one of: active, suspended")
    }
    return nil
}
```

更新常量（移除 `MembershipStatusInvited`）：

```go
const (
    MembershipStatusActive    MembershipStatus = "active"
    MembershipStatusSuspended MembershipStatus = "suspended"
)
```

更新 `NewMembership` — 移除邀请字段，加 `IsAdmin` 参数：

```go
// NewMembership 创建成员关系（直接加入，状态为 active）
func NewMembership(id, userID, orgName string, isAdmin bool) (*Membership, error) {
    now := time.Now()
    m := &Membership{
        ID:        id,
        UserID:    userID,
        OrgName:   orgName,
        IsAdmin:   isAdmin,
        Status:    MembershipStatusActive,
        CreatedAt: now,
        UpdatedAt: now,
    }
    if err := m.Validate(); err != nil {
        return nil, err
    }
    return m, nil
}
```

删除 `NewInvitation`、`AcceptInvitation` 方法（邀请流程废弃）。

保留 `Suspend()`、`IsActive()` 方法。

- [ ] **Step 3: 更新 membership_test.go**

将测试更新为：

```go
package membership_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "modelcraft/internal/domain/membership"
)

func TestNewMembership(t *testing.T) {
    m, err := membership.NewMembership("mid-001", "user-001", "my-org", false)
    require.NoError(t, err)
    assert.Equal(t, "mid-001", m.ID)
    assert.Equal(t, "user-001", m.UserID)
    assert.Equal(t, "my-org", m.OrgName)
    assert.False(t, m.IsAdmin)
    assert.Equal(t, membership.MembershipStatusActive, m.Status)
    assert.NotZero(t, m.CreatedAt)
}

func TestNewMembershipAdmin(t *testing.T) {
    m, err := membership.NewMembership("mid-002", "user-002", "my-org", true)
    require.NoError(t, err)
    assert.True(t, m.IsAdmin)
    assert.Equal(t, membership.MembershipStatusActive, m.Status)
}

func TestMembershipValidate_RequiredFields(t *testing.T) {
    _, err := membership.NewMembership("", "user-001", "my-org", false)
    assert.Error(t, err)

    _, err = membership.NewMembership("mid-001", "", "my-org", false)
    assert.Error(t, err)

    _, err = membership.NewMembership("mid-001", "user-001", "", false)
    assert.Error(t, err)
}
```

- [ ] **Step 4: 运行测试**

```bash
cd modelcraft-backend && go test ./internal/domain/membership/... -v
```

预期：全部通过。

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/domain/membership/
git commit -m "domain: update Membership entity — remove invite fields, add IsAdmin"
```

---

## Task 3: 更新 MembershipToDomain 和 CreateMembership 适配 user_orgs

**Files:**
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_org_repository.go`

`MembershipToDomain` 当前使用 `dbgen.UserOrganization`（已不存在），需改为 `dbgen.UserOrg`（新类型），同时移除邀请字段转换，加 `IsAdmin` 字段。

- [ ] **Step 1: 读取 sql_org_repository.go 中的 MembershipToDomain**

```bash
grep -n "MembershipToDomain\|InvitedBy\|InvitedAt\|JoinedAt\|UserOrganization\|CreateMembership" \
  modelcraft-backend/internal/infrastructure/repository/sql_org_repository.go
```

- [ ] **Step 2: 将 MembershipToDomain 函数替换为**

```go
// MembershipToDomain converts a dbgen.UserOrg row to a domain Membership entity.
func MembershipToDomain(row dbgen.UserOrg) *membership.Membership {
    return &membership.Membership{
        ID:        row.ID,
        UserID:    row.UserID,
        OrgName:   row.OrgName,
        IsAdmin:   row.IsAdmin != 0,
        Status:    membership.MembershipStatus(row.Status),
        CreatedAt: row.CreatedAt,
        UpdatedAt: row.UpdatedAt,
    }
}
```

- [ ] **Step 3: 更新 CreateMembership 调用（移除邀请字段）**

在 `SqlOrganizationRepository` 或 membership repository 中找到 `CreateMembership` 的参数构建，检查是否引用了已移除的 `invited_by`、`invited_at`、`joined_at` 字段：

```bash
grep -n "InvitedBy\|InvitedAt\|JoinedAt\|invited_by\|invited_at\|joined_at" \
  modelcraft-backend/internal/infrastructure/repository/sql_org_repository.go
```

如有引用，将对应的 `dbgen.CreateMembershipParams` 字段移除（schema 已无这些列，dbgen 生成时已不包含）。

- [ ] **Step 4: 检查 MembershipRepository 接口中是否有邀请相关方法**

```bash
cat modelcraft-backend/internal/domain/membership/repository.go 2>/dev/null || \
  grep -rn "MembershipRepository\|NewInvitation\|AcceptInvitation" \
  modelcraft-backend/internal/domain/ modelcraft-backend/internal/app/ --include="*.go" -l
```

如有 `NewInvitation`、`AcceptInvitation` 方法调用，标记为 TODO（Plan 2b 彻底清理）。

- [ ] **Step 5: 编译检查**

```bash
cd modelcraft-backend && go build ./internal/infrastructure/repository/... 2>&1
```

预期：无 `UserOrganization`、`InvitedBy` 相关错误。

- [ ] **Step 6: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/infrastructure/repository/sql_org_repository.go
git commit -m "repo: update MembershipToDomain to use dbgen.UserOrg, add IsAdmin"
```

---

## Task 4: User domain 加 IsAdmin，更新 UserToDomain

**Files:**
- Modify: `modelcraft-backend/internal/domain/user/user.go`
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_org_repository.go` (UserToDomain 部分)

`User` 实体本身不持有 `IsAdmin`（IsAdmin 存在 `user_orgs` 表，不在 `users` 表）。但认证时需要把 `user_orgs.is_admin` 加进 token。这个值通过 `Membership.IsAdmin` 传递，不需要在 `User` struct 上加字段。

本 Task 确认这个设计判断，并在 `PlatformClaims` 上加 `IsAdmin`。

- [ ] **Step 1: 确认 User struct 不需要 IsAdmin 字段**

读取 `internal/domain/user/user.go` 确认 `User` 只含账号信息（无 org 相关字段），`IsAdmin` 通过 `Membership` 对象传递到 token 层。这是正确的 DDD 设计——`IsAdmin` 是 org 上下文的属性，不属于用户实体本身。

```bash
cat modelcraft-backend/internal/domain/user/user.go
```

无需修改 `User` struct。

- [ ] **Step 2: 更新 PlatformClaims 加 IsAdmin**

修改 `modelcraft-backend/internal/domain/auth/platform_claims.go`：

```go
// PlatformClaims 是统一 JWT 的 payload 结构，适用于所有用户类型（tenant / end-user）。
type PlatformClaims struct {
    UserID  string `json:"user_id"`
    OrgName string `json:"org_name"`
    IsAdmin bool   `json:"is_admin"`
    Key     string `json:"key"` // APISIX jwt-auth Consumer key
    jwt.RegisteredClaims
}
```

注意：移除 `EndUserAdminIDs`（合并用户体系后不再需要这个字段）。

- [ ] **Step 3: 检查 EndUserAdminIDs 的使用方**

```bash
grep -rn "EndUserAdminIDs\|end_user_admin_ids" \
  modelcraft-backend/internal/ --include="*.go"
```

如有使用，标记为 TODO（Plan 2b 清理）或直接移除（如果只是 token 填充代码）。

- [ ] **Step 4: 编译检查**

```bash
cd modelcraft-backend && go build ./internal/domain/auth/... 2>&1
```

预期：编译通过。

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/domain/auth/platform_claims.go
git commit -m "domain: add IsAdmin to PlatformClaims, remove EndUserAdminIDs"
```

---

## Task 5: 移除 create_organization_service.go 中的 builtin admin 逻辑

**Files:**
- Modify: `modelcraft-backend/internal/app/organization/create_organization_service.go`

设计决策：builtin end-user 概念已废弃，创建组织时不再需要创建内置管理员账号。

- [ ] **Step 1: 读取文件，找到 maybeCreateBuiltinAdmin 相关代码**

```bash
grep -n "builtin\|Builtin\|EndUserRepo\|EndUserRepoFactory\|maybeCreateBuiltin" \
  modelcraft-backend/internal/app/organization/create_organization_service.go
```

- [ ] **Step 2: 移除 builtin admin 相关代码**

具体步骤：
1. 移除 `EndUserRepoFactory` 字段（如果存在于 service struct）
2. 移除 `maybeCreateBuiltinAdmin` 方法（如果存在）
3. 移除 `Execute` 方法中调用 `maybeCreateBuiltinAdmin` 的代码块
4. 移除相关 import（`enduser` package）

- [ ] **Step 3: 编译检查**

```bash
cd modelcraft-backend && go build ./internal/app/organization/... 2>&1
```

预期：编译通过，无 `enduser` package 引用。

- [ ] **Step 4: 运行现有测试**

```bash
cd modelcraft-backend && go test ./internal/app/organization/... -v 2>&1 | tail -20
```

预期：测试通过（builtin admin 相关测试可能需要删除，如有编译错误先修复）。

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/app/organization/
git commit -m "app: remove builtin end-user creation from create_organization_service"
```

---

## Task 6: 清理 sql_enduser_repository.go 编译错误

**Files:**
- Modify or Delete: `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go`
- Modify or Delete: `modelcraft-backend/internal/infrastructure/repository/sql_enduser_session_repository.go`

这两个文件引用了已不存在的 dbgen 类型（`dbgen.EndUser`、`dbgen.GetEndUserByIDParams` 等）。Plan 2b 会彻底重写这些逻辑，本 Task 只需让它们编译通过。

策略：将这两个文件内容替换为最小存根，保留接口签名但 body 返回 `errors.New("not implemented")`。

- [ ] **Step 1: 查看当前编译错误**

```bash
cd modelcraft-backend && go build ./internal/infrastructure/repository/... 2>&1 | grep "sql_enduser"
```

- [ ] **Step 2: 读取 sql_enduser_repository.go 开头，了解它实现的接口**

```bash
head -30 modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go
```

- [ ] **Step 3: 用最小存根替换 sql_enduser_repository.go**

将文件替换为能通过编译的最小存根（保留 `NewSqlEndUserRepository` 构造函数签名，body 用 panic 或 error 占位）：

```go
// Package repository — sql_enduser_repository.go
// DEPRECATED: EndUser repository is being replaced by unified User repository.
// This file is a stub to maintain compilation during Plan 2a.
// Will be fully replaced in Plan 2b.
package repository

import (
    "context"
    "errors"

    "modelcraft/internal/domain/enduser"
    "modelcraft/internal/infrastructure/dbgen"
)

var errEndUserRepoDeprecated = errors.New("EndUserRepository is deprecated; use UserRepository instead")

// SqlEndUserRepository is a deprecated stub.
// TODO(Plan 2b): Replace with unified user repository.
type SqlEndUserRepository struct{}

// NewSqlEndUserRepository returns a deprecated stub.
func NewSqlEndUserRepository(_ dbgen.Querier) enduser.EndUserRepository {
    return &SqlEndUserRepository{}
}

func (r *SqlEndUserRepository) Save(_ context.Context, _ *enduser.EndUser) error {
    return errEndUserRepoDeprecated
}
func (r *SqlEndUserRepository) GetByID(_ context.Context, _, _ string) (*enduser.EndUser, error) {
    return nil, errEndUserRepoDeprecated
}
func (r *SqlEndUserRepository) GetByUsername(_ context.Context, _, _ string) (*enduser.EndUser, error) {
    return nil, errEndUserRepoDeprecated
}
func (r *SqlEndUserRepository) GetBuiltinByOrg(_ context.Context, _ string) (*enduser.EndUser, error) {
    return nil, nil // builtin concept abolished
}
func (r *SqlEndUserRepository) UpdateStatus(_ context.Context, _, _ string, _ bool) error {
    return errEndUserRepoDeprecated
}
func (r *SqlEndUserRepository) UpdatePassword(_ context.Context, _, _ string, _ enduser.HashedPassword) error {
    return errEndUserRepoDeprecated
}
func (r *SqlEndUserRepository) Delete(_ context.Context, _, _ string) error {
    return errEndUserRepoDeprecated
}
func (r *SqlEndUserRepository) ListWithTotal(_ context.Context, _ enduser.ListEndUsersQuery) ([]*enduser.EndUser, int64, error) {
    return nil, 0, errEndUserRepoDeprecated
}
func (r *SqlEndUserRepository) ListAccessibleProjectsByRoleAssignment(_ context.Context, _, _ string) ([]enduser.AccessibleProject, error) {
    return nil, errEndUserRepoDeprecated
}
func (r *SqlEndUserRepository) HasProjectAccessByRole(_ context.Context, _, _, _ string) (bool, error) {
    return false, errEndUserRepoDeprecated
}
```

- [ ] **Step 4: 用最小存根替换 sql_enduser_session_repository.go**

```bash
# 先查看它实现的接口
head -30 modelcraft-backend/internal/infrastructure/repository/sql_enduser_session_repository.go
grep -n "type.*Session\|SessionRepository" \
  modelcraft-backend/internal/domain/enduser/end_user_session_repository.go 2>/dev/null | head -10
```

类似 Task 6 Step 3，将 session repository 也替换为编译通过的存根。

- [ ] **Step 5: 删除已无法编译的旧 repository 测试文件**

```bash
# 检查是否有测试文件引用了已删除的 dbgen 类型
cd modelcraft-backend && go build ./internal/infrastructure/repository/... 2>&1 | grep "_test.go"
```

如有测试文件引用了已删除类型，直接删除这些测试文件（Plan 2b 重写时重新写测试）：

```bash
# 例如：
rm modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository_test.go
```

- [ ] **Step 6: 编译整个 repository 层**

```bash
cd modelcraft-backend && go build ./internal/infrastructure/repository/... 2>&1
```

预期：编译通过。

- [ ] **Step 7: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/infrastructure/repository/
git commit -m "repo: stub out EndUser repositories pending Plan 2b rewrite"
```

---

## Task 7: 全局编译通过验证

- [ ] **Step 1: 全量编译**

```bash
cd modelcraft-backend && go build ./... 2>&1
```

预期：编译通过，0 errors。如有剩余错误，逐一修复。

- [ ] **Step 2: 处理剩余编译错误（如有）**

常见剩余错误类型和处理方式：
- `enduser.NewInvitation` / `AcceptInvitation` 调用 → 找到调用方，注释或删除
- `dbgen.UserOrganization` 类型引用（非 repository 层）→ 改为 `dbgen.UserOrg`
- `MembershipStatusInvited` 引用 → 删除或改为 `MembershipStatusActive`
- `EndUserAdminIDs` 引用 → 删除（token 填充代码）

每修复一类错误就重新编译确认。

- [ ] **Step 3: 运行单元测试**

```bash
cd modelcraft-backend && go test ./internal/domain/... ./internal/app/... 2>&1 | tail -30
```

预期：domain 层和 app 层测试通过（repository 层测试因为存根可能有部分跳过，属正常）。

- [ ] **Step 4: 最终 commit**

```bash
cd /data/home/lukemxjia/modelcraft
git status
# 如有未提交变更：
git add -A && git commit -m "chore: Plan 2a complete — compilation restored, domain aligned with new schema"
```
