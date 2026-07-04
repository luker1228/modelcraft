# Access Control RLS Policy Management — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the access-control page's 3 tabs (roles/bundles/permissions) with a single RLS Policy V2 management UI. Backend resolvers implemented, frontend built against `rls_policy_v2.graphql`.

**Architecture:** Backend follows existing DDD layers (domain → app → interfaces/graphql). Frontend follows existing patterns (API client → hooks → components). Policy CRUD is per-model, selected via dropdown at project level.

**Tech Stack:** Go (gqlgen + sqlc), Next.js (Apollo Client + shadcn/ui)

---

## File Structure

### Backend — New Files

| File | Responsibility |
|------|---------------|
| `internal/app/rls/policy_crud_service.go` | V2 Policy CRUD app service (List/Upsert/Delete/DeleteByModel) |

### Backend — Modified Files

| File | Change |
|------|--------|
| `internal/infrastructure/persistence/rls/policy_repository.go` | Add V2 CRUD methods: ListByModel, Upsert, Delete, DeleteByModel |
| `internal/domain/rls/policy_repository.go` | Add V2 repository interface |
| `internal/interfaces/graphql/project/rls_policy_v2.resolvers.go` | Replace panic stubs with real implementations |
| `internal/interfaces/graphql/project/resolver.go` | Inject new PolicyCRUDService |

### Frontend — New Files

| File | Responsibility |
|------|---------------|
| `src/api-client/rls-policy/index.ts` | Barrel export |
| `src/api-client/rls-policy/graphql-docs.ts` | GraphQL query/mutation documents |
| `src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/RlsPolicyContent.tsx` | Main content: model selector + policy table |
| `src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/PolicyEditorDialog.tsx` | Create/edit policy dialog |
| `src/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/rls-policy/useRlsPolicyList.ts` | Query hook |
| `src/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/rls-policy/useRlsPolicyManage.ts` | Mutation hook |

### Frontend — Modified Files

| File | Change |
|------|--------|
| `src/app/org/[orgName]/project/[projectSlug]/access-control/page.tsx` | Remove tabs, render RlsPolicyContent |
| `src/app/org/[orgName]/project/[projectSlug]/access-control/_components/index.ts` | Replace exports with rls-policy only |
| `src/web/components/features/layout/AppLayout.tsx` | Simplify sidebar nav (remove role/bundle/permission sub-items) |
| `src/web/lib/route-catalog.ts` | Merge 3 route entries into 1 |

### Frontend — Deleted Paths

| Path | Content |
|------|---------|
| `access-control/_components/roles/` | RolesContent |
| `access-control/_components/bundles/` | BundlesTab |
| `access-control/_components/permissions/` | PermissionsTab, CreatePermissionSheet, ColumnPolicyEditor, RowScopeSelector |
| `access-control/_hooks/roles/` | useRoleList, useRoleEdit |
| `access-control/_hooks/bundles/` | useBundleList, useBundleManage |
| `access-control/_hooks/permissions/` | useCreatePermissionWizard, usePermissionList, usePermissionsView |
| `access-control/[roleId]/` | Role detail page |
| `access-control/bundles/` | Bundle detail page |

---

### Task 1: Backend — V2 Policy Repository Interface

**Files:**
- Create: `modelcraft-backend/internal/domain/rls/policy_repository.go`

- [ ] **Step 1: Define the V2 repository interface**

```go
// internal/domain/rls/policy_repository.go
package rls

import "context"

// PolicyRepositoryV2 V2 多策略 CRUD 接口
type PolicyRepositoryV2 interface {
	// ListByModel 查询模型的所有策略
	ListByModel(ctx context.Context, orgName, projectSlug, modelID string) ([]*Policy, error)

	// Upsert 创建或更新单条策略（按 policy_name 唯一键 upsert）
	Upsert(ctx context.Context, orgName, projectSlug string, policy *Policy) error

	// Delete 按 ID 删除单条策略
	Delete(ctx context.Context, orgName, projectSlug string, id int64) error

	// DeleteByModel 删除模型的所有策略
	DeleteByModel(ctx context.Context, orgName, projectSlug, modelID string) error
}
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-backend/internal/domain/rls/policy_repository.go
git commit -m "feat(rls): add V2 policy repository interface"
```

---

### Task 2: Backend — Extend SqlPolicyRepository with V2 CRUD

**Files:**
- Modify: `modelcraft-backend/internal/infrastructure/persistence/rls/policy_repository.go`

- [ ] **Step 1: Add ListByModel method**

Append to the existing file:

```go
// ListByModel 查询模型的所有策略（用于管理界面）
func (r *SqlPolicyRepository) ListByModel(
	ctx context.Context, orgName, projectSlug, modelID string,
) ([]*rls.Policy, error) {
	rows, err := r.q.ListPoliciesByModel(ctx, dbgen.ListPoliciesByModelParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ModelID:     modelID,
	})
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}

	policies := make([]*rls.Policy, 0, len(rows))
	for _, row := range rows {
		p := &rls.Policy{
			ID:           int64(row.ID),
			OrgName:      row.OrgName,
			ProjectSlug:  row.ProjectSlug,
			ModelID:      row.ModelID,
			PolicyName:   row.PolicyName,
			Action:       rls.Action(row.Action),
			Role:         row.Role,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}
		if row.UsingExpr != nil {
			p.UsingExpr = rls.JsonExpr(*row.UsingExpr)
		}
		if row.WithCheckExpr != nil {
			p.WithCheckExpr = rls.JsonExpr(*row.WithCheckExpr)
		}
		policies = append(policies, p)
	}
	return policies, nil
}
```

- [ ] **Step 2: Add Upsert method**

```go
// Upsert 创建或更新单条策略（按 policy_name 唯一键 upsert）
func (r *SqlPolicyRepository) Upsert(
	ctx context.Context, orgName, projectSlug string, policy *rls.Policy,
) error {
	usingExpr := json.RawMessage(policy.UsingExpr)
	withCheckExpr := json.RawMessage(policy.WithCheckExpr)

	err := r.q.UpsertPolicy(ctx, dbgen.UpsertPolicyParams{
		OrgName:       orgName,
		ProjectSlug:   projectSlug,
		ModelID:       policy.ModelID,
		PolicyName:    policy.PolicyName,
		Action:        dbgen.ModelRlsPoliciesAction(policy.Action),
		Role:          policy.Role,
		UsingExpr:     &usingExpr,
		WithCheckExpr: &withCheckExpr,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}
```

Note: add `"encoding/json"` to imports.

- [ ] **Step 3: Add Delete method**

```go
// Delete 按 ID 删除单条策略
func (r *SqlPolicyRepository) Delete(
	ctx context.Context, orgName, projectSlug string, id int64,
) error {
	err := r.q.DeletePolicy(ctx, dbgen.DeletePolicyParams{
		ID:          uint64(id),
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}
```

- [ ] **Step 4: Add DeleteByModel method**

```go
// DeleteByModel 删除模型的所有策略
func (r *SqlPolicyRepository) DeleteByModel(
	ctx context.Context, orgName, projectSlug, modelID string,
) error {
	err := r.q.DeletePoliciesByModel(ctx, dbgen.DeletePoliciesByModelParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ModelID:     modelID,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/infrastructure/persistence/rls/policy_repository.go
git commit -m "feat(rls): add V2 CRUD methods to SqlPolicyRepository"
```

---

### Task 3: Backend — Policy CRUD App Service

**Files:**
- Create: `modelcraft-backend/internal/app/rls/policy_crud_service.go`

- [ ] **Step 1: Create the CRUD app service**

```go
// internal/app/rls/policy_crud_service.go
package rls

import (
	"context"
	"modelcraft/internal/domain/rls"
)

// PolicyCRUDService V2 策略 CRUD 应用服务
type PolicyCRUDService struct {
	repo rls.PolicyRepositoryV2
}

// NewPolicyCRUDService 创建 PolicyCRUDService
func NewPolicyCRUDService(repo rls.PolicyRepositoryV2) *PolicyCRUDService {
	return &PolicyCRUDService{repo: repo}
}

// ListByModel 查询模型的所有策略
func (s *PolicyCRUDService) ListByModel(
	ctx context.Context, orgName, projectSlug, modelID string,
) ([]*rls.Policy, error) {
	return s.repo.ListByModel(ctx, orgName, projectSlug, modelID)
}

// UpsertInput upsert 输入
type UpsertInput struct {
	ModelID       string
	PolicyName    string
	Action        rls.Action
	Role          string
	UsingExpr     rls.JsonExpr
	WithCheckExpr rls.JsonExpr
}

// Upsert 创建或更新策略
func (s *PolicyCRUDService) Upsert(
	ctx context.Context, orgName, projectSlug string, input UpsertInput,
) (*rls.Policy, error) {
	policy := &rls.Policy{
		ModelID:       input.ModelID,
		PolicyName:    input.PolicyName,
		Action:        input.Action,
		Role:          input.Role,
		UsingExpr:     input.UsingExpr,
		WithCheckExpr: input.WithCheckExpr,
	}

	if err := s.repo.Upsert(ctx, orgName, projectSlug, policy); err != nil {
		return nil, err
	}

	// Re-query to get the persisted record (with ID, timestamps)
	policies, err := s.repo.ListByModel(ctx, orgName, projectSlug, input.ModelID)
	if err != nil {
		return nil, err
	}
	for _, p := range policies {
		if p.PolicyName == input.PolicyName {
			return p, nil
		}
	}
	return policy, nil
}

// Delete 删除单条策略
func (s *PolicyCRUDService) Delete(
	ctx context.Context, orgName, projectSlug string, id int64,
) error {
	return s.repo.Delete(ctx, orgName, projectSlug, id)
}

// DeleteByModel 删除模型的所有策略
func (s *PolicyCRUDService) DeleteByModel(
	ctx context.Context, orgName, projectSlug, modelID string,
) error {
	return s.repo.DeleteByModel(ctx, orgName, projectSlug, modelID)
}
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-backend/internal/app/rls/policy_crud_service.go
git commit -m "feat(rls): add V2 policy CRUD app service"
```

---

### Task 4: Backend — Implement V2 GraphQL Resolvers

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/rls_policy_v2.resolvers.go`
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/resolver.go`

- [ ] **Step 1: Add PolicyCRUDService to Resolver struct**

In `resolver.go`, add to the `Resolver` struct:

```go
// RLS Policy V2
PolicyCRUDService *appRLS.PolicyCRUDService
```

Add import: `appRLS "modelcraft/internal/app/rls"` if not already present.

- [ ] **Step 2: Implement RlsPolicies query resolver**

Replace the panic stub in `rls_policy_v2.resolvers.go`:

```go
// RlsPolicies is the resolver for the rlsPolicies field.
func (r *queryResolver) RlsPolicies(ctx context.Context, modelID string) ([]*generated.RlsPolicy, error) {
	orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	policies, err := r.PolicyCRUDService.ListByModel(ctx, orgName, projectSlug, modelID)
	if err != nil {
		return nil, err
	}

	result := make([]*generated.RlsPolicy, 0, len(policies))
	for _, p := range policies {
		result = append(result, &generated.RlsPolicy{
			ID:            fmt.Sprintf("%d", p.ID),
			PolicyName:    p.PolicyName,
			Action:        generated.RlsAction(p.Action),
			Role:          p.Role,
			UsingExpr:     stringPtr(string(p.UsingExpr)),
			WithCheckExpr: stringPtr(string(p.WithCheckExpr)),
			CreatedAt:     p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}
```

Note: add `"fmt"` and `"time"` to imports. Add helper:

```go
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
```

- [ ] **Step 3: Implement UpsertRlsPolicy mutation resolver**

```go
// UpsertRlsPolicy is the resolver for the upsertRlsPolicy field.
func (r *mutationResolver) UpsertRlsPolicy(ctx context.Context, modelID string, input generated.RlsPolicyInput) (*generated.UpsertRlsPolicyPayload, error) {
	orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var usingExpr domainRLS.JsonExpr
	if input.UsingExpr != nil {
		usingExpr = domainRLS.JsonExpr(*input.UsingExpr)
	}
	var withCheckExpr domainRLS.JsonExpr
	if input.WithCheckExpr != nil {
		withCheckExpr = domainRLS.JsonExpr(*input.WithCheckExpr)
	}

	policy, err := r.PolicyCRUDService.Upsert(ctx, orgName, projectSlug, appRLS.UpsertInput{
		ModelID:       modelID,
		PolicyName:    input.PolicyName,
		Action:        domainRLS.Action(input.Action),
		Role:          input.Role,
		UsingExpr:     usingExpr,
		WithCheckExpr: withCheckExpr,
	})
	if err != nil {
		return &generated.UpsertRlsPolicyPayload{
			Policy: nil,
			Error:  &generated.InvalidInput{Message: err.Error()},
		}, nil
	}

	return &generated.UpsertRlsPolicyPayload{
		Policy: &generated.RlsPolicy{
			ID:            fmt.Sprintf("%d", policy.ID),
			PolicyName:    policy.PolicyName,
			Action:        generated.RlsAction(policy.Action),
			Role:          policy.Role,
			UsingExpr:     stringPtr(string(policy.UsingExpr)),
			WithCheckExpr: stringPtr(string(policy.WithCheckExpr)),
			CreatedAt:     policy.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     policy.UpdatedAt.Format(time.RFC3339),
		},
		Error: nil,
	}, nil
}
```

- [ ] **Step 4: Implement DeleteRlsPolicy mutation resolver**

```go
// DeleteRlsPolicy is the resolver for the deleteRlsPolicy field.
func (r *mutationResolver) DeleteRlsPolicy(ctx context.Context, id string) (*generated.DeleteRlsPolicyPayload, error) {
	orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	policyID, parseErr := strconv.ParseInt(id, 10, 64)
	if parseErr != nil {
		return &generated.DeleteRlsPolicyPayload{
			Success: false,
			Error:   &generated.ResourceNotFound{Message: "policy not found"},
		}, nil
	}

	if err := r.PolicyCRUDService.Delete(ctx, orgName, projectSlug, policyID); err != nil {
		return &generated.DeleteRlsPolicyPayload{
			Success: false,
			Error:   &generated.ResourceNotFound{Message: err.Error()},
		}, nil
	}

	return &generated.DeleteRlsPolicyPayload{
		Success: true,
		Error:   nil,
	}, nil
}
```

Note: add `"strconv"` to imports.

- [ ] **Step 5: Implement DeleteRlsPoliciesByModel mutation resolver**

```go
// DeleteRlsPoliciesByModel is the resolver for the deleteRlsPoliciesByModel field.
func (r *mutationResolver) DeleteRlsPoliciesByModel(ctx context.Context, modelID string) (*generated.DeleteRlsPolicyPayload, error) {
	orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := r.PolicyCRUDService.DeleteByModel(ctx, orgName, projectSlug, modelID); err != nil {
		return &generated.DeleteRlsPolicyPayload{
			Success: false,
			Error:   &generated.ResourceNotFound{Message: err.Error()},
		}, nil
	}

	return &generated.DeleteRlsPolicyPayload{
		Success: true,
		Error:   nil,
	}, nil
}
```

- [ ] **Step 6: Add missing imports to rls_policy_v2.resolvers.go**

Ensure imports include:
```go
import (
	"context"
	"fmt"
	"strconv"
	"time"

	appRLS "modelcraft/internal/app/rls"
	domainRLS "modelcraft/internal/domain/rls"
	"modelcraft/internal/interfaces/graphql/project/generated"
)
```

- [ ] **Step 7: Build and verify**

```bash
cd modelcraft-backend && go build ./...
```

- [ ] **Step 8: Commit**

```bash
git add modelcraft-backend/internal/interfaces/graphql/project/
git commit -m "feat(rls): implement V2 RLS policy GraphQL resolvers"
```

---

### Task 5: Frontend — Contract Sync

**Files:**
- Pull backend contract into frontend

- [ ] **Step 1: Run contract pull**

```bash
cd modelcraft-backend && just generate-gql
```

Then use the `front-contract-pull` skill to sync to frontend.

- [ ] **Step 2: Verify generated types**

Check that `modelcraft-front/src/generated/graphql.ts` now contains:
```typescript
export type RlsPolicy = { ... }
export type RlsAction = ...
export type RlsPolicyInput = ...
export type UpsertRlsPolicyPayload = ...
export type DeleteRlsPolicyPayload = ...
```

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/contract/ modelcraft-front/src/generated/
git commit -m "chore(frontend): sync RLS policy V2 contract from backend"
```

---

### Task 6: Frontend — API Client GraphQL Documents

**Files:**
- Create: `modelcraft-front/src/api-client/rls-policy/index.ts`
- Create: `modelcraft-front/src/api-client/rls-policy/graphql-docs.ts`

- [ ] **Step 1: Create GraphQL documents**

```typescript
// src/api-client/rls-policy/graphql-docs.ts
import { gql } from '@apollo/client'

export const GET_RLS_POLICIES = gql`
  query GetRlsPolicies($modelId: ID!) {
    rlsPolicies(modelId: $modelId) {
      id
      policyName
      action
      role
      usingExpr
      withCheckExpr
      createdAt
      updatedAt
    }
  }
`

export const UPSERT_RLS_POLICY = gql`
  mutation UpsertRlsPolicy($modelId: ID!, $input: RlsPolicyInput!) {
    upsertRlsPolicy(modelId: $modelId, input: $input) {
      policy {
        id
        policyName
        action
        role
        usingExpr
        withCheckExpr
        createdAt
        updatedAt
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on ResourceNotFound {
          message
        }
      }
    }
  }
`

export const DELETE_RLS_POLICY = gql`
  mutation DeleteRlsPolicy($id: ID!) {
    deleteRlsPolicy(id: $id) {
      success
      error {
        __typename
        ... on ResourceNotFound {
          message
        }
      }
    }
  }
`
```

- [ ] **Step 2: Create barrel export**

```typescript
// src/api-client/rls-policy/index.ts
export * from './graphql-docs'
```

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/api-client/rls-policy/
git commit -m "feat(frontend): add RLS policy GraphQL documents"
```

---

### Task 7: Frontend — Hooks

**Files:**
- Create: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/rls-policy/useRlsPolicyList.ts`
- Create: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/rls-policy/useRlsPolicyManage.ts`

- [ ] **Step 1: Create useRlsPolicyList hook**

```typescript
// _hooks/rls-policy/useRlsPolicyList.ts
import { useQuery } from '@apollo/client'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import { GET_RLS_POLICIES } from '@/api-client/rls-policy'
import type { RlsPolicy } from '@/generated/graphql'

interface UseRlsPolicyListProps {
  projectSlug: string
  modelId: string | null
}

interface UseRlsPolicyListReturn {
  policies: RlsPolicy[]
  loading: boolean
  error: Error | undefined
}

export function useRlsPolicyList({ projectSlug, modelId }: UseRlsPolicyListProps): UseRlsPolicyListReturn {
  const client = useProjectScopedClient(projectSlug)

  const { data, loading, error } = useQuery(GET_RLS_POLICIES, {
    client,
    variables: { modelId },
    skip: !modelId,
  })

  const policies: RlsPolicy[] = data?.rlsPolicies ?? []

  return { policies, loading, error }
}
```

- [ ] **Step 2: Create useRlsPolicyManage hook**

```typescript
// _hooks/rls-policy/useRlsPolicyManage.ts
import { useMutation } from '@apollo/client'
import { useCallback } from 'react'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import {
  UPSERT_RLS_POLICY,
  DELETE_RLS_POLICY,
  GET_RLS_POLICIES,
} from '@/api-client/rls-policy'
import type { RlsAction } from '@/generated/graphql'

interface UseRlsPolicyManageProps {
  projectSlug: string
  modelId: string
}

interface UpsertInput {
  policyName: string
  action: RlsAction
  role: string
  usingExpr?: string
  withCheckExpr?: string
}

interface UseRlsPolicyManageReturn {
  upsertPolicy: (input: UpsertInput) => Promise<{ success: boolean; errorMessage?: string }>
  deletePolicy: (id: string) => Promise<{ success: boolean; errorMessage?: string }>
  upserting: boolean
  deleting: boolean
}

export function useRlsPolicyManage({ projectSlug, modelId }: UseRlsPolicyManageProps): UseRlsPolicyManageReturn {
  const client = useProjectScopedClient(projectSlug)

  const [upsertMutation, { loading: upserting }] = useMutation(UPSERT_RLS_POLICY, { client })
  const [deleteMutation, { loading: deleting }] = useMutation(DELETE_RLS_POLICY, { client })

  const upsertPolicy = useCallback(
    async (input: UpsertInput) => {
      const result = await upsertMutation({
        variables: {
          modelId,
          input: {
            policyName: input.policyName,
            action: input.action,
            role: input.role,
            usingExpr: input.usingExpr ?? null,
            withCheckExpr: input.withCheckExpr ?? null,
          },
        },
        refetchQueries: [{ query: GET_RLS_POLICIES, variables: { modelId } }],
      })

      const payload = result.data?.upsertRlsPolicy
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message }
      }
      return { success: true }
    },
    [upsertMutation, modelId],
  )

  const deletePolicy = useCallback(
    async (id: string) => {
      const result = await deleteMutation({
        variables: { id },
        refetchQueries: [{ query: GET_RLS_POLICIES, variables: { modelId } }],
      })

      const payload = result.data?.deleteRlsPolicy
      if (payload?.error) {
        return { success: false, errorMessage: payload.error.message }
      }
      return { success: true }
    },
    [deleteMutation, modelId],
  )

  return { upsertPolicy, deletePolicy, upserting, deleting }
}
```

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/_hooks/rls-policy/
git commit -m "feat(frontend): add RLS policy hooks"
```

---

### Task 8: Frontend — RLS Policy UI Components

**Files:**
- Create: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/RlsPolicyContent.tsx`
- Create: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/PolicyEditorDialog.tsx`

- [ ] **Step 1: Create PolicyEditorDialog**

```tsx
// _components/rls-policy/PolicyEditorDialog.tsx
'use client'

import * as React from 'react'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@web/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { Textarea } from '@web/components/ui/textarea'
import type { RlsAction } from '@/generated/graphql'

const ACTIONS: RlsAction[] = ['read', 'create', 'update', 'delete']

interface PolicyEditorDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (data: {
    policyName: string
    action: RlsAction
    role: string
    usingExpr?: string
    withCheckExpr?: string
  }) => Promise<void>
  saving: boolean
}

export function PolicyEditorDialog({
  open,
  onOpenChange,
  onSave,
  saving,
}: PolicyEditorDialogProps) {
  const [policyName, setPolicyName] = React.useState('')
  const [action, setAction] = React.useState<RlsAction>('read')
  const [role, setRole] = React.useState('')
  const [usingExpr, setUsingExpr] = React.useState('')
  const [withCheckExpr, setWithCheckExpr] = React.useState('')

  // Reset form on open
  React.useEffect(() => {
    if (open) {
      setPolicyName('')
      setAction('read')
      setRole('')
      setUsingExpr('')
      setWithCheckExpr('')
    }
  }, [open])

  const handleSave = async () => {
    if (!policyName.trim()) { toast.error('请输入策略名称'); return }
    if (!role.trim()) { toast.error('请输入角色'); return }
    await onSave({
      policyName: policyName.trim(),
      action,
      role: role.trim(),
      usingExpr: usingExpr.trim() || undefined,
      withCheckExpr: withCheckExpr.trim() || undefined,
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>添加策略</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-1.5">
            <Label>
              策略名称 <span className="text-destructive">*</span>
            </Label>
            <Input
              value={policyName}
              onChange={(e) => setPolicyName(e.target.value)}
              placeholder="例如：admin_full_access"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>
                Action <span className="text-destructive">*</span>
              </Label>
              <Select value={action} onValueChange={(v) => setAction(v as RlsAction)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {ACTIONS.map((a) => (
                    <SelectItem key={a} value={a}>{a}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>
                Role <span className="text-destructive">*</span>
              </Label>
              <Input
                value={role}
                onChange={(e) => setRole(e.target.value)}
                placeholder="例如：admin"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>Using Expr</Label>
            <Textarea
              value={usingExpr}
              onChange={(e) => setUsingExpr(e.target.value)}
              placeholder='例如：{"owner_id": {"equals": "{{user_id}}"}}'
              rows={3}
              className="font-mono text-xs"
            />
          </div>

          <div className="space-y-1.5">
            <Label>Check Expr</Label>
            <Textarea
              value={withCheckExpr}
              onChange={(e) => setWithCheckExpr(e.target.value)}
              placeholder='例如：{"owner_id": {"equals": "{{user_id}}"}}'
              rows={3}
              className="font-mono text-xs"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
          <Button onClick={handleSave} disabled={saving} className="bg-primary text-primary-foreground hover:bg-primary/90">
            {saving ? <><Loader2 className="mr-2 size-4 animate-spin" />保存中...</> : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 2: Create RlsPolicyContent (main page content)**

```tsx
// _components/rls-policy/RlsPolicyContent.tsx
'use client'

import * as React from 'react'
import { Plus, Trash2, ShieldOff, Loader2, Search } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { Skeleton } from '@web/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@web/components/ui/alert-dialog'
import { useRlsPolicyList } from '../../_hooks/rls-policy/useRlsPolicyList'
import { useRlsPolicyManage } from '../../_hooks/rls-policy/useRlsPolicyManage'
import { PolicyEditorDialog } from './PolicyEditorDialog'
import type { RlsAction } from '@/generated/graphql'

interface RlsPolicyContentProps {
  orgName: string
  projectSlug: string
}

// TODO: fetch models list from API; for now using a prop placeholder
const MODELS: { id: string; name: string }[] = []

export function RlsPolicyContent({ orgName, projectSlug }: RlsPolicyContentProps) {
  const [selectedModelId, setSelectedModelId] = React.useState<string | null>(null)
  const [editorOpen, setEditorOpen] = React.useState(false)
  const [deleteTargetId, setDeleteTargetId] = React.useState<string | null>(null)

  const { policies, loading } = useRlsPolicyList({
    projectSlug,
    modelId: selectedModelId,
  })
  const { upsertPolicy, deletePolicy, upserting, deleting } = useRlsPolicyManage({
    projectSlug,
    modelId: selectedModelId ?? '',
  })

  const handleUpsert = async (data: {
    policyName: string
    action: RlsAction
    role: string
    usingExpr?: string
    withCheckExpr?: string
  }) => {
    const result = await upsertPolicy(data)
    if (result.success) {
      toast.success('策略已保存')
      setEditorOpen(false)
    } else {
      toast.error(result.errorMessage ?? '保存失败')
    }
  }

  const handleDelete = async () => {
    if (!deleteTargetId) return
    const result = await deletePolicy(deleteTargetId)
    if (result.success) {
      toast.success('策略已删除')
      setDeleteTargetId(null)
    } else {
      toast.error(result.errorMessage ?? '删除失败')
    }
  }

  const actionLabel = (a: string) => {
    const map: Record<string, string> = {
      read: 'read',
      create: 'create',
      update: 'update',
      delete: 'delete',
    }
    return map[a] ?? a
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Select
            value={selectedModelId ?? ''}
            onValueChange={(v) => setSelectedModelId(v || null)}
          >
            <SelectTrigger className="w-[220px]">
              <SelectValue placeholder="选择模型..." />
            </SelectTrigger>
            <SelectContent>
              {MODELS.map((m) => (
                <SelectItem key={m.id} value={m.id}>{m.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <Button
          onClick={() => setEditorOpen(true)}
          size="sm"
          disabled={!selectedModelId}
          className="bg-primary text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="size-4" strokeWidth={1.5} />
          添加策略
        </Button>
      </div>

      {!selectedModelId && (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <ShieldOff className="mb-3 size-9 text-muted-foreground/30" strokeWidth={1} />
          <p className="text-sm font-semibold text-foreground">请选择一个模型</p>
          <p className="mt-1 text-xs text-muted-foreground">
            选择模型后查看和管理其 RLS 策略
          </p>
        </div>
      )}

      {selectedModelId && loading && (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      )}

      {selectedModelId && !loading && (
        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <Table>
            <TableHeader>
              <TableRow className="border-b-2 border-border bg-card hover:bg-card">
                <TableHead className="h-10 w-[100px] text-[11px] font-medium uppercase tracking-wider text-foreground">Action</TableHead>
                <TableHead className="h-10 w-[140px] text-[11px] font-medium uppercase tracking-wider text-foreground">Role</TableHead>
                <TableHead className="h-10 text-[11px] font-medium uppercase tracking-wider text-foreground">Using Expr</TableHead>
                <TableHead className="h-10 text-[11px] font-medium uppercase tracking-wider text-foreground">Check Expr</TableHead>
                <TableHead className="h-10 w-[80px] text-right text-[11px] font-medium uppercase tracking-wider text-foreground">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {policies.map((policy) => (
                <TableRow key={policy.id} className="group border-b border-border last:border-0 hover:bg-foreground/[0.015]">
                  <TableCell className="h-12 text-[13px]">
                    <span className="inline-flex items-center rounded-md bg-secondary px-2 py-0.5 text-xs font-medium text-secondary-foreground">
                      {actionLabel(policy.action)}
                    </span>
                  </TableCell>
                  <TableCell className="h-12 text-[13px] font-medium text-foreground">
                    {policy.role || <span className="text-muted-foreground/40">默认</span>}
                  </TableCell>
                  <TableCell className="h-12">
                    <code className="text-[11px] text-muted-foreground line-clamp-1">
                      {policy.usingExpr || '—'}
                    </code>
                  </TableCell>
                  <TableCell className="h-12">
                    <code className="text-[11px] text-muted-foreground line-clamp-1">
                      {policy.withCheckExpr || '—'}
                    </code>
                  </TableCell>
                  <TableCell className="h-12 text-right">
                    <div className="flex items-center justify-end opacity-0 transition-opacity group-hover:opacity-100">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 gap-1.5 text-xs text-muted-foreground hover:text-destructive"
                        onClick={() => setDeleteTargetId(policy.id)}
                      >
                        <Trash2 className="size-3.5" strokeWidth={1.5} />
                        删除
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {policies.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-24 text-center">
                    <div className="flex flex-col items-center justify-center py-8">
                      <ShieldOff className="mb-2 size-7 text-muted-foreground/25" strokeWidth={1} />
                      <p className="text-sm font-semibold text-foreground">暂无策略</p>
                      <p className="mt-1 text-xs text-muted-foreground">
                        点击「添加策略」为该模型创建 RLS 策略
                      </p>
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      )}

      <PolicyEditorDialog
        open={editorOpen}
        onOpenChange={setEditorOpen}
        onSave={handleUpsert}
        saving={upserting}
      />

      <AlertDialog open={!!deleteTargetId} onOpenChange={(open) => { if (!open) setDeleteTargetId(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除策略</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除此策略吗？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={handleDelete}
              disabled={deleting}
            >
              {deleting ? <><Loader2 className="mr-2 size-4 animate-spin" />删除中...</> : '确认删除'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
```

Note: The `MODELS` constant is a placeholder. In Task 10, you will wire up the actual model list query from the existing API.

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/_components/rls-policy/
git commit -m "feat(frontend): add RLS policy UI components"
```

---

### Task 9: Frontend — Wire Up Page, Sidebar, Route Catalog

**Files:**
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/page.tsx`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/index.ts`
- Modify: `modelcraft-front/src/web/components/features/layout/AppLayout.tsx`
- Modify: `modelcraft-front/src/web/lib/route-catalog.ts`

- [ ] **Step 1: Rewrite page.tsx**

```tsx
// page.tsx
'use client'

import { useParams } from 'next/navigation'

import { PageHeader, PageLayout } from '@web/components/features/layout'
import { RlsPolicyContent } from './_components'

export default function AccessControlPage() {
  const params = useParams()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  return (
    <PageLayout maxWidth="7xl">
      <PageHeader title="RLS 策略" />

      <div className="mt-6">
        <RlsPolicyContent orgName={orgName} projectSlug={projectSlug} />
      </div>
    </PageLayout>
  )
}
```

- [ ] **Step 2: Update _components/index.ts**

```typescript
export { RlsPolicyContent } from './rls-policy/RlsPolicyContent'
export { PolicyEditorDialog } from './rls-policy/PolicyEditorDialog'
```

- [ ] **Step 3: Update AppLayout.tsx sidebar nav**

Replace lines 189-193 (the 访问控制 block with children) with:

```typescript
{ label: '访问控制', icon: '/icons/icon-shield.svg', href: `/org/${orgName}/project/${projectSlug}/access-control` },
```

- [ ] **Step 4: Update route-catalog.ts**

Replace the 3 entries (lines 102-138) with a single entry:

```typescript
{
  routeTemplate: '/org/:orgName/project/:projectSlug/access-control',
  title: 'RLS 策略管理',
  description: '管理项目内各模型的 RLS 行级安全策略，按 action + role 匹配',
  keywords: ['RLS', '策略', '行级安全', '权限', 'policy', 'action', 'role'],
  examples: [
    '帮我查看 User 模型的 RLS 策略',
    '创建一个 admin 角色的 read 策略',
    '删除某条 RLS 策略',
    'RLS 策略管理在哪',
  ],
  requiresProject: true,
},
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-front/src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/page.tsx \
        modelcraft-front/src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/_components/index.ts \
        modelcraft-front/src/web/components/features/layout/AppLayout.tsx \
        modelcraft-front/src/web/lib/route-catalog.ts
git commit -m "feat(frontend): wire up RLS policy page, simplify sidebar and route catalog"
```

---

### Task 10: Frontend — Wire Up Model List (Database + Model Selectors)

**Files:**
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/RlsPolicyContent.tsx`

`GET_MODELS` requires `databaseName` as input, so we need a cascading selector: database → model.

- [ ] **Step 1: Replace MODELS placeholder with dynamic database+model selectors**

Add imports:
```typescript
import { useQuery } from '@apollo/client'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import { GET_MODELS } from '@/api-client/model'
```

In the component body, add after `const { policies, loading } = ...`:
```typescript
const client = useProjectScopedClient(projectSlug)

// Step 1: derive unique database names from project (models carry databaseName)
// Use a single models query scoped to the project's default database, or
// iterate over all databases. For simplicity, fetch models per selected DB.
const [dbName, setDbName] = React.useState<string | null>(null)

const { data: modelsData, loading: modelsLoading } = useQuery(GET_MODELS, {
  client,
  variables: { input: { databaseName: dbName ?? '' } },
  skip: !dbName,
})

const models = (modelsData?.models?.items ?? []).map(
  (m: { id: string; name: string }) => ({ id: m.id, name: m.name }),
)
```

Replace the single model Select with a database+model cascade:
```tsx
<div className="flex items-center gap-3">
  {/* Database selector — hardcoded for now; fetch from project config */}
  <Input
    placeholder="数据库名"
    value={dbName ?? ''}
    onChange={(e) => { setDbName(e.target.value); setSelectedModelId(null) }}
    className="w-[180px]"
  />
  <Select
    value={selectedModelId ?? ''}
    onValueChange={(v) => setSelectedModelId(v || null)}
    disabled={!dbName || modelsLoading}
  >
    <SelectTrigger className="w-[220px]">
      <SelectValue placeholder={
        modelsLoading ? '加载中...' : dbName ? '选择模型...' : '先输入数据库名'
      } />
    </SelectTrigger>
    <SelectContent>
      {models.map((m) => (
        <SelectItem key={m.id} value={m.id}>{m.name}</SelectItem>
      ))}
    </SelectContent>
  </Select>
</div>
```

Note: If a project-level database list endpoint exists, replace the text-input database selector with a proper Select dropdown. For now, a text input for database name is the simplest approach that works with `GET_MODELS`.

- [ ] **Step 2: Commit**

```bash
git add modelcraft-front/src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/_components/rls-policy/RlsPolicyContent.tsx
git commit -m "feat(frontend): wire model list into RLS policy content"
```

---

### Task 11: Frontend — Delete Old Code

**Files:**
- Delete all old access-control sub-components, hooks, and pages

- [ ] **Step 1: Delete old directories**

```bash
cd modelcraft-front

# Remove old components
rm -rf src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/_components/roles
rm -rf src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/_components/bundles
rm -rf src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/_components/permissions

# Remove old hooks
rm -rf src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/_hooks/roles
rm -rf src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/_hooks/bundles
rm -rf src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/_hooks/permissions

# Remove old sub-pages
rm -rf src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/\[roleId\]
rm -rf src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/bundles
```

- [ ] **Step 2: Check for any remaining imports referencing deleted modules**

```bash
grep -r "RolesContent\|BundlesTab\|PermissionsTab\|CreatePermissionSheet\|ColumnPolicyEditor\|RowScopeSelector" \
  modelcraft-front/src/ --include="*.tsx" --include="*.ts" -l
```

If any files are found, remove those import references.

- [ ] **Step 3: Run lint check**

```bash
cd modelcraft-front && npx eslint src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/ --ext .ts,.tsx
```

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/app/org/\[orgName\]/project/\[projectSlug\]/access-control/
git commit -m "refactor(frontend): remove old roles/bundles/permissions from access-control"
```

---

### Task 12: Backend — Wire Dependency Injection

**Files:**
- Find and modify the DI/wire file that constructs the GraphQL Resolver

- [ ] **Step 1: Find the DI construction site**

```bash
grep -rn "projectgraphql.Resolver\|projectgraphql.NewResolver\|RLSPolicyAppService" \
  modelcraft-backend/cmd/ modelcraft-backend/internal/interfaces/http/ \
  --include="*.go" -l
```

- [ ] **Step 2: Add PolicyCRUDService injection**

At the construction site, add:

```go
policyCRUDService := appRLS.NewPolicyCRUDService(sqlPolicyRepo)
```

And pass it to the Resolver:
```go
PolicyCRUDService: policyCRUDService,
```

- [ ] **Step 3: Build and verify**

```bash
cd modelcraft-backend && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/cmd/ modelcraft-backend/internal/interfaces/http/
git commit -m "feat(rls): wire PolicyCRUDService into GraphQL resolver"
```

---

## Implementation Order

1. **Task 1** — V2 repository interface (domain)
2. **Task 2** — Extend SqlPolicyRepository (infrastructure)
3. **Task 3** — Policy CRUD app service
4. **Task 4** — GraphQL resolvers
5. **Task 12** — DI wiring (verify backend works)
6. **Task 5** — Frontend contract sync
7. **Task 6** — API client GraphQL documents
8. **Task 7** — Hooks
9. **Task 8** — UI components
10. **Task 9** — Wire up page + sidebar + route catalog
11. **Task 10** — Wire model list
12. **Task 11** — Delete old code
