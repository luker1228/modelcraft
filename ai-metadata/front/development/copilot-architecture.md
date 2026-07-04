# Copilot 架构文档

> 描述 `src/web/components/features/copilot/` 目录的 AI 助手体系设计：CopilotKit Provider、知识注入机制、前端工具注册，以及 AI 高亮导航流程。

---

## 概述

ModelCraft Copilot 基于 **CopilotKit** 构建，使用 `CopilotWrapper` 挂载到 `/org/*` 路由（设计态，项目/模型/权限管理）：

| 实例 | Wrapper | Agent | 使用场景 |
|------|---------|-------|----------|
| 管理员（Tenant） | `CopilotWrapper` | `modelcraft_admin_agent` | `/org/*` 路由 — 设计态，项目/模型/权限管理 |

---

## 知识分层架构

知识按**是否依赖运行时**划分为两层：

```
后端 Agent（静态，不走网络）          前端 CopilotKit context（每次请求携带）
─────────────────────────────         ──────────────────────────────────────
system prompt                         routeCatalog（可导航页目录）
  ├── 操作手册                         aiTargets（当前页已注册高亮目标）
  └── 工具：get_page_knowledge         项目依赖策略

_PAGE_KNOWLEDGE 字典（按需加载）
  ├── model-editor: workflow
  ├── databases: workflow
  ├── enums / roles / settings ...
  └── 索引写入 system prompt
      Agent 需要时调用工具取完整内容
```

**只有 aiTargets 必须留前端**——它是运行时 DOM 注册，后端无法感知。

---

## 文件清单

```
前端 src/web/components/features/copilot/
├── CopilotProvider.tsx          # Provider 入口（CopilotWrapper）
├── SharedCopilotActions.tsx     # 共享前端工具：ui_present_proposal
├── AdminCopilotKnowledge.tsx    # 管理员侧边栏快捷提示（静态知识已迁移后端）
├── RoutePageKnowledge.tsx       # ⚠️ DEPRECATED — 已迁移后端，保留空壳
├── AICapabilityReadable.tsx     # 将 aiTargets + routeCatalog 注入到 Agent 上下文
├── AiTarget.tsx                 # 声明式 UI 高亮目标注册组件
├── AIChipMessage.tsx            # 自定义 AssistantMessage：渲染 [ACTION:id] 为可点击 Chip
├── AiProposalCard.tsx           # 渲染 AI 导航提案卡片（候选项列表）
├── types.ts                     # 核心类型（AiAction / ProposalCandidate / AgentUiResponse）
├── ProjectCopilotActions.tsx    # ⚠️ DEPRECATED — 已废弃，保留空壳
├── OrgCopilotActions.tsx        # ⚠️ DEPRECATED — 已废弃，保留空壳
└── AICapabilityDebugPanel.tsx   # 开发调试面板（可选挂载）

后端 modelcraft-agent/agents/
├── tools.py                     # _PAGE_KNOWLEDGE 字典 + get_page_knowledge 工具
└── admin_agent.py               # system prompt 含知识索引 + get_page_knowledge 注册
```

---

## Provider 结构

### `CopilotWrapper`（管理员）

```tsx
<Suspense fallback={children}>
  <CopilotKit runtimeUrl="/api/copilotkit" agent="modelcraft_admin_agent" ...>
    <SharedCopilotActions />       {/* 共享工具 */}
    <AdminCopilotKnowledge />      {/* 快捷提示（仅 UI） */}
    <RoutePageKnowledge />         {/* 空壳，已废弃 */}
    {children}
    <CopilotSidebar AssistantMessage={AIChipMessage} />
  </CopilotKit>
</Suspense>
```

**注意**：`AICapabilityReadable` 不在 `CopilotWrapper` 中直接挂载，而是由各**页面布局**按需挂载（需要在 `<CopilotKit>` 树内）。

---

## 前端工具（CopilotKit Actions）

### 共享工具（`SharedCopilotActions`）

| 工具名 | 用途 |
|--------|------|
| `ui_present_proposal` | Agent 发送导航提案卡片，用户点击候选项执行跳转/高亮/澄清 |

`ui_present_proposal` 是核心工具，参数结构为 `AgentUiResponse`：

```ts
type AgentUiResponse = {
  kind: 'proposal'
  proposalId: string
  proposalType: 'navigation' | 'highlight' | 'clarification' | 'mixed'
  message: string
  query: string
  candidates: ProposalCandidate[]  // ActionCandidate | ClarificationCandidate
}
```

每个 `ActionCandidate` 包含一组 `AiAction`：

```ts
type AiAction =
  | { type: 'ui.navigate'; args: { route: string } }
  | { type: 'ui.highlight'; args: { targetId: string; message?: string; durationMs?: number } }
  | { type: 'ui.guide'; args: { route?: string; targetId?: string; message?: string } }
```

**注意**：`ui.navigate` 的 `route` 必须从 `routeCatalog`（由 `AICapabilityReadable` 注入）中的 `routeTemplate` 派生，将 `:orgName`/`:projectSlug` 替换为实际值。`ui.highlight` 的 `targetId` 必须是 `AICapabilityContext` 中已注册的 AiTarget ID。

---

## 后端知识体系

### 页面知识（`tools.py` → `_PAGE_KNOWLEDGE`）

采用**渐进式索引**设计：索引常驻 system prompt，完整内容按需加载。

```
system prompt 中只写索引：
  "可用页面知识索引: model-editor, databases, enums, roles, ..."
  "需要页面操作指南时调用 get_page_knowledge(page)"

_PAGE_KNOWLEDGE 字典静态存在内存：
  "model-editor" → { name, description, workflow }
  "databases"    → { name, description, workflow }
  ...

get_page_knowledge(page) 工具：
  Agent 识别 current_route 末段 → 调用工具 → 获取完整 workflow → 生成回复
```

当前覆盖页面：`model-editor`、`databases`、`enums`、`roles`、`identity-settings`、`settings`、`workspace`、`developers`、`cluster`

### 静态操作手册

管理员操作手册、常见问题排查直接写在对应 agent 的 `system_msg` 中，不经过前端传输，利用 **prompt cache** 降低 token 成本。

---

## 前端知识注入（仍留前端的部分）

仅以下两类通过 `useCopilotReadable` 注入，因为它们依赖运行时状态：

| 组件 | 注入内容 | 原因 |
|------|---------|------|
| `AICapabilityReadable` | `routeCatalog`（全局路由目录） | 前端维护，Agent 用它生成 route |
| `AICapabilityReadable` | `aiTargets`（当前页高亮目标列表） | 运行时 DOM 注册，后端无法感知 |
| `AICapabilityReadable` | 项目依赖策略 | 配合 routeCatalog 使用 |

---

## AiTarget：高亮目标注册

`AiTarget` 是一个声明式 wrapper，自动向 `AICapabilityContext` 注册/注销当前页面的可高亮 DOM 元素：

```tsx
<AiTarget id="create-model-btn" label="新建模型按钮" type="button" description="点击后弹出新建模型表单">
  <Button>新建模型</Button>
</AiTarget>
```

- 注册的信息（`id`、`label`、`description`、`type`）通过 `AICapabilityReadable` 注入 Agent 上下文
- Agent 在 `ui.highlight` action 中引用的 `targetId` 必须是已注册的 ID
- `AiTarget` 挂载时注册、卸载时自动注销（基于 `useEffect`）

**类型枚举**（`AiTargetType`）：`field` | `button` | `section` | `tableRow` | `tab` | `menu`

---

## AIChipMessage：自定义消息渲染

`AIChipMessage` 替换 CopilotKit 默认的 `AssistantMessage`，支持在 AI 回复中嵌入可点击 Chip 按钮：

- Agent 在消息中插入 `[ACTION:targetId]` 标记
- `AIChipMessage` 解析标记，将其渲染为琥珀色 Chip 按钮
- 点击 Chip 后调用 `highlightElement` 高亮对应的 DOM 元素
- 未注册的 `targetId` 渲染为禁用的灰色 Chip

```
Agent 回复示例：
"请点击这里 [ACTION:create-model-btn] 来新建数据模型。"
→ 渲染为：文字 + ✨新建模型按钮（可点击，触发高亮）
```

---

## AiProposalCard：提案卡片渲染

`AiProposalCard` 在 `ui_present_proposal` 的 `render` 函数中渲染，展示候选操作列表：

- `action_candidate`：点击后执行 `actions` 中的 `ui.navigate` / `ui.highlight` / `ui.guide`
- `clarification_candidate`：点击后向 Copilot 发送澄清消息（用 `payload` 中的 `userMeaning` 作为消息内容）
- `isPrimary=true` 的候选项显示"推荐"标签
- 交互逻辑由 `useNavigationProposal` Hook 处理

---

## 废弃组件

| 组件 | 状态 | 原因 |
|------|------|------|
| `RoutePageKnowledge` | ⚠️ 空壳 | 页面知识迁移到后端 `_PAGE_KNOWLEDGE` + `get_page_knowledge` 工具 |
| `OrgCopilotActions` | ⚠️ 空壳 | 导航 actions 已被 `ui_present_proposal` 替代 |
| `ProjectCopilotActions` | ⚠️ 空壳 | 导航 actions 已被 `ui_present_proposal` 替代 |
| `EndUserCopilotWrapper` | ❌ 已移除 | 终端用户前端页面已移除 |
| `EndUserCopilotActions` | ❌ 已移除 | 终端用户前端页面已移除 |
| `EndUserCopilotKnowledge` | ❌ 已移除 | 终端用户前端页面已移除 |

---

## 扩展指南

### 添加新的 AiTarget（高亮目标）

```tsx
import { AiTarget } from '@web/components/features/copilot/AiTarget'

<AiTarget id="rbac-create-role-btn" label="新建角色按钮" type="button">
  <Button>新建角色</Button>
</AiTarget>
```

### 添加新的页面知识

在 `modelcraft-agent/agents/tools.py` 的 `_PAGE_KNOWLEDGE` 字典中新增条目：

```python
"new-page-segment": {
    "name": "页面中文名",
    "description": "页面功能说明",
    "workflow": "1. 步骤一\n2. 步骤二",
},
```

Agent 会在 system prompt 索引中自动包含新 key，无需其他改动。

### 新增快捷提示

在 `AdminCopilotKnowledge.tsx` 的 `*_SUGGESTIONS` 数组中添加：

```ts
{ title: '按钮显示文字', message: '发送给 Agent 的完整消息' },
```
