## 1. Gateway 信任契约

- [x] 1.1 盘点当前 front、BFF、gateway、backend、CLI 之间的 developer request paths。
- [x] 1.2 定义 gateway 必须注入到 backend design-time request 中的 trusted developer identity contract。
- [x] 1.3 识别所有当前依赖 direct developer JWT validation 的 backend middleware 与 handler entry point。

## 2. Backend Design API 迁移

- [x] 2.1 更新 backend design-time middleware，使受保护请求通过 gateway-trusted identity contract 完成认证。
- [x] 2.2 移除 backend design GraphQL endpoint 对 direct developer bearer token verification 的依赖。
- [x] 2.3 移除受保护 backend design-time REST endpoint 对 direct developer bearer token verification 的依赖。
- [x] 2.4 新增或更新 backend tests，覆盖 gateway-trusted request 被接受、direct-backend request 被拒绝的场景。

## 3. Frontend 与 CLI 对齐

- [x] 3.1 验证 frontend `/api/auth/*` 与 `/api/bff/graphql/*` 流程符合 gateway-authenticated developer path。
- [x] 3.2 识别并移除 frontend 中假设 backend design API 会直接认证 developer bearer token 的逻辑。
- [x] 3.3 盘点 CLI 的 design-time request，并将其统一到 `cli -> gateway -> backend`。
- [x] 3.4 新增或更新 frontend 与 CLI 在 gateway-first model 下的测试与 fixtures。

## 4. Legacy 清理

- [x] 4.1 移除 backend 中仅为 direct JWT verification 保留的 developer-auth compatibility path。
- [x] 4.2 在确认 gateway compatibility 完成后，移除 backend 中与 direct developer JWT validation 绑定的 legacy issuer 假设。
- [x] 4.3 更新配置、注释与文档，明确 gateway 是唯一的 developer authentication authority。
