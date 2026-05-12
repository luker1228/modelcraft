# Onboarding Checklist — Design Spec

**Date:** 2026-05-12  
**Status:** Approved  
**Scope:** Frontend only (no backend changes required)

---

## 1. Overview

A lightweight onboarding checklist for **tenant administrators** that guides them through the full path from project creation to end-user activation. The flow is **on-demand** (not auto-triggered), accessed via a sidebar entry point, and presented as a **floating progress panel** anchored to the bottom-right corner of the screen.

**Design north star:** Stripe Dashboard — Precision Tool. White surface, Indigo as the sole action color, structural borders, no decorative chrome.

---

## 2. User & Trigger

| Property | Value |
|----------|-------|
| Target user | Tenant administrator (Developer/Admin role) |
| Trigger | On-demand — user clicks "快速开始" in the Org sidebar |
| Scope | Org level, spanning into a single bound Project |
| Persistence | `localStorage` — no backend changes |

The "快速开始" entry point renders as a ghost button at the bottom of the Org sidebar, conditionally shown only when onboarding is incomplete and not dismissed.

---

## 3. The 8 Steps

| # | Step ID | Label | Layer | Completion Trigger |
|---|---------|-------|-------|--------------------|
| 1 | `create_project` | 创建项目 | Org | `createProject` mutation `onCompleted` |
| 2 | `create_model` | 创建模型 | Project | `createModel` mutation `onCompleted` |
| 3 | `add_field` | 添加字段 | Project | `createField` mutation `onCompleted` |
| 4 | `apply_preset` | 应用权限预设 | Project | `applyEndUserPresetPolicy` `onCompleted` |
| 5 | `create_role` | 创建角色 | Project | `createEndUserRole` `onCompleted` |
| 6 | `add_end_user` | 添加终端用户 | Org | `createEndUser` `onCompleted` |
| 7 | `assign_role` | 分配角色 | Org | `assignEndUserRole` `onCompleted` |
| 8 | `end_user_login` | 终端用户登录 | Org | Manual confirmation button in panel |

**Step 1 binds the flow to a specific project.** Once `create_project` completes, the `projectSlug` is stored in localStorage. Steps 2–5 are locked to that project until all 8 steps are finished (or the user resets).

Step 8 cannot be auto-detected (it happens in a separate browser session). The panel shows the end-user login URL and a manual "已完成 ✓" confirm button.

---

## 4. UI Design

### Floating Panel — Collapsed State

A white card anchored `fixed bottom-6 right-6 z-50`, rendered as a compact capsule:

- Left 3px Indigo (`#4F46E5`) vertical bar — signals "active"
- Title: "快速开始" (Inter 600, 12px, `#1A1F36`)
- Progress bar: `#EBEEF2` track, `#4F46E5` fill, 4px height
- Subtitle: "X / 8 步完成" (Inter 500, 10px, `#697386`)
- Chevron toggle icon (right side)
- Shadow: `shadow-md` (`0 4px 6px -1px rgba(0,0,0,0.1)`)
- Border: `1px solid #E3E8EE`
- Background: `#ffffff`

### Floating Panel — Expanded State

Expands upward from the collapsed capsule (slide-in-right, 250ms ease-out):

**Header:** Title + close (✕) button + progress bar + "X / 8 步完成"

**Steps list:**
- **Completed:** Green check circle (`rgba(16,185,129,0.1)` bg, `#10b981` icon), strikethrough label in `#8792A2`
- **Current:** Indigo selected-state row — `rgba(79,70,229,0.06)` background, `3px solid #4F46E5` left border, Indigo circle with step number, bold Indigo label + `→` arrow
- **Locked:** `#E3E8EE` border circle with step number, muted label `#8792A2`

**CTA footer:** Full-width primary button (`#4F46E5`), label: "前往：[当前步骤名] →". Navigates directly to the relevant route.

**Panel width:** 232px  
**Shadow:** `shadow-lg` (`0 10px 15px -3px rgba(0,0,0,0.1)`)

### Visual Reference

Mockups saved at:
```
.superpowers/brainstorm/3128168-1778584304/content/panel-v2.html
```

---

## 5. State Schema (localStorage)

**Key:** `mc_onboarding_v1`

```ts
interface OnboardingState {
  orgName: string;
  projectSlug: string | null;      // null until step 1 completes
  completedSteps: OnboardingStepId[];
  dismissed: boolean;              // true = hide sidebar entry point
  panelOpen: boolean;              // tracks collapsed/expanded
}

type OnboardingStepId =
  | 'create_project'
  | 'create_model'
  | 'add_field'
  | 'apply_preset'
  | 'create_role'
  | 'add_end_user'
  | 'assign_role'
  | 'end_user_login';
```

All reads and writes go through `storage.ts`. No direct `localStorage` calls outside that module.

---

## 6. Component Architecture

### File Structure

```
src/shared/onboarding/
├── OnboardingContext.tsx    // React Context + Provider + useOnboarding() hook
├── OnboardingPanel.tsx      // Floating panel UI (collapsed + expanded)
├── steps.ts                 // Step definitions: id, label, route, description
└── storage.ts               // localStorage read/write, typed
```

### Provider Mounting

Mount in the Org-level layout so the panel is visible across all Org and Project pages:

```tsx
// src/app/(tenant)/org/[orgName]/layout.tsx
<OnboardingProvider orgName={orgName}>
  {children}
  <OnboardingPanel />   {/* fixed bottom-6 right-6 z-50 */}
</OnboardingProvider>
```

### `useOnboarding()` Hook API

```ts
interface UseOnboardingReturn {
  steps: OnboardingStep[];          // ordered step definitions with status
  currentStep: OnboardingStep;      // first incomplete step
  completedCount: number;
  totalCount: number;               // 8
  isComplete: boolean;
  panelOpen: boolean;
  markStep: (id: OnboardingStepId) => void;
  openPanel: () => void;
  closePanel: () => void;
  dismiss: () => void;              // hide entry point permanently
  reset: () => void;                // dev/debug only
}
```

### Step Completion — Mutation Integration

In each relevant page/component, add a single `markStep` call in the mutation's `onCompleted` callback. No changes to the component's structure or rendering logic:

```tsx
// Example: model editor page
const { markStep } = useOnboarding();

const [createModel] = useCreateModelMutation({
  onCompleted: () => {
    markStep('create_model');
    // existing logic...
  },
});
```

Pages requiring `markStep` integration:

| Mutation | File location (approximate) |
|----------|-----------------------------|
| `createProject` | Org projects page / create project dialog |
| `createModel` | Project model editor |
| `createField` | Project field creation |
| `applyEndUserPresetPolicy` | Project RBAC permissions page |
| `createEndUserRole` | Project RBAC roles page |
| `createEndUser` | Org end-users page |
| `assignEndUserRole` | Org end-users page |

---

## 7. Sidebar Entry Point

Add to the bottom of the Org sidebar (above the user avatar area), conditionally rendered:

```tsx
{!onboarding.isComplete && !onboarding.dismissed && (
  <button
    className="ghost-button w-full"
    onClick={onboarding.openPanel}
  >
    快速开始
    <span className="badge-default ml-auto">
      {onboarding.completedCount}/{onboarding.totalCount}
    </span>
  </button>
)}
```

---

## 8. Step Routes (CTA Navigation)

| Step | Route |
|------|-------|
| 创建项目 | `/org/[orgName]` (projects list) |
| 创建模型 | `/org/[orgName]/project/[projectSlug]/model-editor` |
| 添加字段 | `/org/[orgName]/project/[projectSlug]/model-editor` |
| 应用权限预设 | `/org/[orgName]/project/[projectSlug]/rbac/permissions` |
| 创建角色 | `/org/[orgName]/project/[projectSlug]/rbac/roles` |
| 添加终端用户 | `/org/[orgName]/end-users` |
| 分配角色 | `/org/[orgName]/end-users` |
| 终端用户登录 | Panel shows login URL: `/end-user/org/[orgName]/login` |

---

## 9. Out of Scope

- No backend changes — no new DB fields, no new GraphQL operations
- No analytics/telemetry on step completion
- No multi-project onboarding (flow binds to one project)
- No dark mode variant (follows system default)
- No mobile-specific layout (panel is desktop-only)
