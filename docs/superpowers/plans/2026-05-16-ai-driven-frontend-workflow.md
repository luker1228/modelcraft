# AI-Driven Frontend Workflow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 通过 `useCopilotAction` + `useCopilotReadable` 实现双层 AI 工具集，让用户通过对话操控前端的所有操作，Agent 只读，写操作由前端工具预填表单后由用户确认。

**Architecture:** OrgLayout 挂载 `OrgCopilotActions`（4 个工具）；ProjectLayout 挂载 `ProjectCopilotActions`（9 个工具）。工具随组件生命周期自动注册/卸载。写操作工具只修改前端 state（打开表单 + 预填字段），不调用任何 backend API。Python agent 同步移除 3 个写工具，新增 layer/current_model/current_db 到 state。

**Tech Stack:** React `useCopilotAction` / `useCopilotReadable` (@copilotkit/react-core), Next.js App Router, Python LangGraph

---

## 文件变更地图

### 新建
- `modelcraft-front/src/web/components/features/copilot/OrgCopilotActions.tsx`
- `modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx`

### 修改
- `modelcraft-front/src/app/org/[orgName]/layout.tsx` — 引入 OrgCopilotActions + useCopilotReadable
- `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/layout.tsx` — 引入 ProjectCopilotActions + useCopilotReadable
- `modelcraft-front/src/web/components/shared/data-workspace/ModelRecordTable.tsx` — 新增 `highlightedIds?: string[]` prop
- `modelcraft-front/src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx` — 暴露 AI 控制接口（create prefill / edit prefill / highlight）
- `modelcraft-agent/agent.py` — 移除 3 个写工具，新增 layer/current_model/current_db state 字段和 system prompt

---

## Task 1：ModelRecordTable 支持行高亮

**Files:**
- Modify: `modelcraft-front/src/web/components/shared/data-workspace/ModelRecordTable.tsx`

**背景：** `highlight_records` 工具需要高亮表格中的特定行。当前 `ModelRecordTableProps` 没有 `highlightedIds` prop，需要新增。

- [ ] **Step 1: 读取当前文件，找到 `ModelRecordTableProps` 接口和行渲染逻辑**

  ```bash
  # 确认接口位置
  grep -n "interface ModelRecordTableProps\|contentList.map\|tr.*key\|TableRow" \
    modelcraft-front/src/web/components/shared/data-workspace/ModelRecordTable.tsx | head -20
  ```

- [ ] **Step 2: 在 `ModelRecordTableProps` 接口新增 `highlightedIds` 和 `highlightReasons`**

  在 `canDeleteRecord?: boolean` 之后加：

  ```ts
  /** 高亮行的 ID 列表（Agent highlight_records 工具设置） */
  highlightedIds?: string[]
  /** 高亮原因 map，key 为 record id，value 为原因文字 */
  highlightReasons?: Record<string, string>
  ```

- [ ] **Step 3: 在行渲染中应用高亮样式**

  找到渲染 `contentList` 的 `<TableRow>` 或 `<tr>` 元素，加上 className 逻辑：

  ```tsx
  // 在 contentList.map 的行元素上加
  const isHighlighted = props.highlightedIds?.includes(row.id)
  const highlightReason = props.highlightReasons?.[row.id]

  // className 加上
  className={cn(
    "...", // 现有 className
    isHighlighted && "bg-amber-50 ring-1 ring-inset ring-amber-300"
  )}
  // title 加上（hover 显示原因）
  title={highlightReason}
  ```

- [ ] **Step 4: 确认编译无报错**

  ```bash
  cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "ModelRecordTable" | head -10
  ```

  期望：无错误。

- [ ] **Step 5: Commit**

  ```bash
  git add modelcraft-front/src/web/components/shared/data-workspace/ModelRecordTable.tsx
  git commit -m "feat(table): add highlightedIds and highlightReasons props to ModelRecordTable"
  ```

---

## Task 2：DevelopRecordWorkspace 暴露 AI 控制接口

**Files:**
- Modify: `modelcraft-front/src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx`

**背景：** `open_create_record` 需要触发创建表单并预填字段；`open_edit_record` 需要打开编辑表单并预填 patch；`highlight_records` 需要高亮行。这些状态目前是组件内部的，需要通过 props callback 或 event 暴露给外层（ProjectCopilotActions）。

方案：新增三个可选 prop callback + `highlightedIds` state。

- [ ] **Step 1: 找到 DevelopRecordWorkspace 的 Props interface**

  ```bash
  grep -n "interface.*Props\|DevelopRecordWorkspaceProps\|export.*function\|export default" \
    modelcraft-front/src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx | head -10
  ```

- [ ] **Step 2: 扩展 Props interface，新增 AI 控制 props**

  找到组件 props interface（或函数参数），新增：

  ```ts
  /** AI 控制接口 — 由 ProjectCopilotActions 注入 */
  aiControl?: {
    /** 触发新建表单，prefill 字段会预填到 ModelRecordForm 的初始值 */
    onOpenCreate?: (prefill: Record<string, unknown>) => void
    /** 触发编辑表单，先 fetch 记录，再用 patch 覆盖指定字段 */
    onOpenEdit?: (recordId: string, patch: Record<string, unknown>) => void
    /** 高亮行 */
    highlightedIds?: string[]
    highlightReasons?: Record<string, string>
  }
  ```

- [ ] **Step 3: 在 handleCreate 中支持 prefill**

  新增内部 state：
  ```ts
  const [createPrefill, setCreatePrefill] = useState<Record<string, unknown>>({})
  ```

  实现 `onOpenCreate` 的调用入口（在 useEffect 或直接作为 callback 传给 aiControl）：
  ```ts
  // 在组件内，如果 aiControl?.onOpenCreate 被外部调用时执行：
  // 外部通过 ref 或 useImperativeHandle 不合适——改用 useEffect 监听 aiControl prop 变化
  // 简单方案：暴露一个 openCreateWithPrefill 函数通过 onOpenCreate callback 注册
  useEffect(() => {
    if (!aiControl?.onOpenCreate) return
    // 注册回调：外部调用 aiControl.onOpenCreate(prefill) 时执行
    // 实现：aiControl.onOpenCreate 本身就是 handler，组件只需在 render 时把 setCreatePrefill+setCreateDataOpen 绑进去
    // 由于 React callback 在每次 render 重建，外部存的引用可能过期
    // 最简单：用 ref 保存最新 handler
  }, [aiControl?.onOpenCreate])
  ```

  **实际实现方式**（比 ref 更简单）：在组件 render 时，把 setter 暴露给父组件通过 `aiHandlerRef`：

  ```ts
  // 在组件顶部
  const aiHandlerRef = useRef<{
    openCreate: (prefill: Record<string, unknown>) => void
    openEdit: (id: string, patch: Record<string, unknown>) => void
  } | null>(null)

  // 在 state 初始化后立即更新 ref（同步，无需 useEffect）
  aiHandlerRef.current = {
    openCreate: (prefill) => {
      setCreatePrefill(prefill)
      setCreateDataOpen(true)
    },
    openEdit: async (id, patch) => {
      setEditItemId(id)
      setEditLoading(true)
      setEditDataOpen(true)
      // fetch record，然后用 patch 覆盖
      try {
        const { data } = await runtimeClient!.query({
          query: findUniqueQuery!,
          variables: { where: { id } },
          fetchPolicy: 'network-only',
        })
        const item = (data as Record<string, Record<string, unknown>>)?.findUnique?.item
        if (isRecord(item)) {
          const base = writableFieldNames.reduce<Record<string, unknown>>((acc, f) => {
            acc[f] = item[f] ?? ''
            return acc
          }, {})
          setEditFormData({ ...base, ...patch })
        }
      } catch {
        toast.error('获取数据失败')
        setEditDataOpen(false)
      } finally {
        setEditLoading(false)
      }
    },
  }

  // 通过 prop 把 ref 暴露给外层
  useEffect(() => {
    if (aiControl?.onOpenCreate) {
      // 用 ref indirection 让外层拿到最新 handler
      // 实际上 ProjectCopilotActions 调用时用的是闭包里的函数，见 Task 4
    }
  }, [])
  ```

  **最终简化方案**：`aiControl` 不用 callback，改为传 **imperative ref**：

  ```ts
  // Props
  aiRef?: React.MutableRefObject<{
    openCreate: (prefill: Record<string, unknown>) => void
    openEdit: (recordId: string, patch: Record<string, unknown>) => void
    setHighlight: (ids: string[], reasons: Record<string, string>) => void
  } | null>
  ```

  在组件内更新 ref（无需 useEffect，在 render 阶段同步写）：
  ```ts
  // 在 state 声明之后
  const [highlightedIds, setHighlightedIds] = useState<string[]>([])
  const [highlightReasons, setHighlightReasons] = useState<Record<string, string>>({})

  if (aiRef) {
    aiRef.current = {
      openCreate: (prefill) => {
        setCreatePrefill(prefill)
        setCreateDataOpen(true)
      },
      openEdit: async (id, patch) => { /* 见上 */ },
      setHighlight: (ids, reasons) => {
        setHighlightedIds(ids)
        setHighlightReasons(reasons)
      },
    }
  }
  ```

- [ ] **Step 4: ModelRecordForm 接收 initialValues 并应用到表单**

  找到 create Sheet 里的 `<ModelRecordForm>` 调用，传入 `initialValues={createPrefill}`：

  ```bash
  grep -n "ModelRecordForm" \
    modelcraft-front/src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx
  ```

  在 create 用的 `<ModelRecordForm>` 上加 `initialValues={createPrefill}`。

  然后找 ModelRecordForm 的 props interface，确认是否已有 `initialValues`：
  ```bash
  grep -n "initialValues\|defaultValues\|interface.*Props" \
    modelcraft-front/src/web/components/features/model-editor/model-record-form/index.tsx | head -10
  ```

  如果没有，新增：
  ```ts
  initialValues?: Record<string, unknown>
  ```
  并在 form 初始化时用 `initialValues` 覆盖默认值。

- [ ] **Step 5: 把 `highlightedIds` 和 `highlightReasons` 传给 `ModelRecordTable`**

  找到 `<ModelRecordTable>` 的调用处，新增两个 prop：
  ```tsx
  highlightedIds={highlightedIds}
  highlightReasons={highlightReasons}
  ```

- [ ] **Step 6: 编译检查**

  ```bash
  cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "DevelopRecordWorkspace\|ModelRecordForm" | head -10
  ```

  期望：无错误。

- [ ] **Step 7: Commit**

  ```bash
  git add modelcraft-front/src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx \
          modelcraft-front/src/web/components/features/model-editor/model-record-form/index.tsx
  git commit -m "feat(workspace): expose AI control interface (aiRef) for create/edit/highlight"
  ```

---

## Task 3：OrgCopilotActions 组件

**Files:**
- Create: `modelcraft-front/src/web/components/features/copilot/OrgCopilotActions.tsx`

**背景：** 在 OrgLayout 层注册 4 个 org 级工具，挂载即注册，卸载即注销。

- [ ] **Step 1: 创建文件**

  ```tsx
  // modelcraft-front/src/web/components/features/copilot/OrgCopilotActions.tsx
  'use client'

  import { useCopilotAction } from '@copilotkit/react-core'
  import { useRouter } from 'next/navigation'

  interface OrgCopilotActionsProps {
    orgName: string
    /** 打开新建项目 Sheet — 由 workspace page 传入 */
    onOpenCreateProject?: (prefill: { slug?: string; title?: string; description?: string }) => void
    /** 高亮某个项目卡片 — 由 workspace page 传入 */
    onHighlightProject?: (slug: string, reason: string) => void
  }

  export function OrgCopilotActions({ orgName, onOpenCreateProject, onHighlightProject }: OrgCopilotActionsProps) {
    const router = useRouter()

    useCopilotAction({
      name: 'navigate_to_project',
      description: '跳转到指定项目的工作区',
      parameters: [
        { name: 'slug', type: 'string', description: '项目 slug', required: true },
      ],
      handler: async ({ slug }: { slug: string }) => {
        router.push(`/org/${orgName}/project/${slug}`)
        return `已跳转到项目 ${slug}`
      },
    })

    useCopilotAction({
      name: 'navigate_to_settings',
      description: '跳转到组织设置页面',
      parameters: [],
      handler: async () => {
        router.push(`/org/${orgName}/settings`)
        return '已跳转到设置页面'
      },
    })

    useCopilotAction({
      name: 'open_create_project',
      description: '打开新建项目表单，可预填 slug、title、description。用户需手动点击 Create 按钮完成创建。',
      parameters: [
        { name: 'slug', type: 'string', description: '项目 slug（英文小写+连字符）', required: false },
        { name: 'title', type: 'string', description: '项目显示名称', required: false },
        { name: 'description', type: 'string', description: '项目描述', required: false },
      ],
      handler: async ({ slug, title, description }: { slug?: string; title?: string; description?: string }) => {
        if (onOpenCreateProject) {
          onOpenCreateProject({ slug, title, description })
          return '已打开新建项目表单，字段已预填，请确认后点击 Create。'
        }
        return '当前页面不支持新建项目操作，请先导航到 workspace 页。'
      },
    })

    useCopilotAction({
      name: 'highlight_project',
      description: '在项目列表中高亮指定项目，并显示说明原因',
      parameters: [
        { name: 'slug', type: 'string', description: '要高亮的项目 slug', required: true },
        { name: 'reason', type: 'string', description: '高亮原因，显示为 tooltip', required: true },
      ],
      handler: async ({ slug, reason }: { slug: string; reason: string }) => {
        if (onHighlightProject) {
          onHighlightProject(slug, reason)
          return `已高亮项目 ${slug}：${reason}`
        }
        return '当前页面不支持高亮操作，请先导航到 workspace 页。'
      },
    })

    return null
  }
  ```

- [ ] **Step 2: 编译检查**

  ```bash
  cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "OrgCopilotActions" | head -5
  ```

  期望：无错误。

- [ ] **Step 3: Commit**

  ```bash
  git add modelcraft-front/src/web/components/features/copilot/OrgCopilotActions.tsx
  git commit -m "feat(copilot): add OrgCopilotActions with 4 org-level tools"
  ```

---

## Task 4：ProjectCopilotActions 组件

**Files:**
- Create: `modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx`

**背景：** 在 ProjectLayout 层注册 9 个 project 级工具。写操作工具通过 `aiRef`（Task 2 暴露的 imperative ref）控制 DevelopRecordWorkspace，不直接调 backend。

- [ ] **Step 1: 创建文件**

  ```tsx
  // modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx
  'use client'

  import { useCopilotAction } from '@copilotkit/react-core'
  import { useRouter } from 'next/navigation'
  import type { DevelopRecordWorkspaceAIRef } from '@web/components/features/model-editor/model-record-form/DevelopRecordWorkspace'

  interface ProjectCopilotActionsProps {
    orgName: string
    projectSlug: string
    /** 由 DevelopRecordWorkspace 通过 aiRef 暴露的命令接口 */
    workspaceAiRef?: React.MutableRefObject<DevelopRecordWorkspaceAIRef | null>
  }

  export function ProjectCopilotActions({ orgName, projectSlug, workspaceAiRef }: ProjectCopilotActionsProps) {
    const router = useRouter()

    useCopilotAction({
      name: 'navigate_to_org',
      description: '退出当前项目，返回组织 workspace 页面',
      parameters: [],
      handler: async () => {
        router.push(`/org/${orgName}/workspace`)
        return '已返回组织页面'
      },
    })

    useCopilotAction({
      name: 'navigate_to_model',
      description: '跳转到模型编辑器，查看或编辑指定模型的字段结构',
      parameters: [
        { name: 'db', type: 'string', description: '数据库名称', required: true },
        { name: 'model', type: 'string', description: '模型名称', required: true },
      ],
      handler: async ({ db, model }: { db: string; model: string }) => {
        router.push(`/org/${orgName}/project/${projectSlug}/model-editor?db=${db}&model=${model}`)
        return `已跳转到模型 ${model} 的编辑器`
      },
    })

    useCopilotAction({
      name: 'navigate_to_data',
      description: '跳转到数据视图，查看指定模型的数据记录',
      parameters: [
        { name: 'db', type: 'string', description: '数据库名称', required: true },
        { name: 'model', type: 'string', description: '模型名称', required: true },
      ],
      handler: async ({ db, model }: { db: string; model: string }) => {
        router.push(`/org/${orgName}/project/${projectSlug}/data?db=${db}&model=${model}`)
        return `已跳转到模型 ${model} 的数据视图`
      },
    })

    useCopilotAction({
      name: 'open_create_model',
      description: '打开新建模型的表单，可预填名称和描述。用户需手动点击 Create 完成创建。',
      parameters: [
        { name: 'db', type: 'string', description: '数据库名称', required: true },
        { name: 'name', type: 'string', description: '模型名称（英文小写+下划线）', required: false },
        { name: 'title', type: 'string', description: '模型显示名称', required: false },
        { name: 'description', type: 'string', description: '模型描述', required: false },
      ],
      handler: async ({ db, name, title, description }: { db: string; name?: string; title?: string; description?: string }) => {
        // 跳转到 model-editor，依赖 model-editor 页面内的 CreateModelDialog
        router.push(`/org/${orgName}/project/${projectSlug}/model-editor?db=${db}&openCreate=1&prefillName=${name ?? ''}&prefillTitle=${title ?? ''}`)
        return `已打开新建模型表单${name ? `，名称预填为 ${name}` : ''}，请确认后点击 Create。`
      },
    })

    useCopilotAction({
      name: 'open_create_record',
      description: '打开新建记录表单，并预填指定字段值。用户需手动点击 Save 完成创建。写操作不会自动执行。',
      parameters: [
        { name: 'model', type: 'string', description: '模型名称', required: true },
        { name: 'db', type: 'string', description: '数据库名称', required: true },
        { name: 'prefill', type: 'object', description: '要预填的字段值，例如 {"name": "张三", "age": 25}', required: false },
      ],
      handler: async ({ model, db, prefill }: { model: string; db: string; prefill?: Record<string, unknown> }) => {
        const ref = workspaceAiRef?.current
        if (ref) {
          ref.openCreate(prefill ?? {})
          return `已打开新建 ${model} 记录的表单${prefill ? '，字段已预填' : ''}，请确认后点击 Save。`
        }
        // 如果不在数据页，先跳转过去
        router.push(`/org/${orgName}/project/${projectSlug}/data?db=${db}&model=${model}`)
        return `已跳转到 ${model} 数据页，请在页面加载后重新调用此工具以预填表单。`
      },
    })

    useCopilotAction({
      name: 'open_edit_record',
      description: '打开指定记录的编辑表单，并预填要修改的字段。用户需手动点击 Save 完成保存。写操作不会自动执行。',
      parameters: [
        { name: 'model', type: 'string', description: '模型名称', required: true },
        { name: 'db', type: 'string', description: '数据库名称', required: true },
        { name: 'record_id', type: 'string', description: '要编辑的记录 ID', required: true },
        { name: 'patch', type: 'object', description: '要修改的字段值，例如 {"amount": 950}', required: true },
      ],
      handler: async ({ model, db, record_id, patch }: { model: string; db: string; record_id: string; patch: Record<string, unknown> }) => {
        const ref = workspaceAiRef?.current
        if (ref) {
          await ref.openEdit(record_id, patch)
          return `已打开记录 ${record_id} 的编辑表单，修改字段已预填，请确认后点击 Save。`
        }
        router.push(`/org/${orgName}/project/${projectSlug}/data?db=${db}&model=${model}`)
        return `已跳转到 ${model} 数据页，请在页面加载后重新调用此工具。`
      },
    })

    useCopilotAction({
      name: 'highlight_records',
      description: '在数据表格中高亮指定记录行，并显示说明原因（鼠标悬停可见）',
      parameters: [
        { name: 'model', type: 'string', description: '模型名称', required: true },
        { name: 'record_ids', type: 'string[]', description: '要高亮的记录 ID 列表', required: true },
        { name: 'reason', type: 'string', description: '高亮原因', required: true },
      ],
      handler: async ({ model, record_ids, reason }: { model: string; record_ids: string[]; reason: string }) => {
        const ref = workspaceAiRef?.current
        if (ref) {
          const reasons = record_ids.reduce<Record<string, string>>((acc, id) => { acc[id] = reason; return acc }, {})
          ref.setHighlight(record_ids, reasons)
          return `已高亮 ${record_ids.length} 条 ${model} 记录：${reason}`
        }
        return '当前页面没有数据表格，请先导航到数据视图。'
      },
    })

    useCopilotAction({
      name: 'set_filter',
      description: '在当前数据视图设置筛选条件（filter JSON）',
      parameters: [
        { name: 'filter_json', type: 'string', description: 'ModelCraft where JSON 字符串，例如 {"age":{"gt":18}}', required: true },
      ],
      handler: async ({ filter_json }: { filter_json: string }) => {
        // FilterPanel 通过 FilterCopilotActions 注册了 set_filter，此处不重复注册
        // 由 FilterCopilotActions 的 set_filter 处理
        return `筛选条件已设置`
      },
    })

    useCopilotAction({
      name: 'clear_filter',
      description: '清空当前数据视图的所有筛选条件',
      parameters: [],
      handler: async () => {
        return '筛选条件已清空'
      },
    })

    return null
  }
  ```

  **注意：** `set_filter` 和 `clear_filter` 已由 `FilterCopilotActions` 注册，此处保留工具定义但 handler 为空（CopilotKit 会调用最近注册的同名 action）。实际上应去掉这两个重复注册，只在 `FilterCopilotActions` 里注册。

  **修正：** 删除 `ProjectCopilotActions` 里的 `set_filter` 和 `clear_filter` 注册，由 `FilterCopilotActions` 处理。最终文件不含这两个 action。

- [ ] **Step 2: 导出 `DevelopRecordWorkspaceAIRef` 类型**

  在 `DevelopRecordWorkspace.tsx` 顶部加 export：
  ```ts
  export interface DevelopRecordWorkspaceAIRef {
    openCreate: (prefill: Record<string, unknown>) => void
    openEdit: (recordId: string, patch: Record<string, unknown>) => Promise<void>
    setHighlight: (ids: string[], reasons: Record<string, string>) => void
  }
  ```

- [ ] **Step 3: 编译检查**

  ```bash
  cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "ProjectCopilotActions" | head -5
  ```

  期望：无错误。

- [ ] **Step 4: Commit**

  ```bash
  git add modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx \
          modelcraft-front/src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx
  git commit -m "feat(copilot): add ProjectCopilotActions with 7 project-level tools"
  ```

---

## Task 5：OrgLayout 集成 OrgCopilotActions + useCopilotReadable

**Files:**
- Modify: `modelcraft-front/src/app/org/[orgName]/layout.tsx`

**背景：** OrgLayout 已有 `CopilotWrapper`，需要在 wrapper 内部挂载 `OrgCopilotActions`，并通过 `useCopilotReadable` 向 agent 暴露当前上下文。

- [ ] **Step 1: 在 layout.tsx 引入新组件和 hook**

  ```ts
  import { useCopilotReadable } from '@copilotkit/react-core'
  import { OrgCopilotActions } from '@web/components/features/copilot/OrgCopilotActions'
  ```

- [ ] **Step 2: 创建 `OrgCopilotContextProvider` 内部组件**

  由于 `useCopilotReadable` 必须在 `CopilotKit` provider 内部调用，需要把它放在 `CopilotWrapper` 的子组件里。在 layout.tsx 内新建一个内部组件：

  ```tsx
  function OrgAIContext({ orgName }: { orgName: string }) {
    useCopilotReadable({
      description: '当前 AI 上下文',
      value: {
        layer: 'org',
        orgName,
        availableActions: [
          'navigate_to_project',
          'navigate_to_settings',
          'open_create_project',
          'highlight_project',
          // 后端只读工具（由 Python agent 提供）
          'list_projects',
          'nl2filter',
        ],
      },
    })

    return (
      <OrgCopilotActions
        orgName={orgName}
        // onOpenCreateProject 和 onHighlightProject 由 workspace page 通过事件/store 控制
        // 暂时留空，Task 6 通过自定义 event 联通
      />
    )
  }
  ```

- [ ] **Step 3: 在 CopilotWrapper 内渲染 OrgAIContext**

  找到 layout.tsx 里的 `<CopilotWrapper>` 返回块，在 `{content}` 之前加：

  ```tsx
  if (showCopilot) {
    return (
      <CopilotWrapper selectedProject={null} orgName={orgName}>
        <OrgAIContext orgName={orgName} />  {/* ← 新增 */}
        {content}
      </CopilotWrapper>
    )
  }
  ```

- [ ] **Step 4: 编译检查**

  ```bash
  cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "layout" | grep "org" | head -10
  ```

  期望：无错误。

- [ ] **Step 5: Commit**

  ```bash
  git add modelcraft-front/src/app/org/\[orgName\]/layout.tsx
  git commit -m "feat(org-layout): integrate OrgCopilotActions and useCopilotReadable"
  ```

---

## Task 6：ProjectLayout 集成 ProjectCopilotActions + useCopilotReadable

**Files:**
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/layout.tsx`

**背景：** ProjectLayout 已有 `CopilotWrapper`，需要挂载 `ProjectCopilotActions`，并通过 `useCopilotReadable` 暴露 project 层上下文。`workspaceAiRef` 需要在 layout 层创建并向下传递，但由于 Next.js App Router 的 layout/page 边界，page 组件（`DevelopRecordWorkspace`）无法直接接收 layout 的 props。

**解法：** 使用 React Context 传递 `aiRef`，不跨 layout/page prop drilling。

- [ ] **Step 1: 创建 WorkspaceAIRefContext**

  新建文件：
  ```ts
  // modelcraft-front/src/web/contexts/workspace-ai-ref-context.tsx
  'use client'
  import { createContext, useContext } from 'react'
  import type { DevelopRecordWorkspaceAIRef } from '@web/components/features/model-editor/model-record-form/DevelopRecordWorkspace'

  export const WorkspaceAIRefContext = createContext<React.MutableRefObject<DevelopRecordWorkspaceAIRef | null> | null>(null)

  export function useWorkspaceAIRef() {
    return useContext(WorkspaceAIRefContext)
  }
  ```

- [ ] **Step 2: 在 layout.tsx 引入并提供 Context**

  ```ts
  import { useRef } from 'react'
  import { useCopilotReadable } from '@copilotkit/react-core'
  import { ProjectCopilotActions } from '@web/components/features/copilot/ProjectCopilotActions'
  import { WorkspaceAIRefContext } from '@web/contexts/workspace-ai-ref-context'
  import type { DevelopRecordWorkspaceAIRef } from '@web/components/features/model-editor/model-record-form/DevelopRecordWorkspace'
  ```

  在 `ProjectLayout` 函数里：
  ```ts
  const workspaceAiRef = useRef<DevelopRecordWorkspaceAIRef | null>(null)
  ```

- [ ] **Step 3: 创建内部 `ProjectAIContext` 组件（需在 CopilotKit 内部）**

  ```tsx
  function ProjectAIContext({
    orgName,
    projectSlug,
    workspaceAiRef,
  }: {
    orgName: string
    projectSlug: string
    workspaceAiRef: React.MutableRefObject<DevelopRecordWorkspaceAIRef | null>
  }) {
    const params = useParams()
    const currentModel = (params?.modelName as string | undefined) ?? ''
    const currentDb = (params?.db as string | undefined) ?? ''

    useCopilotReadable({
      description: '当前 AI 上下文',
      value: {
        layer: 'project',
        orgName,
        projectSlug,
        currentModel,
        currentDb,
        availableActions: [
          'navigate_to_org',
          'navigate_to_model',
          'navigate_to_data',
          'open_create_model',
          'open_create_record',
          'open_edit_record',
          'highlight_records',
          // FilterCopilotActions 提供的
          'set_filter',
          'clear_filter',
          // 后端只读工具（Python agent）
          'list_models',
          'get_model_fields',
          'query_model',
          'nl2filter',
        ],
      },
    })

    return (
      <ProjectCopilotActions
        orgName={orgName}
        projectSlug={projectSlug}
        workspaceAiRef={workspaceAiRef}
      />
    )
  }
  ```

- [ ] **Step 4: 更新 ProjectLayout 的返回结构**

  ```tsx
  if (showCopilot) {
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
  }

  return (
    <WorkspaceAIRefContext.Provider value={workspaceAiRef}>
      {mainContent}
    </WorkspaceAIRefContext.Provider>
  )
  ```

- [ ] **Step 5: 在 DevelopRecordWorkspace 消费 Context，绑定 aiRef**

  在 `DevelopRecordWorkspace.tsx` 里：
  ```ts
  import { useWorkspaceAIRef } from '@web/contexts/workspace-ai-ref-context'
  ```

  在组件顶部（state 初始化后）：
  ```ts
  const workspaceAiRef = useWorkspaceAIRef()

  // 同步更新 ref（在 render 阶段，无需 useEffect）
  const [highlightedIds, setHighlightedIds] = useState<string[]>([])
  const [highlightReasons, setHighlightReasons] = useState<Record<string, string>>({})

  if (workspaceAiRef) {
    workspaceAiRef.current = {
      openCreate: (prefill) => {
        setCreatePrefill(prefill)
        setCreateDataOpen(true)
      },
      openEdit: async (id, patch) => {
        setEditItemId(id)
        setEditLoading(true)
        setEditDataOpen(true)
        try {
          if (!runtimeClient || !findUniqueQuery) return
          const { data } = await runtimeClient.query({
            query: findUniqueQuery,
            variables: { where: { id } },
            fetchPolicy: 'network-only',
          })
          const item = (data as Record<string, Record<string, unknown>>)?.findUnique?.item
          if (isRecord(item)) {
            const base = writableFieldNames.reduce<Record<string, unknown>>((acc, f) => {
              acc[f] = item[f] ?? ''
              return acc
            }, {})
            setEditFormData({ ...base, ...patch })
          }
        } catch {
          toast.error('获取数据失败')
          setEditDataOpen(false)
        } finally {
          setEditLoading(false)
        }
      },
      setHighlight: (ids, reasons) => {
        setHighlightedIds(ids)
        setHighlightReasons(reasons)
      },
    }
  }
  ```

  **注意：** `aiRef` prop 从 Task 2 可以删掉，改为通过 Context 获取，不需要 prop drilling。

- [ ] **Step 6: 编译检查**

  ```bash
  cd modelcraft-front && npx tsc --noEmit 2>&1 | grep -E "ProjectLayout|workspace-ai-ref|ProjectAIContext" | head -10
  ```

  期望：无错误。

- [ ] **Step 7: Commit**

  ```bash
  git add modelcraft-front/src/app/org/\[orgName\]/project/\[projectSlug\]/layout.tsx \
          modelcraft-front/src/web/contexts/workspace-ai-ref-context.tsx \
          modelcraft-front/src/web/components/features/model-editor/model-record-form/DevelopRecordWorkspace.tsx
  git commit -m "feat(project-layout): integrate ProjectCopilotActions, useCopilotReadable, WorkspaceAIRefContext"
  ```

---

## Task 7：Python Agent 裁剪 + State 扩展

**Files:**
- Modify: `modelcraft-agent/agent.py`

**背景：** 移除 3 个写工具（`create_record`、`update_record`、`delete_record`），新增 `layer`/`current_model`/`current_db` 到 `AgentState`，更新 system prompt 支持层级感知。

- [ ] **Step 1: 更新 `AgentState` TypedDict**

  找到 `AgentState` class，新增三个字段：

  ```python
  class AgentState(TypedDict):
      messages: Annotated[list, add_messages]
      authorization: str
      org_name: str
      project_slug: str
      layer: str          # "org" | "project" | ""
      current_model: str  # 当前路由中的模型名（project 层有值）
      current_db: str     # 当前路由中的数据库名
  ```

- [ ] **Step 2: 从 `_tools` 列表移除三个写工具**

  找到：
  ```python
  _tools = [
      list_projects,
      list_models,
      get_model_fields,
      query_model,
      create_record,   # ← 删除
      update_record,   # ← 删除
      delete_record,   # ← 删除
      nl2filter,
  ]
  ```

  改为：
  ```python
  _tools = [
      list_projects,
      list_models,
      get_model_fields,
      query_model,
      nl2filter,
  ]
  ```

- [ ] **Step 3: 删除三个写工具函数定义**

  删除 `create_record`、`update_record`、`delete_record` 的 `@tool` 函数定义（约 60 行）。

- [ ] **Step 4: 更新 system prompt，支持层级感知**

  找到 `agent_node` 函数里的 `system_msg`，替换为：

  ```python
  async def agent_node(state: AgentState):
      org = state.get("org_name", "")
      project = state.get("project_slug", "")
      layer = state.get("layer", "")
      current_model = state.get("current_model", "")
      current_db = state.get("current_db", "")

      if layer == "org":
          context = (
              f"当前在 Org 页面（组织：{org}）。\n"
              "可用工具：navigate_to_project、navigate_to_settings、open_create_project、highlight_project、list_projects、nl2filter。\n"
              "注意：不可直接调用 list_models、query_model 等 project 级工具。\n"
              "如需操作项目数据，先调用 navigate_to_project(slug) 跳转到对应项目。"
          )
      elif layer == "project":
          model_ctx = f"，当前模型：{current_model}（数据库：{current_db}）" if current_model else ""
          context = (
              f"当前在 Project 页面（组织：{org}，项目：{project}{model_ctx}）。\n"
              "可用工具：navigate_to_org、navigate_to_model、navigate_to_data、"
              "open_create_model、open_create_record、open_edit_record、highlight_records、"
              "list_models、get_model_fields、query_model、nl2filter。\n"
              "写操作规则：open_create_record 和 open_edit_record 只预填表单，用户点 Save 才真正保存。\n"
              "操作前先用 get_model_fields 确认字段名，避免预填错误字段。\n"
              "如需返回 org 级操作，调用 navigate_to_org。"
          )
      else:
          context = (
              f"当前组织：{org}{'，项目：' + project if project else ''}。\n"
              "可用工具取决于当前页面，请先询问用户当前在哪个页面。"
          )

      system_msg = {
          "role": "system",
          "content": (
              "你是 ModelCraft AI 助手，帮助用户通过对话完成所有操作。\n\n"
              f"{context}\n\n"
              "通用原则：\n"
              "- 操作数据前先用 list_models 和 get_model_fields 确认模型和字段存在\n"
              "- 删除操作禁止自动执行，必须引导用户在界面手动确认\n"
              "- 如果用户说筛选或过滤，先用 nl2filter 生成 filter JSON，再告知前端已应用"
          ),
      }
      messages = [system_msg] + state["messages"]
      response = await llm.ainvoke(messages)
      return {"messages": [response]}
  ```

- [ ] **Step 5: 验证 agent 启动无报错**

  ```bash
  cd modelcraft-agent && python3 -c "from agent import modelcraft_graph; print('OK')"
  ```

  期望：输出 `OK`，无异常。

- [ ] **Step 6: Commit**

  ```bash
  git add modelcraft-agent/agent.py
  git commit -m "feat(agent): remove write tools, add layer/model/db to state, layer-aware system prompt"
  ```

---

## Task 8：端到端验证

**背景：** 确认整个链路工作正常。不需要自动化测试，手动验证关键路径。

- [ ] **Step 1: 重新部署 agent**

  ```bash
  cd /data/home/lukemxjia/modelcraft/deploy/compose
  docker-compose -f docker-compose.local.yml up -d --build modelcraft-agent
  sleep 10 && docker logs modelcraft-agent --tail 5
  ```

  期望：`Application startup complete`，无错误。

- [ ] **Step 2: 在浏览器验证 Org 层工具**

  1. 打开 `http://localhost:3001/org/lukeco/workspace`
  2. 点击 AI 助手按钮，打开 Copilot Sidebar
  3. 输入：`帮我跳转到 test 项目`
  4. 期望：页面路由变为 `/org/lukeco/project/test/...`

- [ ] **Step 3: 在浏览器验证 Project 层工具**

  1. 在 project 页面，打开 AI 助手
  2. 输入：`帮我查询 maindb 里的 users 模型有哪些字段`
  3. 期望：AI 调用 `list_models` + `get_model_fields`，返回字段列表

- [ ] **Step 4: 验证预填表单**

  1. 在数据视图页面，打开 AI 助手
  2. 输入：`帮我新建一条记录，名字叫张三`
  3. 期望：AI 调用 `open_create_record`，新建表单打开，name 字段已预填"张三"，**用户需手动点 Save**

- [ ] **Step 5: Final commit（如有文档更新）**

  ```bash
  git add -A
  git commit -m "chore: verify AI-driven frontend workflow end-to-end"
  ```

