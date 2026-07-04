# 前端开发规范

本目录包含 ModelCraft 前端项目的开发规范和最佳实践文档。

## 📚 文档索引

### 架构

- **[前端架构总览](./architecture.md)** - 目录分层、组件约定、Hooks 组织、GraphQL 类型生成
  - 目录结构（`app/` / `web/` / `bff/` / `shared/` / `types/` / `generated/`）
  - `features/` vs `common/` 组件分类
  - 页面私有 `_components/` + `_hooks/` 拆分规范
  - BFF 路由
  - GraphQL Codegen 流程
  - Types / Hooks 按业务域组织约定
- **[API Client 层设计](./api-client-design.md)** - GraphQL 文档组织、Apollo 客户端、BFF 双体系路由约定
  - `src/api-client/**` 分层与门面导出
  - Developer 路由到 Gateway 的映射关系
  - 禁止前端直连 Backend 的约束
- **[Workspace 模式边界](./workspace-mode-boundary.md)** - 通过 `workspaceMode` 定义能力边界
  - 能力矩阵（插入列、字段生命周期管理）
  - 调用方强制显式传参约定

### 代码质量

- **[ESLint 规则](./eslint-rules.md)** - ESLint 配置和代码检查规则
  - Tailwind CSS 规则
  - 设计系统强制规则(字体、颜色)
  - 自动修复指南

- **[TypeScript 开发指南](./typescript-guide.md)** - TypeScript 配置和类型系统最佳实践
  - 严格模式配置
  - React 组件类型
  - Hooks 类型定义
  - GraphQL 类型集成

### 代码规范

- **[代码规范](./code-conventions.md)** - 代码风格和命名约定
  - 目录结构规范
  - 命名规范(变量、函数、组件、类型)
  - 代码风格(导入顺序、组件结构、注释)
  - 最佳实践

- **[React 最佳实践](./react-best-practices.md)** - React 开发模式和性能优化
  - 组件设计原则
  - State 管理
  - Hooks 使用
  - 性能优化(memo、useMemo、useCallback)
  - 错误处理

## 🚀 快速开始

### 1. 设置开发环境

```bash
# 安装依赖
npm install

# 运行开发服务器
npm run dev

# 运行 ESLint 检查
npm run lint

# 自动修复问题
npm run lint -- --fix
```

### 2. 编辑器配置

推荐使用 **VS Code** 并安装以下扩展:

- **ESLint** (`dbaeumer.vscode-eslint`) - 实时代码检查
- **Tailwind CSS IntelliSense** (`bradlc.vscode-tailwindcss`) - Tailwind 类名提示
- **Pretty TypeScript Errors** - 更友好的 TypeScript 错误提示

配置 `.vscode/settings.json`:

```json
{
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "eslint.validate": [
    "javascript",
    "javascriptreact",
    "typescript",
    "typescriptreact"
  ]
}
```

### 3. 核心规则速查

#### 字体权重

```tsx
// ❌ 禁止
<h1 className="font-bold">标题</h1>

// ✅ 允许
<h1 className="font-semibold">标题</h1>  // 600
<p className="font-medium">正文</p>     // 500
```

#### 文字颜色

```tsx
// ❌ 禁止:直接使用灰度值
<p className="text-gray-600">文字</p>

// ✅ 允许:使用语义化变量
<p className="text-foreground">主要文字</p>
<p className="text-muted-foreground">次要文字</p>
```

#### 组件命名

```typescript
// ✅ 组件:PascalCase
export function UserCard() { }

// ✅ 函数/变量:camelCase
const userName = 'John'
function fetchUserData() { }

// ✅ Hooks:use 前缀 + camelCase
function useUser() { }
```

#### 导入顺序

```typescript
// 1. React/Next.js
import React from 'react'
import { useRouter } from 'next/navigation'

// 2. 第三方库
import { useQuery } from '@apollo/client'

// 3. 内部模块(@ 别名)
import { Button } from '@/components/ui/button'

// 4. 相对导入
import { Header } from './Header'
```

## 📖 相关文档

### 设计系统

- [设计系统规范](../style/STYLE.md) - 整体设计规范
- [颜色系统](../style/color-system.md) - 颜色使用指南
- [快速开始指南](../style/quick-start.md) - 设计系统快速上手
- [Tailwind 使用规范](../style/tailwind-usage-policy.md) - Tailwind CSS 策略

### 项目文档

- [项目根目录 AGENTS.md](../../AGENTS.md) - 项目整体说明
- [后端 AGENTS.md](../../modelcraft-backend/AGENTS.md) - 后端详细文档
- [前端 AGENTS.md](../../modelcraft-front/AGENTS.md) - 前端详细文档
- [API Contract 共享规范](../../ai-metadata/backend/development/contract-sync.md) - API Contract 同步机制

### API Contract

前端 `contract/` 目录从后端同步，**禁止直接修改**。

```bash
# 拉取后端最新的 API Contract
# 使用 front-contract-pull skill
front-contract-pull
```

> 详细说明见 [front-contract-pull skill](../../.agents/skills/front-contract-pull/SKILL.md)

## 🔧 工具和命令

### ESLint

```bash
# 检查所有文件
npm run lint

# 自动修复可修复的问题
npm run lint -- --fix

# 检查特定文件
npx eslint src/components/Button.tsx

# 查看配置
npx eslint --print-config src/app/page.tsx
```

### TypeScript

```bash
# 类型检查
npx tsc --noEmit

# 查看类型推断
npx tsc --explainFiles
```

## ⚠️ 常见错误

### 1. 使用了禁止的字体权重

```
Error: 禁止使用 font-bold/font-extrabold/font-black。
请使用 font-semibold (600) 或 font-medium (500)。
参见 src/lib/typography.ts。
```

**修复**: 将 `font-bold` 改为 `font-semibold`

### 2. 使用了非语义化颜色

```
Error: 禁止使用 text-gray-* 具体值。
主文本用 text-foreground，次要文本用 text-muted-foreground。
参见 STYLE.md §1.3。
```

**修复**: 将 `text-gray-600` 改为 `text-muted-foreground`

### 3. Tailwind 类名冲突

```
Error: tailwindcss/no-contradicting-classname
```

**修复**: 检查是否同时使用了冲突的类名(如 `flex` 和 `block`)

## 💡 最佳实践提醒

1. **必须使用 shadcn/ui** - 所有基础 UI 组件使用 `@/components/ui`
2. **组件职责单一** - 一个组件只做一件事
3. **提取可复用逻辑** - 使用自定义 Hooks
4. **明确的类型定义** - 避免使用 `any`
5. **性能优化** - 适当使用 `memo`、`useMemo`、`useCallback`
6. **错误处理** - 使用 Error Boundary 和 try-catch
7. **代码复审** - 提交前运行 `npm run lint`

## 🤝 贡献

遵循以下流程提交代码:

1. 创建功能分支
2. 遵循代码规范开发
3. 运行 `npm run lint` 确保无错误
4. 提交代码(提交信息遵循约定式提交)
5. 创建 Pull Request

## 📝 更新日志

- **2025-03-23**: 创建开发规范文档
  - 添加 ESLint 规则文档
  - 添加 TypeScript 开发指南
  - 添加代码规范文档
  - 添加 React 最佳实践文档

---

如有问题或建议,请联系前端团队。
