# Model 数据隔离（RLS）PRD

## 是什么问题

开发者用 ModelCraft 搭建多用户应用（任务管理、电商订单、CRM 等），终端用户（EndUser）通过 Runtime GraphQL API 操作数据。**当前系统无法自动隔离数据行**：

- 用户 A 登录后调用 `orders` 查询，能看到用户 B 的订单
- 用户 A 发起 updateOrder mutation，可以修改任意其他用户的数据
- 数据归属完全靠开发者在每个 query/mutation 里手动加 `where: { owner: $me }` 过滤

这意味着：**开发者必须在每一个涉及用户数据的操作里都记得加 filter，一旦遗漏就是数据泄露**。在多用户场景下，这不是偶发的 bug，而是系统性的安全风险。

类比 Supabase：Supabase 的解法是数据库原生 RLS（`CREATE POLICY ... USING (owner_id = auth.uid())`）+ 列默认值 `DEFAULT auth.uid()`。ModelCraft 在应用层提供等效能力，但体验更简单——开发者不需要写 SQL。

---

## 目标用户

**使用 ModelCraft 构建多用户应用的开发者。**

典型场景：
- 搭建任务管理 App，每个用户只能看到自己创建的任务
- 搭建电商后台，买家只能访问自己的订单
- 搭建多租户 SaaS，数据按用户隔离

---

## 目标与成功标准

> **核心目标**：新建 Model 开箱即用数据隔离，零配置，无法被终端用户绕过。

### 成功指标

| 指标 | 目标 |
|------|------|
| **零配置上手** | 新建 Model 无需任何额外操作，RLS 自动开启 |
| **向后兼容** | 导入的 Model、未配置 RLS 的 Model，EndUser DENY ALL（Default Deny 原则） |
| **数据隔离可靠性** | 启用 RLS 的 Model，终端用户无法通过任何 Runtime 操作访问到不属于自己的行 |
| **开发者免疫** | 持有开发者 JWT 的请求绕过所有 RLS，保持现有调试能力 |

---

## 核心概念：EndUserRef 字段

RLS 依赖一种新的字段 Format：**`EndUserRef`**。

- **语义**：该字段存储一个 EndUser 的 ID，表示"这条数据归属于哪个终端用户"
- **存储**：UUID 字符串，数据库层有外键约束指向 `private_{projectSlug}.users.id`
- **约束**：一个 Model 最多只有一个 `EndUserRef` 字段（第一期）
- **RLS 绑定**：有 `EndUserRef` 字段 = 可配置 Policy；无 = 无 Policy，EndUser DENY ALL，不显示 Policy 配置入口。
- **字段名**：固定为 `owner`，不可改名

### owner 字段的可见性

| 调用者 | Query 结果 | Mutation input |
|--------|-----------|----------------|
| **EndUser** | 可见，但只读（不出现在 input 类型中） | 不暴露，系统自动填充/强制覆盖 |
| **Developer** | 可见，无限制 | 可传入，不强制覆盖 |

---

## 核心概念：ModelRLSPolicy

**`ModelRLSPolicy`** 是 RLS 的执行策略，描述 Runtime 对该 Model 数据的读写行为。

```
ModelRLSPolicy
├── modelId          : ModelID
├── selectPredicate  : JsonExpr   ← SELECT USING（USING 语义）
├── insertCheck      : JsonExpr   ← INSERT WITH CHECK（CHECK 语义）
├── updatePredicate  : JsonExpr   ← UPDATE USING（USING 语义）
├── updateCheck      : JsonExpr   ← UPDATE WITH CHECK（CHECK 语义）
└── deletePredicate  : JsonExpr   ← DELETE USING（USING 语义）
```

### 表达式格式：GraphQL JSON

Policy 的每个字段存储一个 JSON 对象，表示 Boolean 条件表达式。

**比较操作符**：
```json
{ "field": { "_eq": value } }
{ "field": { "_neq": value } }
{ "field": { "_gt": value } }
{ "field": { "_gte": value } }
{ "field": { "_lt": value } }
{ "field": { "_lte": value } }
{ "field": { "_in": [v1, v2] } }
{ "field": { "_nin": [v1, v2] } }
{ "field": { "_is_null": true } }
```

**逻辑操作符**：
```json
{ "_and": [...conditions] }
{ "_or": [...conditions] }
{ "_not": condition }
```

**特殊值（变量引用）**：
```json
{ "_auth": "uid" }              → 当前 EndUser ID（内置，来自 jwt.user_id）
{ "_auth": "tenant_id" }        → auth_schema 声明的扩展变量
{ "_ref": "db.table.column" }   → 跨表字段引用（仅 PREDICATE 允许）
```

**EXISTS 子查询**（仅 PREDICATE 允许，CHECK 不允许）：
```json
{
  "_exists": {
    "model": "mc_meta.org_memberships",
    "where": {
      "org_id": { "_eq": { "_ref": "user_db.projects.org_id" } },
      "user_id": { "_eq": { "_auth": "uid" } }
    }
  }
}
```

**常量简写**：`true` 和 `false` 等价于 `{"_const": true}` / `{"_const": false}`

### 操作符允许矩阵

| 操作 | 比较操作符 | _and/_or/_not | _exists | _ref |
|------|-----------|---------------|---------|------|
| selectPredicate | ✅ | ✅ | ✅ | ✅ |
| insertCheck | ✅ | ✅ | ❌ | ❌ |
| updatePredicate | ✅ | ✅ | ✅ | ✅ |
| updateCheck | ✅ | ✅ | ❌ | ❌ |
| deletePredicate | ✅ | ✅ | ✅ | ✅ |

### auth 上下文

- `uid` 内置，永远来自 `jwt.user_id`，无需声明
- 其他变量在 Project 设置里声明 `auth_schema`（Project 级别）：
```json
{
  "auth_schema": {
    "tenant_id": { "source": "jwt.tenant_id", "type": "uuid" },
    "role":      { "source": "jwt.role",      "type": "string" }
  }
}
```
- 编译期校验：`_auth.uid` 永远合法，其他变量必须在 `auth_schema` 中声明

### USING vs WITH CHECK 错误语义（关键）

```
USING 不通过（selectPredicate / updatePredicate / deletePredicate）：
  → 行"不存在"，静默过滤
  → SELECT 返回空结果（不报错）
  → UPDATE/DELETE 返回 0 行受影响（不报错，不泄露行存在信息）

WITH CHECK 不通过（insertCheck / updateCheck）：
  → 直接抛错，整个操作失败
  → 错误码：RLS_CHECK_VIOLATION
  → INSERT/UPDATE 失败，明确告知"写入数据违反策略约束"
```

### 5 种预设策略（Preset）

| Preset | selectPredicate | insertCheck | updatePredicate | updateCheck | deletePredicate | 标记 |
|--------|----------------|-------------|-----------------|-------------|-----------------|------|
| `READ_WRITE_OWNER`（**默认**） | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | |
| `READ_ALL_WRITE_OWNER` | `true` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | |
| `READ_ALL` | `true` | `false` | `false` | `false` | `false` | |
| `READ_WRITE_ALL` | `true` | `true` | `true` | `true` | `true` | ⚠️ 高危 |
| `NO_ACCESS` | `false` | `false` | `false` | `false` | `false` | |

Preset 是命名快捷方式，底层存储的始终是五件套 JsonExpr 值（`true`/`false` 为常量简写）。选择 `READ_WRITE_ALL` 时需二次确认（高危）。

### Policy 与 owner 字段的关系

- **有 owner 字段**：自动创建 Policy（默认 Preset = `READ_WRITE_OWNER`），显示 Policy 配置入口
- **无 owner 字段**：无 Policy，**EndUser DENY ALL（不是全量访问）**，不显示 Policy 配置入口
- **删除 owner 字段**：Policy 同步删除
- **导入 Model**：无 owner 字段，无 Policy → EndUser DENY ALL，需开发者主动配置

> ⚠️ **Default Deny 原则**：无 Policy = DENY ALL，与有没有 owner 字段无关。owner 字段只是 `{"owner":{"_eq":{"_auth":"uid"}}}` 谓词的依赖，不是 RLS 开关。

### Runtime 执行语义

```
USING 谓词（selectPredicate / updatePredicate / deletePredicate）：
  true（常量）            → 不注入 WHERE（全量）
  {"owner":{"_eq":...}}   → PolicyExecutor.ToSQL() 编译为 AND owner = ? + args
  false（常量）           → 注入 AND 1=0（静默返回空/0行，不报错）
  通用 JSON 表达式        → PolicyCompiler 递归编译为参数化 SQL + args[]

WITH CHECK 谓词（insertCheck / updateCheck）：
  true（常量）            → 允许写入
  {"owner":{"_eq":...}}   → 应用层校验 owner = endUserId，违反时抛 RLS_CHECK_VIOLATION
  false（常量）           → 直接抛 RLS_CHECK_VIOLATION（拒绝所有写入）
  通用 JSON 表达式        → PolicyExecutor 应用层校验（不含 _exists / _ref）

无 Policy（model.getPolicy() == nil）：
  EndUser → DENY ALL（所有操作均被拒绝）
```

---

## 用户故事

### Story 1：新建 Model，零配置开箱即用

> 作为**开发者**，我新建一个 `orders` Model，系统自动生成一个 `owner` 字段（EndUserRef 类型）和默认 Policy（读写自己：readScope=OWNER, writeScope=OWNER）。
>
> 终端用户调用 Runtime GraphQL 时，系统自动加上 `WHERE owner = <当前用户ID>`，我无需在任何查询里手动传这个条件。CreateOne 时 `owner` 字段也自动填充为当前 EndUser ID。我可以在 Model 详情页的"访问控制"Tab 中修改 Policy。

**验收场景**：
- 新建 Model 后，字段列表中自动存在一个 `owner` 字段，类型为 `EndUserRef`
- 新建 Model 后，自动创建默认 Policy（Preset = 读写自己）
- Runtime 查询自动隔离，EndUser 只能看到自己的数据（readScope=OWNER）
- CreateOne 时 `owner` 字段无需传入，系统自动填充当前 EndUser ID（writeScope=OWNER）
- Model 详情页显示"访问控制"Tab，开发者可修改 Preset

---

### Story 2：删除 owner 字段，进入默认拒绝（EndUser DENY ALL）

> 作为**开发者**，我删除 `owner` 字段后，Policy 同步删除。由于 `无 Policy = Default Deny`，该 Model 对 EndUser 的 Runtime 访问将被拒绝；如需公开读取，应显式保留 owner 并将 Policy 调整为 `READ_ALL` 等允许读的策略。

**验收场景**：
- 删除 `EndUserRef` 字段后，该 Model 的 Policy 同步删除
- Runtime 请求不再执行基于 Policy 的 WHERE 注入，但因无 Policy 按 Default Deny 拒绝 EndUser 访问
- EndUser 调用该 Model 的 Runtime 读写均 DENY ALL（非全量访问）
- Model 详情页不再显示"访问控制"Tab

---

### Story 3：终端用户只能访问自己的数据

> 作为**终端用户**，无论我传什么查询条件，都只能看到归属于我的数据行，系统不允许我访问其他用户的数据。

**验收场景**：
- EndUser A 查询 `orders`，只返回 `owner = A` 的行，即使不带任何 filter
- EndUser A 显式传 `where: { owner: B_id }` → 返回空（WHERE 条件取交集）
- EndUser A 尝试 update/delete `owner = B` 的记录 → 返回"记录不存在"（非 403）

---

### Story 4：开发者通过 EndUser 身份调试 Runtime

> 作为**开发者**，我想调试 Runtime GraphQL，需要先在管理后台创建一个测试用的 EndUser 账号，用该账号登录拿到 EndUser JWT，再调用 Runtime。Runtime 不接受开发者 JWT。

**验收场景**：
- Developer JWT 直接调用 Runtime → 401，拒绝访问
- 开发者用测试 EndUser 账号登录，拿到 EndUser JWT → 正常调用，RLS 生效
- 开发者可以在管理后台创建任意数量的测试 EndUser 账号用于调试

---

### Story 5：从表导入的 Model 默认 EndUser DENY ALL

> 作为**开发者**，我从已有数据库表导入 Model，系统不自动添加 owner 字段，也不创建 Policy。**导入后 EndUser 调用 Runtime 将 DENY ALL（无法访问任何数据）**，我需要主动配置 Policy 才能开放访问。

**验收场景**：
- 导入 Model 后，字段列表中无 `owner` 字段，无 Policy
- EndUser 调用 Runtime 操作该 Model → DENY ALL（所有操作被拒绝），而非全量访问
- 开发者在"访问控制"Tab 配置 Policy 后，Runtime 按新 Policy 执行
- 当前版本不支持为导入 Model 自动创建 owner 字段

---

## 功能范围（Must-have）

### M1：EndUserRef 字段 Format

- 新增字段格式类型 `EndUserRef`，字段名固定为 `owner`
- 存储语义：指向当前项目 EndUser 的 ID
- 数据库层外键约束：`REFERENCES private_{projectSlug}.users(id)`
- 一个 Model 最多只能有一个 `EndUserRef` 字段（创建第二个时报错）
- **EndUser 可见性**：`owner` 字段在 Query 结果中对 EndUser 可见（只读），不出现在 EndUser 的 Mutation input 类型中
- **Developer 可见性**：完整暴露，无限制

### M2：新建 Model 自动生成 owner 字段 + 默认 Policy

- 新建 Model 时，系统自动添加名为 `owner` 的 `EndUserRef` 字段，同时创建默认 Policy（Preset = `READ_WRITE_OWNER`：五件套均为 `OWNER_EQUALS_USER`）
- owner 字段与 Policy **同步创建**，两者生命周期绑定
- 该字段可以被开发者删除；删除 owner 字段时，Policy 同步删除
- 无 Policy 的 Model 不显示 Policy 配置入口

### M3：Runtime 只接受 EndUser JWT

- Runtime endpoint 只接受 EndUser JWT，拒绝 Developer JWT（返回 401）
- EndUser ID 从 JWT 中提取，注入 request context
- 不存在"开发者绕过 RLS"的模式，所有 Runtime 调用均受 RLS 约束

### M4：Runtime WHERE 自动注入

检测到 EndUser JWT 时，按以下规则执行：

**有 Policy 时**（`model.getPolicy() != nil`）：

**USING 谓词（selectPredicate / updatePredicate / deletePredicate）**：
- `true`（常量）→ 不注入 WHERE，全量返回/操作
- JSON 表达式 → PolicyExecutor.ToSQL(compiled, authCtx) 生成参数化 SQL WHERE 子句注入；USING 不通过 → 行"不存在"，静默过滤（0行/0受影响，不报错）
- `false`（常量）→ 追加 `AND 1=0`；静默返回空/0行，不报错

**WITH CHECK 谓词（insertCheck / updateCheck）**：
- `true`（常量）→ 允许写入
- JSON 表达式 → PolicyExecutor 应用层校验（不含 `_exists` / `_ref`），违反时返回 `RLS_CHECK_VIOLATION`
- `false`（常量）→ 直接返回 `RLS_CHECK_VIOLATION`

**CreateOne 额外规则**：当 insertCheck 含 `{"_auth":"uid"}` 等于 owner 约束时，强制将 `owner` 填充为当前 EndUser ID（覆盖传入值），`owner` 不允许为空。

**无 Policy 时**（`model.getPolicy() == nil`）：
- EndUser 调用 → **DENY ALL**（所有操作均被拒绝，返回 `PERMISSION_DENIED`）

**触发条件**：请求携带合法 EndUser JWT（`iss = mc-enduser`），`endUserId` 已注入 context。

### M5：导入 Model 默认 EndUser DENY ALL

- 从表导入的 Model 不自动添加 `owner` 字段，也不创建 Policy
- 导入后 EndUser 调用 Runtime → DENY ALL（无 Policy = 拒绝所有访问）
- 开发者可在"访问控制"Tab 主动配置 Policy 后开放访问

### M6：Policy 配置

- 管理后台 Model 详情页新增"访问控制"Tab（仅在有 owner 字段时显示）
- 开发者可选择 5 种 Preset 之一：`READ_WRITE_OWNER` / `READ_ALL_WRITE_OWNER` / `READ_ALL` / `READ_WRITE_ALL`（⚠️高危，需二次确认） / `NO_ACCESS`
- Preset 选择后，底层写入五件套 JsonExpr 字段（JSON 字符串）
- 提供**可视化条件构建器**（类似 Notion filter），不需要手写 JSON；同时展示 JSON 预览
- 提供 `setModelRLSPolicy` mutation，支持直接传五件套 JSON 字符串（不限于 Preset）
- 支持实时 JSON 校验（调用 `validateRLSExpr` API 在编译期验证表达式合法性）
- 删除 owner 字段时，Policy 同步删除，"访问控制"Tab 不再显示

### M7：auth_schema 配置（Project 级别）

- 在 Project 设置页新增"认证变量"配置区
- 开发者可声明扩展 JWT 变量（如 `tenant_id`、`role`）及其来源和类型
- 声明后可在 Policy JSON 中通过 `{"_auth": "tenant_id"}` 引用
- `uid` 内置，永远来自 `jwt.user_id`，无需声明
- 提供 `setProjectAuthSchema` / `getProjectAuthSchema` mutation/query

---

## 不做什么（Out of scope）

| 排除项 | 原因 |
|--------|------|
| **为导入 Model 开启 RLS** | 第一期只支持新建 Model，导入场景后续扩展 |
| **多个 EndUserRef 字段** | 第一期每个 Model 只允许一个，后续扩展 |
| **RLS 开关 UI** | 字段即开关，不需要单独的 toggle |
| **开发者角色级访问控制** | 本功能仅针对终端用户行隔离，开发者访问控制是独立方向 |
| **审计日志** | 访问记录和策略变更历史不在本期 |
| **数据库原生 RLS** | 应用层实现，保持对外部 MySQL 集群的兼容性 |

---

## 验收标准

### AC-1：EndUserRef 字段约束

- [ ] 新建 `EndUserRef` 字段，DB 层生成对应外键约束
- [ ] 同一 Model 尝试添加第二个 `EndUserRef` 字段 → 报错提示"每个 Model 只允许一个归属字段"

### AC-2：新建 Model 自动生成 owner + 默认 Policy

- [ ] 新建 Model 后，字段列表中自动存在名为 `owner` 的 `EndUserRef` 字段
- [ ] 新建 Model 后，自动创建默认 Policy（五件套均为 `{"owner":{"_eq":{"_auth":"uid"}}}`，Preset=READ_WRITE_OWNER）
- [ ] 导入 Model 后，字段列表中无 `owner` 字段，无 Policy，EndUser DENY ALL
- [ ] 删除 `owner` 字段时，弹出二次确认弹窗，提示"删除后数据隔离将关闭"
- [ ] 删除 `owner` 字段后，Policy 同步删除

### AC-3：行级过滤生效（依赖 Policy）

- [ ] selectPredicate=`{"owner":{"_eq":{"_auth":"uid"}}}`：EndUser A 查询该 Model，只返回 `owner = A` 的行
- [ ] selectPredicate=`true`：EndUser A 查询该 Model，返回全量数据
- [ ] selectPredicate=`false`：EndUser A 查询该 Model，返回空集（静默，不报错）
- [ ] EndUser A 显式传入 `where: { owner: B_id }` 且 selectPredicate 含 owner 等于当前 uid 约束 → 返回空，而不是 B 的数据
- [ ] 未配置 Policy 的 Model（无 owner 字段），EndUser DENY ALL（不是全量访问）
- [ ] USING 失败 → 静默（0行/0受影响），不报错，不返回 403

### AC-4：CreateOne 自动填充

- [ ] EndUser A 调用 CreateOne 不传 `owner` → 自动填充为 A 的 ID
- [ ] EndUser A 调用 CreateOne 故意传 `owner = B` → 被覆盖为 A 的 ID

### AC-5：Mutation 保护（依赖 Policy）

- [ ] updatePredicate=owner 等于当前 uid：EndUser A 调用 UpdateOne（指向 `owner = B` 的记录）→ 静默返回 0 行受影响（不报错，不返回 403）
- [ ] deletePredicate=owner 等于当前 uid：EndUser A 调用 DeleteOne（指向 `owner = B` 的记录）→ 静默返回 0 行受影响（不报错，不返回 403）
- [ ] insertCheck=`false`：EndUser A 调用 CreateOne → 返回 `RLS_CHECK_VIOLATION`（整个操作失败）
- [ ] updateCheck=owner 等于当前 uid：EndUser A 调用 UpdateOne 设置 `owner = B` → 返回 `RLS_CHECK_VIOLATION`
- [ ] CHECK 失败（insertCheck/updateCheck 不通过）→ 返回 `RLS_CHECK_VIOLATION`，整个操作失败，明确报错

### AC-9：Policy 配置 UI

- [ ] 有 owner 字段的 Model 详情页显示"访问控制"Tab
- [ ] 无 owner 字段的 Model 详情页不显示"访问控制"Tab
- [ ] 访问控制 Tab 中可选择 5 种 Preset，选择后自动写入对应的五件套 JsonExpr 字段
- [ ] 选择 `READ_WRITE_ALL`（⚠️高危）时弹出二次确认弹窗，文案："此策略允许所有终端用户读写任意数据，包括其他用户的数据，请确认你了解风险"
- [ ] 提供可视化条件构建器（类似 Notion filter）和 JSON 预览，不需要手写 JSON
- [ ] 支持实时 JSON 校验（`validateRLSExpr` API），表达式非法时给出错误提示
- [ ] 调用 `setModelRLSPolicy` 后，Policy 即时生效（Runtime 下一次请求即按新 Policy 执行）
- [ ] 删除 owner 字段后，访问控制 Tab 消失，`model.rlsPolicy` 返回 null

### AC-13：auth_schema 配置

- [ ] Project 设置页显示"认证变量"配置区
- [ ] 开发者可声明扩展变量（名称、JWT 来源、类型），`uid` 内置不可修改
- [ ] Policy 可视化构建器中，值下拉自动显示已声明的 auth 变量
- [ ] 调用 `validateRLSExpr` 时，未在 `auth_schema` 声明的 `_auth` 变量报编译错误

### AC-6：Runtime 只接受 EndUser JWT

- [ ] Developer JWT 调用 Runtime → 401，拒绝访问
- [ ] EndUser JWT 调用 Runtime → 正常执行，RLS 生效
- [ ] CreateOne 时 `owner` 字段强制填充为当前 EndUser ID，不允许为空，传入值被覆盖

### AC-7：owner 字段可见性

- [ ] EndUser 调用 Query → 返回结果中包含 `owner` 字段（值为自己的 ID）
- [ ] EndUser 调用 CreateOne / UpdateOne → input 类型中不暴露 `owner` 字段
- [ ] Developer 调用 Query / Mutation → `owner` 字段完整暴露，行为无变化

### AC-8：向后兼容

- [ ] 升级后，现有无 EndUserRef 字段的 Model，EndUser 调用 Runtime → DENY ALL（无 Policy = 拒绝，非旧行为）
- [ ] 现有 BDD / 集成测试全部通过，无需修改

---

## 待确认

| # | 问题 | 建议 |
|---|------|------|
| Q1 | ✅ **JWT issuer 规范**：两类 JWT 通过 `iss` 字段区分。Developer JWT `iss: "mc-developer"`，EndUser JWT `iss: "mc-enduser"`。Runtime endpoint 只接受 `mc-enduser`，设计态 GraphQL 只接受 `mc-developer`。现有 `iss: "modelcraft"` 的 Developer JWT 需同步迁移。 | 已确认 |
| Q2 | ✅ **无 Policy 的 Model**：EndUser DENY ALL（所有操作被拒绝）。适用于导入 Model 或尚未配置 Policy 的场景，开发者需主动配置 Policy 后才能开放 EndUser 访问。 | 已确认 |
