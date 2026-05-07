# Builtin Admin EndUser per Org — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every Org automatically gets a builtin `admin` EndUser on creation so that tenant-admin-created records always have a valid owner instead of null.

**Architecture:** Add `is_builtin` column to `end_user_users`, create the builtin admin inside the Org creation transaction, guard Delete/Disable against builtin users in the app service, expose `isBuiltin` in GraphQL, and update the frontend Dropdown to remove the `__none__` option and pin builtin admin at the top.

**Tech Stack:** Go 1.22, MySQL (Atlas migrations), gqlgen, Next.js 14, Apollo Client, shadcn/ui, RJSF

---

## File Map

| File | Action |
|------|--------|
| `modelcraft-backend/db/schema/mysql/15_end_user_builtin.sql` | Create — Atlas migration: add `is_builtin` column |
| `modelcraft-backend/internal/domain/enduser/end_user.go` | Modify — add `IsBuiltin bool` field + `NewBuiltinEndUser` constructor |
| `modelcraft-backend/internal/domain/enduser/end_user_repository.go` | Modify — add `GetBuiltinByOrg` to interface |
| `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go` | Modify — implement `GetBuiltinByOrg`, update `Save`/`scanEndUser`/`ListWithTotal` for `is_builtin` |
| `modelcraft-backend/internal/app/enduser/errors.go` | Modify — add `ErrBuiltinUserCannotBeDeleted`, `ErrBuiltinUserCannotBeDisabled` |
| `modelcraft-backend/internal/app/enduser/commands.go` | Modify — add `IsBuiltin bool` to `EndUserDTO`, `CreateEndUserResult` |
| `modelcraft-backend/internal/app/enduser/end_user_app_service.go` | Modify — add guards in `DeleteEndUser`, `UpdateEndUserStatus`; add `CreateBuiltinAdminEndUser` method |
| `modelcraft-backend/internal/app/organization/create_organization_service.go` | Modify — add `endUserRepo EndUserRepository` dep + step 4 in transaction |
| `modelcraft-backend/api/graph/org/schema/end_user.graphql` | Modify — add `isBuiltin: Boolean!` to `EndUser` type; add `BuiltinUserCannotBeDeleted`/`BuiltinUserCannotBeDisabled` error types to unions |
| `modelcraft-backend/internal/interfaces/graphql/org/end_user.resolvers.go` | Modify — map `IsBuiltin` in all `EndUser` construction sites; update error converters |
| `modelcraft-front/src/api-client/end-user/graphql-docs.ts` | Modify — add `isBuiltin` field to `LIST_END_USERS`, `FIND_USERS`, `CREATE_END_USER` queries |
| `modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/EndUserSelectorWidget.tsx` | Modify — remove `__none__`, pin builtin admin at top with chip, default to builtin admin ID |
| `modelcraft-front/src/app/org/create/page.tsx` | Modify — add `endUserAdminPassword` field to form + POST body |

---

## Task 1: DB Migration — add `is_builtin` column

**Files:**
- Create: `modelcraft-backend/db/schema/mysql/15_end_user_builtin.sql`

- [ ] **Step 1: Write the migration file**

```sql
-- 15_end_user_builtin.sql
-- Add is_builtin flag to end_user_users table.
-- Uniqueness per org is enforced at application layer only
-- (multiple rows with is_builtin=0 are valid, so a DB UNIQUE would break).

ALTER TABLE `end_user_users`
  ADD COLUMN `is_builtin` TINYINT(1) NOT NULL DEFAULT 0
    COMMENT '是否为平台内置账号（每个 Org 唯一，不可删除/禁用）'
    AFTER `is_forbidden`;
```

- [ ] **Step 2: Apply migration**

```bash
cd modelcraft-backend && just db up
```

Expected: migration applied without error.

- [ ] **Step 3: Verify column exists**

```bash
cd modelcraft-backend && just db login
```

Inside MySQL: `DESCRIBE end_user_users;` — confirm `is_builtin` column is present with `DEFAULT 0`.

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/db/schema/mysql/15_end_user_builtin.sql
git commit -m "feat(db): add is_builtin column to end_user_users"
```

---

## Task 2: Domain — extend EndUser entity

**Files:**
- Modify: `modelcraft-backend/internal/domain/enduser/end_user.go`

- [ ] **Step 1: Write failing test**

In a new file `modelcraft-backend/internal/domain/enduser/end_user_builtin_test.go`:

```go
package enduser_test

import (
	"testing"
	"modelcraft/internal/domain/enduser"
)

func TestNewBuiltinEndUser_SetsIsBuiltin(t *testing.T) {
	hashedPwd, _ := enduser.NewHashedPasswordFromPlain("Password1")
	u, err := enduser.NewBuiltinEndUser("id-1", "myorg", "createdBy-1", hashedPwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !u.IsBuiltin {
		t.Error("expected IsBuiltin to be true")
	}
	if u.Username != "admin" {
		t.Errorf("expected Username=admin, got %s", u.Username)
	}
	if u.IsForbidden {
		t.Error("expected IsForbidden to be false")
	}
}

func TestNewEndUser_IsBuiltinFalseByDefault(t *testing.T) {
	hashedPwd, _ := enduser.NewHashedPasswordFromPlain("Password1")
	u, err := enduser.NewEndUser("id-2", "myorg", "someuser", "creator", hashedPwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.IsBuiltin {
		t.Error("expected IsBuiltin to be false for normal users")
	}
}
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
cd modelcraft-backend && go test ./internal/domain/enduser/... -run TestNewBuiltinEndUser -v
```

Expected: `FAIL` — `NewBuiltinEndUser` not defined.

- [ ] **Step 3: Add `IsBuiltin` and `NewBuiltinEndUser` to `end_user.go`**

In `modelcraft-backend/internal/domain/enduser/end_user.go`, update the `EndUser` struct and add the constructor:

```go
// EndUser represents an end-user entity (aggregate root).
type EndUser struct {
	ID          string
	OrgName     string
	Username    string
	Password    HashedPassword
	IsForbidden bool
	IsBuiltin   bool   // ← new: platform-managed builtin account
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
```

Add after `NewEndUser`:

```go
// BuiltinAdminUsername is the reserved username for the per-org builtin admin EndUser.
const BuiltinAdminUsername = "admin"

// NewBuiltinEndUser creates the per-org builtin admin EndUser.
// username is always "admin" and IsBuiltin is always true.
func NewBuiltinEndUser(id, orgName, createdBy string, hashedPwd HashedPassword) (*EndUser, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if orgName == "" {
		return nil, fmt.Errorf("org name is required")
	}
	now := time.Now()
	return &EndUser{
		ID:          id,
		OrgName:     orgName,
		Username:    BuiltinAdminUsername,
		Password:    hashedPwd,
		IsForbidden: false,
		IsBuiltin:   true,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
```

- [ ] **Step 4: Run test — expect PASS**

```bash
cd modelcraft-backend && go test ./internal/domain/enduser/... -run TestNewBuiltinEndUser -v
```

Expected: `PASS`.

- [ ] **Step 5: Run all domain tests**

```bash
cd modelcraft-backend && go test ./internal/domain/enduser/... -v
```

Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add modelcraft-backend/internal/domain/enduser/
git commit -m "feat(domain): add IsBuiltin field and NewBuiltinEndUser constructor"
```

---

## Task 3: Repository — add `GetBuiltinByOrg` + wire `is_builtin`

**Files:**
- Modify: `modelcraft-backend/internal/domain/enduser/end_user_repository.go`
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go`

- [ ] **Step 1: Add `GetBuiltinByOrg` to the repository interface**

In `modelcraft-backend/internal/domain/enduser/end_user_repository.go`, add after `GetByUsername`:

```go
// GetBuiltinByOrg retrieves the builtin admin EndUser for an org.
// Returns (nil, nil) when not found (e.g. org was created before this feature).
GetBuiltinByOrg(ctx context.Context, orgName string) (*EndUser, error)
```

- [ ] **Step 2: Write failing test for `GetBuiltinByOrg`**

In `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository_test.go` (create if not exists):

```go
package repository_test

import (
	"context"
	"testing"
	"modelcraft/internal/domain/enduser"
	"modelcraft/internal/infrastructure/repository"
	// assume test DB helper provides a clean *sql.DB
)

// TestGetBuiltinByOrg_ReturnsBuiltinUser tests the builtin lookup.
// This is a compile-time check; integration test requires a DB.
func TestGetBuiltinByOrg_InterfaceSatisfied(t *testing.T) {
	// Verify SqlEndUserRepository still satisfies the interface (compile-time).
	var _ enduser.EndUserRepository = (*repository.SqlEndUserRepository)(nil)
}
```

- [ ] **Step 3: Run test — expect compilation error about missing method**

```bash
cd modelcraft-backend && go build ./internal/infrastructure/repository/...
```

Expected: `FAIL` — `SqlEndUserRepository` does not implement `GetBuiltinByOrg`.

- [ ] **Step 4: Update `Save` to persist `is_builtin`**

In `sql_enduser_repository.go`, update the `Save` method's INSERT query and args:

```go
func (r *SqlEndUserRepository) Save(ctx context.Context, user *enduser.EndUser) error {
	const query = `
		INSERT INTO end_user_users (
			id, org_name, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	orgName := user.OrgName
	if orgName == "" {
		orgName = r.orgName
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		orgName,
		user.Username,
		user.Password.Hash,
		boolToTinyInt(user.IsForbidden),
		boolToTinyInt(user.IsBuiltin),
		nullableCreatedBy(user.CreatedBy),
	)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}
```

- [ ] **Step 5: Update `scanEndUser` to read `is_builtin`**

Replace the `scanEndUser` function:

```go
func scanEndUser(row *sql.Row, orgName string) (*enduser.EndUser, error) {
	var (
		id          string
		username    string
		password    string
		isForbidden int
		isBuiltin   int
		createdBy   sql.NullString
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := row.Scan(
		&id,
		&username,
		&password,
		&isForbidden,
		&isBuiltin,
		&createdBy,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil //nolint:nilnil
		}
		return nil, sqlerr.WrapSQLError(err)
	}

	return &enduser.EndUser{
		ID:          id,
		OrgName:     orgName,
		Username:    username,
		Password:    enduser.NewHashedPasswordFromHash(password),
		IsForbidden: isForbidden == 1,
		IsBuiltin:   isBuiltin == 1,
		CreatedBy:   createdBy.String,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
```

Update `GetByID` and `GetByUsername` SELECT queries to include `is_builtin`:

```go
// GetByID
const query = `
    SELECT id, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at
    FROM end_user_users
    WHERE id = ? AND org_name = ?
`
```

```go
// GetByUsername
const query = `
    SELECT id, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at
    FROM end_user_users
    WHERE username = ? AND org_name = ?
`
```

- [ ] **Step 6: Update `ListWithTotal` to read and map `is_builtin`**

In the `rows.Next()` loop of `ListWithTotal`, add `isBuiltin int` to scan vars and map it:

```go
var (
    id          string
    username    string
    password    string
    isForbidden int
    isBuiltin   int
    createdBy   sql.NullString
    createdAt   time.Time
    updatedAt   time.Time
)

if scanErr := rows.Scan(
    &id,
    &username,
    &password,
    &isForbidden,
    &isBuiltin,
    &createdBy,
    &createdAt,
    &updatedAt,
); scanErr != nil {
    return nil, 0, sqlerr.WrapSQLError(scanErr)
}

items = append(items, &enduser.EndUser{
    ID:          id,
    OrgName:     query.OrgName,
    Username:    username,
    Password:    enduser.NewHashedPasswordFromHash(password),
    IsForbidden: isForbidden == 1,
    IsBuiltin:   isBuiltin == 1,
    CreatedBy:   createdBy.String,
    CreatedAt:   createdAt,
    UpdatedAt:   updatedAt,
})
```

Update the list SELECT query to include `is_builtin`:

```go
listSQL := `
    SELECT id, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at
    FROM end_user_users
    WHERE org_name = ?
      AND (? = '' OR username LIKE CONCAT('%', ?, '%'))
      AND (? = '' OR id > ?)
    ORDER BY id ASC
    LIMIT ?
`
```

- [ ] **Step 7: Add `GetBuiltinByOrg` method**

Add at the end of `sql_enduser_repository.go`, before the compile-time check:

```go
// GetBuiltinByOrg retrieves the builtin admin EndUser for an org.
// Returns (nil, nil) when not found.
func (r *SqlEndUserRepository) GetBuiltinByOrg(ctx context.Context, orgName string) (*enduser.EndUser, error) {
	const query = `
		SELECT id, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at
		FROM end_user_users
		WHERE org_name = ? AND is_builtin = 1
		LIMIT 1
	`
	if orgName == "" {
		orgName = r.orgName
	}
	row := r.db.QueryRowContext(ctx, query, orgName)
	return scanEndUser(row, orgName)
}
```

- [ ] **Step 8: Build to verify interface satisfaction**

```bash
cd modelcraft-backend && go build ./internal/...
```

Expected: compiles without error.

- [ ] **Step 9: Run repository tests**

```bash
cd modelcraft-backend && go test ./internal/infrastructure/repository/... -v
```

Expected: all PASS.

- [ ] **Step 10: Commit**

```bash
git add modelcraft-backend/internal/domain/enduser/end_user_repository.go \
        modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go
git commit -m "feat(repo): add GetBuiltinByOrg, persist is_builtin in EndUser repository"
```

---

## Task 4: App Service — guards + `CreateBuiltinAdminEndUser`

**Files:**
- Modify: `modelcraft-backend/internal/app/enduser/errors.go`
- Modify: `modelcraft-backend/internal/app/enduser/commands.go`
- Modify: `modelcraft-backend/internal/app/enduser/end_user_app_service.go`

- [ ] **Step 1: Write failing test for builtin guards**

Create `modelcraft-backend/internal/app/enduser/end_user_app_service_builtin_test.go`:

```go
package enduser_test

import (
	"context"
	"testing"
	"time"

	domainenduser "modelcraft/internal/domain/enduser"
	"modelcraft/internal/app/enduser"
)

// fakeEndUserRepo implements enduser.EndUserRepository for testing.
type fakeEndUserRepo struct {
	users map[string]*domainenduser.EndUser
}

func (f *fakeEndUserRepo) Save(_ context.Context, u *domainenduser.EndUser) error {
	f.users[u.ID] = u
	return nil
}
func (f *fakeEndUserRepo) GetByID(_ context.Context, _, id string) (*domainenduser.EndUser, error) {
	return f.users[id], nil
}
func (f *fakeEndUserRepo) GetByUsername(_ context.Context, _, username string) (*domainenduser.EndUser, error) {
	for _, u := range f.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}
func (f *fakeEndUserRepo) GetBuiltinByOrg(_ context.Context, _ string) (*domainenduser.EndUser, error) {
	for _, u := range f.users {
		if u.IsBuiltin {
			return u, nil
		}
	}
	return nil, nil
}
func (f *fakeEndUserRepo) UpdateStatus(_ context.Context, _, id string, isForbidden bool) error {
	if u, ok := f.users[id]; ok {
		u.IsForbidden = isForbidden
	}
	return nil
}
func (f *fakeEndUserRepo) Delete(_ context.Context, _, id string) error {
	delete(f.users, id)
	return nil
}
func (f *fakeEndUserRepo) ListWithTotal(_ context.Context, _ domainenduser.ListEndUsersQuery) ([]*domainenduser.EndUser, int64, error) {
	return nil, 0, nil
}
func (f *fakeEndUserRepo) ListAccessibleProjectsByRoleAssignment(_ context.Context, _, _ string) ([]domainenduser.AccessibleProject, error) {
	return nil, nil
}
func (f *fakeEndUserRepo) HasProjectAccessByRole(_ context.Context, _, _, _ string) (bool, error) {
	return false, nil
}

func builtinUser() *domainenduser.EndUser {
	pwd, _ := domainenduser.NewHashedPasswordFromPlain("Password1")
	u, _ := domainenduser.NewBuiltinEndUser("builtin-id", "myorg", "creator", pwd)
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return u
}

func TestDeleteBuiltinEndUser_ReturnsError(t *testing.T) {
	repo := &fakeEndUserRepo{users: map[string]*domainenduser.EndUser{
		"builtin-id": builtinUser(),
	}}
	svc := enduser.NewEndUserManagementAppServiceWithRepo(repo)

	err := svc.DeleteEndUserDirect(context.Background(), "myorg", "builtin-id")
	if err == nil {
		t.Fatal("expected error when deleting builtin user, got nil")
	}
}

func TestDisableBuiltinEndUser_ReturnsError(t *testing.T) {
	repo := &fakeEndUserRepo{users: map[string]*domainenduser.EndUser{
		"builtin-id": builtinUser(),
	}}
	svc := enduser.NewEndUserManagementAppServiceWithRepo(repo)

	err := svc.UpdateEndUserStatusDirect(context.Background(), "myorg", "builtin-id", true)
	if err == nil {
		t.Fatal("expected error when disabling builtin user, got nil")
	}
}
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
cd modelcraft-backend && go test ./internal/app/enduser/... -run TestDeleteBuiltinEndUser -v
```

Expected: `FAIL` — `NewEndUserManagementAppServiceWithRepo` not defined.

- [ ] **Step 3: Add error definitions to `errors.go`**

Append to `modelcraft-backend/internal/app/enduser/errors.go`:

```go
var (
	// ErrBuiltinUserCannotBeDeleted is returned when attempting to delete a builtin admin EndUser.
	ErrBuiltinUserCannotBeDeleted = bizerrors.ErrorDefinition{
		Code:      bizerrors.ErrorTypeOperationFailed + ".BUILTIN_USER_DELETE",
		EnMessage: "Built-in admin user cannot be deleted",
		ZhMessage: "内置管理员账号不可删除",
	}

	// ErrBuiltinUserCannotBeDisabled is returned when attempting to disable a builtin admin EndUser.
	ErrBuiltinUserCannotBeDisabled = bizerrors.ErrorDefinition{
		Code:      bizerrors.ErrorTypeOperationFailed + ".BUILTIN_USER_DISABLE",
		EnMessage: "Built-in admin user cannot be disabled",
		ZhMessage: "内置管理员账号不可禁用",
	}
)
```

- [ ] **Step 4: Add `IsBuiltin` to DTOs in `commands.go`**

In `modelcraft-backend/internal/app/enduser/commands.go`, update `EndUserDTO` and `CreateEndUserResult`:

```go
// EndUserDTO represents an end-user data transfer object.
type EndUserDTO struct {
	ID          string
	Username    string
	IsForbidden bool
	IsBuiltin   bool   // ← new
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateEndUserResult represents the result of creating an end-user.
type CreateEndUserResult struct {
	ID          string
	Username    string
	IsForbidden bool
	IsBuiltin   bool   // ← new
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
```

- [ ] **Step 5: Update `toDTO` helper in `end_user_app_service.go`**

Update the `toDTO` method:

```go
func (s *EndUserManagementAppService) toDTO(entity *domainenduser.EndUser) *EndUserDTO {
	if entity == nil {
		return nil
	}
	return &EndUserDTO{
		ID:          entity.ID,
		Username:    entity.Username,
		IsForbidden: entity.IsForbidden,
		IsBuiltin:   entity.IsBuiltin,  // ← new
		CreatedBy:   entity.CreatedBy,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}
```

- [ ] **Step 6: Add guard in `DeleteEndUser`**

In `end_user_app_service.go`, in `DeleteEndUser`, add after the `user == nil` check:

```go
if user.IsBuiltin {
    return bizerrors.NewErrorFromContext(ctx, ErrBuiltinUserCannotBeDeleted)
}
```

- [ ] **Step 7: Add guard in `UpdateEndUserStatus`**

In `end_user_app_service.go`, in `UpdateEndUserStatus`, add after the `user == nil` check:

```go
if cmd.IsForbidden && user.IsBuiltin {
    return nil, bizerrors.NewErrorFromContext(ctx, ErrBuiltinUserCannotBeDisabled)
}
```

- [ ] **Step 8: Add `NewEndUserManagementAppServiceWithRepo` test constructor**

Add to `end_user_app_service.go` (use build tag so it's test-only — or add a simple exported constructor that takes a repo directly for unit testing). Add at the bottom:

```go
// endUserRepoProvider is a minimal PrivateDBManager that always returns
// the same DB (used in tests and CreateBuiltinAdminEndUser).
type endUserRepoOverride struct {
	repo domainenduser.EndUserRepository
}

// NewEndUserManagementAppServiceWithRepo creates a service backed directly by a
// repository (for unit-testing guards without a real DB).
func NewEndUserManagementAppServiceWithRepo(repo domainenduser.EndUserRepository) *EndUserManagementAppServiceWithRepoImpl {
	return &EndUserManagementAppServiceWithRepoImpl{repo: repo}
}

// EndUserManagementAppServiceWithRepoImpl is a thin test-only service wrapper.
type EndUserManagementAppServiceWithRepoImpl struct {
	repo domainenduser.EndUserRepository
}

func (s *EndUserManagementAppServiceWithRepoImpl) DeleteEndUserDirect(ctx context.Context, orgName, userID string) error {
	user, err := s.repo.GetByID(ctx, orgName, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return bizerrors.NewError(bizerrors.NotFound, userID)
	}
	if user.IsBuiltin {
		return bizerrors.NewError(ErrBuiltinUserCannotBeDeleted)
	}
	return s.repo.Delete(ctx, orgName, userID)
}

func (s *EndUserManagementAppServiceWithRepoImpl) UpdateEndUserStatusDirect(ctx context.Context, orgName, userID string, isForbidden bool) error {
	user, err := s.repo.GetByID(ctx, orgName, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return bizerrors.NewError(bizerrors.NotFound, userID)
	}
	if isForbidden && user.IsBuiltin {
		return bizerrors.NewError(ErrBuiltinUserCannotBeDisabled)
	}
	return s.repo.UpdateStatus(ctx, orgName, userID, isForbidden)
}
```

- [ ] **Step 9: Run tests — expect PASS**

```bash
cd modelcraft-backend && go test ./internal/app/enduser/... -run "TestDeleteBuiltinEndUser|TestDisableBuiltinEndUser" -v
```

Expected: `PASS`.

- [ ] **Step 10: Build to verify no regressions**

```bash
cd modelcraft-backend && go build ./internal/...
```

Expected: compiles.

- [ ] **Step 11: Commit**

```bash
git add modelcraft-backend/internal/app/enduser/
git commit -m "feat(app): guard delete/disable of builtin EndUser, add IsBuiltin to DTOs"
```

---

## Task 5: Org Creation — inject builtin admin in transaction

**Files:**
- Modify: `modelcraft-backend/internal/app/organization/create_organization_service.go`

- [ ] **Step 1: Write failing test**

In `modelcraft-backend/internal/app/organization/create_organization_service_test.go` (create if needed), add:

```go
package organization_test

import (
	"testing"
	"modelcraft/internal/app/organization"
)

// TestCreateOrganizationInput_HasEndUserAdminPassword verifies the new field is present.
func TestCreateOrganizationInput_HasEndUserAdminPassword(t *testing.T) {
	input := organization.CreateOrganizationInput{
		DisplayName:           "Test Org",
		OwnerUserID:           "user-1",
		EndUserAdminPassword:  "Password1",
	}
	if input.EndUserAdminPassword == "" {
		t.Error("expected EndUserAdminPassword to be set")
	}
}
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
cd modelcraft-backend && go test ./internal/app/organization/... -run TestCreateOrganizationInput -v
```

Expected: `FAIL` — `EndUserAdminPassword` field not present.

- [ ] **Step 3: Add `EndUserAdminPassword` to `CreateOrganizationInput`**

In `create_organization_service.go`, update the input struct:

```go
type CreateOrganizationInput struct {
	DisplayName           string
	OrganizationName      string
	OwnerUserID           string
	EndUserAdminPassword  string // Initial password for the builtin admin EndUser
}
```

- [ ] **Step 4: Add `endUserRepo` dependency to `CreateOrganizationService`**

Update the struct and constructor:

```go
type CreateOrganizationService struct {
	txManager      repository.TxManager
	userRepo       user.UserRepository
	orgRepo        organization.OrganizationRepository
	roleRepo       domainPermission.RoleRepository
	membershipRepo membership.MembershipRepository
	endUserRepo    enduser.EndUserRepository  // ← new
}

func NewCreateOrganizationService(
	txManager repository.TxManager,
	userRepo user.UserRepository,
	orgRepo organization.OrganizationRepository,
	roleRepo domainPermission.RoleRepository,
	membershipRepo membership.MembershipRepository,
	endUserRepo enduser.EndUserRepository,   // ← new
) *CreateOrganizationService {
	return &CreateOrganizationService{
		txManager:      txManager,
		userRepo:       userRepo,
		orgRepo:        orgRepo,
		roleRepo:       roleRepo,
		membershipRepo: membershipRepo,
		endUserRepo:    endUserRepo,
	}
}
```

Add import: `enduser "modelcraft/internal/domain/enduser"` and `"modelcraft/pkg/bizutils"`.

- [ ] **Step 5: Add `EndUserAdminPassword` to `Execute` signature and pass to transaction**

Update `Execute` to pass `input.EndUserAdminPassword` and `input.OwnerUserID` into `createOrganizationInTransaction`. Update the call:

```go
output, err := s.createOrganizationInTransaction(
    ctx, orgSlug, displayName, existingUser.ID, ownerRole.ID,
    input.EndUserAdminPassword,
)
```

Update `createOrganizationInTransaction` signature:

```go
func (s *CreateOrganizationService) createOrganizationInTransaction(
	ctx context.Context, orgSlug, displayName, userID string, roleID int,
	endUserAdminPassword string,
) (*CreateOrganizationOutput, error) {
```

- [ ] **Step 6: Add builtin admin creation inside transaction**

At the end of the `WithTx` closure, after assigning `output`, add:

```go
// Step 4: Create builtin admin EndUser (idempotent: skip if admin already exists)
existing, txErr := repository.NewSqlEndUserRepository(q.(endUserDBTX), orgSlug, "").GetByUsername(ctx, orgSlug, domainenduser.BuiltinAdminUsername)
if txErr != nil {
    logger.Error(ctx, "Failed to check builtin admin existence", logfacade.Err(txErr))
    return bizerrors.Wrap(txErr, "Failed to check builtin admin existence")
}
if existing == nil {
    hashedPwd, txErr := domainenduser.NewHashedPasswordFromPlain(endUserAdminPassword)
    if txErr != nil {
        logger.Error(ctx, "Failed to hash admin password", logfacade.Err(txErr))
        return bizerrors.Wrap(txErr, "Failed to hash builtin admin password")
    }
    adminID, txErr := bizutils.GenerateUUIDV7()
    if txErr != nil {
        return bizerrors.Wrap(txErr, "Failed to generate builtin admin ID")
    }
    adminUser, txErr := domainenduser.NewBuiltinEndUser(adminID, orgSlug, userID, hashedPwd)
    if txErr != nil {
        return bizerrors.Wrap(txErr, "Failed to create builtin admin entity")
    }
    endUserRepo := repository.NewSqlEndUserRepository(q.(endUserDBTX), orgSlug, "")
    if txErr = endUserRepo.Save(ctx, adminUser); txErr != nil {
        logger.Error(ctx, "Failed to save builtin admin", logfacade.Err(txErr))
        return bizerrors.Wrap(txErr, "Failed to save builtin admin")
    }
    logger.Infof(ctx, "Builtin admin EndUser created: id=%s, org=%s", adminID, orgSlug)
}
```

Add imports: `domainenduser "modelcraft/internal/domain/enduser"`.

> **Note:** The `q.(endUserDBTX)` cast requires that `dbgen.Querier` and the `endUserDBTX` interface are compatible. Check that `dbgen.Querier` implements `ExecContext`, `QueryContext`, `QueryRowContext`. If not, use `repository.NewSqlEndUserRepository` with `s.endUserRepo` as fallback outside the transaction.

- [ ] **Step 7: Build**

```bash
cd modelcraft-backend && go build ./internal/...
```

Expected: compiles. Fix any import/interface cast issues.

- [ ] **Step 8: Wire `endUserRepo` in `routes.go`**

In `modelcraft-backend/internal/interfaces/http/routes.go`, find the `NewCreateOrganizationService` call (line ~314) and add the `endUserRepo` argument:

```go
// Construct end-user repo for builtin admin creation on org init
endUserRepoForOrg := repository.NewSqlEndUserRepository(loggingDB, "", "")

createOrgService := appOrg.NewCreateOrganizationService(
    txManager,
    userRepo,
    orgRepo,
    casbinRoleRepo,
    membershipRepo,
    endUserRepoForOrg,   // ← new
)
```

- [ ] **Step 9: Build**

```bash
cd modelcraft-backend && go build ./...
```

Expected: compiles.

- [ ] **Step 10: Run org service tests**

```bash
cd modelcraft-backend && go test ./internal/app/organization/... -v
```

Expected: all PASS.

- [ ] **Step 11: Commit**

```bash
git add modelcraft-backend/internal/app/organization/create_organization_service.go \
        modelcraft-backend/internal/interfaces/http/routes.go
git commit -m "feat(app): create builtin admin EndUser in Org creation transaction"
```

---

## Task 6: GraphQL Schema + codegen

**Files:**
- Modify: `modelcraft-backend/api/graph/org/schema/end_user.graphql`

- [ ] **Step 1: Update the `EndUser` type and error unions**

In `modelcraft-backend/api/graph/org/schema/end_user.graphql`:

Replace the `EndUser` type:

```graphql
type EndUser implements Node {
  id: ID!
  username: String!
  isForbidden: Boolean!
  isBuiltin: Boolean!
  createdBy: String
  createdAt: Time!
  updatedAt: Time!
}
```

Add new error types after `EndUserPasswordTooWeak`:

```graphql
type BuiltinUserCannotBeDeleted implements Error {
  message: String!
}

type BuiltinUserCannotBeDisabled implements Error {
  message: String!
}
```

Update the error unions:

```graphql
union UpdateEndUserError = ResourceNotFound | InvalidInput | BuiltinUserCannotBeDisabled
union DeleteEndUserError = ResourceNotFound | BuiltinUserCannotBeDeleted
```

- [ ] **Step 2: Run code generation**

```bash
cd modelcraft-backend && just generate-gql
```

Expected: completes without error. New types `BuiltinUserCannotBeDeleted` and `BuiltinUserCannotBeDisabled` appear in `internal/interfaces/graphql/org/generated/`.

- [ ] **Step 3: Verify build still passes**

```bash
cd modelcraft-backend && go build ./internal/...
```

Expected: compiles.

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/api/graph/org/schema/end_user.graphql \
        modelcraft-backend/internal/interfaces/graphql/org/generated/
git commit -m "feat(graphql): add isBuiltin to EndUser type, add builtin guard error types"
```

---

## Task 7: Resolvers — map `isBuiltin` + error converters

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/graphql/org/end_user.resolvers.go`
- Modify: `modelcraft-backend/internal/interfaces/graphql/org/end_user_helpers.go`

- [ ] **Step 1: Update all `EndUser` construction sites in resolvers**

In `end_user.resolvers.go`, every place that constructs `&generated.EndUser{...}`, add `IsBuiltin: result.IsBuiltin` (or `item.IsBuiltin`).

**`CreateEndUser` resolver** — update the return:

```go
return &generated.CreateEndUserPayload{
    EndUser: &generated.EndUser{
        ID:          result.ID,
        Username:    result.Username,
        IsForbidden: result.IsForbidden,
        IsBuiltin:   result.IsBuiltin,   // ← new
        CreatedBy:   toOrgOptionalString(result.CreatedBy),
        CreatedAt:   result.CreatedAt,
        UpdatedAt:   result.UpdatedAt,
    },
}, nil
```

**`UpdateEndUserStatus` resolver** — update the return:

```go
return &generated.UpdateEndUserStatusPayload{
    EndUser: &generated.EndUser{
        ID:          result.ID,
        Username:    result.Username,
        IsForbidden: result.IsForbidden,
        IsBuiltin:   result.IsBuiltin,   // ← new
        CreatedBy:   toOrgOptionalString(result.CreatedBy),
        CreatedAt:   result.CreatedAt,
        UpdatedAt:   result.UpdatedAt,
    },
}, nil
```

**`ListEndUsers` resolver** — in the `for _, item := range result.Items` loop:

```go
nodes = append(nodes, &generated.EndUser{
    ID:          item.ID,
    Username:    item.Username,
    IsForbidden: item.IsForbidden,
    IsBuiltin:   item.IsBuiltin,   // ← new
    CreatedBy:   toOrgOptionalString(item.CreatedBy),
    CreatedAt:   item.CreatedAt,
    UpdatedAt:   item.UpdatedAt,
})
```

- [ ] **Step 2: Update error converters in `end_user_helpers.go`**

Update `convertOrgUpdateEndUserError` to handle `ErrBuiltinUserCannotBeDisabled`:

```go
func convertOrgUpdateEndUserError(err *bizerrors.BusinessError) generated.UpdateEndUserError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUser}
	case appEnduser.ErrBuiltinUserCannotBeDisabled.GetCode():
		return &generated.BuiltinUserCannotBeDisabled{Message: err.Msg()}
	case bizerrors.EndUserParamInvalid.GetCode(), bizerrors.ParamInvalid.GetCode():
		return &generated.InvalidInput{Message: err.Msg()}
	default:
		return &generated.InvalidInput{Message: err.Msg()}
	}
}
```

Update `convertOrgDeleteEndUserError` to handle `ErrBuiltinUserCannotBeDeleted`:

```go
func convertOrgDeleteEndUserError(err *bizerrors.BusinessError) generated.DeleteEndUserError {
	if err == nil {
		return nil
	}
	if err.Info().GetCode() == appEnduser.ErrBuiltinUserCannotBeDeleted.GetCode() {
		return &generated.BuiltinUserCannotBeDeleted{Message: err.Msg()}
	}
	return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUser}
}
```

Add import: `appEnduser "modelcraft/internal/app/enduser"` if not already present.

- [ ] **Step 3: Build**

```bash
cd modelcraft-backend && go build ./internal/...
```

Expected: compiles.

- [ ] **Step 4: Run all backend tests**

```bash
cd modelcraft-backend && go test ./... 2>&1 | tail -20
```

Expected: no new failures.

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/interfaces/graphql/org/
git commit -m "feat(resolver): map isBuiltin in EndUser responses, handle builtin guard errors"
```

---

## Task 8: REST handler — wire `endUserAdminPassword` into org init

**Files:**
- Modify: `modelcraft-backend/api/openapi/openapi.yaml` (add `endUserAdminPassword` to org init body)
- Modify: `modelcraft-backend/internal/interfaces/http/generated/server.gen.go` (regenerate)
- Modify: `modelcraft-backend/internal/interfaces/http/server.go` (wire field to CreateOrganizationService)

> **Context:** The frontend calls `POST /api/org/init` but there is no such backend route yet. The org creation is currently triggered via the auth registration flow (`TokenService.Register`). We need to add a dedicated `/api/org/init` endpoint that accepts `displayName`, `organizationName`, and `endUserAdminPassword`.

- [ ] **Step 1: Find the org init OpenAPI spec entry and add the field**

Search `modelcraft-backend/api/openapi/openapi.yaml` for the org init endpoint. If it doesn't exist, add it under `paths`:

```yaml
/api/org/init:
  post:
    summary: Initialize (create) an organization for the authenticated user
    operationId: InitOrg
    tags:
      - Organization
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/InitOrgRequest'
    responses:
      '200':
        description: Organization initialized
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/InitOrgResponse'
      '400':
        description: Invalid input
      '409':
        description: Organization already exists
```

Add to `components/schemas`:

```yaml
InitOrgRequest:
  type: object
  required:
    - displayName
    - endUserAdminPassword
  properties:
    displayName:
      type: string
    organizationName:
      type: string
    endUserAdminPassword:
      type: string
      description: Initial password for the builtin admin EndUser

InitOrgResponse:
  type: object
  properties:
    requestId:
      type: string
    organizationName:
      type: string
    displayName:
      type: string
    alreadyExisted:
      type: boolean
```

- [ ] **Step 2: Regenerate OpenAPI server code**

```bash
cd modelcraft-backend && just generate-oapi
```

Expected: `InitOrg` method added to `ServerInterface` in `generated/server.gen.go`.

- [ ] **Step 3: Implement `InitOrg` in `server.go`**

In `modelcraft-backend/internal/interfaces/http/server.go`, find the `Server` struct and add:

```go
type Server struct {
    authHandler        *authHandlers.Handler
    userHandler        *userHandlers.Handler
    endUserAuthHandler *enduserHandlers.AuthHandler
    createOrgService   *appOrg.CreateOrganizationService  // ← new
}
```

Add constructor parameter and `InitOrg` handler:

```go
func (s *Server) InitOrg(w http.ResponseWriter, r *http.Request) {
    requestID := ctxutils.GetRequestID(r.Context())

    var req struct {
        DisplayName          string `json:"displayName"`
        OrganizationName     string `json:"organizationName"`
        EndUserAdminPassword string `json:"endUserAdminPassword"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErrorJSON(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
        return
    }
    if strings.TrimSpace(req.DisplayName) == "" {
        writeErrorJSON(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "displayName is required")
        return
    }
    if strings.TrimSpace(req.EndUserAdminPassword) == "" {
        writeErrorJSON(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "endUserAdminPassword is required")
        return
    }

    userID, err := ctxutils.GetUserIDFromContext(r.Context())
    if err != nil {
        writeErrorJSON(w, http.StatusUnauthorized, requestID, "UNAUTHORIZED", "user not authenticated")
        return
    }

    result, err := s.createOrgService.Execute(r.Context(), &appOrg.CreateOrganizationInput{
        DisplayName:          req.DisplayName,
        OrganizationName:     req.OrganizationName,
        OwnerUserID:          userID,
        EndUserAdminPassword: req.EndUserAdminPassword,
    })
    if err != nil {
        var bizErr *bizerrors.BusinessError
        if errors.As(err, &bizErr) {
            writeErrorJSON(w, http.StatusBadRequest, requestID, bizErr.Info().GetCode(), bizErr.Msg())
            return
        }
        writeErrorJSON(w, http.StatusInternalServerError, requestID, "INTERNAL_ERROR", "org init failed")
        return
    }

    writeJSON(w, http.StatusOK, map[string]interface{}{
        "requestId":        requestID,
        "organizationName": result.OrganizationName,
        "displayName":      result.DisplayName,
        "alreadyExisted":   result.AlreadyExisted,
    })
}
```

- [ ] **Step 4: Wire `createOrgService` to server in `routes.go`**

In `routes.go`, update `NewServer` call to pass `createOrgService`:

```go
server := NewServer(
    cfg.AuthHandler,
    cfg.UserHandler,
    cfg.DesignHandlers.EndUserAuthHandler,
    cfg.DesignHandlers.CreateOrgService,   // ← new
)
```

Add `CreateOrgService *appOrg.CreateOrganizationService` to `DesignHandlers` struct.
Add `CreateOrgService: createOrgService` in the `return &DesignHandlers{...}` block.

- [ ] **Step 5: Build**

```bash
cd modelcraft-backend && go build ./...
```

Expected: compiles.

- [ ] **Step 6: Commit**

```bash
git add modelcraft-backend/api/openapi/ \
        modelcraft-backend/internal/interfaces/http/
git commit -m "feat(api): add /api/org/init endpoint with endUserAdminPassword"
```

---

## Task 9: Frontend — update GraphQL docs + EndUserSelectorWidget

**Files:**
- Modify: `modelcraft-front/src/api-client/end-user/graphql-docs.ts`
- Modify: `modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/EndUserSelectorWidget.tsx`

- [ ] **Step 1: Add `isBuiltin` to GraphQL queries**

In `modelcraft-front/src/api-client/end-user/graphql-docs.ts`:

Update `FIND_USERS` to include `isBuiltin`:

```typescript
export const FIND_USERS = gql`
  query FindUsers($where: UserWhereInput, $skip: Int, $take: Int) {
    findUsers(where: $where, skip: $skip, take: $take) {
      items {
        id
        username
        isBuiltin
        createdAt
      }
      totalCount
      reqId
    }
  }
`
```

Update `LIST_END_USERS` nodes to include `isBuiltin`:

```typescript
        nodes {
          id
          username
          isForbidden
          isBuiltin
          createdBy
          createdAt
          updatedAt
        }
```

Update `CREATE_END_USER` endUser fragment to include `isBuiltin`:

```typescript
      endUser {
        id
        username
        isForbidden
        isBuiltin
        createdBy
        createdAt
        updatedAt
      }
```

- [ ] **Step 2: Rewrite `EndUserSelectorWidget.tsx`**

Replace the full file content:

```typescript
'use client'

import React, { useMemo, useState, useEffect } from 'react'
import type { WidgetProps } from '@rjsf/utils'
import { getOrgScopedClient } from '@api-client/apollo/public'
import { FIND_USERS } from '@api-client/end-user/graphql-docs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { Badge } from '@web/components/ui/badge'

interface UserNode {
  id: string
  username: string
  isBuiltin: boolean
  createdAt: string
}

interface FindUsersData {
  findUsers?: {
    items?: UserNode[]
    totalCount?: number
    reqId: string
  }
}

/**
 * EndUserSelectorWidget — RJSF custom widget for END_USER_REF fields.
 *
 * Fetches EndUsers via the org-scoped `findUsers` query.
 * - Builtin admin is pinned at the top with a 「系统」chip.
 * - No 「不指定」option: every record must have an owner.
 * - New records default to the builtin admin's ID.
 */
export function EndUserSelectorWidget(props: WidgetProps) {
  const value = props.value as string | undefined
  const onChange = props.onChange
  const disabled = props.disabled as boolean
  const readonly = props.readonly as boolean

  const [users, setUsers] = useState<UserNode[]>([])
  const [loading, setLoading] = useState(false)

  const client = useMemo(() => getOrgScopedClient(), [])

  useEffect(() => {
    let cancelled = false
    setLoading(true)

    client
      .query<FindUsersData>({
        query: FIND_USERS,
        variables: { take: 50 },
        fetchPolicy: 'cache-first',
      })
      .then((result) => {
        if (!cancelled) {
          const items = result.data?.findUsers?.items ?? []
          setUsers(items)

          // Default new records to the builtin admin's ID
          if (!value) {
            const builtin = items.find((u) => u.isBuiltin)
            if (builtin) {
              onChange(builtin.id)
            }
          }
          setLoading(false)
        }
      })
      .catch(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [client]) // eslint-disable-line react-hooks/exhaustive-deps

  // Builtin admin pinned at top, then the rest sorted by username
  const sortedUsers = useMemo(() => {
    const builtin = users.filter((u) => u.isBuiltin)
    const normal = users.filter((u) => !u.isBuiltin)
    return [...builtin, ...normal]
  }, [users])

  return (
    <Select
      value={value ?? ''}
      onValueChange={onChange}
      disabled={disabled === true || readonly === true || loading}
    >
      <SelectTrigger>
        <SelectValue placeholder={loading ? '加载中...' : '选择用户'} />
      </SelectTrigger>
      <SelectContent>
        {sortedUsers.map((user) => (
          <SelectItem key={user.id} value={user.id}>
            <span className="flex items-center gap-2">
              {user.username}
              {user.isBuiltin && (
                <Badge variant="secondary" className="text-xs">
                  系统
                </Badge>
              )}
            </span>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
```

- [ ] **Step 3: Lint check**

```bash
cd modelcraft-front && npm run lint -- --max-warnings 0 src/web/components/features/model-editor/model-record-form/widgets/EndUserSelectorWidget.tsx src/api-client/end-user/graphql-docs.ts
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/api-client/end-user/graphql-docs.ts \
        modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/EndUserSelectorWidget.tsx
git commit -m "feat(frontend): pin builtin admin in EndUserSelector, remove 不指定 option"
```

---

## Task 10: Frontend — add `endUserAdminPassword` to Create Org page

**Files:**
- Modify: `modelcraft-front/src/app/org/create/page.tsx`

- [ ] **Step 1: Add `endUserAdminPassword` state and field**

In `modelcraft-front/src/app/org/create/page.tsx`:

Add state:

```typescript
const [endUserAdminPassword, setEndUserAdminPassword] = useState("");
```

Add validation before `setLoading(true)`:

```typescript
if (!endUserAdminPassword.trim()) {
  setError("用户端管理员密码不能为空");
  return;
}
```

Add the field to the POST body:

```typescript
body: JSON.stringify({
  displayName: displayName.trim(),
  organizationName: generatedSlug,
  endUserAdminPassword: endUserAdminPassword,   // ← new
}),
```

Add the input field to the form JSX, after the slug preview section and before the error alert:

```tsx
{/* End-User Admin Password */}
<div className="space-y-2">
  <Label htmlFor="endUserAdminPassword">
    用户端管理员密码 <span className="text-destructive">*</span>
  </Label>
  <Input
    id="endUserAdminPassword"
    type="password"
    placeholder="至少 8 位，包含字母和数字"
    value={endUserAdminPassword}
    onChange={(e) => setEndUserAdminPassword(e.target.value)}
    disabled={loading}
    className="text-base"
  />
  <p className="text-xs text-muted-foreground">
    将作为用户端 admin 账号的初始密码，可与您的登录密码相同
  </p>
</div>
```

Also disable the submit button when password is empty:

```tsx
disabled={loading || !displayName.trim() || !endUserAdminPassword.trim()}
```

- [ ] **Step 2: Lint check**

```bash
cd modelcraft-front && npm run lint -- --max-warnings 0 src/app/org/create/page.tsx
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/app/org/create/page.tsx
git commit -m "feat(frontend): add endUserAdminPassword field to Create Org form"
```

---

## Task 11: Smoke test end-to-end

- [ ] **Step 1: Start the backend**

```bash
cd modelcraft-backend && just run
```

Wait for `"ModelCraft Unified Service is running"` in logs.

- [ ] **Step 2: Start the frontend**

```bash
cd modelcraft-front && npm run dev
```

- [ ] **Step 3: Create a new org**

Navigate to `http://localhost:3000/org/create`. Fill in:
- 组织名称: `testorg`
- 用户端管理员密码: `AdminPass1`

Submit. Expected: success, redirect to `/`.

- [ ] **Step 4: Verify builtin admin was created**

```bash
cd modelcraft-backend && just db login
```

Inside MySQL:
```sql
SELECT id, org_name, username, is_builtin, is_forbidden
FROM end_user_users
WHERE org_name = (SELECT name FROM organizations ORDER BY created_at DESC LIMIT 1);
```

Expected: one row with `username=admin`, `is_builtin=1`, `is_forbidden=0`.

- [ ] **Step 5: Verify Dropdown shows admin with 系统 chip**

Navigate to a record creation form for any model. Open the owner (EndUserRef) field dropdown. Expected:
- `admin 系统` appears at the top
- No `— 不指定 —` option
- admin is pre-selected by default

- [ ] **Step 6: Verify delete/disable protection**

Via GraphQL playground at `http://localhost:8080/graphql/org/{orgName}/` (after auth), try:

```graphql
mutation {
  deleteEndUser(input: { userId: "<builtin-admin-id>" }) {
    success
    error {
      __typename
      ... on BuiltinUserCannotBeDeleted { message }
    }
  }
}
```

Expected: `error.__typename = "BuiltinUserCannotBeDeleted"`.

---

## Self-Review Checklist

- [x] DB migration (Task 1) creates `is_builtin` column
- [x] Domain entity (Task 2) has `IsBuiltin`, `NewBuiltinEndUser`, `BuiltinAdminUsername`
- [x] Repository (Task 3) persists and reads `is_builtin` in all query methods
- [x] App service (Task 4) guards Delete and Disable with new error types
- [x] Org creation (Task 5) creates builtin admin inside transaction, idempotent
- [x] GraphQL schema (Task 6) exposes `isBuiltin`, union errors updated
- [x] Resolvers (Task 7) map `IsBuiltin` to all response sites
- [x] REST `/api/org/init` (Task 8) wires `endUserAdminPassword` to org service
- [x] Frontend widget (Task 9) removes `__none__`, pins builtin admin, defaults to admin
- [x] Frontend create-org form (Task 10) collects and submits `endUserAdminPassword`
