# Policy 配置层

> 依赖：`01-enduserref-field.md`（EndUserRef 字段语义）、`02-model-owner-lifecycle.md`（owner 字段生命周期）
> 对应主 PRD 章节：M2（默认 Policy）、M6（Policy 配置 UI）

---

## 背景

RLS 的执行行为由 **ModelRLSPolicy** 决定，而不是由 owner 字段的存在/缺失二元控制。Policy 允许开发者在"完全隔离""读全部""完全封锁"等多种模式之间切换，满足不同业务场景需求。

---

## Policy 数据模型

```
ModelRLSPolicy
├── modelId          : ModelID       ← 唯一标识，与 Model 1:1 绑定
├── selectPredicate  : JsonExpr      ← SELECT USING（JSON String）
├── insertCheck      : JsonExpr      ← INSERT WITH CHECK（JSON String）
├── updatePredicate  : JsonExpr      ← UPDATE USING（JSON String）
├── updateCheck      : JsonExpr      ← UPDATE WITH CHECK（JSON String）
└── deletePredicate  : JsonExpr      ← DELETE USING（JSON String）
```

Policy 不是独立的聚合根，而是 Model 聚合的内部 Entity，生命周期与 owner 字段绑定。

### 表达式格式：GraphQL JSON

每个字段存储一个 JSON 对象，表示 Boolean 条件表达式。`true` / `false` 为常量简写。

**比较操作符**：
```json
{ "field": { "_eq": value } }
{ "field": { "_neq": value } }
{ "field": { "_gt": value } }
{ "field": { "_in": [v1, v2] } }
```

**逻辑操作符**：
```json
{ "_and": [...conditions] }
{ "_or": [...conditions] }
{ "_not": condition }
```

**特殊值**：
```json
{ "_auth": "uid" }           → 当前 EndUser ID（内置）
{ "_auth": "tenant_id" }     → auth_schema 声明的扩展变量
{ "_ref": "db.table.col" }   → 跨表字段引用（仅 PREDICATE 允许）
```

**EXISTS 子查询**（仅 PREDICATE 允许，CHECK 不允许）：
```json
{
  "_exists": {
    "model": "mc_meta.org_memberships",
    "where": { "user_id": { "_eq": { "_auth": "uid" } } }
  }
}
```

### 操作符允许矩阵

| 操作 | 比较操作符 | _and/_or/_not | _exists | _ref |
|------|-----------|---------------|---------|------|
| selectPredicate | ✅ | ✅ | ✅ | ✅ |
| insertCheck | ✅ | ✅ | ❌ | ❌ |
| updatePredicate | ✅ | ✅ | ✅ | ✅ |
| updateCheck | ✅ | ✅ | ❌ | ❌ |
| deletePredicate | ✅ | ✅ | ✅ | ✅ |

### USING vs WITH CHECK 错误语义

| 语义类型 | 谓词 | 不通过行为 |
|---------|------|---------|
| **USING** | selectPredicate / updatePredicate / deletePredicate | 行"不存在"，静默过滤（0行/0受影响），**不报错** |
| **WITH CHECK** | insertCheck / updateCheck | 整个操作失败，抛 `RLS_CHECK_VIOLATION` |

---

## 5 种预设策略（Preset）

Preset 是五件套 JsonExpr 的命名组合，底层存储始终是五个 JSON 字符串。

| Preset | selectPredicate | insertCheck | updatePredicate | updateCheck | deletePredicate | 典型场景 | 标记 |
|--------|----------------|-------------|-----------------|-------------|-----------------|---------|------|
| `READ_WRITE_OWNER`（**默认**） | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | 任务、订单、个人笔记 | |
| `READ_ALL_WRITE_OWNER` | `true` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | `{"owner":{"_eq":{"_auth":"uid"}}}` | 评论、帖子（可被所有人读取，仅自己可写） | |
| `READ_ALL` | `true` | `false` | `false` | `false` | `false` | 公告、商品目录（只读公共数据） | |
| `READ_WRITE_ALL` | `true` | `true` | `true` | `true` | `true` | 完全开放（无隔离） | ⚠️ 高危 |
| `NO_ACCESS` | `false` | `false` | `false` | `false` | `false` | 系统内部表（完全对终端用户隐藏） | |

**自定义组合**：`setModelRLSPolicy` mutation 允许直接传五件套 JSON 字符串，不限于上述 Preset。当组合不匹配任何 Preset 时，`preset` 字段返回 null。

> ⚠️ **READ_WRITE_ALL 高危说明**：选择此 Preset 时，前端需弹出二次确认弹窗，文案："此策略允许所有终端用户读写任意数据，包括其他用户的数据，请确认你了解风险"。

---

## Policy 生命周期规则

1. **新建 Model**：自动同时创建 owner 字段 + 默认 Policy（Preset = `READ_WRITE_OWNER`），两者同步创建
2. **有 owner 字段**：必然存在 Policy（不变量）
3. **无 owner 字段（导入 Model 或删除 owner 后）**：必然无 Policy（不变量），EndUser **DENY ALL**（非全量访问）
4. **删除 owner 字段**：Policy 同步删除（级联）
5. **导入 Model**：不生成 owner 字段，不创建 Policy → EndUser DENY ALL，开发者需主动配置

> ⚠️ **Default Deny 原则**：无 Policy ≠ 全量访问。无 Policy = DENY ALL，与 owner 字段是否存在无关。

---

## 管理后台 UI 入口

### Policy 配置 Tab

**入口**：Model 详情页 → "访问控制"Tab（仅在 `model.rlsPolicy != null` 时显示）

**Tab 内容**：
- **Preset 选择器**：展示 5 种 Preset 卡片，当前激活的高亮显示
  - 卡片内容：Preset 名称 + 适用场景描述
  - 选中 Preset → 调用 `setModelRLSPolicy`，写入对应的五件套 JsonExpr
  - 选中 `READ_WRITE_ALL`（⚠️高危）→ 先弹二次确认弹窗，确认后再调用
- **可视化条件构建器**：类似 Notion filter，不需要手写 JSON
  ```
  ┌─────────────────────────────────────────────┐
  │ SELECT 条件                                  │
  │  ┌──────────┐ ┌─────┐ ┌───────────────┐    │
  │  │ owner_id │ │ = ▾ │ │ auth.uid()  ▾ │    │
  │  └──────────┘ └─────┘ └───────────────┘    │
  │  + 添加条件   + 添加 EXISTS                 │
  │  关系：● AND  ○ OR                         │
  └─────────────────────────────────────────────┘
  ```
- **JSON 预览**：实时展示当前表达式对应的 JSON，便于高级用户理解底层存储
- **实时校验**：输入或修改表达式后调用 `validateRLSExpr` API，编译期校验：
  - JSON Schema 结构合法性
  - 字段名白名单（对照 Model 字段列表）
  - `_auth` 变量白名单（`uid` 内置 + `auth_schema` 声明）
  - `insertCheck` / `updateCheck` 不含 `_exists` / `_ref`
- **保存反馈**：mutation 成功后 toast 提示"访问控制已更新"

**隐藏条件**：`model.rlsPolicy == null`（即无 owner 字段）时，"访问控制"Tab 不渲染

---

## auth_schema 配置

**入口**：Project 设置页 → "认证变量"配置区

**说明**：开发者可声明额外的 JWT 变量，用于 Policy JSON 中的 `_auth` 引用。`uid` 永远内置，来自 `jwt.user_id`，无需声明。

**配置格式**：
```json
{
  "auth_schema": {
    "tenant_id": { "source": "jwt.tenant_id", "type": "uuid" },
    "role":      { "source": "jwt.role",      "type": "string" }
  }
}
```

**支持类型**：`uuid` | `string` | `integer`

**GraphQL API**：
```graphql
type ProjectAuthSchema {
  variables: [AuthVariable!]!
}
type AuthVariable {
  name:   String!
  source: String!
  type:   String!
}
input SetProjectAuthSchemaInput {
  projectId: ID!
  variables: [AuthVariableInput!]!
}
```

**编译期校验行为**：`_auth.uid` 永远合法；其他 `_auth.<name>` 必须在 `auth_schema` 中声明，否则 `validateRLSExpr` 返回编译错误。

---

## 验收标准

### AC-9：Policy 配置 UI

- [ ] 有 owner 字段的 Model 详情页显示"访问控制"Tab
- [ ] 无 owner 字段的 Model 详情页不显示"访问控制"Tab
- [ ] 访问控制 Tab 中展示 5 种 Preset，当前 Policy 对应的 Preset 高亮
- [ ] 选择 `READ_WRITE_ALL`（⚠️高危）时弹出二次确认弹窗，用户确认后才调用 mutation
- [ ] 选择其他 Preset 后调用 `setModelRLSPolicy`，成功后 Policy 即时生效
- [ ] 提供可视化条件构建器（条件行 + 操作符下拉 + 值输入），支持 AND/OR 关系切换
- [ ] 同时展示 JSON 预览，实时反映构建器中的表达式
- [ ] 调用 `validateRLSExpr` 进行实时校验，表达式不合法时展示具体错误信息
- [ ] `insertCheck` / `updateCheck` 中添加 `_exists` 操作符时，校验报错

### AC-10：Policy 默认值

- [ ] 新建 Model 后，`model.rlsPolicy` 五件套均为 `{"owner":{"_eq":{"_auth":"uid"}}}`
- [ ] `model.rlsPolicy.preset = READ_WRITE_OWNER`

### AC-11：Policy 级联删除

- [ ] 删除 owner 字段后，`model.rlsPolicy` 返回 null
- [ ] 删除 owner 字段后，"访问控制"Tab 消失
- [ ] 无 Policy 后 EndUser 调用 Runtime → DENY ALL（非全量访问）

### AC-12：Policy 修改生效

- [ ] 将 Policy 改为 `READ_ALL`（selectPredicate=`true`，其余=`false`）后，EndUser 可读取全量数据，但写操作返回 `RLS_CHECK_VIOLATION`
- [ ] 将 Policy 改为 `NO_ACCESS` 后，EndUser 的读写操作均受限（静默空结果 + CHECK 报错）

### AC-13：auth_schema 配置

- [ ] Project 设置页显示"认证变量"配置区，可添加/删除变量
- [ ] `uid` 内置，不可删除，不可覆盖
- [ ] 添加变量 `tenant_id`（source: `jwt.tenant_id`, type: `uuid`）后，Policy 构建器中可引用 `_auth.tenant_id`
- [ ] 调用 `validateRLSExpr` 时，使用未声明变量（如 `_auth.org_id`）报编译错误

---

## 不做什么（本子页 Out of scope）

- **Policy 变更历史审计**：不记录变更日志
- **多 Policy 并存**：每个 Model 只有一个 Policy，不支持按角色/条件分支的复合 Policy
