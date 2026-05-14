# 筛选面板（Filter Panel）设计文档

**日期**: 2026-05-14
**状态**: 待实现
**范围**: `modelcraft-front` — EndUser 数据视图筛选功能

---

## 背景

`EndUserRecordWorkspace` 组件目前有一个"筛选"按钮占位，但无任何实现。搜索功能是客户端内存过滤（只过滤已加载的 50 条），`useQuery` 的 `where` 变量从未被使用。

后端 GraphQL 运行时已完整支持 `T{ModelName}WhereInput`，包括 `equals`、`contains`、`gt`/`lt`、`in`、`AND`/`OR`/`NOT` 等任意嵌套逻辑，前端只需接入。

---

## 设计目标

1. **MVP**：JSON 编辑器直接编写 `where` 条件，后端执行真实过滤
2. **AI-ready**：数据结构和组件接口提前为 AI 接入设计好，后续加一个按钮即可

---

## 核心决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 面板形式 | 内联展开（工具栏下方） | 筛选状态始终可见，不遮挡表格 |
| AND/OR | 全局切换（写在 JSON 顶层） | 交互最简，JSON 天然支持 |
| 编辑方式 | JSON 编辑器 + 字段速查面板 | AI 可直接读写 JSON；速查面板解决字段名记忆问题 |
| AI 接入 | 本期不做，接口预留 | MVP 先上线，AI 按钮后续叠加 |

---

## 交互设计

### 筛选按钮三态

| 状态 | 样式 | 触发条件 |
|------|------|----------|
| 默认态 | ghost button，灰色文字 | `whereJson` 为 null 或空 |
| 激活态 | 蓝色边框 + 蓝色文字 + 角标（条件数） | `whereJson` 非空且合法，面板关闭 |
| 展开态 | 同激活态 + 外发光 ring | 面板打开中 |

> 条件数 = 顶层 AND/OR 数组的长度；如果 where JSON 不是标准 AND/OR 结构，则角标显示 `•`。

### 面板布局

```
┌─────────────────────────────────────────────────────────┐
│ [Where JSON]              [格式化] [清空]                │
│ ┌──────────────────────────────┐ ┌─────────────────────┐│
│ │  代码编辑区（语法高亮）       │ │  字段 & 操作符      ││
│ │                              │ │                     ││
│ │  {                           │ │  name      String   ││
│ │    "AND": [                  │ │  status    Enum     ││
│ │      { "name": {             │ │  age       Int      ││
│ │          "contains": "张" }} │ │  createdAt Date     ││
│ │    ]                         │ │  ─────────────────  ││
│ │  }                           │ │  equals / not       ││
│ │                              │ │  contains           ││
│ │                              │ │  gt / gte / lt / lte││
│ │                              │ │  in: [...]          ││
│ │                              │ │  AND / OR / NOT     ││
│ └──────────────────────────────┘ └─────────────────────┘│
│  ✓ 有效 JSON                              [应用筛选]     │
└─────────────────────────────────────────────────────────┘
```

### 行为规范

- **展开/收起**：点击工具栏"筛选"按钮切换面板
- **实时校验**：JSON 输入时校验格式，底部显示 `✓ 有效 JSON` 或 `✗ JSON 格式错误`
- **应用筛选**：点击按钮后将 `whereJson` 传入 `useQuery` variables，触发后端查询
- **清空**：清空编辑器内容，同时清除当前筛选（重新 fetch 全量数据）
- **格式化**：调用 `JSON.stringify(JSON.parse(...), null, 2)` 美化 JSON
- **字段点击**：点击速查面板中的字段名，将 `"fieldName": {}` 片段插入编辑器光标处
- **面板收起时保留状态**：`whereJson` 不因面板收起而清除，按钮保持激活态

---

## 数据结构

```typescript
// EndUserRecordWorkspace 新增状态
const [filterOpen, setFilterOpen] = useState(false)

// 草稿态：编辑器内容，随用户输入实时变化，不触发查询
const [whereJsonDraft, setWhereJsonDraft] = useState<string>('')

// 已提交态：点击"应用筛选"后才更新，驱动实际查询
const [whereJsonCommitted, setWhereJsonCommitted] = useState<string | null>(null)

// 点击"应用筛选"时：将草稿提交
function handleApplyFilter() {
  setWhereJsonCommitted(whereJsonDraft || null)
}

// 派生：解析已提交的 where（传给 useQuery）
const whereInput = useMemo(() => {
  if (!whereJsonCommitted) return undefined
  try {
    return JSON.parse(whereJsonCommitted)
  } catch {
    return undefined
  }
}, [whereJsonCommitted])

// useQuery 接入（现有代码只需加 where）
variables: {
  take: 50,
  skip: 0,
  where: whereInput,  // ← 新增
}
```

> **AI 预留说明**：`whereJsonCommitted: string | null` 是 AI 的读写接口。AI agent 直接写入合法的 where JSON 字符串并调用 `handleApplyFilter()`，组件无需任何改动即可支持 AI 驱动的筛选。速查面板的字段列表和操作符参考，可直接作为 AI prompt 的 schema context。

---

## 组件结构

```
EndUserRecordWorkspace (已有，小改)
├── 搜索栏（已有，保持不变）
├── 工具栏（已有，接入 filterOpen / whereJson 状态）
├── FilterPanel (新组件)             ← src/web/components/features/end-user-data/FilterPanel.tsx
│   ├── WhereJsonEditor              ← 编辑器 + JSON 校验 + 格式化/清空
│   └── FieldSchemaPanel             ← 字段列表（来自 runtimeFields）+ 操作符速查
└── ModelRecordTable（已有，不变）
```

### `FilterPanel` Props

```typescript
interface FilterPanelProps {
  // 当前模型的字段定义（来自 runtimeFields，已有）
  fields: FieldDefinition[]
  // 编辑器草稿内容（随输入变化，不触发查询）
  whereJsonDraft: string
  // 用户编辑时回调
  onWhereJsonDraftChange: (json: string) => void
  // 点击"应用筛选"时回调（将草稿提交为已提交态）
  onApply: () => void
}
```

### `WhereJsonEditor` Props

```typescript
interface WhereJsonEditorProps {
  value: string
  onChange: (value: string) => void
  onFormat: () => void
  onClear: () => void
  isValid: boolean
}
```

### `FieldSchemaPanel` Props

```typescript
interface FieldSchemaPanelProps {
  fields: FieldDefinition[]
  // 点击字段名时，返回要插入编辑器的片段
  onFieldClick: (snippet: string) => void
}
```

---

## 筛选按钮激活态角标计算

```typescript
function getFilterCount(whereJson: string | null): number | '•' | null {
  if (!whereJson) return null
  try {
    const parsed = JSON.parse(whereJson)
    if (parsed.AND) return parsed.AND.length
    if (parsed.OR) return parsed.OR.length
    // 单字段条件（非 AND/OR 结构）
    const keys = Object.keys(parsed).filter(k => k !== 'NOT')
    return keys.length > 0 ? '•' : null
  } catch {
    return null
  }
}
```

---

## 编辑器选型

优先使用 `textarea` + 简单语法着色（CodeMirror `@codemirror/lang-json`），**不引入 Monaco**（包体积 ~2MB，过重）。

若项目已有 CodeMirror，直接复用；若无，用带行号的 `<textarea>` + JSON 校验也可接受作为 MVP。

---

## 超出范围（本期不做）

- 自然语言 → JSON（AI 生成）
- 排序（Sort）功能
- 分页（Pagination）
- 筛选条件持久化（localStorage / URL 参数）
- 图形化条件构建器（Airtable 风格）

---

## 验收标准

1. 点击"筛选"按钮展开面板，再次点击收起，状态保留
2. 在编辑器输入合法 JSON，点击"应用筛选"，表格数据更新为后端过滤结果
3. 输入非法 JSON 时，底部显示错误提示，"应用筛选"按钮禁用
4. 清空后表格恢复全量数据（不带 where 的请求）
5. 有激活筛选时，"筛选"按钮显示蓝色角标
6. 速查面板正确展示当前模型的字段名和类型
7. 点击字段名，将片段插入编辑器
