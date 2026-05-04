# Developer / EndUser 双体系说明

> 适用范围：认证、网关代理、前端 BFF 联调。本文定义 ModelCraft 当前并行的两套用户体系。

## 1. 总览

ModelCraft 当前有两套并行体系：

1. **Developer 体系**：面向设计态/管理态用户。
2. **EndUser 体系**：面向业务终端用户。

两套体系在登录入口、Token 策略、GraphQL 路由和后端识别方式上都不同，但都必须通过 Gateway 进入 Backend。

---

## 2. 对照表

| 维度 | Developer 体系 | EndUser 体系 |
|---|---|---|
| 登录入口（Gateway） | `/auth/*` | `/api/end-user/auth/*` |
| 前端 BFF 入口 | `/api/auth/*` | `/api/bff/org/{orgName}/end-user/auth/*` |
| Access Token 验证 | ES256（网关使用 `JWT_PUBLIC_KEY` 验签） | HMAC-SHA256（网关使用 `JWT_SECRET` 验签） |
| Refresh Cookie | `mc_refresh_token` | `mc_enduser_refresh_token` |
| GraphQL 路由 | `/graphql/org/{orgName}`、`/graphql/org/{orgName}/project/{projectSlug}` | `/graphql/end-user/org/{orgName}/project/{projectSlug}` |
| 后端识别头 | `X-User-ID` | `X-User-ID` + `X-Internal-Token` |

---

## 3. 请求链路

### 3.1 Developer 链路

`Browser -> Front BFF(/api/auth/*) -> Gateway(/auth/*, /graphql/org/*) -> Backend`

### 3.2 EndUser 链路

`Browser -> Front BFF(/api/bff/org/*/end-user/*) -> Gateway(/api/end-user/auth/*, /graphql/end-user/*) -> Backend`

---

## 4. 强制边界（必须遵守）

1. 前端（浏览器侧 + 前端服务侧）**必须先访问 Gateway，再转发到 Backend**。
2. 禁止前端任何业务请求直连 Backend（包括 GraphQL/REST）。
3. Backend 对外联调视角只接受来自 Gateway 的受控流量。
4. **Gateway 是唯一的 Developer JWT 验签者**。Backend design-time 端点不接受 direct bearer token，只信任 Gateway 注入的 `X-User-ID`。
5. **CLI** 必须走 `cli -> gateway -> backend` 路径，不得直连 Backend design-time 端点。

---

## 5. 代码锚点（当前实现）

### Gateway
- 路由装配：`modelcraft-gateway/cmd/gateway/main.go`
- Developer/EndUser Token 校验：`modelcraft-gateway/internal/auth/service.go`
- GraphQL 代理：`modelcraft-gateway/internal/proxy/handler.go`

### Frontend BFF
- Developer auth 代理：`modelcraft-front/src/app/api/auth/[...path]/route.ts`
- EndUser auth 代理：`modelcraft-front/src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts`
- EndUser GraphQL 代理：`modelcraft-front/src/app/api/bff/graphql/end-user/org/[orgName]/project/[projectSlug]/route.ts`

### Backend
- EndUser 路由与 JWT 签发（HS256）：`modelcraft-backend/internal/interfaces/http/routes.go`
- EndUser HTTP Handler：`modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go`

---

## 6. 文档关系

- Gateway 架构细节：`./gateway-architecture.md`
- 部署联调检查项（含“前端必须经 Gateway”）：`../deployment/README.md`
