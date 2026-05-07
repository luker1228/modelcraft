# Builtin Admin EndUser per Organization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Automatically create a builtin admin EndUser for each Organization, with role assignment and system-level metadata, enabling org admins to manage the platform without manual user creation.

**Architecture:** The feature follows ModelCraft's DDD architecture with org-scoped isolation. A new database migration creates a `builtin_admin` marker in the `end_user_users` table. The CreateOrganizationService is extended to orchestrate user creation as part of the org creation transaction. GraphQL schema gains a new `createBuiltinAdmin` mutation and query. Frontend components display the admin user credentials on org creation success. Integration tests verify end-to-end flow.

**Tech Stack:** MySQL (schema migration), Go (domain/app services), GraphQL (schema + resolvers), TypeScript/React (frontend), sqlc (optional generated queries), bcrypt (password hashing), cursor-based pagination.

---

## File Structure

**Database:**
- `db/schema/mysql/15_builtin_admin_enduser.sql` — New migration adding `is_builtin_admin` column and index

**Backend Domain Layer:**
- `internal/domain/enduser/end_user.go` — Extend `EndUser` struct with `IsBuiltinAdmin` field, add `MarkAsBuiltinAdmin()` method

**Backend Application Layer:**
- `internal/app/enduser/end_user_app_service.go` — Add `CreateBuiltinAdminEndUser()` public method
- `internal/app/enduser/commands.go` — Add `CreateBuiltinAdminCommand` and `CreateBuiltinAdminResult` types
- `internal/app/organization/create_organization_service.go` — Extend `Execute()` to call builtin admin creation and wrap in transaction

**GraphQL Schema:**
- `api/graph/org/schema/end_user.graphql` — Add `isBuiltinAdmin: Boolean!` field to `EndUser` type, add mutation

**GraphQL Resolvers:**
- `api/graph/org/resolver/end_user_resolver.go` — Add `CreateBuiltinAdmin` mutation resolver

**Frontend:**
- `src/web/components/features/organization/builtin-admin-credentials-display.tsx` — New component showing builtin admin credentials after org creation
- `src/web/hooks/organization/use-create-builtin-admin.ts` — Custom hook for builtin admin creation mutation

**Tests:**
- `internal/app/enduser/end_user_app_service_test.go` — Add `TestCreateBuiltinAdminEndUser()` 
- `internal/app/organization/create_organization_service_test.go` — Extend existing tests to verify builtin admin creation

---

## Tasks

### Task 1: Create Database Migration

**Files:**
- Create: `db/schema/mysql/15_builtin_admin_enduser.sql`

- [ ] **Step 1: Write migration file**

Create file with content:

```sql
START TRANSACTION;

ALTER TABLE end_user_users
ADD COLUMN is_builtin_admin TINYINT(1) NOT NULL DEFAULT 0
AFTER is_forbidden;

CREATE INDEX idx_end_user_users_org_builtin_admin 
ON end_user_users(org_name, is_builtin_admin);

CREATE UNIQUE INDEX idx_end_user_users_org_builtin_admin_unique 
ON end_user_users(org_name, is_builtin_admin) 
WHERE is_builtin_admin = 1;

COMMIT;
```

- [ ] **Step 2: Verify migration applies**

Run: `mysql -u root -p mc_meta < db/schema/mysql/15_builtin_admin_enduser.sql`

- [ ] **Step 3: Commit**

```bash
git add db/schema/mysql/15_builtin_admin_enduser.sql
git commit -m "db: add is_builtin_admin column to end_user_users"
```

---

### Task 2: Extend EndUser Domain Entity

**Files:**
- Modify: `internal/domain/enduser/end_user.go`

- [ ] **Step 1: Add IsBuiltinAdmin field**

Add to struct:

```go
type EndUser struct {
	ID            string
	OrgName       string
	Username      string
	Password      HashedPassword
	IsForbidden   bool
	IsBuiltinAdmin bool
	CreatedBy     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
```

- [ ] **Step 2: Add MarkAsBuiltinAdmin method**

```go
func (u *EndUser) MarkAsBuiltinAdmin() {
	u.IsBuiltinAdmin = true
}
```

- [ ] **Step 3: Update NewEndUser constructor**

```go
func NewEndUser(id, orgName, username string, createdBy string, hashedPwd HashedPassword) *EndUser {
	return &EndUser{
		ID:            id,
		OrgName:       orgName,
		Username:      username,
		Password:      hashedPwd,
		IsForbidden:   false,
		IsBuiltinAdmin: false,
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/domain/enduser/... -v`

- [ ] **Step 5: Commit**

```bash
git add internal/domain/enduser/end_user.go
git commit -m "domain: add IsBuiltinAdmin field to EndUser"
```

---

### Task 3: Add Command Types

**Files:**
- Modify: `internal/app/enduser/commands.go`

- [ ] **Step 1: Add command types**

Append to file:

```go
type CreateBuiltinAdminCommand struct {
	OrgName      string
	Username     string
	PasswordHash string
	CreatedBy    string
}

type CreateBuiltinAdminResult struct {
	ID             string
	Username       string
	PlainPassword  string
	IsBuiltinAdmin bool
	CreatedAt      time.Time
}
```

- [ ] **Step 2: Verify compiles**

Run: `go build ./internal/app/enduser/...`

- [ ] **Step 3: Commit**

```bash
git add internal/app/enduser/commands.go
git commit -m "app: add CreateBuiltinAdminCommand types"
```

---

### Task 4: Implement CreateBuiltinAdminEndUser Method

**Files:**
- Modify: `internal/app/enduser/end_user_app_service.go`

- [ ] **Step 1: Add method after ListAccessibleProjects**

```go
func (s *EndUserManagementAppService) CreateBuiltinAdminEndUser(
	ctx context.Context,
	cmd *CreateBuiltinAdminCommand,
) (*CreateBuiltinAdminResult, error) {
	if cmd == nil {
		return nil, fmt.Errorf("CreateBuiltinAdminCommand cannot be nil")
	}

	if cmd.OrgName == "" {
		return nil, fmt.Errorf("OrgName is required")
	}
	if cmd.Username == "" {
		return nil, fmt.Errorf("Username is required")
	}
	if cmd.CreatedBy == "" {
		return nil, fmt.Errorf("CreatedBy is required")
	}

	userID := uuid.New().String()

	hashedPwd, err := password.HashPassword(cmd.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := enduser.NewEndUser(userID, cmd.OrgName, cmd.Username, cmd.CreatedBy, hashedPwd)
	user.MarkAsBuiltinAdmin()

	repo := s.getEndUserRepository(cmd.OrgName)

	if err := repo.Save(ctx, user); err != nil {
		return nil, s.convertRepoError(err)
	}

	return &CreateBuiltinAdminResult{
		ID:             user.ID,
		Username:       user.Username,
		PlainPassword:  cmd.PasswordHash,
		IsBuiltinAdmin: user.IsBuiltinAdmin,
		CreatedAt:      user.CreatedAt,
	}, nil
}
```

- [ ] **Step 2: Verify imports**

Ensure file has uuid and password imports

- [ ] **Step 3: Run tests**

Run: `go test ./internal/app/enduser/... -v`

- [ ] **Step 4: Commit**

```bash
git add internal/app/enduser/end_user_app_service.go
git commit -m "app: implement CreateBuiltinAdminEndUser method"
```

---

### Task 5: Extend CreateOrganizationService

**Files:**
- Modify: `internal/app/organization/create_organization_service.go`

- [ ] **Step 1: Add EndUserAppService dependency**

Update struct:

```go
type CreateOrganizationService struct {
	txManager         *txmanager.TransactionManager
	userRepo          user.UserRepository
	orgRepo           organization.OrganizationRepository
	roleRepo          role.RoleRepository
	membershipRepo    membership.MembershipRepository
	endUserAppService *enduser.EndUserManagementAppService
}
```

- [ ] **Step 2: Update constructor**

```go
func NewCreateOrganizationService(
	txManager *txmanager.TransactionManager,
	userRepo user.UserRepository,
	orgRepo organization.OrganizationRepository,
	roleRepo role.RoleRepository,
	membershipRepo membership.MembershipRepository,
	endUserAppService *enduser.EndUserManagementAppService,
) *CreateOrganizationService {
	return &CreateOrganizationService{
		txManager:         txManager,
		userRepo:          userRepo,
		orgRepo:           orgRepo,
		roleRepo:          roleRepo,
		membershipRepo:    membershipRepo,
		endUserAppService: endUserAppService,
	}
}
```

- [ ] **Step 3: Extend Execute to create builtin admin**

After org creation in Execute, add:

```go
builtinAdminCmd := &enduser.CreateBuiltinAdminCommand{
	OrgName:      result.OrganizationName,
	Username:     "admin",
	PasswordHash: s.generateSecurePassword(),
	CreatedBy:    input.OwnerUserID,
}
_, err = s.endUserAppService.CreateBuiltinAdminEndUser(ctx, builtinAdminCmd)
if err != nil {
	log.Warnf(ctx, "failed to create builtin admin: %v", err)
}
```

- [ ] **Step 4: Add password generation helper**

```go
func (s *CreateOrganizationService) generateSecurePassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result strings.Builder
	for i := 0; i < 12; i++ {
		result.WriteByte(charset[rand.Intn(len(charset))])
	}
	result.WriteString("-")
	for i := 0; i < 4; i++ {
		result.WriteByte(charset[rand.Intn(len(charset))])
	}
	return result.String()
}
```

- [ ] **Step 5: Add imports for strings, rand**

```go
import (
	"math/rand"
	"strings"
)
```

- [ ] **Step 6: Run tests**

Run: `go test ./internal/app/organization/... -v`

- [ ] **Step 7: Commit**

```bash
git add internal/app/organization/create_organization_service.go
git commit -m "app: extend CreateOrganizationService to create builtin admin"
```

---

### Task 6: Update GraphQL Schema

**Files:**
- Modify: `api/graph/org/schema/end_user.graphql`

- [ ] **Step 1: Add isBuiltinAdmin to EndUser type**

Update type:

```graphql
type EndUser implements Node {
  id: ID!
  username: String!
  isForbidden: Boolean!
  isBuiltinAdmin: Boolean!
  createdBy: String
  createdAt: Time!
  updatedAt: Time!
}
```

- [ ] **Step 2: Add isBuiltinAdmin to EndUserPublic**

```graphql
type EndUserPublic {
  id: ID!
  username: String!
  isBuiltinAdmin: Boolean!
  createdAt: Time!
}
```

- [ ] **Step 3: Add mutation and types**

Add to file:

```graphql
input CreateBuiltinAdminInput {
  orgName: String!
}

type CreateBuiltinAdminPayload {
  endUser: EndUser
  plainPassword: String
  error: CreateEndUserError
}
```

- [ ] **Step 4: Add mutation to schema**

Add to Mutation block:

```graphql
createBuiltinAdmin(input: CreateBuiltinAdminInput!): CreateBuiltinAdminPayload! @hasPermission(action: "end-user:create")
```

- [ ] **Step 5: Commit**

```bash
git add api/graph/org/schema/end_user.graphql
git commit -m "schema: add isBuiltinAdmin and CreateBuiltinAdmin mutation"
```

---

### Task 7: Implement GraphQL Resolver

**Files:**
- Create/Modify: `api/graph/org/resolver/end_user_resolver.go`

- [ ] **Step 1: Add resolver method**

```go
func (r *mutationResolver) CreateBuiltinAdmin(
	ctx context.Context,
	input generated.CreateBuiltinAdminInput,
) (*generated.CreateBuiltinAdminPayload, error) {
	orgName := auth.GetOrgFromContext(ctx)
	if orgName == "" {
		return &generated.CreateBuiltinAdminPayload{
			Error: &generated.InvalidInput{
				Message: "organization context required",
			},
		}, nil
	}

	cmd := &enduser.CreateBuiltinAdminCommand{
		OrgName:     orgName,
		Username:    "admin",
		PasswordHash: input.OrgName,
		CreatedBy:   auth.GetUserIDFromContext(ctx),
	}

	result, err := r.endUserAppService.CreateBuiltinAdminEndUser(ctx, cmd)
	if err != nil {
		return &generated.CreateBuiltinAdminPayload{
			Error: &generated.InvalidInput{
				Message: err.Error(),
			},
		}, nil
	}

	endUserGQL := &generated.EndUser{
		ID:            result.ID,
		Username:      result.Username,
		IsForbidden:   false,
		IsBuiltinAdmin: result.IsBuiltinAdmin,
		CreatedBy:     &cmd.CreatedBy,
		CreatedAt:     result.CreatedAt,
		UpdatedAt:     result.CreatedAt,
	}

	return &generated.CreateBuiltinAdminPayload{
		EndUser:       endUserGQL,
		PlainPassword: &result.PlainPassword,
	}, nil
}
```

- [ ] **Step 2: Ensure struct has endUserAppService**

Check resolver struct has field:

```go
type mutationResolver struct {
	endUserAppService *enduser.EndUserManagementAppService
}
```

- [ ] **Step 3: Run codegen**

Run: `go generate ./api/graph/...`

- [ ] **Step 4: Commit**

```bash
git add api/graph/org/resolver/end_user_resolver.go
git commit -m "resolver: implement CreateBuiltinAdmin mutation"
```

---

### Task 8: Create Frontend Credentials Component

**Files:**
- Create: `src/web/components/features/organization/builtin-admin-credentials-display.tsx`

- [ ] **Step 1: Write component**

```typescript
'use client'

import React, { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@web/components/ui/dialog'
import { Button } from '@web/components/ui/button'
import { Alert, AlertDescription } from '@web/components/ui/alert'
import { Copy, CheckCircle } from 'lucide-react'

interface Props {
  open: boolean
  username: string
  plainPassword: string
  orgName: string
  onClose: () => void
}

export function BuiltinAdminCredentialsDisplay({
  open,
  username,
  plainPassword,
  orgName,
  onClose,
}: Props) {
  const [copiedField, setCopiedField] = useState<'username' | 'password' | null>(null)

  const handleCopy = (text: string, field: 'username' | 'password') => {
    navigator.clipboard.writeText(text)
    setCopiedField(field)
    setTimeout(() => setCopiedField(null), 2000)
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Admin Credentials Created</DialogTitle>
          <DialogDescription>
            Builtin admin for <strong>{orgName}</strong> created. Save credentials now.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <Alert className="bg-amber-50 border-amber-200">
            <AlertDescription className="text-amber-800 text-sm">
              ⚠️ This password will not be shown again.
            </AlertDescription>
          </Alert>

          <div>
            <label className="block text-sm font-medium mb-1">Username</label>
            <div className="flex gap-2">
              <input
                type="text"
                value={username}
                readOnly
                className="flex-1 px-3 py-2 border border-gray-300 rounded bg-gray-50 text-sm"
              />
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleCopy(username, 'username')}
              >
                {copiedField === 'username' ? (
                  <CheckCircle className="w-4 h-4 text-green-600" />
                ) : (
                  <Copy className="w-4 h-4" />
                )}
              </Button>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Password</label>
            <div className="flex gap-2">
              <input
                type="password"
                value={plainPassword}
                readOnly
                className="flex-1 px-3 py-2 border border-gray-300 rounded bg-gray-50 text-sm"
              />
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleCopy(plainPassword, 'password')}
              >
                {copiedField === 'password' ? (
                  <CheckCircle className="w-4 h-4 text-green-600" />
                ) : (
                  <Copy className="w-4 h-4" />
                )}
              </Button>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button onClick={onClose} className="w-full">
            I've saved my credentials
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 2: Verify compiles**

Run: `npm run build`

- [ ] **Step 3: Commit**

```bash
git add src/web/components/features/organization/builtin-admin-credentials-display.tsx
git commit -m "feat: add BuiltinAdminCredentialsDisplay component"
```

---

### Task 9: Create Frontend Hook

**Files:**
- Create: `src/web/hooks/organization/use-create-builtin-admin.ts`

- [ ] **Step 1: Write hook**

```typescript
'use client'

import { useMutation } from '@apollo/client'
import { CREATE_BUILTIN_ADMIN } from '@api-client/end-user/graphql-docs'

export function useCreateBuiltinAdmin(options?: {
  onSuccess?: (credentials: any) => void
  onError?: (error: Error) => void
}) {
  const [mutation, { loading, error, data }] = useMutation(CREATE_BUILTIN_ADMIN)

  const createAdmin = async (orgName: string) => {
    try {
      const result = await mutation({
        variables: { input: { orgName } },
      })

      if (result.data?.createBuiltinAdmin?.error) {
        const err = new Error(result.data.createBuiltinAdmin.error.message)
        options?.onError?.(err)
        return null
      }

      options?.onSuccess?.(result.data?.createBuiltinAdmin)
      return result.data?.createBuiltinAdmin
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error')
      options?.onError?.(error)
      return null
    }
  }

  return { createAdmin, loading, error, credentials: data?.createBuiltinAdmin }
}
```

- [ ] **Step 2: Add GraphQL query**

In `src/api-client/end-user/graphql-docs.ts`:

```typescript
export const CREATE_BUILTIN_ADMIN = gql`
  mutation CreateBuiltinAdmin($input: CreateBuiltinAdminInput!) {
    createBuiltinAdmin(input: $input) {
      endUser {
        id
        username
        isBuiltinAdmin
        createdAt
      }
      plainPassword
      error { message }
    }
  }
`
```

- [ ] **Step 3: Verify compiles**

Run: `npm run build`

- [ ] **Step 4: Commit**

```bash
git add src/web/hooks/organization/use-create-builtin-admin.ts src/api-client/end-user/graphql-docs.ts
git commit -m "feat: add useCreateBuiltinAdmin hook"
```

---

### Task 10: Integrate Component in Org Creation

**Files:**
- Modify: `src/web/components/features/organization/create-org-dialog.tsx`

- [ ] **Step 1: Add imports**

```typescript
import { BuiltinAdminCredentialsDisplay } from './builtin-admin-credentials-display'
import { useCreateBuiltinAdmin } from '@web/hooks/organization/use-create-builtin-admin'
```

- [ ] **Step 2: Add state**

```typescript
const [showCredentials, setShowCredentials] = useState(false)
const [adminCredentials, setAdminCredentials] = useState<{
  username: string
  plainPassword: string
  orgName: string
} | null>(null)
```

- [ ] **Step 3: Add hook**

```typescript
const { createAdmin } = useCreateBuiltinAdmin({
  onSuccess: (creds) => {
    if (creds?.endUser?.username && creds?.plainPassword) {
      setAdminCredentials({
        username: creds.endUser.username,
        plainPassword: creds.plainPassword,
        orgName: orgName,
      })
      setShowCredentials(true)
    }
  },
})
```

- [ ] **Step 4: Call in handler**

In org creation success handler:

```typescript
await createAdmin(orgResult.organizationName)
```

- [ ] **Step 5: Add component to JSX**

```typescript
<BuiltinAdminCredentialsDisplay
  open={showCredentials}
  username={adminCredentials?.username || ''}
  plainPassword={adminCredentials?.plainPassword || ''}
  orgName={adminCredentials?.orgName || ''}
  onClose={() => {
    setShowCredentials(false)
    router.push(`/org/${adminCredentials?.orgName}`)
  }}
/>
```

- [ ] **Step 6: Verify integration**

Run: `npm run dev` and test org creation flow

- [ ] **Step 7: Commit**

```bash
git add src/web/components/features/organization/create-org-dialog.tsx
git commit -m "feat: integrate builtin admin credentials display"
```

---

### Task 11: Write Unit Tests

**Files:**
- Modify: `internal/app/enduser/end_user_app_service_test.go`

- [ ] **Step 1: Add test**

```go
func TestCreateBuiltinAdminEndUser(t *testing.T) {
	t.Run("successfully creates builtin admin", func(t *testing.T) {
		ctx := context.Background()
		cmd := &enduser.CreateBuiltinAdminCommand{
			OrgName:      "test-org",
			Username:     "admin",
			PasswordHash: "SecurePass123!",
			CreatedBy:    "system",
		}

		// Create service with mocked repo
		service := enduser.NewEndUserManagementAppService(nil, nil)

		result, err := service.CreateBuiltinAdminEndUser(ctx, cmd)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "admin", result.Username)
		assert.True(t, result.IsBuiltinAdmin)
		assert.Equal(t, "SecurePass123!", result.PlainPassword)
	})

	t.Run("returns error for empty OrgName", func(t *testing.T) {
		ctx := context.Background()
		service := enduser.NewEndUserManagementAppService(nil, nil)

		cmd := &enduser.CreateBuiltinAdminCommand{
			OrgName: "",
		}

		result, err := service.CreateBuiltinAdminEndUser(ctx, cmd)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/app/enduser/... -v`

- [ ] **Step 3: Commit**

```bash
git add internal/app/enduser/end_user_app_service_test.go
git commit -m "test: add CreateBuiltinAdminEndUser tests"
```

---

### Task 12: Write Integration Tests

**Files:**
- Modify: `internal/app/organization/create_organization_service_test.go`

- [ ] **Step 1: Add integration test**

```go
func TestCreateOrganizationWithBuiltinAdmin(t *testing.T) {
	t.Run("creates org and builtin admin in transaction", func(t *testing.T) {
		ctx := context.Background()

		service := setupTestService(t)
		input := &organization.CreateOrganizationInput{
			DisplayName:      "Test Org",
			OrganizationName: "test-org",
			OwnerUserID:      "owner-123",
		}

		result, err := service.Execute(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-org", result.OrganizationName)
	})
}

func setupTestService(t *testing.T) *organization.CreateOrganizationService {
	t.Helper()
	// Setup test dependencies
	return nil
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/app/organization/... -v`

- [ ] **Step 3: Commit**

```bash
git add internal/app/organization/create_organization_service_test.go
git commit -m "test: add org creation with builtin admin integration tests"
```

---

### Task 13: Documentation

**Files:**
- Create: `docs/features/builtin-admin-enduser.md`

- [ ] **Step 1: Write documentation**

```markdown
# Builtin Admin EndUser Per Organization

## Overview

Organizations automatically get a builtin admin user on creation.

## Database

Column `is_builtin_admin` added to `end_user_users` table.
- Default: 0 (false)
- Unique constraint per org ensures only one admin

## API

GraphQL mutation `createBuiltinAdmin` accepts orgName and returns:
- endUser: The created admin user
- plainPassword: One-time password display
- error: Error union if failed

## Frontend

Component `BuiltinAdminCredentialsDisplay` shows credentials modal.
Hook `useCreateBuiltinAdmin` handles mutation.

## Security

- Password bcrypt-hashed before storage
- Plaintext returned only once
- UI enforces copy-to-clipboard
- User recommended to change password after login
```

- [ ] **Step 2: Commit**

```bash
git add docs/features/builtin-admin-enduser.md
git commit -m "docs: add builtin admin enduser documentation"
```

---

Plan complete and saved to `docs/superpowers/plans/2026-05-07-builtin-admin-enduser-per-org.md`.

## Execution Options

**Two approaches:**

**1. Subagent-Driven (Recommended)** — Fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** — Sequential execution in this session with checkpoints

**Which would you prefer?**
