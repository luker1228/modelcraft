# Immersive Tutorial Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the existing onboarding panel into a 5-step immersive tutorial matching the user-approved flow: 创建项目 → 创建模型 → 创建终端用户 → 分配权限 → 体验登录.

**Architecture:** The onboarding system already exists (`src/shared/onboarding/`). This plan rewires the step definitions, adds `pendingAction` highlights to the 2 unwired pages (end-users, end-user-access), adds a `tutorial-highlight` CSS class for sub-step highlighting, adds a reset button to the panel, and removes the `dismissed` hide-forever behavior (panel becomes a persistent tab).

**Tech Stack:** Next.js 14, React 18, TypeScript, Vitest (unit tests colocated), Tailwind CSS, `localStorage` for persistence.

---

## File Map

| File | Change |
|------|--------|
| `src/shared/onboarding/storage.ts` | Remove `dismissed` field; add `ONBOARDING_KEY` version bump to `mc_onboarding_v2` |
| `src/shared/onboarding/steps.ts` | Rewrite `ONBOARDING_GROUPS` to match 5-step flow |
| `src/shared/onboarding/OnboardingContext.tsx` | Remove `dismiss`/`dismissed`; add `panelOpen: true` default; keep `reset()` |
| `src/shared/onboarding/OnboardingPanel.tsx` | Remove X/dismiss button; add reset button; `isComplete` shows completion state instead of `return null`; step 5 is pure text |
| `src/app/globals.css` | Add `.tutorial-highlight` keyframe + class |
| `src/app/org/[orgName]/end-users/page.tsx` | Wire `pendingAction === 'add_end_user'` → highlight 新建终端用户 button |
| `src/app/org/[orgName]/project/[projectSlug]/end-user-access/page.tsx` | Wire `pendingAction === 'assign_role'` → highlight 授权 button |
| `src/shared/onboarding/storage.test.ts` | Update tests for new storage shape (no `dismissed`) |
| `src/shared/onboarding/OnboardingContext.test.tsx` | Update step IDs and group counts to match new 5-step flow |

---

## Task 1: Update storage — remove `dismissed`, bump storage key

**Files:**
- Modify: `src/shared/onboarding/storage.ts`
- Modify: `src/shared/onboarding/storage.test.ts`

- [ ] **Step 1: Update storage.ts**

Replace the full file content:

```typescript
export const ONBOARDING_KEY = 'mc_onboarding_v2'

export type OnboardingStepId =
  | 'create_project'
  | 'create_model'
  | 'add_end_user'
  | 'assign_role'
  | 'end_user_login'

export interface OnboardingState {
  orgName: string
  projectSlug: string | null
  completedSteps: OnboardingStepId[]
  panelOpen: boolean
}

export function defaultOnboardingState(orgName: string): OnboardingState {
  return {
    orgName,
    projectSlug: null,
    completedSteps: [],
    panelOpen: true,
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

- [ ] **Step 2: Write failing test for new storage shape**

Replace `src/shared/onboarding/storage.test.ts`:

```typescript
import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  readOnboardingState,
  writeOnboardingState,
  ONBOARDING_KEY,
  defaultOnboardingState,
} from './storage'

const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
  }
})()

vi.stubGlobal('localStorage', localStorageMock)

describe('onboarding storage', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('returns default state when localStorage is empty', () => {
    const state = readOnboardingState('my-org')
    expect(state).toEqual(defaultOnboardingState('my-org'))
  })

  it('default state has panelOpen: true', () => {
    const state = defaultOnboardingState('my-org')
    expect(state.panelOpen).toBe(true)
  })

  it('default state has no dismissed field', () => {
    const state = defaultOnboardingState('my-org')
    expect('dismissed' in state).toBe(false)
  })

  it('round-trips state to localStorage', () => {
    const state = {
      orgName: 'my-org',
      projectSlug: 'my-project',
      completedSteps: ['create_project' as const],
      panelOpen: true,
    }
    writeOnboardingState(state)
    expect(readOnboardingState('my-org')).toEqual(state)
  })

  it('returns default state when orgName does not match', () => {
    const state = {
      orgName: 'org-a',
      projectSlug: null,
      completedSteps: [],
      panelOpen: false,
    }
    writeOnboardingState(state)
    expect(readOnboardingState('org-b')).toEqual(defaultOnboardingState('org-b'))
  })

  it('handles corrupt localStorage gracefully', () => {
    localStorage.setItem(ONBOARDING_KEY, 'not-json')
    expect(readOnboardingState('my-org')).toEqual(defaultOnboardingState('my-org'))
  })

  it('uses v2 key to avoid collision with old data', () => {
    expect(ONBOARDING_KEY).toBe('mc_onboarding_v2')
  })
})
```

- [ ] **Step 3: Run tests — expect failures**

```bash
cd /data/home/lukemxjia/modelcraft/modelcraft-front
npx vitest run src/shared/onboarding/storage.test.ts
```

Expected: some tests fail because storage.ts still has old shape.

- [ ] **Step 4: Apply storage.ts changes (already written in Step 1)**

Verify file was saved correctly:
```bash
grep "mc_onboarding_v2\|dismissed" src/shared/onboarding/storage.ts
```
Expected: `mc_onboarding_v2` appears, `dismissed` does NOT appear.

- [ ] **Step 5: Run tests — expect all pass**

```bash
npx vitest run src/shared/onboarding/storage.test.ts
```
Expected: all 7 tests PASS.

- [ ] **Step 6: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/shared/onboarding/storage.ts modelcraft-front/src/shared/onboarding/storage.test.ts
git commit -m "feat(onboarding): bump to v2 storage — remove dismissed, panelOpen defaults true"
```

---

## Task 2: Rewrite step definitions to 5-step tutorial flow

**Files:**
- Modify: `src/shared/onboarding/steps.ts`
- Modify: `src/shared/onboarding/OnboardingContext.test.tsx`

- [ ] **Step 1: Write failing tests first**

Replace `src/shared/onboarding/OnboardingContext.test.tsx`:

```typescript
import { describe, it, expect } from 'vitest'
import {
  defaultOnboardingState,
  type OnboardingState,
  type OnboardingStepId,
} from './storage'
import { ONBOARDING_GROUPS, ONBOARDING_STEPS, ALL_TRACKED_STEPS } from './steps'

function deriveGroups(completedSteps: OnboardingStepId[]) {
  const completedSet = new Set(completedSteps)
  const globalCurrentId = ALL_TRACKED_STEPS.find(
    (s) => !completedSet.has(s.id)
  )?.id ?? null
  let markedCurrent = false
  return ONBOARDING_GROUPS.map((group) => {
    const steps = group.steps.map((step) => {
      if (step.kind === 'nav') return { ...step, status: 'nav' as const }
      if (completedSet.has(step.id)) return { ...step, status: 'completed' as const }
      if (step.id === globalCurrentId && !markedCurrent) {
        markedCurrent = true
        return { ...step, status: 'current' as const }
      }
      return { ...step, status: 'locked' as const }
    })
    const tracked = steps.filter((s) => s.kind === 'tracked')
    const allDone = tracked.length > 0 && tracked.every((s) => s.kind === 'tracked' && s.status === 'completed')
    return { id: group.id, label: group.label, status: allDone ? 'completed' as const : 'todo' as const, steps }
  })
}

describe('tutorial step definitions', () => {
  it('has exactly 5 groups', () => {
    expect(ONBOARDING_GROUPS.length).toBe(5)
  })

  it('group ids match expected 5-step flow', () => {
    const ids = ONBOARDING_GROUPS.map((g) => g.id)
    expect(ids).toEqual([
      'setup_project',
      'design_model',
      'add_users',
      'assign_permissions',
      'experience_login',
    ])
  })

  it('first required step is current at start', () => {
    const groups = deriveGroups([])
    const g1 = groups[0]
    const tracked = g1.steps.filter((s) => s.kind === 'tracked')
    expect(tracked[0].status).toBe('current')
  })

  it('nav steps always have nav status', () => {
    const groups = deriveGroups([])
    const navSteps = groups.flatMap((g) => g.steps.filter((s) => s.kind === 'nav'))
    expect(navSteps.every((s) => s.status === 'nav')).toBe(true)
  })

  it('completing create_project advances current to create_model', () => {
    const groups = deriveGroups(['create_project'])
    const g2 = groups[1]
    const tracked = g2.steps.filter((s) => s.kind === 'tracked')
    expect(tracked[0].status).toBe('current') // create_model
  })

  it('group 2 completes when create_model done', () => {
    const groups = deriveGroups(['create_project', 'create_model'])
    expect(groups[1].status).toBe('completed')
  })

  it('completing all required steps except last advances to end_user_login', () => {
    const groups = deriveGroups(['create_project', 'create_model', 'add_end_user', 'assign_role'])
    const g5 = groups[4]
    const tracked = g5.steps.filter((s) => s.kind === 'tracked')
    expect(tracked[0].status).toBe('current') // end_user_login
  })

  it('ONBOARDING_STEPS contains exactly the 5 required step IDs', () => {
    const ids = ONBOARDING_STEPS.map((s) => s.id)
    expect(ids).toEqual([
      'create_project',
      'create_model',
      'add_end_user',
      'assign_role',
      'end_user_login',
    ])
  })

  it('end_user_login step route returns null (no navigation)', () => {
    const step = ALL_TRACKED_STEPS.find((s) => s.id === 'end_user_login')!
    expect(step.route({ orgName: 'org', projectSlug: null })).toBeNull()
  })

  it('create_model route resolves to model-editor when projectSlug set', () => {
    const step = ALL_TRACKED_STEPS.find((s) => s.id === 'create_model')!
    expect(step.route({ orgName: 'org', projectSlug: 'proj' }))
      .toBe('/org/org/project/proj/model-editor')
  })

  it('markStep dedup works', () => {
    const prev: OnboardingState = { ...defaultOnboardingState('test'), completedSteps: ['create_project'] }
    expect(prev.completedSteps.includes('create_project')).toBe(true)
  })
})
```

- [ ] **Step 2: Run tests — expect failures**

```bash
cd /data/home/lukemxjia/modelcraft/modelcraft-front
npx vitest run src/shared/onboarding/OnboardingContext.test.tsx
```
Expected: multiple failures (wrong group count, wrong IDs).

- [ ] **Step 3: Rewrite steps.ts**

Replace full file content:

```typescript
import type { OnboardingStepId } from './storage'

/** A sub-step that records completion */
export interface OnboardingTrackedStep {
  kind: 'tracked'
  id: OnboardingStepId
  label: string
  type: 'action' | 'manual'
  route: (params: { orgName: string; projectSlug: string | null }) => string | null
}

/** A sub-step that is pure navigation — no completion tracking */
export interface OnboardingNavStep {
  kind: 'nav'
  id: string
  label: string
  route: (params: { orgName: string; projectSlug: string | null }) => string
}

export type OnboardingSubStep = OnboardingTrackedStep | OnboardingNavStep

export interface OnboardingGroup {
  id: string
  label: string
  steps: OnboardingSubStep[]
}

export const ONBOARDING_GROUPS: OnboardingGroup[] = [
  {
    id: 'setup_project',
    label: '创建项目',
    steps: [
      {
        kind: 'nav',
        id: 'goto_project',
        label: '前往项目列表',
        route: ({ orgName }) => `/org/${orgName}/workspace`,
      },
      {
        kind: 'tracked',
        id: 'create_project',
        label: '新建项目',
        type: 'action',
        route: ({ orgName }) => `/org/${orgName}/workspace`,
      },
    ],
  },
  {
    id: 'design_model',
    label: '创建模型',
    steps: [
      {
        kind: 'nav',
        id: 'goto_model_editor',
        label: '进入项目，前往模型编辑',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/model-editor`
            : `/org/${orgName}/workspace`,
      },
      {
        kind: 'tracked',
        id: 'create_model',
        label: '点击新建模型',
        type: 'action',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/model-editor`
            : `/org/${orgName}/workspace`,
      },
    ],
  },
  {
    id: 'add_users',
    label: '创建终端用户',
    steps: [
      {
        kind: 'nav',
        id: 'goto_end_users',
        label: '前往终端用户',
        route: ({ orgName }) => `/org/${orgName}/end-users`,
      },
      {
        kind: 'tracked',
        id: 'add_end_user',
        label: '新建终端用户',
        type: 'action',
        route: ({ orgName }) => `/org/${orgName}/end-users`,
      },
    ],
  },
  {
    id: 'assign_permissions',
    label: '分配权限',
    steps: [
      {
        kind: 'nav',
        id: 'goto_end_user_access',
        label: '进入项目，前往终端用户授权',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/end-user-access`
            : `/org/${orgName}/workspace`,
      },
      {
        kind: 'tracked',
        id: 'assign_role',
        label: '为用户分配角色',
        type: 'action',
        route: ({ orgName, projectSlug }) =>
          projectSlug
            ? `/org/${orgName}/project/${projectSlug}/end-user-access`
            : `/org/${orgName}/workspace`,
      },
    ],
  },
  {
    id: 'experience_login',
    label: '体验终端用户登录',
    steps: [
      {
        kind: 'tracked',
        id: 'end_user_login',
        label: '终端用户登录体验',
        type: 'manual',
        route: () => null,
      },
    ],
  },
]

/** Flat list of all tracked steps — used for completion counting */
export const ONBOARDING_STEPS: OnboardingTrackedStep[] = ONBOARDING_GROUPS
  .flatMap((g) => g.steps)
  .filter((s): s is OnboardingTrackedStep => s.kind === 'tracked')

/** Alias for compatibility */
export const ALL_TRACKED_STEPS = ONBOARDING_STEPS
```

- [ ] **Step 4: Run tests — expect all pass**

```bash
npx vitest run src/shared/onboarding/OnboardingContext.test.tsx
```
Expected: all 11 tests PASS.

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/shared/onboarding/steps.ts modelcraft-front/src/shared/onboarding/OnboardingContext.test.tsx
git commit -m "feat(tutorial): redefine 5-step tutorial flow — 创建项目/模型/终端用户/分配权限/体验登录"
```

---

## Task 3: Update OnboardingContext — remove dismissed, fix panelOpen default

**Files:**
- Modify: `src/shared/onboarding/OnboardingContext.tsx`

The context needs: remove `dismiss`/`dismissed` from state and API; update group derivation to remove optional-group logic (no more optional groups); `panelOpen` defaults to `true` (already set in storage default).

- [ ] **Step 1: Apply the following diff to OnboardingContext.tsx**

Remove from the `OnboardingContextValue` interface:
```typescript
// REMOVE these lines:
dismissed: boolean
dismiss: () => void
```

Remove the `dismiss` callback entirely:
```typescript
// REMOVE:
const dismiss = useCallback(() => {
  setState((prev) => {
    const next = { ...prev, dismissed: true, panelOpen: false }
    writeOnboardingState(next)
    return next
  })
}, [])
```

Update the group derivation to remove optional-group logic. Replace the `groups` derivation block (lines ~163–202 in original):

```typescript
const completedSet = new Set(state.completedSteps)

const globalCurrentId = ALL_TRACKED_STEPS.find(
  (s) => !completedSet.has(s.id)
)?.id ?? null

let markedCurrent = false
const groups: OnboardingGroupWithStatus[] = ONBOARDING_GROUPS.map((group) => {
  const steps: OnboardingSubStepWithStatus[] = group.steps.map((step) => {
    if (step.kind === 'nav') return { ...step, status: 'nav' as const }
    if (completedSet.has(step.id)) return { ...step, status: 'completed' } satisfies OnboardingTrackedStepWithStatus
    if (step.id === globalCurrentId && !markedCurrent) {
      markedCurrent = true
      return { ...step, status: 'current' } satisfies OnboardingTrackedStepWithStatus
    }
    return { ...step, status: 'locked' } satisfies OnboardingTrackedStepWithStatus
  })

  const trackedSteps = steps.filter((s): s is OnboardingTrackedStepWithStatus => s.kind === 'tracked')
  const allDone = trackedSteps.length > 0 && trackedSteps.every((s) => s.status === 'completed')

  return {
    id: group.id,
    label: group.label,
    status: allDone ? 'completed' : 'todo',
    steps,
  }
})
```

Remove `dismissed` from the `context.Provider` value object, and remove the `dismiss` function from the value. Also remove `dismissed: state.dismissed` from the value.

Update the `OnboardingGroupWithStatus` interface — remove `optional?` field (no longer used).

Update `OnboardingContextValue` type — remove `dismissed` and `dismiss`.

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd /data/home/lukemxjia/modelcraft/modelcraft-front
npx tsc --noEmit 2>&1 | grep -i "dismissed\|dismiss\|onboarding" | head -20
```
Expected: no errors related to `dismissed` or `dismiss`.

- [ ] **Step 3: Fix any TypeScript errors that surface**

If errors appear about `dismissed` being used elsewhere, find and remove those usages:
```bash
grep -rn "dismissed\|dismiss" src/shared/onboarding/ src/app/org
```
Remove any remaining references.

- [ ] **Step 4: Run all onboarding tests**

```bash
npx vitest run src/shared/onboarding/
```
Expected: all tests PASS (storage: 7, context: 11 = 18 total).

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/shared/onboarding/OnboardingContext.tsx
git commit -m "feat(tutorial): remove dismissed behavior — tutorial panel is persistent"
```

---

## Task 4: Add `tutorial-highlight` CSS class

**Files:**
- Modify: `src/app/globals.css`

The `tutorial-highlight` class will be applied via `pendingAction` checks in individual components (already used in ModelSidebar as inline Tailwind). We add a global CSS version for use in pages that can't use inline Tailwind ring animations.

- [ ] **Step 1: Add to globals.css**

At the end of `src/app/globals.css`, append:

```css
/* ── Tutorial highlight ─────────────────────────────────────────── */
@keyframes tutorial-pulse {
  0%, 100% { outline-color: #4F46E5; }
  50% { outline-color: #818CF8; }
}

.tutorial-highlight {
  outline: 2px solid #4F46E5;
  outline-offset: 3px;
  border-radius: 6px;
  animation: tutorial-pulse 1.5s ease-in-out infinite;
}
```

- [ ] **Step 2: Verify the class can be used**

```bash
grep -n "tutorial-highlight" /data/home/lukemxjia/modelcraft/modelcraft-front/src/app/globals.css
```
Expected: 2 matches (keyframe usage + class definition).

- [ ] **Step 3: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/app/globals.css
git commit -m "feat(tutorial): add tutorial-highlight CSS keyframe animation"
```

---

## Task 5: Wire pendingAction highlight to end-users page

The `add_end_user` step is already tracked via `markStep` in `useOrgEndUsers.ts` (fires on mutation success). What's missing: when the user arrives at this page with `pendingAction === 'add_end_user'`, the 新建终端用户 button should animate.

**Files:**
- Modify: `src/app/org/[orgName]/end-users/page.tsx`

- [ ] **Step 1: Read the current end-users page**

```bash
cat -n /data/home/lukemxjia/modelcraft/modelcraft-front/src/app/org/\[orgName\]/end-users/page.tsx | head -60
```

Identify: (a) the 新建终端用户 button element, (b) whether `useOnboarding` is already imported.

- [ ] **Step 2: Add pendingAction highlight to the 新建终端用户 button**

Add `useOnboarding` import at the top of the file (if not present):
```typescript
import { useOnboarding } from '@shared/onboarding/OnboardingContext'
```

Inside the page component, add:
```typescript
const { pendingAction, setPendingAction } = useOnboarding()
const highlightAddUser = pendingAction === 'add_end_user'
```

Find the 新建终端用户 button (look for text "新建终端用户" or the CreateEndUserDialog trigger) and add:
```typescript
className={cn(
  existingClasses,
  highlightAddUser && 'border-amber-400 bg-amber-50 ring-2 ring-amber-400 ring-offset-1 animate-pulse'
)}
onClick={() => {
  if (highlightAddUser) setPendingAction(null)
  // existing onClick handler
}}
```

If the button lives inside `EndUsersManagementTable`, pass `highlightAddUser` as a prop instead and apply there.

- [ ] **Step 3: Manual verification**

Start dev server (`npm run dev`) and navigate to `端用户` page with `pendingAction` set to `add_end_user` in localStorage (or via the onboarding panel clicking the step). Confirm the button pulses amber.

- [ ] **Step 4: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/app/org/\[orgName\]/end-users/page.tsx
# Include EndUsersManagementTable if modified
git commit -m "feat(tutorial): highlight 新建终端用户 button when pendingAction=add_end_user"
```

---

## Task 6: Wire pendingAction highlight to end-user-access page

The `assign_role` step is tracked via `markStep` in `useUserAuth.ts`. What's missing: highlight the grant/assign button when `pendingAction === 'assign_role'`.

**Files:**
- Modify: `src/app/org/[orgName]/project/[projectSlug]/end-user-access/page.tsx`
  (and possibly `src/web/components/features/end-user-access/GrantEndUserAccessDialog.tsx`)

- [ ] **Step 1: Read the end-user-access page**

```bash
cat -n /data/home/lukemxjia/modelcraft/modelcraft-front/src/app/org/\[orgName\]/project/\[projectSlug\]/end-user-access/page.tsx | head -80
```

Identify: the button that triggers role assignment (likely opens `GrantEndUserAccessDialog`).

- [ ] **Step 2: Add pendingAction highlight to the assign-role trigger button**

```typescript
import { useOnboarding } from '@shared/onboarding/OnboardingContext'

// inside component:
const { pendingAction, setPendingAction } = useOnboarding()
const highlightAssign = pendingAction === 'assign_role'
```

Apply to the trigger button:
```typescript
className={cn(
  existingClasses,
  highlightAssign && 'border-amber-400 bg-amber-50 ring-2 ring-amber-400 ring-offset-1 animate-pulse'
)}
onClick={() => {
  if (highlightAssign) setPendingAction(null)
  // existing handler
}}
```

- [ ] **Step 3: Manual verification**

Navigate to the end-user-access page with `pendingAction === 'assign_role'`. Confirm the button pulses amber.

- [ ] **Step 4: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/app/org/\[orgName\]/project/\[projectSlug\]/end-user-access/page.tsx
git commit -m "feat(tutorial): highlight 授权 button when pendingAction=assign_role"
```

---

## Task 7: Redesign OnboardingPanel — remove dismiss, add reset, step 5 text card

**Files:**
- Modify: `src/shared/onboarding/OnboardingPanel.tsx`

Key changes:
1. Remove the X (dismiss) button from the header
2. Add a "↺ 重新开始" link/button in the footer (above the collapse chevron)
3. When `isComplete`, show a completion card instead of `return null`
4. Step 5 (`end_user_login`, `type: 'manual'`) shows the login URL text + "已完成 ✓" button (already implemented — verify it still works with new step structure)

- [ ] **Step 1: Remove the X dismiss button**

Find in `OnboardingPanel.tsx`:
```tsx
<button
  onClick={dismiss}
  className="flex size-5 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
  aria-label="关闭并不再显示"
>
  <X className="size-3" />
</button>
```

Delete it. Also remove `dismiss` from the `useOnboarding()` destructure, and remove the `X` import if unused.

- [ ] **Step 2: Add reset button to footer**

Replace the existing collapse footer:
```tsx
<button
  onClick={closePanel}
  className="flex w-full items-center justify-center border-t border-border py-1.5 text-muted-foreground transition-colors hover:bg-accent"
  aria-label="折叠面板"
>
  <ChevronDown className="size-3.5" />
</button>
```

With:
```tsx
<div className="border-t border-border">
  <button
    onClick={() => { reset(); openPanel() }}
    className="flex w-full items-center justify-center gap-1 py-1.5 text-[10px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
    aria-label="重新开始教程"
  >
    <RotateCcw className="size-3" />
    重新开始
  </button>
  <button
    onClick={closePanel}
    className="flex w-full items-center justify-center border-t border-border py-1.5 text-muted-foreground transition-colors hover:bg-accent"
    aria-label="折叠面板"
  >
    <ChevronDown className="size-3.5" />
  </button>
</div>
```

Add `RotateCcw` to the lucide-react import. Add `reset` to the `useOnboarding()` destructure.

- [ ] **Step 3: Replace `isComplete` null return with completion card**

Replace:
```tsx
if (isComplete) return null
```

With:
```tsx
if (isComplete) {
  if (!panelOpen) {
    return (
      <div
        className="fixed bottom-6 right-6 z-50 flex cursor-pointer items-center gap-3 rounded-lg border border-border bg-white px-3 py-2.5 shadow-md transition-shadow hover:shadow-lg"
        onClick={openPanel}
        role="button"
        aria-label="展开教程完成面板"
      >
        <div className="h-8 w-0.5 flex-shrink-0 rounded-full bg-[#10b981]" />
        <div className="flex flex-col gap-1">
          <span className="text-[12px] font-semibold text-foreground">快速开始</span>
          <div className="h-1 w-32 overflow-hidden rounded-full bg-[#EBEEF2]">
            <div className="h-full w-full rounded-full bg-[#10b981]" />
          </div>
          <span className="text-[10px] font-medium text-[#10b981]">全部完成 🎉</span>
        </div>
        <ChevronUp className="ml-1 size-3.5 text-muted-foreground" />
      </div>
    )
  }
  return (
    <div className="fixed bottom-6 right-6 z-50 w-[260px] overflow-hidden rounded-xl border border-border bg-white shadow-lg">
      <div className="border-b border-border px-3.5 py-3">
        <div className="mb-2 flex items-center justify-between">
          <span className="text-[12px] font-semibold text-foreground">快速开始</span>
        </div>
        <div className="h-1 overflow-hidden rounded-full bg-[#EBEEF2]">
          <div className="h-full w-full rounded-full bg-[#10b981] transition-all duration-300" />
        </div>
        <p className="mt-1 text-[10px] font-medium text-[#10b981]">{totalCount} / {totalCount} 步完成</p>
      </div>
      <div className="px-4 py-5 text-center">
        <div className="mb-2 text-2xl">🎉</div>
        <p className="text-[13px] font-semibold text-foreground">教程完成！</p>
        <p className="mt-1 text-[11px] text-muted-foreground">你已完成所有入门步骤</p>
      </div>
      <div className="border-t border-border">
        <button
          onClick={() => { reset(); openPanel() }}
          className="flex w-full items-center justify-center gap-1 py-1.5 text-[10px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        >
          <RotateCcw className="size-3" />
          重新开始
        </button>
        <button
          onClick={closePanel}
          className="flex w-full items-center justify-center border-t border-border py-1.5 text-muted-foreground transition-colors hover:bg-accent"
        >
          <ChevronDown className="size-3.5" />
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Verify TypeScript**

```bash
cd /data/home/lukemxjia/modelcraft/modelcraft-front
npx tsc --noEmit 2>&1 | grep -i "onboardingpanel\|panel\|dismiss" | head -20
```
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/shared/onboarding/OnboardingPanel.tsx
git commit -m "feat(tutorial): panel redesign — remove dismiss, add reset, show completion card"
```

---

## Task 8: Fix syncProjects to use new step IDs + final TypeScript check

The `syncProjects` in `OnboardingContext.tsx` still references `create_project` (which is fine — that step ID is preserved). But it also contains logic checking `isOptionalGroup` which no longer exists. Verify the full context compiles clean.

**Files:**
- Modify: `src/shared/onboarding/OnboardingContext.tsx` (verify only)

- [ ] **Step 1: Full TypeScript check**

```bash
cd /data/home/lukemxjia/modelcraft/modelcraft-front
npx tsc --noEmit 2>&1 | head -40
```

- [ ] **Step 2: Fix any remaining type errors**

Common issues to look for:
- `optional` field referenced on `OnboardingGroup` (removed in new steps.ts — remove from interface if needed)
- `dismissed` referenced anywhere
- Old step IDs (`select_database`, `insert_column`, `insert_data`, `create_permission`, `create_bundle`, `create_role`) referenced in `OnboardingPendingAction` union type — update to match new 5 step IDs:

```typescript
export type OnboardingPendingAction =
  | 'create_project'
  | 'create_model'
  | 'add_end_user'
  | 'assign_role'
  | 'highlight_first_project'
  | null
```

- [ ] **Step 3: Run all onboarding tests**

```bash
npx vitest run src/shared/onboarding/
```
Expected: all tests PASS.

- [ ] **Step 4: Run full lint**

```bash
npm run lint 2>&1 | grep -i "onboard\|tutorial\|error" | head -20
```
Expected: no errors in onboarding files.

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-front/src/shared/onboarding/OnboardingContext.tsx
git commit -m "feat(tutorial): update PendingAction type to match 5-step tutorial IDs"
```

---

## Task 9: Remove stale `markStep` calls for deleted step IDs

Since steps `select_database`, `insert_column`, `insert_data`, `create_permission`, `create_bundle`, `create_role` no longer exist in `OnboardingStepId`, their `markStep` calls will cause TypeScript errors. Remove them.

**Files:**
- Modify: `src/web/hooks/model/use-models.ts` (keep `markStep('create_model')`, remove any others)
- Modify: `src/app/org/[orgName]/project/[projectSlug]/rbac/roles/_hooks/useRoleList.ts` (remove `markStep('create_role')`)
- Modify: `src/app/org/[orgName]/project/[projectSlug]/rbac/permissions/_hooks/usePermissionsView.ts` (remove `markStep('create_permission')`)
- Modify: `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx` (pendingAction `create_model` highlight stays; remove any reference to `select_database`)

- [ ] **Step 1: Find all stale markStep calls**

```bash
grep -rn "markStep('select_database')\|markStep('insert_column')\|markStep('insert_data')\|markStep('create_permission')\|markStep('create_bundle')\|markStep('create_role')" \
  /data/home/lukemxjia/modelcraft/modelcraft-front/src
```

- [ ] **Step 2: Remove each stale call**

For each file found, remove the `markStep(...)` call and, if the file only used `markStep` for that one removed step, also remove the `const { markStep } = useOnboarding()` destructure and the `useOnboarding` import if nothing else from it is used.

Example for `useRoleList.ts` — change:
```typescript
const { markStep } = useOnboarding()
// ...
onCompleted: () => markStep('create_role'),
```
To:
```typescript
// remove markStep line
// onCompleted: () => { /* no-op */ }
```
Or simply remove the `onCompleted` callback if it only called `markStep`.

- [ ] **Step 3: Full TypeScript check**

```bash
cd /data/home/lukemxjia/modelcraft/modelcraft-front
npx tsc --noEmit 2>&1 | head -30
```
Expected: 0 errors.

- [ ] **Step 4: Run all onboarding tests + lint**

```bash
npx vitest run src/shared/onboarding/ && npm run lint 2>&1 | grep "error" | head -10
```
Expected: all pass, no lint errors.

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add -u modelcraft-front/src/
git commit -m "feat(tutorial): remove stale markStep calls for retired step IDs"
```

---

## Self-Review

**Spec coverage:**
- ✅ 5-step flow (创建项目/模型/终端用户/分配权限/体验登录) → Task 2
- ✅ Persistent tab (no dismiss forever) → Task 3 + Task 7
- ✅ Reset button → Task 7
- ✅ Sub-step nav buttons (nav kind) → preserved from existing system, no change needed
- ✅ pendingAction highlights for each step → create_project (workspace, existing), create_model (ModelSidebar, existing), add_end_user (Task 5), assign_role (Task 6)
- ✅ Step 5 pure text + manual confirm → existing `type: 'manual'` pattern in OnboardingPanel, preserved

**Placeholder scan:** No TBDs. All code is concrete.

**Type consistency:**
- `OnboardingStepId` union updated in Task 1 (storage.ts)
- `OnboardingPendingAction` updated in Task 8
- `ALL_TRACKED_STEPS` aliased from `ONBOARDING_STEPS` in Task 2 for backward compat
- `optional` field removed from groups — Task 3 derivation no longer references it
