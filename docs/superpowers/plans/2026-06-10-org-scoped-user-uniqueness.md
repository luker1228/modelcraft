# Org-Scoped User Uniqueness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `users.phone` 和 `users.name` 的唯一性从全局改为 org-scoped；`organizations` 表新增全局唯一 phone；Admin 登录改为仅手机号；前端首页删除 EndUser 登录入口。

**Architecture:** 数据库新增两列（`organizations.phone`、`users.org_name`），唯一索引替换。SQL 查询层（sqlc）、Repository 接口、Application Service、HTTP Handler、OpenAPI spec、前端登录/注册页面依次更新。Admin 登录变为两步：先用 `org.phone` 定位 org，再用 `users.org_name + phone` 取 user。

**Tech Stack:** Go 1.22, sqlc, Atlas (db migration), oapi-codegen (OpenAPI → Go types), Next.js 15, React Hook Form, Zod

---

## File Map

| 文件 | 变更类型 |
|------|---------|
| `db/schema/mysql/05_organizations.sql` | 新增 `phone` 列 + `uk_org_phone` 索引 |
| `db/schema/mysql/06_users.sql` | 新增 `org_name` 列，替换唯一索引，更新查询索引 |
| `db/queries/org.sql` | 新增/修改 sqlc 查询：`GetOrgByPhone`、`ExistsByOrgPhone`、`CreateOrganization`（含 phone）、`CreateUser`（含 org_name）、`ExistsByPhone`/`ExistsByUserName` 改为 org-scoped |
| `db/queries/user_auth.sql` | 修改 `GetUserByPhone`、`GetUserByName`、`ExistsByPhone`、`ExistsByUserName` 为 org-scoped |
| `internal/infrastructure/dbgen/` | sqlc 自动生成（`just generate-gql` 类似，实际用 `just generate-sqlc`） |
| `internal/domain/organization/organization.go` | `Organization` struct 新增 `Phone` 字段 |
| `internal/domain/organization/repository.go` | 新增 `GetByPhone`、`ExistsByPhone` 接口方法 |
| `internal/domain/user/user.go` | `User` struct 新增 `OrgName` 字段 |
| `internal/domain/user/repository.go` | `ExistsByPhone`/`ExistsByName` 改为 org-scoped 签名；新增 `ExistsByPhoneInOrg`/`ExistsByNameInOrg` |
| `internal/infrastructure/repository/sql_org_repository.go` | 实现 org `GetByPhone`/`ExistsByPhone`；修改 user `ExistsByPhone`/`ExistsByName` 为 org-scoped |
| `internal/app/auth/commands.go` | `RegisterCommand` 新增 `OrgDisplayName`；`LoginCommand` 移除 `IdentifierType`（只保留 phone） |
| `internal/app/auth/token_service.go` | `Register`：先检查 `org.phone` 全局唯一；`Login`：改为两步手机号查找 |
| `internal/app/organization/create_organization_service.go` | `CreateOrganizationInput` 新增 `Phone`；`Execute` 中传入 phone 创建 org |
| `api/openapi/auth.yaml` | `RegisterRequest` 新增 `orgDisplayName`；`LoginRequest` 移除 `identifierType`，只保留 `phone` |
| `internal/interfaces/http/handlers/auth/handler.go` | `HandleRegister` 传入 `OrgDisplayName`；`HandleLogin` 只走 phone 路径 |
| `modelcraft-front/src/app/page.tsx` | 删除「组织员工」section |
| `modelcraft-front/src/app/tenant/login/page.tsx` | 删除用户名 Tab，只保留手机号登录 |
| `modelcraft-front/src/app/tenant/register/page.tsx` | 新增 `orgDisplayName` 输入框 |
| `modelcraft-front/src/shared/validation/auth.ts` | `loginFormSchema` 移除 `identifierType`；`registerFormSchema` 新增 `orgDisplayName` |

---

## Task 1: DB Schema — organizations 新增 phone

**Files:**
- Modify: `db/schema/mysql/05_organizations.sql`

- [ ] **Step 1: 修改 organizations schema 文件**

在 `05_organizations.sql` 中，在 `status` 字段后、`created_at` 之前新增 `phone` 列，并在末尾 `INDEX` 区新增 `uk_org_phone`：

```sql
CREATE TABLE IF NOT EXISTS `organizations` (
  `name` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '唯一标识符，随机slug',
  `display_name` VARCHAR(255) COMMENT '用于 UI 显示的名称',
  `owner_id` VARCHAR(36) COMMENT '组织创建者/所有者（引用 users.id）',
  `phone` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'Org 注册手机号，全局唯一',
  `status` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态：active、suspended、deleted',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',

  UNIQUE INDEX `uk_org_phone` (`phone`, `delete_token`) COMMENT 'Org 手机号全局唯一',
  INDEX `idx_org_owner` (`owner_id`) COMMENT '按所有者查找组织',
  INDEX `idx_org_status` (`status`) COMMENT '按状态筛选',
  INDEX `idx_org_live_status` (`deleted_at`, `status`) COMMENT '按活跃状态筛选组织'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织表（多租户容器）';
```

- [ ] **Step 2: 运行 db migration**

```bash
cd modelcraft-backend && just db up
```

Expected: 无报错，`organizations` 表新增 `phone` 列和 `uk_org_phone` 唯一索引

- [ ] **Step 3: 验证**

```bash
cd modelcraft-backend && just db login
# 在 MySQL 中执行：
SHOW CREATE TABLE organizations\G
```

Expected: 输出包含 `phone` 列和 `UNIQUE KEY uk_org_phone`

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/db/schema/mysql/05_organizations.sql
git commit -m "feat(schema): add phone to organizations for global uniqueness"
```

---

## Task 2: DB Schema — users 新增 org_name，替换唯一索引

**Files:**
- Modify: `db/schema/mysql/06_users.sql`

- [ ] **Step 1: 修改 users schema 文件**

将 `06_users.sql` 中的 `users` 表定义替换为以下内容（新增 `org_name`，删除旧唯一索引，新增 org-scoped 唯一索引）：

```sql
CREATE TABLE IF NOT EXISTS `users` (
  `id` VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '内部 UUID',
  `external_id` VARCHAR(255) NULL COMMENT '外部认证提供者用户 ID（来自 JWT.sub，AuthProvider 用户有值，本地注册用户为 NULL）',
  `name` VARCHAR(255) NOT NULL COMMENT '用户名（userName）',
  `phone` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '用户手机号',
  `password_hash` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'bcrypt 密码哈希（本地注册用户有值，AuthProvider 用户为空）',
  `display_name` VARCHAR(255) COMMENT '用于 UI 显示的名称',
  `org_name` VARCHAR(36) NOT NULL DEFAULT '' COMMENT '所属 Org，创建时绑定（引用 organizations.name）',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
  `delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',

  UNIQUE INDEX `uk_org_user_phone` (`org_name`, `phone`, `delete_token`) COMMENT 'Org 内手机号唯一',
  UNIQUE INDEX `uk_org_user_name` (`org_name`, `name`, `delete_token`) COMMENT 'Org 内用户名唯一',
  INDEX `idx_external_id` (`external_id`) COMMENT '按外部 ID 快速查找',
  INDEX `idx_users_live_name` (`deleted_at`, `org_name`, `name`) COMMENT '活跃用户查询索引',
  CONSTRAINT `fk_users_org` FOREIGN KEY (`org_name`) REFERENCES `organizations`(`name`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
```

注意：`users` schema 文件末尾有 `ALTER TABLE organizations ADD CONSTRAINT fk_org_owner...`，保持不变。

- [ ] **Step 2: 运行 db migration**

```bash
cd modelcraft-backend && just db up
```

Expected: 无报错，`users` 表新增 `org_name` 列，旧的 `uk_phone` / `uk_user_name` 索引被新的 org-scoped 索引替换

- [ ] **Step 3: 验证**

```bash
cd modelcraft-backend && just db login
# 在 MySQL 中执行：
SHOW CREATE TABLE users\G
```

Expected: 包含 `org_name` 列、`uk_org_user_phone`、`uk_org_user_name` 索引

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/db/schema/mysql/06_users.sql
git commit -m "feat(schema): org-scope users phone/name uniqueness, add org_name fk"
```

---

## Task 3: sqlc 查询层更新

**Files:**
- Modify: `db/queries/org.sql`
- Modify: `db/queries/user_auth.sql`
- Modify: `internal/infrastructure/dbgen/` (自动生成，运行 `just generate-sqlc`)

- [ ] **Step 1: 更新 org.sql — CreateOrganization 含 phone，新增 org phone 查询**

在 `db/queries/org.sql` 中：

（1）修改 `CreateOrganization` 含 phone：
```sql
-- name: CreateOrganization :exec
INSERT INTO organizations (name, display_name, owner_id, phone, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, NOW(3), NOW(3));
```

（2）新增两个查询（追加到文件末尾）：
```sql
-- name: GetOrganizationByPhone :one
SELECT * FROM organizations WHERE phone = ? AND `organizations`.`deleted_at` = 0 LIMIT 1;

-- name: ExistsOrganizationByPhone :one
SELECT COUNT(*) FROM organizations WHERE phone = ? AND `organizations`.`deleted_at` = 0;
```

（3）修改 `CreateUser` 含 org_name：
```sql
-- name: CreateUser :exec
INSERT INTO users (id, external_id, name, phone, password_hash, display_name, org_name, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));
```

- [ ] **Step 2: 更新 user_auth.sql — 查询改为 org-scoped**

将 `db/queries/user_auth.sql` 替换为：

```sql
-- name: GetUserByPhoneInOrg :one
SELECT id, phone, password_hash, name, external_id, org_name, created_at, updated_at
FROM users
WHERE org_name = ? AND phone = ? AND `users`.`deleted_at` = 0 LIMIT 1;

-- name: GetUserByNameInOrg :one
SELECT id, phone, password_hash, name, external_id, org_name, created_at, updated_at
FROM users
WHERE org_name = ? AND name = ? AND `users`.`deleted_at` = 0 LIMIT 1;

-- name: ExistsByPhoneInOrg :one
SELECT EXISTS(SELECT 1 FROM users WHERE org_name = ? AND phone = ?) AS phone_exists;

-- name: ExistsByUserNameInOrg :one
SELECT EXISTS(SELECT 1 FROM users WHERE org_name = ? AND name = ?) AS name_exists;
```

- [ ] **Step 3: 重新生成 sqlc 代码**

```bash
cd modelcraft-backend && just generate-sqlc
```

Expected: `internal/infrastructure/dbgen/` 下的文件更新，包含新的方法签名（`GetOrganizationByPhone`、`ExistsOrganizationByPhone`、`GetUserByPhoneInOrg`、`ExistsByPhoneInOrg` 等）

- [ ] **Step 4: 编译验证**

```bash
cd modelcraft-backend && go build ./...
```

Expected: 编译报错，因为 Repository 层还在调用旧方法名（`GetUserByPhone`、`ExistsByPhone` 等）。**这是预期的**，说明我们需要继续更新 Repository。

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/db/queries/org.sql modelcraft-backend/db/queries/user_auth.sql modelcraft-backend/internal/infrastructure/dbgen/
git commit -m "feat(sqlc): org-scoped phone/name queries, org phone lookup"
```

---

## Task 4: Domain 层更新

**Files:**
- Modify: `modelcraft-backend/internal/domain/organization/organization.go`
- Modify: `modelcraft-backend/internal/domain/organization/organization_test.go`
- Modify: `modelcraft-backend/internal/domain/organization/repository.go`
- Modify: `modelcraft-backend/internal/domain/user/user.go`
- Modify: `modelcraft-backend/internal/domain/user/repository.go`

- [ ] **Step 1: 写失败测试 — Organization.Phone 字段**

在 `internal/domain/organization/organization_test.go` 中，在 `TestNewOrganization` 已有的 test 后追加：

```go
func TestNewOrganization_WithPhone(t *testing.T) {
    t.Run("should store phone in organization", func(t *testing.T) {
        org, err := NewOrganization("my-company", "My Company", "user-uuid-001", "13800138000")
        assert.NoError(t, err)
        assert.Equal(t, "13800138000", org.Phone)
    })

    t.Run("should create organization with empty phone", func(t *testing.T) {
        org, err := NewOrganization("my-company", "My Company", "user-uuid-001", "")
        assert.NoError(t, err)
        assert.Equal(t, "", org.Phone)
    })
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/domain/organization/... -run TestNewOrganization_WithPhone -v
```

Expected: 编译失败，`NewOrganization` 不接受第 4 个参数

- [ ] **Step 3: 更新 Organization 实体和 NewOrganization**

在 `internal/domain/organization/organization.go` 中：

（1）`Organization` struct 新增 `Phone` 字段：
```go
type Organization struct {
    Name        string
    DisplayName string
    OwnerID     string
    Phone       string    // Org 注册手机号，全局唯一
    Status      OrgStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

（2）`NewOrganization` 函数签名新增 `phone string` 参数：
```go
func NewOrganization(name, displayName, ownerID, phone string) (*Organization, error) {
    now := time.Now()
    org := &Organization{
        Name:        name,
        DisplayName: displayName,
        OwnerID:     ownerID,
        Phone:       phone,
        Status:      OrgStatusActive,
        CreatedAt:   now,
        UpdatedAt:   now,
    }
    if err := org.Validate(); err != nil {
        return nil, err
    }
    return org, nil
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd modelcraft-backend && go test ./internal/domain/organization/... -v
```

Expected: 所有 organization domain tests PASS（包括旧的测试，因为它们会报编译错误直到调用处被修复）

注意：此时 `NewOrganization` 的调用处（`create_organization_service.go`、`organization_test.go` 内的旧调用）会编译失败，后续 Task 修复。

- [ ] **Step 5: 更新 Organization Repository 接口**

在 `internal/domain/organization/repository.go` 中新增两个方法：

```go
type OrganizationRepository interface {
    // Create 创建组织
    Create(ctx context.Context, org *Organization) error

    // GetByName 根据名称获取组织
    GetByName(ctx context.Context, name string) (*Organization, error)

    // GetByPhone 根据手机号获取组织（全局唯一）
    // 返回 nil, shared.NewNotFoundError 当不存在时
    GetByPhone(ctx context.Context, phone string) (*Organization, error)

    // ListByUser 获取用户所属的所有组织
    ListByUser(ctx context.Context, userID string) ([]*Organization, error)

    // Update 更新组织
    Update(ctx context.Context, org *Organization) error

    // ExistsByName 检查组织名称是否已存在
    ExistsByName(ctx context.Context, name string) (bool, error)

    // ExistsByPhone 检查 org phone 是否已被注册
    ExistsByPhone(ctx context.Context, phone string) (bool, error)
}
```

- [ ] **Step 6: 更新 User struct — 新增 OrgName 字段**

在 `internal/domain/user/user.go` 中，`User` struct 新增 `OrgName`：

```go
type User struct {
    ID           string
    ExternalID   string
    Name         string
    Phone        PhoneNumber
    PasswordHash string
    OrgName      string    // 所属 Org，创建时绑定
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

`NewUser` 函数签名新增 `orgName string`，并在 user 赋值时加入：

```go
func NewUser(id, userName string, phone PhoneNumber, passwordHash, orgName string) (*User, error) {
    if phone.IsZero() {
        return nil, fmt.Errorf("phone number is required")
    }
    if passwordHash == "" {
        return nil, fmt.Errorf("password hash is required")
    }
    if orgName == "" {
        return nil, fmt.Errorf("org name is required")
    }
    if err := ValidateUserName(userName); err != nil {
        return nil, err
    }
    now := time.Now()
    u := &User{
        ID:           id,
        Name:         userName,
        Phone:        phone,
        PasswordHash: passwordHash,
        OrgName:      orgName,
        CreatedAt:    now,
        UpdatedAt:    now,
    }
    if err := u.Validate(); err != nil {
        return nil, err
    }
    return u, nil
}
```

- [ ] **Step 7: 更新 User Repository 接口 — org-scoped 签名**

在 `internal/domain/user/repository.go` 中，将 `ExistsByPhone` 和 `ExistsByName` 改为 org-scoped：

```go
type UserRepository interface {
    // Create 创建用户
    Create(ctx context.Context, user *User) error

    // GetByID 根据内部 UUID 获取用户
    GetByID(ctx context.Context, id string) (*User, error)

    // GetByExternalID 根据外部认证提供者 ID 获取用户
    GetByExternalID(ctx context.Context, externalID string) (*User, error)

    // ExistsByExternalID 检查外部 ID 是否已存在
    ExistsByExternalID(ctx context.Context, externalID string) (bool, error)

    // FindIDByExternalID retrieves the internal user ID by external authentication provider ID.
    FindIDByExternalID(ctx context.Context, externalID string) (string, bool, error)

    // GetByPhone 根据 org 和手机号获取用户
    GetByPhone(ctx context.Context, orgName, phone string) (*User, error)

    // GetByName 根据 org 和用户名获取用户
    GetByName(ctx context.Context, orgName, name string) (*User, error)

    // ExistsByPhone 检查 org 内手机号是否已被注册
    ExistsByPhone(ctx context.Context, orgName, phone string) (bool, error)

    // ExistsByName 检查 org 内用户名是否已被占用
    ExistsByName(ctx context.Context, orgName, name string) (bool, error)
}
```

- [ ] **Step 8: Commit**

```bash
git add modelcraft-backend/internal/domain/
git commit -m "feat(domain): add Phone to Organization, OrgName to User, org-scoped repo interfaces"
```

---

## Task 5: Repository 实现层更新

**Files:**
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_org_repository.go`

- [ ] **Step 1: 更新 `CreateOrganization` — 传 phone**

在 `sql_org_repository.go` 的 `SqlOrganizationRepository.Create` 方法中，找到调用 `q.CreateOrganization` 的地方，更新参数：

```go
func (r *SqlOrganizationRepository) Create(ctx context.Context, org *organization.Organization) error {
    return r.q.CreateOrganization(ctx, dbgen.CreateOrganizationParams{
        Name:        org.Name,
        DisplayName: sql.NullString{String: org.DisplayName, Valid: org.DisplayName != ""},
        OwnerID:     sql.NullString{String: org.OwnerID, Valid: org.OwnerID != ""},
        Phone:       org.Phone,
        Status:      string(org.Status),
    })
}
```

- [ ] **Step 2: 新增 `GetByPhone` 和 `ExistsByPhone` — Organization**

在 `sql_org_repository.go` 中，在 `ExistsByName` 方法之后新增：

```go
// GetByPhone retrieves an organization by phone number.
func (r *SqlOrganizationRepository) GetByPhone(ctx context.Context, phone string) (*organization.Organization, error) {
    row, err := r.q.GetOrganizationByPhone(ctx, phone)
    if err != nil {
        if sqlerr.IsNotFoundError(err) {
            return nil, shared.NewNotFoundError("organization not found by phone: " + phone)
        }
        return nil, bizerrors.Wrapf(err, "failed to get organization by phone: %s", phone)
    }
    return OrgToDomain(row), nil
}

// ExistsByPhone checks whether an organization with the given phone already exists.
func (r *SqlOrganizationRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
    count, err := r.q.ExistsOrganizationByPhone(ctx, phone)
    if err != nil {
        return false, bizerrors.Wrapf(err, "failed to check org phone existence: %s", phone)
    }
    return count > 0, nil
}
```

注：`OrgToDomain` 函数需要更新以映射 `Phone` 字段（见 Step 3）。

- [ ] **Step 3: 更新 `OrgToDomain` helper — 映射 Phone 字段**

找到 `sql_org_repository.go` 中的 `OrgToDomain` 函数（或类似名称），添加 `Phone` 映射：

```go
func OrgToDomain(row dbgen.Organization) *organization.Organization {
    var displayName string
    if row.DisplayName.Valid {
        displayName = row.DisplayName.String
    }
    var ownerID string
    if row.OwnerID.Valid {
        ownerID = row.OwnerID.String
    }
    return &organization.Organization{
        Name:        row.Name,
        DisplayName: displayName,
        OwnerID:     ownerID,
        Phone:       row.Phone,
        Status:      organization.OrgStatus(row.Status),
        CreatedAt:   row.CreatedAt,
        UpdatedAt:   row.UpdatedAt,
    }
}
```

- [ ] **Step 4: 更新 `SqlUserRepository` — org-scoped 查询**

在 `sql_org_repository.go` 中，将 `GetByPhone`、`GetByName`、`ExistsByPhone`、`ExistsByName` 改为 org-scoped 版本：

```go
func (r *SqlUserRepository) GetByPhone(ctx context.Context, orgName, phone string) (*user.User, error) {
    row, err := r.q.GetUserByPhoneInOrg(ctx, dbgen.GetUserByPhoneInOrgParams{OrgName: orgName, Phone: phone})
    if err != nil {
        if sqlerr.IsNotFoundError(err) {
            return nil, shared.NewNotFoundError("user not found by phone in org: " + orgName)
        }
        return nil, bizerrors.Wrapf(err, "failed to get user by phone in org: %s", orgName)
    }
    return userRowToDomain(row), nil
}

func (r *SqlUserRepository) GetByName(ctx context.Context, orgName, name string) (*user.User, error) {
    row, err := r.q.GetUserByNameInOrg(ctx, dbgen.GetUserByNameInOrgParams{OrgName: orgName, Name: name})
    if err != nil {
        if sqlerr.IsNotFoundError(err) {
            return nil, shared.NewNotFoundError("user not found by name in org: " + orgName)
        }
        return nil, bizerrors.Wrapf(err, "failed to get user by name in org: %s", orgName)
    }
    return userRowToDomain(row), nil
}

func (r *SqlUserRepository) ExistsByPhone(ctx context.Context, orgName, phone string) (bool, error) {
    exists, err := r.q.ExistsByPhoneInOrg(ctx, dbgen.ExistsByPhoneInOrgParams{OrgName: orgName, Phone: phone})
    if err != nil {
        return false, bizerrors.Wrapf(err, "failed to check phone existence in org: %s", orgName)
    }
    return exists, nil
}

func (r *SqlUserRepository) ExistsByName(ctx context.Context, orgName, name string) (bool, error) {
    exists, err := r.q.ExistsByUserNameInOrg(ctx, dbgen.ExistsByUserNameInOrgParams{OrgName: orgName, Name: name})
    if err != nil {
        return false, bizerrors.Wrapf(err, "failed to check user name existence in org: %s", orgName)
    }
    return exists, nil
}
```

同时更新 `userRowToDomain` helper 以包含 `OrgName`：

```go
func userRowToDomain(row interface{ /* GetUserByPhoneInOrgRow or GetUserByNameInOrgRow */ }) *user.User {
    // 根据 sqlc 生成的行类型适配，关键是映射 OrgName
    // 示例（以 GetUserByPhoneInOrgRow 为例）：
    var externalID string
    if row.ExternalID.Valid {
        externalID = row.ExternalID.String
    }
    var phoneVO user.PhoneNumber
    if row.Phone != "" {
        p, err := user.NewPhoneNumber(row.Phone)
        if err == nil {
            phoneVO = p
        }
    }
    return &user.User{
        ID:           row.ID,
        ExternalID:   externalID,
        Name:         row.Name,
        Phone:        phoneVO,
        PasswordHash: row.PasswordHash,
        OrgName:      row.OrgName,
        CreatedAt:    row.CreatedAt,
        UpdatedAt:    row.UpdatedAt,
    }
}
```

注意：sqlc 为不同查询生成不同的 row struct，如果类型不同，需要为每种类型分别写转换代码，或使用 interface 统一字段访问。以实际 sqlc 生成的类型为准。

- [ ] **Step 5: 更新 `Create` user — 传 org_name**

找到 `sql_org_repository.go` 中 `SqlUserRepository.Create` 方法，更新 `CreateUserParams`：

```go
func (r *SqlUserRepository) Create(ctx context.Context, u *user.User) error {
    return r.q.CreateUser(ctx, dbgen.CreateUserParams{
        ID:           u.ID,
        ExternalID:   sql.NullString{String: u.ExternalID, Valid: u.ExternalID != ""},
        Name:         u.Name,
        Phone:        u.Phone.String(),
        PasswordHash: u.PasswordHash,
        DisplayName:  sql.NullString{},
        OrgName:      u.OrgName,
    })
}
```

- [ ] **Step 6: 编译验证**

```bash
cd modelcraft-backend && go build ./...
```

Expected: 编译报错集中在 Application Service 层（`token_service.go`、`create_organization_service.go` 调用 `NewOrganization`、`NewUser`、`ExistsByPhone` 等时参数数量不对）。后续 Task 修复。

- [ ] **Step 7: Commit**

```bash
git add modelcraft-backend/internal/infrastructure/repository/sql_org_repository.go
git commit -m "feat(repo): org-scoped user queries, org phone lookup, CreateUser with org_name"
```

---

## Task 6: Application Service 层更新

**Files:**
- Modify: `modelcraft-backend/internal/app/auth/commands.go`
- Modify: `modelcraft-backend/internal/app/auth/token_service.go`
- Modify: `modelcraft-backend/internal/app/organization/create_organization_service.go`

- [ ] **Step 1: 更新 `commands.go`**

（1）`RegisterCommand` 新增 `OrgDisplayName`（手机号已有，不变）：

```go
type RegisterCommand struct {
    Phone            string
    Password         string
    UserName         string
    OrgDisplayName   string // Org 展示名称（新增）
    OrganizationName string // 可选 org slug，为空时自动生成
}
```

（2）`LoginCommand` 移除 `IdentifierType`，只保留 `Phone` + `Password`：

```go
type LoginCommand struct {
    Phone    string
    Password string
}
```

同时删除 `IdentifierType` 常量（`IdentifierTypePhone`、`IdentifierTypeUsername`）和 `IdentifierType` 类型，**除非** EndUser login 还在使用它们。

检查 EndUser login 是否使用 `IdentifierType`：

```bash
grep -rn "IdentifierType\|IdentifierTypePhone\|IdentifierTypeUsername" \
  modelcraft-backend/internal/app/auth/token_service_enduser.go
```

若 EndUser login 仍使用这些类型，则**只从 `LoginCommand` 中移除 `IdentifierType` 字段**，不删除类型定义本身（将其移至 enduser 相关 command 内）。

- [ ] **Step 2: 更新 `Register` 方法**

在 `token_service.go` 的 `Register` 方法中做以下修改：

（1）将步骤 4（检查 userName 唯一）改为 org-scoped：
```go
// 4. 检查 org 内 userName 是否已被占用（org 在后面创建，此时先跳过；userName 唯一性靠 DB 约束保证）
// 注意：org 在用户创建之前创建，因此这里先不校验 userName，DB 唯一约束会兜底。
```

（2）将步骤 5（检查手机号）改为检查 `organizations.phone` 全局唯一：
```go
// 5. 检查手机号是否已注册 Org
phoneExists, err := s.orgRepo.ExistsByPhone(ctx, cmd.Phone)
if err != nil {
    return nil, bizerrors.ConvertRepositoryError(ctx, err)
}
if phoneExists {
    return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.PhoneAlreadyExists, phone.Masked())
}
```

（3）在 `orgInput` 中传入 phone 和 display name：
```go
orgInput := &organization.CreateOrganizationInput{
    DisplayName:      cmd.OrgDisplayName,
    OrganizationName: cmd.OrganizationName,
    OwnerUserID:      u.ID,
    Phone:            cmd.Phone,
}
```

（4）`NewUser` 调用新增 `orgName` 参数（org 在 user 之前创建）：

这里有顺序问题：org 需要先创建才能知道 `orgName`，而 user 创建需要 `orgName`。当前的 `registerWithTxOrgService` 在事务内先创建 user 再创建 org。需要调整为先创建 org，再创建 user。

更新后的事务顺序（在 `registerWithTxOrgService` 中）：
```
1. 创建 Org（含 phone）→ 取得 orgName
2. 创建 User（含 orgName）
3. 创建 Profile
4. 创建 user_orgs
```

具体：找到 `registerWithTxOrgService` 中 `txOrgService.ExecuteWithQuerier` 的调用，将 user + profile 的持久化移到 org 创建之后，并在 `NewUser` 调用时传入 org 返回的 `orgName`。

- [ ] **Step 3: 更新 `Login` 方法 — 两步手机号查找**

将 `token_service.go` 的 `Login` 方法替换为：

```go
// Login 管理员手机号登录：先用 org.phone 定位 org，再查 user。
func (s *TokenService) Login(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
    logger := logfacade.GetLogger(ctx)

    if cmd.Phone == "" {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthParamInvalid, "phone is required")
    }

    // Step 1: 用手机号定位 org
    org, err := s.orgRepo.GetByPhone(ctx, cmd.Phone)
    if err != nil {
        if shared.IsNotFoundError(err) {
            return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthenticationFailed, "phone not registered")
        }
        return nil, bizerrors.ConvertRepositoryError(ctx, err)
    }

    // Step 2: 在该 org 内查找 user
    u, err := s.userRepo.GetByPhone(ctx, org.Name, cmd.Phone)
    if err != nil {
        if shared.IsNotFoundError(err) {
            return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthenticationFailed, "user not found")
        }
        return nil, bizerrors.ConvertRepositoryError(ctx, err)
    }

    // Step 3: 验证密码（直接用 bcrypt）
    if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(cmd.Password)); err != nil {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthenticationFailed, "incorrect password")
    }

    // Step 4: 签发 token（与原逻辑相同）
    plaintext, hash, err := GenerateRefreshToken()
    if err != nil {
        return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate refresh token")
    }
    tokenID, err := bizutils.GenerateUUIDV7()
    if err != nil {
        return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate token id")
    }
    expiresAt := time.Now().Add(s.refreshTTL)
    token := &domainauth.RefreshToken{
        ID:        tokenID,
        UserID:    u.ID,
        TokenHash: hash,
        ExpiresAt: expiresAt,
        CreatedAt: time.Now(),
    }
    if err := s.refreshTokenRepo.Save(ctx, token); err != nil {
        return nil, bizerrors.ConvertRepositoryError(ctx, err)
    }

    accessToken, expiresIn, err := s.issueAccessToken(u.ID, u.OrgName, true)
    if err != nil {
        return nil, err
    }

    logger.Infof(ctx, "Admin login success: user_id=%s, org_name=%s", u.ID, u.OrgName)

    return &LoginResult{
        UserID:       u.ID,
        UserName:     u.Name,
        OrgName:      u.OrgName,
        AccessToken:  accessToken,
        RefreshToken: plaintext,
        ExpiresIn:    expiresIn,
    }, nil
}
```

注：`s.orgRepo` 需要作为 `TokenService` 的依赖注入（若当前不存在，需新增 `orgRepo organization.OrganizationRepository` 字段并在构造函数中注入）。

- [ ] **Step 4: 更新 `CreateOrganizationInput` — 新增 Phone**

在 `create_organization_service.go` 中：

```go
type CreateOrganizationInput struct {
    DisplayName      string
    OrganizationName string
    OwnerUserID      string
    Phone            string // Org 注册手机号
}
```

在 `Execute`/`ExecuteWithQuerier` 中，创建 `organization.Organization` 时传入 phone：

```go
org, err := domainOrg.NewOrganization(resolvedName, input.DisplayName, input.OwnerUserID, input.Phone)
```

- [ ] **Step 5: 修复所有 `NewOrganization` 调用处（4 参数）**

```bash
grep -rn "NewOrganization(" modelcraft-backend/internal/ --include="*.go"
```

对每个调用处补齐第 4 个参数 `phone`，测试文件中使用空字符串 `""` 即可。

- [ ] **Step 6: 修复所有 `NewUser` 调用处（新增 orgName 参数）**

```bash
grep -rn "NewUser(" modelcraft-backend/internal/ --include="*.go"
```

对每个调用处补齐 `orgName` 参数，测试文件中使用 `"test-org"` 即可。

- [ ] **Step 7: 编译验证**

```bash
cd modelcraft-backend && go build ./...
```

Expected: 无编译错误，或只剩 handler 层的类型不匹配（`IdentifierType` 相关）

- [ ] **Step 8: Commit**

```bash
git add modelcraft-backend/internal/app/
git commit -m "feat(app): register org-phone uniqueness check, login two-step phone lookup"
```

---

## Task 7: OpenAPI spec + Handler 更新

**Files:**
- Modify: `modelcraft-backend/api/openapi/auth.yaml`
- Modify: `modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go`
- Run: `just generate-oapi` (重新生成 Go types)

- [ ] **Step 1: 更新 auth.yaml — RegisterRequest 新增 orgDisplayName**

在 `api/openapi/auth.yaml` 的 `RegisterRequest` schema 中新增 `orgDisplayName` 字段：

```yaml
RegisterRequest:
  type: object
  required:
    - phone
    - password
    - userName
    - orgDisplayName
  properties:
    phone:
      type: string
      description: "11-digit mainland China phone number"
      pattern: "^1[3-9]\\d{9}$"
      example: "13800138000"
    password:
      type: string
      description: "Password (minimum 8 characters)"
      minLength: 8
      example: "mypassword123"
    userName:
      type: string
      description: "User name (3-32 chars)"
      pattern: "^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$"
      minLength: 3
      maxLength: 32
      example: "john_doe"
    orgDisplayName:
      type: string
      description: "Organization display name (human-readable, e.g. '我的公司')"
      minLength: 1
      maxLength: 64
      example: "我的公司"
    organizationName:
      type: string
      description: "Optional org name slug. Auto-generated if omitted."
      pattern: "^[a-z][a-z0-9_]{5,23}$"
      minLength: 6
      maxLength: 24
      example: "my_company"
```

- [ ] **Step 2: 更新 auth.yaml — LoginRequest 只保留 phone**

将 `LoginRequest` 替换为：

```yaml
LoginRequest:
  type: object
  required:
    - phone
    - password
  properties:
    phone:
      type: string
      description: "11-digit mainland China phone number"
      pattern: "^1[3-9]\\d{9}$"
      example: "13800138000"
    password:
      type: string
      description: "User password"
      example: "mypassword123"
```

- [ ] **Step 3: 重新生成 OpenAPI Go types**

```bash
cd modelcraft-backend && just generate-oapi
```

Expected: `internal/interfaces/http/generated/` 下 `RegisterRequest` 新增 `OrgDisplayName` 字段，`LoginRequest` 只有 `Phone` + `Password`

- [ ] **Step 4: 更新 `HandleRegister` — 传入 OrgDisplayName**

在 `internal/interfaces/http/handlers/auth/handler.go` 的 `HandleRegister` 中：

```go
result, err := h.tokenService.Register(r.Context(), appAuth.RegisterCommand{
    Phone:            req.Phone,
    Password:         req.Password,
    UserName:         req.UserName,
    OrgDisplayName:   req.OrgDisplayName,
    OrganizationName: derefString(req.OrganizationName),
})
```

- [ ] **Step 5: 更新 `HandleLogin` — 只走 phone 路径**

将 `HandleLogin` 中构建 `LoginCommand` 的部分替换为：

```go
cmd := appAuth.LoginCommand{
    Phone:    req.Phone,
    Password: req.Password,
}
```

删除原来的 `identifierType` switch/case 逻辑。

- [ ] **Step 6: 编译验证**

```bash
cd modelcraft-backend && go build ./...
```

Expected: 无编译错误

- [ ] **Step 7: 运行后端测试**

```bash
cd modelcraft-backend && go test ./internal/app/auth/... -v
```

Expected: 相关测试通过（若有测试 mock `GetByPhoneGlobal`，需更新 mock 为新的两步查找逻辑）

- [ ] **Step 8: Commit**

```bash
git add modelcraft-backend/api/openapi/auth.yaml \
        modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go \
        modelcraft-backend/internal/interfaces/http/generated/
git commit -m "feat(api): register requires orgDisplayName, login phone-only"
```

---

## Task 8: 前端 — 首页删除 EndUser 登录入口

**Files:**
- Modify: `modelcraft-front/src/app/page.tsx`

- [ ] **Step 1: 写测试（可选但推荐）**

由于这是纯 UI 删除，快速验证方式是目视检查。如果项目有 snapshot 测试，先运行：

```bash
cd modelcraft-front && npx jest --testPathPattern="app/page" -u
```

- [ ] **Step 2: 删除「组织员工」section**

将 `modelcraft-front/src/app/page.tsx` 替换为：

```tsx
'use client'

import NextLink from 'next/link'
import { AuthLayout } from '@/web/components/features/auth/auth-layout'
import { Button } from '@web/components/ui/button'
import {
  TENANT_LOGIN_PATH,
  TENANT_REGISTER_PATH,
} from '@shared/constants/routes'

export default function Home() {
  return (
    <AuthLayout
      title="欢迎使用 ModelCraft"
      subtitle="让 AI 安全、可控地使用数据库"
      showCliPromo
    >
      <div className="flex flex-col gap-4">
        <section className="rounded-xl border border-border bg-muted/20 p-4">
          <div className="mb-3">
            <h2 className="text-sm font-semibold text-foreground">组织管理员</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              登录管理后台，管理组织、权限和数据访问。
            </p>
          </div>
          <div className="flex flex-col gap-2">
            <Button asChild className="w-full">
              <NextLink href={TENANT_LOGIN_PATH}>管理员登录</NextLink>
            </Button>
            <Button asChild variant="outline" className="w-full">
              <NextLink href={TENANT_REGISTER_PATH}>注册组织</NextLink>
            </Button>
          </div>
        </section>
      </div>
    </AuthLayout>
  )
}
```

（移除了 `END_USER_LOGIN_PATH` import 和整个「组织员工」section）

- [ ] **Step 3: 检查 lint**

```bash
cd modelcraft-front && npm run lint
```

Expected: 无 lint 错误（`END_USER_LOGIN_PATH` 的 import 已移除，不会产生 unused import 错误）

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/app/page.tsx
git commit -m "feat(front): remove end-user login entry from homepage"
```

---

## Task 9: 前端 — 管理员登录页改为仅手机号

**Files:**
- Modify: `modelcraft-front/src/shared/validation/auth.ts`
- Modify: `modelcraft-front/src/app/tenant/login/page.tsx`
- Modify: `modelcraft-front/src/web/hooks/auth/use-auth-form.ts` (if identifierType logic is there)

- [ ] **Step 1: 更新 validation — loginFormSchema 移除 identifierType**

在 `modelcraft-front/src/shared/validation/auth.ts` 中，将 `loginFormSchema` 替换为：

```ts
/** 登录表单 — 仅手机号 */
export const loginFormSchema = z.object({
  phone: phoneNumberSchema,
  password: z.string().min(1, '请输入密码'),
})

export type LoginFormValues = z.infer<typeof loginFormSchema>
```

同时删除 `identifierSchema`（若无其他地方使用）。

- [ ] **Step 2: 更新 tenant login 页面 — 移除 Tab，只保留手机号输入**

将 `modelcraft-front/src/app/tenant/login/page.tsx` 替换为：

```tsx
'use client'

import { useSearchParams } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2 } from 'lucide-react'
import { loginFormSchema, type LoginFormValues } from '@/shared/validation/auth'
import { useLogin } from '@/web/hooks/auth/use-auth-form'
import { AuthLayout } from '@/web/components/features/auth/auth-layout'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { PasswordInput } from '@/web/components/common/password-input'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from '@web/components/ui/form'

export default function TenantLoginPage() {
  const searchParams = useSearchParams()
  const redirect = searchParams.get('redirect')
  const backHref = redirect ? `/?redirect=${encodeURIComponent(redirect)}` : '/'

  const { login, isLoading, error } = useLogin()

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginFormSchema),
    defaultValues: { phone: '', password: '' },
  })

  const handleSubmit = form.handleSubmit(async (values) => {
    await login(values)
  })

  return (
    <AuthLayout
      title="欢迎回来，管理员"
      subtitle="登录管理控制台"
      showCliPromo
      backLink={{ href: backHref, label: '返回登录选择' }}
    >
      <Form {...form}>
        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          {error && (
            <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          )}

          <FormField
            control={form.control}
            name="phone"
            render={({ field }) => (
              <FormItem>
                <FormLabel>手机号</FormLabel>
                <FormControl>
                  <Input
                    placeholder="请输入手机号"
                    autoComplete="tel"
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="password"
            render={({ field }) => (
              <FormItem>
                <FormLabel>密码</FormLabel>
                <FormControl>
                  <PasswordInput placeholder="请输入密码" autoComplete="current-password" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <Button
            type="submit"
            className="mt-1 h-10 w-full"
            disabled={isLoading}
          >
            {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
            登录
          </Button>
        </form>
      </Form>
    </AuthLayout>
  )
}
```

- [ ] **Step 3: 更新 `useLogin` hook — 传 phone 而非 identifier + identifierType**

找到 `modelcraft-front/src/web/hooks/auth/use-auth-form.ts`（或类似文件），更新 `login` 函数的请求 body 构建：

```bash
grep -n "identifierType\|identifier\|useLogin" modelcraft-front/src/web/hooks/auth/use-auth-form.ts | head -20
```

将 `useLogin` 中发送到 BFF 的 body 从 `{ identifier, identifierType, password }` 改为 `{ phone, password }`。

- [ ] **Step 4: 检查 lint**

```bash
cd modelcraft-front && npm run lint
```

Expected: 无错误

- [ ] **Step 5: Commit**

```bash
git add modelcraft-front/src/shared/validation/auth.ts \
        modelcraft-front/src/app/tenant/login/page.tsx \
        modelcraft-front/src/web/hooks/auth/
git commit -m "feat(front): tenant login phone-only, remove username tab"
```

---

## Task 10: 前端 — 注册页新增 orgDisplayName 字段

**Files:**
- Modify: `modelcraft-front/src/shared/validation/auth.ts`
- Modify: `modelcraft-front/src/app/tenant/register/page.tsx`
- Modify: `modelcraft-front/src/web/hooks/auth/use-auth-form.ts`

- [ ] **Step 1: 更新 registerFormSchema — 新增 orgDisplayName**

在 `modelcraft-front/src/shared/validation/auth.ts` 的 `registerFormSchema` 中新增字段：

```ts
export const orgDisplayNameSchema = z
  .string()
  .min(1, '请输入组织名称')
  .max(64, '组织名称最多 64 个字符')

export const registerFormSchema = z
  .object({
    phone: phoneNumberSchema,
    userName: userNameSchema,
    orgDisplayName: orgDisplayNameSchema,  // 新增
    orgName: orgNameSchema,
    password: passwordSchema,
    confirmPassword: z.string().min(1, '请确认密码'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: '两次输入的密码不一致',
    path: ['confirmPassword'],
  })

export type RegisterFormValues = z.infer<typeof registerFormSchema>
```

- [ ] **Step 2: 更新注册页 — 新增 orgDisplayName 输入框**

在 `modelcraft-front/src/app/tenant/register/page.tsx` 的 form defaultValues 中新增 `orgDisplayName: ''`，并在表单中添加输入框：

在 `form.useForm` 的 `defaultValues` 中：
```ts
defaultValues: { phone: '', userName: '', orgDisplayName: '', orgName: '', password: '', confirmPassword: '' },
```

在 `userName` 字段之后、`orgName` 字段之前插入：

```tsx
<FormField
  control={form.control}
  name="orgDisplayName"
  render={({ field }) => (
    <FormItem>
      <FormLabel>组织名称</FormLabel>
      <FormControl>
        <Input
          placeholder="请输入组织名称，如「我的公司」"
          autoComplete="organization"
          {...field}
        />
      </FormControl>
      <FormMessage />
    </FormItem>
  )}
/>
```

- [ ] **Step 3: 更新 `useRegister` hook — 传 orgDisplayName**

找到 `modelcraft-front/src/web/hooks/auth/use-auth-form.ts` 中的 `useRegister` hook，将 `orgDisplayName` 加入发送给 BFF 的 body。

- [ ] **Step 4: 检查 lint**

```bash
cd modelcraft-front && npm run lint
```

Expected: 无错误

- [ ] **Step 5: Commit**

```bash
git add modelcraft-front/src/shared/validation/auth.ts \
        modelcraft-front/src/app/tenant/register/page.tsx \
        modelcraft-front/src/web/hooks/auth/
git commit -m "feat(front): register page add orgDisplayName field"
```

---

## Task 11: 端到端验证

- [ ] **Step 1: 启动后端**

```bash
cd modelcraft-backend && just run
```

Expected: 后端在配置的端口启动，无错误日志

- [ ] **Step 2: 启动前端**

```bash
cd modelcraft-front && npm run dev
```

Expected: 前端在 http://localhost:3000 启动

- [ ] **Step 3: 验证首页**

浏览器打开 http://localhost:3000，确认：
- ✅ 只有「组织管理员」section（管理员登录、注册组织按钮）
- ❌ 「组织员工」section 已不存在

- [ ] **Step 4: 验证注册流程**

打开 http://localhost:3000/tenant/register，确认：
- ✅ 表单包含：手机号、用户名、**组织名称（orgDisplayName）**、组织 slug、密码、确认密码
- 填写全部字段后提交
- ✅ 注册成功，跳转至控制台
- ✅ 数据库 `organizations` 表有 phone 字段值
- ✅ 数据库 `users` 表有 org_name 字段值

- [ ] **Step 5: 验证登录流程**

打开 http://localhost:3000/tenant/login，确认：
- ✅ 只有手机号和密码两个输入框，无「用户名登录」Tab
- 输入注册时的手机号 + 密码
- ✅ 登录成功

- [ ] **Step 6: 验证同手机号只能注册一个 Org**

再次尝试用相同手机号注册另一个 Org：
- ✅ 返回手机号已注册错误

- [ ] **Step 7: 运行后端单元测试**

```bash
cd modelcraft-backend && go test ./... -count=1
```

Expected: 全部 PASS

- [ ] **Step 8: 最终 Commit（如有遗留变更）**

```bash
git status
# 如有未提交变更：
git add -p
git commit -m "chore: final cleanup for org-scoped user uniqueness"
```
