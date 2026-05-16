# AI-Driven Frontend Workflow Design

**日期：** 2026-05-16
**状态：** 已确认，待实施

---

## 背景与目标

用户希望通过对话完成 ModelCraft 中的所有操作——只需点击最终确认，其余由 AI 引导完成。

**核心原则：**
- Agent（Python）只做**读**：查询数据、发现资源、生成建议
- **写操作由前端工具完成**：Agent 调用前端注册的语义工具，修改的是 draft state（预填表单），用户点 Save/Delete 才真正走 mutation
- Agent **永远不直接写后端数据**

---

## 整体架构

```
OrgLayout (CopilotWrapper)
  ├── useCopilotReadable: { layer: "org", orgName, availableActions }
  └── OrgCopilotActions          ← 新组件，注册 org 层工具

ProjectLayout (CopilotWrapper)
  ├── useCopilotReadable: { layer: "project", orgName, projectSlug, currentModel, currentDb, availableActions }
  └── ProjectCopilotActions      ← 新组件，注册 project 层工具
```

工具随组件挂载/卸载自动注册/销毁——进入 project 页时 org 工具自动卸载，project 工具自动注册。无需手动管理。

---

## Org 层工具集（`OrgCopilotActions`）

注册位置：`OrgLayout`，随 org 页面生命周期存在。

| 工具名 | 参数 | 前端效果 | 需用户确认 |
|--------|------|----------|-----------|
| `navigate_to_project` | `slug: string` | `router.push(.../project/[slug])` | 否 |
| `navigate_to_settings` | — | `router.push(.../settings)` | 否 |
| `open_create_project` | `prefill?: {slug, title, description}` | 打开新建项目 Sheet，字段预填 | 是（用户点 Create） |
| `highlight_project` | `slug: string, reason: string` | 项目卡片高亮 + tooltip | 否 |

**`useCopilotReadable` 内容：**
```ts
{
  layer: "org",
  orgName: string,
  availableActions: ["navigate_to_project", "open_create_project", "highlight_project", "navigate_to_settings"]
}
```

---

## Project 层工具集（`ProjectCopilotActions`）

注册位置：`ProjectLayout`，随 project 页面生命周期存在。

| 工具名 | 参数 | 前端效果 | 需用户确认 |
|--------|------|----------|-----------|
| `navigate_to_org` | — | `router.push(.../workspace)` | 否 |
| `navigate_to_model` | `db: string, model: string` | 跳转到模型编辑器 | 否 |
| `navigate_to_data` | `db: string, model: string` | 跳转到数据视图 | 否 |
| `open_create_model` | `db: string, prefill?: {name, title, description}` | 打开新建模型 Sheet，预填 | 是（用户点 Create） |
| `open_create_record` | `model: string, db: string, prefill?: Record<string, unknown>` | 打开新建记录 Sheet，预填字段值 | 是（用户点 Save） |
| `open_edit_record` | `model: string, db: string, record_id: string, patch: Record<string, unknown>` | 打开编辑 Sheet，patch 字段已填入 | 是（用户点 Save） |
| `highlight_records` | `model: string, record_ids: string[], reason: string` | 表格对应行高亮，hover 显示 reason | 否 |
| `set_filter` | `filter_json: string` | FilterPanel 应用筛选 | 否（已有） |
| `clear_filter` | — | 清空筛选 | 否（已有） |

**Draft State 机制：**
- `open_create_record` = 等价于用户点"新建"按钮 + 预填字段
- `open_edit_record` = 等价于用户点"编辑"按钮 + 填入 patch 字段
- 表单内的 Save/Delete 按钮不变，仍走现有 mutation 逻辑
- **无需引入新的 draft state 管理**，直接复用现有表单 state

**`useCopilotReadable` 内容：**
```ts
{
  layer: "project",
  orgName: string,
  projectSlug: string,
  currentModel?: string,
  currentDb?: string,
  availableActions: ["navigate_to_org", "navigate_to_model", "navigate_to_data",
                     "open_create_model", "open_create_record", "open_edit_record",
                     "highlight_records", "set_filter", "clear_filter"]
}
```

---

## Python Agent 工具裁剪

### 保留（只读）

| 工具 | 理由 |
|------|------|
| `list_projects` | 知道有哪些项目可跳转 |
| `list_models` | 知道有哪些模型 |
| `get_model_fields` | 知道字段名才能预填表单 |
| `query_model` | 读数据展示给用户 |
| `nl2filter` | 生成 filter JSON 配合 `set_filter` |

### 移除（写操作转移前端）

| 工具 | 替代 |
|------|------|
| `create_record` | → `open_create_record` 前端工具 |
| `update_record` | → `open_edit_record` 前端工具 |
| `delete_record` | 删除必须用户手动触发，agent 无权调用 |

---

## AgentState 新增字段

```python
class AgentState(TypedDict):
    messages: Annotated[list, add_messages]
    authorization: str
    org_name: str
    project_slug: str
    layer: str          # "org" | "project" | ""
    current_model: str  # 当前路由中的模型名
    current_db: str     # 当前路由中的数据库名
```

## System Prompt 层级逻辑

```python
if state.get("layer") == "org":
    # 限制：不可调用 list_models、query_model 等 project 工具
    # 引导：如需操作项目数据，先调 navigate_to_project

elif state.get("layer") == "project":
    # 写操作只预填表单，用户确认后才真正保存
    # 如需 org 级操作，调用 navigate_to_org
```

---

## 典型交互流程

### 场景：用户在 org 页说"帮我看 test 项目的 users 表数据"

```
1. Agent 读 state.layer = "org"
2. Agent 调 navigate_to_project("test")  → 页面跳转
3. Project 工具集加载，state.layer = "project"
4. Agent 调 list_models("maindb")        → 确认 users 存在
5. Agent 调 navigate_to_data("maindb", "users")  → 跳转数据视图
6. Agent 调 query_model(...)             → 读数据，展示在 sidebar
```

### 场景：用户说"帮我新建一条 users 记录，名字叫张三，年龄 25"

```
1. Agent 调 get_model_fields(users)      → 知道字段 name, age
2. Agent 调 open_create_record("users", "maindb", {name: "张三", age: 25})
3. 前端：新建表单打开，字段已预填
4. 用户确认 → 点 Save → 走已有 mutation
```

### 场景：用户说"把筛选条件设为年龄大于 18"

```
1. Agent 调 nl2filter("年龄大于18", ["id", "name", "age"])
   → {"age": {"gt": 18}}
2. Agent 调 set_filter('{"age":{"gt":18}}')
3. 前端：FilterPanel 自动更新，表格刷新
```

---

## 文件变更地图

### 新建
- `modelcraft-front/src/web/components/features/copilot/OrgCopilotActions.tsx`
- `modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx`

### 修改
- `modelcraft-front/src/app/org/[orgName]/layout.tsx` — 引入 `OrgCopilotActions`，`useCopilotReadable`
- `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/layout.tsx` — 引入 `ProjectCopilotActions`，`useCopilotReadable`
- `modelcraft-agent/agent.py` — 移除 3 个写工具，新增 `layer`/`current_model`/`current_db` 到 state 和 system prompt
