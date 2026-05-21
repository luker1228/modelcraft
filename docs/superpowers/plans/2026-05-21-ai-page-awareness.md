# AI 页面感知与功能高亮 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 AI 助手感知当前页面的可用功能，在回复中渲染可点击的 chip 按钮，点击后琥珀色高亮对应 UI 元素。

**Architecture:** 组件挂载时通过 `useRegisterAICapability` 向 `AICapabilityContext` 注册能力 + DOM ref；`CopilotProvider` 内的 `AICapabilityReadable` 组件通过 `useCopilotReadable` 把能力列表注入 AI；AI 回复中嵌入 `[ACTION:id]` 标记，`AIChipMessage` 解析后渲染为 amber chip，点击直接对 ref 调用 `highlightElement`。

**Tech Stack:** React 18, TypeScript, CopilotKit v1.51.3 (`AssistantMessageProps`), Vitest 4, TailwindCSS

---

## 文件地图

| 操作 | 路径 | 职责 |
|------|------|------|
| 新建 | `modelcraft-front/src/web/contexts/ai-capability-context.tsx` | 能力注册表 Context + Provider |
| 新建 | `modelcraft-front/src/web/hooks/ai/use-register-ai-capability.ts` | 组件注册 hook |
| 新建 | `modelcraft-front/src/web/lib/highlight-element.ts` | 琥珀色高亮工具函数 |
| 新建 | `modelcraft-front/src/web/components/features/copilot/AIChipMessage.tsx` | [ACTION:id] 解析 + chip 渲染 |
| 新建 | `modelcraft-front/src/web/components/features/copilot/AICapabilityReadable.tsx` | 把能力列表注入 CopilotKit |
| 改动 | `modelcraft-front/src/web/components/features/copilot/CopilotProvider.tsx` | 挂载 AICapabilityReadable + AssistantMessage prop |
| 改动 | `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/layout.tsx` | 包裹 AICapabilityProvider |
| 改动 | `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx` | 注册 create_model + select_database |
| 改动 | `modelcraft-agent/agents/admin_agent.py` | 系统 prompt 加 [ACTION:] 规则 |

---

## Task 1: AICapabilityContext + useRegisterAICapability

**Files:**
- Create: `modelcraft-front/src/web/contexts/ai-capability-context.tsx`
- Create: `modelcraft-front/src/web/hooks/ai/use-register-ai-capability.ts`
- Test: `modelcraft-front/src/web/contexts/ai-capability-context.test.ts`

- [ ] **Step 1: 写失败测试**

新建 `modelcraft-front/src/web/contexts/ai-capability-context.test.ts`：

```typescript
import { describe, it, expect } from 'vitest'
import { createCapabilityStore } from './ai-capability-context'

describe('createCapabilityStore', () => {
  it('starts empty', () => {
    const store = createCapabilityStore()
    expect(store.getAll()).toEqual([])
  })

  it('register adds a capability', () => {
    const store = createCapabilityStore()
    const ref = { current: null }
    store.register({ id: 'create_model', label: '新建模型', ref })
    expect(store.getAll()).toHaveLength(1)
    expect(store.getAll()[0].id).toBe('create_model')
  })

  it('unregister removes by id', () => {
    const store = createCapabilityStore()
    const ref = { current: null }
    store.register({ id: 'create_model', label: '新建模型', ref })
    store.unregister('create_model')
    expect(store.getAll()).toEqual([])
  })

  it('getRef returns the registered ref', () => {
    const store = createCapabilityStore()
    const ref = { current: document.createElement('button') }
    store.register({ id: 'create_model', label: '新建模型', ref })
    expect(store.getRef('create_model')).toBe(ref)
  })

  it('later registration overwrites earlier for same id', () => {
    const store = createCapabilityStore()
    const ref1 = { current: null }
    const ref2 = { current: document.createElement('button') }
    store.register({ id: 'create_model', label: '旧标签', ref: ref1 })
    store.register({ id: 'create_model', label: '新标签', ref: ref2 })
    expect(store.getAll()).toHaveLength(1)
    expect(store.getAll()[0].label).toBe('新标签')
  })

  it('getRef returns undefined for unknown id', () => {
    const store = createCapabilityStore()
    expect(store.getRef('nonexistent')).toBeUndefined()
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd modelcraft-front && npm test -- --reporter=verbose src/web/contexts/ai-capability-context.test.ts
```

Expected: FAIL — `createCapabilityStore` not found

- [ ] **Step 3: 实现 ai-capability-context.tsx**

新建 `modelcraft-front/src/web/contexts/ai-capability-context.tsx`：

```typescript
'use client'

import { createContext, useContext, useState, useCallback, type RefObject } from 'react'

export type AICapability = {
  id: string
  label: string
  ref: RefObject<HTMLElement>
  description?: string
}

interface AICapabilityStore {
  register: (capability: AICapability) => void
  unregister: (id: string) => void
  getAll: () => AICapability[]
  getRef: (id: string) => RefObject<HTMLElement> | undefined
}

/** Pure factory — used both in the React context and in tests. */
export function createCapabilityStore(): AICapabilityStore {
  const map = new Map<string, AICapability>()
  return {
    register: (cap) => map.set(cap.id, cap),
    unregister: (id) => map.delete(id),
    getAll: () => Array.from(map.values()),
    getRef: (id) => map.get(id)?.ref,
  }
}

const AICapabilityContext = createContext<{
  register: (capability: AICapability) => void
  unregister: (id: string) => void
  getAll: () => AICapability[]
  getRef: (id: string) => RefObject<HTMLElement> | undefined
} | null>(null)

export function AICapabilityProvider({ children }: { children: React.ReactNode }) {
  // version counter drives re-renders so useCopilotReadable consumers pick up changes
  const [, setVersion] = useState(0)
  const [store] = useState(() => createCapabilityStore())

  const register = useCallback((cap: AICapability) => {
    store.register(cap)
    setVersion((v) => v + 1)
  }, [store])

  const unregister = useCallback((id: string) => {
    store.unregister(id)
    setVersion((v) => v + 1)
  }, [store])

  return (
    <AICapabilityContext.Provider value={{
      register,
      unregister,
      getAll: store.getAll,
      getRef: store.getRef,
    }}>
      {children}
    </AICapabilityContext.Provider>
  )
}

export function useAICapabilityContext() {
  const ctx = useContext(AICapabilityContext)
  if (!ctx) throw new Error('useAICapabilityContext must be used inside AICapabilityProvider')
  return ctx
}
```

- [ ] **Step 4: 实现 use-register-ai-capability.ts**

新建 `modelcraft-front/src/web/hooks/ai/use-register-ai-capability.ts`：

```typescript
'use client'

import { useEffect, type RefObject } from 'react'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'

/**
 * Register a UI element's capability with the AI assistant.
 * Automatically unregisters when the component unmounts.
 *
 * @param id       Unique action identifier, e.g. "create_model"
 * @param label    Display label for AI suggestions, e.g. "新建模型"
 * @param ref      Ref to the DOM element that will be highlighted on click
 * @param description  Optional hint for the AI about what this action does
 */
export function useRegisterAICapability(
  id: string,
  label: string,
  ref: RefObject<HTMLElement>,
  description?: string,
) {
  const { register, unregister } = useAICapabilityContext()

  useEffect(() => {
    register({ id, label, ref, description })
    return () => unregister(id)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, label, description])
  // Note: `ref` identity is stable (useRef), no need to add it to deps.
  // `register`/`unregister` are stable callbacks from useCallback.
}
```

- [ ] **Step 5: 运行测试确认通过**

```bash
cd modelcraft-front && npm test -- --reporter=verbose src/web/contexts/ai-capability-context.test.ts
```

Expected: 6 tests PASS

- [ ] **Step 6: 提交**

```bash
cd modelcraft-front
git add src/web/contexts/ai-capability-context.tsx src/web/contexts/ai-capability-context.test.ts src/web/hooks/ai/use-register-ai-capability.ts
git commit -m "feat: add AICapabilityContext and useRegisterAICapability hook"
```

---

## Task 2: highlightElement 工具函数

**Files:**
- Create: `modelcraft-front/src/web/lib/highlight-element.ts`
- Test: `modelcraft-front/src/web/lib/highlight-element.test.ts`

- [ ] **Step 1: 写失败测试**

新建 `modelcraft-front/src/web/lib/highlight-element.test.ts`：

```typescript
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { highlightElement, HIGHLIGHT_CLASSES } from './highlight-element'

describe('highlightElement', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('adds highlight classes to the element', () => {
    const el = document.createElement('button')
    highlightElement({ current: el })
    for (const cls of HIGHLIGHT_CLASSES) {
      expect(el.classList.contains(cls)).toBe(true)
    }
  })

  it('removes highlight classes after durationMs', () => {
    const el = document.createElement('button')
    highlightElement({ current: el }, 1000)
    vi.advanceTimersByTime(1000)
    for (const cls of HIGHLIGHT_CLASSES) {
      expect(el.classList.contains(cls)).toBe(false)
    }
  })

  it('does not throw when ref.current is null', () => {
    expect(() => highlightElement({ current: null })).not.toThrow()
  })

  it('default duration is 5000ms', () => {
    const el = document.createElement('button')
    highlightElement({ current: el })
    vi.advanceTimersByTime(4999)
    expect(el.classList.contains(HIGHLIGHT_CLASSES[0])).toBe(true)
    vi.advanceTimersByTime(1)
    expect(el.classList.contains(HIGHLIGHT_CLASSES[0])).toBe(false)
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd modelcraft-front && npm test -- --reporter=verbose src/web/lib/highlight-element.test.ts
```

Expected: FAIL — `highlightElement` not found

- [ ] **Step 3: 实现 highlight-element.ts**

新建 `modelcraft-front/src/web/lib/highlight-element.ts`：

```typescript
import type { RefObject } from 'react'

/**
 * Tailwind classes applied to highlighted elements.
 * Matches the existing table-row highlight style in DevelopRecordWorkspace.
 */
export const HIGHLIGHT_CLASSES = [
  'bg-amber-50',
  'ring-2',
  'ring-amber-300',
  'ring-offset-1',
  'transition-all',
] as const

/**
 * Apply amber highlight to a DOM element and auto-remove after durationMs.
 * Silently skips if ref.current is null (element unmounted).
 */
export function highlightElement(
  ref: RefObject<HTMLElement>,
  durationMs = 5000,
): void {
  const el = ref.current
  if (!el) return

  // Scroll into view if not visible
  el.scrollIntoView({ behavior: 'smooth', block: 'nearest' })

  // Apply highlight
  el.classList.add(...HIGHLIGHT_CLASSES)

  // Auto-remove after duration
  setTimeout(() => {
    el.classList.remove(...HIGHLIGHT_CLASSES)
  }, durationMs)
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd modelcraft-front && npm test -- --reporter=verbose src/web/lib/highlight-element.test.ts
```

Expected: 4 tests PASS

- [ ] **Step 5: 提交**

```bash
cd modelcraft-front
git add src/web/lib/highlight-element.ts src/web/lib/highlight-element.test.ts
git commit -m "feat: add highlightElement utility with amber ring style"
```

---

## Task 3: parseActionMarkers + AIChipMessage 组件

**Files:**
- Create: `modelcraft-front/src/web/components/features/copilot/AIChipMessage.tsx`
- Test: `modelcraft-front/src/web/components/features/copilot/AIChipMessage.test.ts`

- [ ] **Step 1: 写失败测试（只测纯函数部分）**

新建 `modelcraft-front/src/web/components/features/copilot/AIChipMessage.test.ts`：

```typescript
import { describe, it, expect } from 'vitest'
import { parseActionMarkers, type MessageSegment } from './AIChipMessage'

describe('parseActionMarkers', () => {
  it('returns plain text segment when no markers', () => {
    const result = parseActionMarkers('Hello world')
    expect(result).toEqual([{ type: 'text', content: 'Hello world' }])
  })

  it('extracts a single ACTION marker', () => {
    const result = parseActionMarkers('Click [ACTION:create_model] to start')
    expect(result).toEqual([
      { type: 'text', content: 'Click ' },
      { type: 'action', id: 'create_model' },
      { type: 'text', content: ' to start' },
    ])
  })

  it('extracts multiple ACTION markers', () => {
    const result = parseActionMarkers('[ACTION:create_model] or [ACTION:connect_db]')
    expect(result).toEqual([
      { type: 'action', id: 'create_model' },
      { type: 'text', content: ' or ' },
      { type: 'action', id: 'connect_db' },
    ])
  })

  it('ignores empty text segments', () => {
    const result = parseActionMarkers('[ACTION:create_model]')
    expect(result).toEqual([{ type: 'action', id: 'create_model' }])
  })

  it('handles ACTION at end of string', () => {
    const result = parseActionMarkers('请点击 [ACTION:create_model]')
    expect(result).toEqual([
      { type: 'text', content: '请点击 ' },
      { type: 'action', id: 'create_model' },
    ])
  })

  it('returns empty array for empty input', () => {
    expect(parseActionMarkers('')).toEqual([])
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd modelcraft-front && npm test -- --reporter=verbose src/web/components/features/copilot/AIChipMessage.test.ts
```

Expected: FAIL — `parseActionMarkers` not found

- [ ] **Step 3: 实现 AIChipMessage.tsx**

新建 `modelcraft-front/src/web/components/features/copilot/AIChipMessage.tsx`：

```tsx
'use client'

import { memo } from 'react'
import { AssistantMessage } from '@copilotkit/react-ui'
import type { AssistantMessageProps } from '@copilotkit/react-ui'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'
import { highlightElement } from '@web/lib/highlight-element'

/** A parsed segment of an AI message. */
export type MessageSegment =
  | { type: 'text'; content: string }
  | { type: 'action'; id: string }

const ACTION_REGEX = /\[ACTION:([^\]]+)\]/g

/**
 * Parse a message string into text and ACTION marker segments.
 * Pure function — exported for testing.
 */
export function parseActionMarkers(text: string): MessageSegment[] {
  const segments: MessageSegment[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null

  ACTION_REGEX.lastIndex = 0
  while ((match = ACTION_REGEX.exec(text)) !== null) {
    if (match.index > lastIndex) {
      segments.push({ type: 'text', content: text.slice(lastIndex, match.index) })
    }
    segments.push({ type: 'action', id: match[1] })
    lastIndex = match.index + match[0].length
  }

  if (lastIndex < text.length) {
    segments.push({ type: 'text', content: text.slice(lastIndex) })
  }

  return segments
}

/**
 * Drop-in replacement for CopilotKit's AssistantMessage.
 * Renders [ACTION:id] markers as clickable amber chip buttons.
 * Unknown action IDs (not registered in AICapabilityContext) render as disabled chips.
 *
 * Usage in CopilotProvider:
 *   <CopilotSidebar AssistantMessage={AIChipMessage} ... />
 */
export const AIChipMessage = memo(function AIChipMessage(props: AssistantMessageProps) {
  const { getRef, getAll } = useAICapabilityContext()
  const content: string = props.message?.content ?? ''

  // Only process messages that contain ACTION markers
  if (!content.includes('[ACTION:')) {
    return <AssistantMessage {...props} />
  }

  const segments = parseActionMarkers(content)
  const registeredIds = new Set(getAll().map((c) => c.id))

  // Reconstruct the text-only content for the default renderer (removes ACTION markers)
  const textOnly = segments
    .filter((s): s is { type: 'text'; content: string } => s.type === 'text')
    .map((s) => s.content)
    .join('')

  const handleChipClick = (actionId: string) => {
    const ref = getRef(actionId)
    if (ref) {
      highlightElement(ref)
    }
  }

  return (
    <div>
      {/* Render text segments through the default AssistantMessage */}
      <AssistantMessage {...props} message={props.message ? { ...props.message, content: textOnly } : props.message} />
      {/* Render chip buttons below the text */}
      <div className="mt-2 flex flex-wrap gap-2 px-3 pb-2">
        {segments
          .filter((s): s is { type: 'action'; id: string } => s.type === 'action')
          .map((seg) => {
            const known = registeredIds.has(seg.id)
            const capability = getAll().find((c) => c.id === seg.id)
            const label = capability?.label ?? seg.id
            return (
              <button
                key={seg.id}
                type="button"
                disabled={!known}
                onClick={() => handleChipClick(seg.id)}
                title={known ? `高亮 ${label}` : '该操作当前不可用'}
                className={
                  known
                    ? 'inline-flex items-center gap-1.5 rounded-full border border-amber-300 bg-amber-50 px-3 py-1 text-xs font-medium text-amber-900 transition-colors hover:bg-amber-100 cursor-pointer'
                    : 'inline-flex items-center gap-1.5 rounded-full border border-muted bg-muted/50 px-3 py-1 text-xs font-medium text-muted-foreground cursor-not-allowed'
                }
              >
                {known && <span aria-hidden>✨</span>}
                {label}
              </button>
            )
          })}
      </div>
    </div>
  )
})
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd modelcraft-front && npm test -- --reporter=verbose src/web/components/features/copilot/AIChipMessage.test.ts
```

Expected: 6 tests PASS

- [ ] **Step 5: 提交**

```bash
cd modelcraft-front
git add src/web/components/features/copilot/AIChipMessage.tsx src/web/components/features/copilot/AIChipMessage.test.ts
git commit -m "feat: add AIChipMessage component with ACTION marker parsing"
```

---

## Task 4: AICapabilityReadable + CopilotProvider 改动

**Files:**
- Create: `modelcraft-front/src/web/components/features/copilot/AICapabilityReadable.tsx`
- Modify: `modelcraft-front/src/web/components/features/copilot/CopilotProvider.tsx`

- [ ] **Step 1: 实现 AICapabilityReadable.tsx**

新建 `modelcraft-front/src/web/components/features/copilot/AICapabilityReadable.tsx`：

```tsx
'use client'

import { memo } from 'react'
import { useCopilotReadable } from '@copilotkit/react-core'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'

/**
 * Must be mounted INSIDE a <CopilotKit> provider tree.
 * Reads the current page's registered capabilities and injects them
 * into the AI context via useCopilotReadable on every change.
 */
export const AICapabilityReadable = memo(function AICapabilityReadable() {
  const { getAll } = useAICapabilityContext()
  const capabilities = getAll()

  useCopilotReadable({
    description: '当前页面可用的 UI 操作（点击 [ACTION:id] chip 可高亮对应元素）',
    value: capabilities.map((c) => ({
      id: c.id,
      label: c.label,
      description: c.description,
    })),
    // Empty array is fine — CopilotKit injects it but AI will see an empty list
    // and correctly skip using [ACTION:] markers.
  })

  return null
})
```

- [ ] **Step 2: 修改 CopilotProvider.tsx**

读取当前文件 `modelcraft-front/src/web/components/features/copilot/CopilotProvider.tsx`，做两处修改：

**修改一**：在 import 区域末尾追加两行新 import：

```tsx
// 在最后一个 import 行之后加入
import { AICapabilityReadable } from './AICapabilityReadable'
import { AIChipMessage } from './AIChipMessage'
```

**修改二**：在 admin `CopilotProvider` 的 `<CopilotKit>` 内的 `<AdminCopilotKnowledge />` 后加 `<AICapabilityReadable />`，并给 `<CopilotSidebar>` 添加 `AssistantMessage` prop：

将：
```tsx
    <CopilotKit
      runtimeUrl="/api/copilotkit"
      agent="modelcraft_admin_agent"
      headers={headers}
      properties={copilotContext}
    >
      <SharedCopilotActions />
      <AdminCopilotKnowledge />
      {children}
      <CopilotSidebar
        labels={{
          title: 'ModelCraft AI 助手',
          initial: initialMessage,
        }}
        defaultOpen={false}
        clickOutsideToClose={true}
      />
    </CopilotKit>
```

替换为：
```tsx
    <CopilotKit
      runtimeUrl="/api/copilotkit"
      agent="modelcraft_admin_agent"
      headers={headers}
      properties={copilotContext}
    >
      <SharedCopilotActions />
      <AdminCopilotKnowledge />
      <AICapabilityReadable />
      {children}
      <CopilotSidebar
        labels={{
          title: 'ModelCraft AI 助手',
          initial: initialMessage,
        }}
        defaultOpen={false}
        clickOutsideToClose={true}
        AssistantMessage={AIChipMessage}
      />
    </CopilotKit>
```

**修改三**：同样修改 `EndUserCopilotWrapper` 里的 `<CopilotSidebar>`（加 `AssistantMessage={AIChipMessage}`，不加 `<AICapabilityReadable />` — end-user 侧暂不使用能力注册）：

```tsx
          <CopilotSidebar
            labels={{
              title: 'ModelCraft AI 助手',
              initial: initialMessage,
            }}
            defaultOpen={false}
            AssistantMessage={AIChipMessage}
          />
```

- [ ] **Step 3: TypeScript 检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep -E "AICapabilityReadable|AIChipMessage|CopilotProvider" | head -20
```

Expected: 无错误输出（或只有不相关的已有错误）

- [ ] **Step 4: 提交**

```bash
cd modelcraft-front
git add src/web/components/features/copilot/AICapabilityReadable.tsx src/web/components/features/copilot/CopilotProvider.tsx
git commit -m "feat: wire AICapabilityReadable and AIChipMessage into CopilotProvider"
```

---

## Task 5: Project Layout 接入 AICapabilityProvider

**Files:**
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/layout.tsx`

- [ ] **Step 1: 修改 layout.tsx**

在现有 import 列表末尾加入：
```tsx
import { AICapabilityProvider } from '@web/contexts/ai-capability-context'
```

将现有 return 语句：
```tsx
  return (
    <WorkspaceAIRefContext.Provider value={workspaceAiRef}>
      <CopilotWrapper selectedProject={selectedProject} orgName={orgName}>
        <ProjectAIContext
          orgName={orgName}
          projectSlug={projectSlug}
          workspaceAiRef={workspaceAiRef}
        />
        {mainContent}
      </CopilotWrapper>
    </WorkspaceAIRefContext.Provider>
  )
```

替换为：
```tsx
  return (
    <AICapabilityProvider>
      <WorkspaceAIRefContext.Provider value={workspaceAiRef}>
        <CopilotWrapper selectedProject={selectedProject} orgName={orgName}>
          <ProjectAIContext
            orgName={orgName}
            projectSlug={projectSlug}
            workspaceAiRef={workspaceAiRef}
          />
          {mainContent}
        </CopilotWrapper>
      </WorkspaceAIRefContext.Provider>
    </AICapabilityProvider>
  )
```

- [ ] **Step 2: TypeScript 检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "layout.tsx" | head -10
```

Expected: 无错误

- [ ] **Step 3: 提交**

```bash
cd modelcraft-front
git add src/app/org/\[orgName\]/project/\[projectSlug\]/layout.tsx
git commit -m "feat: wrap project layout with AICapabilityProvider"
```

---

## Task 6: ModelSidebar P0 能力注册

**Files:**
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx`

在此组件里注册两个 P0 能力：
- `create_model` → 指向"新建模型"按钮
- `select_database` → 指向数据库选择器触发按钮

- [ ] **Step 1: 在 ModelSidebar.tsx 添加 import**

在文件顶部的 React import 区域（`'use client'` 后第一行）加入：
```tsx
import { useRef } from 'react'
import { useRegisterAICapability } from '@web/hooks/ai/use-register-ai-capability'
```

- [ ] **Step 2: 添加 refs 和注册调用**

在 `ModelSidebar` 函数体内（`const { pendingAction, setPendingAction } = useOnboarding()` 之后），加入：

```tsx
  // AI capability refs for chip highlighting
  const createModelBtnRef = useRef<HTMLButtonElement>(null)
  const selectDbBtnRef = useRef<HTMLButtonElement>(null)

  useRegisterAICapability('create_model', '新建模型', createModelBtnRef, '点击打开新建模型表单')
  useRegisterAICapability('select_database', '选择数据库', selectDbBtnRef, '点击选择要操作的数据库')
```

- [ ] **Step 3: 给"新建模型"按钮添加 ref**

找到现有的"新建模型"按钮（当前在 `<div className="flex flex-col gap-1 px-3 py-2.5">` 内），将其 `<Button` 改为携带 ref：

将：
```tsx
          <Button
            size="sm"
            variant="outline"
            className={cn(
              'h-7 w-full justify-start px-2.5 text-xs font-normal transition-colors',
              !state.selectedDatabase && 'pointer-events-none opacity-40',
              pendingAction === 'nav_create_model' && state.selectedDatabase && 'ring-2 ring-amber-400 ring-offset-1 animate-pulse border-amber-400'
            )}
            onClick={handleCreateModelClick}
            disabled={!state.selectedDatabase}
            >
```

替换为：
```tsx
          <Button
            ref={createModelBtnRef}
            size="sm"
            variant="outline"
            className={cn(
              'h-7 w-full justify-start px-2.5 text-xs font-normal transition-colors',
              !state.selectedDatabase && 'pointer-events-none opacity-40',
              pendingAction === 'nav_create_model' && state.selectedDatabase && 'ring-2 ring-amber-400 ring-offset-1 animate-pulse border-amber-400'
            )}
            onClick={handleCreateModelClick}
            disabled={!state.selectedDatabase}
            >
```

- [ ] **Step 4: 给数据库选择器触发按钮添加 ref**

找到 `<PopoverTrigger asChild>` 内的 `<Button`（数据库选择器），添加 ref：

将：
```tsx
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              size="sm"
              className={cn(
```

替换为：
```tsx
          <PopoverTrigger asChild>
            <Button
              ref={selectDbBtnRef}
              variant="outline"
              size="sm"
              className={cn(
```

- [ ] **Step 5: TypeScript 检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "ModelSidebar" | head -10
```

Expected: 无错误

- [ ] **Step 6: 提交**

```bash
cd modelcraft-front
git add src/app/org/\[orgName\]/project/\[projectSlug\]/model-editor/_components/ModelSidebar.tsx
git commit -m "feat: register create_model and select_database AI capabilities in ModelSidebar"
```

---

## Task 6b: 其余 P0 能力注册（ClusterPanel + FieldListPage）

**Files:**
- Find + Modify: cluster 页面的"连接数据库"按钮 → `connect_db`
- Find + Modify: field 列表页的"新建字段"按钮 → `create_field`

- [ ] **Step 1: 找到 cluster 连接按钮**

```bash
grep -rn "navigate_to_cluster\|连接\|connect\|AddCluster\|CreateCluster" modelcraft-front/src/app --include="*.tsx" -l | head -5
```

打开找到的文件，定位"连接数据库"或"添加集群"按钮元素。

- [ ] **Step 2: 给 ClusterPanel 组件注册 connect_db 能力**

在找到的组件顶部加 import（同 Task 6 Step 1 的模式）：
```tsx
import { useRef } from 'react'
import { useRegisterAICapability } from '@web/hooks/ai/use-register-ai-capability'
```

在函数体内加：
```tsx
  const connectDbBtnRef = useRef<HTMLButtonElement>(null)
  useRegisterAICapability('connect_db', '连接数据库', connectDbBtnRef, '点击配置数据库集群连接')
```

给目标按钮加 `ref={connectDbBtnRef}`。

- [ ] **Step 3: 找到 FieldListPage 新建字段按钮**

```bash
grep -rn "新建字段\|create.*field\|CreateField\|AddField" modelcraft-front/src/app --include="*.tsx" -l | head -5
```

- [ ] **Step 4: 给 FieldListPage 注册 create_field 能力**

同 Step 2 模式，注册：
```tsx
  const createFieldBtnRef = useRef<HTMLButtonElement>(null)
  useRegisterAICapability('create_field', '新建字段', createFieldBtnRef, '点击打开新建字段表单')
```

给目标按钮加 `ref={createFieldBtnRef}`。

- [ ] **Step 5: TypeScript 检查**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | tail -5
```

Expected: 无新增错误

- [ ] **Step 6: 提交**

```bash
cd modelcraft-front
git add -u
git commit -m "feat: register connect_db and create_field AI capabilities"
```

---

## Task 7: Admin Agent 系统 Prompt 更新

**Files:**
- Modify: `modelcraft-agent/agents/admin_agent.py:129-147`

- [ ] **Step 1: 在 admin_agent.py 的 project context 字符串里加入 [ACTION:] 规则**

在 `admin_agent.py` 的 `agent_node` 函数中，找到 `elif layer == "project":` 块里的 `context` 变量定义，在最后拼接的字符串末尾追加以下内容（在 `"引导工具说明..."` 那段之后）：

将：
```python
                "引导工具说明：guide_select_database / guide_create_model 只是高亮 UI 元素，\n"
                "  不替代用户点击——高亮后必须用文字告知用户需要执行什么操作。"
```

替换为：
```python
                "引导工具说明：guide_select_database / guide_create_model 只是高亮 UI 元素，\n"
                "  不替代用户点击——高亮后必须用文字告知用户需要执行什么操作。\n\n"
                "UI 操作 Chip 规则（[ACTION:id] 标记）：\n"
                "  当你需要引导用户使用页面上的某个功能时，可以在回复文本中插入 [ACTION:action_id] 标记。\n"
                "  前端会把它渲染成可点击的按钮，用户点击后自动高亮对应 UI 元素。\n"
                "  只使用系统上下文「当前页面可用的 UI 操作」列表里的 action_id，不要编造。\n"
                "  示例：「点击 [ACTION:create_model] 即可打开新建模型表单。」\n"
                "  如果当前页面没有相关 action_id，正常用文字回答，不使用此标记。"
```

- [ ] **Step 2: 验证修改位置正确**

```bash
grep -n "\[ACTION" modelcraft-agent/agents/admin_agent.py
```

Expected: 输出包含类似 `119:  "  示例：「点击 [ACTION:create_model] 即可打开新建模型表单。」\n"` 的行

- [ ] **Step 3: 运行已有 agent 测试**

```bash
cd modelcraft-agent && python -m pytest tests/agents/test_admin_agent.py -v 2>&1 | tail -20
```

Expected: 已有测试全部通过（本次修改不改变工具列表，只改 system prompt）

- [ ] **Step 4: 提交**

```bash
git add modelcraft-agent/agents/admin_agent.py
git commit -m "feat: add [ACTION:id] chip guidance to admin agent system prompt"
```

---

## 验收检查清单

完成所有 Task 后，在浏览器中手动验证：

- [ ] 打开项目页面，进入 model-editor，打开 AI 助手侧边栏
- [ ] 发送消息："如何新建模型？"
- [ ] 确认 AI 回复中出现 `✨ 新建模型` amber chip 按钮
- [ ] 点击 chip，确认"新建模型"按钮被琥珀色光环包围
- [ ] 确认 5 秒后高亮自动消失
- [ ] 发送消息："如何选择数据库？"，确认出现 `✨ 选择数据库` chip
- [ ] 点击 chip，确认数据库选择器按钮被高亮
- [ ] 在没有注册能力的页面（如 RBAC 页）验证 chip 不出现

---

## 运行所有单元测试

```bash
cd modelcraft-front && npm test
```

Expected: 所有测试通过（包括新增的 3 个测试文件）
