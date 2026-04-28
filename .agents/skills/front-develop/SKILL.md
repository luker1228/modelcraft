---
name: front-develop
description: >
  ModelCraft 前端业务开发总指南。仅在"前端业务代码实现/修改"时触发：
  (1) 新增或修改页面、组件、Hooks、API Client 等前端业务逻辑；
  (2) 需要理解前端分层架构（app/ bff/ web/ shared/ generated/）；
  (3) GraphQL 查询/变更文档编写或调整（需对照 contract/graph/ 校验）；
  (4) 涉及 shadcn/ui 组件使用、Tailwind 样式、设计系统规范；
  (5) 涉及 MSW mock 开发策略或页面级 mock 控制；
  (6) 涉及 TypeScript 类型、ESLint 规则、代码规范问题。
  不要用于：纯调试排错（用 backend-debug）、纯后端任务、数据库迁移、部署运维、Contract 同步（用 front-contract-pull）。
---

# ModelCraft 前端开发指南

> 本 skill 是前端开发的**导航层**——先读简介定位，再按需深入对应文档。

---

## 快速定位

遇到前端任务，先判断属于哪个类别，直接跳转对应文档：

| 我要做的事 | 优先读 |
|-----------|--------|
| 理解目录分层、路由结构、依赖规则 | @ai-metadata/front/development/architecture.md |
| 写 GraphQL 查询/变更，理解 API Client 层 | @ai-metadata/front/development/api-client-design.md |
| 组件设计、State 管理、Hooks、性能优化 | @ai-metadata/front/development/react-best-practices.md |
| 命名规范、导入顺序、文件结构 | @ai-metadata/front/development/code-conventions.md |
| TypeScript 类型写法、严格模式 | @ai-metadata/front/development/typescript-guide.md |
| ESLint 规则、字体/颜色限制、自动修复 | @ai-metadata/front/development/eslint-rules.md |
| UI 组件规范、设计 token、禁止事项 | @ai-metadata/front/style/ui-spec.md |
| Design / End User 模式的能力边界 | @ai-metadata/front/development/workspace-mode-boundary.md |
| 已知 Bug、临时解决方案 | @ai-metadata/front/development/known-issues.md |

---

## 架构速览

```
src/
├── app/          # Next.js 路由层（页面入口 + layout）
├── bff/          # BFF 层（服务端数据聚合，靠近 HTTP）
├── web/          # 纯前端 UI 层（组件、Hooks、页面逻辑）
├── shared/       # app/bff/web 三层共用工具
├── api-client/   # GraphQL 文档 + Apollo Client + Auth
└── generated/    # codegen 自动生成，禁止手动编辑
```

**依赖方向（单向）：** `app → web → shared`，`bff` 独立，`api-client` 被 web/bff 调用。

> 详细目录结构、路径别名、认证流程见 @ai-metadata/front/development/architecture.md

---

## GraphQL 开发关键规则

1. **以 `contract/graph/` 为唯一参照物** — 编写查询前先查 Schema，不得凭记忆推测字段
2. **查询文件位置** — `src/api-client/{module}/graphql-docs.ts`
3. **验证方式** — 写完后运行 `npm run codegen`，报 `Cannot query field` 即字段非法
4. **生成代码只读** — `src/generated/` 和 `contract/` 禁止手动编辑

### Contract 同步

`contract/` 目录从后端单向同步，**禁止直接修改**。当后端修改了 Schema（新增字段、重命名类型、修改 mutation）后，需先同步 contract 再开发：

```
后端改 Schema → front-contract-pull → npm run codegen → 修复查询 → 开发
```

**何时需要同步：**
- 后端告知 Schema 有变更
- codegen 报错提示类型缺失或字段不存在
- 想用一个字段但在 `contract/graph/` 里找不到

**同步方式：** 使用 `front-contract-pull` skill（不要手动 cp 文件）。

> 完整规范（端点约定、门面模式、MSW mock）见 @ai-metadata/front/development/api-client-design.md

---

## 设计系统关键规则

```tsx
// ❌ 禁止
<h1 className="font-bold text-gray-600">标题</h1>

// ✅ 正确
<h1 className="font-semibold text-foreground">标题</h1>
```

- **字体**：只用 `font-semibold`（600）或 `font-medium`（500），禁止 `font-bold`
- **颜色**：用语义变量（`text-foreground`、`text-muted-foreground`），禁止 `text-gray-*`
- **组件**：所有基础 UI 必须用 `@/components/ui`（shadcn/ui），禁止自造轮子

> 完整 token、组件规范、Anti-Pattern 见 @ai-metadata/front/style/ui-spec.md  
> 提交前速查 → @ai-metadata/front/style/ui-checklist.md

---

## 开发命令速查

```bash
npm run dev          # 启动开发服务器
npm run codegen      # 重新生成 GraphQL TypeScript 类型
npm run lint         # ESLint 检查
npm run lint -- --fix  # 自动修复
npx tsc --noEmit     # TypeScript 类型检查
```

---

## 常见错误速查

| 错误 | 原因 | 修复 |
|------|------|------|
| `Cannot query field "xxx" on type "YYY"` | 查询了 Schema 中不存在的字段 | 查 `contract/graph/` 确认字段，删除非法字段后重跑 codegen |
| `Too many re-renders` | Apollo Client 实例在 render 中新建 | 用 `useMemo` 稳定 client 实例 |
| ESLint `font-bold` 报错 | 禁止使用 font-bold | 改为 `font-semibold` |
| ESLint `text-gray-*` 报错 | 禁止硬编码灰度颜色 | 改为 `text-muted-foreground` 等语义变量 |

> 更多已知问题见 @ai-metadata/front/development/known-issues.md
