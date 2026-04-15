🗂️  Checklist Review — `modelcraft-backend/internal/infrastructure/repository/sql_modelruntime_repository.go`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

共 1 条规则（来自错题本 `ai-metadata/backend/common-mistakes.md`）

---

## 规则清单

**[BM-20260415-0001]** 凡是查询带 `project_slug` 的 SQL，必须同时带 `org_name`。
适用表：`model_enums`、`field_definitions`、`models`、`logical_foreign_keys`、`model_field_enum_associations` 等所有 project 域资源表。

---

## 审查结果

❌ 命中 2 处：

### 命中 #1 — `GetFieldsByModelID` 仅按 model_id 过滤，缺少 org_name

- **位置**：`sql_modelruntime_repository.go:68`（调用处）；底层 SQL：`dbgen/field.sql.go`
- **问题**：`GetFieldsByModelID` 的 SQL 为：
  ```sql
  SELECT ... FROM field_definitions
  WHERE model_id = ?
  ORDER BY display_order ASC
  ```
  只用 `model_id` 过滤。`model_id` 是 UUID，在数据库层面唯一性依赖于生成策略，但这里没有 `org_name` 作为额外保障。虽然 `model_id` 在设计上全局唯一，但 `field_definitions` 是 project 域资源表，按 checklist 规则其 `project_slug` 查询必须带 `org_name`。更实质的问题是：该 SQL 没有同时过滤 `project_slug`/`org_name`，若 `model_id` 因任何原因（脏数据、bug）出现碰撞，将造成跨 org 数据污染。
- **参考案例**：BM-20260415-0001

### 命中 #2 — `GetModelByName` 带 `project_slug` 但缺少 `org_name`

- **位置**：`sql_modelruntime_repository.go:44-48`（调用处）；底层 SQL：`dbgen/model.sql.go`
- **问题**：`GetModelByName` 的 SQL 为：
  ```sql
  SELECT ... FROM models
  WHERE database_name = ? AND name = ? AND project_slug = ?
  LIMIT 1
  ```
  使用了 `project_slug` 但**缺少 `org_name`**。`project_slug` 在不同 org 间可以重名，`database_name` 也可能在不同 org 中相同（如都叫 `dev`、`main`）。这是 BM-20260415-0001 描述的完全相同模式：`project_slug` 不是全局唯一键，必须与 `org_name` 联合使用才能确保租户隔离。
  
  调用处（`sql_modelruntime_repository.go:44-48`）传入的 `ModelLocator` 中有 `ProjectSlug`，但没有 `OrgName`：
  ```go
  row, err := r.q.GetModelByName(ctx, dbgen.GetModelByNameParams{
      DatabaseName: modelLocator.DatabaseName,
      Name:         modelLocator.ModelName,
      ProjectSlug:  modelLocator.ProjectSlug,
  })
  ```
  `ModelLocator` 是否包含 `OrgName` 字段需确认；如果有，没有传入即为 bug。
- **参考案例**：BM-20260415-0001

---

✅ 通过 0 条（无完全通过项）

> **注意**：`GetEnumsByNames` 调用已正确传入 `OrgName`（`sql_modelruntime_repository.go:84-88`），且底层 SQL 包含 `WHERE org_name = ? AND project_slug = ?`，这一处**符合规则**。

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## 修改建议

### 针对命中 #2（优先级 CRITICAL）

确认 `modeldesign.ModelLocator` 结构体中是否包含 `OrgName` 字段。如果有，立即修改调用处：

```go
// ❌ 当前写法（缺少 org_name）
row, err := r.q.GetModelByName(ctx, dbgen.GetModelByNameParams{
    DatabaseName: modelLocator.DatabaseName,
    Name:         modelLocator.ModelName,
    ProjectSlug:  modelLocator.ProjectSlug,
})

// ✅ 修正写法（必须带 org_name）
row, err := r.q.GetModelByName(ctx, dbgen.GetModelByNameParams{
    OrgName:      modelLocator.OrgName,
    DatabaseName: modelLocator.DatabaseName,
    Name:         modelLocator.ModelName,
    ProjectSlug:  modelLocator.ProjectSlug,
})
```

同时修改 `GetModelByNameParams` struct 和底层 SQL：

```sql
-- ✅ 修正后的 SQL
SELECT ... FROM models
WHERE org_name = ? AND database_name = ? AND name = ? AND project_slug = ?
LIMIT 1
```

### 针对命中 #1（优先级 HIGH）

`GetFieldsByModelID` 使用 UUID `model_id`，全局唯一性风险相对较低，但建议：
1. 在代码注释中明确说明 `model_id` 是 UUID，全局唯一，不需要 org_name 二次过滤；
2. 或者增加 `org_name` + `project_slug` 复合过滤作为防御性保障，避免脏数据扩散。

---

## 附：审查范围说明

本次 review 扫描了以下调用链：

| 调用点 | 底层 SQL | `org_name` 是否在 WHERE |
|---|---|---|
| `GetModelByID`（L25） | `WHERE id = ?` | 不适用（id 为全局唯一 PK）|
| `GetModelByName`（L44） | `WHERE database_name=? AND name=? AND project_slug=?` | **❌ 缺失** |
| `GetFieldsByModelID`（L68） | `WHERE model_id = ?` | **⚠️ 未使用 project_slug，依赖 UUID 唯一性** |
| `GetEnumsByNames`（L84） | `WHERE org_name=? AND project_slug=? AND name IN(...)` | **✅ 正确** |
