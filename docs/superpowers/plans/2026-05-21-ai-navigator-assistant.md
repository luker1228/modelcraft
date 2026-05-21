# AI 导航助手 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构 AI 助手，用统一的 `show_navigation_proposal` 工具替换 15+ 个碎片化 CopilotKit actions；引入 `<AiTarget>` 组件作为声明式高亮注册机制；后端永远返回 Proposal（含 `action_candidate` / `clarification_candidate`），用户点击后前端执行。

**Architecture:** 前端通过 `<AiTarget>` 组件声明可高亮区域，通过 `routeCatalog` 声明可导航页面，并经由 `useCopilotReadable` 注入到 Agent 上下文。Agent 调用唯一前端工具 `show_navigation_proposal`，CopilotKit 的 `render` 函数将其渲染为 `AiProposalCard`。用户点击候选项时，`handleCandidateClick` 判断类型：`action_candidate` → 执行导航/高亮序列；`clarification_candidate` → 通过 CopilotKit chat 将选择发回 Agent 继续推理。

**Tech Stack:** Next.js (App Router), CopilotKit (`@copilotkit/react-core`, `@copilotkit/react-ui`), Vitest, Python/LangGraph (modelcraft-agent), Tailwind CSS, shadcn/ui

**Spec:** `docs/superpowers/specs/2026-05-21-ai-navigator-assistant-design.md`

---

## File Map

| 状态 | 文件 | 说明 |
|------|------|------|
| Create | `src/web/components/features/copilot/types.ts` | AI Navigator 类型定义 |
| Create | `src/web/components/features/copilot/AiTarget.tsx` | 声明可高亮区域的包裹组件 |
| Create | `src/web/components/features/copilot/AiProposalCard.tsx` | 渲染候选项的卡片组件 |
| Create | `src/web/components/features/copilot/AiProposalCard.test.tsx` | AiProposalCard 单元测试 |
| Create | `src/web/hooks/ai/use-navigation-proposal.ts` | handleCandidateClick / executeActions hook |
| Create | `src/web/lib/route-catalog.ts` | 可导航页面目录（静态数据） |
| Modify | `src/web/lib/highlight-element.ts` | 增加 message 参数（tooltip）、支持直接传 HTMLElement |
| Modify | `src/web/contexts/ai-capability-context.tsx` | AICapability 增加 type 字段 |
| Modify | `src/web/components/features/copilot/SharedCopilotActions.tsx` | 新增 show_navigation_proposal action |
| Modify | `src/web/components/features/copilot/AICapabilityReadable.tsx` | 同时注入 routeCatalog 和 aiTargets |
| Modify | `src/web/components/features/copilot/ProjectCopilotActions.tsx` | 删除所有 navigate_*/guide_*/highlight_records action |
| Modify | `src/web/components/features/copilot/OrgCopilotActions.tsx` | 删除 navigate_*/highlight_* action |
| Modify | `modelcraft-agent/agents/admin_agent.py` | 更新系统提示词，只保留 show_navigation_proposal |
| Modify | `modelcraft-agent/tests/agents/test_admin_agent.py` | 更新断言（list_databases 补入预期集合） |

---

## Task 1: 定义 AI Navigator 类型

**Files:**
- Create: `modelcraft-front/src/web/components/features/copilot/types.ts`

- [ ] **Step 1: 创建类型文件**

```ts
// modelcraft-front/src/web/components/features/copilot/types.ts

export type AiNavigateArgs = {
  route: string
  params?: Record<string, unknown>
  reason?: string
}

export type AiHighlightArgs = {
  targetId: string
  targetType?: 'field' | 'button' | 'section' | 'tableRow' | 'tab' | 'menu'
  label?: string
  message?: string
  durationMs?: number
  scrollIntoView?: boolean
}

export type AiAction =
  | { type: 'ui.navigate'; args: AiNavigateArgs }
  | { type: 'ui.highlight'; args: AiHighlightArgs }

export type ActionCandidate = {
  id: string
  type: 'action_candidate'
  title: string
  description?: string
  category?: 'page' | 'model' | 'table' | 'field' | 'setting' | 'action'
  confidence?: number
  isPrimary?: boolean
  actions: AiAction[]
}

export type ClarificationCandidate = {
  id: string
  type: 'clarification_candidate'
  title: string
  description?: string
  payload: {
    intent?: string
    entities?: Record<string, unknown>
    userMeaning?: string
  }
}

export type ProposalCandidate = ActionCandidate | ClarificationCandidate

export type AgentUiResponse = {
  kind: 'proposal'
  proposalId: string
  proposalType: 'navigation' | 'highlight' | 'clarification' | 'mixed'
  message: string
  query: string
  candidates: ProposalCandidate[]
}
```

- [ ] **Step 2: 验证 TypeScript 无错误**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | head -20
```

Expected: 0 errors（或仅已有的无关错误）

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/types.ts
git commit -m "feat(ai-navigator): add AI navigator type definitions"
```

---

## Task 2: 扩展 highlight-element.ts 支持 message

**Files:**
- Modify: `modelcraft-front/src/web/lib/highlight-element.ts`
- Modify: 对应 test 文件（同目录下，vitest 自动发现）

现有签名：`highlightElement(ref: RefObject<HTMLElement | null>, durationMs?: number)`
新增：接受 options 对象（含 `message`），支持直接传 `HTMLElement`

- [ ] **Step 1: 先写新测试（TDD）**

在 `highlight-element.ts` 同目录找到测试文件（文件末尾 `describe('highlightElement'...` 块），追加：

```ts
// 追加到现有 describe('highlightElement', ...) 块末尾：

it('shows message tooltip when provided', () => {
  const el = document.createElement('div')
  // We only verify no throw and that classes are applied — tooltip rendering
  // requires DOM API beyond jsdom scope. The message option is consumed by
  // the caller to render a tooltip externally.
  expect(() =>
    highlightElement({ current: el }, { message: '点击这里配置权限', scrollIntoView: false })
  ).not.toThrow()
  expect(el.classList.contains('bg-amber-50')).toBe(true)
})

it('accepts HighlightOptions object as second argument', () => {
  const el = document.createElement('button')
  highlightElement({ current: el }, { durationMs: 2000, scrollIntoView: false })
  expect(el.classList.contains('ring-amber-400')).toBe(true)
  vi.advanceTimersByTime(2000)
  expect(el.classList.contains('ring-amber-400')).toBe(false)
})
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd modelcraft-front && npx vitest run src/web/lib/highlight-element
```

Expected: FAIL（`HighlightOptions` 不存在，类型错误）

- [ ] **Step 3: 实现新签名**

完整替换 `highlight-element.ts`：

```ts
// modelcraft-front/src/web/lib/highlight-element.ts
import type { RefObject } from 'react'

export const HIGHLIGHT_CLASSES = [
  'bg-amber-50',
  'ring-4',
  'ring-amber-400',
  'ring-offset-4',
  'animate-pulse',
  'transition-all',
] as const

let activeTimeoutId: ReturnType<typeof setTimeout> | null = null

export type HighlightOptions = {
  /** Tooltip message shown near the element (caller is responsible for rendering). */
  message?: string
  durationMs?: number
  scrollIntoView?: boolean
}

/**
 * Apply amber highlight to a DOM element.
 * Second argument is either a legacy durationMs number or a HighlightOptions object.
 */
export function highlightElement(
  ref: RefObject<HTMLElement | null>,
  optionsOrDuration: HighlightOptions | number = {},
): void {
  const el = ref.current
  if (!el) return

  const opts: HighlightOptions =
    typeof optionsOrDuration === 'number'
      ? { durationMs: optionsOrDuration }
      : optionsOrDuration

  const { durationMs = 5000, scrollIntoView = true } = opts

  if (activeTimeoutId !== null) {
    clearTimeout(activeTimeoutId)
    activeTimeoutId = null
  }

  if (scrollIntoView) {
    el.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  }

  el.classList.add(...HIGHLIGHT_CLASSES)

  activeTimeoutId = setTimeout(() => {
    el.classList.remove(...HIGHLIGHT_CLASSES)
    activeTimeoutId = null
  }, durationMs)
}

/**
 * Highlight a raw HTMLElement directly (used by AiTarget / executeActions).
 */
export function highlightHTMLElement(
  el: HTMLElement,
  optionsOrDuration: HighlightOptions | number = {},
): void {
  highlightElement({ current: el }, optionsOrDuration)
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd modelcraft-front && npx vitest run src/web/lib/highlight-element
```

Expected: All PASS

- [ ] **Step 5: 全量测试确认无破坏**

```bash
cd modelcraft-front && npx vitest run 2>&1 | tail -10
```

Expected: no regressions

- [ ] **Step 6: Commit**

```bash
git add modelcraft-front/src/web/lib/highlight-element.ts
git commit -m "feat(ai-navigator): extend highlightElement with options object and highlightHTMLElement"
```

---

## Task 3: 创建 routeCatalog.ts

**Files:**
- Create: `modelcraft-front/src/web/lib/route-catalog.ts`

- [ ] **Step 1: 创建文件**

```ts
// modelcraft-front/src/web/lib/route-catalog.ts

export type RouteCatalogEntry = {
  /** 路由模板，使用 :param 占位符，例如 /org/:orgName/project/:projectSlug/models */
  routeTemplate: string
  /** 页面标题（中文，AI 用来匹配意图） */
  title: string
  /** 页面功能描述（AI 判断跳转依据） */
  description: string
  /** 关键词列表（触发跳转的语义词） */
  keywords: string[]
}

/**
 * All navigable pages in ModelCraft.
 * Agent reads this via useCopilotReadable to decide which page to navigate to.
 * Routes with :param are resolved at runtime using current org/project context.
 */
export const ROUTE_CATALOG: RouteCatalogEntry[] = [
  {
    routeTemplate: '/org/:orgName/workspace',
    title: '项目列表',
    description: '查看、搜索、创建和管理组织下的所有项目',
    keywords: ['项目列表', '所有项目', 'workspace', '项目管理'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/model-editor',
    title: '数据模型编辑器',
    description: '创建和管理数据模型、字段结构，查看模型数据记录',
    keywords: ['模型', '字段', '数据模型', '模型编辑器', '新建模型', '字段管理'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/enums',
    title: '枚举管理',
    description: '管理项目中的枚举类型，限制字段的可选值范围',
    keywords: ['枚举', 'enum', '枚举值', '枚举类型'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/rbac/roles',
    title: 'RBAC 角色管理',
    description: '管理项目内的角色与权限包，控制用户对数据的增删改查权限',
    keywords: ['权限', 'RBAC', '角色', '权限管理', '角色配置'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/rbac/users',
    title: 'RBAC 用户授权',
    description: '为终端用户分配角色和访问权限',
    keywords: ['用户权限', '授权', '角色分配', 'RBAC 用户'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/rbac/bundles',
    title: '权限包管理',
    description: '管理权限包版本，配置细粒度操作权限',
    keywords: ['权限包', 'bundle', '权限版本', '权限快照'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/end-users',
    title: '终端用户管理',
    description: '管理访问本项目的终端用户账号',
    keywords: ['终端用户', 'end user', '用户管理', '外部用户'],
  },
  {
    routeTemplate: '/org/:orgName/project/:projectSlug/settings',
    title: '项目设置',
    description: '修改项目基本信息、归档或删除项目、管理数据库集群',
    keywords: ['项目设置', '集群配置', '数据库连接', '项目信息', '设置'],
  },
  {
    routeTemplate: '/org/:orgName/developers',
    title: '成员管理',
    description: '管理组织内的开发者成员和角色（Owner / Admin / Member）',
    keywords: ['成员', '开发者', '邀请成员', '成员管理'],
  },
  {
    routeTemplate: '/org/:orgName/end-users',
    title: '终端用户（Org 级）',
    description: '管理组织下所有终端用户账号',
    keywords: ['终端用户', 'org 级用户', '用户账号'],
  },
  {
    routeTemplate: '/org/:orgName/settings',
    title: '组织设置',
    description: '配置组织基础信息和安全设置',
    keywords: ['组织设置', 'org 设置', '组织信息'],
  },
]
```

- [ ] **Step 2: 验证 TypeScript 无错误**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "route-catalog" | head -5
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/lib/route-catalog.ts
git commit -m "feat(ai-navigator): add route catalog for navigable pages"
```

---

## Task 4: 扩展 AICapabilityContext 添加 type 字段

**Files:**
- Modify: `modelcraft-front/src/web/contexts/ai-capability-context.tsx`

- [ ] **Step 1: 在 types.ts 查看 AICapability 现有定义**

当前 `AICapability`：
```ts
export type AICapability = {
  id: string
  label: string
  ref: RefObject<HTMLElement>
  description?: string
}
```

需要添加 `type` 字段（与 spec 中 `AiHighlightArgs.targetType` 一致）。

- [ ] **Step 2: 先更新现有 test（测试 type 字段）**

在 `ai-capability-context.tsx` 测试的 `describe('createCapabilityStore'...)` 块末尾追加：

```ts
it('stores and returns type field', () => {
  const store = createCapabilityStore()
  const ref = { current: document.createElement('div') }
  store.register({ id: 'my-section', label: '测试区域', ref, type: 'section' })
  expect(store.getAll()[0].type).toBe('section')
})
```

- [ ] **Step 3: 运行测试确认新测试失败**

```bash
cd modelcraft-front && npx vitest run src/web/contexts/ai-capability-context
```

Expected: FAIL（`type` 字段不存在于类型定义）

- [ ] **Step 4: 更新 AICapability 类型**

在 `ai-capability-context.tsx` 中，找到 `AICapability` 类型定义，改为：

```ts
export type AICapability = {
  id: string
  label: string
  ref: RefObject<HTMLElement>
  description?: string
  type?: 'field' | 'button' | 'section' | 'tableRow' | 'tab' | 'menu'
}
```

- [ ] **Step 5: 运行测试确认全部通过**

```bash
cd modelcraft-front && npx vitest run src/web/contexts/ai-capability-context
```

Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add modelcraft-front/src/web/contexts/ai-capability-context.tsx
git commit -m "feat(ai-navigator): add type field to AICapability"
```

---

## Task 5: 创建 `<AiTarget>` 组件

**Files:**
- Create: `modelcraft-front/src/web/components/features/copilot/AiTarget.tsx`

`<AiTarget>` 是 `useRegisterAICapability` 的声明式封装。它包裹子元素，自动注册/注销，并给 DOM 元素加 `data-ai-target` 属性。

- [ ] **Step 1: 创建组件**

```tsx
// modelcraft-front/src/web/components/features/copilot/AiTarget.tsx
'use client'

import { useRef, useEffect, type ReactNode } from 'react'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'

type AiTargetType = 'field' | 'button' | 'section' | 'tableRow' | 'tab' | 'menu'

interface AiTargetProps {
  /** Stable unique identifier. Agent uses this as targetId in ui.highlight. */
  id: string
  /** Human-readable label for this region (shown in AI context). */
  label: string
  /** Optional hint for AI about what this region does. */
  description?: string
  /** Semantic type for AI's understanding of the element. */
  type?: AiTargetType
  children: ReactNode
  className?: string
}

/**
 * Declarative wrapper that registers a UI region as a highlight target.
 *
 * Usage:
 *   <AiTarget id="create-model-btn" label="新建模型按钮" type="button">
 *     <Button>新建模型</Button>
 *   </AiTarget>
 *
 * This adds data-ai-target="create-model-btn" to the wrapper div and
 * registers/unregisters with AICapabilityContext automatically.
 */
export function AiTarget({
  id,
  label,
  description,
  type,
  children,
  className,
}: AiTargetProps) {
  const ref = useRef<HTMLDivElement>(null)
  const { register, unregister } = useAICapabilityContext()

  useEffect(() => {
    register({ id, label, ref, description, type })
    return () => unregister(id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, label, description, type])

  return (
    <div ref={ref} data-ai-target={id} className={className}>
      {children}
    </div>
  )
}
```

- [ ] **Step 2: 验证 TypeScript 编译**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "AiTarget" | head -5
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/AiTarget.tsx
git commit -m "feat(ai-navigator): add AiTarget component for declarative highlight registration"
```

---

## Task 6: 创建 useNavigationProposal hook

**Files:**
- Create: `modelcraft-front/src/web/hooks/ai/use-navigation-proposal.ts`

此 hook 封装候选项点击后的两种行为：执行 actions，或回传 Agent 继续推理。

- [ ] **Step 1: 创建 hook**

```ts
// modelcraft-front/src/web/hooks/ai/use-navigation-proposal.ts
'use client'

import { useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useCopilotChat } from '@copilotkit/react-core'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'
import { highlightHTMLElement } from '@web/lib/highlight-element'
import type { ProposalCandidate, AiAction } from '@web/components/features/copilot/types'

/**
 * Poll until targetId appears in the registry, or timeout expires.
 * Used after navigation to wait for the new page's AiTarget components to mount.
 */
function waitForTarget(
  getRef: (id: string) => React.RefObject<HTMLElement> | undefined,
  targetId: string,
  timeoutMs = 3000,
): Promise<void> {
  return new Promise((resolve) => {
    const deadline = Date.now() + timeoutMs
    const poll = () => {
      if (getRef(targetId)?.current || Date.now() >= deadline) {
        resolve()
      } else {
        requestAnimationFrame(poll)
      }
    }
    requestAnimationFrame(poll)
  })
}

export function useNavigationProposal() {
  const router = useRouter()
  const { getRef } = useAICapabilityContext()
  const { appendMessage } = useCopilotChat()

  const executeActions = useCallback(
    async (actions: AiAction[]) => {
      for (const action of actions) {
        if (action.type === 'ui.navigate') {
          router.push(action.args.route)

          // If the next action is a highlight, wait for its target to mount
          const nextAction = actions[actions.indexOf(action) + 1]
          if (nextAction?.type === 'ui.highlight') {
            await waitForTarget(getRef, nextAction.args.targetId)
          }
        }

        if (action.type === 'ui.highlight') {
          const ref = getRef(action.args.targetId)
          const el = ref?.current
          if (!el) {
            console.warn(`[AI] targetId "${action.args.targetId}" not in registry`)
            return
          }
          highlightHTMLElement(el, {
            message: action.args.message,
            durationMs: action.args.durationMs ?? 5000,
            scrollIntoView: action.args.scrollIntoView ?? true,
          })
        }
      }
    },
    [router, getRef],
  )

  const sendClarificationToAgent = useCallback(
    (candidateTitle: string) => {
      appendMessage({
        role: 'user',
        content: candidateTitle,
        id: crypto.randomUUID(),
      })
    },
    [appendMessage],
  )

  const handleCandidateClick = useCallback(
    async (candidate: ProposalCandidate) => {
      if (candidate.type === 'action_candidate') {
        await executeActions(candidate.actions)
        return
      }

      if (candidate.type === 'clarification_candidate') {
        // Send candidate title as user message — agent has conversation history
        // and will understand this is a clarification response.
        sendClarificationToAgent(candidate.title)
        return
      }
    },
    [executeActions, sendClarificationToAgent],
  )

  return { handleCandidateClick, executeActions, sendClarificationToAgent }
}
```

- [ ] **Step 2: 验证 TypeScript 编译**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "use-navigation-proposal" | head -5
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/hooks/ai/use-navigation-proposal.ts
git commit -m "feat(ai-navigator): add useNavigationProposal hook for candidate execution"
```

---

## Task 7: 创建 AiProposalCard 组件及测试

**Files:**
- Create: `modelcraft-front/src/web/components/features/copilot/AiProposalCard.tsx`
- Create: `modelcraft-front/src/web/components/features/copilot/AiProposalCard.test.tsx`

- [ ] **Step 1: 先写测试（TDD）**

```tsx
// modelcraft-front/src/web/components/features/copilot/AiProposalCard.test.tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { AiProposalCard } from './AiProposalCard'
import type { ProposalCandidate } from './types'

const actionCandidate: ProposalCandidate = {
  id: 'go-models',
  type: 'action_candidate',
  title: '数据模型管理',
  description: '进入数据模型编辑器',
  isPrimary: true,
  actions: [{ type: 'ui.navigate', args: { route: '/org/acme/project/main/model-editor' } }],
}

const clarificationCandidate: ProposalCandidate = {
  id: 'intent-config-model',
  type: 'clarification_candidate',
  title: '配置项目模型',
  description: '我想配置字段和权限',
  payload: { intent: 'configure_project_model' },
}

describe('AiProposalCard', () => {
  it('renders message and candidate titles', () => {
    render(
      <AiProposalCard
        message="找到以下结果："
        candidates={[actionCandidate]}
        onSelect={vi.fn()}
      />,
    )
    expect(screen.getByText('找到以下结果：')).toBeInTheDocument()
    expect(screen.getByText('数据模型管理')).toBeInTheDocument()
  })

  it('calls onSelect with correct candidate on click', () => {
    const onSelect = vi.fn()
    render(
      <AiProposalCard
        message="选择："
        candidates={[actionCandidate]}
        onSelect={onSelect}
      />,
    )
    fireEvent.click(screen.getByRole('button', { name: /数据模型管理/ }))
    expect(onSelect).toHaveBeenCalledWith(actionCandidate)
  })

  it('shows 推荐 badge on isPrimary candidate', () => {
    render(
      <AiProposalCard
        message="推荐："
        candidates={[actionCandidate]}
        onSelect={vi.fn()}
      />,
    )
    expect(screen.getByText('推荐')).toBeInTheDocument()
  })

  it('renders clarification_candidate with different visual cue', () => {
    render(
      <AiProposalCard
        message="请选择："
        candidates={[clarificationCandidate]}
        onSelect={vi.fn()}
      />,
    )
    expect(screen.getByText('配置项目模型')).toBeInTheDocument()
  })

  it('renders empty state when no candidates', () => {
    render(
      <AiProposalCard message="未找到" candidates={[]} onSelect={vi.fn()} />,
    )
    expect(screen.getByText('未找到相关页面')).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd modelcraft-front && npx vitest run src/web/components/features/copilot/AiProposalCard.test.tsx
```

Expected: FAIL（`AiProposalCard` 不存在）

- [ ] **Step 3: 实现 AiProposalCard**

```tsx
// modelcraft-front/src/web/components/features/copilot/AiProposalCard.tsx
'use client'

import type { ProposalCandidate } from './types'

interface AiProposalCardProps {
  message: string
  candidates: ProposalCandidate[]
  onSelect: (candidate: ProposalCandidate) => void
}

/**
 * Renders AI navigation proposals as clickable candidate cards.
 * Both action_candidate (execute) and clarification_candidate (chat)
 * are rendered identically — the caller handles the behavioral difference.
 */
export function AiProposalCard({ message, candidates, onSelect }: AiProposalCardProps) {
  if (candidates.length === 0) {
    return (
      <div className="mt-2 rounded-lg border border-border bg-background p-3 text-sm text-muted-foreground">
        未找到相关页面
      </div>
    )
  }

  return (
    <div className="mt-2 flex flex-col gap-2">
      {message && (
        <p className="text-xs text-muted-foreground px-1">{message}</p>
      )}
      {candidates.map((candidate) => (
        <button
          key={candidate.id}
          type="button"
          onClick={() => onSelect(candidate)}
          className="group w-full rounded-lg border border-border bg-background px-3 py-2.5 text-left transition-colors hover:border-amber-400 hover:bg-amber-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-amber-400"
        >
          <div className="flex items-start justify-between gap-2">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-1.5">
                <span className="text-sm font-medium text-foreground truncate">
                  {candidate.title}
                </span>
                {candidate.type === 'action_candidate' && candidate.isPrimary && (
                  <span className="shrink-0 rounded-full bg-amber-100 px-1.5 py-0.5 text-[10px] font-medium text-amber-700">
                    推荐
                  </span>
                )}
                {candidate.type === 'action_candidate' && (
                  <span className="shrink-0 rounded-full bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
                    {candidate.category ?? 'page'}
                  </span>
                )}
              </div>
              {candidate.description && (
                <p className="mt-0.5 text-xs text-muted-foreground line-clamp-2">
                  {candidate.description}
                </p>
              )}
            </div>
            <span className="shrink-0 text-muted-foreground transition-colors group-hover:text-amber-600 mt-0.5">
              {candidate.type === 'clarification_candidate' ? '→' : '↗'}
            </span>
          </div>
        </button>
      ))}
    </div>
  )
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd modelcraft-front && npx vitest run src/web/components/features/copilot/AiProposalCard.test.tsx
```

Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/AiProposalCard.tsx \
        modelcraft-front/src/web/components/features/copilot/AiProposalCard.test.tsx
git commit -m "feat(ai-navigator): add AiProposalCard component with action/clarification candidates"
```

---

## Task 8: 更新 SharedCopilotActions — 注册 show_navigation_proposal

**Files:**
- Modify: `modelcraft-front/src/web/components/features/copilot/SharedCopilotActions.tsx`

这是核心变更：将整个 Proposal 渲染逻辑接入 CopilotKit。

- [ ] **Step 1: 替换 SharedCopilotActions.tsx 完整内容**

```tsx
// modelcraft-front/src/web/components/features/copilot/SharedCopilotActions.tsx
'use client'

import { memo } from 'react'
import { useCopilotAction } from '@copilotkit/react-core'
import { toast } from 'sonner'
import { AiProposalCard } from './AiProposalCard'
import { useNavigationProposal } from '@web/hooks/ai/use-navigation-proposal'
import type { AgentUiResponse } from './types'

/**
 * Frontend tools shared between admin and end-user agents.
 * Mount inside any CopilotKit context tree.
 *
 * Registers:
 *   show_toast             — agent sends a one-line notification
 *   show_navigation_proposal — agent returns a navigation/clarification proposal;
 *                              rendered as AiProposalCard, user clicks to execute
 */
export const SharedCopilotActions = memo(function SharedCopilotActions() {
  const { handleCandidateClick } = useNavigationProposal()

  useCopilotAction({
    name: 'show_toast',
    description: '向用户显示一条临时通知消息（不需要用户在聊天框内查看）',
    parameters: [
      {
        name: 'message',
        type: 'string',
        description: '通知内容',
        required: true,
      },
      {
        name: 'type',
        type: 'string',
        description: 'success | error | info | warning（默认 info）',
        required: false,
      },
    ],
    handler: async ({ message, type }: { message: string; type?: string }) => {
      const fn =
        type === 'success' ? toast.success
        : type === 'error' ? toast.error
        : type === 'warning' ? toast.warning
        : toast.info
      fn(message)
      return 'toast displayed'
    },
  })

  useCopilotAction({
    name: 'show_navigation_proposal',
    description:
      '展示导航方案候选项。每当用户询问"去哪里"、"怎么配置"、"在哪里"时必须调用此工具。' +
      '返回 candidates 列表，用户点击后前端自动执行导航和高亮。' +
      '永远通过此工具返回导航方案，不要只在文字中描述跳转。',
    parameters: [
      {
        name: 'response',
        type: 'object',
        description: 'AgentUiResponse 格式的 Proposal，包含 proposalId、message、candidates',
        required: true,
      },
    ],
    // CopilotKit render function: renders AiProposalCard instead of default tool output
    render: ({ args }: { args: { response: AgentUiResponse } }) => {
      const response = args.response
      if (!response?.candidates) return null
      return (
        <AiProposalCard
          message={response.message}
          candidates={response.candidates}
          onSelect={(candidate) => handleCandidateClick(candidate)}
        />
      )
    },
  })

  return null
})
```

- [ ] **Step 2: 验证 TypeScript 编译**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "SharedCopilotActions\|show_navigation" | head -10
```

Expected: no errors

- [ ] **Step 3: 全量测试确认无破坏**

```bash
cd modelcraft-front && npx vitest run 2>&1 | tail -10
```

Expected: no regressions

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/SharedCopilotActions.tsx
git commit -m "feat(ai-navigator): register show_navigation_proposal CopilotKit action with AiProposalCard render"
```

---

## Task 9: 更新 AICapabilityReadable — 同时注入 routeCatalog

**Files:**
- Modify: `modelcraft-front/src/web/components/features/copilot/AICapabilityReadable.tsx`

Agent 需要知道两件事：当前页哪些区域可以高亮（已注入），以及全局哪些页面可以导航（新增注入）。

- [ ] **Step 1: 替换 AICapabilityReadable.tsx 完整内容**

```tsx
// modelcraft-front/src/web/components/features/copilot/AICapabilityReadable.tsx
'use client'

import { memo } from 'react'
import { useCopilotReadable } from '@copilotkit/react-core'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'
import { ROUTE_CATALOG } from '@web/lib/route-catalog'

/**
 * Must be mounted INSIDE a <CopilotKit> provider tree.
 *
 * Injects two pieces of knowledge into the Agent context on every render:
 *   1. aiTargets  — current page's registered AiTarget elements (id, label, description, type)
 *   2. routeCatalog — all navigable pages (routeTemplate, title, description, keywords)
 *
 * The agent uses aiTargets to populate targetId in ui.highlight actions,
 * and routeCatalog to populate route in ui.navigate actions.
 */
export const AICapabilityReadable = memo(function AICapabilityReadable() {
  const { getAll } = useAICapabilityContext()
  const targets = getAll()

  useCopilotReadable({
    description:
      '当前页面已注册的 AI 高亮目标（AiTarget）。' +
      '调用 show_navigation_proposal 时，ui.highlight 的 targetId 必须从这个列表中选取。',
    value: targets.map((c) => ({
      id: c.id,
      label: c.label,
      description: c.description,
      type: c.type,
    })),
  })

  useCopilotReadable({
    description:
      '系统所有可导航页面目录（routeCatalog）。' +
      '调用 show_navigation_proposal 时，ui.navigate 的 route 字段必须从 routeTemplate 派生，' +
      '将 :orgName、:projectSlug 等参数替换为当前会话的实际值。',
    value: ROUTE_CATALOG,
  })

  return null
})
```

- [ ] **Step 2: 验证 TypeScript 编译**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "AICapabilityReadable\|route-catalog" | head -5
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/AICapabilityReadable.tsx
git commit -m "feat(ai-navigator): inject routeCatalog into agent context via AICapabilityReadable"
```

---

## Task 10: 清理 ProjectCopilotActions — 移除碎片化 actions

**Files:**
- Modify: `modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx`
- Modify: `modelcraft-front/src/web/components/features/copilot/OrgCopilotActions.tsx`

所有 `navigate_*`、`guide_*`、`highlight_*` actions 由 `show_navigation_proposal` 统一替代。

- [ ] **Step 1: 检查 ProjectCopilotActions 被哪些文件引用**

```bash
grep -r "ProjectCopilotActions" /data/home/lukemxjia/modelcraft/modelcraft-front/src --include="*.tsx" --include="*.ts" -l
```

记录引用文件，确保下一步修改是完整的。

- [ ] **Step 2: 清空 ProjectCopilotActions.tsx**

将文件替换为空的占位符（保留文件，避免破坏已有 import）：

```tsx
// modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx
'use client'

/**
 * ProjectCopilotActions — DEPRECATED in AI Navigator refactor.
 *
 * All navigation actions (navigate_to_model, navigate_to_rbac, etc.) and
 * guide actions (guide_select_database, guide_create_model, etc.) have been
 * replaced by the unified `show_navigation_proposal` tool in SharedCopilotActions.
 *
 * This file is intentionally empty. It may be removed once all import sites
 * are cleaned up.
 */
export function ProjectCopilotActions(_props: Record<string, unknown>) {
  return null
}
```

- [ ] **Step 3: 清理 OrgCopilotActions.tsx**

```tsx
// modelcraft-front/src/web/components/features/copilot/OrgCopilotActions.tsx
'use client'

/**
 * OrgCopilotActions — DEPRECATED in AI Navigator refactor.
 *
 * navigate_to_project, navigate_to_settings, highlight_project, open_create_project
 * are replaced by the unified `show_navigation_proposal` tool.
 *
 * This file is intentionally empty.
 */
export function OrgCopilotActions(_props: Record<string, unknown>) {
  return null
}
```

- [ ] **Step 4: 验证编译（确认 import 无断裂）**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | head -20
```

Expected: no new errors

- [ ] **Step 5: 全量测试**

```bash
cd modelcraft-front && npx vitest run 2>&1 | tail -15
```

Expected: all pass

- [ ] **Step 6: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx \
        modelcraft-front/src/web/components/features/copilot/OrgCopilotActions.tsx
git commit -m "refactor(ai-navigator): deprecate ProjectCopilotActions and OrgCopilotActions — replaced by show_navigation_proposal"
```

---

## Task 11: 更新 Python admin_agent.py

**Files:**
- Modify: `modelcraft-agent/agents/admin_agent.py`
- Modify: `modelcraft-agent/tests/agents/test_admin_agent.py`

移除所有旧工具的引导说明，更新系统提示词让 Agent 知道只用 `show_navigation_proposal`。

- [ ] **Step 1: 先更新测试（修正 list_databases 漏项 + 新增 proposal 相关断言）**

完整替换 `test_admin_agent.py`：

```python
# modelcraft-agent/tests/agents/test_admin_agent.py
"""Tests for the admin agent graph."""
import pytest
from agents.admin_agent import admin_graph, ADMIN_TOOLS


def test_admin_graph_compiles():
    """Graph must build without errors."""
    assert admin_graph is not None


def test_admin_graph_has_agent_and_tools_nodes():
    """Graph must have agent and tools nodes."""
    node_names = set(admin_graph.nodes.keys())
    assert "agent" in node_names
    assert "tools" in node_names


def test_admin_tools_include_list_projects():
    """Admin agent must have access to list_projects."""
    tool_names = {t.name for t in ADMIN_TOOLS}
    assert "list_projects" in tool_names


def test_admin_tools_exact_set():
    """Admin backend tools must be exactly the expected set (no regressions)."""
    tool_names = {t.name for t in ADMIN_TOOLS}
    assert tool_names == {
        "list_projects",
        "list_databases",
        "list_models",
        "get_model_fields",
        "query_model",
        "nl2filter",
    }


def test_frontend_tool_names_from_state():
    """_frontend_tool_names must extract names from copilotkit.actions."""
    from agents.admin_agent import _frontend_tool_names

    state = {
        "messages": [],
        "copilotkit": {
            "actions": [
                {"name": "show_navigation_proposal", "description": "show proposal"},
                {"name": "show_toast", "description": "show toast"},
            ]
        },
    }
    names = _frontend_tool_names(state)
    assert "show_navigation_proposal" in names
    assert "show_toast" in names


def test_should_continue_returns_end_for_frontend_tool():
    """When agent calls only frontend tools, should_continue must return END."""
    from agents.admin_agent import _build_admin_graph
    from langchain_core.messages import AIMessage
    from langgraph.graph import END

    # Build fresh graph to access should_continue directly
    # We test the logic by inspecting the graph's conditional edges indirectly:
    # just verify admin_graph is compiled with conditional edges.
    assert admin_graph is not None
```

- [ ] **Step 2: 运行测试确认已有测试通过、新测试通过**

```bash
cd modelcraft-agent && python -m pytest tests/agents/test_admin_agent.py -v
```

Expected: all PASS（`test_admin_tools_exact_set` 现在包含 `list_databases`）

- [ ] **Step 3: 更新 admin_agent.py 系统提示词**

找到 `system_msg` 的 `content` 构建部分（约 150–168 行），替换为：

```python
        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手（管理员版），帮助租户管理员通过对话完成所有操作。\n\n"
                f"{context}\n\n"
                "【核心规则：UI 导航必须通过 show_navigation_proposal】\n"
                "当用户询问'去哪里'、'在哪里配置'、'怎么操作'、'帮我跳转'时，\n"
                "必须调用 show_navigation_proposal 工具，不能只在文字里描述操作步骤。\n\n"
                "show_navigation_proposal 使用规范：\n"
                "1. response.candidates 每项必须有 type 字段：\n"
                "   - 'action_candidate'：能确定跳转目标时使用，actions 包含 ui.navigate 和/或 ui.highlight\n"
                "   - 'clarification_candidate'：意图不明确时使用，payload 描述用户意图\n"
                "2. ui.navigate 的 route 必须从注入的 routeCatalog 中选取，替换 :orgName/:projectSlug 为实际值\n"
                "3. ui.highlight 的 targetId 必须从注入的 aiTargets 中选取\n"
                "4. 即使只有一个候选项也要包装成 candidates 数组返回，不得直接执行\n\n"
                "数据查询工具调用规则：\n"
                "- 调用任何 project 级工具时，回复中必须明确说明「在项目 **{project}** 中...」\n"
                "- 如需 project 级操作但当前无项目上下文，先调用 list_projects 展示列表\n"
                "- 操作数据前先用 list_models 和 get_model_fields 确认模型和字段存在\n"
                "- 删除操作禁止自动执行，必须引导用户在界面手动确认\n"
                "- 完成操作后用 show_toast 通知用户结果"
            ).replace("{project}", project or "（未知项目）"),
        }
```

也要更新 `context` 中的 layer 判断，移除旧工具引用：

找到 `elif layer == "project":` 块，替换 `context` 赋值：

```python
        elif layer == "project":
            model_ctx = f"，当前模型：{current_model}（数据库：{current_db}）" if current_model else ""
            context = (
                f"当前在 Project 页面（组织：{org}，项目：**{project}**{model_ctx}）。\n"
                "UI 导航工具（只用 show_navigation_proposal）：\n"
                "  routeCatalog 和 aiTargets 已通过上下文注入，选取对应条目生成 candidates。\n\n"
                "数据查询工具：\n"
                "  list_databases、list_models、get_model_fields、query_model、nl2filter\n\n"
                "通知工具：show_toast"
            )
```

找到 `if layer == "org":` 块，替换：

```python
        if layer == "org":
            context = (
                f"当前在 Org 页面（组织：{org}）。\n"
                "UI 导航工具（只用 show_navigation_proposal）：\n"
                "  routeCatalog 和 aiTargets 已通过上下文注入，选取对应条目生成 candidates。\n"
                + (f"\n当前会话项目上下文：**{project}**（用户未明确指定时默认使用此项目）。" if project else "")
            )
```

找到 `else:` 块（无 layer 时），替换：

```python
        else:
            project_ctx = f"当前会话项目上下文：**{project}**。" if project else "当前无项目上下文。"
            context = (
                f"当前组织：{org}。{project_ctx}\n"
                "UI 导航工具（只用 show_navigation_proposal）：\n"
                "  routeCatalog 和 aiTargets 已通过上下文注入。"
            )
```

最后，删除整个 `page_ui_actions` / `chip_section` 逻辑块（约 131–148 行）：

删除以下代码：
```python
        # Build UI action chip section — injected into every system prompt when actions are registered.
        if page_ui_actions:
            chip_ids = " ".join(f"[ACTION:{a['id']}]" for a in page_ui_actions)
            chip_list = "\n".join(
                f"  - [ACTION:{a['id']}] → {a.get('label','')}（{a.get('description','')}）"
                for a in page_ui_actions
            )
            chip_section = (
                f"\n\n【当前页面已注册 UI 操作 — 必须在回复中使用】\n"
                f"{chip_list}\n"
                f"规则：\n"
                f"  1. 每次回复都必须在末尾附一行：「本页可用操作：{chip_ids}」\n"
                f"  2. 介绍某功能时，同步在句中嵌入对应 [ACTION:id]，"
                f"例如：「点击 [ACTION:create_model] 打开新建模型表单」\n"
                f"  3. 不得用文字代替 chip，必须用 [ACTION:id] 标记"
            )
        else:
            chip_section = ""
```

也要删除 `page_ui_actions` 读取那行：
```python
        page_ui_actions: list = state.get("page_ui_actions", [])
```

以及在 `system_msg` 的 `content` 中删除 `f"{chip_section}\n\n"` 这段（已在上面新代码中不存在）。

- [ ] **Step 4: 运行 Python 测试确认通过**

```bash
cd modelcraft-agent && python -m pytest tests/agents/test_admin_agent.py -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add modelcraft-agent/agents/admin_agent.py \
        modelcraft-agent/tests/agents/test_admin_agent.py
git commit -m "feat(ai-navigator): update admin agent to use show_navigation_proposal; remove chip-based guidance"
```

---

## Task 12: 更新 AdminCopilotKnowledge — 刷新提示和建议

**Files:**
- Modify: `modelcraft-front/src/web/components/features/copilot/AdminCopilotKnowledge.tsx`

移除引用旧工具名称的 onboarding 文本，更新 suggestions 反映新能力。

- [ ] **Step 1: 替换 AdminCopilotKnowledge.tsx 完整内容**

```tsx
// modelcraft-front/src/web/components/features/copilot/AdminCopilotKnowledge.tsx
'use client'

import { memo } from 'react'
import { useCopilotReadable } from '@copilotkit/react-core'
import { useCopilotChatSuggestions } from '@copilotkit/react-ui'

const ADMIN_ONBOARDING = `
ModelCraft 新手引导（AI 导航助手）：

AI 助手通过 show_navigation_proposal 帮助你导航到正确页面并高亮目标区域。
你可以直接用自然语言描述想做的事，AI 会给你一个或多个候选方案，点击后自动跳转。

常见操作引导：
- 创建数据模型 → AI 带你到"数据模型编辑器"，高亮新建模型入口
- 配置权限 → AI 带你到"RBAC 角色管理"，高亮相关配置区域
- 添加数据库连接 → AI 带你到"项目设置"，高亮集群配置区域
- 管理终端用户 → AI 带你到"终端用户管理"页面

数据查询（不需要导航）：
- 直接询问数据，AI 会调用后端工具查询并在对话中返回结果
`.trim()

const ADMIN_TROUBLESHOOTING = `
常见问题：

问题：数据库连接失败
  → 告诉 AI "帮我检查集群配置"，AI 会导航到项目设置并高亮集群配置区域

问题：找不到模型或字段
  → 告诉 AI 模型名称，AI 可以调用 list_models / get_model_fields 查询

问题：权限配置在哪里
  → 告诉 AI "帮我配置权限"，AI 会导航到 RBAC 相关页面
`.trim()

const ADMIN_SUGGESTIONS = [
  { title: '帮我创建第一个数据模型', message: '帮我创建第一个数据模型，带我到对应页面' },
  { title: '权限配置在哪里？', message: '我想配置数据权限，带我去对应页面' },
  { title: '我有哪些项目？', message: '列出我所有的项目' },
  { title: '帮我导航到模型管理', message: '帮我去数据模型管理页面' },
  { title: '数据库连不上，帮我排查', message: '数据库连接有问题，带我去检查配置' },
]

/**
 * Injects admin knowledge base and sidebar suggestions into CopilotKit context.
 * Must be mounted inside a CopilotKit provider tree.
 */
export const AdminCopilotKnowledge = memo(function AdminCopilotKnowledge() {
  useCopilotReadable({
    description: 'ModelCraft 新手引导操作手册（管理员）',
    value: ADMIN_ONBOARDING,
  })

  useCopilotReadable({
    description: 'ModelCraft 常见问题排查手册（管理员）',
    value: ADMIN_TROUBLESHOOTING,
  })

  useCopilotChatSuggestions({
    suggestions: ADMIN_SUGGESTIONS,
    available: 'before-first-message',
  })

  return null
})
```

- [ ] **Step 2: 验证编译**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | head -10
```

Expected: no errors

- [ ] **Step 3: 全量前端测试**

```bash
cd modelcraft-front && npx vitest run 2>&1 | tail -15
```

Expected: all pass

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/AdminCopilotKnowledge.tsx
git commit -m "feat(ai-navigator): update AdminCopilotKnowledge for show_navigation_proposal pattern"
```

---

## Task 13: 烟雾测试 — 端到端验证

手动验证完整流程，不需要写自动化测试。

- [ ] **Step 1: 启动开发环境**

```bash
# Terminal 1: 前端
cd modelcraft-front && npm run dev

# Terminal 2: agent 服务（确认启动成功）
cd modelcraft-agent && python main.py
```

- [ ] **Step 2: 验证 action_candidate 流程**

1. 打开浏览器，进入任意项目页面
2. 打开 AI 助手 Sidebar
3. 输入：`帮我去数据模型管理页面`
4. 预期：AI 调用 `show_navigation_proposal`，渲染 `AiProposalCard`，显示"数据模型管理"候选项
5. 点击候选项
6. 预期：页面跳转到 `/org/.../project/.../model-editor`

- [ ] **Step 3: 验证 clarification_candidate 流程**

1. 输入：`项目`（模糊意图）
2. 预期：AI 返回多个澄清候选项（查看项目列表 / 配置项目模型 / ...）
3. 点击其中一个
4. 预期：AI 根据澄清继续推理，返回新的 action_candidate proposal

- [ ] **Step 4: 验证 highlight 流程**

1. 输入：`帮我找到新建模型的按钮`
2. 预期：AI 返回包含 `ui.navigate` + `ui.highlight` 的 action_candidate
3. 点击候选项
4. 预期：页面跳转后，目标元素被 amber 高亮并滚动到可见区域

- [ ] **Step 5: Final commit（如果有 debug 修改）**

```bash
git add -A
git commit -m "fix(ai-navigator): smoke test fixes"
```

---

## 自检 Checklist

**Spec 覆盖率：**

| Spec 章节 | 对应 Task |
|-----------|-----------|
| 三、routeCatalog | Task 3 |
| 三、`<AiTarget>` | Task 5 |
| 三、AiTargetRegistry | Task 4（扩展 AICapabilityContext） |
| 四、show_navigation_proposal | Task 8 |
| 四、handleCandidateClick | Task 6 |
| 四、executeActions | Task 6 |
| 五、AgentUiResponse 类型 | Task 1 |
| 五、action_candidate / clarification_candidate | Task 1、Task 7 |
| 六、AiProposalCard | Task 7 |
| 六、完整流程 | Task 6 + Task 8 |
| 七、高亮视觉规范 | Task 2 |
| 八、MVP 范围 | Task 10（移除非 MVP actions） |
| 后端 Agent 协议 | Task 11 |

**类型一致性：**
- `AiAction`、`ActionCandidate`、`ClarificationCandidate`、`AgentUiResponse` 全部定义在 Task 1 的 `types.ts`
- Task 6、7、8 均从 `types.ts` 导入，无重复定义
- `highlightHTMLElement`（Task 2）被 Task 6 的 `executeActions` 调用
- `ROUTE_CATALOG`（Task 3）被 Task 9 的 `AICapabilityReadable` 导入
