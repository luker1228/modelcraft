# CopilotKit 前端工具设计

> 状态更新：本文档描述的是“双 agent”设计稿。当前实际代码只保留 `modelcraft_admin_agent`，`modelcraft_enduser_agent` 与 `/copilotkit/enduser` 未启用。

**日期**: 2026-05-20  
**范围**: modelcraft-front + modelcraft-agent  
**状态**: 已批准，待实现

---

## 背景

ModelCraft 存在两类用户：租户管理员（tenant-admin）和终端用户（end-user），两者使用场景完全不同。本文是当时为“双 agent”方向准备的设计稿。

本设计将其拆分为两个独立 agent，各自有专属工具集、知识库和引导内容，同时补全缺失的导航工具，并通过 CopilotSidebar `suggestions` 提供开箱即用的快捷入口。

---

## 一、Agent 架构

### 两个独立 Agent

| 项目 | `modelcraft_admin_agent` | `modelcraft_enduser_agent` |
|------|--------------------------|---------------------------|
| 服务对象 | 租户管理员 | 终端用户 |
| 挂载位置 | `/org/*` 布局（`CopilotWrapper`） | `/end-user/*` 布局（`EndUserCopilotWrapper`） |
| 前端组件 | `agent="modelcraft_admin_agent"` | `agent="modelcraft_enduser_agent"` |

两个 agent 注册在同一 FastAPI 服务，`main.py` 通过 `LangGraphAGUIAgent` 分别命名注册。Agent 图定义分拆到独立文件：

```
modelcraft-agent/
├── main.py                  # 注册两个 agent
├── agents/
│   ├── admin_agent.py       # admin graph + tools + knowledge
│   └── enduser_agent.py     # end-user graph + tools + knowledge
└── client/
    └── graphql_client.py    # 共享 GraphQL 客户端（不变）
```

### Authorization 注入

两个 agent 沿用现有 `main.py` 的 header 注入方式：每次请求从 HTTP `Authorization` header 读取 token，始终写入 `state["authorization"]`。

---

## 二、前端工具全集

### 2.1 共享工具（两个 agent 均注册）

| 工具名 | 参数 | 说明 |
|--------|------|------|
| `show_toast` | `message: string`, `type?: "success"\|"error"\|"info"\|"warning"` | Agent 向用户推送临时通知，无需用户在聊天框内查看 |
| `set_filter` | `filter_json: string` | 设置数据表格的 where 筛选条件（ModelCraft filter JSON） |
| `clear_filter` | — | 清空所有筛选条件，恢复全量数据展示 |
| `highlight_records` | `record_ids: string[]`, `reason: string` | 在数据表格中高亮指定记录行 |

### 2.2 Admin Agent 专属工具

#### Org 层（已有，保留）

| 工具名 | 参数 | 说明 |
|--------|------|------|
| `navigate_to_project` | `slug: string` | 跳转到指定项目工作区 |
| `navigate_to_settings` | — | 跳转到 org 设置页 |
| `open_create_project` | `slug?, title?, description?` | 打开新建项目表单（预填，用户手动确认） |
| `highlight_project` | `slug: string`, `reason: string` | 在项目列表中高亮指定项目 |

#### Project 层（已有，保留）

| 工具名 | 参数 | 说明 |
|--------|------|------|
| `navigate_to_org` | — | 返回 org workspace 页面 |
| `navigate_to_model` | `db: string`, `model: string` | 跳转到模型编辑器 |
| `navigate_to_data` | `db: string`, `model: string` | 跳转到数据视图 |
| `open_create_model` | `db: string`, `name?, title?` | 打开新建模型表单（预填） |
| `open_create_record` | `model: string`, `db: string`, `prefill?` | 打开新建记录表单（预填） |
| `open_edit_record` | `model: string`, `db: string`, `record_id: string`, `patch: object` | 打开编辑记录表单（预填修改字段） |

#### Project 层（新增）

| 工具名 | 参数 | 说明 |
|--------|------|------|
| `navigate_to_enums` | `db?: string` | 跳转到枚举管理页 |
| `navigate_to_cluster` | — | 跳转到数据库集群配置页 |
| `navigate_to_rbac` | `section?: "roles"\|"users"\|"bundles"\|"permissions"` | 跳转到 RBAC 对应子页，默认 roles |
| `navigate_to_end_users` | — | 跳转到 end-user 管理页 |

### 2.3 End User Agent 专属工具

| 工具名 | 参数 | 说明 |
|--------|------|------|
| `navigate_to_project` | `slug: string` | 切换项目（end-user 路由） |
| `navigate_to_workspace` | — | 返回项目选择页 |

---

## 三、CopilotSidebar Suggestions（快捷入口）

通过 `CopilotSidebar` 的 `suggestions` prop 在聊天框底部展示可点击的引导入口，用户无需打字即可触发常用场景。

### Admin Agent

```
新手引导：带我完成初始配置
帮我创建第一个数据模型
数据库连不上，帮我排查
我有哪些项目？
解释当前页面的功能
```

### End User Agent

```
新手引导：带我了解这个系统
用自然语言帮我筛选数据
我看不到想要的数据，帮我排查
这些字段分别是什么意思？
帮我统计一下数据
```

---

## 四、知识库（useCopilotReadable）

通过 `useCopilotReadable` 在 agent 上下文中注入操作手册。Agent 读手册后，用工具执行对应步骤。知识库组件与 `*CopilotActions` 同级，挂载在 layout 层。

### 4.1 Admin Agent 新手引导

```
新手引导共 5 步：

Step 1 [创建项目]
  目标：创建第一个项目
  工具：open_create_project(slug, title)
  验证：调用 list_projects，确认项目出现

Step 2 [配置数据库集群]
  目标：连接一个数据库
  工具：navigate_to_cluster，引导用户在页面上手动填写连接信息
  提示：集群配置需要用户提供数据库连接信息，agent 无法代替操作

Step 3 [创建数据模型]
  目标：在项目下创建第一个模型
  工具：navigate_to_project → open_create_model(db, name)
  验证：调用 list_models 确认模型存在

Step 4 [添加字段]
  目标：给模型添加字段
  工具：navigate_to_model(db, model)
  提示：字段编辑在右侧面板，由用户手动操作

Step 5 [查看数据]
  目标：进入数据视图，确认配置完成
  工具：navigate_to_data(db, model)
```

### 4.2 Admin Agent 问题排查

```
问题：数据库连接失败
  → show_toast("正在带你去检查集群配置", "info")
  → navigate_to_cluster，引导检查 host/port/credentials

问题：找不到模型或字段
  → 调用 list_models(db) 确认模型名是否正确
  → list_models 返回空则模型未创建，建议执行 Step 3

问题：权限被拒绝
  → navigate_to_rbac(section="users")，检查用户角色分配

问题：字段显示异常
  → navigate_to_model(db, model)，检查字段类型和配置
```

### 4.3 End User Agent 新手引导

```
新手引导共 3 步：

Step 1 [了解当前数据]
  目标：知道项目里有哪些数据
  工具：调用 list_models，用自然语言介绍每个模型的用途

Step 2 [学会筛选]
  目标：用自然语言筛选数据
  步骤：
    1. 询问用户想查什么
    2. 调用 nl2filter(natural_language, field_names) 生成 filter JSON
    3. 调用 set_filter(filter_json) 应用筛选
  示例引导语：「你可以说"帮我找金额大于 1000 的订单"，我来帮你筛选」

Step 3 [理解字段含义]
  目标：用户看懂表格里的每一列
  工具：get_model_fields(model) → 逐字段用中文解释
```

### 4.4 End User Agent 问题排查

```
问题：看不到数据
  → 先调用 clear_filter 排除筛选遮挡
  → 若仍无数据，说明可能没有访问权限，提示联系管理员

问题：不知道怎么筛选
  → 引导用户用自然语言描述需求
  → 执行 nl2filter + set_filter

问题：字段看不懂
  → 调用 get_model_fields(model)，逐字段解释含义和示例值

问题：数据量太大，加载慢
  → 引导用户说出筛选条件
  → 用 nl2filter 缩小数据范围后再查看
```

---

## 五、文件变更一览

### modelcraft-agent

```
agents/
  admin_agent.py          新建：admin graph、tools、system prompt、knowledge
  enduser_agent.py        新建：end-user graph、tools、system prompt、knowledge
main.py                   修改：注册两个 agent，删除旧 modelcraft_agent
agent.py                  删除（逻辑迁移到 agents/）
```

### modelcraft-front

```
src/web/components/features/copilot/
  AdminCopilotKnowledge.tsx      新建：admin useCopilotReadable 知识库组件
  EndUserCopilotKnowledge.tsx    新建：end-user useCopilotReadable 知识库组件
  AdminCopilotActions.tsx        新建：整合现有 OrgCopilotActions + ProjectCopilotActions + 新增工具
  EndUserCopilotActions.tsx      新建：end-user 专属工具（navigate_to_project、navigate_to_workspace）
  CopilotProvider.tsx            修改：admin/enduser 分别传 agent name 和 suggestions
  OrgCopilotActions.tsx          保留（被 AdminCopilotActions 复用）
  ProjectCopilotActions.tsx      保留（被 AdminCopilotActions 复用）
```

---

## 六、关键约束

1. **写操作不自动执行** — `open_create_*` 和 `open_edit_*` 只预填表单，用户必须手动点击 Save/Create
2. **show_toast 是单向通知** — agent 不等待 toast 的用户响应，仅作为信息反馈
3. **knowledge base 只读** — `useCopilotReadable` 注入的内容仅供 LLM 参考，不暴露给用户
4. **两个 agent 完全独立** — 不共享 graph state，不共享 checkpointer thread
