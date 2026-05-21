# AI 导航助手设计文档

**日期**: 2026-05-21  
**状态**: 已确认，待实现  
**范围**: MVP 第一阶段（导航 + 高亮，不修改数据）

---

## 一、核心原则

> **前端声明"哪些页面可以导航、哪些区域可以高亮"，后端 Agent 根据用户问题生成候选方案，用户点击后前端才真正执行导航和高亮。**

系统角色分工：

```
Agent         = 方案生成器（理解意图，返回 Proposal）
Frontend      = 方案展示器（渲染 AiProposalCard）
User          = 最终决策者（点击后执行）
ActionExecutor = 受控执行器（执行 actions 序列）
```

**MVP 阶段约束**：AI 不修改表单数据、不提交、不删除、不直接点击按钮。只做两件事：

1. 导航：带用户去正确页面
2. 高亮：指出用户应该看哪里、点哪里

---

## 二、技术栈

- **前端框架**: 保留 CopilotKit（UI + Action 注册）
- **Action 数量**: 只有 2 个（`ui.navigate`、`ui.highlight`），替换现有 15+ 个碎片化 action
- **后端**: 现有 Python Agent（LangGraph），重新设计响应协议

---

## 三、前端声明层

### 3.1 `routeCatalog` — 可导航页面目录

前端手动维护（或从路由配置自动生成），是 Agent 判断"去哪儿"的唯一依据。

```ts
type RouteCatalogEntry = {
  route: string
  title: string
  description: string
  keywords: string[]
  params?: Record<string, string>  // 动态参数说明，如 { tab: "data-permission" }
}

const routeCatalog: RouteCatalogEntry[] = [
  {
    route: "/org/:orgName/settings",
    title: "租户设置",
    description: "配置租户基础信息、权限、数据源、脱敏规则",
    keywords: ["租户", "权限", "脱敏", "敏感字段", "数据权限"],
  },
  {
    route: "/org/:orgName/project/:projectSlug/models",
    title: "数据模型管理",
    description: "创建和管理数据模型、字段、枚举、外键关系",
    keywords: ["模型", "字段", "枚举", "外键", "数据建模"],
  },
  {
    route: "/org/:orgName/project/:projectSlug/permissions",
    title: "权限配置",
    description: "配置模型级别的访问权限、行级安全策略",
    keywords: ["权限", "RBAC", "行级安全", "RLS", "访问控制"],
  },
  // ... 其他页面
]
```

### 3.2 `<AiTarget>` 组件 — 可高亮区域声明

页面组件主动声明自己可以被 AI 高亮，不依赖 CSS selector。

```tsx
// 用法：包裹任何需要被 AI 指引到的 UI 区域
<AiTarget
  id="sensitive-mask-section"
  label="敏感字段脱敏配置"
  description="配置字段级脱敏规则，控制敏感数据的展示方式"
  type="section"
>
  <SensitiveFieldMaskPanel />
</AiTarget>
```

**`<AiTarget>` 内部行为**：
1. 给 DOM 根元素加 `data-ai-target="sensitive-mask-section"` 属性
2. 组件挂载时自动注册到 `AiTargetRegistry`
3. 组件卸载时自动从 `AiTargetRegistry` 移除

### 3.3 `AiTargetRegistry` — 运行时高亮目标注册表

```ts
type AiTargetEntry = {
  id: string
  label: string
  description?: string
  type: "field" | "button" | "section" | "tableRow" | "tab" | "menu"
  element: HTMLElement
}

// 运行时结构，由 <AiTarget> 组件自动维护
const aiTargetRegistry = new Map<string, AiTargetEntry>()
```

**为什么不用 CSS Selector**：

```
❌ 不推荐
{ "selector": ".ant-form-item:nth-child(3) .ant-input" }
// 页面结构一改就失效

✅ 推荐
{ "targetId": "customer-phone-field" }
// UI 随便改，只要 id 稳定 AI 就能找到
```

---

## 四、CopilotKit 集成方式

### 4.1 一个注册 Action：`show_navigation_proposal`

后端 Agent 不直接调用 `ui.navigate` / `ui.highlight`，而是调用**一个** CopilotKit action 把整个 Proposal 传给前端渲染。

```ts
useCopilotAction({
  name: "show_navigation_proposal",
  description: "展示导航方案候选项，让用户选择后再执行。",
  parameters: [
    {
      name: "response",
      type: "object",
      description: "AgentUiResponse 格式的 Proposal",
    },
  ],
  // CopilotKit render 函数：把结构化数据渲染成 AiProposalCard
  render: ({ args }) => (
    <AiProposalCard
      message={args.response.message}
      candidates={args.response.candidates}
      onSelect={(candidate) => executeActions(candidate.actions)}
    />
  ),
})
```

### 4.2 `executeActions` — 前端受控执行函数

用户点击候选项后，由前端直接调用执行，不经过 CopilotKit 工具调用：

```ts
async function executeActions(actions: AiAction[]) {
  for (const action of actions) {
    if (action.type === "ui.navigate") {
      const url = buildUrl(action.args.route, action.args.params)
      await router.push(url)
      // 等待新页面的 AiTarget 组件挂载完成再继续
      // 实现方式：监听 AiTargetRegistry 的 "ready" 事件，或等待特定 targetId 出现
      await waitForAiTargetsReady()
    }
    if (action.type === "ui.highlight") {
      const entry = aiTargetRegistry.get(action.args.targetId)
      if (!entry) {
        console.warn(`[AI] targetId "${action.args.targetId}" not in registry`)
        return
      }
      highlightElement(entry.element, {
        message: action.args.message,
        scrollIntoView: action.args.scrollIntoView ?? true,
        durationMs: action.args.durationMs ?? 5000,
      })
    }
  }
}
```

**安全约束**：

```
1. Agent 只能调用 show_navigation_proposal，不能直接操作 DOM
2. candidates.actions 只能包含 routeCatalog 里的路由
3. candidates.actions 只能包含 AiTargetRegistry 里注册的 targetId
4. 不允许 Agent 传 CSS selector
5. 不允许 Agent 修改表单、提交、删除
6. executeActions 由前端用户点击触发，不自动执行
```

---

## 五、后端响应协议

### 5.1 统一响应类型

后端**永远**返回 Proposal，不直接返回立即执行的 Action。

```ts
type AgentUiResponse = {
  kind: "proposal"
  proposalType: "navigation" | "highlight" | "mixed"
  message: string       // 自然语言解释
  query: string         // 用户原始问题
  candidates: NavigationCandidate[]
}

type NavigationCandidate = {
  id: string
  title: string
  description?: string
  category?: "page" | "model" | "table" | "field" | "setting" | "action"
  confidence?: number   // 0-1，Agent 的置信度
  isPrimary?: boolean   // 是否为首推方案
  actions: AiAction[]   // 用户点击后要执行的动作序列
}

type AiAction =
  | {
      type: "ui.navigate"
      args: {
        route: string
        params?: Record<string, any>
        reason?: string
      }
    }
  | {
      type: "ui.highlight"
      args: {
        targetId: string
        targetType?: "field" | "button" | "section" | "tableRow" | "tab" | "menu"
        label?: string
        message?: string
        durationMs?: number
        scrollIntoView?: boolean
      }
    }
```

### 5.2 精确结果示例（1 个候选）

```json
{
  "kind": "proposal",
  "proposalType": "navigation",
  "query": "去项目管理",
  "message": "我找到了最可能的位置：",
  "candidates": [
    {
      "id": "page-project-management",
      "title": "项目管理",
      "category": "page",
      "description": "查看、创建和管理项目。",
      "confidence": 0.96,
      "isPrimary": true,
      "actions": [
        {
          "type": "ui.navigate",
          "args": {
            "route": "/org/acme/projects",
            "reason": "用户想进入项目管理页面"
          }
        }
      ]
    }
  ]
}
```

### 5.3 模糊结果示例（多个候选）

```json
{
  "kind": "proposal",
  "proposalType": "mixed",
  "query": "项目",
  "message": "你说的"项目"可能指以下几个位置：",
  "candidates": [
    {
      "id": "page-project-management",
      "title": "项目管理",
      "category": "page",
      "description": "查看、创建和管理项目。",
      "confidence": 0.78,
      "isPrimary": true,
      "actions": [
        { "type": "ui.navigate", "args": { "route": "/org/acme/projects" } }
      ]
    },
    {
      "id": "model-project",
      "title": "项目模型",
      "category": "model",
      "description": "配置项目模型的字段、表单和权限。",
      "confidence": 0.71,
      "actions": [
        { "type": "ui.navigate", "args": { "route": "/org/acme/project/main/models" } },
        {
          "type": "ui.highlight",
          "args": {
            "targetId": "model-project-field-list",
            "message": "这里可以配置项目模型字段。"
          }
        }
      ]
    }
  ]
}
```

---

## 六、前端执行流程

### 6.1 `AiProposalCard` 组件

```tsx
<AiProposalCard
  message={response.message}
  candidates={response.candidates}
  onSelect={(candidate) => executeActions(candidate.actions)}
/>
```

展示逻辑：
- 候选项 ≥ 2：显示多个卡片，用户选择
- 候选项 = 1：也显示一张卡片（不自动执行），`isPrimary: true` 时加"推荐"标签
- 无候选项：显示"未找到相关页面"提示

### 6.2 完整流程

```
用户输入自然语言
  ↓
前端把当前页面 + routeCatalog + 当前页 aiTargetRegistry 传给 Agent
  ↓
Agent 解析意图，生成 candidates
  ↓
前端渲染 AiProposalCard
  ↓
用户点击某一候选项
  ↓
ActionExecutor 串行执行该候选的 actions
  ├─ ui.navigate → 跳转页面，等待 AiTargets 挂载
  └─ ui.highlight → 滚动到目标元素并高亮
```

---

## 七、高亮视觉效果规范

沿用现有 `highlight-element.ts`，但统一样式：

- Amber ring：`ring-4 ring-amber-400 ring-offset-4`
- 背景淡显：`bg-amber-50`
- 入场动画：`animate-pulse`（持续 1s 后停止）
- 工具提示：在元素旁显示 `message` 文字（Tooltip 样式）
- 自动消失：`durationMs` 后移除所有样式
- 自动滚动：`scrollIntoView({ behavior: "smooth", block: "center" })`

---

## 八、MVP 范围

**本次实现**：

| 模块 | 内容 |
|------|------|
| `routeCatalog` | 维护所有可导航页面的目录文件 |
| `<AiTarget>` 组件 | 声明可高亮区域，自动注册/注销 |
| `AiTargetRegistry` | 运行时 Map，收集当前页的高亮目标 |
| `ui.navigate` action | 替换现有 15+ 个导航 action |
| `ui.highlight` action | 替换现有碎片化 highlight action |
| `ActionExecutor` | 串行执行 actions，处理 navigate→highlight 时序 |
| `AiProposalCard` | 渲染候选项，用户点击触发执行 |
| 后端 Agent 协议 | 返回 `NavigationProposalResponse` 格式 |

**暂不实现**：

```
suggest 接口（后续）
form.patchFields
form.submit
delete / update
自动点击按钮
自动填写表单
```

---

## 九、为什么用 Proposal 而不是 Direct Action

| 维度 | Direct Action | Proposal（本方案）|
|------|--------------|------------------|
| 安全性 | AI 可能理解错直接跳转 | 用户点击才执行，不误操作 |
| 交互一致性 | 需要区分明确/模糊两种响应 | 前端永远只处理一种结构 |
| 用户信任感 | AI "擅自"操作页面 | AI 提建议，用户做决定 |
| 可扩展性 | 每种新能力要新响应格式 | actions 数组可以持续扩展 |
