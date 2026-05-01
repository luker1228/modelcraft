# ModelCraft 前端"模型导入"代码探索报告

## 📋 概述
探索了 modelcraft-front 中与"模型导入"（Import Model）相关的代码，梳理了 UI 组件、GraphQL 查询、数据类型和业务流程。

---

## 1️⃣ 模型列表选择组件

### 📍 主要文件位置
- **模型列表组件**：`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx`
- **导入对话框**：`modelcraft-front/src/web/components/features/model-editor/ImportModelDialog.tsx`
- **编辑器主视图**：`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelEditorView.tsx`

### 🎨 模型列表 UI 结构（两行显示）

在 **ModelSidebar.tsx** 中（第 195-225 行），每个模型项显示为：

```tsx
<div className="group flex items-center gap-2 h-7 pl-2 pr-1 rounded-md cursor-pointer ...">
  <Table2 className="size-[15px] shrink-0" />
  
  {/* 第一行：模型英文标识（name） */}
  <span className="min-w-0 flex-1 truncate text-xs">
    {model.name}  // 如：123startsWithNumb3ba18ee
  </span>

  {/* 导入状态标签（如果是导入的模型） */}
  {model.createdVia === 'IMPORTED' && (
    <span className="rounded border border-warning/30 bg-warning/10 px-1 py-0 text-[10px] text-warning">
      托管
    </span>
  )}

  {/* 第二行：模型中文名（title） */}
  {model.title && model.title !== model.name && (
    <span className="max-w-[56px] shrink-0 truncate text-xs text-muted-foreground/60" title={model.title}>
      {model.title}  // 如：用户表 / Orders
    </span>
  )}

  {/* 更多操作菜单 */}
  <DropdownMenu>
    {/* ... */}
  </DropdownMenu>
</div>
```

### 📐 数据类型

**EditorModel** 定义在 `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_hooks/types.ts`：

```typescript
export interface EditorModel {
  id: string
  name: string                    // 英文标识（数据库表名标准化）
  title: string                   // 中文显示名
  description?: string
  displayField?: string
  databaseName: string
  storageType?: string
  createdVia?: 'NEW' | 'IMPORTED' // ← 导入状态标记
}
```

---

## 2️⃣ 相关 GraphQL 查询

### 📝 GET_MODELS（获取模型列表）

位置：`modelcraft-front/src/api-client/model/graphql-docs.ts` 第 5-74 行

```graphql
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
    pageInfo { ... }
    totalCount
  }
}
```

**变量示例**：
```json
{
  "input": {
    "databaseName": "public",
    "limit": 100
  }
}
```

### 🔄 LIST_TABLES（获取可导入的表列表）

位置：`modelcraft-front/src/api-client/project/graphql-docs.ts` 第 5-14 行

```graphql
query ListTables($input: ListTablesInput!) {
  listTables(input: $input) {
    items {
      name
    }
    totalCount
  }
}
```

**变量示例**：
```json
{
  "input": {
    "databaseName": "public",
    "excludeExisting": true,    // ← 排除已导入的表
    "limit": 20,
    "offset": 0
  }
}
```

### 💾 IMPORT_MODEL（执行模型导入）

位置：`modelcraft-front/src/api-client/model/graphql-docs.ts` 第 619-628 行

```graphql
mutation ImportModel($input: ImportModelInput!) {
  importModel(input: $input) {
    modelId
    modelName
    fieldsCount
    skippedFields
  }
}
```

**变量示例**：
```json
{
  "input": {
    "databaseName": "public",
    "tableName": "users"
  }
}
```

---

## 3️⃣ 已导入状态字段和逻辑

### 🏷️ 导入状态标记

在 **EditorModel** 中：
- **字段名**：`createdVia`
- **类型**：`'NEW' | 'IMPORTED'`
- **含义**：
  - `'NEW'`：手动创建的模型（可编辑、可删除）
  - `'IMPORTED'`：从数据库表导入的模型（只读、"托管"）

### 🔒 权限控制逻辑

在 **ModelSidebar.tsx** 中（第 215-288 行）：

```typescript
// 导入的模型显示"托管"标签
{model.createdVia === 'IMPORTED' && (
  <span className="...">托管</span>
)}

// 导入的模型的编辑、删除按钮被禁用
<DropdownMenuItem
  disabled={model.createdVia === 'IMPORTED'}
  className={cn(
    'text-xs focus:bg-accent',
    model.createdVia === 'IMPORTED'
      ? 'cursor-not-allowed text-muted-foreground/50'
      : 'cursor-pointer text-foreground'
  )}
>
  <Edit className="mr-2 size-3.5" />
  编辑模型
</DropdownMenuItem>

<DropdownMenuItem
  disabled={model.createdVia === 'IMPORTED'}
  className={cn(
    'text-xs focus:bg-accent',
    model.createdVia === 'IMPORTED'
      ? 'cursor-not-allowed text-muted-foreground/50'
      : 'cursor-pointer text-destructive'
  )}
>
  删除模型
</DropdownMenuItem>
```

---

## 4️⃣ 数据流和来源

### 🔀 完整数据流

```
ModelEditorView (主容器)
  ├── useModelEditorState() (状态管理)
  ├── useModelCRUD() (数据获取和操作)
  │   ├── useQuery(GET_MODELS) → 获取模型列表
  │   │   └── 调用 /graphql/org/{orgName}/project/{projectSlug}/
  │   ├── useMutation(CREATE_MODEL)
  │   ├── useMutation(UPDATE_MODEL)
  │   └── useMutation(DELETE_MODEL)
  │
  ├── ModelSidebar (左侧模型列表)
  │   └── 显示 filteredModels（过滤后的模型列表）
  │
  └── ImportModelDialog (导入对话框)
      ├── useQuery(LIST_TABLES) → 获取可导入的表
      │   └── 变量：excludeExisting: true
      └── useMutation(IMPORT_MODEL) → 执行导入
```

### 📊 数据来源

1. **模型列表** → `GET_MODELS` GraphQL 查询
   - 从 Apollo Client 获取，使用 `projectScopedClient`
   - 路径：`/graphql/org/{orgName}/project/{projectSlug}/`

2. **可导入表列表** → `LIST_TABLES` GraphQL 查询
   - 项目级 GraphQL 端点
   - 自动排除已导入的表（`excludeExisting: true`）

3. **导入执行** → `IMPORT_MODEL` mutation
   - 调用后自动 refetch 模型列表

---

## 5️⃣ 导入流程实现

### 📝 ImportModelDialog 详细流程

位置：`modelcraft-front/src/web/components/features/model-editor/ImportModelDialog.tsx`

```typescript
interface ImportModelDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectSlug: string
  databaseName: string
  onSuccess: () => void
}

export function ImportModelDialog({
  open,
  onOpenChange,
  projectSlug,
  databaseName,
  onSuccess,
}: ImportModelDialogProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedTable, setSelectedTable] = useState<string | null>(null)
  const [currentPage, setCurrentPage] = useState(1)

  // 1. 获取可导入的表列表（分页）
  const { data, loading: tablesLoading } = useQuery<ListTablesQueryData>(
    LIST_TABLES,
    {
      variables: {
        input: {
          databaseName,
          excludeExisting: true,  // ← 关键：排除已导入
          limit: PAGE_SIZE,
          offset: (currentPage - 1) * PAGE_SIZE,
        },
      },
      skip: !open || !projectSlug || !databaseName,
      fetchPolicy: 'network-only',
    }
  )

  // 2. 执行导入 mutation
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
        onSuccess()  // ← 刷新父级的模型列表
        onOpenChange(false)
      } else {
        toast.error('导入失败，请重试')
      }
    } catch (error) {
      // 错误处理
    }
  }

  // UI: 搜索 → 表列表 → 分页 → 导入按钮
}
```

### 🔗 集成到 ModelEditor

在 **ModelEditorView.tsx** 中（第 135-142 行）：

```tsx
<ImportModelDialog
  open={state.importDialogOpen}
  onOpenChange={state.setImportDialogOpen}
  projectSlug={projectSlug}
  databaseName={state.selectedDatabase}
  onSuccess={() => crud.refetchModels()}  // ← 导入后刷新列表
/>
```

在 **ModelSidebar.tsx** 中（第 147-159 行）：

```tsx
<Button
  size="sm"
  variant="outline"
  onClick={() => state.setImportDialogOpen(true)}
  disabled={!state.selectedDatabase}
>
  <Download className="mr-1 size-3.5" strokeWidth={1.5} />
  导入模型
</Button>
```

---

## 6️⃣ 使用的 Apollo Client

### 🌐 项目级 Client

```typescript
// 位置：src/api-client/apollo/public.ts
const projectClient = useProjectScopedClient(projectSlug)

// 特点：
// - 针对单个项目的 GraphQL endpoint
// - 路径：/graphql/org/{orgName}/project/{projectSlug}/
// - 用于查询该项目的模型、表等信息
```

---

## 7️⃣ 关键代码位置速查表

| 功能 | 文件路径 | 关键行/函数 |
|------|---------|-----------|
| 模型列表组件 | `_components/ModelSidebar.tsx` | 第 195-225 行，`model.name` / `model.title` |
| 导入对话框 | `@web/components/features/model-editor/ImportModelDialog.tsx` | 整个文件 |
| 数据类型定义 | `_hooks/types.ts` | `EditorModel` 接口 |
| GET_MODELS 查询 | `api-client/model/graphql-docs.ts` | 第 5-74 行 |
| LIST_TABLES 查询 | `api-client/project/graphql-docs.ts` | 第 5-14 行 |
| IMPORT_MODEL 变更 | `api-client/model/graphql-docs.ts` | 第 619-628 行 |
| 模型 CRUD | `_hooks/use-model-crud.ts` | 第 46-180 行 |
| 编辑器主容器 | `_components/ModelEditorView.tsx` | 第 31-200 行 |
| 导入状态字段 | `_hooks/types.ts` | `EditorModel.createdVia` |
| 权限控制逻辑 | `_components/ModelSidebar.tsx` | 第 241-288 行 |

---

## 8️⃣ 关键发现总结

### ✅ 已导入状态的完整实现

1. **状态标记**：通过 `createdVia` 字段标记（'NEW' 或 'IMPORTED'）
2. **UI 反馈**：显示"托管"标签
3. **权限管理**：禁用编辑和删除操作
4. **查询过滤**：`LIST_TABLES` 使用 `excludeExisting: true` 自动排除已导入

### 🎯 模型列表两行显示逻辑

```
┌─────────────────────────────────┐
│ 🔷 123startsWithNumb3ba18ee 托管 │ ← name(英文) + 状态标签
│            用户表              │ ← title(中文)
└─────────────────────────────────┘
```

- **第一行**：`model.name`（数据库表标准化名称）+ 状态标签
- **第二行**：`model.title`（中文显示名，仅当与 name 不同时显示）
- **条件**：`model.title && model.title !== model.name`

### 📈 数据路由

- 前端：`projectScopedClient` → Apollo
- 后端：`/graphql/org/{orgName}/project/{projectSlug}/`
- Resolver：处理 `models`、`listTables`、`importModel`

---

## 🎓 建议查看的相关文件

1. **后端实现**：`modelcraft-backend/pkg/resolver/` 中的 `models.resolver.ts`
2. **GraphQL Schema**：查看 `schema.graphqls` 中 `ImportModel` 和 `importModel` 定义
3. **数据库模型**：`modelcraft-backend` 中的数据库迁移文件

