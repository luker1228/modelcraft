# AI 页面感知与功能高亮 — 设计文档

**日期**: 2026-05-21  
**状态**: 已确认，待实现  
**范围**: `modelcraft-front` 前端 + LangGraph Agent system prompt

---

## 背景

当前 AI 助手（CopilotKit + LangGraph）在回答"如何新建模型"之类的问题时，只能输出纯文本导航说明，无法：

1. 感知当前页面有哪些功能可用
2. 在回复中提供可点击的操作入口
3. 高亮对应 UI 元素引导用户操作

本设计解决以上三个问题。

---

## 核心决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 页面感知机制 | 组件注册制 | 动态感知，随组件生命周期，直接持有 DOM ref |
| 建议呈现位置 | Chat 内嵌 Chips | 融入现有对话流，实现最简 |
| 高亮风格 | 琥珀色（`bg-amber-50 + ring-amber-300`） | 沿用现有表格行高亮，零新增视觉语言 |
| Chip→高亮连接 | 纯前端处理 | 点击即时响应，0 延迟，不经过 AI |

---

## 架构

### 数据流

```
① 组件挂载
   └─ useRegisterAICapability("create_model", "新建模型", buttonRef)
      └─ AICapabilityContext.register(...)

② 用户发消息
   └─ CopilotProvider 通过 useCopilotReadable 自动注入能力列表
      └─ AI 收到：「当前页面可用操作：create_model, connect_db, set_permissions」

③ AI 回复
   └─ 「可以点击 [ACTION:create_model] 新建模型」

④ 前端渲染
   └─ AIChipMessage 解析 [ACTION:id] → amber chip 按钮

⑤ 用户点击 chip
   └─ AICapabilityContext.getRef("create_model") → ref
      └─ highlightElement(ref) → 加 amber 样式，5s 后自动清除
```

---

## 新增 & 改动文件

### 新增文件

#### `src/web/contexts/AICapabilityContext.tsx`

全局能力注册表。提供 React Context + Provider。

```typescript
type AICapability = {
  id: string           // 唯一标识，如 "create_model"
  label: string        // 展示名，如 "新建模型"
  ref: RefObject<HTMLElement>
  description?: string // 可选，给 AI 的补充说明
}

interface AICapabilityContextValue {
  register: (capability: AICapability) => void
  unregister: (id: string) => void
  getAll: () => AICapability[]
  getRef: (id: string) => RefObject<HTMLElement> | undefined
}
```

- `AICapabilityProvider` 包裹整个应用，**必须是 `CopilotProvider` 的外层**（这样 `CopilotProvider` 内部才能读到 context）
- 内部用 `Map<string, AICapability>` 存储，`useState` 驱动重渲染（确保 `useCopilotReadable` 能感知到注册变化）

#### `src/web/hooks/useRegisterAICapability.ts`

组件专用 hook，mount 时注册，unmount 时自动注销。

```typescript
function useRegisterAICapability(
  id: string,
  label: string,
  ref: RefObject<HTMLElement>,
  description?: string
): void
```

用法示例：
```tsx
// ModelListPage.tsx
const createButtonRef = useRef<HTMLButtonElement>(null)
useRegisterAICapability("create_model", "新建模型", createButtonRef, "打开新建模型表单")

return <Button ref={createButtonRef}>+ 新建模型</Button>
```

#### `src/web/components/features/copilot/AIChipMessage.tsx`

Chat 消息渲染器，通过 CopilotKit 的 `renderMessage` prop 接入，替换默认的纯文本渲染。

```tsx
// CopilotProvider.tsx 中
<CopilotPopup
  renderMessage={(message) => <AIChipMessage message={message} />}
  ...
/>
```

- 解析消息文本中的 `[ACTION:action_id]` 标记
- 将其渲染为 amber chip 按钮（`bg-amber-50 border border-amber-300 text-amber-900 rounded-full`）
- 点击时：
  1. 通过 `AICapabilityContext.getRef(id)` 获取 ref
  2. 调用 `highlightElement(ref)` 高亮元素
  3. 自动 scroll 到目标元素（`scrollIntoView({ behavior: 'smooth', block: 'center' })`）

Chip 外观：
```
┌─────────────────────┐
│  ✨ 新建模型         │  ← bg-amber-50, border-amber-300, rounded-full
└─────────────────────┘
```

### 扩展文件

#### `src/web/contexts/WorkspaceAIRefContext.tsx`

增加通用 `highlightElement` 方法（现有 `setHighlight` 只作用于表格行）：

```typescript
highlightElement: (ref: RefObject<HTMLElement>, durationMs?: number) => void
```

实现：
- 对 `ref.current` 添加 class：`bg-amber-50 ring-2 ring-amber-300 transition-all`
- `durationMs` 默认 5000ms，timeout 后移除 class
- 若元素不在视口内，先 `scrollIntoView`

#### `src/web/components/features/copilot/CopilotProvider.tsx`

增加 `useCopilotReadable`，从 `AICapabilityContext` 读取能力列表并注入 AI：

```typescript
const capabilities = useAICapabilityContext()

useCopilotReadable({
  description: "当前页面可用的 UI 操作",
  value: capabilities.getAll().map(c => ({
    id: c.id,
    label: c.label,
    description: c.description,
  }))
})
```

#### 各页面组件（第一批，优先级 P0）

首批接入以下组件，覆盖核心用户旅程：

| 组件 | action_id | label |
|------|-----------|-------|
| `ModelListPage` 新建按钮 | `create_model` | 新建模型 |
| `ClusterPanel` 连接按钮 | `connect_db` | 连接数据库 |
| `ProjectSettings` 权限入口 | `set_permissions` | 设置权限 |
| `ModelListPage` 侧边导航 | `goto_models` | 进入数据模型 |
| `FieldListPage` 新建按钮 | `create_field` | 新建字段 |

---

## 后端 Agent 变更

### LangGraph Agent system prompt 增加

```
## UI 操作引导规则

当你需要引导用户使用页面上的某个功能时，在回复中插入 [ACTION:action_id] 标记。
前端会将其渲染为可点击按钮，用户点击后自动高亮对应 UI 元素。

规则：
1. 只使用系统提供的"当前页面可用操作"列表中的 action_id
2. 不要编造不存在的 action_id
3. 标记可以出现在句子中间，例如：「点击 [ACTION:create_model] 即可开始」
4. 如果当前页面没有相关操作，正常回答，不使用标记
```

---

## 边界情况处理

| 场景 | 处理方式 |
|------|----------|
| AI 使用了不存在的 action_id | chip 渲染为灰色禁用状态，显示 tooltip「该操作当前不可用」 |
| ref 对应元素已被卸载 | `highlightElement` 检查 `ref.current` 是否为 null，静默跳过 |
| 同一 action_id 被多次注册 | 后注册的覆盖前者（最新挂载的组件优先） |
| 能力列表为空 | `useCopilotReadable` 不注入（避免 AI 输出无效 `[ACTION:...]`） |

---

## 不在范围内（v1）

- 能力列表不持久化，刷新后重新注册（组件挂载驱动）
- 不支持键盘导航 chip
- 不支持 chip 触发后的自动下一步操作（仅高亮，不自动点击）
- 不支持动态更新 chip label（label 在注册时确定）

---

## 实现顺序建议

1. `AICapabilityContext` + `useRegisterAICapability` — 纯 React，无副作用，可独立测试
2. `WorkspaceAIRefContext` 增加 `highlightElement` — 扩展现有，影响面小
3. `AIChipMessage` 渲染器 — 接入 Context，视觉验证
4. `CopilotProvider` 增加 `useCopilotReadable` — 接通 AI
5. 后端 Agent system prompt 修改 — 最后，前后端联调
6. 各页面组件接入 `useRegisterAICapability` — 按 P0 列表逐一完成
