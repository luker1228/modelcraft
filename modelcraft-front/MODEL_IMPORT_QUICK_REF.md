# 模型导入（Import Model）- 快速查询表

## 📌 核心文件索引表

| # | 组件/功能 | 文件路径 | 关键代码/行数 | 说明 |
|---|---------|--------|-----------|-----|
| 1 | 模型列表UI | `_components/ModelSidebar.tsx` | L195-225 | 渲染模型项，显示 name + title |
| 2 | 两行显示逻辑 | `ModelSidebar.tsx` | L211-225 | 第一行：name，第二行：title（条件显示） |
| 3 | 导入状态标签 | `ModelSidebar.tsx` | L215-219 | `createdVia === 'IMPORTED'` → 显示"托管" |
| 4 | 权限控制 | `ModelSidebar.tsx` | L241-288 | 禁用导入模型的编辑/删除 |
| 5 | 导入对话框 | `@web/components/features/model-editor/ImportModelDialog.tsx` | 整个文件 | 表选择、搜索、分页、导入 |
| 6 | EditorModel 类型 | `_hooks/types.ts` | L13-22 | 包含 `createdVia` 字段 |
| 7 | GET_MODELS 查询 | `api-client/model/graphql-docs.ts` | L5-74 | 获取项目的所有模型 |
| 8 | LIST_TABLES 查询 | `api-client/project/graphql-docs.ts` | L5-14 | 获取可导入表，支持 excludeExisting |
| 9 | IMPORT_MODEL 变更 | `api-client/model/graphql-docs.ts` | L619-628 | 执行导入操作 |
| 10 | 模型CRUD | `_hooks/use-model-crud.ts` | L46-180 | 使用 GET_MODELS + mutations |
| 11 | 编辑器主视图 | `_components/ModelEditorView.tsx` | L31-142 | 集成所有组件，状态管理 |
| 12 | 导入按钮 | `ModelSidebar.tsx` | L147-159 | 触发 `setImportDialogOpen(true)` |

---

## 🎯 模型列表两行显示逻辑

### UI 结构
```
┌─────────────────────────────────────┐
│ 🔷 123startsWithNumb3ba18ee 托管     │ ← name + 导入状态标签
│             用户表                  │ ← title（条件显示）
└─────────────────────────────────────┘
```

### 渲染条件
| 字段 | 来源 | 显示规则 |
|-----|------|---------|
| name | `model.name` | 总是显示 |
| 状态标签 | `model.createdVia === 'IMPORTED'` | createdVia === 'IMPORTED' 时显示"托管" |
| title | `model.title` | 仅当 `title` 存在 && `title !== name` 时显示 |

---

## 🔄 GraphQL 查询完整对照

| 查询/变更 | 来源 | 用途 | 关键参数 | 响应字段 |
|---------|------|------|--------|---------|
| **GET_MODELS** | `api-client/model/graphql-docs.ts:5-74` | 获取模型列表 | `databaseName`, `limit` | `models.edges[].node` |
| **LIST_TABLES** | `api-client/project/graphql-docs.ts:5-14` | 获取可导入表 | `databaseName`, `excludeExisting: true` | `listTables.items[]` |
| **IMPORT_MODEL** | `api-client/model/graphql-docs.ts:619-628` | 执行导入 | `databaseName`, `tableName` | `modelId`, `modelName`, `fieldsCount` |

---

## 🔐 导入状态的字段和权限

### 状态标记字段
- **字段名**: `EditorModel.createdVia`
- **类型**: `'NEW' \| 'IMPORTED'`
- **位置**: `_hooks/types.ts:21`

### 权限逻辑
| 操作 | NEW模型 | IMPORTED模型 | 代码位置 |
|-----|--------|-----------|--------|
| 编辑 | ✅ 允许 | ❌ 禁用 | `ModelSidebar.tsx:241-256` |
| 删除 | ✅ 允许 | ❌ 禁用 | `ModelSidebar.tsx:271-288` |
| 复制名称 | ✅ 允许 | ✅ 允许 | `ModelSidebar.tsx:257-269` |
| 显示"托管"标签 | ❌ 无 | ✅ 显示 | `ModelSidebar.tsx:215-219` |

---

## 📊 数据流完整图

```
用户进入模型编辑器
    ↓
ModelEditorView 初始化
    ├─ useModelEditorState() → 状态容器
    ├─ useModelCRUD()
    │   └─ useQuery(GET_MODELS)
    │       └─ 查询 /graphql/org/{org}/project/{slug}/
    │           └─ 返回 models[] 包含 createdVia 字段
    │
    ├─ ModelSidebar 渲染模型列表
    │   └─ 过滤 + 显示 filteredModels
    │       └─ 显示 name + title + 导入标签
    │
    └─ 用户点击"导入模型"
        └─ setImportDialogOpen(true)
            └─ ImportModelDialog 打开
                ├─ useQuery(LIST_TABLES, { excludeExisting: true })
                │   └─ 加载可导入表列表
                │       └─ 已导入的表自动排除
                │
                ├─ 用户搜索 + 选择表
                └─ 用户点击"导入"
                    └─ useMutation(IMPORT_MODEL)
                        └─ 导入成功
                            └─ onSuccess() → crud.refetchModels()
                                └─ 重新查询 GET_MODELS
                                    └─ 新模型出现在列表（createdVia='IMPORTED'）
```

---

## 🌐 Apollo Client 路由

| 端点 | 用途 | 代码 |
|-----|------|------|
| `projectScopedClient` | 项目级 GraphQL 端点 | `api-client/apollo/public.ts` |
| 路径 | `/graphql/org/{orgName}/project/{projectSlug}/` | - |
| 查询 | GET_MODELS, LIST_TABLES, IMPORT_MODEL | 各 graphql-docs.ts |

---

## ⚡ 快速查找

### "已导入"状态在哪里？
1. **标记字段**: `EditorModel.createdVia` → `_hooks/types.ts:21`
2. **UI显示**: `ModelSidebar.tsx:215-219` → 显示"托管"标签
3. **权限控制**: `ModelSidebar.tsx:241-288` → 禁用编辑/删除

### 模型列表怎么查询？
1. **查询定义**: `api-client/model/graphql-docs.ts:5-74` → `GET_MODELS`
2. **使用位置**: `use-model-crud.ts:110-119`
3. **变量示例**: `{ input: { databaseName, limit: 100 } }`

### 导入流程从哪里开始？
1. **按钮**: `ModelSidebar.tsx:147-159` → 点击打开对话框
2. **对话框**: `ImportModelDialog.tsx` 整个文件
3. **核心逻辑**: L115-136 → `handleImport()` 函数

### excludeExisting 参数在哪里？
1. **位置**: `ImportModelDialog.tsx:68-75`
2. **关键代码**: `excludeExisting: true` 在 LIST_TABLES 变量中
3. **效果**: 自动排除已导入的表，用户只能看到未导入的表

---

## 📝 前端数据类型对应

| TypeScript 类型 | 后端 GraphQL 类型 | 文件位置 |
|----------------|------------------|--------|
| `EditorModel` | `Model` | `_hooks/types.ts:13-22` |
| `ListTablesQueryData` | - | `ImportModelDialog.tsx:36-41` |
| `ImportModelMutationData` | - | `ImportModelDialog.tsx:43-45` |
| `ModelsQueryData` | - | `_hooks/types.ts:52-56` |

---

## 🎓 学习路径推荐

1. **理解 UI**: 看 `ModelSidebar.tsx` 第 195-225 行
2. **理解状态**: 看 `EditorModel` 接口和 `createdVia` 字段
3. **理解查询**: 看 `GET_MODELS` 和 `LIST_TABLES` GraphQL 定义
4. **理解导入**: 看 `ImportModelDialog.tsx` 的完整实现
5. **理解权限**: 看 `ModelSidebar.tsx` 第 241-288 行的禁用逻辑
6. **理解集成**: 看 `ModelEditorView.tsx` 如何组织这些组件

