# Merge user_orgs into users Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `user_orgs` 表的 `is_admin` 和 `status` 两列合并进 `users` 表，删掉 `user_orgs` 表，消除所有 JOIN。

**Architecture:** 在 `users` 表新增 `is_admin TINYINT(1)` 和 `status VARCHAR(20)` 两列；把所有 `JOIN user_orgs` 的 SQL 改写为直接读 `users`；删除 `membership` domain / repository / sqlc query 层，改为在 `user.User` 上持有 `IsAdmin`/`Status`；用一张 Atlas 迁移脚本同步 schema。

**Tech Stack:** Go 1.22, sqlc, Atlas, MySQL 8, `just` task runner

---

## File Map

| 文件 | 操作 |
|------|------|
| `db/schema/mysql/06_users.sql` | Modify — users 表加 is_admin/status，删 user_orgs 建表 DDL |
| `db/queries/org.sql` | Modify — 删 user_orgs 相关 query，新增/改写 user query |
| `internal/infrastructure/dbgen/` | Re-generate via `just generate-sqlc` |
| `internal/infrastructure/dbgenwrap/safe_querier_gen.go` | Re-generate via `just generate-safe-querier` |
| `internal/domain/user/user.go` | Modify — User struct 加 IsAdmin bool / Status UserStatus |
| `internal/domain/membership/membership.go` | Delete |
| `internal/domain/membership/membership_test.go` | Delete |
| `internal/domain/membership/repository.go` | Delete |
| `internal/infrastructure/repository/sql_org_repository.go` | Modify — 删 MembershipToDomain / SqlMembershipRepository，相关 org query 改写 |
| `internal/infrastructure/repository/org_convert_test.go` | Modify — 删 TestMembershipToDomain |
| `internal/infrastructure/repository/sql_enduser_repository.go` | Modify — 所有 JOIN user_orgs 改为直接读 users |
| `internal/infrastructure/repository/sql_end_user_permission_repository.go` | Modify — IsOrgAdmin 改用 users.is_admin |
| `internal/app/organization/create_organization_service.go` | Modify — 删 membershipRepo 依赖，user 创建时直接写 is_admin/status |
| `internal/app/organization/organization_service.go` | Modify — ListByOrgWithUserName 改读 users |
| `internal/app/enduser/end_user_app_service.go` | Modify — CreateUser 的 is_admin 更新改为 UPDATE users |
| `internal/interfaces/http/routes.go` | Modify — 删 membershipRepo 初始化和传参 |
| `internal/interfaces/http/handlers/user/handler.go` | Modify — 替换 ListByUserWithDetails |

---

## Task 1: Schema 变更 — users 表加列，删 user_orgs 表

**Files:**
- Modify: `db/schema/mysql/06_users.sql`

- [ ] **Step 1: 修改 06_users.sql**

  在 `users` 表的 `org_name` 字段**之后**加入两列，并删除 `user_orgs` 整张表的建表 DDL。

  在 `users` CREATE TABLE 的 `org_name VARCHAR(36)` 行后面加：

  ```sql
  `is_admin`   TINYINT(1)  NOT NULL DEFAULT 0 COMMENT '是否为管理员',
  `status`     VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态：active | suspended',
  ```

  同时把 `idx_users_live_name` 索引改为也覆盖 status（可选，先不改也行）。

  然后**删除整个 `user_orgs` 建表块**（第 35-60 行左右，`-- 2. 用户-组织绑定表` 开始到 `COMMENT='用户-组织绑定表...';` 为止）。

- [ ] **Step 2: 确认 schema 文件只有 users / profile 表**

  ```bash
  grep -n "CREATE TABLE" modelcraft-backend/db/schema/mysql/06_users.sql
  ```

  Expected output（只剩三个表，没有 user_orgs）：
  ```
  XX: CREATE TABLE IF NOT EXISTS `users`
  XX: CREATE TABLE IF NOT EXISTS `profile`
  ```
  （ALTER TABLE organizations 也在，不是 CREATE，也没问题）

---

## Task 2: SQL Query 层 — 删 user_orgs query，给 users 新增 is_admin/status 读写

**Files:**
- Modify: `db/queries/org.sql`

- [ ] **Step 1: 删除所有 user_orgs query**

  从 `org.sql` 中删除以下所有 query（完整块，包含注释行）：

  - `CreateMembership`
  - `GetMembershipByID`
  - `GetMembershipByUserAndOrg`
  - `ListMembershipsByOrg`
  - `ListMembershipsWithUserName`
  - `ListMembershipsByUser`
  - `CountMembershipsByUser`
  - `ListMembershipsWithOrgDetails`
  - `UpdateMembership`
  - `DeleteMembership`
  - `GetUserOrgByUserID`
  - `CreateUserOrg`
  - `UpdateUserOrgAdmin`

- [ ] **Step 2: 修改 CreateUser query — 加 is_admin/status 列**

  将原来的：
  ```sql
  -- name: CreateUser :exec
  INSERT INTO users (id, external_id, name, phone, password_hash, display_name, org_name, created_at, updated_at)
  VALUES (?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));
  ```

  改为：
  ```sql
  -- name: CreateUser :exec
  INSERT INTO users (id, external_id, name, phone, password_hash, display_name, org_name, is_admin, status, created_at, updated_at)
  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));
  ```

- [ ] **Step 3: 修改 ListOrganizationsByUser query — 去掉 JOIN user_orgs**

  将原来的：
  ```sql
  -- name: ListOrganizationsByUser :many
  SELECT o.* FROM organizations o
  INNER JOIN user_orgs m ON o.name = m.org_name
  WHERE m.user_id = ? AND m.status = 'active' AND `o`.`deleted_at` = 0 ORDER BY o.created_at DESC;
  ```

  改为：
  ```sql
  -- name: ListOrganizationsByUser :many
  SELECT o.* FROM organizations o
  INNER JOIN users u ON u.org_name = o.name
  WHERE u.id = ? AND u.status = 'active' AND u.deleted_at = 0 AND `o`.`deleted_at` = 0 ORDER BY o.created_at DESC;
  ```

- [ ] **Step 4: 新增 GetUserWithMembership、UpdateUserStatus、UpdateUserAdmin、GetUserByIDWithAdmin query**

  在 org.sql 末尾追加：

  ```sql
  -- name: UpdateUserStatus :exec
  UPDATE users SET status = ?, updated_at = NOW(3)
  WHERE id = ? AND deleted_at = 0;

  -- name: UpdateUserAdmin :exec
  UPDATE users SET is_admin = ?, updated_at = NOW(3)
  WHERE id = ? AND deleted_at = 0;

  -- name: ListUsersByOrg :many
  SELECT id, user_id_alias, name, org_name, is_admin, status, created_at, updated_at
  FROM users
  WHERE org_name = ? AND deleted_at = 0
  ORDER BY created_at DESC;
  ```

  > 注意：sqlc 会根据 SELECT 字段自动生成对应的 Row struct。`GetUserByID` 已经 SELECT *，所以已包含 is_admin/status，无需新增。

- [ ] **Step 5: 新增 ListUsersWithOrgDetails query（替代 ListMembershipsWithOrgDetails）**

  ```sql
  -- name: ListUsersWithOrgDetails :many
  SELECT u.id, u.org_name, u.is_admin, u.status, u.created_at,
         o.display_name AS org_display_name
  FROM users u
  INNER JOIN organizations o ON o.name = u.org_name
  WHERE u.id = ? AND u.status = 'active' AND o.deleted_at = 0 AND u.deleted_at = 0
  ORDER BY u.created_at DESC
  LIMIT ?;
  ```

- [ ] **Step 6: 新增 ListUsersByOrgWithName query（替代 ListMembershipsWithUserName）**

  ```sql
  -- name: ListUsersByOrgWithName :many
  SELECT id, name, org_name, is_admin, status, created_at, updated_at
  FROM users
  WHERE org_name = ? AND deleted_at = 0
  ORDER BY created_at DESC;
  ```

- [ ] **Step 7: 新增 CountUsersByOrg query（替代 CountMembershipsByUser）**

  ```sql
  -- name: CountUsersByOrg :one
  SELECT COUNT(*) FROM users WHERE org_name = ? AND deleted_at = 0;
  ```

  > `CountByUser(userID)` 实际上就是"该用户属于几个 Org"，合并后永远是 1，但接口保持兼容，改为 "该 Org 下有多少用户" 来服务原来的唯一调用场景（check user already has org）。详见 Task 5。

---

## Task 3: 重新生成 sqlc 代码

**Files:**
- Modify: `internal/infrastructure/dbgen/` (auto-generated)
- Modify: `internal/infrastructure/dbgenwrap/safe_querier_gen.go` (auto-generated)

- [ ] **Step 1: 运行 sqlc 生成**

  ```bash
  cd modelcraft-backend && just generate-sqlc
  ```

  Expected: 无报错，`internal/infrastructure/dbgen/org.sql.go` 和 `models.go` 更新；`UserOrg` struct 不再存在。

- [ ] **Step 2: 重新生成 safe querier wrapper**

  ```bash
  cd modelcraft-backend && just generate-safe-querier
  ```

  Expected: `internal/infrastructure/dbgenwrap/safe_querier_gen.go` 更新。

- [ ] **Step 3: 确认 UserOrg struct 已消失**

  ```bash
  grep -r "UserOrg\b" modelcraft-backend/internal/infrastructure/dbgen/
  ```

  Expected: 无输出。

---

## Task 4: Domain 层 — 扩展 User，删除 Membership

**Files:**
- Modify: `internal/domain/user/user.go`
- Delete: `internal/domain/membership/membership.go`
- Delete: `internal/domain/membership/membership_test.go`
- Delete: `internal/domain/membership/repository.go`

- [ ] **Step 1: User struct 加 IsAdmin / Status 字段**

  在 `internal/domain/user/user.go` 的 `User` struct 中，在 `OrgName string` 之后加：

  ```go
  IsAdmin bool   // 是否为管理员
  Status  string // active | suspended
  ```

- [ ] **Step 2: NewUser 函数加 isAdmin / status 参数**

  `NewUser` 现在只用于管理员注册，所以 isAdmin=true，status="active"。修改签名和函数体：

  ```go
  func NewUser(id, userName string, phone PhoneNumber, passwordHash, orgName string) (*User, error) {
      // ... 已有校验 ...
      now := time.Now()
      user := &User{
          ID:           id,
          Name:         userName,
          Phone:        phone,
          PasswordHash: passwordHash,
          OrgName:      orgName,
          IsAdmin:      true,   // 通过 Register 创建的用户都是管理员
          Status:       "active",
          CreatedAt:    now,
          UpdatedAt:    now,
      }
      // ...
  }
  ```

  > `NewOAuthUser` 同理，加 `IsAdmin: false, Status: "active"`。

- [ ] **Step 3: 删除 membership 包**

  ```bash
  rm modelcraft-backend/internal/domain/membership/membership.go
  rm modelcraft-backend/internal/domain/membership/membership_test.go
  rm modelcraft-backend/internal/domain/membership/repository.go
  ```

- [ ] **Step 4: 尝试编译，收集编译错误列表**

  ```bash
  cd modelcraft-backend && go build ./... 2>&1 | grep "cannot\|undefined\|has no field\|no field\|declared\|imported" | head -40
  ```

  Expected: 一系列编译错误（下面各 Task 逐一修复）。

---

## Task 5: Infrastructure — sql_org_repository.go

**Files:**
- Modify: `internal/infrastructure/repository/sql_org_repository.go`
- Modify: `internal/infrastructure/repository/org_convert_test.go`

- [ ] **Step 1: 删除 MembershipToDomain 函数和 SqlMembershipRepository 全部内容**

  删除从 `// MembershipToDomain converts a dbgen.UserOrg row...` 到文件末尾的 `_ membership.MembershipRepository = (*SqlMembershipRepository)(nil)` 那行（含该行）。

- [ ] **Step 2: UserToDomain 加 IsAdmin / Status 映射**

  在 `UserToDomain` 函数的 return 语句里加：

  ```go
  return &user.User{
      // ... 已有字段 ...
      IsAdmin:      row.IsAdmin,
      Status:       row.Status,
  }
  ```

  对 `userPhoneRowToDomain` 和 `userNameRowToDomain` 做同样处理（它们的 row 类型是 `GetUserByPhoneInOrgRow` / `GetUserByNameInOrgRow`，sqlc 重新生成后这两个 Row struct 也会带上 is_admin/status 字段）。

- [ ] **Step 3: SqlUserRepository.Create 加 IsAdmin / Status 参数**

  修改 `Create` 方法中的 `CreateUserParams`：

  ```go
  params := dbgen.CreateUserParams{
      ID:           u.ID,
      ExternalID:   StrToNullStr(u.ExternalID),
      Name:         u.Name,
      Phone:        u.Phone.String(),
      PasswordHash: u.PasswordHash,
      DisplayName:  sql.NullString{},
      OrgName:      u.OrgName,
      IsAdmin:      u.IsAdmin,
      Status:       u.Status,
  }
  ```

- [ ] **Step 4: SqlOrganizationRepository.ListByUser 签名不变，SQL 已在 query 层改写（Task 2 Step 3），这里无需修改逻辑**

  但需要删除 import `"modelcraft/internal/domain/membership"` 行（如果存在）。

- [ ] **Step 5: 删除 org_convert_test.go 中的 TestMembershipToDomain 测试**

  找到并删除 `// TestMembershipToDomain verifies...` 到对应的 `}` 整个函数块。

- [ ] **Step 6: 编译确认无 membership import 残留**

  ```bash
  cd modelcraft-backend && go build ./internal/infrastructure/repository/... 2>&1
  ```

  Expected: 无错误（或只有其他 Task 尚未完成的错误）。

---

## Task 6: Infrastructure — sql_enduser_repository.go

**Files:**
- Modify: `internal/infrastructure/repository/sql_enduser_repository.go`

所有 `JOIN user_orgs uo` 的 SQL 全部改为直接读 `users` 表字段。

- [ ] **Step 1: 修改 saveExec — 删除 insertUserOrg，在 insertUser 里加 org_name/is_admin/status**

  ```go
  func (r *SqlEndUserOrgRepository) saveExec(
      ctx context.Context,
      userID, username, phone, passwordHash, _, orgName string, // userOrgID 参数保留签名不改，忽略即可（或改造调用方）
  ) error {
      const insertUser = `
          INSERT INTO users (id, name, phone, password_hash, org_name, is_admin, status, deleted_at, delete_token, created_at, updated_at)
          VALUES (?, ?, ?, ?, ?, 0, 'active', 0, 0, NOW(3), NOW(3))
      `
      if _, err := r.db.ExecContext(ctx, insertUser, userID, username, phone, passwordHash, orgName); err != nil {
          return sqlerr.WrapSQLError(err)
      }
      return nil
  }
  ```

  > `userOrgID` 参数不再需要，但调用方 `Save` 方法还在生成它（`bizutils.GenerateUUIDV7()`），只需删掉生成这行并从调用处移除。

- [ ] **Step 2: saveExec 的签名简化，同步修改 Save 调用方**

  `Save` 方法中：
  1. 删除 `userOrgID, err := bizutils.GenerateUUIDV7()` 及错误处理
  2. 调用 `saveExec(ctx, ...)` 去掉 `userOrgID` 参数

- [ ] **Step 3: saveExecTx 同样修改（删 insertUserOrg，insertUser 加 org_name/is_admin/status）**

  ```go
  func (r *SqlEndUserOrgRepository) saveExecTx(
      ctx context.Context,
      tx txExecer,
      userID, username, phone, passwordHash, orgName string,
  ) error {
      const insertUser = `
          INSERT INTO users (id, name, phone, password_hash, org_name, is_admin, status, deleted_at, delete_token, created_at, updated_at)
          VALUES (?, ?, ?, ?, ?, 0, 'active', 0, 0, NOW(3), NOW(3))
      `
      if _, err := tx.ExecContext(ctx, insertUser, userID, username, phone, passwordHash, orgName); err != nil {
          return sqlerr.WrapSQLError(err)
      }
      return nil
  }
  ```

- [ ] **Step 4: 修改所有 SELECT 中 `JOIN user_orgs` 的 query**

  每处 query 形如：
  ```sql
  SELECT u.id, u.name, u.password_hash, uo.status, uo.is_admin, u.created_at, u.updated_at, uo.org_name
  FROM users u
  JOIN user_orgs uo ON uo.user_id = u.id AND uo.org_name = ? AND uo.deleted_at = 0
  WHERE u.id = ? AND u.deleted_at = 0
  ```

  改为：
  ```sql
  SELECT u.id, u.name, u.password_hash, u.status, u.is_admin, u.created_at, u.updated_at, u.org_name
  FROM users u
  WHERE u.org_name = ? AND u.id = ? AND u.deleted_at = 0
  ```

  对应方法：`GetByID`, `GetByIDGlobal`, `GetByUsername`, `GetByPhone`, `GetByPhoneGlobal`, `GetByUsernameGlobal`。

  - `GetByIDGlobal` / `GetByPhoneGlobal` / `GetByUsernameGlobal`（无 org 过滤）改为：
    ```sql
    SELECT id, name, password_hash, status, is_admin, created_at, updated_at, org_name
    FROM users
    WHERE id = ? AND deleted_at = 0
    LIMIT 1
    ```

- [ ] **Step 5: 修改 UpdateStatus — 改为更新 users.status**

  ```go
  func (r *SqlEndUserOrgRepository) UpdateStatus(ctx context.Context, orgName, id string, isForbidden bool) error {
      if orgName == "" {
          orgName = r.orgName
      }
      status := endUserStatusActive
      if isForbidden {
          status = "suspended"
      }
      const q = `UPDATE users SET status = ?, updated_at = NOW(3) WHERE id = ? AND org_name = ? AND deleted_at = 0`
      result, err := r.db.ExecContext(ctx, q, status, id, orgName)
      if err != nil {
          return sqlerr.WrapSQLError(err)
      }
      if rows, _ := result.RowsAffected(); rows == 0 {
          return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, fmt.Sprintf("end user not found: %s", id))
      }
      return nil
  }
  ```

- [ ] **Step 6: 修改 Delete — 软删除只更新 users 一张表**

  ```go
  func (r *SqlEndUserOrgRepository) Delete(ctx context.Context, orgName, id string) error {
      if orgName == "" {
          orgName = r.orgName
      }
      const softDeleteUser = `
          UPDATE users
          SET deleted_at   = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
              delete_token = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED),
              updated_at   = NOW(3)
          WHERE id = ? AND org_name = ? AND deleted_at = 0
      `
      if _, err := r.db.ExecContext(ctx, softDeleteUser, id, orgName); err != nil {
          return sqlerr.WrapSQLError(err)
      }
      return nil
  }
  ```

- [ ] **Step 7: 修改 ListWithTotal — 去掉 JOIN user_orgs**

  countSQL 改为：
  ```sql
  SELECT COUNT(*) FROM users u WHERE u.org_name = ? AND u.deleted_at = 0
  ```

  listSQL（找到对应 SELECT）同理，用 `users` 直读字段。

- [ ] **Step 8: 编译**

  ```bash
  cd modelcraft-backend && go build ./internal/infrastructure/repository/... 2>&1
  ```

---

## Task 7: Infrastructure — sql_end_user_permission_repository.go

**Files:**
- Modify: `internal/infrastructure/repository/sql_end_user_permission_repository.go`

- [ ] **Step 1: 修改 IsOrgAdmin — 改用 users.is_admin**

  将原来的：
  ```go
  row, err := r.q.GetMembershipByUserAndOrg(ctx, dbgen.GetMembershipByUserAndOrgParams{
      UserID:  userID,
      OrgName: orgName,
  })
  if err != nil {
      if sqlerr.IsNotFoundError(err) {
          return false, nil
      }
      return false, err
  }
  return row.IsAdmin, nil
  ```

  改为（使用 sqlc 生成的 `GetUserByID` query，然后检查 org_name 和 is_admin）：
  ```go
  row, err := r.q.GetUserByID(ctx, userID)
  if err != nil {
      if sqlerr.IsNotFoundError(err) {
          return false, nil
      }
      return false, err
  }
  if row.OrgName != orgName {
      return false, nil
  }
  return row.IsAdmin, nil
  ```

- [ ] **Step 2: 编译**

  ```bash
  cd modelcraft-backend && go build ./internal/infrastructure/repository/... 2>&1
  ```

---

## Task 8: App 层 — create_organization_service.go

**Files:**
- Modify: `internal/app/organization/create_organization_service.go`

`CreateOrganizationService` 依赖 `membershipRepo`，用于在注册时创建 `user_orgs` 记录。合并后，Org 创建流程不再需要写 `user_orgs`——用户创建时 `users.is_admin=1` 和 `users.org_name` 已经记录了所有信息。

- [ ] **Step 1: 删除 membershipRepo 字段及构造参数**

  ```go
  type CreateOrganizationService struct {
      txManager repository.TxManager
      userRepo  user.UserRepository
      orgRepo   organization.OrganizationRepository
      roleRepo  domainPermission.RoleRepository
      // 删除: membershipRepo membership.MembershipRepository
  }

  func NewCreateOrganizationService(
      txManager repository.TxManager,
      userRepo user.UserRepository,
      orgRepo organization.OrganizationRepository,
      roleRepo domainPermission.RoleRepository,
      // 删除: membershipRepo membership.MembershipRepository
  ) *CreateOrganizationService {
      return &CreateOrganizationService{
          txManager: txManager,
          userRepo:  userRepo,
          orgRepo:   orgRepo,
          roleRepo:  roleRepo,
      }
  }
  ```

- [ ] **Step 2: 删除所有 membershipRepo.Create / membershipRepository.Create 调用**

  搜索并删除：
  - `membership.NewMembership(...)` 调用
  - `membershipRepo.Create(ctx, ms)` 调用
  - `membershipRepository.Create(ctx, ms)` 调用
  - `membershipID, _ := bizutils.GenerateUUIDV7()` 如果只为 membership 服务

- [ ] **Step 3: CountByUser 调用替换**

  原来用 `membershipRepo.CountByUser(ctx, userID)` 来检查"该用户是否已有 Org"。
  改为直接用 `orgRepo`：

  ```go
  // resolveUserAndCheckOrg 里：
  orgs, err := orgRepo.ListByUser(ctx, existingUser.ID)
  if err == nil && len(orgs) > 0 {
      return &CreateOrganizationOutput{
          OrganizationName: orgs[0].Name,
          AlreadyExisted:   true,
      }, nil
  }
  ```

  同理，`handleExistingOrganization` 方法里的 `s.membershipRepo.CountByUser` 也改为 `s.orgRepo.ListByUser`。

- [ ] **Step 4: 删除所有 membership import**

  ```bash
  grep -n "membership" modelcraft-backend/internal/app/organization/create_organization_service.go
  ```

  删除 `"modelcraft/internal/domain/membership"` 的 import 行，以及所有 `repository.NewSqlMembershipRepository(q)` 调用。

- [ ] **Step 5: 编译**

  ```bash
  cd modelcraft-backend && go build ./internal/app/organization/... 2>&1
  ```

---

## Task 9: App 层 — organization_service.go

**Files:**
- Modify: `internal/app/organization/organization_service.go`

- [ ] **Step 1: 替换 ListByOrgWithUserName 实现**

  原来调用 `s.membershipRepo.ListByOrgWithUserName(ctx, orgName)` 然后把 `MembershipWithUserName` 映射成 GraphQL 响应。

  改为查询 `users` 表，接口仍然返回同样结构的数据。在 organization_service.go 里：

  1. 删除 `membershipRepo membership.MembershipRepository` 字段
  2. 删除构造函数里的 membershipRepo 参数
  3. 把 `ListByOrgWithUserName` 改为调用 `userRepo.ListByOrg(ctx, orgName)`（需要先在 `user.UserRepository` 接口和 `SqlUserRepository` 实现里添加 `ListByOrg` 方法，见 Step 2）

- [ ] **Step 2: 在 user.UserRepository 接口和实现里增加 ListByOrg**

  在 `internal/domain/user/` 找到 repository interface 文件（如 `repository.go`），加：

  ```go
  // ListByOrg 返回 org 下所有活跃用户列表
  ListByOrg(ctx context.Context, orgName string) ([]*User, error)
  ```

  在 `internal/infrastructure/repository/sql_org_repository.go` 的 `SqlUserRepository` 里实现：

  ```go
  func (r *SqlUserRepository) ListByOrg(ctx context.Context, orgName string) ([]*user.User, error) {
      rows, err := r.q.ListUsersByOrgWithName(ctx, orgName)
      if err != nil {
          return nil, bizerrors.Wrapf(err, "failed to list users by org: %s", orgName)
      }
      result := make([]*user.User, len(rows))
      for i, row := range rows {
          result[i] = &user.User{
              ID:      row.ID,
              Name:    row.Name,
              OrgName: row.OrgName,
              IsAdmin: row.IsAdmin,
              Status:  row.Status,
          }
      }
      return result, nil
  }
  ```

  > `ListUsersByOrgWithName` 是 Task 2 Step 6 中加的 query，sqlc 会生成对应函数。

- [ ] **Step 3: organization_service.go 里 ListByOrgWithUserName 改为 userRepo**

  原来返回 `[]*membership.MembershipWithUserName`，现在返回结构体不变（GraphQL resolver 那边已有 mapping），只需让 service 返回 `[]orgUserEntry`（可以是 org-local struct）：

  - 找到 `organization_service.go` 里哪个函数用到了 membershipRepo，把它改为调用 `userRepo.ListByOrg`
  - 映射字段：`UserName = u.Name`, `IsAdmin = u.IsAdmin`, `Status = u.Status`

- [ ] **Step 4: 删除 membershipRepo 字段和 membership import**

  ```bash
  grep -n "membership" modelcraft-backend/internal/app/organization/organization_service.go
  ```

- [ ] **Step 5: 编译**

  ```bash
  cd modelcraft-backend && go build ./internal/app/organization/... 2>&1
  ```

---

## Task 10: App 层 — end_user_app_service.go

**Files:**
- Modify: `internal/app/enduser/end_user_app_service.go`

- [ ] **Step 1: 修改 CreateUser 的 is_admin 更新**

  原来：
  ```go
  if cmd.IsAdmin {
      const updateIsAdmin = `UPDATE user_orgs SET is_admin = 1, updated_at = NOW(3) ` +
          `WHERE user_id = ? AND org_name = ? AND deleted_at = 0`
      if _, execErr := s.db.ExecContext(ctx, updateIsAdmin, user.ID, cmd.OrgName); execErr != nil {
          return nil, bizerrors.Wrapf(execErr, "failed to set user as admin")
      }
  }
  ```

  改为：
  ```go
  if cmd.IsAdmin {
      const updateIsAdmin = `UPDATE users SET is_admin = 1, updated_at = NOW(3) WHERE id = ? AND org_name = ? AND deleted_at = 0`
      if _, execErr := s.db.ExecContext(ctx, updateIsAdmin, user.ID, cmd.OrgName); execErr != nil {
          return nil, bizerrors.Wrapf(execErr, "failed to set user as admin")
      }
  }
  ```

- [ ] **Step 2: 编译**

  ```bash
  cd modelcraft-backend && go build ./internal/app/enduser/... 2>&1
  ```

---

## Task 11: Interfaces 层 — routes.go 和 handlers

**Files:**
- Modify: `internal/interfaces/http/routes.go`
- Modify: `internal/interfaces/http/handlers/user/handler.go`

- [ ] **Step 1: routes.go 删除 membershipRepo 初始化**

  找到：
  ```go
  membershipRepo := repository.NewSqlMembershipRepository(dbgen.New(loggingDB))
  ```
  删除这行。

  同时在所有传 `membershipRepo` 的地方（第 284、306 行左右），把参数删掉。

- [ ] **Step 2: handler.go 替换 membershipRepo 依赖**

  `internal/interfaces/http/handlers/user/handler.go` 里的 `membershipRepo` 用于 `ListByUserWithDetails`（获取当前用户所属的 org + role 信息，用于 token exchange）。

  合并后，单用户只属于一个 org，`ListByUserWithDetails` 改为直接用 `userRepo` 查询：

  1. 把 handler 的 `membershipRepo membership.MembershipRepository` 字段替换为 `userRepo user.UserRepository`
  2. 构造函数参数对应修改
  3. `ListByUserWithDetails` 调用改为：

  ```go
  u, err := h.userRepo.GetByID(ctx, userID)
  if err != nil {
      // ... error handling
  }
  // 构建单条 MembershipWithDetails（保持 GraphQL 返回格式兼容）
  details := []*membership.MembershipWithDetails{
      {
          OrgName:     u.OrgName,
          IsAdmin:     u.IsAdmin,
          JoinedAt:    u.CreatedAt,
      },
  }
  ```

  > 如果 `MembershipWithDetails` 定义在 `domain/membership` 包里，而该包已被删除，则在 handler 里定义一个本地 struct 或改用 `user.User` 直接映射到 GraphQL response model。

- [ ] **Step 3: 修改 routes.go 传参给 NewHandler**

  ```go
  UserHandler: userHandlers.NewHandler(userRepo, logger),
  ```

- [ ] **Step 4: 编译**

  ```bash
  cd modelcraft-backend && go build ./internal/interfaces/... 2>&1
  ```

---

## Task 12: 全量编译 & 测试

- [ ] **Step 1: 全量编译**

  ```bash
  cd modelcraft-backend && go build ./... 2>&1
  ```

  Expected: 0 errors.

- [ ] **Step 2: 运行全部单测**

  ```bash
  cd modelcraft-backend && go test ./... 2>&1
  ```

  Expected: PASS（或只有需要 DB 的集成测试被 skip）。

- [ ] **Step 3: 运行 lint**

  ```bash
  cd modelcraft-backend && just lint 2>&1
  ```

  Expected: 无错误。

- [ ] **Step 4: Commit**

  ```bash
  git add modelcraft-backend/
  git commit -m "refactor: merge user_orgs into users table, remove membership layer"
  ```

---

## Task 13: Atlas Schema 迁移脚本（生产数据库同步）

**Files:**
- Create: `db/schema/mysql/migrations/20260610_merge_user_orgs.sql` (或根据项目 Atlas 惯例)

- [ ] **Step 1: 写 Atlas/SQL 迁移脚本**

  ```sql
  -- 迁移：user_orgs 合并进 users
  -- 方向：将 user_orgs 数据 backfill 进 users，再删表

  -- 1. 给 users 表加两列（如果尚未存在）
  ALTER TABLE users
    ADD COLUMN IF NOT EXISTS `is_admin` TINYINT(1)  NOT NULL DEFAULT 0 COMMENT '是否为管理员',
    ADD COLUMN IF NOT EXISTS `status`   VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态：active | suspended';

  -- 2. 从 user_orgs backfill 数据（以防旧数据库）
  UPDATE users u
  INNER JOIN user_orgs uo ON uo.user_id = u.id AND uo.deleted_at = 0
  SET u.is_admin = uo.is_admin,
      u.status   = uo.status,
      u.updated_at = NOW(3)
  WHERE u.deleted_at = 0;

  -- 3. 删除 user_orgs 表（确保上面 backfill 已成功后执行）
  DROP TABLE IF EXISTS user_orgs;
  ```

- [ ] **Step 2: 验证迁移脚本语法**

  ```bash
  mysql -u root -p modelcraft_dev < db/schema/mysql/migrations/20260610_merge_user_orgs.sql
  ```

  Expected: 无报错。

- [ ] **Step 3: Commit**

  ```bash
  git add db/schema/mysql/migrations/
  git commit -m "chore: add Atlas migration to drop user_orgs and backfill users"
  ```

---

## 自检清单

- [ ] `grep -r "user_orgs" modelcraft-backend/` 应为空（或只剩迁移脚本）
- [ ] `grep -r "UserOrg\b" modelcraft-backend/internal/` 应为空
- [ ] `grep -r "membership\." modelcraft-backend/internal/` 应为空（domain/membership 包已删）
- [ ] `go build ./...` 通过
- [ ] `go test ./...` 通过
- [ ] `just lint` 通过
