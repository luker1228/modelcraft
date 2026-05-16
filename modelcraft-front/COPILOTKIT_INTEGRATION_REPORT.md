# ModelCraft 前端 CopilotKit 集成探索报告

探索时间: 2026-05-16
项目路径: /data/home/lukemxjia/modelcraft/modelcraft-front

---

## 📋 执行摘要

### 核心发现：
1. **CopilotKit 已在 end-user（终端用户）端集成，租户端已有基础但需完善**
2. **tenant-admin 使用路径**: `/org/[orgName]/...` 系列路由
3. **end-user 使用路径**: `/end-user/[orgName]/projects/[projectSlug]/data` 及相关路由
4. CopilotKit 的引入采用了**按需加载和条件渲染**的策略
5. 有完整的 API 代理层 (`/api/copilotkit/route.ts`) 连接到 Python Agent 服务

---

## 🔍 详细探索结果

### 1️⃣ CopilotKit 依赖配置

**文件**: `package.json`

```json
{
  "@copilotkit/react-core": "^1.51.3",
  "@copilotkit/react-ui": "^1.51.3",
  "@copilotkit/runtime": "^1.51.3"
}
```

**Next.js 构建配置**: `next.config.mjs` (第 47-54 行)
```javascript
transpilePackages: [
  '@copilotkit/react-core',
  '@copilotkit/react-ui',
  '@copilotkit/runtime',
  'streamdown',
  'shiki',
  'mermaid',
],
```

---

### 2️⃣ CopilotKit 后端代理

**文件**: `src/app/api/copilotkit/route.ts` ✅

**功能**:
- BFF 代理层，将前端请求转发到 Python modelcraft-agent 服务
- 处理 SSE（Server-Sent Events）流式响应
- 注入授权信息：`Authorization` header、`Cookie` 等
- 在请求体中注入状态：`org_name`, `project_slug`, `authorization`

**关键代码片段**:
```typescript
const AGENT_SERVICE_URL = process.env.AGENT_SERVICE_URL ?? 'http://localhost:8000'
const upstreamUrl = `${AGENT_SERVICE_URL}/copilotkit/`

// 转发 Authorization
const authHeader = req.headers.get('Authorization') ?? ''
if (authHeader) headers.set('Authorization', authHeader)

// 注入状态到请求体
parsed.state = {
  authorization: authHeader,
  org_name: properties.orgName ?? '',
  project_slug: properties.projectSlug ?? '',
}
```

**⚠️ 当前问题**: `orgName` 在 route.ts 中等待接收（见源码注释第 51-52 行）

---

### 3️⃣ CopilotProvider 组件（核心包装器）

**文件**: `src/web/components/features/copilot/CopilotProvider.tsx` ✅

**导出 3 个主要组件**:

#### a) `CopilotProvider` (租户端使用)
- 完整的 CopilotKit 体验：sidebar + AI 助手
- 传递 `projectId`, `projectSlug`, `orgName` 给 AI agent

#### b) `CopilotWrapper` (租户端使用 - 带加载边界)
- 包装 CopilotProvider
- 提供 CopilotAvailableContext

#### c) `EndUserCopilotWrapper` (终端用户端使用 - 无侧边栏)
- 轻量级集成：仅提供上下文，不显示 sidebar
- 仅传递 `projectSlug`, `orgName`（不传 `projectId`）
- 用于 FilterPanel 中的 "✨ AI 查询" 功能

---

### 4️⃣ 当前集成状态

#### 🟢 End-User（终端用户端）- 已集成 ✅

**主要路由**: `/end-user/[orgName]/projects/[projectSlug]/data/...`

集成点：
1. `src/app/end-user/[orgName]/projects/[projectSlug]/data/layout.tsx` - 使用 `EndUserCopilotWrapper`
2. `src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx` - 挂载 `FilterCopilotActions`
3. `src/web/components/features/end-user-data/FilterPanel.tsx` - 显示 "✨ AI 查询" 按钮

#### 🟡 Tenant-Admin（租户端）- 部分集成

**主要路由**: `/org/[orgName]/project/[projectSlug]/...`

现状分析：
- ✅ 项目级 Layout (`src/app/org/[orgName]/project/[projectSlug]/layout.tsx`):
  - 已导入 `CopilotWrapper` 和 `AIAssistantButton`
  - 有按需加载逻辑
  - 导入了 CSS 样式
  
- ❌ 顶层 Layout (`src/app/org/[orgName]/layout.tsx`):
  - 只有 `OnboardingProvider` 包装
  - 未集成 CopilotKit

---

### 5️⃣ CopilotKit 可用性检测机制

**文件**: `src/web/components/features/end-user-data/FilterCopilotActions.tsx`

**方案**: 自定义 `CopilotAvailableContext`
```typescript
export const CopilotAvailableContext = createContext<boolean>(false)

export function useCopilotKitAvailable(): boolean {
  return useContext(CopilotAvailableContext)
}
```

**为什么需要**：
- `@copilotkit/react-core` 的 `CopilotContext` 有非 null 的默认值
- 直接使用 `useContext(CopilotContext) !== null` 不可靠
- 自定义上下文提供可靠的检测机制

---

### 6️⃣ 项目结构总览

```
src/app/
├── layout.tsx                              # ✅ 根 layout（Apollo + Query 提供者）
├── org/
│   └── [orgName]/
│       ├── layout.tsx                      # ❌ 无 CopilotKit
│       └── project/
│           └── [projectSlug]/
│               ├── layout.tsx              # ✅ CopilotWrapper 按需加载
│               ├── model-editor/
│               ├── settings/
│               └── ...
├── end-user/
│   └── [orgName]/
│       ├── projects/
│       │   └── [projectSlug]/
│       │       ├── data/
│       │       │   └── layout.tsx          # ✅ EndUserCopilotWrapper 包装
│       │       │       └── [modelId]/
│       │       │           └── page.tsx    # EndUserRecordWorkspace
│       │       └── ...
│       └── workspace/
├── api/
│   └── copilotkit/
│       └── route.ts                        # ✅ BFF 代理到 Python Agent
└── ...
```

---

## 💡 关键发现与设计模式

### 1. 两套 CopilotKit 集成方案

| 维度 | 租户端 (CopilotWrapper) | 终端用户端 (EndUserCopilotWrapper) |
|------|------------------------|----------------------------------|
| UI 体验 | Sidebar + 浮动按钮 | 仅上下文（无 UI） |
| 加载策略 | 按需加载（用户点击激活） | 始终加载 |
| 传递数据 | projectId, projectSlug, orgName | projectSlug, orgName |
| Agent | 有 agent="modelcraft_agent" | 无 agent 参数 |
| CSS 导入 | 是 | 否 |

### 2. 条件渲染约定

所有使用 CopilotKit hooks 的组件都必须：
```typescript
const hasCopilot = useCopilotKitAvailable()
// ...
{hasCopilot && <ComponentThatUsesCopilotKit />}
```

### 3. 已实现的 AI 功能

- ✅ End-User 端：AI 自然语言过滤（"✨ AI 查询"）
- ⚠️ Tenant 端：仅基础架构，无具体功能

---

## 📝 要在 Tenant-Admin 引入 CopilotKit 需要改哪些文件

### 当前状态
- ✅ 基础架构已到位（Provider、route、配置）
- ✅ 项目级 layout 已实现按需加载
- ⚠️ 完整功能需要补充

### 需要改进的文件

#### 1. 可选：在租户端其他路由添加 CopilotKit

**目标文件**:
- `src/app/org/[orgName]/settings/layout.tsx`
- `src/app/org/[orgName]/developers/layout.tsx`

**改进方式**: 参考 `src/app/org/[orgName]/project/[projectSlug]/layout.tsx` 的模式

#### 2. 需要检查：API 代理的 orgName 转发

**文件**: `src/app/api/copilotkit/route.ts` (第 53 行)

当前逻辑：
```typescript
org_name: properties.orgName ?? ''
```

需要验证：CopilotKit 是否正确转发 `properties` 中的 `orgName`

#### 3. 可选：在租户端数据工作区集成 AI 过滤

**参考实现**: `src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx` (第 794-798 行)

```typescript
{hasCopilot && (
  <FilterCopilotActions
    onSetFilter={(json) => { setWhereJsonCommitted(json) }}
    onClearFilter={handleClearFilter}
  />
)}
```

#### 4. 注意：CSS 样式导入

已在租户端项目 layout 中导入：
```typescript
import "@copilotkit/react-ui/styles.css"  // 第 11 行
```

如果需要在其他 layout 使用 CopilotSidebar，添加相同导入。

---

## 🔗 依赖链

```
User Action (点击 AI Assistant 按钮)
    ↓
ProjectLayout (src/app/org/.../project/.../layout.tsx)
    ↓ setShowCopilot(true)
    ↓
CopilotWrapper (src/.../CopilotProvider.tsx)
    ↓
CopilotKit (来自 @copilotkit/react-core)
    ↓ runtimeUrl="/api/copilotkit"
    ↓
BFF Handler (src/app/api/copilotkit/route.ts)
    ↓ 转发请求 + 注入授权
    ↓
AGENT_SERVICE_URL (默认 http://localhost:8000)
    ↓
Python modelcraft-agent (/copilotkit/ endpoint)
```

---

## 📊 文件映射表

| 文件路径 | 用途 | 当前状态 |
|---------|------|--------|
| `package.json` | CopilotKit 依赖 (v1.51.3) | ✅ |
| `next.config.mjs` | ESM transpile 配置 | ✅ |
| `src/app/api/copilotkit/route.ts` | BFF 代理 | ✅ |
| `src/web/components/features/copilot/CopilotProvider.tsx` | 核心 Provider | ✅ |
| `src/app/org/[orgName]/project/[projectSlug]/layout.tsx` | 租户项目 layout | ✅ |
| `src/app/end-user/[orgName]/projects/[projectSlug]/data/layout.tsx` | 用户数据 layout | ✅ |
| `src/web/components/features/end-user-data/FilterCopilotActions.tsx` | AI 过滤动作 | ✅ |
| `src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx` | 用户工作区 | ✅ |
| `src/web/components/features/end-user-data/FilterPanel.tsx` | 过滤栏 | ✅ |
| `src/app/org/[orgName]/layout.tsx` | 租户顶层 layout | ⚠️ |

---

## ⚠️ 已知问题 & 待做项

1. **orgName 转发验证**: 需要确认 CopilotProvider 的 `orgName` 是否被正确转发到 API

2. **CSS 样式冲突风险**: 如果在多个 layout 中导入 `@copilotkit/react-ui/styles.css` 会造成重复

3. **Agent 启动不可靠性**: EndUserCopilotWrapper 没有传 `agent=` 以避免启动时崩溃，导致运行时才发现问题

4. **租户端 AI 过滤未实现**: 只在终端用户端的 FilterPanel 集成了 AI 查询

---

## 总结

CopilotKit 采用**双层集成架构**，很好地平衡了性能和功能。基础设施已完整，后续主要工作是：
- ✅ 验证现有功能是否正常运作
- 🎯 可选：在其他租户路由复用同样的集成模式
- ⚠️ 需要解决的问题见上文

