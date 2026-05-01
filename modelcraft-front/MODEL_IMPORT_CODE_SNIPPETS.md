# ModelCraft 模型导入 - 代码片段速查

## 核心代码片段

### 1️⃣ EditorModel 类型定义

**文件**：`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_hooks/types.ts`

```typescript
export interface EditorModel {
  id: string
  name: string                    // 英文标识（如 123startsWithNumb3ba18ee）
  title: string                   // 中文名（如 用户表）
  description?: string
  displayField?: string
  databaseName: string
  storageType?: string
  createdVia?: 'NEW' | 'IMPORTED' // ← 核心字段：导入状态
}
```

---

### 2️⃣ ModelSidebar 中的模型列表项渲染

**文件**：`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx` (L195-225)

```tsx
{!modelsLoading && filteredModels.map((model) => (
  <div
    key={model.id}
    role="button"
    tabIndex={0}
    onClick={() => handleModelDetailClick(model.id)}
    onKeyDown={(e) => e.key === 'Enter' && handleModelDetailClick(model.id)}
    className={cn(
      'group flex items-center gap-2 h-7 pl-2 pr-1 rounded-md cursor-pointer transition-colors select-none border-l-[3px]',
      state.selectedModelId === model.id
        ? 'bg-primary/[0.08] text-primary border-l-primary'
        : 'text-muted-foreground hover:bg-accent/60 hover:text-foreground border-l-transparent'
    )}
  >
    <Table2 className={cn('size-[15px] shrink-0 transition-colors', state.selectedModelId === model.id ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground')} />

    {/* 第一行：模型英文标识 */}
    <span className="min-w-0 flex-1 truncate text-xs">
      {model.name}  // 如：123startsWithNumb3ba18ee
    </span>

    {/* 导入状态标签 */}
    {model.createdVia === 'IMPORTED' && (
      <span className="rounded border border-warning/30 bg-warning/10 px-1 py-0 text-[10px] text-warning">
        托管
      </span>
    )}

    {/* 第二行：模型中文名（可选） */}
    {model.title && model.title !== model.name && (
      <span className="max-w-[56px] shrink-0 truncate text-xs text-muted-foreground/60" title={model.title}>
        {model.title}  // 如：用户表
      </span>
    )}

    {/* 更多操作菜单 */}
    <DropdownMenu>
      {/* ... */}
    </DropdownMenu>
  </div>
))}
```

---

### 3️⃣ 权限控制 - 禁用导入模型的编辑

**文件**：`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx` (L241-256)

```tsx
<DropdownMenuItem
  className={cn(
    'text-xs focus:bg-accent',
    model.createdVia === 'IMPORTED'
      ? 'cursor-not-allowed text-muted-foreground/50 focus:text-muted-foreground/50'
      : 'cursor-pointer text-foreground focus:text-foreground'
  )}
  onClick={(e) => {
    e.stopPropagation()
    if (model.createdVia === 'IMPORTED') {
      return  // ← 导入的模型禁止编辑
    }
    crud.handleEditModel(model.id)
  }}
  disabled={model.createdVia === 'IMPORTED'}  // ← UI 禁用
>
  <Edit className="mr-2 size-3.5" />
  编辑模型
</DropdownMenuItem>
```

---

### 4️⃣ GraphQL - GET_MODELS 查询

**文件**：`modelcraft-front/src/api-client/model/graphql-docs.ts` (L5-74)

```graphql
export const GET_MODELS = gql`
  query GetModels($input: ModelQueryInput) {
    models(input: $input) {
      edges {
        node {
          id
          projectSlug
          name
          title
          description
          databaseName
          storageType
          fields { ... }
          group { ... }
          dbTable
          createdAt
          updatedAt
        }
        cursor
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      totalCount
    }
  }
`
```

**使用方式**：
```typescript
const { data, loading, error, refetch } = useQuery<GetModelsData>(GET_MODELS, {
  variables: {
    input: {
      databaseName: state.selectedDatabase,
      limit: 100,
    },
  },
  skip: !projectSlug || !state.selectedDatabase,
  client: projectClient,
})
```

---

### 5️⃣ GraphQL - LIST_TABLES 查询（带 excludeExisting）

**文件**：`modelcraft-front/src/api-client/project/graphql-docs.ts` (L5-14)

```graphql
export const LIST_TABLES = gql`
  query ListTables($input: ListTablesInput!) {
    listTables(input: $input) {
      items {
        name
      }
      totalCount
    }
  }
`
```

**使用方式**（关键：排除已导入的表）：
```typescript
const { data, loading: tablesLoading } = useQuery<ListTablesQueryData>(LIST_TABLES, {
  client: projectClient,
  variables: {
    input: {
      databaseName,
      excludeExisting: true,      // ← 自动排除已导入的表！
      limit: PAGE_SIZE,
      offset,
    },
  },
  skip: !open || !projectSlug || !databaseName,
  fetchPolicy: 'network-only',
})
```

---

### 6️⃣ GraphQL - IMPORT_MODEL 变更

**文件**：`modelcraft-front/src/api-client/model/graphql-docs.ts` (L619-628)

```graphql
export const IMPORT_MODEL = gql`
  mutation ImportModel($input: ImportModelInput!) {
    importModel(input: $input) {
      modelId
      modelName
      fieldsCount
      skippedFields
    }
  }
`
```

**使用方式**：
```typescript
const [importModel, { loading: importing }] = useMutation<ImportModelMutationData>(
  IMPORT_MODEL,
  { client: projectClient }
)

const handleImport = async () => {
  if (!selectedTable) return

  try {
    const result = await importModel({
      variables: {
        input: {
          databaseName,
          tableName: selectedTable,
        },
      },
    })

    if (result.data?.importModel) {
      toast.success('模型导入成功')
      onSuccess()  // ← 刷新模型列表
      onOpenChange(false)
    } else {
      toast.error('导入失败，请重试')
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : '导入失败，请重试'
    toast.error(message)
  }
}
```

---

### 7️⃣ ImportModelDialog 完整组件

**文件**：`modelcraft-front/src/web/components/features/model-editor/ImportModelDialog.tsx` (完整)

```tsx
'use client'

import { useState, useEffect, useMemo } from 'react'
import { useParams } from 'next/navigation'
import { useQuery, useMutation, ApolloClient } from '@apollo/client'
import { ChevronLeft, ChevronRight, Loader2, Search, Table2, X } from 'lucide-react'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetFooter } from '@web/components/ui/sheet'
import { Input } from '@web/components/ui/input'
import { Button } from '@web/components/ui/button'
import { toast } from 'sonner'
import { LIST_TABLES } from '@/api-client/project'
import { IMPORT_MODEL } from '@/api-client/model'
import { useProjectScopedClient } from '@api-client/apollo/public'

const PAGE_SIZE = 20

interface ImportModelDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectSlug: string
  databaseName: string
  onSuccess: () => void
}

interface TableInfo {
  name: string
}

interface ListTablesQueryData {
  listTables?: {
    items: TableInfo[]
    totalCount: number
  } | null
}

interface ImportModelMutationData {
  importModel?: boolean | null
}

export function ImportModelDialog({
  open,
  onOpenChange,
  projectSlug,
  databaseName,
  onSuccess,
}: ImportModelDialogProps) {
  const params = useParams()
  const orgName = params?.orgName as string

  const [searchQuery, setSearchQuery] = useState('')
  const [selectedTable, setSelectedTable] = useState<string | null>(null)
  const [currentPage, setCurrentPage] = useState(1)

  const projectClient = useProjectScopedClient(projectSlug) as ApolloClient<object>
  const offset = (currentPage - 1) * PAGE_SIZE

  const { data, loading: tablesLoading } = useQuery<ListTablesQueryData>(LIST_TABLES, {
    client: projectClient,
    variables: {
      input: {
        databaseName,
        excludeExisting: true,  // ← 关键：排除已导入的表
        limit: PAGE_SIZE,
        offset,
      },
    },
    skip: !open || !projectSlug || !databaseName,
    fetchPolicy: 'network-only',
  })

  const [importModel, { loading: importing }] = useMutation<ImportModelMutationData>(
    IMPORT_MODEL,
    { client: projectClient }
  )

  useEffect(() => {
    if (!open) {
      setSearchQuery('')
      setSelectedTable(null)
      setCurrentPage(1)
    }
  }, [open])

  useEffect(() => {
    setCurrentPage(1)
  }, [searchQuery])

  const tables: TableInfo[] = useMemo(() => {
    if (!data?.listTables?.items) return []
    return data.listTables.items as TableInfo[]
  }, [data])

  const totalCount: number = data?.listTables?.totalCount ?? 0
  const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE))

  const filteredTables = useMemo(() => {
    if (!searchQuery) return tables
    const q = searchQuery.toLowerCase()
    return tables.filter((t) => t.name.toLowerCase().includes(q))
  }, [tables, searchQuery])

  const handleImport = async () => {
    if (!selectedTable) return

    try {
      const result = await importModel({
        variables: {
          input: {
            databaseName,
            tableName: selectedTable,
          },
        },
      })

      if (result.data?.importModel) {
        toast.success('模型导入成功')
        onSuccess()
        onOpenChange(false)
      } else {
        toast.error('导入失败，请重试')
      }
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : '导入失败，请重试'
      toast.error(message)
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="flex w-[400px] flex-col sm:max-w-[400px]">
        <SheetHeader>
          <SheetTitle className="text-base">导入模型</SheetTitle>
          <SheetDescription className="text-sm">
            从数据库 <span className="font-mono text-blue-600">{databaseName}</span> 选择一张表导入为模型
          </SheetDescription>
        </SheetHeader>

        <div className="flex flex-1 flex-col gap-3 overflow-hidden py-4">
          {/* Search input */}
          <div className="relative">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" strokeWidth={1.5} />
            <Input
              type="text"
              placeholder="搜索表名..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-8 px-8 text-sm"
            />
            {searchQuery && (
              <button
                type="button"
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                onClick={() => setSearchQuery('')}
              >
                <X className="size-3.5" strokeWidth={1.5} />
              </button>
            )}
          </div>

          {/* Table list */}
          <div className="flex-1 overflow-y-auto rounded-md border border-border bg-muted/10">
            {tablesLoading ? (
              <div className="flex h-full min-h-[200px] flex-col items-center justify-center gap-2 text-muted-foreground">
                <Loader2 className="size-5 animate-spin" strokeWidth={1.5} />
                <span className="text-sm">加载表列表...</span>
              </div>
            ) : filteredTables.length === 0 ? (
              <div className="flex h-full min-h-[200px] flex-col items-center justify-center gap-2 text-muted-foreground">
                <Table2 className="size-8 opacity-30" strokeWidth={1.5} />
                <p className="text-sm">
                  {totalCount === 0 ? '所有表已导入' : '未找到匹配的表'}
                </p>
              </div>
            ) : (
              <div className="divide-y divide-border">
                {filteredTables.map((table) => (
                  <button
                    key={table.name}
                    type="button"
                    className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-[#dadee5]"
                    style={
                      selectedTable === table.name
                        ? { backgroundColor: '#dadee5' }
                        : undefined
                    }
                    onClick={() => setSelectedTable(table.name)}
                  >
                    <Table2 className="size-3.5 shrink-0 text-muted-foreground" strokeWidth={1.5} />
                    <span className="font-mono text-sm text-foreground">{table.name}</span>
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Pagination */}
          {!searchQuery && totalPages > 1 && (
            <div className="flex items-center justify-between px-1">
              <span className="text-xs text-muted-foreground">
                第 {currentPage} / {totalPages} 页，共 {totalCount} 张表
              </span>
              <div className="flex items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  className="size-7 p-0"
                  onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                  disabled={currentPage <= 1 || tablesLoading}
                  aria-label="上一页"
                >
                  <ChevronLeft className="size-3.5" strokeWidth={1.5} />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="size-7 p-0"
                  onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                  disabled={currentPage >= totalPages || tablesLoading}
                  aria-label="下一页"
                >
                  <ChevronRight className="size-3.5" strokeWidth={1.5} />
                </Button>
              </div>
            </div>
          )}
        </div>

        <SheetFooter>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onOpenChange(false)}
            disabled={importing}
          >
            取消
          </Button>
          <Button
            size="sm"
            className="border-0 bg-[#2563eb] text-white transition-colors duration-200 hover:bg-[#1d4ed8]"
            onClick={handleImport}
            disabled={!selectedTable || importing}
          >
            {importing && <Loader2 className="mr-1.5 size-3.5 animate-spin" strokeWidth={1.5} />}
            {importing ? '导入中...' : '导入'}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
```

---

### 8️⃣ 在 ModelEditorView 中集成 ImportModelDialog

**文件**：`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelEditorView.tsx` (L135-142)

```tsx
{/* Import Model Dialog */}
<ImportModelDialog
  open={state.importDialogOpen}
  onOpenChange={state.setImportDialogOpen}
  projectSlug={projectSlug}
  databaseName={state.selectedDatabase}
  onSuccess={() => crud.refetchModels()}  // ← 导入后刷新列表
/>
```

---

### 9️⃣ 在 ModelSidebar 中显示导入按钮

**文件**：`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx` (L147-159)

```tsx
<Button
  size="sm"
  variant="outline"
  className={cn(
    'h-7 w-full justify-start px-2.5 text-xs font-normal transition-colors',
    !state.selectedDatabase && 'pointer-events-none opacity-40'
  )}
  onClick={() => state.setImportDialogOpen(true)}
  disabled={!state.selectedDatabase}
>
  <Download className="mr-1 size-3.5" strokeWidth={1.5} />
  导入模型
</Button>
```

---

## 🔍 查询变量示例

### GET_MODELS 变量

```json
{
  "input": {
    "databaseName": "public",
    "limit": 100
  }
}
```

### LIST_TABLES 变量（导入对话框）

```json
{
  "input": {
    "databaseName": "public",
    "excludeExisting": true,
    "limit": 20,
    "offset": 0
  }
}
```

### IMPORT_MODEL 变量

```json
{
  "input": {
    "databaseName": "public",
    "tableName": "users"
  }
}
```

---

## 📊 状态流转图

```
用户点击"导入模型"按钮
  ↓
ImportModelDialog 打开
  ↓
LIST_TABLES 查询加载表列表（excludeExisting=true）
  ↓
用户搜索、选择表
  ↓
用户点击"导入"按钮
  ↓
IMPORT_MODEL mutation 执行
  ↓
成功 → toast.success() + onSuccess() → crud.refetchModels()
  ↓
模型列表刷新 → GET_MODELS 重新查询
  ↓
新导入的模型出现在列表中（createdVia = 'IMPORTED'）
```

