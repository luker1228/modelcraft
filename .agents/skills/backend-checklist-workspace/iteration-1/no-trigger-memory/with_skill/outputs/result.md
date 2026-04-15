# 回答：之前做了哪些改动，以及记住的重要信息

## backend-checklist skill 触发判断

**未触发 backend-checklist skill。**

理由：用户的问题是「你之前帮我做了哪些改动？有没有记住什么重要信息？」，这是一个**回顾历史/查询记忆**的请求，不符合 backend-checklist skill 的触发词（「加入错题本」、「记录这个错误」、「用错题本 check」、「checklist review」、「有没有历史 bug」），也不涉及错误录入或代码审查任务，因此无需触发该 skill。

---

## 之前帮你做过的改动（来自 git log）

根据 git 历史记录，近期主要提交包括：

| 提交 | 说明 |
|------|------|
| `2ab5ac3` | feat(front): 重构枚举字段创建流程并优化组件 |
| `9729231` | docs: 更新图谱报告 |
| `aa6323f` | feat(db): 在枚举关联表添加组织字段 |
| `9e2465d` | docs: 添加 BDD 测试指南并修复文档链接 |
| `62e9305` | feat(graphify): 添加知识图谱生成技能文档 |
| `eb1d113` | feat: 新增模型展示字段功能并优化组织路由跳转逻辑 |
| `af4c9a8` | feat(field): add display field label protocol end-to-end |
| `3cda330` | feat(project): add database connection test and step-by-step creation |
| `cb8a373` | feat(auth): 支持用户名登录及注册时创建个人组织 |
| `2b15d59` | fix(graphql): 处理字段验证错误 |
| `7ded48f` | refactor(modeldesign): 优化模型创建流程以支持惰性加载 |
| `5e7bcf2` | feat(api): implement phone-based authentication |

---

## 记住的重要信息（来自 memory 系统）

持久记忆中存有以下三条重要信息：

### 1. BDD 测试行为规范（feedback 类型）

- **不要自动修复后端代码** — `modelcraft-backend/` 完全禁止在 BDD 场景下修改。
- **默认只报告，不修复** — 除非用户明确要求 AND 失败分类为 `[TEST]`（测试代码本身的 bug）。
- **诊断优先用日志** — 失败时先从 requestId 追踪后端日志，再读代码。
- **不以注册为默认 setup** — BDD 测试应通过 login 或 `TEST_ACCESS_TOKEN` 获取认证状态。
- **Why:** 用户要求严格区分 BDD 诊断与修复职责，历史上曾发生过 mock/prod 不一致导致的问题。

### 2. GraphQL 字段选择与条件加载模式（project 类型）

- 使用 `graphql.CollectFields` + `graphql.GetOperationContext` 检测客户端请求的字段。
- 通过 `FieldSelectionChecker` 工具类实现 resolver 层的按需加载（仅在客户端请求时才查询数据库）。
- Dataloader 模式用于防止 N+1 问题：每请求懒初始化，用 `WHERE IN` 批量查询聚合外键。
- 关键文件：`internal/interfaces/graphql/project/field_selection.go`、`model.resolvers.go`、`internal/domain/modelruntime/relation_loader.go`。

### 3. 枚举 Order 字段架构（reference 类型）

- 枚举选项的 `order` 字段**不在组件状态中存储**，而是在提交 mutation 时**按数组下标自动赋值**。
- 前端枚举管理页：`modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/enums/page.tsx`
- 展示时按 `order` 字段排序（`.sort((a, b) => a.order - b.order)`）。
- `EnumOptionInput.order` 来源于 `options.map((opt, i) => ({ ..., order: i }))`。

---

## 总结

- **backend-checklist skill 未触发**：当前任务是查询历史改动和记忆内容，不匹配该 skill 的任何触发词。
- 持久记忆中共记录了 3 条重要信息，涵盖 BDD 测试规范、GraphQL 条件加载模式和枚举字段架构。
- git 历史显示近期工作聚焦于：枚举字段关联、展示字段协议、用户名认证、项目创建流程和 GraphQL 惰性加载优化。
