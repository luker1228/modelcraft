# Onboarding Checklist Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an on-demand floating progress panel that guides tenant admins through 8 steps — from project creation to end-user activation — with state persisted in localStorage.

**Architecture:** A React Context (`OnboardingContext`) wraps the Org-level layout, exposing a `useOnboarding()` hook. A floating `OnboardingPanel` component renders fixed bottom-right. Each mutation's `onCompleted` callback calls `markStep()` to advance the checklist. All state lives in `localStorage` key `mc_onboarding_v1`; no backend changes required.

**Tech Stack:** React Context, localStorage, Vitest, Tailwind CSS semantic tokens, shadcn/ui Button, lucide-react icons, Next.js `useRouter`

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `modelcraft-front/src/shared/onboarding/storage.ts` | Typed localStorage read/write |
| Create | `modelcraft-front/src/shared/onboarding/steps.ts` | Step definitions (id, label, route) |
| Create | `modelcraft-front/src/shared/onboarding/OnboardingContext.tsx` | Context, Provider, `useOnboarding()` hook |
| Create | `modelcraft-front/src/shared/onboarding/OnboardingPanel.tsx` | Floating panel UI (collapsed + expanded) |
| Create | `modelcraft-front/src/shared/onboarding/storage.test.ts` | Unit tests for storage module |
| Create | `modelcraft-front/src/shared/onboarding/OnboardingContext.test.tsx` | Unit tests for context logic |
| Modify | `modelcraft-front/src/app/org/[orgName]/layout.tsx` | Mount Provider + Panel |
| Modify | `modelcraft-front/src/web/components/features/layout/AppLayout.tsx` | Sidebar "快速开始" entry point |
| Modify | `modelcraft-front/src/web/hooks/project/use-projects-with-error-handling.ts` | `markStep('create_project')` |
| Modify | `modelcraft-front/src/web/hooks/model/use-models.ts` | `markStep('create_model')` |
| Modify | `modelcraft-front/src/web/components/features/model-editor/InsertFieldSheet.tsx` | `markStep('add_field')` |
| Modify | `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/permissions/_hooks/usePermissionsView.ts` | `markStep('apply_preset')` |
| Modify | `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/roles/_hooks/useRoleList.ts` | `markStep('create_role')` |
| Modify | `modelcraft-front/src/web/hooks/end-users/useOrgEndUsers.ts` | `markStep('add_end_user')` |
| Modify | `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/users/_hooks/useUserAuth.ts` | `markStep('assign_role')` |

---

## Task 1: storage.ts — typed localStorage wrapper

**Files:**
- Create: `modelcraft-front/src/shared/onboarding/storage.ts`
- Create: `modelcraft-front/src/shared/onboarding/storage.test.ts`

- [ ] **Step 1: Write failing tests**

Create `modelcraft-front/src/shared/onboarding/storage.test.ts`:

```ts
import { describe, it, expect, beforeEach } from 'vitest'
import {
  readOnboardingState,
  writeOnboardingState,
  ONBOARDING_KEY,
  defaultOnboardingState,
} from './storage'

describe('onboarding storage', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('returns default state when localStorage is empty', () => {
    const state = readOnboardingState('my-org')
    expect(state).toEqual(defaultOnboardingState('my-org'))
  })

  it('round-trips state to localStorage', () => {
    const state = {
      orgName: 'my-org',
      projectSlug: 'my-project',
      completedSteps: ['create_project' as const],
      dismissed: false,
      panelOpen: true,
    }
    writeOnboardingState(state)
    expect(readOnboardingState('my-org')).toEqual(state)
  })

  it('returns default state when orgName does not match stored state', () => {
    const state = {
      orgName: 'org-a',
      projectSlug: null,
      completedSteps: [],
      dismissed: false,
      panelOpen: false,
    }
    writeOnboardingState(state)
    const result = readOnboardingState('org-b')
    expect(result).toEqual(defaultOnboardingState('org-b'))
  })

  it('handles corrupt localStorage gracefully', () => {
    localStorage.setItem(ONBOARDING_KEY, 'not-json')
    const result = readOnboardingState('my-org')
    expect(result).toEqual(defaultOnboardingState('my-org'))
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd modelcraft-front && npx vitest run src/shared/onboarding/storage.test.ts
```
Expected: FAIL — "Cannot find module './storage'"

- [ ] **Step 3: Implement storage.ts**

Create `modelcraft-front/src/shared/onboarding/storage.ts`:

```ts
export const ONBOARDING_KEY = 'mc_onboarding_v1'

export type OnboardingStepId =
  | 'create_project'
  | 'create_model'
  | 'add_field'
  | 'apply_preset'
  | 'create_role'
  | 'add_end_user'
  | 'assign_role'
  | 'end_user_login'

export interface OnboardingState {
  orgName: string
  projectSlug: string | null
  completedSteps: OnboardingStepId[]
  dismissed: boolean
  panelOpen: boolean
}

export function defaultOnboardingState(orgName: string): OnboardingState {
  return {
    orgName,
    projectSlug: null,
    completedSteps: [],
    dismissed: false,
    panelOpen: false,
  }
}

export function readOnboardingState(orgName: string): OnboardingState {
  try {
    const raw = localStorage.getItem(ONBOARDING_KEY)
    if (!raw) return defaultOnboardingState(orgName)
    const parsed = JSON.parse(raw) as OnboardingState
    if (parsed.orgName !== orgName) return defaultOnboardingState(orgName)
    return parsed
  } catch {
    return defaultOnboardingState(orgName)
  }
}

export function writeOnboardingState(state: OnboardingState): void {
  localStorage.setItem(ONBOARDING_KEY, JSON.stringify(state))
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd modelcraft-front && npx vitest run src/shared/onboarding/storage.test.ts
```
Expected: PASS — 4 tests passing

- [ ] **Step 5: Commit**

```bash
git add modelcraft-front/src/shared/onboarding/storage.ts modelcraft-front/src/shared/onboarding/storage.test.ts
git commit -m "feat(onboarding): add typed localStorage storage module"
```

---

## Task 2: steps.ts — step definitions

**Files:**
- Create: `modelcraft-front/src/shared/onboarding/steps.ts`

- [ ] **Step 1: Create steps.ts**

Create `modelcraft-front/src/shared/onboarding/steps.ts`:

```ts
import type { OnboardingStepId } from './storage'

export interface OnboardingStepDef {
  id: OnboardingStepId
  label: string
  description: string
  /** Build the navigation route for the CTA button */
  route: (params: { orgName: string; projectSlug: string | null }) => string | null
}

export const ONBOARDING_STEPS: OnboardingStepDef[] = [
  {
    id: 'create_project',
    label: '创建项目',
    description: '在组织下创建第一个项目',
    route: ({ orgName }) => `/org/${orgName}/workspace`,
  },
  {
    id: 'create_model',
    label: '创建模型',
    description: '在项目中定义第一个数据模型',
    route: ({ orgName, projectSlug }) =>
      projectSlug ? `/org/${orgName}/project/${projectSlug}/model-editor` : null,
  },
  {
    id: 'add_field',
    label: '添加字段',
    description: '为模型添加至少一个字段',
    route: ({ orgName, projectSlug }) =>
      projectSlug ? `/org/${orgName}/project/${projectSlug}/model-editor` : null,
  },
  {
    id: 'apply_preset',
    label: '应用权限预设',
    description: '为模型选择终端用户权限预设策略',
    route: ({ orgName, projectSlug }) =>
      projectSlug ? `/org/${orgName}/project/${projectSlug}/rbac/permissions` : null,
  },
  {
    id: 'create_role',
    label: '创建角色',
    description: '创建一个终端用户角色',
    route: ({ orgName, projectSlug }) =>
      projectSlug ? `/org/${orgName}/project/${projectSlug}/rbac/roles` : null,
  },
  {
    id: 'add_end_user',
    label: '添加终端用户',
    description: '在组织中创建第一个终端用户账号',
    route: ({ orgName }) => `/org/${orgName}/end-users`,
  },
  {
    id: 'assign_role',
    label: '分配角色',
    description: '将角色授予终端用户',
    route: ({ orgName }) => `/org/${orgName}/end-users`,
  },
  {
    id: 'end_user_login',
    label: '终端用户登录',
    description: '以终端用户身份登录，验证整条链路',
    route: () => null, // manual confirmation — panel shows login URL
  },
]
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-front/src/shared/onboarding/steps.ts
git commit -m "feat(onboarding): add step definitions"
```

---

## Task 3: OnboardingContext.tsx — context, provider, hook

**Files:**
- Create: `modelcraft-front/src/shared/onboarding/OnboardingContext.tsx`
- Create: `modelcraft-front/src/shared/onboarding/OnboardingContext.test.tsx`

- [ ] **Step 1: Write failing tests**

Create `modelcraft-front/src/shared/onboarding/OnboardingContext.test.tsx`:

```tsx
import { describe, it, expect, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { OnboardingProvider, useOnboarding } from './OnboardingContext'
import React from 'react'

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <OnboardingProvider orgName="test-org">{children}</OnboardingProvider>
)

describe('useOnboarding', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('starts with 0 completed steps', () => {
    const { result } = renderHook(() => useOnboarding(), { wrapper })
    expect(result.current.completedCount).toBe(0)
    expect(result.current.isComplete).toBe(false)
  })

  it('markStep advances completedCount', () => {
    const { result } = renderHook(() => useOnboarding(), { wrapper })
    act(() => {
      result.current.markStep('create_project')
    })
    expect(result.current.completedCount).toBe(1)
  })

  it('markStep with create_project stores projectSlug', () => {
    const { result } = renderHook(() => useOnboarding(), { wrapper })
    act(() => {
      result.current.markStep('create_project', 'my-project')
    })
    expect(result.current.projectSlug).toBe('my-project')
  })

  it('markStep is idempotent', () => {
    const { result } = renderHook(() => useOnboarding(), { wrapper })
    act(() => {
      result.current.markStep('create_project')
      result.current.markStep('create_project')
    })
    expect(result.current.completedCount).toBe(1)
  })

  it('isComplete is true after all 8 steps', () => {
    const { result } = renderHook(() => useOnboarding(), { wrapper })
    act(() => {
      result.current.markStep('create_project')
      result.current.markStep('create_model')
      result.current.markStep('add_field')
      result.current.markStep('apply_preset')
      result.current.markStep('create_role')
      result.current.markStep('add_end_user')
      result.current.markStep('assign_role')
      result.current.markStep('end_user_login')
    })
    expect(result.current.isComplete).toBe(true)
  })

  it('dismiss sets dismissed flag', () => {
    const { result } = renderHook(() => useOnboarding(), { wrapper })
    act(() => {
      result.current.dismiss()
    })
    expect(result.current.dismissed).toBe(true)
  })

  it('openPanel and closePanel toggle panelOpen', () => {
    const { result } = renderHook(() => useOnboarding(), { wrapper })
    act(() => result.current.openPanel())
    expect(result.current.panelOpen).toBe(true)
    act(() => result.current.closePanel())
    expect(result.current.panelOpen).toBe(false)
  })

  it('currentStep is the first incomplete step', () => {
    const { result } = renderHook(() => useOnboarding(), { wrapper })
    act(() => {
      result.current.markStep('create_project')
      result.current.markStep('create_model')
    })
    expect(result.current.currentStep?.id).toBe('add_field')
  })

  it('reset clears all state', () => {
    const { result } = renderHook(() => useOnboarding(), { wrapper })
    act(() => {
      result.current.markStep('create_project')
      result.current.reset()
    })
    expect(result.current.completedCount).toBe(0)
    expect(result.current.projectSlug).toBeNull()
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd modelcraft-front && npx vitest run src/shared/onboarding/OnboardingContext.test.tsx
```
Expected: FAIL — "Cannot find module './OnboardingContext'"

- [ ] **Step 3: Implement OnboardingContext.tsx**

Create `modelcraft-front/src/shared/onboarding/OnboardingContext.tsx`:

```tsx
'use client'

import React, { createContext, useCallback, useContext, useEffect, useState } from 'react'
import {
  type OnboardingState,
  type OnboardingStepId,
  defaultOnboardingState,
  readOnboardingState,
  writeOnboardingState,
} from './storage'
import { ONBOARDING_STEPS, type OnboardingStepDef } from './steps'

export interface OnboardingStep extends OnboardingStepDef {
  status: 'completed' | 'current' | 'locked'
  index: number // 1-based
}

interface OnboardingContextValue {
  steps: OnboardingStep[]
  currentStep: OnboardingStep | null
  projectSlug: string | null
  completedCount: number
  totalCount: number
  isComplete: boolean
  panelOpen: boolean
  dismissed: boolean
  markStep: (id: OnboardingStepId, projectSlug?: string) => void
  openPanel: () => void
  closePanel: () => void
  dismiss: () => void
  reset: () => void
}

const OnboardingContext = createContext<OnboardingContextValue | null>(null)

export function OnboardingProvider({
  orgName,
  children,
}: {
  orgName: string
  children: React.ReactNode
}) {
  const [state, setState] = useState<OnboardingState>(() =>
    defaultOnboardingState(orgName)
  )

  // Hydrate from localStorage on mount (client only)
  useEffect(() => {
    setState(readOnboardingState(orgName))
  }, [orgName])

  const persist = useCallback(
    (next: OnboardingState) => {
      setState(next)
      writeOnboardingState(next)
    },
    []
  )

  const markStep = useCallback(
    (id: OnboardingStepId, projectSlug?: string) => {
      setState((prev) => {
        if (prev.completedSteps.includes(id)) return prev
        const next: OnboardingState = {
          ...prev,
          completedSteps: [...prev.completedSteps, id],
          projectSlug:
            id === 'create_project' && projectSlug ? projectSlug : prev.projectSlug,
        }
        writeOnboardingState(next)
        return next
      })
    },
    []
  )

  const openPanel = useCallback(() => {
    persist({ ...state, panelOpen: true })
  }, [state, persist])

  const closePanel = useCallback(() => {
    persist({ ...state, panelOpen: false })
  }, [state, persist])

  const dismiss = useCallback(() => {
    persist({ ...state, dismissed: true, panelOpen: false })
  }, [state, persist])

  const reset = useCallback(() => {
    persist(defaultOnboardingState(orgName))
  }, [orgName, persist])

  // Derive step statuses
  const completedSet = new Set(state.completedSteps)
  let foundCurrent = false
  const steps: OnboardingStep[] = ONBOARDING_STEPS.map((def, i) => {
    const completed = completedSet.has(def.id)
    let status: OnboardingStep['status']
    if (completed) {
      status = 'completed'
    } else if (!foundCurrent) {
      status = 'current'
      foundCurrent = true
    } else {
      status = 'locked'
    }
    return { ...def, status, index: i + 1 }
  })

  const currentStep = steps.find((s) => s.status === 'current') ?? null
  const completedCount = state.completedSteps.length
  const totalCount = ONBOARDING_STEPS.length
  const isComplete = completedCount === totalCount

  return (
    <OnboardingContext.Provider
      value={{
        steps,
        currentStep,
        projectSlug: state.projectSlug,
        completedCount,
        totalCount,
        isComplete,
        panelOpen: state.panelOpen,
        dismissed: state.dismissed,
        markStep,
        openPanel,
        closePanel,
        dismiss,
        reset,
      }}
    >
      {children}
    </OnboardingContext.Provider>
  )
}

export function useOnboarding(): OnboardingContextValue {
  const ctx = useContext(OnboardingContext)
  if (!ctx) {
    throw new Error('useOnboarding must be used within OnboardingProvider')
  }
  return ctx
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd modelcraft-front && npx vitest run src/shared/onboarding/OnboardingContext.test.tsx
```
Expected: PASS — 9 tests passing

- [ ] **Step 5: Commit**

```bash
git add modelcraft-front/src/shared/onboarding/OnboardingContext.tsx modelcraft-front/src/shared/onboarding/OnboardingContext.test.tsx
git commit -m "feat(onboarding): add context, provider, and useOnboarding hook"
```

---

## Task 4: OnboardingPanel.tsx — floating UI

**Files:**
- Create: `modelcraft-front/src/shared/onboarding/OnboardingPanel.tsx`

- [ ] **Step 1: Create OnboardingPanel.tsx**

Create `modelcraft-front/src/shared/onboarding/OnboardingPanel.tsx`:

```tsx
'use client'

import { useRouter } from 'next/navigation'
import { ChevronUp, ChevronDown, X, Check, ArrowRight } from 'lucide-react'
import { cn } from '@/shared/utils'
import { Button } from '@web/components/ui/button'
import { useOnboarding } from './OnboardingContext'

export function OnboardingPanel({ orgName }: { orgName: string }) {
  const {
    steps,
    currentStep,
    projectSlug,
    completedCount,
    totalCount,
    isComplete,
    panelOpen,
    openPanel,
    closePanel,
    dismiss,
    markStep,
  } = useOnboarding()

  const router = useRouter()

  if (isComplete) return null

  const progressPct = (completedCount / totalCount) * 100

  const handleCta = () => {
    if (!currentStep) return
    const route = currentStep.route({ orgName, projectSlug })
    if (route) {
      router.push(route)
      closePanel()
    }
  }

  // Collapsed state
  if (!panelOpen) {
    return (
      <div
        className="fixed bottom-6 right-6 z-50 flex cursor-pointer items-center gap-3 rounded-lg border border-border bg-white px-3 py-2.5 shadow-md transition-shadow hover:shadow-lg"
        onClick={openPanel}
        role="button"
        aria-label="展开快速开始面板"
      >
        {/* Left indigo accent bar */}
        <div className="h-8 w-0.5 flex-shrink-0 rounded-full bg-primary" />
        <div className="flex flex-col gap-1">
          <span className="text-[12px] font-semibold text-foreground">快速开始</span>
          {/* Progress bar */}
          <div className="h-1 w-32 overflow-hidden rounded-full bg-[#EBEEF2]">
            <div
              className="h-full rounded-full bg-primary transition-all duration-300"
              style={{ width: `${progressPct}%` }}
            />
          </div>
          <span className="text-[10px] font-medium text-muted-foreground">
            {completedCount} / {totalCount} 步完成
          </span>
        </div>
        <ChevronUp className="ml-1 size-3.5 text-muted-foreground" />
      </div>
    )
  }

  // Expanded state
  return (
    <div className="fixed bottom-6 right-6 z-50 w-[232px] overflow-hidden rounded-xl border border-border bg-white shadow-lg">
      {/* Header */}
      <div className="border-b border-border px-3.5 py-3">
        <div className="mb-2 flex items-center justify-between">
          <span className="text-[12px] font-semibold text-foreground">快速开始</span>
          <button
            onClick={dismiss}
            className="flex size-5 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            aria-label="关闭并不再显示"
          >
            <X className="size-3" />
          </button>
        </div>
        {/* Progress bar */}
        <div className="h-1 overflow-hidden rounded-full bg-[#EBEEF2]">
          <div
            className="h-full rounded-full bg-primary transition-all duration-300"
            style={{ width: `${progressPct}%` }}
          />
        </div>
        <p className="mt-1 text-[10px] font-medium text-muted-foreground">
          {completedCount} / {totalCount} 步完成
        </p>
      </div>

      {/* Steps list */}
      <div className="py-1.5">
        {steps.map((step) => (
          <div
            key={step.id}
            className={cn(
              'flex items-center gap-2 px-3.5 py-1.5',
              step.status === 'current' && 'border-l-[3px] border-primary bg-primary/[0.06] pl-[11px]'
            )}
          >
            {/* Step indicator */}
            {step.status === 'completed' ? (
              <div className="flex size-4 flex-shrink-0 items-center justify-center rounded-full border border-[#10b981]/30 bg-[#10b981]/10">
                <Check className="size-2.5 text-[#10b981]" strokeWidth={2.5} />
              </div>
            ) : step.status === 'current' ? (
              <div className="flex size-4 flex-shrink-0 items-center justify-center rounded-full border-[1.5px] border-primary bg-primary/10">
                <span className="text-[9px] font-semibold text-primary">{step.index}</span>
              </div>
            ) : (
              <div className="flex size-4 flex-shrink-0 items-center justify-center rounded-full border border-border">
                <span className="text-[9px] text-muted-foreground">{step.index}</span>
              </div>
            )}

            {/* Label */}
            <span
              className={cn(
                'flex-1 text-[11px]',
                step.status === 'completed' && 'text-muted-foreground line-through',
                step.status === 'current' && 'font-semibold text-primary',
                step.status === 'locked' && 'text-muted-foreground'
              )}
            >
              {step.label}
            </span>

            {/* Arrow for current step */}
            {step.status === 'current' && (
              <ArrowRight className="size-3 flex-shrink-0 text-primary" />
            )}
          </div>
        ))}

        {/* Step 8 manual confirm button (shown when step 8 is current) */}
        {currentStep?.id === 'end_user_login' && (
          <div className="mx-3.5 mt-2 rounded-md border border-border bg-[#F6F8FA] px-3 py-2">
            <p className="mb-1.5 text-[10px] text-muted-foreground">终端用户登录地址：</p>
            <code className="block break-all font-mono text-[10px] text-foreground">
              /end-user/org/{orgName}/login
            </code>
            <Button
              size="sm"
              variant="outline"
              className="mt-2 h-7 w-full text-[11px]"
              onClick={() => markStep('end_user_login')}
            >
              已完成 ✓
            </Button>
          </div>
        )}
      </div>

      {/* CTA footer */}
      {currentStep && currentStep.id !== 'end_user_login' && (
        <div className="border-t border-border px-3.5 py-2.5">
          <Button
            size="sm"
            className="h-8 w-full text-[11px]"
            onClick={handleCta}
          >
            前往：{currentStep.label} →
          </Button>
        </div>
      )}

      {/* Collapse chevron */}
      <button
        onClick={closePanel}
        className="flex w-full items-center justify-center border-t border-border py-1.5 text-muted-foreground transition-colors hover:bg-accent"
        aria-label="折叠面板"
      >
        <ChevronDown className="size-3.5" />
      </button>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-front/src/shared/onboarding/OnboardingPanel.tsx
git commit -m "feat(onboarding): add floating panel UI component"
```

---

## Task 5: Mount Provider and Panel in Org layout

**Files:**
- Modify: `modelcraft-front/src/app/org/[orgName]/layout.tsx`

Current file ends with:

```tsx
  return <>{children}</>
```

- [ ] **Step 1: Add Provider and Panel to layout**

Edit `modelcraft-front/src/app/org/[orgName]/layout.tsx` — add imports and wrap children:

```tsx
'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useRequireAuth } from '@web/hooks/auth/use-auth'
import { useOrganizationStore } from '@shared/stores/organization'
import { TENANT_LOGIN_PATH } from '@shared/constants/routes'
import { OnboardingProvider } from '@shared/onboarding/OnboardingContext'
import { OnboardingPanel } from '@shared/onboarding/OnboardingPanel'

export default function OrgLayout({ children }: { children: React.ReactNode }) {
  const params = useParams()
  const router = useRouter()
  const orgName = params.orgName as string

  const { isLoading: authLoading, user } = useRequireAuth()
  const [isVerifying, setIsVerifying] = useState(true)
  const { setCurrentOrg, loadMemberships } = useOrganizationStore()

  useEffect(() => {
    if (authLoading) return

    async function verifyOrgAccess() {
      console.log('[OrgLayout] Verifying org access:', orgName, 'user:', user?.id)

      try {
        const { getToken } = await import('@api-client/auth/public')
        const token = getToken()
        if (!token) {
          console.warn('[OrgLayout] No token after auth restore')
          router.push(TENANT_LOGIN_PATH)
          return
        }

        console.log('[OrgLayout] Loading memberships...')
        const memberships = await loadMemberships(token, false)
        console.log('[OrgLayout] Memberships:', memberships.map((m) => m.orgName))

        const hasAccess = memberships.some((m) => m.orgName === orgName)
        console.log('[OrgLayout] Has access to org:', orgName, '→', hasAccess)

        if (!hasAccess) {
          const fallbackOrgName = memberships[0]?.orgName
          if (fallbackOrgName) {
            console.warn(`[OrgLayout] Access denied to "${orgName}", redirecting to fallback org "${fallbackOrgName}"`)
            localStorage.setItem('defaultOrgName', fallbackOrgName)
            router.push(`/org/${fallbackOrgName}/workspace`)
            return
          }
          console.warn(`[OrgLayout] Access denied to "${orgName}" and no memberships found, redirecting to login`)
          localStorage.removeItem('defaultOrgName')
          router.push(TENANT_LOGIN_PATH)
          return
        }

        setCurrentOrg(orgName)
        localStorage.setItem('defaultOrgName', orgName)
        console.log('[OrgLayout] Org access verified ✓')
        setIsVerifying(false)
      } catch (error) {
        console.error('[OrgLayout] Error verifying org access:', error)
        localStorage.removeItem('defaultOrgName')
        router.push(TENANT_LOGIN_PATH)
      }
    }

    verifyOrgAccess()
  }, [authLoading, orgName, router, setCurrentOrg, loadMemberships, user?.id])

  if (authLoading || isVerifying) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  return (
    <OnboardingProvider orgName={orgName}>
      {children}
      <OnboardingPanel orgName={orgName} />
    </OnboardingProvider>
  )
}
```

- [ ] **Step 2: Verify the page compiles**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep -i onboarding
```
Expected: no errors mentioning onboarding files

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/app/org/\[orgName\]/layout.tsx
git commit -m "feat(onboarding): mount OnboardingProvider and panel in Org layout"
```

---

## Task 6: Sidebar entry point in AppLayout

**Files:**
- Modify: `modelcraft-front/src/web/components/features/layout/AppLayout.tsx`

The sidebar footer section (lines 487–500) renders only the collapse toggle button. We add the "快速开始" entry above the border-t footer.

- [ ] **Step 1: Add imports and entry point**

In `modelcraft-front/src/web/components/features/layout/AppLayout.tsx`:

**Add import** (near the top with other imports, after existing imports):
```tsx
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
```

**Add hook call** inside the `AppLayout` component function body, after existing hook calls:
```tsx
const { completedCount, totalCount, isComplete, dismissed, openPanel } = useOnboarding()
```

**Replace the sidebar footer div** (the `div` at line 487 that contains the toggle button) with:

```tsx
          {/* Sidebar footer */}
          <div className="flex-shrink-0 border-t border-border">
            {/* Quick start entry — shown when onboarding is incomplete and not dismissed */}
            {!isComplete && !dismissed && (
              <button
                onClick={openPanel}
                className={cn(
                  'flex w-full items-center gap-2 px-3 py-2 text-left transition-colors',
                  sidebarCollapsed ? 'justify-center' : '',
                  'text-muted-foreground hover:bg-accent hover:text-foreground'
                )}
                title="快速开始"
              >
                <Sparkles className="size-4 flex-shrink-0" strokeWidth={1.5} />
                {!sidebarCollapsed && (
                  <>
                    <span className="flex-1 text-[12px] font-medium">快速开始</span>
                    <span className="rounded bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium text-primary">
                      {completedCount}/{totalCount}
                    </span>
                  </>
                )}
              </button>
            )}
            {/* Collapse toggle */}
            <div className="flex h-11 items-center px-2">
              <button
                onClick={toggleSidebar}
                className="flex size-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                title={sidebarCollapsed ? '展开侧边栏' : '折叠侧边栏'}
              >
                {sidebarCollapsed ? (
                  <PanelLeft className="size-4" strokeWidth={1.5} />
                ) : (
                  <PanelLeftClose className="size-4" strokeWidth={1.5} />
                )}
              </button>
            </div>
          </div>
```

- [ ] **Step 2: Verify compilation**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | head -20
```
Expected: no type errors

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/components/features/layout/AppLayout.tsx
git commit -m "feat(onboarding): add quick-start sidebar entry point"
```

---

## Task 7: Wire markStep into mutations (7 integration points)

**Files:**
- Modify: `modelcraft-front/src/web/hooks/project/use-projects-with-error-handling.ts`
- Modify: `modelcraft-front/src/web/hooks/model/use-models.ts`
- Modify: `modelcraft-front/src/web/components/features/model-editor/InsertFieldSheet.tsx`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/permissions/_hooks/usePermissionsView.ts`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/roles/_hooks/useRoleList.ts`
- Modify: `modelcraft-front/src/web/hooks/end-users/useOrgEndUsers.ts`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/users/_hooks/useUserAuth.ts`

> **Important:** `useOnboarding()` may throw if called outside `OnboardingProvider`. All 7 files are rendered within the Org layout (which now wraps with `OnboardingProvider`), so this is safe. If a hook is also used in non-Org contexts, wrap the call in a try-catch or add a null-safe context check.

### 7a — createProject (`use-projects-with-error-handling.ts`)

The `onCompleted` already exists. Add `markStep` call. The `createProject` mutation returns the new project's slug at `mutationData.createProject.project.slug`.

- [ ] **Step 1: Add import and markStep call**

In `modelcraft-front/src/web/hooks/project/use-projects-with-error-handling.ts`, add the import at the top:
```ts
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
```

Add hook call inside the hook function body:
```ts
const { markStep } = useOnboarding()
```

In the existing `createProject` `onCompleted` callback, add `markStep`:
```ts
onCompleted: (mutationData) => {
  if (mutationData?.createProject?.project) {
    addProject(mutationData.createProject.project)
    markStep('create_project', mutationData.createProject.project.slug)
  }
},
```

### 7b — createModel (`use-models.ts`)

- [ ] **Step 2: Add import and markStep call**

In `modelcraft-front/src/web/hooks/model/use-models.ts`, add import:
```ts
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
```

Add hook call inside the hook function body:
```ts
const { markStep } = useOnboarding()
```

In the existing `createModel` `onCompleted` callback:
```ts
onCompleted: (mutationData) => {
  if (mutationData?.createModel?.model) {
    addModel(mutationData.createModel.model)
    markStep('create_model')
  }
},
```

### 7c — addFields (`InsertFieldSheet.tsx`)

- [ ] **Step 3: Add import and markStep call**

In `modelcraft-front/src/web/components/features/model-editor/InsertFieldSheet.tsx`, add import:
```ts
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
```

Add hook call inside the component body:
```ts
const { markStep } = useOnboarding()
```

In the existing `onCompleted` callback (around line 199), add after `onSuccess?.()`:
```ts
onCompleted: (data) => {
  const bizError = data?.addFields?.error
  if (bizError) {
    setSubmitError(bizError.message ?? '添加字段失败')
    setSaving(false)
    return
  }
  setFieldData(DEFAULT_FIELD_DATA)
  setSubmitError(null)
  setSaving(false)
  markStep('add_field')   // ← add this line
  onSuccess?.()
  if (!continueModeRef.current) {
    onOpenChange(false)
  }
},
```

### 7d — applyEndUserPresetPolicy (`usePermissionsView.ts`)

The `applyPresetPolicyMutation` currently has no `onCompleted`. Use the mutation's returned promise instead by modifying the call site.

- [ ] **Step 4: Add import and markStep call**

In `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/permissions/_hooks/usePermissionsView.ts`, add import:
```ts
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
```

Add hook call inside the hook function body:
```ts
const { markStep } = useOnboarding()
```

Add `onCompleted` to the `applyPresetPolicyMutation` useMutation call:
```ts
const [applyPresetPolicyMutation] = useMutation(APPLY_END_USER_PRESET_POLICY, {
  client,
  refetchQueries: [GET_END_USER_PERMISSIONS],
  onCompleted: () => markStep('apply_preset'),
})
```

### 7e — createEndUserRole (`useRoleList.ts`)

- [ ] **Step 5: Add import and markStep call**

In `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/roles/_hooks/useRoleList.ts`, add import:
```ts
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
```

Add hook call inside the hook function body:
```ts
const { markStep } = useOnboarding()
```

Add `onCompleted` to the `createRoleMutation` useMutation call:
```ts
const [createRoleMutation] = useMutation(CREATE_END_USER_ROLE, {
  client,
  refetchQueries: [GET_END_USER_ROLES],
  onCompleted: () => markStep('create_role'),
})
```

### 7f — createEndUser (`useOrgEndUsers.ts`)

This hook uses `client.mutate()` directly (no `onCompleted` callback style). Add `markStep` call after the successful mutation.

- [ ] **Step 6: Add import and markStep call**

In `modelcraft-front/src/web/hooks/end-users/useOrgEndUsers.ts`, add import:
```ts
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
```

Add hook call inside the hook function body:
```ts
const { markStep } = useOnboarding()
```

In the existing `createUser` callback, add `markStep` after `reload()`:
```ts
const createUser = useCallback(async (payload: CreateEndUserPayload) => {
  const client = getOrgScopedClient()
  const { data } = await client.mutate<CreateEndUserData>({
    mutation: CREATE_END_USER,
    variables: { input: { username: payload.username, password: payload.password } },
  })
  const err = data?.createEndUser?.error
  if (err) {
    throw new Error(err.message ?? '创建用户失败')
  }
  markStep('add_end_user')   // ← add this line
  reload()
}, [reload, markStep])
```

### 7g — assignEndUserRole (`useUserAuth.ts`)

- [ ] **Step 7: Add import and markStep call**

In `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/users/_hooks/useUserAuth.ts`, add import:
```ts
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
```

Add hook call inside the hook function body:
```ts
const { markStep } = useOnboarding()
```

In the existing `assignRole` callback, add `markStep` on success:
```ts
const assignRole = useCallback(
  async (endUserId: string, roleId: string): Promise<MutationResult> => {
    const result = await assignRoleMutation({
      variables: { endUserId, roleId },
    })
    const payload = result.data?.assignEndUserRole
    if (payload?.error) {
      return { success: false, errorMessage: payload.error.message ?? '分配角色失败' }
    }
    markStep('assign_role')   // ← add this line
    return { success: true }
  },
  [assignRoleMutation, markStep]
)
```

- [ ] **Step 8: Verify compilation**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | head -30
```
Expected: no type errors

- [ ] **Step 9: Commit all 7 integration files**

```bash
git add \
  modelcraft-front/src/web/hooks/project/use-projects-with-error-handling.ts \
  modelcraft-front/src/web/hooks/model/use-models.ts \
  modelcraft-front/src/web/components/features/model-editor/InsertFieldSheet.tsx \
  "modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/permissions/_hooks/usePermissionsView.ts" \
  "modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/roles/_hooks/useRoleList.ts" \
  modelcraft-front/src/web/hooks/end-users/useOrgEndUsers.ts \
  "modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/rbac/users/_hooks/useUserAuth.ts"
git commit -m "feat(onboarding): wire markStep into all 7 mutation completion handlers"
```

---

## Task 8: Full test run and smoke check

- [ ] **Step 1: Run all onboarding unit tests**

```bash
cd modelcraft-front && npx vitest run src/shared/onboarding/
```
Expected: all tests pass (storage.test.ts + OnboardingContext.test.tsx)

- [ ] **Step 2: Run full type check**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1
```
Expected: 0 errors

- [ ] **Step 3: Run lint**

```bash
cd modelcraft-front && npm run lint
```
Expected: no errors

- [ ] **Step 4: Manual smoke check**

Start the dev server (`npm run dev` in `modelcraft-front`), log in as a tenant admin, and verify:

1. Sidebar shows "快速开始" button with "0/8" badge
2. Clicking it opens the floating panel (collapsed → expanded)
3. Creating a project marks step 1 ✓ and stores `projectSlug`
4. Creating a model marks step 2 ✓
5. Adding a field marks step 3 ✓
6. Applying a preset marks step 4 ✓
7. Creating a role marks step 5 ✓
8. Adding an end user marks step 6 ✓
9. Assigning a role marks step 7 ✓
10. Step 8 shows login URL + "已完成 ✓" button; clicking marks complete
11. Panel disappears after all 8 steps; sidebar entry hides
12. Refresh browser — progress is preserved (localStorage)
13. Clicking ✕ dismisses panel and hides sidebar entry permanently

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "feat(onboarding): complete onboarding checklist feature"
```
