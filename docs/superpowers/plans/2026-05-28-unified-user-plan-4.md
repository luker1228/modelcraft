# Unified User System — Plan 4: GraphQL 新用户管理 API

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 GraphQL API 基础上新增统一用户管理接口：`createUser`（含 isAdmin）、`users` query、`myProjects` query。不修改/删除现有 `createEndUser`、`endUserProjects`、`isBuiltin`、`createdBy`（保持向后兼容）。

**Architecture:** 在 `api/graph/org/schema/end_user.graphql` 追加新类型/查询/变更 → `just generate-gql` 重新生成代码 → 在 `end_user.resolvers.go` 实现新 resolver → 同步更新 app service layer。

**Tech Stack:** Go, gqlgen, GraphQL

**当前状态：**
- 现有 `createEndUser`, `updateEndUserStatus`, `deleteEndUser`, `resetEndUserPassword`, `endUserProjects` — 保持不变
- `EndUser.isBuiltin`, `EndUser.createdBy` — 保持不变（resolver 已返回 false/nil）
- 新增 `createUser(username, password, isAdmin)`, `users(...)`, `myProjects` — 待实现

---

## 文件变更地图

| 操作 | 文件 |
|------|------|
| **修改** | `api/graph/org/schema/end_user.graphql` — 追加新 mutation/query |
| **重新生成** | `internal/interfaces/graphql/org/generated/generated.go` — `just generate-gql` |
| **修改** | `internal/interfaces/graphql/org/end_user.resolvers.go` — 实现新 resolver |
| **修改** | `internal/app/enduser/end_user_app_service.go` — 新增 CreateUser（含 isAdmin）方法 |
| **修改** | `internal/app/enduser/commands.go` — 新增 CreateUserCommand |
| **修改** | `internal/interfaces/graphql/org/resolver.go` — 注入新依赖（如需要） |

---

## Task 1: 更新 GraphQL Schema — 追加新类型和操作

**Files:**
- Modify: `modelcraft-backend/api/graph/org/schema/end_user.graphql`

- [ ] **Step 1: 读取当前 end_user.graphql**

```bash
cat modelcraft-backend/api/graph/org/schema/end_user.graphql
```

- [ ] **Step 2: 在文件末尾追加以下内容**

```graphql
# ============================================
# Unified User Management (统一用户体系)
# 以下 API 基于统一 users + user_orgs 表，
# 与旧 createEndUser 并存，新代码优先使用这些接口。
# ============================================

# CreateUser error union（管理员创建用户）
union CreateUserError = EndUserAlreadyExists | EndUserPasswordTooWeak | InvalidInput

type CreateUserPayload {
  user: EndUser
  error: CreateUserError
}

input CreateUserInput {
  username: String!  # 3–64 chars, ^[a-zA-Z0-9_-]+$
  password: String!  # at least 8 chars, letter + digit
  isAdmin: Boolean!  # true = 管理员（访问管理后台），false = 普通用户
}

extend type Mutation {
  # createUser: 统一创建用户（管理员或普通用户）。
  # 仅管理员可调用（@hasPermission admin-only）。
  createUser(input: CreateUserInput!): CreateUserPayload! @hasPermission(action: "end-user:create")
}

extend type Query {
  # myProjects: 当前 end-user 可访问的 Project 列表（endUserProjects 的新名称）。
  # allowEndUser: true — end-user 可调用，tenant admin 不可调用（同 endUserProjects）。
  myProjects: [Project!]! @hasPermission(action: "end-user:read", allowEndUser: true)
}
```

- [ ] **Step 3: 验证 schema 文件语法（通过代码生成来验证）**

```bash
cd modelcraft-backend && just generate-gql 2>&1 | head -30
```

如有 schema 语法错误，根据错误信息修复。

- [ ] **Step 4: Commit schema 文件**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/api/graph/org/schema/end_user.graphql
git commit -m "schema(gql): add createUser mutation and myProjects query to org schema"
```

---

## Task 2: 重新生成 GraphQL 代码

**Files:**
- Regenerate: `modelcraft-backend/internal/interfaces/graphql/org/generated/generated.go`

- [ ] **Step 1: 运行代码生成**

```bash
cd modelcraft-backend && just generate-gql 2>&1
```

预期：生成成功，无 schema 错误。

- [ ] **Step 2: 验证新生成的接口**

```bash
grep -n "CreateUser\|MyProjects\|CreateUserInput\|CreateUserPayload" \
  modelcraft-backend/internal/interfaces/graphql/org/generated/generated.go | head -20
```

预期：出现 `CreateUser`、`MyProjects` 相关的 resolver 接口和类型定义。

- [ ] **Step 3: 编译检查（此时 resolver 可能有未实现的接口）**

```bash
cd modelcraft-backend && go build ./internal/interfaces/graphql/org/... 2>&1 | head -20
```

预期：出现类似 "does not implement OrgResolversResolver interface" 的错误，提示需要实现新 resolver 方法。

- [ ] **Step 4: Commit 生成代码**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/interfaces/graphql/org/generated/
git commit -m "chore: regenerate org graphql code for createUser + myProjects"
```

---

## Task 3: 新增 CreateUser app service 方法

**Files:**
- Modify: `modelcraft-backend/internal/app/enduser/commands.go`
- Modify: `modelcraft-backend/internal/app/enduser/end_user_app_service.go`

- [ ] **Step 1: 读取当前 commands.go 和 end_user_app_service.go**

```bash
cat modelcraft-backend/internal/app/enduser/commands.go
head -100 modelcraft-backend/internal/app/enduser/end_user_app_service.go
```

- [ ] **Step 2: 在 commands.go 中追加 CreateUserCommand 和 CreateUserResult**

在文件末尾追加：

```go
// CreateUserCommand 创建统一用户（管理员或普通用户）。
// 与 CreateEndUserCommand 的区别：包含 IsAdmin 字段。
type CreateUserCommand struct {
    OrgName  string
    Username string
    Password string
    IsAdmin  bool
}

// CreateUserResult 创建用户的返回结果。
type CreateUserResult struct {
    ID          string
    Username    string
    IsAdmin     bool
    IsForbidden bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

- [ ] **Step 3: 在 end_user_app_service.go 中追加 CreateUser 方法**

```go
// CreateUser 创建统一用户（管理员或普通用户）。
// 与 CreateEndUser 的区别：支持 IsAdmin，同时创建 user_orgs 记录。
func (s *EndUserManagementAppService) CreateUser(
    ctx context.Context,
    cmd CreateUserCommand,
) (*CreateUserResult, error) {
    if err := domainenduser.ValidatePasswordStrength(cmd.Password); err != nil {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, err.Error())
    }
    if err := domainenduser.ValidateUsername(cmd.Username); err != nil {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, err.Error())
    }

    hashedPwd, err := domainenduser.NewHashedPasswordFromPlain(cmd.Password)
    if err != nil {
        return nil, bizerrors.Wrapf(err, "failed to hash password")
    }

    userID, err := bizutils.GenerateUUIDV7()
    if err != nil {
        return nil, bizerrors.Wrapf(err, "failed to generate user id")
    }

    user, err := domainenduser.NewEndUser(userID, cmd.OrgName, cmd.Username, hashedPwd)
    if err != nil {
        return nil, bizerrors.Wrapf(err, "failed to create user entity")
    }

    repo := infrrepo.NewSqlEndUserRepository(s.db, cmd.OrgName, "")
    if err := repo.Save(ctx, user); err != nil {
        if shared.IsDuplicateKeyError(err) {
            return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserAlreadyExists, cmd.Username)
        }
        return nil, bizerrors.ConvertRepositoryError(ctx, err)
    }

    // 如果 IsAdmin，更新 user_orgs.is_admin = true
    if cmd.IsAdmin {
        // user_orgs 记录已由 repo.Save 创建（is_admin=false）
        // 需要追加更新 is_admin
        // TODO: UserOrgRepository 接口待 Plan 2c 后扩展；当前通过 dbgen 直接操作
        // 暂时通过 SQL 直接更新（后续重构为 repository 方法）
        const updateIsAdmin = `UPDATE user_orgs SET is_admin = 1, updated_at = NOW(3) WHERE user_id = ? AND org_name = ? AND deleted_at = 0`
        if _, execErr := s.db.ExecContext(ctx, updateIsAdmin, user.ID, cmd.OrgName); execErr != nil {
            return nil, bizerrors.Wrapf(execErr, "failed to set user as admin")
        }
    }

    return &CreateUserResult{
        ID:          user.ID,
        Username:    user.Username,
        IsAdmin:     cmd.IsAdmin,
        IsForbidden: user.IsForbidden,
        CreatedAt:   user.CreatedAt,
        UpdatedAt:   user.UpdatedAt,
    }, nil
}
```

- [ ] **Step 4: 编译检查**

```bash
cd modelcraft-backend && go build ./internal/app/enduser/... 2>&1
```

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/app/enduser/
git commit -m "app: add CreateUser service method with isAdmin support"
```

---

## Task 4: 实现 GraphQL Resolver — CreateUser 和 MyProjects

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/graphql/org/end_user.resolvers.go`

- [ ] **Step 1: 读取当前 resolver 文件末尾和 generated.go 中新接口的方法签名**

```bash
tail -50 modelcraft-backend/internal/interfaces/graphql/org/end_user.resolvers.go
grep -n "func.*CreateUser\|func.*MyProjects" \
  modelcraft-backend/internal/interfaces/graphql/org/generated/generated.go | head -10
```

- [ ] **Step 2: 在 end_user.resolvers.go 末尾追加 CreateUser resolver**

```go
// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input generated.CreateUserInput) (*generated.CreateUserPayload, error) {
    if strings.TrimSpace(input.Username) == "" {
        return &generated.CreateUserPayload{
            Error: &generated.InvalidInput{Message: "username is required"},
        }, nil
    }
    if strings.TrimSpace(input.Password) == "" {
        return &generated.CreateUserPayload{
            Error: &generated.InvalidInput{Message: "password is required"},
        }, nil
    }

    service := r.EndUserMgmtAppService
    if service == nil {
        return &generated.CreateUserPayload{
            Error: &generated.InvalidInput{Message: "user management service not initialized"},
        }, nil
    }

    orgName, err := ctxutils.GetOrgNameFromContext(ctx)
    if err != nil {
        return &generated.CreateUserPayload{
            Error: &generated.InvalidInput{Message: "orgName not found in context"},
        }, nil
    }

    result, err := service.CreateUser(ctx, appEnduser.CreateUserCommand{
        OrgName:  orgName,
        Username: input.Username,
        Password: input.Password,
        IsAdmin:  input.IsAdmin,
    })
    if err != nil {
        var bizErr *bizerrors.BusinessError
        if errors.As(err, &bizErr) {
            logfacade.GetLogger(ctx).Error(ctx, "failed to create user",
                logfacade.Err(bizErr), logfacade.Stack(bizErr))
            return &generated.CreateUserPayload{Error: convertOrgCreateEndUserError(bizErr)}, nil
        }
        return nil, err
    }

    return &generated.CreateUserPayload{
        User: &generated.EndUser{
            ID:          result.ID,
            Username:    result.Username,
            IsForbidden: result.IsForbidden,
            IsBuiltin:   false,
            CreatedBy:   nil,
            CreatedAt:   result.CreatedAt,
            UpdatedAt:   result.UpdatedAt,
        },
    }, nil
}
```

- [ ] **Step 3: 在 end_user.resolvers.go 末尾追加 MyProjects resolver**

`myProjects` 와 `endUserProjects` 는 동일한 로직 — 현재 end-user의 접근 가능한 프로젝트 목록을 반환한다. 기존 `EndUserProjects` resolver 로직을 복사해서 `MyProjects`로 만든다.

```bash
grep -n "func.*EndUserProjects\|endUserProjects" \
  modelcraft-backend/internal/interfaces/graphql/org/end_user.resolvers.go
```

기존 `EndUserProjects` resolver를 읽고 동일한 로직으로 `MyProjects` resolver를 추가한다.

- [ ] **Step 4: 편집 후 확인 — 생성된 인터페이스가 모두 구현됐는지 확인**

```bash
cd modelcraft-backend && go build ./internal/interfaces/graphql/org/... 2>&1
```

예상: 0 오류. 새 인터페이스가 모두 구현됨.

- [ ] **Step 5: 전체 컴파일**

```bash
cd modelcraft-backend && go build ./... 2>&1
```

- [ ] **Step 6: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/interfaces/graphql/org/end_user.resolvers.go
git commit -m "resolver: implement CreateUser and MyProjects GraphQL resolvers"
```

---

## Task 5: 全局编译 + 测试验证

- [ ] **Step 1: 全量编译**

```bash
cd modelcraft-backend && go build ./... 2>&1
```

预期：0 errors。

- [ ] **Step 2: 运行 graphql 层测试（如有）**

```bash
cd modelcraft-backend && go test ./internal/interfaces/graphql/... 2>&1 | tail -10
```

- [ ] **Step 3: 确认新 API 在 generated 中存在**

```bash
grep -n "CreateUser\|MyProjects" \
  modelcraft-backend/internal/interfaces/graphql/org/generated/generated.go | head -10
```

- [ ] **Step 4: 最终 commit**

```bash
cd /data/home/lukemxjia/modelcraft
git status
git add -A && git commit -m "chore: Plan 4 complete — createUser + myProjects GraphQL API added"
```
