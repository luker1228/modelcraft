## 1. Schema 与错误模型收敛

- [x] 1.1 在 Org/Project/EndUser Schema 中新增 `ResourceType` 枚举与 `ResourceNotFound` 类型
- [x] 1.2 将各 Query/Mutation 错误 union 中的 `*NotFound` 成员替换为 `ResourceNotFound`
- [x] 1.3 清理不再使用的 `*NotFound` 类型定义并确保 schema 可通过校验

## 2. 后端错误映射改造

- [x] 2.1 修改各 GraphQL 错误适配器，将 `NOT_FOUND.*` 统一映射到 `generated.ResourceNotFound`
- [ ] 2.2 建立业务错误码到 `ResourceType` 的映射函数，并补充 `UNKNOWN` 兜底策略
- [ ] 2.3 保持 `extensions.code` 细粒度透传，补充必要单元测试覆盖映射分支

## 3. 前端与 BDD 迁移

- [x] 3.1 批量更新前端 `graphql-docs.ts`：将 `... on XxxNotFound` 改为 `... on ResourceNotFound { message resourceType }`
- [ ] 3.2 运行前端 codegen 并修复类型报错，统一运行时分支判断到 `ResourceNotFound + resourceType`
- [x] 3.3 更新 BDD feature 与 step-definitions 的错误断言，移除对具体 `XxxNotFound` 的依赖

## 4. 回归验证与收口

- [x] 4.1 执行后端 GraphQL 代码生成并完成编译检查
- [ ] 4.2 执行前端 lint/类型检查，确认 GraphQL 客户端与 UI 逻辑通过
- [ ] 4.3 执行 BDD 回归，确认 not-found 场景全部通过并记录破坏性变更说明

## 验证阻塞说明（2026-05-03）

- 前端 `npm run codegen` 失败：`contract/graph/*` 仍未包含 `ResourceNotFound` 与多条 RBAC 新能力定义，导致文档校验报错。
- 前端 `npm run lint` 失败：当前分支存在大量与本次变更无关的既有 lint error（RBAC 页面类型安全问题）。
- BDD 运行失败：测试环境存在鉴权与 schema 不匹配（含 `Unknown type "ResourceNotFound"`、`401 Unauthorized`、既有 undefined steps）。