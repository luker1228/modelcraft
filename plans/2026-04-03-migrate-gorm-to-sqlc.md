# Plan: Migrate GORM to sqlc

**Date:** 2026-04-03  
**Scope:** `modelcraft-backend`  
**Goal:** Completely remove GORM from the codebase. All DB access must go through sqlc-generated code (`internal/infrastructure/dbgen`).

---

## Current State

All sqlc query files already exist in `db/queries/` and generated code is in `internal/infrastructure/dbgen/`. Several sqlc-based repositories already exist. GORM is only used in legacy paths.

### Remaining GORM Usages

| File | What it does | Action |
|------|-------------|--------|
| `cmd/server/main.go` | `NewGormConnection()` → `*gorm.DB`, then `.DB()` to extract `*sql.DB` | Swap to `NewSQLConnection()` |
| `repository/gorm_connection.go` | Creates `*gorm.DB` from config | **Delete** |
| `repository/gorm_logger.go` | GORM logger adapter | **Delete** |
| `repository/connection_factory.go` | Wraps `*sql.DB` in GORM; exposes `DB *gorm.DB` | Remove `DB` field; keep `SqlDB` only |
| `repository/user_model.go` | `GormUserRepository` + `UserModel` PO | **Delete** (sqlc impl exists in `sql_org_repository.go`) |
| `repository/organization_model.go` | `GormOrganizationRepository` + `OrganizationModel` PO | **Delete** (sqlc impl exists) |
| `repository/membership_model.go` | `GormMembershipRepository` + `MembershipModel` PO + helpers | **Delete** (sqlc impl exists) |
| `repository/role_model.go` | `GormRoleRepository` + `RoleModel` PO | **Replace** with new `SqlRoleRepository` |
| `repository/field_enum_association_model.go` | `FieldEnumAssociationPO` with `gorm:` tags (no repo) | Remove GORM tags only |
| `repository/model_group_model.go` | `ModelGroupPO` with `gorm:` tags (no repo) | Remove GORM tags only |
| `repository/membership_with_username_test.go` | Test using GORM in-memory SQLite | **Delete** |
| `interfaces/http/routes.go` | `NewGormRoleRepository(repoFactory.DB)` + `repoFactory.DB` → `ReverseEngineerAppService` | Switch to sqlc |
| `app/modeldesign/reverse_engineer_app.go` | Holds `db *gorm.DB`, wraps sqlDB in GORM before calling `IntrospectTable` | Change to `*sql.DB` |
| `database/ddl/introspector.go` | Interface + impl takes `*gorm.DB` | Refactor to `*sql.DB` |
| `domain/modeldesign/comparison_service.go` | Dead code: `ClusterConnectionManagerCompat` + `UseClause` import GORM | Delete those two + remove imports |

---

## Key Observations

### User / Organization / Membership: already migrated
`sql_org_repository.go` already contains `SqlOrganizationRepository`, `SqlUserRepository`, `SqlMembershipRepository`. `routes.go` already calls the `NewSql*` constructors. The GORM versions are dead production code — only referenced by one test file.

### Role repository: needs new sqlc implementation
`routes.go:169` still calls `repository.NewGormRoleRepository(repoFactory.DB)`. `casbin.sql.go` already has all needed queries: `GetRoleByID`, `GetRoleByName`, `GetSystemRoleByName`, `ListRoles`, `CreateRole` (`:execresult`), `UpdateRole`, `DeleteRole`.

**Schema change:** Old `RoleModel` had a `permissions` JSON column. New `roles` table has **no permissions column** — permissions live in `role_permissions` (Casbin). Therefore `role.Role.Permissions` must be set to `[]string{}` in the new repo.

**ID type mismatch:** `dbgen.Role.ID` is `int64` (BIGINT AUTO_INCREMENT). `role.Role.ID` is `string`. Convert with `strconv.FormatInt` / `strconv.ParseInt`.

### `introspector.go`: pure SQL refactor
All three queries can be done with `db.QueryRowContext` / `db.QueryContext` using `database/sql`.

### Dead code in `comparison_service.go`
`ClusterConnectionManagerCompat` and `UseClause` are never called outside their file. Safe to delete.

---

## Go Module Dependencies to Remove

After migration, `go mod tidy` removes:
- `gorm.io/gorm`
- `gorm.io/driver/mysql`
- `gorm.io/driver/sqlite` (only in deleted test)

---

## Implementation Tasks

### Task 1 — Refactor `ddl.SchemaIntrospector` to use `*sql.DB`

**File:** `internal/infrastructure/database/ddl/introspector.go`

1. Change interface: `IntrospectTable(ctx context.Context, db *sql.DB, tableName string)`
2. Change all private methods to take `*sql.DB`
3. Replace GORM raw query calls with standard `database/sql`:
   - `db.Raw("SELECT DATABASE()").Scan(&dbName).Error` → `db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&dbName)`
   - `db.Raw(query, args...).Scan(&slice).Error` → `rows, err := db.QueryContext(...)` + manual scan loop
   - `db.Raw(query, args...).Rows()` → `db.QueryContext(ctx, query, args...)`
4. Remove `gorm.io/gorm` import; keep `"database/sql"`

### Task 2 — Refactor `ReverseEngineerAppService` to use `*sql.DB`

**File:** `internal/app/modeldesign/reverse_engineer_app.go`

1. Change `db *gorm.DB` field → `db *sql.DB`
2. Change constructor parameter `db *gorm.DB` → `db *sql.DB`
3. In `getTableDefinition`, `sqlDB` already comes from `clusterManager.GetConnectionWithDatabase` (returns `*sql.DB`). Remove the GORM wrapping:
   ```go
   // DELETE:
   gormDB, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB}), &gorm.Config{})
   if err != nil { return nil, ... }
   tableDef, err := s.schemaIntrospector.IntrospectTable(ctx, gormDB, cmd.TableName)
   // REPLACE WITH:
   tableDef, err := s.schemaIntrospector.IntrospectTable(ctx, sqlDB, cmd.TableName)
   ```
4. The struct's `db` field is no longer used (connection comes from clusterManager) — remove it from struct and constructor entirely.
5. Remove `gorm.io/driver/mysql` and `gorm.io/gorm` imports.

### Task 3 — Create `SqlRoleRepository`

**File (new):** `internal/infrastructure/repository/sql_role_repository.go`

```go
package repository

import (
    "context"
    "strconv"

    "modelcraft/internal/domain/role"
    "modelcraft/internal/domain/shared"
    "modelcraft/internal/infrastructure/dbgen"
)

type SqlRoleRepository struct{ q dbgen.Querier }

func NewSqlRoleRepository(q dbgen.Querier) role.RoleRepository {
    return &SqlRoleRepository{q: q}
}

func roleToDomain(row dbgen.Role) *role.Role {
    desc := ""
    if row.Description.Valid { desc = row.Description.String }
    return &role.Role{
        ID:          strconv.FormatInt(row.ID, 10),
        Name:        row.Name,
        Description: desc,
        Permissions: []string{},   // permissions column removed; Casbin handles this
        IsSystem:    row.IsSystem,
        CreatedAt:   row.CreatedAt,
        UpdatedAt:   row.UpdatedAt,
    }
}
```

Method implementations (use `QueryWithSQLErrorHandling` / `ExecWithErrorHandling`):

- `GetByID(ctx, id string)` — `strconv.ParseInt(id, 10, 64)` → `q.GetRoleByID(ctx, int64)` — NotFound → `shared.NewNotFoundError`
- `GetByName(ctx, name)` — `q.GetRoleByName(ctx, name)` — NotFound → error
- `GetSystemRoleByName(ctx, name)` — `q.GetSystemRoleByName(ctx, name)` — NotFound → error
- `List(ctx)` — `q.ListRoles(ctx)` → convert all
- `Create(ctx, entity)` — `q.CreateRole(ctx, dbgen.CreateRoleParams{Name, Description: sql.NullString{...}, IsSystem, OrgName})` returns `sql.Result`; call `.LastInsertId()` and set `entity.ID = strconv.FormatInt(id, 10)`
- `Update(ctx, entity)` — parse entity.ID, `q.UpdateRole(ctx, dbgen.UpdateRoleParams{Description: ..., ID: int64})`
- `Delete(ctx, id)` — parse id, `q.DeleteRole(ctx, int64)`

### Task 4 — Simplify `connection_factory.go`

**File:** `internal/infrastructure/repository/connection_factory.go`

Replace entire file:
```go
package repository

import "database/sql"

// ConnectionFactory holds the database connection for repository construction.
type ConnectionFactory struct {
    SqlDB *sql.DB
}

// NewConnectionFactory creates a ConnectionFactory from a *sql.DB connection.
func NewConnectionFactory(sqlDB *sql.DB) *ConnectionFactory {
    return &ConnectionFactory{SqlDB: sqlDB}
}
```

Signature changes from `(*ConnectionFactory, error)` to `*ConnectionFactory` — update `main.go` accordingly.

### Task 5 — Update `routes.go`

**File:** `internal/interfaces/http/routes.go`

1. Change `repoFactory.DB` → `repoFactory.SqlDB` at the `NewReverseEngineerAppService` call:
   ```go
   reverseEngineerApp := modeldesign.NewReverseEngineerAppService(
       appService, clusterManager, clusterRepository, modelRepository,
       // db *gorm.DB parameter removed in Task 2 — no db arg at all
   )
   ```
   (After Task 2, `NewReverseEngineerAppService` no longer takes a `db` argument at all.)

2. Replace lines ~169-170:
   ```go
   // OLD:
   legacyRoleRepo := repository.NewGormRoleRepository(repoFactory.DB)
   // NEW:
   roleRepo := repository.NewSqlRoleRepository(dbgen.New(loggingDB))
   roleAppService := appRole.NewRoleAppService(roleRepo)
   ```

3. Remove any remaining `gorm.io/*` imports.

### Task 6 — Update `main.go`

**File:** `cmd/server/main.go`

Replace the GORM-based DB initialization:
```go
// OLD:
db, err := repository.NewGormConnection(&cfg.Database)
if err != nil { log.Fatal(...) }
sqlDb, err := db.DB()
if err != nil { log.Fatal(...) }
defer sqlDb.Close()
// ...
repoFactory := repository.NewConnectionFactory(db)

// NEW:
sqlDb, err := repository.NewSQLConnection(&cfg.Database)
if err != nil { log.Fatal("Database connection failed", logfacade.Err(err)) }
defer sqlDb.Close()
// ...
repoFactory := repository.NewConnectionFactory(sqlDb)
```

Remove `gorm.io/driver/mysql`, `gorm.io/gorm` imports.

### Task 7 — Delete obsolete files

```bash
rm modelcraft-backend/internal/infrastructure/repository/gorm_connection.go
rm modelcraft-backend/internal/infrastructure/repository/gorm_logger.go
rm modelcraft-backend/internal/infrastructure/repository/user_model.go
rm modelcraft-backend/internal/infrastructure/repository/organization_model.go
rm modelcraft-backend/internal/infrastructure/repository/membership_model.go
rm modelcraft-backend/internal/infrastructure/repository/membership_with_username_test.go
rm modelcraft-backend/internal/infrastructure/repository/role_model.go
```

### Task 8 — Clean GORM remnants in remaining files

**`repository/field_enum_association_model.go`:**
- Remove all `gorm:"..."` struct tags (leave other struct tags untouched)
- Remove `gorm.io/gorm` import

**`repository/model_group_model.go`:**
- Remove all `gorm:"..."` struct tags
- Remove `gorm.io/gorm` import

**`domain/modeldesign/comparison_service.go`:**
- Delete `ClusterConnectionManagerCompat` struct + `DB()` method (dead code)
- Delete `UseClause` function (dead code)
- Remove imports: `gorm.io/driver/mysql`, `gorm.io/gorm`, `gorm.io/gorm/clause`

### Task 9 — Run `go mod tidy`

```bash
cd modelcraft-backend && go mod tidy
```

Verify `go.mod` no longer references `gorm.io/*`.

### Task 10 — Verify build and tests pass

```bash
just build
just lint
just test-unit
```

---

## Execution Order

Dependencies:
```
Task 1 (introspector: independent)
  └─→ Task 2 (reverse_engineer: depends on Task 1 interface change)

Task 3 (SqlRoleRepository: independent)
Task 4 (connection_factory: independent)

Task 5 (routes.go: depends on Tasks 2 + 3 + 4)
Task 6 (main.go: depends on Task 4)

Task 7 (delete files: safe after Tasks 5 + 6 compile cleanly)
Task 8 (PO cleanup + dead code: safe after Task 7)

Task 9 (go mod tidy: after all GORM removed)
Task 10 (verify: last)
```

Parallel-safe execution groups:
- **Round 1 (parallel):** Tasks 1, 3, 4
- **Round 2 (parallel):** Tasks 2, 5, 6 (after Round 1)
- **Round 3 (parallel):** Tasks 7, 8 (after Round 2 compiles)
- **Round 4 (sequential):** Task 9 → Task 10
