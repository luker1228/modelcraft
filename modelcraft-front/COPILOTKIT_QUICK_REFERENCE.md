# CopilotKit 快速参考指南

## 🎯 快速问题与答案

### Q1: CopilotKit 目前在哪个 layout 被引入？
**A**: 
- ✅ **End-User 端**: `src/app/end-user/[orgName]/projects/[projectSlug]/data/layout.tsx` 使用 `EndUserCopilotWrapper`
- ✅ **Tenant 端**: `src/app/org/[orgName]/project/[projectSlug]/layout.tsx` 使用 `CopilotWrapper`（按需加载）

### Q2: tenant-admin 的顶层 layout 文件路径是什么？
**A**: `src/app/org/[orgName]/layout.tsx`
- 当前只有 `OnboardingProvider` 包装，**没有** CopilotKit 集成

### Q3: 要在 tenant-admin 引入 CopilotKit 需要改哪些文件？
**A**: 
1. **已完成** ✅
   - `src/app/org/[orgName]/project/[projectSlug]/layout.tsx` - 已集成

2. **可选** 🎯
   - `src/app/org/[orgName]/settings/layout.tsx` - 参考项目 layout 的模式
   - `src/app/org/[orgName]/developers/layout.tsx` - 参考项目 layout 的模式

3. **需验证** ⚠️
   - `src/app/api/copilotkit/route.ts` - 检查 `orgName` 是否被正确转发

---

## 📁 关键文件一览

### 核心文件（不要改！）

| 文件 | 作用 |
|------|------|
| `src/app/api/copilotkit/route.ts` | BFF 代理，转发请求到 Python Agent |
| `src/web/components/features/copilot/CopilotProvider.tsx` | Provider 组件，导出 `CopilotWrapper` 和 `EndUserCopilotWrapper` |
| `package.json` | CopilotKit 依赖 (^1.51.3) |
| `next.config.mjs` | ESM transpile 配置 |

### 集成文件（可参考）

| 文件 | 用途 |
|------|------|
| `src/app/org/[orgName]/project/[projectSlug]/layout.tsx` | 租户项目 layout（参考实现） |
| `src/app/end-user/[orgName]/projects/[projectSlug]/data/layout.tsx` | 用户数据 layout（参考实现） |
| `src/web/components/features/end-user-data/FilterCopilotActions.tsx` | AI 过滤动作（参考实现） |

---

## 🔧 复制粘贴模板

### 模板 1: 在新的 layout 中添加 CopilotKit 支持

```typescript
'use client'

import { useEffect, useState, useMemo } from 'react'
import { useParams } from 'next/navigation'
import { CopilotWrapper, AIAssistantButton } from '@web/components/features/copilot/CopilotProvider'
import "@copilotkit/react-ui/styles.css"

export default function YourLayout({ children }: { children: React.ReactNode }) {
  const params = useParams()
  const orgName = params.orgName as string
  const projectSlug = params.projectSlug as string
  
  // TODO: 获取 selectedProject，参考 ProjectLayout 的实现
  const selectedProject = null
  
  const [showCopilot, setShowCopilot] = useState(false)

  const mainContent = (
    <>
      {children}
      {!showCopilot && (
        <AIAssistantButton onClick={() => setShowCopilot(true)} />
      )}
    </>
  )

  if (showCopilot) {
    return (
      <CopilotWrapper selectedProject={selectedProject} orgName={orgName}>
        {mainContent}
      </CopilotWrapper>
    )
  }

  return mainContent
}
```

### 模板 2: 在组件中挂载 CopilotKit 动作

```typescript
'use client'

import { useCopilotKitAvailable, FilterCopilotActions } from './FilterCopilotActions'

export function YourComponent() {
  const hasCopilot = useCopilotKitAvailable()
  
  const handleApplyFilter = (filterJson: string) => {
    // 处理过滤逻辑
  }

  const handleClearFilter = () => {
    // 清除过滤逻辑
  }

  return (
    <>
      {/* 你的组件内容 */}
      
      {/* 仅在 CopilotKit 可用时挂载 AI 动作 */}
      {hasCopilot && (
        <FilterCopilotActions
          onSetFilter={handleApplyFilter}
          onClearFilter={handleClearFilter}
        />
      )}
    </>
  )
}
```

### 模板 3: 创建自定义 CopilotKit 动作

参考: `src/web/components/features/end-user-data/FilterCopilotActions.tsx`

```typescript
import { useCopilotAction } from '@copilotkit/react-core'

export function YourCustomActions() {
  useCopilotAction({
    name: 'your_action_name',
    description: '你的动作描述',
    parameters: [
      {
        name: 'param_name',
        type: 'string',
        description: '参数描述',
        required: true,
      },
    ],
    handler: async ({ param_name }: { param_name: string }) => {
      // 处理逻辑
    },
  })
  
  return null  // 不渲染任何 UI
}
```

---

## ⚠️ 常见坑

### 坑 1: 使用 CopilotKit hooks 但忘记 Provider

❌ **错误**:
```typescript
export function MyComponent() {
  // 这会在没有 CopilotWrapper 时崩溃！
  useCopilotAction({ ... })
  return <div>...</div>
}
```

✅ **正确**:
```typescript
export function MyComponent() {
  const hasCopilot = useCopilotKitAvailable()
  
  return (
    <>
      {hasCopilot && <MyActions />}
    </>
  )
}

function MyActions() {
  useCopilotAction({ ... })
  return null
}
```

### 坑 2: 在多个 layout 导入 CSS

❌ **错误**（会造成重复导入）:
```typescript
// layout1.tsx
import "@copilotkit/react-ui/styles.css"

// layout2.tsx
import "@copilotkit/react-ui/styles.css"
```

✅ **正确**（只在需要 UI 的 layout 导入）:
```typescript
// src/app/org/[orgName]/project/[projectSlug]/layout.tsx
import "@copilotkit/react-ui/styles.css"
```

### 坑 3: useContext(CopilotContext) 不可靠

❌ **错误**:
```typescript
const ctx = useContext(CopilotContext)
if (ctx !== null) {  // 这总是 true！
  useCopilotAction({ ... })
}
```

✅ **正确**:
```typescript
const hasCopilot = useCopilotKitAvailable()
if (hasCopilot) {  // 这才是正确的检测方式
  useCopilotAction({ ... })
}
```

---

## 🚀 测试 CopilotKit 集成

### 步骤 1: 验证环境

```bash
# 检查 AGENT_SERVICE_URL 是否设置
echo $AGENT_SERVICE_URL

# 应该输出: http://localhost:8000（本地）或其他 agent 地址
```

### 步骤 2: 启动 Agent 服务

```bash
# 假设 Python agent 在 modelcraft-agent 项目
cd ../modelcraft-agent
python -m uvicorn main:app --port 8000
```

### 步骤 3: 测试租户端

1. 启动前端: `npm run dev`
2. 登录租户后台: `http://localhost:3000/org/your-org/project/your-project`
3. 点击右下角 "AI Assistant" 按钮
4. 检查浏览器控制台是否有错误

### 步骤 4: 测试终端用户端

1. 访问: `http://localhost:3000/end-user/your-org/projects/your-project/data`
2. 查看 FilterPanel 是否显示 "✨ AI 查询" 按钮
3. 尝试使用 AI 过滤功能

---

## 📋 检查清单

- [ ] 已读完全部集成报告
- [ ] 确认 `AGENT_SERVICE_URL` 环境变量已设置
- [ ] Python Agent 服务能正常启动
- [ ] 租户端 AI Assistant 按钮可以点击
- [ ] 终端用户端的 "✨ AI 查询" 按钮显示
- [ ] 浏览器控制台无 CopilotKit 相关错误
- [ ] 了解两套集成方案的区别
- [ ] 知道何时使用 `useCopilotKitAvailable()`

---

## 🔗 相关资源

- CopilotKit 官方文档: https://docs.copilotkit.ai/
- 项目中的完整报告: `COPILOTKIT_INTEGRATION_REPORT.md`
- 核心 Provider 组件: `src/web/components/features/copilot/CopilotProvider.tsx`
- API 代理实现: `src/app/api/copilotkit/route.ts`
- End-User 集成参考: `src/app/end-user/[orgName]/projects/[projectSlug]/data/layout.tsx`

