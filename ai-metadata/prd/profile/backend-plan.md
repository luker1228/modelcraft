# backend-plan.md

## Profile 分表后端实现方案（不写代码）

## 0. 目标与约束确认（实施前共识）

### 目标
落地 `user` / `profile` 分表，形成最小闭环：注册自动建档、资料可编辑、可一次查询 User+Profile。

### 已确认约束（本方案严格遵循）
1. `user` 仅保留账号主体字段，至少包含 `id/phone/userName`，注册不需要 email。  
2. 注册成功必须自动创建 `profile`，并写入默认昵称（固定前缀 + 随机后缀）。  
3. `myUserProfile` 查询中若 profile 缺失，返回 `ProfileNotFound`。  
4. 保留 `me`，新增 payload 型查询 `myUserProfile`。  
5. 头像能力暂不细化，服务端可 mock。  
6. 暂不考虑历史迁移（不做 backfill/数据修复脚本）。

---

## 1. 变更范围（按分层）

## 1.1 Schema（GraphQL）
- 范围：**Org GraphQL**，新增 Profile 相关类型、错误联合、`myUserProfile`、`updateMyProfile`。
- 保持：`me` 原查询不删不改（兼容）。
- 关键语义：
  - `myUserProfile`：成功与错误互斥；
  - `updateMyProfile`：PATCH 语义；
  - 输入至少一个字段，空输入返回 `InvalidProfileInput`；
  - profile 缺失返回 `ProfileNotFound`。

## 1.2 OpenAPI（Auth）
- 范围：增强 `POST /api/auth/register` 契约，明确注册成功语义为 **user+profile 已创建**。
- 保持：注册入参为 `phone + userName + password`，不引入 email。
- 错误语义：400/409/500 与现有 Auth 错误体系对齐。

## 1.3 Domain
- 新增/补充 `Profile` 聚合（`id/userId/nickname/avatarUrl/bio`）与不变量：
  - 一个 user 只能有一个 profile（`userId` 唯一）；
  - 注册阶段必须初始化默认昵称。
- User 聚合保持账号主体职责，不回流资料字段。

## 1.4 Repo（含 DB 查询层）
- 新增 profile 持久化能力：
  - `CreateInitialProfile`
  - `FindByUserID`
  - `UpdateByUserID`（按 PATCH 语义）
- 注册链路需要事务编排：创建 user 与 profile 原子提交。
- 数据约束：
  - `user.phone` 唯一
  - `user.userName` 唯一
  - `profile.user_id` 唯一 + FK -> `user.id`

## 1.5 App（用例编排）
- 注册用例：`Register` 成功后同事务创建 profile（含默认昵称生成）。
- 查询用例：`GetMyUserProfile` 聚合 user+profile；profile 缺失转业务错误 `NOT_FOUND.PROFILE`。
- 更新用例：`UpdateMyProfile` 进行输入校验与部分字段更新。

## 1.6 Resolver / Handler
- GraphQL Resolver：
  - 新增 `myUserProfile`
  - 新增 `updateMyProfile`
  - 错误映射到 union（含 `ProfileNotFound`）
- OpenAPI Register Handler：
  - 保持路由与认证策略；
  - 返回增强后的注册响应（含 profile 快照）。

## 1.7 Tests
- Domain：Profile 不变量（1:1、默认昵称存在）。
- Repo：唯一约束/FK/更新语义/查询不存在分支。
- App：注册原子性、myUserProfile 缺失策略、更新输入校验。
- Resolver/Handler：GraphQL union 错误分支、REST 状态码与响应结构。
- 回归：`me` 兼容性（行为不变）。

---

## 2. 推荐实现顺序（可执行步骤 + 每步验收标准）

## Step 1：冻结协议边界（GraphQL + OpenAPI）
**动作**
- 先确认并固定合同：Org GraphQL 的 `myUserProfile/updateMyProfile`，Auth REST 的 register 增强。

**验收标准**
- 协议文档/Schema 可表达全部已确认约束；
- 明确 `me` 保留，`myUserProfile` 新增；
- `ProfileNotFound` 已在 `myUserProfile` 错误模型中可返回。

---

## Step 2：数据库结构与约束落地（仅前向变更）
**动作**
- 新增 `profile` 表及索引/FK/唯一约束；
- `user` 结构确认仅账号主体字段（不扩展 email 依赖）。

**验收标准**
- 能在数据库层保证 1:1（`profile.user_id` 唯一）；
- 能在数据库层保证手机号/用户名唯一；
- 不包含历史数据迁移/backfill 脚本。

---

## Step 3：Domain 建模与错误码对齐
**动作**
- 落地 Profile 聚合与默认昵称规则（前缀 + 随机后缀）；
- 明确业务错误：`NOT_FOUND.PROFILE`、`PARAM_INVALID.PROFILE`。

**验收标准**
- 领域不变量可被单测覆盖；
- 昵称规则可重复调用且不破坏约束；
- 错误语义能映射到 GraphQL union / REST 状态码。

---

## Step 4：Repo 能力实现（含事务支持）
**动作**
- 实现 profile 的创建/查询/更新仓储；
- 注册路径支持“创建 user + 创建 profile”同事务提交。

**验收标准**
- 任一子步骤失败时事务整体回滚；
- profile 查询不到可被上层识别为 NotFound 分支；
- 更新仅修改输入字段（PATCH 语义）。

---

## Step 5：App 用例编排
**动作**
- `Register`：注册成功后自动建档；
- `GetMyUserProfile`：返回 user+profile 或 `ProfileNotFound`；
- `UpdateMyProfile`：输入最小校验（至少一项字段）。

**验收标准**
- 注册成功响应可带 profile 快照；
- 缺失 profile 时不返回空成功，必须是 `ProfileNotFound`；
- 空输入不落库，返回 `InvalidProfileInput`。

---

## Step 6：接口层接入（Resolver/Handler）
**动作**
- 接入 Org GraphQL 新查询/新变更；
- 增强 Auth register handler 返回体；
- 保持 `me` 不变。

**验收标准**
- GraphQL 返回结构满足 payload + union 约定；
- OpenAPI register 响应与合同一致；
- 现有依赖 `me` 的调用无行为回归。

---

## Step 7：测试与回归验收
**动作**
- 运行分层测试与关键路径回归；
- 重点覆盖失败路径（冲突、缺失、输入非法、事务回滚）。

**验收标准**
- 新增能力测试通过；
- `me` 回归通过；
- 无违反约束的已知缺口（特别是注册原子性与 ProfileNotFound 策略）。

---

## 3. 风险点与回滚点

| 风险点 | 触发信号 | 影响 | 回滚点（建议按提交粒度） |
|---|---|---|---|
| 合同先后不一致（前后端理解偏差） | codegen 或联调字段不匹配 | 联调阻塞 | **R1**：仅回滚 GraphQL/OpenAPI 合同提交 |
| 事务未闭合导致“有 user 无 profile” | 注册后 `myUserProfile` 返回 ProfileNotFound | 数据不一致 | **R2**：回滚 App+Repo 注册链路提交 |
| 默认昵称冲突概率未处理 | 唯一约束冲突报错 | 注册失败率上升 | **R3**：回滚昵称生成策略提交，恢复上版规则 |
| PATCH 实现误覆盖空值 | 未传字段被清空 | 资料丢失 | **R4**：回滚 profile 更新仓储/用例提交 |
| `me` 被误改 | 既有客户端回归失败 | 兼容性问题 | **R5**：回滚 resolver 接口改动提交 |
| 头像能力误扩展到真实存储 | 引入额外依赖与失败面 | 超范围交付 | **R6**：回滚头像扩展提交，仅保留 mock |

> 回滚原则：优先按“合同层 / 数据层 / 用例层 / 接口层”独立提交，确保可局部回退而非整包回退。

---

## 4. 明确边界：哪些先做、哪些不做

## 4.1 先做（本期必须）
- user/profile 分表与 1:1 约束；
- 注册自动创建 profile + 默认昵称；
- Org GraphQL：`myUserProfile`、`updateMyProfile`；
- `myUserProfile` 的 `ProfileNotFound` 策略；
- 保留 `me`；
- 头像 mock 可用；
- 分层测试与回归。

## 4.2 不做（本期明确排除）
- 历史数据迁移/backfill；
- 社交关系（关注/粉丝）；
- 一对多 profile；
- 复杂隐私可见性策略；
- 真实头像上传/存储/CDN 流程；
- 批量/分页版 User+Profile 查询（后续单独评审）。

---

## 5. 实施完成定义（DoD）

满足以下即视为可进入联调：
1. 注册成功后可立即查询到 profile（同事务语义达成）；  
2. `myUserProfile` 在 profile 缺失时返回 `ProfileNotFound`；  
3. `me` 不受影响；  
4. `updateMyProfile` 为 PATCH 语义且空输入可识别；  
5. 头像字段可读写（mock）；  
6. 相关分层测试通过，关键回归通过。

---

## 6. 方案依据（文档定位）

- PRD 总述与范围：`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/profile.md:16`  
- 不做项：`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/profile.md:21`  
- 领域不变量（注册无 email、默认昵称、1:1）：`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/profile-domain.puml:26`、`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/profile-domain.puml:60`、`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/profile-domain.puml:61`  
- 协议归属（Org GraphQL + Auth REST）：`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/api-contract.md:6`、`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/api-contract.md:7`  
- `myUserProfile` / `updateMyProfile` 契约：`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/api-contract.md:110`、`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/api-contract.md:115`  
- 缺失策略与 me 兼容：`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/api-contract.md:344`、`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/api-contract.md:345`  
- 数据约束与默认昵称规则：`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/api-contract.md:273`、`/data/home/lukemxjia/modelcraft/ai-metadata/prd/profile/api-contract.md:277`