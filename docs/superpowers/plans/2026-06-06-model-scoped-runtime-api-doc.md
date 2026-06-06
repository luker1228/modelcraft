# Model-Scoped Runtime API Doc Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a model-scoped API documentation modal to the end-user record workspace that teaches the current runtime endpoint, token header format, one working `findMany` curl example, and an AI prompt.

**Architecture:** Keep all logic in the frontend. Extract the generated documentation strings into a small pure helper module so the runtime URL, curl snippet, and AI prompt can be unit tested independently from the UI. Render the docs in a dedicated dialog component, then mount that dialog from `EndUserRecordWorkspace` with the already-loaded `orgName`, `projectSlug`, `databaseName`, and `modelName`.

**Tech Stack:** Next.js 14, React 18, TypeScript, Radix Dialog via shadcn/ui, Vitest, Testing Library

---

## File Structure

- Create: `modelcraft-front/src/web/components/features/end-user-data/model-api-docs.ts`
  Responsibility: pure string builders for server URL, full endpoint, token header, `findMany` curl example, and AI prompt.

- Create: `modelcraft-front/src/web/components/features/end-user-data/model-api-docs.test.ts`
  Responsibility: unit tests for generated endpoint, curl, and AI prompt content.

- Create: `modelcraft-front/src/web/components/features/end-user-data/ModelApiDocsDialog.tsx`
  Responsibility: documentation-first modal UI with copy actions and five fixed sections.

- Create: `modelcraft-front/src/web/components/features/end-user-data/ModelApiDocsDialog.test.tsx`
  Responsibility: render tests for the modal content, copy buttons, and disabled/incomplete-context fallback.

- Modify: `modelcraft-front/src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx`
  Responsibility: add the `API 文档` trigger, compute dialog context from the loaded model, and open/close the modal.

- Modify: `modelcraft-front/src/web/components/features/end-user-data/index.ts`
  Responsibility: export the new dialog only if needed by local imports; otherwise leave unchanged.

---

### Task 1: Add and verify pure runtime-doc builders

**Files:**
- Create: `modelcraft-front/src/web/components/features/end-user-data/model-api-docs.ts`
- Create: `modelcraft-front/src/web/components/features/end-user-data/model-api-docs.test.ts`

- [ ] **Step 1: Write the failing unit tests for endpoint, curl, and AI prompt generation**

```ts
import { describe, expect, it } from 'vitest'
import {
  API_DOC_SERVER_URL,
  buildModelRuntimeEndpoint,
  buildFindManyCurlSnippet,
  buildModelApiAiPrompt,
} from './model-api-docs'

describe('model-api-docs', () => {
  const input = {
    orgName: 'luke_5l0o',
    projectSlug: 'shop',
    databaseName: 'prod',
    modelName: 'orders',
  }

  it('builds the full runtime endpoint with real values', () => {
    expect(buildModelRuntimeEndpoint(input)).toBe(
      'http://lukemxjia.devcloud.woa.com:9080/end-user/graphql/org/luke_5l0o/project/shop/db/prod/model/orders'
    )
  })

  it('builds a single findMany curl example that only requires token replacement', () => {
    const curl = buildFindManyCurlSnippet(input)

    expect(curl).toContain('SERVER_URL="http://lukemxjia.devcloud.woa.com:9080"')
    expect(curl).toContain('TOKEN="replace-with-your-api-token"')
    expect(curl).toContain('Authorization: Bearer ${TOKEN}')
    expect(curl).toContain('findMany(take: 5, skip: 0)')
    expect(curl).toContain('items { id }')
    expect(curl).not.toContain('<projectSlug>')
  })

  it('builds an AI prompt with the endpoint, auth header, and working example', () => {
    const prompt = buildModelApiAiPrompt(input)

    expect(prompt).toContain(API_DOC_SERVER_URL)
    expect(prompt).toContain(buildModelRuntimeEndpoint(input))
    expect(prompt).toContain('Authorization: Bearer <API_TOKEN>')
    expect(prompt).toContain('findMany(take: 5, skip: 0)')
    expect(prompt).toContain('[把你想查询或写入的业务目标写在这里]')
  })
})
```

- [ ] **Step 2: Run the new test file to verify it fails**

Run:

```bash
cd modelcraft-front
npx vitest run src/web/components/features/end-user-data/model-api-docs.test.ts
```

Expected: FAIL with module-not-found or missing-export errors for `./model-api-docs`.

- [ ] **Step 3: Implement the pure builder module**

```ts
export const API_DOC_SERVER_URL = 'http://lukemxjia.devcloud.woa.com:9080'

export interface ModelApiDocContext {
  orgName: string
  projectSlug: string
  databaseName: string
  modelName: string
}

export function buildModelRuntimeEndpoint({
  orgName,
  projectSlug,
  databaseName,
  modelName,
}: ModelApiDocContext): string {
  return (
    `${API_DOC_SERVER_URL}/end-user/graphql` +
    `/org/${orgName}/project/${projectSlug}` +
    `/db/${databaseName}/model/${modelName}`
  )
}

export function buildFindManyCurlSnippet(context: ModelApiDocContext): string {
  const endpoint = buildModelRuntimeEndpoint(context)

  return `SERVER_URL="${API_DOC_SERVER_URL}"
TOKEN="replace-with-your-api-token"

curl -X POST "${endpoint}" \\
  -H "Authorization: Bearer \${TOKEN}" \\
  -H "Content-Type: application/json" \\
  -d '{"query":"query { findMany(take: 5, skip: 0) { items { id } } }"}'`
}

export function buildModelApiAiPrompt(context: ModelApiDocContext): string {
  const endpoint = buildModelRuntimeEndpoint(context)

  return `你是 GraphQL 助手。请基于下面这个 Runtime GraphQL 接口，帮我写查询。

Server URL:
${API_DOC_SERVER_URL}

Endpoint:
${endpoint}

认证方式:
Authorization: Bearer <API_TOKEN>

当前已知可用示例:
query { findMany(take: 5, skip: 0) { items { id } } }

我的目标是:
[把你想查询或写入的业务目标写在这里]

请输出:
1. GraphQL 查询或变更
2. 对应 curl 命令
3. 如有字段需要替换，请明确标出`
}
```

- [ ] **Step 4: Run the tests again to verify they pass**

Run:

```bash
cd modelcraft-front
npx vitest run src/web/components/features/end-user-data/model-api-docs.test.ts
```

Expected: PASS with 3 passing tests.

- [ ] **Step 5: Commit the pure builders**

```bash
git add modelcraft-front/src/web/components/features/end-user-data/model-api-docs.ts modelcraft-front/src/web/components/features/end-user-data/model-api-docs.test.ts
git commit -m "feat: add model scoped runtime api doc builders"
```

---

### Task 2: Build the model-scoped API documentation dialog

**Files:**
- Create: `modelcraft-front/src/web/components/features/end-user-data/ModelApiDocsDialog.tsx`
- Create: `modelcraft-front/src/web/components/features/end-user-data/ModelApiDocsDialog.test.tsx`
- Modify: `modelcraft-front/src/web/components/features/end-user-data/model-api-docs.ts`

- [ ] **Step 1: Write the failing dialog tests**

```tsx
import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ModelApiDocsDialog } from './ModelApiDocsDialog'

describe('ModelApiDocsDialog', () => {
  it('renders the five documentation sections for a complete model context', () => {
    render(
      <ModelApiDocsDialog
        open
        onOpenChange={() => {}}
        context={{
          orgName: 'luke_5l0o',
          projectSlug: 'shop',
          databaseName: 'prod',
          modelName: 'orders',
        }}
      />
    )

    expect(screen.getByText('Server URL')).toBeInTheDocument()
    expect(screen.getByText('URL 含义')).toBeInTheDocument()
    expect(screen.getByText('Token 怎么填')).toBeInTheDocument()
    expect(screen.getByText('curl 示例')).toBeInTheDocument()
    expect(screen.getByText('如何让 AI 帮你继续写')).toBeInTheDocument()
  })

  it('shows the model-specific endpoint and findMany curl example', () => {
    render(
      <ModelApiDocsDialog
        open
        onOpenChange={() => {}}
        context={{
          orgName: 'luke_5l0o',
          projectSlug: 'shop',
          databaseName: 'prod',
          modelName: 'orders',
        }}
      />
    )

    expect(
      screen.getByText(/http:\/\/lukemxjia\.devcloud\.woa\.com:9080\/end-user\/graphql\/org\/luke_5l0o\/project\/shop\/db\/prod\/model\/orders/)
    ).toBeInTheDocument()
    expect(screen.getByText(/findMany\(take: 5, skip: 0\)/)).toBeInTheDocument()
  })

  it('copies the curl example', async () => {
    vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue()

    render(
      <ModelApiDocsDialog
        open
        onOpenChange={() => {}}
        context={{
          orgName: 'luke_5l0o',
          projectSlug: 'shop',
          databaseName: 'prod',
          modelName: 'orders',
        }}
      />
    )

    fireEvent.click(screen.getByRole('button', { name: '复制 curl' }))

    expect(navigator.clipboard.writeText).toHaveBeenCalled()
  })
})
```

- [ ] **Step 2: Run the dialog tests to verify they fail**

Run:

```bash
cd modelcraft-front
npx vitest run src/web/components/features/end-user-data/ModelApiDocsDialog.test.tsx
```

Expected: FAIL with module-not-found for `ModelApiDocsDialog`.

- [ ] **Step 3: Implement the dialog component using existing shadcn dialog patterns**

```tsx
'use client'

import { useState } from 'react'
import { Check, Copy, FileCode2 } from 'lucide-react'
import {
  API_DOC_SERVER_URL,
  buildFindManyCurlSnippet,
  buildModelApiAiPrompt,
  buildModelRuntimeEndpoint,
  type ModelApiDocContext,
} from './model-api-docs'
import { Button } from '@web/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'

interface ModelApiDocsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  context: ModelApiDocContext | null
}

function CopyButton({ label, text }: { label: string; text: string }) {
  const [copied, setCopied] = useState(false)

  return (
    <Button
      variant="outline"
      size="sm"
      className="h-7 gap-1.5 px-2.5 text-xs"
      onClick={() => {
        void navigator.clipboard.writeText(text).then(() => {
          setCopied(true)
          setTimeout(() => setCopied(false), 2000)
        })
      }}
    >
      {copied ? <Check className="size-3.5 text-emerald-500" /> : <Copy className="size-3.5" />}
      {copied ? '已复制' : label}
    </Button>
  )
}

export function ModelApiDocsDialog({
  open,
  onOpenChange,
  context,
}: ModelApiDocsDialogProps) {
  const endpoint = context ? buildModelRuntimeEndpoint(context) : ''
  const curlSnippet = context ? buildFindManyCurlSnippet(context) : ''
  const aiPrompt = context ? buildModelApiAiPrompt(context) : ''

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileCode2 className="size-4 text-primary" />
            API 文档
          </DialogTitle>
          <DialogDescription>
            当前模型的 Runtime GraphQL 调用说明。
          </DialogDescription>
        </DialogHeader>

        {!context ? (
          <div className="rounded-md border bg-muted/40 p-4 text-sm text-muted-foreground">
            当前模型的运行时上下文不可用，暂时无法生成 API 文档。
          </div>
        ) : (
          <div className="space-y-5">
            {/* five fixed sections here */}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
```

Implementation details for the five fixed sections:

- `Server URL`: show `API_DOC_SERVER_URL` with a `复制 Server URL` button
- `URL 含义`: show `endpoint` plus 4 bullet explanations for `org/project/db/model`
- `Token 怎么填`: show `Authorization: Bearer <API_TOKEN>` plus a `复制 Header` button
- `curl 示例`: show `curlSnippet` plus a `复制 curl` button
- `如何让 AI 帮你继续写`: show `aiPrompt` plus a `复制 Prompt` button

- [ ] **Step 4: Run the dialog tests and then the builder tests**

Run:

```bash
cd modelcraft-front
npx vitest run src/web/components/features/end-user-data/ModelApiDocsDialog.test.tsx src/web/components/features/end-user-data/model-api-docs.test.ts
```

Expected: PASS with dialog and builder tests both green.

- [ ] **Step 5: Commit the dialog component**

```bash
git add modelcraft-front/src/web/components/features/end-user-data/ModelApiDocsDialog.tsx modelcraft-front/src/web/components/features/end-user-data/ModelApiDocsDialog.test.tsx modelcraft-front/src/web/components/features/end-user-data/model-api-docs.ts
git commit -m "feat: add model scoped runtime api docs dialog"
```

---

### Task 3: Integrate the API docs trigger into the end-user record workspace

**Files:**
- Modify: `modelcraft-front/src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx`
- Test: `modelcraft-front/src/web/components/features/end-user-data/ModelApiDocsDialog.test.tsx`

- [ ] **Step 1: Add a failing integration-oriented render test for the missing trigger behavior**

Extend `ModelApiDocsDialog.test.tsx` or add a narrow workspace interaction test with this shape:

```tsx
it('does not open the dialog with missing runtime context', () => {
  render(
    <ModelApiDocsDialog
      open
      onOpenChange={() => {}}
      context={null}
    />
  )

  expect(
    screen.getByText('当前模型的运行时上下文不可用，暂时无法生成 API 文档。')
  ).toBeInTheDocument()
})
```

This test should already fail if the fallback branch was not implemented in Task 2.

- [ ] **Step 2: Run the focused test file to verify the failure**

Run:

```bash
cd modelcraft-front
npx vitest run src/web/components/features/end-user-data/ModelApiDocsDialog.test.tsx
```

Expected: FAIL until the incomplete-context branch exists.

- [ ] **Step 3: Wire the trigger and dialog into `EndUserRecordWorkspace.tsx`**

Add imports near the existing toolbar imports:

```tsx
import { BookOpen } from 'lucide-react'
import { ModelApiDocsDialog } from './ModelApiDocsDialog'
import type { ModelApiDocContext } from './model-api-docs'
```

Add local dialog state next to the existing sheet/dialog state:

```tsx
const [apiDocsOpen, setApiDocsOpen] = useState(false)
```

Derive the model doc context after `model` is resolved:

```tsx
const apiDocContext = useMemo<ModelApiDocContext | null>(() => {
  if (!model?.databaseName || !model?.name) return null

  return {
    orgName,
    projectSlug,
    databaseName: model.databaseName,
    modelName: model.name,
  }
}, [orgName, projectSlug, model?.databaseName, model?.name])
```

Add the trigger button into the right side of the workspace toolbar:

```tsx
<Button
  variant="outline"
  size="sm"
  className="h-[26px] gap-1.5 px-2.5 text-xs"
  onClick={() => setApiDocsOpen(true)}
  disabled={!apiDocContext}
>
  <BookOpen className="size-3.5" />
  <span>API 文档</span>
</Button>
```

Mount the dialog near the existing delete dialog:

```tsx
<ModelApiDocsDialog
  open={apiDocsOpen}
  onOpenChange={setApiDocsOpen}
  context={apiDocContext}
/>
```

- [ ] **Step 4: Run targeted tests and lint-style verification**

Run:

```bash
cd modelcraft-front
npx vitest run src/web/components/features/end-user-data/model-api-docs.test.ts src/web/components/features/end-user-data/ModelApiDocsDialog.test.tsx
npm run lint
```

Expected:

- Vitest: PASS
- Next lint: PASS with no new errors in `EndUserRecordWorkspace.tsx`

- [ ] **Step 5: Commit the workspace integration**

```bash
git add modelcraft-front/src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx modelcraft-front/src/web/components/features/end-user-data/ModelApiDocsDialog.tsx modelcraft-front/src/web/components/features/end-user-data/ModelApiDocsDialog.test.tsx
git commit -m "feat: add model scoped api docs entry to end user workspace"
```

---

## Self-Review

### Spec coverage

- Record-page scoped entry point: covered in Task 3
- Fixed server URL: covered in Task 1 + Task 2
- Full runtime URL meaning: covered in Task 2
- `Authorization: Bearer <API_TOKEN>` guidance: covered in Task 2
- Single working `findMany` curl example: covered in Task 1 + Task 2
- AI prompt section: covered in Task 1 + Task 2
- No GraphiQL / no Python / no field reference: preserved by scope in all tasks

### Placeholder scan

- No `TODO`/`TBD`
- No “write tests” without concrete test code
- No unnamed files or generic commands

### Type consistency

- Shared context type is `ModelApiDocContext`
- Dialog prop is `context: ModelApiDocContext | null`
- Pure builders accept `ModelApiDocContext`
- Workspace derives `apiDocContext` from loaded `model.databaseName` and `model.name`

---

Plan complete and saved to `docs/superpowers/plans/2026-06-06-model-scoped-runtime-api-doc.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
