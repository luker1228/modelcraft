## Why

当前开发者认证体系处于半迁移状态：一部分逻辑已经转向新的 JWT 约定，一部分仍依赖 backend 侧的 legacy 校验，还有一部分行为散落在 front 代理层。目标运行拓扑现在已经明确：web 流量必须走 `front -> bff -> gateway -> backend`，CLI 流量必须走 `cli -> gateway -> backend`，并且 gateway 必须成为开发者认证的唯一边界。

现在推进这个变更的原因是，当前混合模型仍让 legacy `modelcraft` issuer 假设残留在 backend design API 内部，导致信任边界不清晰，也阻碍了向单一 `mc-developer` 身份体系的彻底迁移。

## What Changes

- 让 gateway 成为开发者认证的唯一验签者和兼容层。
- 让 backend design API 以 gateway-trusted resource server 的方式运行，而不是直接校验 developer JWT。
- 定义一套从 gateway 到 backend 的稳定 design-time trusted identity contract。
- 移除 backend design API 对浏览器或 CLI 直接传入 developer bearer token 校验的依赖。
- 让 frontend BFF 与 CLI 的访问路径统一到同一条 gateway-authenticated developer 流程。
- **BREAKING** 移除 backend design API 对 direct developer JWT verification 的 legacy 支持。
- **BREAKING** 在 gateway 迁移完成后，允许删除对 legacy `modelcraft` developer issuer 的兼容。

## Capabilities

### New Capabilities
- `developer-auth-gateway-routing`: 定义 front、BFF、gateway、CLI、backend 之间开发者认证与请求路由的标准模型。
- `backend-gateway-trusted-design-api`: 定义 backend design API 如何信任 gateway 注入的 developer identity，而不是直接校验 developer JWT。

### Modified Capabilities
无。

## Impact

- backend design-time GraphQL 与受保护 REST 端点的中间件
- gateway 上的 developer login / refresh / token validation 流程
- frontend 对 `/api/auth/*` 与 `/api/bff/graphql/*` 的 BFF 代理
- CLI 访问 gateway-backed design API 的路径
- JWT issuer 从 legacy `modelcraft` 向 canonical `mc-developer` 的迁移
