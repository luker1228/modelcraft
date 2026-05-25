# Model Editor 双视角设计方案

**日期**: 2026-05-25  
**状态**: 设计已确认，待实现  
**路由**: `/org/[orgName]/project/[projectSlug]/model-editor`

---

## 背景与目标

当前 `/model-editor` 页面将"模型管理"（结构编辑：字段/关系/元数据）和"数据管理"（记录增删改查）混合在同一视角下。用户需要先打开模型 Tab 才能操作数据，边界不清晰。

**目标**：将两种用途拆分为左侧导航可切换的两个子视角，与现有"访问控制 → 角色/权限包"保持一致的交互模式。

---

## 设计决策

### 1. 动线重新定义

```
进入页面
  └── 选择数据库（左栏顶部 Popover，共用）
        └── 在左侧导航选择视角
              ├── 模型管理  →  字段/关系/元数据编辑
              └── 数据管理  →  记录查询/增删改
```

两个视角都依赖数据库选择，数据库选择器在两个视角下共用，不重复渲染。

### 2. 导航实现方式：query param（方案 B）

- 保持同一个路由 `/model-editor`，用 `?view=schema`（默认）和 `?view=data` 区分视角
- 与"访问控制"子项的 `tabParam` 机制完全一致，无需新路由
- 切换视角时页面不重载，`selectedDatabase`、模型列表、Apollo 缓存完整保留

URL 示例：
```
/org/lukeco/project/onboardceshi/model-editor           → 默认进入模型管理
/org/lukeco/project/onboardceshi/model-editor?view=data → 数据管理视角
```

### 3. 左侧导航变更

`AppLayout.tsx` 中 `projectNavSections` 的「数据模型」条目添加 `children`：

```ts
// 变更前
{ label: '数据模型', icon: '/icons/icon-table2.svg', href: `.../model-editor` }

// 变更后
{
  label: '数据模型',
  icon: '/icons/icon-table2.svg',
  href: `.../model-editor`,
  children: [
    { label: '模型管理', href: `.../model-editor`, tabParam: 'view=schema' },
    { label: '数据管理', href: `.../model-editor?view=data`, tabParam: 'view=data' },
  ]
}
```

自动展开逻辑：参考「访问控制」已有的 `useEffect` 模式，当路径匹配 `/model-editor` 时自动展开。

「模型管理」为默认 tab：当 `?view` 参数缺失或为 `schema` 时高亮「模型管理」（与 `roles` 默认 tab 逻辑对称）。

### 4. 能力边界（严格隔离）

| 能力 | 模型管理视角 | 数据管理视角 |
|------|------------|------------|
| 字段 CRUD | ✅ | ❌ |
| 外键管理 | ✅ | ❌ |
| 元数据编辑（title/desc/displayField） | ✅ | ❌ |
| 新建模型 | ✅（左栏显示） | ❌（左栏隐藏） |
| 导入模型 | ✅（左栏显示） | ❌（左栏隐藏） |
| 删除模型 | ✅（右键菜单） | ❌ |
| 记录查询 | ❌ | ✅ |
| 记录增删改 | ❌ | ✅ |

### 5. ModelEditorView 内部切换逻辑

`ModelEditorView` 读取 `searchParams.get('view')`（或 `useSearchParams()`），决定渲染哪个主区内容：

```
view === 'data' (或 ?view=data)
  → 主区渲染 DataWorkspacePanel + DevelopRecordWorkspace（现有逻辑）
  → ModelSidebar 隐藏「新建模型」「导入模型」按钮

view === 'schema'（或参数缺失，默认）
  → 主区渲染内嵌字段列表面板 ModelSchemaPanel（新建组件）
  → 选中模型后主区直接展示：元信息区（标识/标题/描述/展示字段）+ 字段定义表格 + 外键关系
  → ModelSidebar 显示「新建模型」「导入模型」按钮
  → 原 ModelDetailPanel Drawer 仍保留，供「编辑字段」等 Sheet 的 z-index 嵌套使用
```

### 6. 切换视角时的状态处理

切换到「数据管理」视角时，需 reset 以下模型编辑相关状态，避免 Drawer/Sheet 残留：
- `editModelOpen` → `false`
- `insertFieldOpen` → `false`
- `editFieldOpen` → `false`
- `fkFormOpen` → `false`

切换到「模型管理」视角时，`openedTabs`（数据管理的多模型 Tab 列表）保持不变，下次回到数据管理时可恢复。

### 7. 默认落点

- 进入 `/model-editor`（无参数）：默认进入「模型管理」视角
- `selectedDatabase` 通过 Zustand persist 跨路由保留，两个视角之间切换时无需重新选择

---

## 文件变更清单

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `src/web/components/features/layout/AppLayout.tsx` | 修改 | 给「数据模型」NavItem 添加 children（模型管理/数据管理）；添加 model-editor 路径自动展开逻辑 |
| `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelEditorView.tsx` | 修改 | 读取 `useSearchParams()` 的 `view` 参数，条件渲染主区（`schema` → ModelSchemaPanel，`data` → DataWorkspacePanel）；切换到 data 视角时 reset 浮层状态 |
| `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx` | 修改 | 接收 `viewMode: 'schema' \| 'data'` prop，`data` 视角下隐藏「新建模型」「导入模型」按钮 |
| `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSchemaPanel.tsx` | 新建 | 内嵌字段列表主区面板，提取自 ModelDetailPanel 的内容区（元信息 + 字段表格 + 外键面板），无 Drawer 包装 |

---

## 不在本次范围内

- 数据管理视角的新功能
- 路由级别的独立拆分（方案 A，本次不做）
- ModelDetailPanel Drawer 的删除（仍作为 InsertFieldSheet/FieldEditSheet 的宿主容器保留）
