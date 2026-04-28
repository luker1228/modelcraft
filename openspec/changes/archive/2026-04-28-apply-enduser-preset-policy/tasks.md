## 并行开发策略

使用 multi-agent 并行开发，3 条独立 worker 线同时推进，最终由 backend-reviewer 审计。

```
backend-worker-1          backend-worker-2          backend-worker-3
─────────────────         ─────────────────         ─────────────────
数据层 + Domain 层         App 层                    Infra 层 + Adapter + Resolver
       ↓                        ↓                          ↓
              ──────────────────────────────────────────────
                              backend-reviewer
                           （编译 + lint + 单元测试审计）
```

**依赖规则：**
- Worker-2 和 Worker-3 依赖 Worker-1 完成 Domain 接口定义后才能实现
- Worker-1 可最先独立启动（SQL 查询 → 生成 → Domain）

---

## Worker-1：数据层 + Domain 层

### 1. SQL 查询更新

- [x] 1.1 修改 `db/queries/rbac/permission.sql`：
  - `CreateEndUserPermission`：替换 `action`/`row_scope` 为 `type`/`row_policy`/`preset`（参数和 VALUES 同步更新）
  - 新增 `DeleteEndUserPermissionsByModelAndType`：按 `(model_id, org_name, type='PRESET')` 批量删除
  - `ListEndUserPermissionsByModel` / `GetEndUserPermissionByID`：SELECT * 自动包含新字段，无需改动
  - `UpdateEndUserPermission`：确认 SET 子句不包含 `action`/`row_scope`（已无该字段）

- [x] 1.2 运行 `just generate-safe-querier` 重新生成 `internal/infrastructure/dbgen/` 代码

### 2. Domain 层

- [x] 2.1 `internal/domain/rbac/permission.go`：
  - `EndUserPermission` struct：删除 `Action`/`RowScope` 字段，新增 `Type PermissionType`、`RowPolicy *RowPolicy`、`Preset *PermissionPreset`
  - 定义 `PermissionType` 类型：`PermissionTypePreset = "PRESET"`、`PermissionTypeCustom = "CUSTOM"`
  - 定义 4 个预设常量：`PresetReadWriteAll`、`PresetReadAll`、`PresetReadWriteOwner`、`PresetReadAllWriteOwner`

- [x] 2.2 新建 `internal/domain/rbac/row_policy.go`：
  - 定义 `RowPolicy` 值对象及子结构（`SelectPolicy`、`InsertPolicy`、`UpdatePolicy`、`DeletePolicy`）
  - 每个 action 含 `Allowed bool`、`Scope PolicyScope`（`ScopeAll`|`ScopeCustom`）、`Predicate`/`Check` JSON
  - `UpdatePolicy` 额外含 `CheckScope PolicyScope`、`Check` JSON
  - 实现 `Validate()` 方法：
    - `select.allowed=false` 时其余 action 必须全为 `allowed=false`（隐藏规则）
    - `scope=custom` 时 predicate/check 不能为空
    - `update.check_scope=custom` 时 check 不能为空
  - 实现 `Normalize()` 方法：`scope=all` 时清除 predicate/check；`allowed=false` 时清除其余字段

- [x] 2.3 `internal/domain/rbac/repository.go`：
  - 新增接口方法 `DeletePresetPermissionsByModel(ctx context.Context, orgName, modelID string) error`

### 3. 单元测试

- [x] 3.1 新建 `internal/domain/rbac/row_policy_test.go`，覆盖 design.md §Unit Tests 中所有用例：
  - `TestRowPolicy_Validate`（6 个子用例）
  - `TestRowPolicy_Normalization`（3 个子用例）
  - `TestExpandPreset`（5 个子用例）
  - `TestExpandPreset_OwnerField`（2 个子用例）
  - `TestRowPolicy_HiddenRule`（4 个子用例）
  - `TestColumnPolicy_Writable`（4 个子用例）

---

## Worker-2：App 层（依赖 Worker-1 完成 Domain 接口）

- [x] 4.1 `pkg/bizerrors/common_errors.go`：RBAC 错误块末尾新增 `EndUserPresetRequiresOwnerField` 错误定义

- [x] 4.2 `internal/app/rbac/commands.go`：新增 `ApplyPresetPolicyCommand` struct：
  ```go
  type ApplyPresetPolicyCommand struct {
      ProjectScope ProjectScope
      ModelID      string
      Preset       rbac.PermissionPreset
  }
  ```

- [x] 4.3 `internal/app/rbac/permission_app.go`：实现 `ApplyPresetPolicy(ctx, cmd ApplyPresetPolicyCommand)` 方法：
  1. 校验 model 存在
  2. 若 preset 为 `*_OWNER` 类，调用 `validateEndUserRefField` 校验 END_USER_REF 字段存在；缺失时返回 `EndUserPresetRequiresOwnerField`
  3. 调用 `repo.DeletePresetPermissionsByModel(ctx, orgName, modelID)`
  4. 调用 `expandPreset(preset, ownerFieldName)` 生成 `RowPolicy`（硬编码 4 种预设展开，见 design.md 展开表）
  5. 调用 `repo.CreatePermission(ctx, ...)` 插入新权限点（type=PRESET，name=`"preset:<PRESET_NAME>"`）
  6. 返回该 model 下所有权限点

- [x] 4.4 `internal/app/rbac/permission_app.go`：实现私有函数 `expandPreset(preset PermissionPreset, ownerField string) (*RowPolicy, error)`：
  - 4 种预设对应的 `RowPolicy` 固定值（见 design.md §4 种预设展开表）
  - `*_OWNER` 类将 `ownerField` 注入谓词的字段名

### 单元测试

- [x] 4.5 新建 `internal/app/rbac/apply_preset_policy_test.go`，覆盖 design.md §ApplyPresetPolicy App 层逻辑 所有用例（6 个子用例），使用 mock repo

---

## Worker-3：Infra 层 + Adapter 层 + Resolver 层（依赖 Worker-1 完成 sqlc 生成）

### Infrastructure 层

- [x] 5.1 `internal/infrastructure/repository/sql_end_user_permission_repository.go`：
  - `toDomainPermission`：映射 `Type`、`RowPolicy`（JSON 反序列化为 `*domain.RowPolicy`）、`Preset`（NullString → `*domain.PermissionPreset`）；删除 `Action`/`RowScope` 映射
  - `CreatePermission`：更新 params 加入 `Type`、`RowPolicy`（JSON 序列化）、`Preset`；删除 `Action`/`RowScope`
  - 实现 `DeletePresetPermissionsByModel` 方法，调用新 sqlc 查询 `DeleteEndUserPermissionsByModelAndType`

### Adapter 层

- [x] 6.1 `internal/interfaces/graphql/project/adapter/rbac_mapper.go`：
  - `ToEndUserPermissionDTO`：映射 `Type`、`RowPolicy`（序列化为 GraphQL 类型）、`Preset`（`*domain.PermissionPreset` → `*generated.EndUserPermissionPreset`）；删除 `Action`/`RowScope` 映射

- [x] 6.2 `internal/interfaces/graphql/project/adapter/rbac_error_adapter.go`：
  - 新增 `ConvertToApplyPresetPolicyError` 函数，处理以下错误映射：
    - `EndUserPresetRequiresOwnerField` → `generated.PresetRequiresOwnerField`
    - `EndUserRowScopeFieldMissing` → `generated.PresetRequiresOwnerField`
    - `ModelNotFound` → `generated.ModelNotFound`
    - `ProjectNotFound` → `generated.ProjectNotFound`

### Resolver 层

- [x] 7.1 `internal/interfaces/graphql/project/rbac.resolvers.go`：
  - 替换 `ApplyEndUserPresetPolicy` panic stub
  - 调用 `permissionSvc.ApplyPresetPolicy(ctx, cmd)` 并通过 adapter 转换返回值

---

## Reviewer：审计（依赖所有 Worker 完成）

- [x] 8.1 运行 `just lint` 确保整条链路编译无误
- [x] 8.2 运行单元测试：`just test-unit-pkg internal/domain/rbac` 和 `just test-unit-pkg internal/app/rbac`
- [x] 8.3 审计 `row_policy` 序列化/反序列化路径（Domain ↔ Infra ↔ GraphQL）的类型一致性
- [x] 8.4 审计隐藏规则（`select.allowed=false` 级联）是否在 `Validate()` 和 App 层均有校验
- [x] 8.5 审计 `DeletePresetPermissionsByModel` 的 CASCADE 副作用是否在日志中有记录
