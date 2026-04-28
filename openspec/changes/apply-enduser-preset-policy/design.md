## Context

当前 RBAC 权限系统已支持手动创建权限点（`EndUserPermission`）和权限包，管理员可精细配置行/列访问控制。但缺乏快速配置入口——为一个模型建立基础权限需要手动创建多个权限点。

预设策略在此基础上引入"一键应用标准权限组合"能力。

**本次变更同时对 `end_user_permissions` 表结构做了重构（见 D6）**：
- 删除 `action` / `row_scope` 字段（原始行为抽象）
- 新增 `type` ENUM(PRESET, CUSTOM)（来源标记）
- 新增 `row_policy` JSON（直接存储 Runtime where 谓词，结构见下方）

## Goals / Non-Goals

**Goals:**
- 管理员可对指定模型应用 4 种预设，一次替换所有 preset 来源的权限点
- 保留 `type = CUSTOM`（手动创建）的权限点不受影响
- 依赖 `END_USER_REF` 字段的预设（`*_OWNER`）在字段缺失时返回结构化错误
- `CreatePermission` 统一支持 `type` / `row_policy` / `preset` 字段（不拆分接口）

**Non-Goals:**
- 预设策略的前端 UI（本变更仅实现后端链路）
- 支持自定义/扩展预设（固定 4 种）
- 预设权限点创建后的独立编辑（需先删除再手动新建）

## 数据模型

### `end_user_permissions` 终态字段（`db/schema/mysql/13_rbac_permissions.sql`）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | VARCHAR(36) | UUID |
| `org_name` | VARCHAR(64) | 所属组织 |
| `project_slug` | VARCHAR(64) | 所属项目 |
| `model_id` | VARCHAR(36) | FK → models.id，CASCADE |
| `name` | VARCHAR(128) | 人类可读名称 |
| `description` | TEXT | 描述 |
| `type` | ENUM(PRESET, CUSTOM) | **新增**：来源标记 |
| `column_policy` | JSON | 列策略（保留） |
| `row_policy` | JSON | **新增**：行策略五件套谓词 |
| `preset` | ENUM(四种预设) NULL | **原有**：仅 type=PRESET 时有值 |

**删除字段：** `action`（ENUM select/insert/update/delete/export）、`row_scope`（ENUM ALL/SELF/DEPT/DEPT_AND_CHILDREN）

**UNIQUE KEY 变更：** `(model_id, action, row_scope, name)` → `(model_id, type, name)`

### `row_policy` JSON 结构

每个 action 用两个固定字段表达语义，predicate/check 为 GraphQL Runtime where 条件 JSON：

```
allowed : boolean         是否允许该操作；false 时其余字段忽略
scope   : "all"|"custom"  all=全通行（predicate/check 忽略），custom=条件生效
```

**隐藏规则**：`select.allowed = false` 时，insert / update / delete 全部强制 `allowed = false`，UI 置灰不可配置。

```json
{
  "select": {
    "allowed": true,
    "scope": "custom",
    "predicate": { "status": { "_eq": "active" } }
  },
  "insert": {
    "allowed": true,
    "scope": "custom",
    "check": { "owner_id": { "_eq": "$endUserId" } }
  },
  "update": {
    "allowed": true,
    "scope": "custom",
    "predicate": { "owner_id": { "_eq": "$endUserId" } },
    "check_scope": "custom",
    "check": { "owner_id": { "_eq": "$endUserId" } }
  },
  "delete": {
    "allowed": false
  }
}
```

`update` 有两层独立配置：
- `scope` + `predicate`：能操作哪些已有的行（UPDATE USING）
- `check_scope` + `check`：写入后数据必须满足什么（UPDATE WITH CHECK，防止把数据"改出"约束）

`check_scope` 只有 `"all"` | `"custom"`，无 `none`（不允许修改直接用 `update.allowed=false`）。

### `column_policy` JSON 结构（新增 `writable` 字段）

```json
{
  "defaultMode": "VISIBLE",
  "rules": [
    { "fieldName": "salary",  "mode": "MASKED",  "maskPattern": "***", "writable": false },
    { "fieldName": "id_card", "mode": "HIDDEN",  "writable": false },
    { "fieldName": "status",  "mode": "VISIBLE", "writable": false }
  ]
}
```

`writable`：该列是否允许被 UPDATE SET，默认 `true`；仅 `mode = VISIBLE` 时有实际意义。

### UI 语义层 → 底层字段映射

```
┌──────────┬────────────────────────────────┬──────────────────────────────────────┐
│  操作    │  UI 选项                       │  row_policy 字段                     │
├──────────┼────────────────────────────────┼──────────────────────────────────────┤
│          │ 允许查看全部                   │ select.allowed=true,  scope="all"    │
│  查看    │ 允许查看，设定范围             │ select.allowed=true,  scope="custom" │
│          │ 不允许查看 ─────────────────► │ select.allowed=false                 │
│          │     └─ 级联禁用所有其他操作    │   → 其余 action.allowed 强制 false   │
├──────────┼────────────────────────────────┼──────────────────────────────────────┤
│          │ 允许创建（无约束）             │ insert.allowed=true,  scope="all"    │
│  创建    │ 允许创建（约束归属）           │ insert.allowed=true,  scope="custom" │
│          │ 不允许创建                     │ insert.allowed=false                 │
├──────────┼────────────────────────────────┼──────────────────────────────────────┤
│          │ 允许修改全部范围               │ update.allowed=true,  scope="all"    │
│  修改    │ 允许修改，限定可操作范围       │ update.allowed=true,  scope="custom" │
│  范围    │ 不允许修改                     │ update.allowed=false                 │
│  ──────  │ ─────────────────────────────  │ ─────────────────────────────────    │
│  写入    │ 写入后无约束                   │ check_scope="all"                    │
│  约束    │ 写入后归属不变                 │ check_scope="custom", check={...}    │
│  ──────  │ ─────────────────────────────  │ ─────────────────────────────────    │
│  列限制  │ 哪些列可被 SET                 │ column_policy.rules[].writable=false │
├──────────┼────────────────────────────────┼──────────────────────────────────────┤
│          │ 允许删除全部                   │ delete.allowed=true,  scope="all"    │
│  删除    │ 允许删除，限定范围             │ delete.allowed=true,  scope="custom" │
│          │ 不允许删除                     │ delete.allowed=false                 │
└──────────┴────────────────────────────────┴──────────────────────────────────────┘
```

### 4 种预设的 row_policy 展开（App 层硬编码，固定值）

每种预设 apply 后在表中生成 **1 条记录**，name 格式固定为 `"preset:<PRESET_NAME>"`。

| 预设 | select | insert | update（范围） | update（写入约束） | delete |
|------|--------|--------|---------------|-------------------|--------|
| READ_WRITE_ALL | allowed=true, scope=all | allowed=true, scope=all | allowed=true, scope=all | check_scope=all | allowed=true, scope=all |
| READ_ALL | allowed=true, scope=all | allowed=false | allowed=false | — | allowed=false |
| READ_WRITE_OWNER | allowed=true, scope=custom, predicate={owner=$eu} | allowed=true, scope=custom, check={owner=$eu} | allowed=true, scope=custom, predicate={owner=$eu} | check_scope=custom, check={owner=$eu} | allowed=true, scope=custom, predicate={owner=$eu} |
| READ_ALL_WRITE_OWNER | allowed=true, scope=all | allowed=true, scope=custom, check={owner=$eu} | allowed=true, scope=custom, predicate={owner=$eu} | check_scope=custom, check={owner=$eu} | allowed=true, scope=custom, predicate={owner=$eu} |

> `$eu` = Runtime 注入的当前 EndUser ID；`owner` = 模型上 END_USER_REF 字段名（apply 前校验存在性）

## ER 图

```
┌──────────────────┐         ┌─────────────────────────────────────────────────────┐
│    models        │         │  end_user_permissions                                │
│──────────────────│         │─────────────────────────────────────────────────────│
│ id (PK)          │ 1     N │ id (PK)                                              │
│ org_name         ├─────────┤ model_id  FK → models.id  CASCADE                   │
│ project_slug     │         │ org_name / project_slug                              │
│ name             │         │ name                                                 │
└──────────────────┘         │ type          PRESET | CUSTOM                        │
                             │ column_policy JSON                                   │
┌──────────────────┐         │ row_policy    JSON  ← 替代 action + row_scope        │
│    fields        │  校验   │   { select:  { allowed, scope, predicate }           │
│──────────────────│◄────────│     insert:  { allowed, scope, check }               │
│ format =         │ *_OWNER │     update:  { allowed, scope, predicate,            │
│ END_USER_REF     │ 预设时  │               check_scope, check }                   │
└──────────────────┘         │     delete:  { allowed, scope, predicate } }         │
                             │ preset        NULL | READ_WRITE_ALL | ...            │
                             │                                                      │
                             │ UNIQUE (model_id, type, name)                        │
                             │ INDEX  (model_id, preset)                            │
                             └────────────────────┬────────────────────────────────┘
                                                   │ ON DELETE CASCADE
                                                   │ 1 : N
                             ┌─────────────────────▼──────────────────────────────┐
                             │  end_user_bundle_permissions  (中间表)              │
                             │ bundle_id → end_user_permission_bundles              │
                             │ permission_id → end_user_permissions  CASCADE        │
                             └──────────────────────────────────────────────────── ┘
                                                   │ N : 1
                             ┌─────────────────────▼──────────────────────────────┐
                             │  end_user_permission_bundles  (权限包)              │
                             └──────────────────┬───────────────────┬─────────────┘
                                                │                   │
                              ┌─────────────────▼──┐  ┌────────────▼────────────┐
                              │ end_user_role_bundles│  │ end_user_user_bundles  │
                              │  role_id → roles    │  │  user_id → end_users   │
                              └────────────────────┘  └────────────────────────┘
```

## `applyEndUserPresetPolicy` 操作涉及的 SQL

```
applyEndUserPresetPolicy(orgName, projectSlug, modelId, preset)
                │
    ┌───────────┼──────────────────┐
    │           │                  │
 [校验只读]  [步骤1] DELETE      [步骤2] INSERT × 1
    │           │                  │
models      end_user_permissions  end_user_permissions
fields(若   WHERE model_id = ?    INSERT 1 行：
 *_OWNER)     AND org_name = ?      type    = 'PRESET'
              AND type = 'PRESET'   preset  = <预设枚举>
                │                   name    = 'preset:<PRESET>'
                │ ON DELETE CASCADE  row_policy = <展开值>
                ▼                   column_policy = NULL (全可见)
  end_user_bundle_permissions
  （引用旧权限点的中间行被级联删除）
       ⚠️ 权限包静默缩减风险（可接受）
```

**不涉及的表：**
- `end_user_permission_bundles` — 只操作裸权限点
- `end_user_role_bundles` / `end_user_user_bundles` — 不分配给角色或用户
- `model_rls_policies` — 独立管理，不关联

## Decisions

### D1：`CreatePermission` 统一写 type / row_policy / preset

**决策**：修改 `CreateEndUserPermission` SQL 查询，加入 `type`、`row_policy`、`preset` 字段，同时移除 `action`、`row_scope`；不拆分接口。

**理由**：统一接口避免 repo 层割裂；字段变更集中在 SQL 和 sqlc 生成，上层同步更新映射。

---

### D2：删除旧 preset 权限点用批量 DELETE，按 type='PRESET' 过滤

**决策**：新增 `DeleteEndUserPermissionsByModelAndType` 查询，按 `(model_id, org_name, type='PRESET')` 批量删除。

**理由**：`type` 字段语义清晰；原子性强，无需先查再删；`idx_permissions_model_preset` 索引可优化查询路径（preset IS NOT NULL = type=PRESET）。

---

### D3：预设展开逻辑（row_policy 生成）在 App 层硬编码

**决策**：4 种预设的 `row_policy` 固定值定义在 `internal/app/rbac/permission_app.go`。Domain 层只持有 `Preset` 常量和 `RowPolicy` 值对象，不感知展开逻辑。

**理由**：预设是业务策略，属于应用层协调职责；固定值无需 DB 存储配置，App 层 map 即可。

---

### D4：`*_OWNER` 预设在 apply 前校验 END_USER_REF 字段

**决策**：`ApplyPresetPolicy` 展开含 owner 行策略的预设前，调用现有 `validateEndUserRefField(ctx, modelID)` 校验。GraphQL 层将 `EndUserRowScopeFieldMissing` 错误转换为 `PresetRequiresOwnerField` 返回前端。

---

### D5：preset 权限点 name 格式固定，每种预设 1 条记录

**决策**：`name = "preset:<PRESET_NAME>"`（如 `"preset:READ_WRITE_ALL"`）。每次 apply 后该模型最多有 4 条 PRESET 记录（每种预设 1 条）。

**理由**：不再按 action 拆多行，row_policy 五件套一次性表达所有行为；name 唯一稳定，规避 DB unique constraint 冲突。

---

### D6：删除 action / row_scope，引入 type / row_policy

**决策**：
- 删除 `action`（ENUM: select/insert/update/delete/export）
- 删除 `row_scope`（ENUM: ALL/SELF/DEPT/DEPT_AND_CHILDREN）
- 新增 `type` ENUM(PRESET/CUSTOM)
- 新增 `row_policy` JSON（存储 Runtime where 条件五件套）

**理由**：
- `action` + `row_scope` 是对 Runtime 行为的二次抽象，最终仍需展开为五件套谓词才能应用到 RLS 引擎；中间层徒增转换逻辑
- `row_policy` 直接存储 Runtime where 条件 JSON，格式与 `model_rls_policies` 对齐，消除阻抗失配
- `type` 字段语义明确，比 `preset IS NOT NULL` 的隐式判断更可靠

**影响：**
- DB：`13_rbac_permissions.sql` 已更新终态，`preset` 列已合并入 CREATE TABLE（`14_rbac_preset.sql` 已删除）
- sqlc：`CreateEndUserPermission` / `UpdateEndUserPermission` / `ListEndUserPermissionsByModel` 签名全部更新
- Domain：`EndUserPermission` struct 删除 Action/RowScope，新增 Type/RowPolicy
- App：删除 action×row_scope → row_policy 展开逻辑，改为直接传递 row_policy 值

## Risks / Trade-offs

- **[风险] preset 权限点被加入权限包后被删除**：apply 替换旧 PRESET 权限点时，若该权限点已被某权限包引用，FK CASCADE 自动删除 `end_user_bundle_permissions` 关联行，权限包静默缩减。→ 缓解：目前认为可接受，未来可在删除前检查并返回警告。
- **[风险] sqlc 重新生成影响已有代码**：修改 `CreateEndUserPermission` SQL 签名会导致 sqlc 重新生成，所有调用方需同步更新。→ 缓解：改动集中，可控。
- **[风险] action/row_scope 删除是破坏性变更**：存量 CUSTOM 权限点若有数据需迁移（action+row_scope → row_policy）。→ 缓解：当前无生产数据，`just db up` 直接应用终态 schema 即可。

## Migration Plan

1. 运行 `just db up`（Atlas 对比终态 schema，自动生成并应用 DDL diff）
2. 修改 SQL 查询文件（`db/queries/rbac/permission.sql`）→ 运行 `just generate-safe-querier`
3. 按层实现：Domain → Infra → App → Adapter → Resolver
4. 运行 `just lint` 验证编译

## Unit Tests

`row_policy` 的解析和校验逻辑较细，需要完整的单元测试覆盖。测试位置：`internal/domain/rbac/row_policy_test.go`。

### RowPolicy 值对象解析

```
TestRowPolicy_Validate
├── select.allowed=false 时，其余 action.allowed 必须全为 false（隐藏规则）
├── select.allowed=false 且其他 action 存在 allowed=true → 报错
├── scope=custom 但 predicate 为空 → 报错
├── scope=all 时 predicate 非空 → 忽略（不报错，但序列化时清除）
├── update.check_scope=custom 但 check 为空 → 报错
└── update.allowed=false 时 check_scope / check 必须被忽略

TestRowPolicy_Normalization
├── scope=all 时序列化不写入 predicate 字段
├── allowed=false 时序列化只写 allowed=false，其余字段省略
└── update.check_scope 默认值为 "all"（未显式指定时）
```

### 预设展开

```
TestExpandPreset
├── READ_WRITE_ALL     → select/insert/update/delete 全为 allowed=true, scope=all
├── READ_ALL           → select allowed=true scope=all；其余 allowed=false
├── READ_WRITE_OWNER   → 所有 action allowed=true, scope=custom, predicate/check={owner=$eu}
│                        → 需传入 ownerFieldName，注入到谓词
├── READ_ALL_WRITE_OWNER → select all；insert/update/delete custom owner
└── 所有预设展开结果通过 RowPolicy.Validate() 无错误

TestExpandPreset_OwnerField
├── *_OWNER 预设且 ownerFieldName="" → 返回 ErrOwnerFieldRequired
└── 非 *_OWNER 预设传入 ownerFieldName → 忽略（不影响结果）
```

### ApplyPresetPolicy App 层逻辑

```
TestApplyPresetPolicy
├── 正常 apply READ_WRITE_ALL
│   → 删除旧 PRESET 权限点（type=PRESET）
│   → 插入 1 条新记录，name="preset:READ_WRITE_ALL"
├── 正常 apply READ_ALL（无 END_USER_REF 字段也可以）
├── apply READ_WRITE_OWNER，模型有 END_USER_REF → 成功
├── apply READ_WRITE_OWNER，模型无 END_USER_REF → 返回 PresetRequiresOwnerField 错误
├── apply 后 CUSTOM 权限点不受影响（type=CUSTOM 的记录保留）
└── 重复 apply 同一预设 → 幂等（删旧插新，结果一致）
```

### RowPolicy 隐藏规则校验

```
TestRowPolicy_HiddenRule
├── select.allowed=false, insert.allowed=true  → 校验失败
├── select.allowed=false, insert.allowed=false → 校验通过
├── select.allowed=false, update.allowed=false → 校验通过
└── select.allowed=true  → 其他 action 不受限制
```

### column_policy writable 字段

```
TestColumnPolicy_Writable
├── writable 默认为 true（不写时）
├── mode=HIDDEN 时 writable=true → 无意义但不报错
├── mode=MASKED 时 writable=true → 无意义但不报错
└── mode=VISIBLE, writable=false → 正常，列可见但不可 SET
```
