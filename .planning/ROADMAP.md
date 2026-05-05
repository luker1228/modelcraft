# 路线图：统一 Token 体系 & Workspace 入口重构

## 概览

本里程碑将 ModelCraft 的双轨认证体系（`mc-developer` + `mc-enduser`）统一为单一 `mc-platform` issuer，新增 `scope` claim 实现 Org/Project 两级 Token，彻底删除冗余的 `end_user` GraphQL schema，并为端用户提供完整的 Workspace 入口（登录 → 项目列表 → 进入 Workspace）。

---

## 阶段列表

**阶段编号说明：**
- 整数阶段（1、2、3）：里程碑规划工作
- 小数阶段（如 2.1）：紧急插入（标注 INSERTED）

- [ ] **阶段 1：Token 核心统一** — 统一 JWT issuer，新增 scope claim，实现 Org Token 登录路径（平台管理员 + 端用户）
- [ ] **阶段 2：中间件 & Exchange 端点** — scope 校验中间件 + `/api/auth/exchange` 换 Project Token
- [ ] **阶段 3：end_user Schema 清理** — 删除 `api/graph/end_user/` 及关联代码，核查存活 query 迁移情况
- [ ] **阶段 4：前端 Workspace 入口** — 新建端用户登录页、项目列表页、Project Token 换取、Workspace 跳转
- [ ] **阶段 5：测试 & 收口** — BDD 验收场景 + 集成测试，确认 scope 边界强执行

---

## 阶段详情

### 阶段 1：Token 核心统一
**目标**：后端统一 JWT issuer 为 `mc-platform`，`PlatformClaims` 新增 `scope` 字段，平台管理员登录和端用户登录均返回 `scope=org` 的 Token
**依赖**：无（起始阶段）
**需求**：TOKEN-01, TOKEN-02, TOKEN-03
**验收标准**（以下条件必须为真）：
  1. `POST /api/auth/login` 返回的 JWT，其 `iss` 为 `mc-platform`，`scope` 字段为 `"org"`
  2. `POST /api/end-user/{orgSlug}/auth/login` 返回的 JWT，`iss` 同样为 `mc-platform`，`scope` 为 `"org"`，不再出现 `mc-enduser`
  3. 旧 `mc-developer` / `mc-enduser` issuer 签发的 Token 被服务端拒绝（401）
  4. Token 使用 ES256（ECDSA P-256）签名，与现有 `jwt_signer.go` 保持一致
**计划**：3 个计划，3 个 wave
Plans:
- [ ] 01-PLAN-01.md — Wave 1：领域层 PlatformClaims 定义、issuer 常量迁移、jwt_signer 签名扩展
- [ ] 01-PLAN-02.md — Wave 2：应用层（token_service Refresh/Login）+ 接口层（enduser auth 迁移 ES256、middleware、rls）
- [ ] 01-PLAN-03.md — Wave 3：Gateway 端用户 token 验证从 HMAC 切换为 ES256

---

### 阶段 2：中间件 & Exchange 端点
**目标**：实现 scope 强制校验中间件，新增 `/api/auth/exchange` 端点，让 Org Token 可换取 Project Token（`scope=project`）
**依赖**：阶段 1
**需求**：TOKEN-04, TOKEN-05
**验收标准**（以下条件必须为真）：
  1. `POST /api/auth/exchange`（携带有效 Org Token）返回 `scope=project` 的新 Token，TTL 1h，且 token payload 不含 `projectSlug`
  2. 携带 `scope=org` 的 Token 访问 `/graphql/org/{orgName}/project/*` 返回 403
  3. 携带 `scope=project` 的 Token 访问 org 管理路由（如项目 CRUD）返回 403
  4. 携带有效 `scope=project` Token 可正常访问 Runtime GraphQL 端点
**计划**：2 个计划，2 个 wave
Plans:
- [ ] 02-PLAN-01.md — Wave 1：Gateway Claims 扩展 + X-Token-Scope 注入 + backend ParsePlatformClaims
- [ ] 02-PLAN-02.md — Wave 2：RequireScope 中间件 + ExchangeToken app 层 + handler/server/路由注册 + gateway 透传

---

### 阶段 3：end_user Schema 清理
**目标**：彻底删除 `api/graph/end_user/` 目录及所有关联 handler / resolver / 路由注册，核查并清理原 end_user schema 中的 query
**依赖**：阶段 1、阶段 2
**需求**：SCHEMA-01, SCHEMA-02
**验收标准**（以下条件必须为真）：
  1. `api/graph/end_user/` 目录在代码库中不存在
  2. 所有原 end_user 路由（如 `/graphql/end-user/*`）均已从路由注册中移除，访问返回 404
  3. 原 end_user schema 中 6 条 query 均已确认：要么对应到 org/project schema 的现有接口，要么已明确删除（无残留死代码）
  4. `just generate-gql` 和 `just build` 通过，无编译错误
**计划**：3 个计划，3 个 wave
Plans:
- [ ] 03-PLAN-01.md — Wave 1：删除 Backend routes.go/chi_setup.go 中的 enduser GraphQL 调用和 import
- [ ] 03-PLAN-02.md — Wave 2：删除 enduser 包目录、schema 目录、gqlgen 配置，更新 justfile
- [ ] 03-PLAN-03.md — Wave 3：清理 Gateway 路由注册、EndUserGraphQLHandler 及 Deprecated auth 代码

---

### 阶段 4：前端 Workspace 入口
**目标**：新建端用户完整登录-项目选择-Workspace 访问链路；BFF 服务端完成 exchange，Project Token 以 httpOnly cookie 存储
**依赖**：阶段 2（exchange 端点就绪）
**需求**：WORKSPACE-01, WORKSPACE-02, WORKSPACE-03, WORKSPACE-04
**验收标准**（以下条件必须为真）：
  1. 访问 `/end-user/{orgSlug}/login` 可看到端用户登录表单，输入凭证后登录成功
  2. 登录成功后跳转到项目列表页，仅展示当前用户有 RBAC membership 的项目
  3. 用户点击项目后，BFF 服务端静默调用 `/api/auth/exchange`，Project Token 以 httpOnly cookie 存储（浏览器 JS 无法读取）
  4. 成功跳转至 `/workspace/{orgSlug}/{projectSlug}`，页面仅展示 runtime GraphQL CRUD tab，不显示模型设计、枚举管理等设计时功能
**计划**：待定
**UI 提示**：yes

---

### 阶段 5：测试 & 收口
**目标**：BDD 验收场景覆盖完整登录→exchange→workspace 流程及 scope 边界强执行，集成测试确认无回归
**依赖**：阶段 1、阶段 2、阶段 3、阶段 4
**需求**：TEST-01
**验收标准**（以下条件必须为真）：
  1. BDD 场景：端用户完成登录 → exchange → 进入 workspace 全流程通过
  2. BDD 场景：`scope=org` token 访问 project 路由，服务端返回 403（有对应 feature 文件）
  3. BDD 场景：`scope=project` token 访问 org 路由，服务端返回 403
  4. BDD 场景：无效/过期 token 访问任意受保护端点，服务端返回 401
  5. `just check-all` 通过（lint + build + test）
**计划**：待定

---

## 进度

**执行顺序：** 阶段 1 → 阶段 2 → 阶段 3（可与阶段 4 并行）→ 阶段 4 → 阶段 5

| 阶段 | 计划完成数 | 状态 | 完成日期 |
|------|-----------|------|----------|
| 1. Token 核心统一 | 0/TBD | 未开始 | - |
| 2. 中间件 & Exchange 端点 | 0/TBD | 未开始 | - |
| 3. end_user Schema 清理 | 0/TBD | 未开始 | - |
| 4. 前端 Workspace 入口 | 0/TBD | 未开始 | - |
| 5. 测试 & 收口 | 0/TBD | 未开始 | - |
