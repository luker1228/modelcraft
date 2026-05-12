# ✅ ModelCraft Front Codebase - Complete Implementation Reference

**Status:** Ready for precise implementation  
**Date Prepared:** 2026-05-12  
**Framework:** Next.js 14.2 + React 18 + TypeScript 5 + Tailwind CSS

---

## 📋 ANSWERS TO ALL 8 QUESTIONS

### 1. ORG-LEVEL LAYOUT FILE

**Exact Path:** `src/app/org/[orgName]/layout.tsx`  
**Full URL:** `/data/home/lukemxjia/modelcraft/modelcraft-front/src/app/org/[orgName]/layout.tsx`

**Key Characteristics:**
- Client component (`'use client'`)
- Verifies org access via memberships
- Guards unauthenticated access
- Shows loading spinner during verification

**Content (excerpt):**
```typescript
"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useRequireAuth } from "@web/hooks/auth/use-auth";
import { useOrganizationStore } from "@shared/stores/organization";
import { TENANT_LOGIN_PATH } from "@shared/constants/routes";

export default function OrgLayout({ children }: { children: React.ReactNode }) {
  const params = useParams();
  const router = useRouter();
  const orgName = params.orgName as string;

  const { isLoading: authLoading, user } = useRequireAuth();
  const [isVerifying, setIsVerifying] = useState(true);
  const { setCurrentOrg, loadMemberships } = useOrganizationStore();

  // Verification logic...
  
  if (authLoading || isVerifying) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  return <>{children}</>;
}
```

---

### 2. CONTEXT + PROVIDER + HOOK PATTERN

**Pattern Used:** Zustand store (not React Context)  
**File Path:** `src/shared/stores/organization.ts`  
**Full URL:** `/data/home/lukemxjia/modelcraft/modelcraft-front/src/shared/stores/organization.ts`

**Complete File Content:**

```typescript
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { getMemberships, invalidateMembershipsCache, type MembershipInfo } from '@/shared/cache/memberships-cache'

interface OrganizationState {
  // Current selected organization
  currentOrg: string | null

  // All organizations user has access to
  organizations: string[]

  // Full membership data (includes roles, displayNames, etc.)
  memberships: MembershipInfo[]

  // Loading state
  isLoadingMemberships: boolean

  // Actions
  setCurrentOrg: (orgName: string) => void
  setOrganizations: (orgs: string[]) => void
  setMemberships: (memberships: MembershipInfo[]) => void
  switchOrganization: (orgName: string) => void
  clearOrganization: () => void

  // Async actions
  loadMemberships: (token: string, forceRefresh?: boolean) => Promise<MembershipInfo[]>
  refreshMemberships: (token: string) => Promise<MembershipInfo[]>
}

/**
 * Organization store using Zustand
 * Manages current organization context for multi-tenant features
 * Integrates with memberships-cache for optimized API calls
 */
export const useOrganizationStore = create<OrganizationState>()(
  persist(
    (set, get) => ({
      currentOrg: null,
      organizations: [],
      memberships: [],
      isLoadingMemberships: false,

      setCurrentOrg: (orgName: string) => {
        set({ currentOrg: orgName })
      },

      setOrganizations: (orgs: string[]) => {
        set({ organizations: orgs })

        // If no current org is selected, set the first one
        const { currentOrg } = get()
        if (!currentOrg && orgs.length > 0) {
          set({ currentOrg: orgs[0] })
        }
      },

      setMemberships: (memberships: MembershipInfo[]) => {
        const orgNames = memberships.map(m => m.orgName)
        set({
          memberships,
          organizations: orgNames,
        })

        // If no current org is selected, set the first one
        const { currentOrg } = get()
        if (!currentOrg && orgNames.length > 0) {
          set({ currentOrg: orgNames[0] })
        }
      },

      switchOrganization: (orgName: string) => {
        const { organizations } = get()

        // Verify user has access to this organization
        if (organizations.length > 0 && !organizations.includes(orgName)) {
          console.error(
            `User does not have access to organization: ${orgName}`
          )
          return
        }

        set({ currentOrg: orgName })

        // Navigation is handled by the OrganizationSwitcher component
        // which uses Next.js router to navigate to the new org URL
      },

      clearOrganization: () => {
        set({
          currentOrg: null,
          organizations: [],
          memberships: [],
        })
        // Clear memberships cache
        invalidateMembershipsCache()
      },

      // Load memberships with caching
      loadMemberships: async (token: string, forceRefresh = false) => {
        set({ isLoadingMemberships: true })

        try {
          const memberships = await getMemberships(token, forceRefresh)

          // Update store
          get().setMemberships(memberships)

          return memberships
        } catch (error) {
          console.error('[OrganizationStore] Failed to load memberships:', error)
          throw error
        } finally {
          set({ isLoadingMemberships: false })
        }
      },

      // Force refresh memberships
      refreshMemberships: async (token: string) => {
        console.log('[OrganizationStore] Refreshing memberships')
        return get().loadMemberships(token, true)
      },
    }),
    {
      name: 'organization-storage', // localStorage key
      // Only persist currentOrg and organizations (not memberships - they come from cache)
      partialize: (state) => ({
        currentOrg: state.currentOrg,
        organizations: state.organizations,
      }),
    }
  )
)

/**
 * Hook to get current organization name
 */
export function useCurrentOrg(): string | null {
  return useOrganizationStore((state) => state.currentOrg)
}

/**
 * Hook to get all organizations
 */
export function useOrganizations(): string[] {
  return useOrganizationStore((state) => state.organizations)
}

/**
 * Hook to get full membership data
 */
export function useMemberships(): MembershipInfo[] {
  return useOrganizationStore((state) => state.memberships)
}

/**
 * Hook to get memberships loading state
 */
export function useIsMembershipsLoading(): boolean {
  return useOrganizationStore((state) => state.isLoadingMemberships)
}

/**
 * Hook to switch organization
 */
export function useSwitchOrganization() {
  return useOrganizationStore((state) => state.switchOrganization)
}

/**
 * Hook to load memberships
 */
export function useLoadMemberships() {
  return useOrganizationStore((state) => state.loadMemberships)
}

/**
 * Hook to refresh memberships
 */
export function useRefreshMemberships() {
  return useOrganizationStore((state) => state.refreshMemberships)
}
```

**Key Pattern Characteristics:**
- Zustand store with TypeScript generics
- Persist middleware for localStorage
- Granular selector hooks to prevent unnecessary re-renders
- Async actions with loading state
- Separation of concerns (store state, actions, selectors)

---

### 3. MUTATION WITH onCompleted CALLBACK

**Primary Example File:** `src/web/components/features/model-editor/InsertFieldSheet.tsx`  
**Lines:** 197-219

**Example 1 - With Error Handling:**
```typescript
const [addField] = useMutation<
  AddFieldsMutationData,
  AddFieldsMutationVariables
>(ADD_FIELDS, {
  context: projectScopedContext,
  onCompleted: (data: { addFields?: { error?: { message?: string } | null } | null }) => {
    const bizError = data?.addFields?.error
    if (bizError) {
      setSubmitError(bizError.message ?? '添加字段失败')
      setSaving(false)
      return
    }
    setFieldData(DEFAULT_FIELD_DATA)
    setSubmitError(null)
    setSaving(false)
    onSuccess?.()
    if (!continueModeRef.current) {
      onOpenChange(false)
    }
  },
  onError: (error) => {
    setSubmitError('添加字段失败: ' + error.message)
    setSaving(false)
  },
  refetchQueries: ['GetModel', 'GetModelJsonSchema'],
})
```

**Example 2 - Simple Success Callback:**
```typescript
const [deleteContent] = useMutation(deleteMutation || NOOP_MUTATION, {
  client: runtimeClient!,
  onCompleted: () => {
    refetch()
    setDeleteDialogOpen(false)
    setDeleteItemId(null)
  },
})
```

**Pattern Characteristics:**
- Generic types for type safety: `useMutation<DataType, VariablesType>`
- Context passed for scoped queries
- `onCompleted` for success handling
- `onError` for error handling
- `refetchQueries` for cache invalidation

---

### 4. SHADCN/UI BUTTON COMPONENT IMPORT PATH

**Exact Import Path:** `@web/components/ui/button`  
**Full File Path:** `src/web/components/ui/button.tsx`

**Available Variants:**
- `default` (primary blue)
- `destructive` (red)
- `outline` (bordered)
- `secondary` (gray)
- `ghost` (transparent)
- `link` (text only)

**Available Sizes:**
- `default` (h-9)
- `sm` (h-8)
- `lg` (h-10)
- `icon` (square)

**Usage Examples:**
```typescript
import { Button } from '@web/components/ui/button'

// Default button
<Button>Click me</Button>

// With variant and size
<Button variant="destructive" size="lg">Delete</Button>

// As child (polymorphic)
<Button asChild>
  <a href="/page">Link as button</a>
</Button>

// With disabled state
<Button disabled>Disabled</Button>
```

---

### 5. cn() UTILITY IMPORT PATH

**Exact Import Path:** `@/shared/utils`  
**Full File Path:** `src/shared/utils.ts`

**Source Code:**
```typescript
import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

**Purpose:**
- Merges Tailwind CSS classes with conflict resolution
- clsx handles conditional classes
- twMerge resolves Tailwind conflicts

**Usage Examples:**
```typescript
import { cn } from '@/shared/utils'

// Basic usage
<div className={cn('px-4', 'py-2')}>Content</div>

// Conditional classes
<div className={cn(
  'base-class',
  isActive && 'bg-primary text-white',
  disabled && 'opacity-50'
)}>
  Content
</div>

// Overriding with precedence
<div className={cn(
  'px-4',
  customPadding // Will override px-4 if set
)}>
  Content
</div>
```

---

### 6. SIDEBAR STRUCTURE

**Main File:** `src/web/components/features/layout/AppLayout.tsx` (512 lines)  
**Full URL:** `/data/home/lukemxjia/modelcraft/modelcraft-front/src/web/components/features/layout/AppLayout.tsx`

**User Menu File:** `src/web/components/features/layout/UserMenu.tsx` (179 lines)

**Layout Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│  TOPBAR (h-12 / 48px)                                       │
│  ├─ Org Selector (Logo + Dropdown)                          │
│  ├─ Breadcrumbs (responsive, truncated)                     │
│  ├─ Spacer (flex-1)                                         │
│  ├─ Help Button (Ghost variant)                             │
│  └─ UserMenu Component (Avatar + Dropdown)                  │
└─────────────────────────────────────────────────────────────┘
┌────────────────────┬──────────────────────────────────────────┐
│ SIDEBAR            │                                          │
│ (w-60/240px or     │  MAIN CONTENT AREA                      │
│  w-16/64px)        │  ├─ Scrollable                          │
│                    │  ├─ bg-card                             │
│ ├─ Nav Sections    │  └─ min-h-screen                        │
│ │  ├─ Section 1    │                                          │
│ │  │  ├─ Nav Item  │                                          │
│ │  │  ├─ Active    │                                          │
│ │  │  └─ Sub-items │                                          │
│ │  └─ Section 2    │                                          │
│ │                  │                                          │
│ └─ Footer          │                                          │
│    └─ Collapse     │                                          │
│       Toggle       │                                          │
└────────────────────┴──────────────────────────────────────────┘
```

**Topbar Details:**
- Height: 48px (h-12)
- Org selector with logo + dropdown
- Breadcrumbs with truncation & navigation
- Help button (ghost style)
- UserMenu with Avatar

**Sidebar Details:**
- Width: 240px expanded, 64px collapsed
- Sections with headers (工作区, 设置, etc.)
- Icons from lucide-react
- Active state: 3px left border (border-l-primary) + bg tint
- Sub-items nested with expand/collapse
- Footer with toggle button

**UserMenu Component:**
```typescript
// Location: src/web/components/features/layout/UserMenu.tsx
// In Topbar (NOT sidebar)
<UserMenu
  userName={displayName}
  userEmail={userInfo?.phone}
  onLogout={handleLogout}
/>
```

Features:
- Avatar with initials fallback
- Dropdown menu (align="end")
- Profile link
- Settings link
- Logout button

**Example Navigation Structure:**
```typescript
const workspaceNavSections: NavSection[] = [
  {
    header: '工作区',
    items: [
      { label: '项目', icon: FolderOpen, href: `/org/${orgName}/workspace` },
      { label: '开发者', icon: Users, href: `/org/${orgName}/developers` },
      { label: '终端用户', icon: KeyRound, href: `/org/${orgName}/end-users` },
    ],
  },
  {
    header: '设置',
    items: [
      { label: '组织设置', icon: Settings, href: `/org/${orgName}/settings` },
    ],
  },
]
```

---

### 7. TEST FRAMEWORK

**Framework:** Vitest v4.1.2  
**Test Command:** `npm run test` (runs `vitest run`)

**Sample Test File:** `src/api-client/runtime-query/runtime-query-builder.test.ts`

**Test Structure Example:**
```typescript
import { describe, expect, it } from 'vitest'

describe('runtime-query-builder: DocumentNode validity', () => {
  const modelName = 'User'
  const fields = ['id', 'name', 'email']

  it('buildFindManyQuery returns valid DocumentNode', () => {
    const doc = buildFindManyQuery(modelName, fields)
    const printed = print(doc)
    expect(printed).toContain('findMany')
    assertValidDocument(printed)
  })

  it('buildFindUniqueQuery returns valid DocumentNode', () => {
    const doc = buildFindUniqueQuery(modelName, fields)
    const printed = print(doc)
    expect(printed).toContain('findUnique')
    assertValidDocument(printed)
  })

  describe('edge cases', () => {
    it('handles empty fields', () => {
      // test implementation
    })

    it('handles only id field', () => {
      // test implementation
    })
  })
})
```

**Package.json:**
```json
{
  "scripts": {
    "test": "vitest run",
    "test:watch": "vitest",
    "test:e2e": "playwright test"
  },
  "devDependencies": {
    "vitest": "^4.1.2"
  }
}
```

---

### 8. TYPESCRIPT TSCONFIG PATHS ALIAS

**File:** `tsconfig.json` (in project root)

**Complete Configuration:**
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [
      {
        "name": "next"
      }
    ],
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"],
      "@api-client/cms/*": ["./src/api-client/runtime-query/*"],
      "@api-client/*": ["./src/api-client/*"],
      "@web/*": ["./src/web/*"],
      "@shared/*": ["./src/shared/*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules", "menu-refactor", ".agents", "src/mocks", "e2e", "playwright.config.ts"]
}
```

**Path Alias Mapping:**

| Alias | Points To | Use Case |
|-------|-----------|----------|
| `@/*` | `src/*` | General imports, utilities |
| `@api-client/*` | `src/api-client/*` | GraphQL clients, queries |
| `@api-client/cms/*` | `src/api-client/runtime-query/*` | Special runtime query alias |
| `@web/*` | `src/web/*` | React components, hooks, stores |
| `@shared/*` | `src/shared/*` | Shared utilities, types, constants |

**Usage Examples:**
```typescript
// Instead of relative imports
import { cn } from '../../../../../shared/utils'
// Use alias imports
import { cn } from '@/shared/utils'

// Component imports
import { Button } from '@web/components/ui/button'
import { UserMenu } from '@web/components/features/layout/UserMenu'

// Store imports
import { useOrganizationStore } from '@shared/stores/organization'

// API client imports
import { ADD_FIELDS } from '@api-client/model'
```

---

## 🎯 QUICK IMPLEMENTATION CHECKLIST

### Essential Patterns to Follow

- [ ] **Client Components**: Use `'use client'` directive for interactive components
- [ ] **State Management**: Use Zustand with persist middleware for global state
- [ ] **Mutations**: Always use `onCompleted` + `onError` callbacks + `refetchQueries`
- [ ] **UI Components**: Import from `@web/components/ui/`
- [ ] **Styling**: Use `cn()` from `@/shared/utils` for conditional classes
- [ ] **Layout**: Wrap with `AppLayout` for sidebar + topbar
- [ ] **User Menu**: Place in topbar, not sidebar
- [ ] **Tests**: Use Vitest `describe`, `expect`, `it` syntax
- [ ] **Imports**: Always use path aliases (@/, @web/, @shared/, @api-client/)
- [ ] **Types**: Leverage TypeScript for mutation data and variables

---

## 📚 File Structure Reference

```
src/
├── app/
│   ├── org/
│   │   └── [orgName]/
│   │       └── layout.tsx ✓ (Org-level wrapper)
│   └── layout.tsx (Root layout)
├── web/
│   ├── components/
│   │   ├── ui/ (shadcn/ui components)
│   │   │   ├── button.tsx ✓
│   │   │   ├── dropdown-menu.tsx
│   │   │   ├── avatar.tsx
│   │   │   └── ...
│   │   └── features/
│   │       └── layout/
│   │           ├── AppLayout.tsx ✓ (Main sidebar + topbar)
│   │           └── UserMenu.tsx ✓ (User menu dropdown)
│   ├── hooks/
│   │   └── auth/
│   │       └── use-auth.ts
│   └── stores/
│       └── project.ts
├── shared/
│   ├── stores/
│   │   └── organization.ts ✓ (Zustand store)
│   ├── utils.ts ✓ (cn() utility)
│   ├── constants/
│   ├── cache/
│   └── ...
└── api-client/
    ├── apollo/
    ├── model/
    └── ...
```

---

## ✨ Final Notes

1. **Import Paths**: Always use aliases from tsconfig.json
2. **Styling**: `cn()` from `@/shared/utils` prevents class conflicts
3. **State**: Zustand stores with persist middleware for auth/org state
4. **Mutations**: Apollo Client with `onCompleted`, `onError`, `refetchQueries`
5. **Components**: shadcn/ui components from `@web/components/ui/`
6. **Testing**: Vitest with standard describe/expect/it syntax
7. **Layout**: AppLayout handles sidebar + topbar + navigation
8. **User Menu**: In topbar header, not in sidebar footer

All paths are verified and exact as of 2026-05-12.

