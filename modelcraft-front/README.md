# ModelCraft React

一个基于 Next.js + TypeScript 构建的现代化模型管理平台，提供直观的用户界面来管理模型、集群、枚举等资源。

## ✨ 特性

- 🚀 **现代技术栈**: Next.js 14 + React 18 + TypeScript
- 🎨 **精美UI**: 基于 Tailwind CSS + Radix UI 组件
- 📊 **数据管理**: 集成 Apollo Client (GraphQL) + TanStack Query
- 🔄 **状态管理**: Zustand 轻量级状态管理
- 📝 **表单处理**: React Hook Form + Zod 验证
- 🛣️ **路由管理**: Next.js App Router 文件系统路由
- 🎯 **TypeScript**: 完整的类型安全支持
- ⚡ **性能优化**: 服务端渲染 (SSR) + 静态生成 (SSG)

## 🏗️ 项目结构

```
src/
├── app/                    # Next.js App Router 页面
│   ├── layout.tsx         # 根布局
│   ├── page.tsx           # 首页（重定向）
│   ├── welcome/           # 欢迎页
│   ├── projects/          # 项目管理页
│   └── (main)/            # 主应用路由组
│       ├── layout.tsx     # 主布局（侧边栏+头部）
│       ├── dashboard/     # 仪表板
│       ├── clusters/      # 集群管理
│       ├── models/        # 模型管理
│       ├── enums/         # 枚举管理
│       ├── guide/         # 新手引导
│       └── accounting/    # 记账系统
├── components/            # 可复用组件
│   ├── layout/           # 布局组件
│   ├── project/          # 项目相关组件
│   └── ui/               # 基础UI组件
├── graphql/              # GraphQL 查询和变更
│   ├── queries/          # 查询操作
│   └── mutations/        # 变更操作
├── hooks/                # 自定义 React Hooks
├── lib/                  # 工具库和配置
│   ├── apollo-wrapper.tsx # Apollo Client 配置
│   ├── query-wrapper.tsx  # React Query 配置
│   └── utils.ts          # 工具函数
├── stores/               # Zustand 状态存储
├── styles/               # 样式文件
└── types/                # TypeScript 类型定义
```

## 🚀 快速开始

### 环境要求

- Node.js >= 18.0.0
- npm 或 yarn 或 pnpm

### 安装依赖

```bash
npm install
```

### 开发环境

```bash
npm run dev
```

应用将在 `http://localhost:3000` 启动

### 同步 Tailwind 配置

保持原型和 React 项目配置同步：

```bash
npm run sync-tailwind
```

详见：[Tailwind 配置同步指南](docs/TAILWIND_SYNC_GUIDE.md)

### 构建生产版本

```bash
npm run build
```

### 启动生产服务器

```bash
npm run start
```

### 代码检查

```bash
npm run lint
```

## 📋 主要功能

### 🏠 仪表板
- 系统概览和统计信息
- 快速导航到各个功能模块

### 📦 模型管理
- 创建、编辑、删除模型
- 模型字段配置
- 模型关系管理

### 🖥️ 集群管理
- 集群资源监控
- 集群配置管理
- 数据库连接测试

### 🏷️ 枚举管理
- 枚举类型定义
- 枚举值管理

### 🧭 其他功能
- **新手引导**: 快速入门教程
- **记账系统**: 财务管理演示

## 🛠️ 技术栈详情

### 核心框架
- **Next.js 14**: React 框架，支持 App Router、SSR、SSG
- **React 18**: 最新的 React 版本，支持并发特性
- **TypeScript**: 提供类型安全和更好的开发体验

### UI 和样式
- **Tailwind CSS**: 实用优先的 CSS 框架
- **Radix UI**: 无样式、可访问的组件库
- **Lucide React**: 精美的图标库
- **tailwindcss-animate**: CSS 动画支持

### 数据管理
- **Apollo Client**: GraphQL 客户端，支持缓存和状态管理
- **TanStack Query**: 强大的数据获取和缓存库
- **Zustand**: 轻量级状态管理解决方案

### 表单和验证
- **React Hook Form**: 高性能的表单库
- **Zod**: TypeScript 优先的模式验证
- **@hookform/resolvers**: 表单验证解析器

## 🔧 开发指南

### HTML-First Prototype 工作流

ModelCraft 使用 HTML 原型优先的开发流程：

1. **创建原型**: 在 `prototypes/` 目录下创建 HTML 原型
2. **确认设计**: 在浏览器中预览和调整样式
3. **实现组件**: 将原型中的 Tailwind 类复制到 React 组件

详见：[HTML-First Prototype 工作流](.codebuddy/skills/html-first-prototype/SKILL.md)

### 添加新页面

1. 在 `src/app/` 目录下创建新的页面目录
2. 在目录中添加 `page.tsx` 文件
3. 如需使用主布局，放在 `src/app/(main)/` 目录下
4. 如需要，在侧边栏组件中添加导航链接

### 创建新组件

1. 在相应的 `src/components/` 子目录中创建组件
2. 使用 TypeScript 定义 props 接口
3. 客户端组件需要在文件顶部添加 `'use client'`
4. 遵循项目的命名约定

### 状态管理

使用 Zustand 进行全局状态管理：

```typescript
// 在 stores/ 中定义 store
export const useMyStore = create<MyState>((set) => ({
  // state and actions
}))

// 在组件中使用
const { data, actions } = useMyStore()
```

### GraphQL 集成

1. 在 `src/graphql/` 中定义查询和变更
2. 使用 Apollo Client hooks 进行数据获取
3. 结合 TanStack Query 进行额外的缓存控制

### API 代理配置

在 `next.config.mjs` 中配置 rewrites 规则：

```javascript
async rewrites() {
  return [
    {
      source: '/api/copilotkit/:path*',
      destination: 'http://localhost:8001/api/copilotkit/:path*',
    },
    {
      source: '/api/design/:path*',
      destination: 'http://localhost:8080/api/design/:path*',
    },
  ]
}
```

## 📝 代码规范

- 使用 ESLint 进行代码检查
- 遵循 TypeScript 严格模式
- 组件使用 PascalCase 命名
- 文件和目录使用 kebab-case 或 camelCase
- 优先使用函数组件和 Hooks
- 客户端组件标记 `'use client'`

### 设计系统规范

- **实色为主** - 禁止渐变、毛玻璃效果
- **边框优先** - 使用边框而非阴影定义边界
- **B2B 风格** - 专业、简洁、克制

详见：[ModelCraft 设计系统规范](ai-metadata/style/STYLE.md)

## 📚 文档

- [完整文档索引](docs/INDEX.md) - 所有文档的入口
- [Tailwind 配置同步](docs/TAILWIND_SYNC_GUIDE.md) - 配置同步完整指南
- [快速同步参考](docs/QUICK_SYNC_REFERENCE.md) - 一行命令快速参考
- [HTML-First Prototype](.codebuddy/skills/html-first-prototype/SKILL.md) - 原型优先工作流
- [设计系统规范](ai-metadata/style/STYLE.md) - 完整设计系统

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关链接

- [Next.js 文档](https://nextjs.org/docs)
- [React 文档](https://react.dev/)
- [TypeScript 文档](https://www.typescriptlang.org/)
- [Tailwind CSS 文档](https://tailwindcss.com/)
- [Apollo Client 文档](https://www.apollographql.com/docs/react/)
