# Developer / EndUser 双体系说明

> 适用范围：认证、网关代理、前端 BFF 联调。本文定义 ModelCraft 当前并行的两套用户体系。

## 1. 总览

ModelCraft 有两类用户，但共享同一套 JWT 格式（统一 issuer `mc-platform`，`scope` claim 区分权限级别）：

1. **Developer 体系（平台管理者）**：负责 Org 管理、Project 创建、模型设计，拥有全量权限。
2. **EndUser 体系（端用户）**：访问被授权的 Project 内的业务数据，不感知平台管理细节。

两套体系使用**完全独立的前端入口**，通过路由前缀实现物理隔离。

---

## 2. 用户能力边界

| 能力 | Developer（平台管理者） | EndUser（端用户） |
|------|------------------------|------------------|
| 创建/删除 Org | ✅ | ❌ |
| 邀请注册 / 管理账号 | ✅ | ❌ |
| 创建 / 删除 Project | ✅ | ❌ |
| 设计模型 / 字段 / 数据库结构 | ✅ | ❌ |
| **List 可访问的 Project** | ✅ | ✅ |
| 在 Project 内操作业务数据（CRUD） | ❌（他是设计者） | ✅ |
| 查看 / 管理自己的 API Key | ❌ | ✅ |
| 查看自己的 Profile | ❌ | ✅ |

---

## 3. 入口隔离

两套体系使用完全独立的前端入口，**URL 本身承载了 Org 上下文**：

| 体系 | 前端入口 | 登录后路由 |
|------|---------|-----------|
| Developer | `app.example.com/` | `/org/{orgName}/dashboard` |
| EndUser | `app.example.com/end-user/{orgName}/login` | `/end-user/{orgName}/projects` |

**EndUser 不需要在登录表单里填写 Org 名称**——Org 已经编码在 URL 里。用户访问平台分配给他的链接，orgName 已固定。访问错误 Org 的链接直接返回"用户不存在"。

---

## 4. 对照表

| 维度 | Developer 体系 | EndUser 体系 |
|------|---------------|--------------|
| **前端入口** | `app.example.com/` | `app.example.com/end-user/{orgName}/login` |
| **前端路由前缀** | `/org/{orgName}/...` | `/end-user/{orgName}/...` |
| **登录入口（Gateway）** | `/auth/login` | `/end-user/{orgSlug}/auth/login` |
| **前端 BFF 入口** | `/api/auth/*` | `/api/bff/org/{orgName}/end-user/auth/*` |
| **JWT issuer** | `mc-platform` | `mc-platform` |
| **JWT scope（登录后）** | `scope=org` | `scope=org` |
| **JWT scope（进入 Project 后）** | `scope=project`（exchange 换取） | `scope=project`（exchange 换取） |
| **Access Token 签名** | ES256（ECDSA P-256） | ES256（ECDSA P-256） |
| **Org 级 GraphQL 路由** | `/graphql/org/{orgName}/` | `/graphql/end-user/org/{orgName}/` |
| **Project 级 GraphQL 路由** | `/graphql/org/{orgName}/project/{projectSlug}/` | `/graphql/end-user/org/{orgName}/project/{projectSlug}/` |
| **后端识别头** | `X-User-ID` + `X-Token-Scope` | `X-User-ID` + `X-Token-Scope` |

---

## 5. GraphQL Schema 分层

EndUser 拥有 Org 级 + Project 级两层端点，对应两套 Schema：

```
Developer Org Schema（全量）            EndUser Org Schema（子集）
/graphql/org/{orgName}/                /graphql/end-user/org/{orgName}/
  createProject                          listProject（可访问的）
  deleteProject                          me（个人 Profile）
  manageUsers                            listApiKey / createApiKey
  ...                                    ...

Developer Project Schema                EndUser Project Schema
/graphql/org/{orgName}/project/{slug}/  /graphql/end-user/org/{orgName}/project/{slug}/
  模型设计 / 字段管理                      业务数据 CRUD（query / get / create / update / delete）
  数据库结构                               Runtime GraphQL（动态 Schema）
  ...
```

EndUser 的 Org Schema 是 Developer Org Schema 的子集，**不包含任何管理操作**（创建/删除 Project、用户管理等）。

---

## 6. 请求链路

### 6.1 Developer 链路

```
Browser -> Front BFF(/api/auth/*) -> Gateway(/auth/*, /graphql/org/*) -> Backend
```

### 6.2 EndUser 链路

```
Browser -> Front BFF(/api/bff/org/*/end-user/*) 
        -> Gateway(/end-user/{orgSlug}/auth/*, /graphql/end-user/*) 
        -> Backend
```

### 6.3 Token exchange 流程（两体系共用）

登录后持有 `scope=org` Token，访问具体 Project 前需调用 exchange 换取 `scope=project` Token：

```
POST /api/auth/exchange
  { projectSlug: "sales" }
  Authorization: Bearer <scope=org token>

→ 200 { accessToken: <scope=project token> }
  Project Token 以 httpOnly cookie 存储，BFF 服务端完成 exchange，JS 不可读
```

---

## 7. 强制边界（必须遵守）

1. 前端（浏览器侧 + 前端服务侧）**必须先访问 Gateway，再转发到 Backend**。
2. 禁止前端任何业务请求直连 Backend（包括 GraphQL/REST）。
3. **Gateway 是唯一的 JWT 验签者**，使用 ES256（`JWT_PUBLIC_KEY`）验签 `mc-platform` issuer。
4. Backend design-time 端点只信任 Gateway 注入的 `X-User-ID` + `X-Token-Scope`，不接受直接 bearer token。
5. **CLI** 走 `cli -> gateway -> backend` 路径，不得直连 Backend。
6. EndUser **不填写 Org 名称**——Org 由 URL 路径决定，或由 JWT claim 中的 `orgId` 确定。

---

## 8. 代码锚点（当前实现）

### Gateway
- 路由装配：`modelcraft-gateway/cmd/gateway/main.go`
- Token 校验（统一 mc-platform ES256）：`modelcraft-gateway/internal/auth/service.go`
- GraphQL 代理：`modelcraft-gateway/internal/proxy/handler.go`

### Frontend BFF
- Developer auth 代理：`modelcraft-front/src/app/api/auth/[...path]/route.ts`
- EndUser auth 代理：`modelcraft-front/src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts`
- EndUser GraphQL 代理：`modelcraft-front/src/app/api/bff/graphql/end-user/org/[orgName]/project/[projectSlug]/route.ts`

### Backend
- EndUser 路由与 JWT 签发（ES256）：`modelcraft-backend/internal/interfaces/http/routes.go`
- EndUser HTTP Handler：`modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go`

---

## 9. 文档关系

- Gateway 架构细节：`./gateway-architecture.md`
- 部署联调检查项：`../deployment/README.md`
- EndUser v2 PRD：`../../prd/enduser-v2/`
- Token 统一里程碑路线图：`.planning/ROADMAP.md`
