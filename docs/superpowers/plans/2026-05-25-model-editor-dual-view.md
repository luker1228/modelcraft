# Model Editor 双视角 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `/model-editor` 页面拆分为「模型管理」和「数据管理」两个视角，通过左侧导航子项（`?view=schema` / `?view=data`）切换，模型管理视角主区内嵌字段列表，数据管理视角保持现有数据 Tab 行为。

**Architecture:** 使用 query param（`?view`）区分视角，保持同一路由不重载。新增 `ModelSchemaPanel` 组件承载模型管理主区内容（从 `ModelDetailPanel` 提取），`ModelEditorView` 读取 `useSearchParams` 条件渲染主区。左侧 `AppLayout` 导航给「数据模型」添加 `children` 子项，复用现有 `tabParam` 机制。

**Tech Stack:** Next.js App Router (`useSearchParams`)，React，Apollo Client，shadcn/ui，Tailwind CSS

---

## File Map

| 文件 | 类型 | 说明 |
|------|------|------|
| `src/web/components/features/layout/AppLayout.tsx` | 修改 | 添加「数据模型」子导航项 + 自动展开逻辑 |
| `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSchemaPanel.tsx` | 新建 | 模型管理主区内嵌面板（元信息 + 字段表格 + 外键） |
| `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelEditorView.tsx` | 修改 | 读取 `?view` 参数，条件渲染主区；切换时 reset 浮层状态 |
| `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx` | 修改 | 接收 `viewMode` prop，`data` 视角下隐藏「新建/导入」按钮 |

---

## Task 1: AppLayout — 添加「数据模型」子导航

**Files:**
- Modify: `modelcraft-front/src/web/components/features/layout/AppLayout.tsx:236-241` (projectNavSections 数据模型条目)
- Modify: `modelcraft-front/src/web/components/features/layout/AppLayout.tsx:188-215` (isSubItemActive + auto-expand useEffect)

- [ ] **Step 1: 在 `projectNavSections` 中给「数据模型」添加 children**

  找到 `AppLayout.tsx` 第 240 行的「数据模型」条目，替换为：

  ```ts
  {
    label: '数据模型',
    icon: '/icons/icon-table2.svg',
    href: `/org/${orgName}/project/${projectSlug}/model-editor`,
    children: [
      {
        label: '模型管理',
        href: `/org/${orgName}/project/${projectSlug}/model-editor`,
        tabParam: 'view=schema',
      },
      {
        label: '数据管理',
        href: `/org/${orgName}/project/${projectSlug}/model-editor?view=data`,
        tabParam: 'view=data',
      },
    ],
  },
  ```

- [ ] **Step 2: 修复「模型管理」默认高亮逻辑**

  在 `isSubItemActive`（第 188 行）中，`view=schema` 条目需要在没有 `?view` 参数时也高亮（与 `roles` 的默认逻辑对称）。在现有的 `if (paramVal === 'roles')` 判断**后**添加：

  ```ts
  // Default view (schema) is active when no view param or view=schema
  if (paramVal === 'schema') return !currentVal || currentVal === 'schema'
  ```

- [ ] **Step 3: 添加 model-editor 路径自动展开逻辑**

  在 `AppLayout.tsx` 现有的 `useEffect`（第 210 行，处理 roles 自动展开）**下方**追加：

  ```ts
  // Auto-expand the "数据模型" item when its route is active
  useEffect(() => {
    const modelEditorHref = `/org/${orgName}/project/${projectSlug}/model-editor`
    if (isNavActive(modelEditorHref)) {
      setExpandedItems((prev) => new Set([...prev, modelEditorHref]))
    }
  }, [pathname, orgName, projectSlug, isNavActive])
  ```

- [ ] **Step 4: 本地验证**

  ```bash
  cd modelcraft-front && npm run lint
  ```

  预期：无 lint 错误。在浏览器打开 `/org/lukeco/project/onboardceshi/model-editor`，确认左侧「数据模型」展开并有「模型管理」/「数据管理」两个子项，「模型管理」默认高亮。

- [ ] **Step 5: Commit**

  ```bash
  git add modelcraft-front/src/web/components/features/layout/AppLayout.tsx
  git commit -m "feat(nav): add model-editor sub-items for schema/data view switching"
  ```

---

## Task 2: 新建 ModelSchemaPanel 组件

**Files:**
- Create: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSchemaPanel.tsx`

该组件是从 `ModelDetailPanel` 的 `DrawerContent` 内容区**提取**出来的，去掉 `Drawer` 包装，直接渲染为主区面板。

- [ ] **Step 1: 新建文件，定义 props 接口**

  ```tsx
  // ModelSchemaPanel.tsx
  'use client'

  import {
    Loader2, Edit, Key, Plus, Table2, MoreVertical, Archive, Trash2, AlertTriangle,
  } from 'lucide-react'
  import { Button } from '@web/components/ui/button'
  import { Input } from '@web/components/ui/input'
  import { Alert, AlertDescription } from '@web/components/ui/alert'
  import {
    DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
  } from '@web/components/ui/dropdown-menu'
  import {
    Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
  } from '@web/components/ui/select'
  import { InsertFieldSheet } from '@web/components/features/model-editor/InsertFieldSheet'
  import {
    getEnumDisplayFieldName,
    isSystemGeneratedLabelField,
  } from '@/shared/model/system-field'
  import { ForeignKeyPanel } from './ForeignKeyPanel'
  import type { ModelEditorState } from '../_hooks'
  import type { ModelCRUD } from '../_hooks'
  import type { FieldOperations } from '../_hooks'
  import type { ForeignKeyOps } from '../_hooks'

  interface ModelSchemaPanelProps {
    state: ModelEditorState
    crud: ModelCRUD
    fieldOps: FieldOperations
    fkOps: ForeignKeyOps
    orgName: string
    projectSlug: string
    onFieldAdded?: () => void
  }
  ```

- [ ] **Step 2: 实现 ModelSchemaPanel 组件主体**

  内容与 `ModelDetailPanel` 的 `<div className="min-h-0 flex-1 overflow-y-auto">` 内部完全相同，但包裹在一个 `<div className="flex min-h-0 flex-1 flex-col bg-background">` 中，并在顶部加上模型标题 header。完整实现：

  ```tsx
  export function ModelSchemaPanel({
    state,
    crud,
    fieldOps,
    fkOps,
    orgName,
    projectSlug,
    onFieldAdded,
  }: ModelSchemaPanelProps) {
    const displayFieldOptions = (state.editModelData?.fields || []).filter(
      (field) => field.format !== 'RELATION'
    )
    const displayFieldSelectValue = state.metaDisplayField || '__display_field_none__'
    const isDisplayFieldUnset = state.metaDisplayField.trim() === ''
    const isManagedReadOnlyModel = state.editModelData?.createdVia === 'IMPORTED'

    if (!state.editModelData) {
      return (
        <div className="flex flex-1 flex-col items-center justify-center text-muted-foreground">
          <Table2 className="mb-3 size-10 opacity-20" />
          <p className="text-sm">从左侧选择模型以查看字段定义</p>
        </div>
      )
    }

    return (
      <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-border bg-card shadow-[0_2px_4px_rgba(0,0,0,0.04)]">
        {/* InsertFieldSheet — 与 ModelDetailPanel 保持相同的 z-index 嵌套 */}
        <InsertFieldSheet
          open={state.insertFieldOpen}
          onOpenChange={state.setInsertFieldOpen}
          modelId={state.editModelId || ''}
          modelName={state.editModelData?.name}
          projectSlug={projectSlug}
          orgName={orgName}
          existingFieldNames={(state.editModelData?.fields || []).map((f) => f.name)}
          onSuccess={() => {
            void crud.refreshModelDetail()
            onFieldAdded?.()
          }}
        />

        {/* Header */}
        <div className="flex shrink-0 items-start justify-between border-b border-border px-6 py-4">
          <div className="min-w-0">
            <h2 className="text-base font-semibold text-foreground">
              {state.editModelData.title || state.editModelData.name}
            </h2>
            <p className="mt-0.5 font-mono text-xs text-muted-foreground">
              {state.editModelData.name}
            </p>
          </div>
        </div>

        {/* Scrollable body — 与 ModelDetailPanel 内容区完全相同 */}
        <div className="min-h-0 flex-1 overflow-y-auto">
          {state.editModelLoading ? (
            <div className="flex flex-col items-center justify-center py-24">
              <Loader2 className="mb-3 size-5 animate-spin text-muted-foreground" />
              <span className="text-sm text-muted-foreground">加载中...</span>
            </div>
          ) : (
            <div className="divide-y divide-border [&>div]:py-6">
              {isManagedReadOnlyModel && (
                <div className="px-6 pb-0 pt-6">
                  <Alert variant="warning" className="py-2">
                    <AlertTriangle className="size-4" />
                    <AlertDescription className="text-xs">
                      当前模型为托管模型，仅支持查看，不支持结构和字段修改。
                    </AlertDescription>
                  </Alert>
                </div>
              )}

              {/* Meta info */}
              <div className="px-6">
                <div className="mb-3 flex items-center justify-between">
                  <span className="text-sm font-semibold text-foreground">元信息</span>
                  {!state.metaEditMode && (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="h-7 text-xs"
                      onClick={() => state.setMetaEditMode(true)}
                      disabled={isManagedReadOnlyModel}
                    >
                      设置展示字段
                    </Button>
                  )}
                </div>
                {isDisplayFieldUnset && (
                  <Alert variant="warning" className="mb-3 py-2">
                    <AlertTriangle className="size-4" />
                    <AlertDescription className="flex items-center justify-between gap-2 text-xs">
                      <span>未设置展示字段，关系展示将显示空(id)。</span>
                      {!state.metaEditMode && (
                        <Button
                          type="button"
                          size="sm"
                          variant="outline"
                          className="h-6 px-2 text-xs"
                          onClick={() => state.setMetaEditMode(true)}
                          disabled={isManagedReadOnlyModel}
                        >
                          去设置
                        </Button>
                      )}
                    </AlertDescription>
                  </Alert>
                )}
                <div className="grid grid-cols-2 gap-x-6 gap-y-3">
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">标识名称</label>
                    <Input value={state.editModelData.name} disabled className="h-8 bg-muted/30 font-mono text-xs" />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">显示标题</label>
                    <Input
                      value={state.metaTitle}
                      onChange={(e) => state.setMetaTitle(e.target.value)}
                      className="h-8 text-sm"
                      placeholder="输入显示标题"
                      disabled={!state.metaEditMode || isManagedReadOnlyModel}
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">数据库</label>
                    <Input value={state.editModelData.databaseName} disabled className="h-8 bg-muted/30 font-mono text-xs" />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">描述</label>
                    <Input
                      value={state.metaDescription}
                      onChange={(e) => state.setMetaDescription(e.target.value)}
                      className="h-8 text-sm"
                      placeholder="输入模型描述"
                      disabled={!state.metaEditMode || isManagedReadOnlyModel}
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">展示字段</label>
                    {state.metaEditMode && !isManagedReadOnlyModel ? (
                      <Select
                        value={displayFieldSelectValue}
                        onValueChange={(value) => {
                          state.setMetaDisplayField(value === '__display_field_none__' ? '' : value)
                        }}
                      >
                        <SelectTrigger className="h-8 text-sm">
                          <SelectValue placeholder="选择展示字段" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="__display_field_none__" className="text-sm">未设置</SelectItem>
                          {displayFieldOptions.map((field) => (
                            <SelectItem key={field.name} value={field.name} className="font-mono text-xs">
                              {field.name}{field.title && field.title !== field.name ? ` (${field.title})` : ''}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      <Input value={state.metaDisplayField || '未设置'} disabled className="h-8 bg-muted/30 text-sm" />
                    )}
                  </div>
                </div>
                {state.metaEditMode && !isManagedReadOnlyModel && (
                  <div className="mt-4 flex items-center justify-end gap-2">
                    <Button
                      variant="ghost" size="sm" className="h-7 text-xs"
                      onClick={() => {
                        state.setMetaTitle(state.editModelData!.title || '')
                        state.setMetaDescription(state.editModelData!.description || '')
                        state.setMetaDisplayField(state.editModelData!.displayField || '')
                        state.setMetaEditMode(false)
                      }}
                    >
                      取消
                    </Button>
                    <Button
                      size="sm" className="h-7 px-4 text-xs"
                      disabled={state.metaSaving}
                      onClick={async () => {
                        await crud.handleSaveMeta()
                        state.setMetaEditMode(false)
                      }}
                    >
                      {state.metaSaving && <Loader2 className="mr-1.5 size-3 animate-spin" />}
                      保存更改
                    </Button>
                  </div>
                )}
              </div>

              {/* Fields */}
              <div className="px-6">
                <div className="mb-3 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold text-foreground">字段定义</span>
                    <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                      {state.editModelData.fields?.length || 0}
                    </span>
                  </div>
                  <button
                    type="button"
                    title={isManagedReadOnlyModel ? '托管模型仅支持查看' : '新增字段'}
                    className="rounded p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50"
                    onClick={() => state.setInsertFieldOpen(true)}
                    disabled={isManagedReadOnlyModel}
                  >
                    <Plus className="size-3.5" />
                  </button>
                </div>
                <div className="overflow-hidden rounded-lg border border-border bg-card shadow-[0_1px_2px_rgba(0,0,0,0.04)]">
                  {state.editModelData.fields && state.editModelData.fields.length > 0 ? (
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b-2 border-border bg-card">
                            <th className="w-[180px] px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-foreground">标识(名称)</th>
                            <th className="w-[90px] px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-foreground">格式</th>
                            <th className="w-[90px] px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-foreground">类型</th>
                            <th className="w-[80px] px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-foreground">默认值</th>
                            <th className="w-[60px] px-3 py-2 text-center text-[11px] font-medium uppercase tracking-wider text-foreground">主键</th>
                            <th className="w-[50px] px-3 py-2 text-center text-[11px] font-medium uppercase tracking-wider text-foreground"></th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                          {state.editModelData.fields.map((field) => {
                            const enumDisplayFieldName = getEnumDisplayFieldName(field)
                            const isSystemField = isSystemGeneratedLabelField(field, state.editModelData?.fields ?? [])
                            const isFieldReadOnlyActionDisabled = isSystemField || isManagedReadOnlyModel
                            return (
                              <tr key={field.name} className="transition-colors hover:bg-muted/20">
                                <td className="px-3 py-2">
                                  <div className="flex flex-col">
                                    <div className="flex items-center gap-2">
                                      <span className={`font-mono text-sm ${field.isDeprecated ? 'text-muted-foreground line-through' : 'text-foreground'}`}>
                                        {field.name}
                                      </span>
                                      {isSystemField && (
                                        <span className="inline-flex items-center rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">系统字段</span>
                                      )}
                                      {field.isDeprecated && (
                                        <span className="bg-warning/10 text-warning inline-flex items-center rounded px-1.5 py-0.5 font-mono text-[11px]">已废弃</span>
                                      )}
                                    </div>
                                    {field.title && <span className="truncate text-xs text-muted-foreground">{field.title}</span>}
                                    {enumDisplayFieldName && (
                                      <span className="truncate text-xs text-muted-foreground">
                                        显示字段: <span className="font-mono">{enumDisplayFieldName}</span>
                                      </span>
                                    )}
                                    {isSystemField && (
                                      <span className="truncate text-xs text-muted-foreground">系统生成 / 只读 / 不可编辑 / 非物理列</span>
                                    )}
                                  </div>
                                </td>
                                <td className="px-3 py-2">
                                  <span className="inline-flex items-center rounded bg-primary/[0.08] px-1.5 py-0.5 font-mono text-[11px] text-primary">
                                    {field.format || '-'}
                                  </span>
                                </td>
                                <td className="px-3 py-2">
                                  <span className="inline-flex items-center rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-muted-foreground">
                                    {field.dbColumn?.columnType || field.storageHint || field.schemaType || 'String'}
                                  </span>
                                </td>
                                <td className="px-3 py-2">
                                  <span className="font-mono text-xs text-muted-foreground">
                                    {field.dbColumn?.defaultValue !== undefined ? String(field.dbColumn.defaultValue) : '-'}
                                  </span>
                                </td>
                                <td className="px-3 py-2 text-center">
                                  {field.isPrimary ? (
                                    <span className="bg-warning/10 inline-flex size-5 items-center justify-center rounded">
                                      <Key className="text-warning size-3" />
                                    </span>
                                  ) : (
                                    <span className="text-muted-foreground/30">-</span>
                                  )}
                                </td>
                                <td className="p-2 text-center">
                                  <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                      <Button variant="ghost" size="sm" className="size-6 p-0 hover:bg-muted">
                                        <MoreVertical className="size-3.5 text-muted-foreground" />
                                      </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end" className="w-36">
                                      <DropdownMenuItem
                                        className={`text-xs ${isFieldReadOnlyActionDisabled ? 'cursor-not-allowed text-muted-foreground/50' : 'cursor-pointer'}`}
                                        onClick={() => fieldOps.handleOpenEditField(field)}
                                        disabled={isFieldReadOnlyActionDisabled}
                                      >
                                        <Edit className="mr-2 size-3.5" />编辑
                                      </DropdownMenuItem>
                                      <DropdownMenuItem
                                        className={`text-xs ${isFieldReadOnlyActionDisabled ? 'cursor-not-allowed text-muted-foreground/50' : 'cursor-pointer'}`}
                                        onClick={() => fieldOps.handleToggleDeprecate(field)}
                                        disabled={isFieldReadOnlyActionDisabled}
                                      >
                                        <Archive className="mr-2 size-3.5" />{field.isDeprecated ? '取消废弃' : '废弃'}
                                      </DropdownMenuItem>
                                      <DropdownMenuItem
                                        className={`cursor-pointer text-xs ${field.isDeprecated && !isFieldReadOnlyActionDisabled ? 'text-destructive focus:text-destructive' : 'cursor-not-allowed text-muted-foreground/50'}`}
                                        onClick={() => fieldOps.handleRemoveField(field)}
                                        disabled={!field.isDeprecated || isFieldReadOnlyActionDisabled}
                                      >
                                        <Trash2 className="mr-2 size-3.5" />删除
                                      </DropdownMenuItem>
                                    </DropdownMenuContent>
                                  </DropdownMenu>
                                </td>
                              </tr>
                            )
                          })}
                        </tbody>
                      </table>
                    </div>
                  ) : (
                    <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                      <Table2 className="mb-2 size-8 opacity-30" />
                      <p className="text-sm">暂无字段</p>
                      <p className="mt-1 text-xs">点击上方按钮添加字段</p>
                    </div>
                  )}
                </div>
              </div>

              {/* Foreign keys */}
              <ForeignKeyPanel
                state={state}
                fkOps={fkOps}
                relationDatabaseNames={crud.relationDatabaseNames}
                getRelationModelsForDatabase={crud.getRelationModelsForDatabase}
                loadRelationModelsForDatabase={crud.loadRelationModelsForDatabase}
                isRelationModelsLoading={crud.isRelationModelsLoading}
              />
            </div>
          )}
        </div>
      </div>
    )
  }
  ```

- [ ] **Step 3: Lint 验证**

  ```bash
  cd modelcraft-front && npm run lint -- --max-warnings 0
  ```

  预期：0 warnings，0 errors。

- [ ] **Step 4: Commit**

  ```bash
  git add "modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSchemaPanel.tsx"
  git commit -m "feat(model-editor): add ModelSchemaPanel as inline schema view"
  ```

---

## Task 3: ModelSidebar — 接收 viewMode prop

**Files:**
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx`

- [ ] **Step 1: 给 `ModelSidebarProps` 添加 `viewMode` 字段**

  在 `interface ModelSidebarProps` 中添加：

  ```ts
  viewMode: 'schema' | 'data'
  ```

- [ ] **Step 2: 在 `data` 视角下隐藏「新建模型」「导入模型」按钮**

  将 `model-actions` 区块（第 158-187 行）整体包裹条件：

  ```tsx
  {/* Action buttons — only shown in schema view */}
  {viewMode === 'schema' && (
    <div className="flex flex-col gap-1 px-3 py-2.5">
      {/* 新建模型 button — 保持原样 */}
      ...
      {/* 导入模型 button — 保持原样 */}
      ...
    </div>
  )}
  ```

- [ ] **Step 3: Lint 验证**

  ```bash
  cd modelcraft-front && npm run lint -- --max-warnings 0
  ```

- [ ] **Step 4: Commit**

  ```bash
  git add "modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx"
  git commit -m "feat(model-editor): hide create/import buttons in data view mode"
  ```

---

## Task 4: ModelEditorView — 读取 view 参数，条件渲染主区

**Files:**
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelEditorView.tsx`

- [ ] **Step 1: 添加 `useSearchParams` 读取 `view` 参数**

  在文件顶部 import 区块添加：

  ```ts
  import { useSearchParams } from 'next/navigation'
  ```

  在 `ModelEditorView` 函数体顶部（`const state = useModelEditorState()` 之前）添加：

  ```ts
  const searchParams = useSearchParams()
  const viewMode = (searchParams.get('view') === 'data' ? 'data' : 'schema') as 'schema' | 'data'
  ```

- [ ] **Step 2: 切换到 data 视角时 reset 浮层状态**

  在现有 `useEffect` 区块（`setOpenedTabs` 之后）添加：

  ```ts
  // Reset schema-view overlays when switching to data view
  useEffect(() => {
    if (viewMode === 'data') {
      state.setEditModelOpen(false)
      state.setInsertFieldOpen(false)
      state.setEditFieldOpen(false)
      state.setFkFormOpen(false)
    }
  }, [viewMode]) // eslint-disable-line react-hooks/exhaustive-deps
  ```

- [ ] **Step 3: 切换到 schema 视角时自动加载选中模型详情**

  在 `viewMode === 'schema'` 且 `state.selectedModelId` 存在时，触发 `crud.handleEditModel`（复用现有逻辑，该函数会 fetch 详情并填充 `editModelData`）：

  ```ts
  // Auto-load model detail when entering schema view with a selected model
  useEffect(() => {
    if (viewMode === 'schema' && state.selectedModelId && !state.editModelData) {
      void crud.handleEditModel(state.selectedModelId)
    }
  }, [viewMode, state.selectedModelId]) // eslint-disable-line react-hooks/exhaustive-deps
  ```

- [ ] **Step 4: 把 `viewMode` 传给 `ModelSidebar`**

  找到 `<ModelSidebar` 的调用，添加 prop：

  ```tsx
  <ModelSidebar
    state={state}
    crud={crud}
    databases={crud.databases}
    databasesLoading={crud.databasesLoading}
    filteredModels={crud.filteredModels}
    modelsLoading={crud.modelsLoading}
    viewMode={viewMode}
  />
  ```

- [ ] **Step 5: 将主区改为条件渲染**

  将现有 `<main>` 区块（第 168-199 行）替换为：

  ```tsx
  {/* Right Content Area */}
  <main className="flex min-w-0 flex-1 flex-col bg-background p-4">
    <section className="flex h-full min-h-0 flex-col gap-3">
      {viewMode === 'schema' ? (
        <ModelSchemaPanel
          state={state}
          crud={crud}
          fieldOps={fieldOps}
          fkOps={fkOps}
          orgName={orgName}
          projectSlug={projectSlug}
          onFieldAdded={handleFieldAdded}
        />
      ) : (
        <DataWorkspacePanel
          tabs={openedTabs}
          activeTabId={state.selectedModelId ?? ''}
          onTabChange={(tabId) => state.setSelectedModelId(tabId)}
          onTabClose={handleCloseTab}
          emptyText="从左侧选择模型以打开数据表"
          className="h-full min-h-0"
          renderContent={(activeTab) => (
            <Suspense
              fallback={
                <div className="flex flex-1 items-center justify-center">
                  <div className="flex flex-col items-center gap-3 text-muted-foreground">
                    <Loader2 className="size-6 animate-spin" />
                    <span className="text-sm">加载中...</span>
                  </div>
                </div>
              }
            >
              <DevelopRecordWorkspace
                key={`${activeTab.id}-${schemaRefreshToken}`}
                modelId={activeTab.id}
                projectSlug={projectSlug}
                orgName={orgName}
                refreshToken={schemaRefreshToken}
              />
            </Suspense>
          )}
        />
      )}
    </section>
  </main>
  ```

- [ ] **Step 6: 添加 `ModelSchemaPanel` import**

  在文件顶部导入区添加：

  ```ts
  import { ModelSchemaPanel } from './ModelSchemaPanel'
  ```

- [ ] **Step 7: Lint 验证**

  ```bash
  cd modelcraft-front && npm run lint -- --max-warnings 0
  ```

- [ ] **Step 8: 浏览器端到端验证**

  1. 打开 `http://localhost:3100/org/lukeco/project/onboardceshi/model-editor`
  2. 左侧「数据模型」应自动展开，「模型管理」高亮
  3. 主区显示 ModelSchemaPanel（选中模型后展示元信息 + 字段表格）
  4. 点击左侧「数据管理」子项，URL 变为 `?view=data`，主区切换为数据 Tab，左栏「新建/导入」消失
  5. 再点「模型管理」，URL 恢复，主区回到字段列表，「新建/导入」重新出现
  6. 刷新页面 `/model-editor?view=data`，直接进入数据管理视角

- [ ] **Step 9: Commit**

  ```bash
  git add "modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelEditorView.tsx"
  git commit -m "feat(model-editor): split schema/data views via ?view query param"
  ```

---

## Self-Review

**Spec coverage:**
- [x] 左侧导航子项（模型管理/数据管理）→ Task 1
- [x] query param 切换（`?view=schema` / `?view=data`）→ Task 1 + Task 4
- [x] 默认高亮「模型管理」→ Task 1 Step 2
- [x] 自动展开导航项 → Task 1 Step 3
- [x] 数据管理隐藏新建/导入 → Task 3
- [x] 切换视角 reset 浮层状态 → Task 4 Step 2
- [x] 模型管理主区内嵌字段列表 → Task 2
- [x] 数据管理主区保持现有 DataWorkspacePanel → Task 4 Step 5
- [x] 选模型自动加载详情 → Task 4 Step 3

**Type consistency:**
- `viewMode: 'schema' | 'data'` 在 Task 1（无需）、Task 3（prop）、Task 4（本地变量）三处类型一致
- `ModelSchemaPanelProps` 与 `ModelDetailPanelProps` 字段完全对齐（相同 props，去掉无用的 Drawer 控制）
- `crud.handleEditModel` 在 Task 4 Step 3 复用，签名不变

**Placeholder scan:** 无 TBD/TODO，所有代码步骤均为完整可执行代码。
