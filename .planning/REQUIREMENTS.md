# 需求：统一 Token 体系 & Workspace 入口重构

**定义时间**：2026-05-05  
**核心价值**：所有用户通过统一 Token 管道认证，端用户可通过 Workspace 直接访问 Runtime GraphQL 数据，不再存在独立认证路径。

---

## v1 需求

### Token 统一（后端）

- [ ] **TOKEN-01**：统一 JWT issuer 为 `mc-platform`；`PlatformClaims` 新增 `scope` 字段（`"org"` | `"project"`），废弃 `mc-developer` 和 `mc-enduser` issuer
- [ ] **TOKEN-02**：平台管理员 `POST /api/auth/login` 返回 `scope=org` 的 JWT，格式与端用户一致
- [ ] **TOKEN-03**：端用户 `POST /api/end-user/{orgSlug}/auth/login` 返回 `scope=org` 的 JWT，不再使用 `mc-enduser` issuer
- [ ] **TOKEN-04**：`POST /api/auth/exchange` — 凭 Org Token 换取 Project Token（`scope=project`，token 不携带 projectSlug，TTL 与 Org Token 一致为 1h）
- [ ] **TOKEN-05**：中间件 scope 强制校验 — `scope=org` 的 token 无法访问 `/graphql/org/{orgName}/project/*`（返回 403）；`scope=project` 的 token 无法访问 org 管理路由组

### Schema 清理（后端）

- [ ] **SCHEMA-01**：删除 `api/graph/end_user/` 目录及所有关联的 handler / resolver / 路由注册
- [ ] **SCHEMA-02**：核查原 end_user schema 中存活的 query，确认已迁移至 org/project schema 或可直接删除（项目列表复用 org schema，无独立迁移成本）

### Workspace 入口（前端）

- [ ] **WORKSPACE-01**：新增 `/end-user/{orgSlug}/login` 页面 — 端用户登录入口
- [ ] **WORKSPACE-02**：登录成功后展示 project 列表（通过 org GraphQL `projects` query，RBAC 过滤只显示有 membership 的项目）
- [ ] **WORKSPACE-03**：用户选择 project → BFF 调用 `/api/auth/exchange` 获取 Project Token → 以 httpOnly cookie 存储（BFF 服务端完成 exchange，不暴露给前端 JS）
- [ ] **WORKSPACE-04**：跳转至 `/workspace/{orgSlug}/{projectSlug}` — workspace 页仅展示 runtime GraphQL CRUD tab，不显示模型设计、枚举管理等设计时功能

### 测试

- [ ] **TEST-01**：BDD 验收场景覆盖：端用户登录 → exchange → 进入 workspace 完整流程；`scope=org` token 访问 project 路由返回 403；`scope=project` token 访问 org 路由返回 403；无效 token 返回 401

---

## v2 需求（本期不做）

### 功能权限

- **FEAT-01**：功能权限 RBAC — project admin 与数据人员的 tab 差异展示

### Service Key

- **SVCKEY-01**：应用级 token（`scope=service_key`），当前 claims 结构已预留字段

---

## 明确不做

| 功能 | 原因 |
|------|------|
| 渐进式 issuer 兼容（旧 token 保留期） | 硬切，1h TTL 内自然过期，无需兼容 |
| refresh_tokens 表迁移 | 无需 DB 变更，旧 token 自然失效 |
| 端用户自注册（`/api/end-user/auth/register`） | 保留现有端点不动，本期不修改 |
| 行级数据隔离（RLS） | 独立子系统，与 token 统一无依赖 |

---

## 需求溯源

| 需求 | 阶段 | 状态 |
|------|------|------|
| TOKEN-01 | 阶段 1 — Token 核心统一 | 待开发 |
| TOKEN-02 | 阶段 1 — Token 核心统一 | 待开发 |
| TOKEN-03 | 阶段 1 — Token 核心统一 | 待开发 |
| TOKEN-04 | 阶段 2 — 中间件 & Exchange 端点 | 待开发 |
| TOKEN-05 | 阶段 2 — 中间件 & Exchange 端点 | 待开发 |
| SCHEMA-01 | 阶段 3 — end_user Schema 清理 | 待开发 |
| SCHEMA-02 | 阶段 3 — end_user Schema 清理 | 待开发 |
| WORKSPACE-01 | 阶段 4 — 前端 Workspace 入口 | 待开发 |
| WORKSPACE-02 | 阶段 4 — 前端 Workspace 入口 | 待开发 |
| WORKSPACE-03 | 阶段 4 — 前端 Workspace 入口 | 待开发 |
| WORKSPACE-04 | 阶段 4 — 前端 Workspace 入口 | 待开发 |
| TEST-01 | 阶段 5 — 测试 & 收口 | 待开发 |

**覆盖率**：
- v1 需求：12 条
- 已映射至阶段：12 条
- 未映射：0 ✓

---
*需求定义时间：2026-05-05*  
*最后更新：2026-05-05 初始定义*
