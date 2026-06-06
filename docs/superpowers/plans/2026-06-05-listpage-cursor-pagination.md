# ListPage Cursor Pagination Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 Runtime GraphQL 新增稳定 cursor 分页接口 `listPage`，并在管理态模型配置中加入"插入序字段"配置项（未配置时显示风险提示，对齐 `displayField` 模式）。

**Architecture:** 
`listPage` 始终生成到 Runtime GraphQL schema，resolver 读取模型的 `InsertionOrderField` 配置决定 cursor 策略：有配置则使用双字段 cursor（sortField + insertionOrderField）确保稳定性，无配置则使用单字段 cursor（sortField 必须唯一）。cursor 对消费者透明，编码为 base64 JSON。管理态复用 `displayField` 的更新链路（GraphQL mutation → App Service → Domain → DB），增加 `insertionOrderField` 字段。

**Tech Stack:** Go (graphql-go, goqu), MySQL, gqlgen (管理态 GraphQL), Next.js/React (前端配置 UI)

---

## 文件变更地图

### 后端 — 数据层

| 文件 | 操作 | 职责 |
|------|------|------|
| `db/schema/mysql/03_model_domain.sql` | 修改 | 加 `insertion_order_field VARCHAR(64) NULL` 列 |
| `db/queries/model.sql` | 修改 | `GetModelByID`、`GetModelByName` 加新列；加 `UpdateInsertionOrderField` query |
| `internal/infrastructure/dbgen/models.go` | 生成 | sqlc 生成，含新字段 |
| `internal/infrastructure/dbgen/model.sql.go` | 生成 | sqlc 生成新 query |

### 后端 — 领域 & 应用层

| 文件 | 操作 | 职责 |
|------|------|------|
| `internal/domain/modeldesign/model.go` | 修改 | `DataModel` 加 `InsertionOrderField *string`；加 `UpdateInsertionOrderField`、`ValidateInsertionOrderField` 方法 |
| `internal/app/modeldesign/commands.go` | 修改 | `UpdateModelMetaCommand` 加 `InsertionOrderField *string` |
| `internal/app/modeldesign/model_app.go` | 修改 | `UpdateModelMeta` 处理 `InsertionOrderField` 更新 |
| `internal/infrastructure/repository/sql_model_repository.go` | 修改 | 调用 `UpdateInsertionOrderField` sqlc query |
| `internal/infrastructure/repository/sql_modelruntime_repository.go` | 修改 | `DbgenModelToRuntimeModel` 映射新字段 |

### 后端 — Runtime modelruntime

| 文件 | 操作 | 职责 |
|------|------|------|
| `internal/domain/modelruntime/runtimemodel.go` | 修改 | `RuntimeModel` 加 `InsertionOrderField *string` |
| `internal/domain/modelruntime/graphql_constants.go` | 修改 | 加 `OperationListPage`、`FieldAfter`、`FieldNextCursor`、`FieldHasNextPage` 常量 |
| `internal/domain/modelruntime/cursor.go` | 新建 | cursor 编解码逻辑（base64 JSON）；`encodeCursor`、`decodeCursor` |
| `internal/domain/modelruntime/cursor_test.go` | 新建 | cursor 编解码单元测试 |
| `internal/domain/modelruntime/graphql_input.go` | 修改 | 加 `ListPageInput` struct；加 `newListPageInput` 函数 |
| `internal/domain/modelruntime/graphql_input_types.go` | 修改 | 加 `GenerateListPageArgs` 方法（单字段 orderBy + after） |
| `internal/domain/modelruntime/model_resolver.go` | 修改 | 加 `createListPageField`、`executeListPage`、`createListPageResultType` 方法；`createRootQuery` 注册 `listPage` |
| `internal/infrastructure/database/dml/sql_mapper.go` | 修改 | 加 `convertListPageInputToSQL` 函数，构造 keyset WHERE |
| `internal/infrastructure/database/dml/client_db_repo_impl.go` | 修改 | 加 `ListPage` 方法到 `ClientDBRepoImpl` |
| `internal/domain/modelruntime/graphql_repository.go` | 修改 | `ClientDatabaseRepository` 接口加 `ListPage` 方法 |

### 后端 — 管理态 GraphQL Schema

| 文件 | 操作 | 职责 |
|------|------|------|
| `api/graph/project/schema/model.graphql` | 修改 | `Model` 类型加 `insertionOrderField`；`CreateModelInput`、`UpdateModelMetaInput` 加 `insertionOrderField` |
| `internal/interfaces/graphql/project/generated/` | 生成 | `just generate-gql` 生成 |
| `internal/interfaces/graphql/project/model.resolvers.go` | 修改 | `UpdateModelMeta` resolver 传递 `InsertionOrderField` |

### 前端 — 管理态配置 UI

| 文件 | 操作 | 职责 |
|------|------|------|
| `modelcraft-front/contract/` | 同步 | `front-contract-pull` 同步新 schema |
| `modelcraft-front/src/generated/graphql.ts` | 生成 | codegen 生成 |
| `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model/[modelId]/settings/_components/ModelMetaForm.tsx` (或对应配置组件) | 修改 | 加 `insertionOrderField` 选择器 + 风险提示（对齐 `displayField` 模式） |

---

## Task 1: DB Schema — 加 `insertion_order_field` 列

**Files:**
- Modify: `modelcraft-backend/db/schema/mysql/03_model_domain.sql`

- [ ] **Step 1: 在 models 表加列**

打开 `db/schema/mysql/03_model_domain.sql`，找到 `display_field` 列，紧跟其后加一行：

```sql
  `insertion_order_field` VARCHAR(64) NULL COMMENT '用于 listPage cursor 分页的插入序字段名（单调递增字段，如 created_at）',
```

完整上下文（在 `display_field` 行后面）：
```sql
  `display_field` VARCHAR(64) NULL COMMENT '用于 runtime _displayName 解析的字段名（必须是模型中存在且可字符串化的字段）',
  `insertion_order_field` VARCHAR(64) NULL COMMENT '用于 listPage cursor 分页的插入序字段名（单调递增字段，如 created_at）',
```

- [ ] **Step 2: 同步 DB**

```bash
cd modelcraft-backend
just db up
```

预期输出：`Schema applied`（或 Atlas diff 显示加了 1 列）

- [ ] **Step 3: Commit**

```bash
git add db/schema/mysql/03_model_domain.sql
git commit -m "feat(db): add insertion_order_field column to models table"
```

---

## Task 2: sqlc — 更新 queries 并重新生成

**Files:**
- Modify: `modelcraft-backend/db/queries/model.sql`
- Generate: `modelcraft-backend/internal/infrastructure/dbgen/`

- [ ] **Step 1: 在 `db/queries/model.sql` 中更新 GetModelByID、GetModelByName**

找到 `GetModelByID` 和 `GetModelByName` 的 SELECT 语句，在 `display_field` 后面加 `insertion_order_field`。示例（两个 query 均需修改）：

```sql
-- 在 SELECT 列表中，display_field 后加：
, m.insertion_order_field
```

- [ ] **Step 2: 在 `db/queries/model.sql` 中加 UpdateInsertionOrderField query**

在文件末尾追加：

```sql
-- name: UpdateInsertionOrderField :exec
UPDATE models
SET insertion_order_field = sqlc.narg('insertion_order_field'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
  AND delete_token = 0;
```

- [ ] **Step 3: 重新生成 sqlc**

```bash
cd modelcraft-backend
just generate-sqlc
```

预期：`internal/infrastructure/dbgen/models.go` 的 `Model` struct 加了 `InsertionOrderField sql.NullString` 字段；`internal/infrastructure/dbgen/model.sql.go` 加了 `UpdateInsertionOrderField` 方法。

- [ ] **Step 4: Commit**

```bash
git add db/queries/model.sql internal/infrastructure/dbgen/
git commit -m "feat(sqlc): add insertion_order_field to model queries"
```

---

## Task 3: Domain — DataModel & RuntimeModel 加新字段

**Files:**
- Modify: `modelcraft-backend/internal/domain/modeldesign/model.go`
- Modify: `modelcraft-backend/internal/domain/modelruntime/runtimemodel.go`
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_modelruntime_repository.go`

- [ ] **Step 1: 在 `DataModel` struct 加字段**

打开 `internal/domain/modeldesign/model.go`，找到 `DataModel` struct（约 89 行），在 `DisplayField *string` 后加：

```go
InsertionOrderField *string `json:"insertionOrderField"` // 用于 listPage cursor 分页的插入序字段名
```

- [ ] **Step 2: 加 UpdateInsertionOrderField 方法**

在 `UpdateDisplayField` 方法（约 215 行）后面加：

```go
// UpdateInsertionOrderField 更新 insertionOrderField
func (m *DataModel) UpdateInsertionOrderField(field *string) {
    m.InsertionOrderField = field
}

// ValidateInsertionOrderField 验证 insertionOrderField 是否有效（字段名必须存在于模型字段中）
func (m *DataModel) ValidateInsertionOrderField() error {
    if m.InsertionOrderField == nil || *m.InsertionOrderField == "" {
        return nil // nil 表示未配置，允许
    }
    for _, f := range m.Fields {
        if f.Name == *m.InsertionOrderField {
            return nil
        }
    }
    return bizerrors.Errorf("insertionOrderField %q 不存在于模型字段中", *m.InsertionOrderField)
}
```

- [ ] **Step 3: 在 RuntimeModel 加字段**

打开 `internal/domain/modelruntime/runtimemodel.go`，在 `DisplayField *string` 后加：

```go
InsertionOrderField *string `json:"insertionOrderField"` // 用于 listPage cursor 的插入序字段名
```

- [ ] **Step 4: 更新 DbgenModelToRuntimeModel**

打开 `internal/infrastructure/repository/sql_modelruntime_repository.go`，找到 `DbgenModelToRuntimeModel` 函数（约 124 行），在 `DisplayField: displayField,` 后加：

```go
var insertionOrderField *string
if row.InsertionOrderField.Valid && row.InsertionOrderField.String != "" {
    insertionOrderField = &row.InsertionOrderField.String
}
```

并在 return 的 struct 里加：

```go
InsertionOrderField: insertionOrderField,
```

- [ ] **Step 5: 编译验证**

```bash
cd modelcraft-backend
go build ./...
```

预期：0 errors

- [ ] **Step 6: Commit**

```bash
git add internal/domain/modeldesign/model.go \
        internal/domain/modelruntime/runtimemodel.go \
        internal/infrastructure/repository/sql_modelruntime_repository.go
git commit -m "feat(domain): add InsertionOrderField to DataModel and RuntimeModel"
```

---

## Task 4: App & Repository — UpdateModelMeta 处理新字段

**Files:**
- Modify: `modelcraft-backend/internal/app/modeldesign/commands.go`
- Modify: `modelcraft-backend/internal/app/modeldesign/model_app.go`
- Modify: `modelcraft-backend/internal/infrastructure/repository/` (sql_model_repository.go)

- [ ] **Step 1: 更新 UpdateModelMetaCommand**

打开 `internal/app/modeldesign/commands.go`，在 `UpdateModelMetaCommand` struct 的 `DisplayField *string` 后加：

```go
InsertionOrderField *string // 用于 listPage cursor 的插入序字段名（nil 表示不更新，空字符串表示清除）
```

- [ ] **Step 2: 更新 model_app.go 的 UpdateModelMeta**

打开 `internal/app/modeldesign/model_app.go`，找到处理 `DisplayField` 更新的代码段（通常是 `if cmd.DisplayField != nil { model.UpdateDisplayField(...) }`），在其后仿照加：

```go
if cmd.InsertionOrderField != nil {
    model.UpdateInsertionOrderField(cmd.InsertionOrderField)
    if err := model.ValidateInsertionOrderField(); err != nil {
        return bizerrors.InvalidInput(err.Error())
    }
}
```

- [ ] **Step 3: 更新 Repository 的 Update 方法**

找到 `internal/infrastructure/repository/` 中 model repository 的 update 逻辑，在处理 `DisplayField` 的地方仿照加对 `InsertionOrderField` 的更新调用（调用 `r.q.UpdateInsertionOrderField`）：

```go
if cmd.InsertionOrderField != nil {
    if err := r.q.UpdateInsertionOrderField(ctx, dbgen.UpdateInsertionOrderFieldParams{
        ID:                   id,
        InsertionOrderField:  dbgenwrap.ToNullString(cmd.InsertionOrderField),
    }); err != nil {
        return bizerrors.Wrapf(err, "UpdateInsertionOrderField")
    }
}
```

- [ ] **Step 4: 编译验证**

```bash
cd modelcraft-backend
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add internal/app/modeldesign/ \
        internal/infrastructure/repository/
git commit -m "feat(app): UpdateModelMeta supports InsertionOrderField"
```

---

## Task 5: 管理态 GraphQL Schema — 加 insertionOrderField

**Files:**
- Modify: `modelcraft-backend/api/graph/project/schema/model.graphql`
- Generate: `modelcraft-backend/internal/interfaces/graphql/project/generated/`
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/model.resolvers.go`

- [ ] **Step 1: 更新 model.graphql**

打开 `api/graph/project/schema/model.graphql`：

1. 在 `Model` type 的 `displayField: String` 后加：
```graphql
  insertionOrderField: String  # 用于 listPage cursor 分页的插入序字段名（未配置时 listPage 稳定性无法保证）
```

2. 在 `CreateModelInput` input 的 `displayField: String` 后加：
```graphql
  insertionOrderField: String  # 用于 listPage cursor 分页的插入序字段名
```

3. 在 `UpdateModelMetaInput` input 的 `displayField: String` 后加：
```graphql
  insertionOrderField: String  # 用于 listPage cursor 分页的插入序字段名
```

- [ ] **Step 2: 生成 gqlgen 代码**

```bash
cd modelcraft-backend
just generate-gql
```

预期：`internal/interfaces/graphql/project/generated/model_gen.go` 中 `Model` struct 加了 `InsertionOrderField *string`；`UpdateModelMetaInput` 加了 `InsertionOrderField *string`。

- [ ] **Step 3: 更新 model.resolvers.go 的 UpdateModelMeta**

打开 `internal/interfaces/graphql/project/model.resolvers.go`，找到构建 `UpdateModelMetaCommand` 的地方（约 119 行），在 `DisplayField: input.DisplayField` 后加：

```go
InsertionOrderField: input.InsertionOrderField,
```

- [ ] **Step 4: 编译验证**

```bash
cd modelcraft-backend
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add api/graph/project/schema/model.graphql \
        internal/interfaces/graphql/project/generated/ \
        internal/interfaces/graphql/project/model.resolvers.go
git commit -m "feat(gql): add insertionOrderField to Model schema and UpdateModelMeta"
```

---

## Task 6: cursor.go — cursor 编解码

**Files:**
- Create: `modelcraft-backend/internal/domain/modelruntime/cursor.go`
- Create: `modelcraft-backend/internal/domain/modelruntime/cursor_test.go`

- [ ] **Step 1: 写 cursor_test.go 的失败测试**

```go
package modelruntime

import (
    "testing"
)

func TestEncodeCursor_SingleField(t *testing.T) {
    c := cursorData{SortField: "price", SortValue: "100", IOField: "", IOValue: ""}
    encoded := encodeCursor(c)
    decoded, err := decodeCursor(encoded)
    if err != nil {
        t.Fatalf("decodeCursor error: %v", err)
    }
    if decoded.SortField != "price" || decoded.SortValue != "100" {
        t.Errorf("got %+v, want price=100", decoded)
    }
}

func TestEncodeCursor_DualField(t *testing.T) {
    c := cursorData{SortField: "price", SortValue: "100", IOField: "created_at", IOValue: "2026-06-05T10:00:00Z"}
    encoded := encodeCursor(c)
    decoded, err := decodeCursor(encoded)
    if err != nil {
        t.Fatalf("decodeCursor error: %v", err)
    }
    if decoded.IOField != "created_at" || decoded.IOValue != "2026-06-05T10:00:00Z" {
        t.Errorf("got %+v", decoded)
    }
}

func TestDecodeCursor_Invalid(t *testing.T) {
    _, err := decodeCursor("not-valid-base64!!")
    if err == nil {
        t.Error("expected error for invalid cursor")
    }
}
```

- [ ] **Step 2: 运行确认失败**

```bash
cd modelcraft-backend
go test ./internal/domain/modelruntime/ -run TestEncodeCursor -v
```

预期：`FAIL` — `encodeCursor undefined`

- [ ] **Step 3: 实现 cursor.go**

```go
package modelruntime

import (
    "encoding/base64"
    "encoding/json"

    bizerrors "modelcraft/pkg/bizerrors"
)

// cursorData cursor 的内部数据结构（对 API 消费者不可见）
type cursorData struct {
    SortField string `json:"sf"`           // 用户指定的排序字段名
    SortValue string `json:"sv"`           // 该字段的当前值（最后一条记录）
    IOField   string `json:"iof,omitempty"` // 插入序字段名（有配置时填充）
    IOValue   string `json:"iov,omitempty"` // 插入序字段值（最后一条记录）
}

// encodeCursor 将 cursorData 编码为 base64 字符串
func encodeCursor(c cursorData) string {
    b, _ := json.Marshal(c)
    return base64.RawURLEncoding.EncodeToString(b)
}

// decodeCursor 将 base64 字符串解码为 cursorData
func decodeCursor(s string) (cursorData, error) {
    b, err := base64.RawURLEncoding.DecodeString(s)
    if err != nil {
        return cursorData{}, bizerrors.Errorf("invalid cursor: %w", err)
    }
    var c cursorData
    if err := json.Unmarshal(b, &c); err != nil {
        return cursorData{}, bizerrors.Errorf("invalid cursor format: %w", err)
    }
    if c.SortField == "" {
        return cursorData{}, bizerrors.Errorf("invalid cursor: sortField is empty")
    }
    return c, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd modelcraft-backend
go test ./internal/domain/modelruntime/ -run TestEncodeCursor -v
go test ./internal/domain/modelruntime/ -run TestDecodeCursor -v
```

预期：`PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/domain/modelruntime/cursor.go \
        internal/domain/modelruntime/cursor_test.go
git commit -m "feat(modelruntime): cursor encode/decode for listPage"
```

---

## Task 7: SQL — convertListPageInputToSQL keyset WHERE

**Files:**
- Modify: `modelcraft-backend/internal/infrastructure/database/dml/sql_mapper.go`
- Create: `modelcraft-backend/internal/infrastructure/database/dml/sql_mapper_listpage_test.go`

- [ ] **Step 1: 加 ListPageInput struct 到 graphql_input.go**

打开 `internal/domain/modelruntime/graphql_input.go`，在 `FindManyInput` 的定义后加：

```go
// ListPageInput cursor 分页查询输入参数
type ListPageInput struct {
    TableName           string
    Selection           *Selection
    Where               map[string]any  // 额外 WHERE 过滤（RLS 等）
    SortField           string          // 必填，用于排序的字段
    SortDirection       string          // "asc" 或 "desc"
    InsertionOrderField string          // 插入序字段名（空表示无配置）
    After               *cursorData     // nil 表示第一页
    Limit               uint
}
```

- [ ] **Step 2: 写 sql_mapper_listpage_test.go 失败测试**

```go
package dml

import (
    "context"
    "testing"

    "modelcraft/internal/domain/modelruntime"
)

func TestConvertListPageInputToSQL_FirstPage_SingleField(t *testing.T) {
    input := &modelruntime.ListPageInput{
        TableName:     "products",
        SortField:     "price",
        SortDirection: "asc",
        Limit:         10,
    }
    sql, args, err := convertListPageInputToSQL(context.Background(), input)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if sql == "" {
        t.Error("expected non-empty SQL")
    }
    _ = args
    // 应包含 ORDER BY price ASC LIMIT 11（取 limit+1 判断 hasNextPage）
    t.Logf("sql: %s args: %v", sql, args)
}

func TestConvertListPageInputToSQL_AfterCursor_DualField(t *testing.T) {
    after := &modelruntime.CursorDataForTest{
        SortField: "price", SortValue: "100",
        IOField: "created_at", IOValue: "2026-06-05T10:00:00Z",
    }
    input := &modelruntime.ListPageInput{
        TableName:           "products",
        SortField:           "price",
        SortDirection:       "asc",
        InsertionOrderField: "created_at",
        After:               after.ToCursorData(),
        Limit:               10,
    }
    sql, args, err := convertListPageInputToSQL(context.Background(), input)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // 应包含 WHERE (price > ?) OR (price = ? AND created_at > ?)
    t.Logf("sql: %s args: %v", sql, args)
}
```

> **注意：** `CursorDataForTest` 是在 cursor.go 中暴露的测试辅助类型，在 step 3 中定义。

- [ ] **Step 3: 在 cursor.go 末尾加测试辅助类型**

```go
// CursorDataForTest 仅供测试使用，暴露内部 cursorData 构造
type CursorDataForTest struct {
    SortField, SortValue, IOField, IOValue string
}

func (c *CursorDataForTest) ToCursorData() *cursorData {
    return &cursorData{
        SortField: c.SortField,
        SortValue: c.SortValue,
        IOField:   c.IOField,
        IOValue:   c.IOValue,
    }
}
```

- [ ] **Step 4: 运行确认失败**

```bash
cd modelcraft-backend
go test ./internal/infrastructure/database/dml/ -run TestConvertListPage -v
```

预期：`FAIL` — `convertListPageInputToSQL undefined`

- [ ] **Step 5: 实现 convertListPageInputToSQL**

在 `internal/infrastructure/database/dml/sql_mapper.go` 末尾加：

```go
// convertListPageInputToSQL 将 ListPageInput 转换为 keyset cursor 分页 SQL
// 取 limit+1 条记录，调用方通过判断结果数 > limit 确定 hasNextPage。
func convertListPageInputToSQL(
    ctx context.Context,
    input *modelruntime.ListPageInput,
) (sql string, args []any, err error) {
    dialectWrapper := goqu.Dialect("mysql")
    ds := dialectWrapper.Select("*").From(input.TableName)

    // ORDER BY：有插入序字段则双排序，否则单排序
    if input.InsertionOrderField != "" {
        if input.SortDirection == modelruntime.OrderByDesc {
            ds = ds.Order(goqu.C(input.SortField).Desc(), goqu.C(input.InsertionOrderField).Desc())
        } else {
            ds = ds.Order(goqu.C(input.SortField).Asc(), goqu.C(input.InsertionOrderField).Asc())
        }
    } else {
        if input.SortDirection == modelruntime.OrderByDesc {
            ds = ds.Order(goqu.C(input.SortField).Desc())
        } else {
            ds = ds.Order(goqu.C(input.SortField).Asc())
        }
    }

    // LIMIT: 多取 1 条用于 hasNextPage 判断
    ds = ds.Limit(input.Limit + 1)

    // CURSOR WHERE（仅在 after != nil 时添加）
    if input.After != nil {
        after := input.After
        var cursorExpr goqu.Expression
        if input.InsertionOrderField != "" && after.IOField != "" {
            // 双字段: (sortField > sv) OR (sortField = sv AND ioField > iov)
            if input.SortDirection == modelruntime.OrderByDesc {
                cursorExpr = goqu.Or(
                    goqu.C(input.SortField).Lt(after.SortValue),
                    goqu.And(
                        goqu.C(input.SortField).Eq(after.SortValue),
                        goqu.C(after.IOField).Lt(after.IOValue),
                    ),
                )
            } else {
                cursorExpr = goqu.Or(
                    goqu.C(input.SortField).Gt(after.SortValue),
                    goqu.And(
                        goqu.C(input.SortField).Eq(after.SortValue),
                        goqu.C(after.IOField).Gt(after.IOValue),
                    ),
                )
            }
        } else {
            // 单字段: sortField > sv（仅适合唯一字段）
            if input.SortDirection == modelruntime.OrderByDesc {
                cursorExpr = goqu.C(input.SortField).Lt(after.SortValue)
            } else {
                cursorExpr = goqu.C(input.SortField).Gt(after.SortValue)
            }
        }
        ds = ds.Where(cursorExpr)
    }

    // 额外 WHERE（RLS 等）
    if len(input.Where) > 0 {
        whereExpr, werr := convertWhereToExpression(input.Where)
        if werr != nil {
            return "", nil, bizerrors.Errorf("listPage where: %w", werr)
        }
        ds = ds.Where(whereExpr)
    }

    sql, args, err = ds.Prepared(true).ToSQL()
    return
}
```

> **重要：** `input.After` 的类型是 `*cursorData`，需要将 `cursorData` 从 `modelruntime` 包导出（将首字母大写）或通过 getter 暴露。**修改 cursor.go 将 `cursorData` 改为 `CursorData`（导出）**，并相应更新所有引用。

- [ ] **Step 6: 导出 CursorData**

打开 `cursor.go`，将 `cursorData` → `CursorData`，所有字段保持不变（json tag 不变），更新 `encodeCursor`、`decodeCursor`、`CursorDataForTest` 的引用。

更新 `ListPageInput` 中的 `After *CursorData`。

- [ ] **Step 7: 运行测试确认通过**

```bash
cd modelcraft-backend
go test ./internal/infrastructure/database/dml/ -run TestConvertListPage -v
go test ./internal/domain/modelruntime/ -run TestEncodeCursor -v
```

预期：`PASS`

- [ ] **Step 8: Commit**

```bash
git add internal/domain/modelruntime/cursor.go \
        internal/domain/modelruntime/cursor_test.go \
        internal/domain/modelruntime/graphql_input.go \
        internal/infrastructure/database/dml/sql_mapper.go \
        internal/infrastructure/database/dml/sql_mapper_listpage_test.go
git commit -m "feat(modelruntime): ListPageInput and keyset SQL generation"
```

---

## Task 8: ClientDatabaseRepository — 加 ListPage 方法

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/graphql_repository.go`
- Modify: `modelcraft-backend/internal/infrastructure/database/dml/client_db_repo_impl.go`

- [ ] **Step 1: 接口加 ListPage**

打开 `internal/domain/modelruntime/graphql_repository.go`，在 `FindMany` 方法后加：

```go
// ListPage 执行 cursor 分页查询，返回最多 limit+1 条记录（调用方判断 hasNextPage）。
ListPage(ctx context.Context, input *ListPageInput) ([]map[string]any, error)
```

- [ ] **Step 2: 实现 ListPage**

打开 `internal/infrastructure/database/dml/client_db_repo_impl.go`，在 `FindMany` 方法后加：

```go
// ListPage 执行 cursor 分页查询
func (c *ClientDBRepoImpl) ListPage(ctx context.Context, input *modelruntime.ListPageInput) ([]map[string]any, error) {
    logger := logfacade.GetLogger(ctx)
    return execute[*modelruntime.ListPageInput, []map[string]any](
        ctx, logger, input,
        func() ([]map[string]any, error) {
            sql, args, err := convertListPageInputToSQL(ctx, input)
            if err != nil {
                return nil, err
            }
            logger.Infof(ctx, "sql=%v args=%v", sql, args)

            rows, err := c.stdDB.Queryx(sql, args...)
            if err != nil {
                logger.Error(ctx, "listPage query fail", logfacade.Err(err))
                return nil, err
            }
            defer rows.Close()

            results := make([]map[string]any, 0, input.Limit+1)
            for rows.Next() {
                record := make(map[string]any)
                if err := rows.MapScan(record); err != nil {
                    return nil, err
                }
                results = append(results, convertBytesToString(record))
            }
            if err := rows.Err(); err != nil {
                return nil, err
            }
            return results, nil
        },
    )
}
```

- [ ] **Step 3: 编译验证（接口满足检查）**

```bash
cd modelcraft-backend
go build ./...
```

预期：0 errors（`ClientDBRepoImpl` 已实现 `ClientDatabaseRepository` 接口）

- [ ] **Step 4: Commit**

```bash
git add internal/domain/modelruntime/graphql_repository.go \
        internal/infrastructure/database/dml/client_db_repo_impl.go
git commit -m "feat(modelruntime): add ListPage to ClientDatabaseRepository"
```

---

## Task 9: GraphQL — createListPageField resolver

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/graphql_constants.go`
- Modify: `modelcraft-backend/internal/domain/modelruntime/graphql_input_types.go`
- Modify: `modelcraft-backend/internal/domain/modelruntime/model_resolver.go`

- [ ] **Step 1: 加常量**

打开 `internal/domain/modelruntime/graphql_constants.go`，在 `OperationCount` 后加：

```go
// OperationListPage cursor 分页查询操作
OperationListPage = "listPage"
// FieldAfter listPage after cursor 参数
FieldAfter = "after"
// FieldNextCursor listPage 返回的下一页 cursor
FieldNextCursor = "nextCursor"
// FieldHasNextPage listPage 是否有下一页
FieldHasNextPage = "hasNextPage"
// FieldSortField listPage 排序字段参数
FieldSortField = "sortField"
// FieldSortDirection listPage 排序方向参数
FieldSortDirection = "sortDirection"
```

- [ ] **Step 2: 加 GenerateListPageArgs**

打开 `internal/domain/modelruntime/graphql_input_types.go`，在 `GenerateFindManyArgs` 方法后加：

```go
// GenerateListPageArgs 为 listPage 查询生成参数配置
// sortField: 必填，用于排序的字段名
// sortDirection: 必填，"asc" 或 "desc"
// limit: 可选，默认 20
// after: 可选，上一页返回的 nextCursor 字符串
func (g *inputTypeGenerator) GenerateListPageArgs(model *RuntimeModel) graphql.FieldConfigArgument {
    return graphql.FieldConfigArgument{
        FieldSortField: &graphql.ArgumentConfig{
            Type:        graphql.NewNonNull(graphql.String),
            Description: "排序字段名（必填）",
        },
        FieldSortDirection: &graphql.ArgumentConfig{
            Type:         graphql.NewNonNull(graphql.String),
            Description:  "排序方向：asc 或 desc（必填）",
            DefaultValue: OrderByAsc,
        },
        FieldLimit: &graphql.ArgumentConfig{
            Type:         graphql.Int,
            DefaultValue: 20,
            Description:  "每页条数，默认 20",
        },
        FieldAfter: &graphql.ArgumentConfig{
            Type:        graphql.String,
            Description: "上一页返回的 nextCursor（第一页不传）",
        },
    }
}
```

- [ ] **Step 3: 加 createListPageResultType、executeListPage、createListPageField**

打开 `internal/domain/modelruntime/model_resolver.go`，在 `createFindManyResultType` 方法后加：

```go
// createListPageResultType 创建 listPage 结果类型
func (r *graphqlModelResolver) createListPageResultType(modelType graphql.Type) *graphql.Object {
    return graphql.NewObject(graphql.ObjectConfig{
        Name: gqlTypeName(r.model.Name) + "ListPageResult",
        Fields: graphql.Fields{
            FieldItems: &graphql.Field{
                Type:        graphql.NewList(graphql.NewNonNull(modelType)),
                Description: "当前页数据",
            },
            FieldNextCursor: &graphql.Field{
                Type:        graphql.String,
                Description: "下一页 cursor（nil 表示已是最后一页）",
            },
            FieldHasNextPage: &graphql.Field{
                Type:        graphql.NewNonNull(graphql.Boolean),
                Description: "是否有下一页",
            },
            FieldTimeCost: &graphql.Field{
                Type:        graphql.NewNonNull(graphql.Int),
                Description: "查询执行时间（毫秒）",
            },
            FieldReqId: &graphql.Field{
                Type:        graphql.NewNonNull(graphql.String),
                Description: "请求追踪 ID",
            },
        },
        Description: "listPage cursor 分页结果",
    })
}

func (m *graphqlModelResolver) executeListPage(p graphql.ResolveParams) (map[string]any, error) {
    rctx, _ := getGraphqlRequestContext(p.Context)
    if err := rctx.EndUserPerms.CheckAction(ActionSelect); err != nil {
        return nil, err
    }
    startTime := time.Now()

    // 解析参数
    sortField, _ := p.Args[FieldSortField].(string)
    sortDirection, _ := p.Args[FieldSortDirection].(string)
    if sortDirection != OrderByAsc && sortDirection != OrderByDesc {
        sortDirection = OrderByAsc
    }
    limitRaw, _ := p.Args[FieldLimit].(int)
    if limitRaw <= 0 {
        limitRaw = 20
    }
    limit := uint(limitRaw)

    // 解析 after cursor
    var after *CursorData
    if afterStr, ok := p.Args[FieldAfter].(string); ok && afterStr != "" {
        decoded, err := decodeCursor(afterStr)
        if err != nil {
            return nil, bizerrors.Errorf("invalid after cursor: %w", err)
        }
        after = &decoded
    }

    // 获取插入序字段配置
    insertionOrderField := ""
    if m.model.InsertionOrderField != nil {
        insertionOrderField = *m.model.InsertionOrderField
    }

    input := &ListPageInput{
        TableName:           m.model.Name,
        SortField:           sortField,
        SortDirection:       sortDirection,
        InsertionOrderField: insertionOrderField,
        After:               after,
        Limit:               limit,
    }

    // 注入 RLS WHERE
    if rf := BuildRowFilter(
        rctx.EndUserPerms, ActionSelect, m.endUserRefFieldName(), rctx.CurrentEndUserID,
    ); rf != nil {
        maps.Copy(input.Where, rf)
    }

    // 取 limit+1 判断 hasNextPage
    rows, err := rctx.ClientRepo.ListPage(p.Context, input)
    if err != nil {
        return nil, err
    }

    hasNextPage := len(rows) > int(limit)
    if hasNextPage {
        rows = rows[:limit] // 截掉多取的那一条
    }

    // 构造 nextCursor（取最后一条记录的字段值）
    var nextCursorStr *string
    if hasNextPage && len(rows) > 0 {
        last := rows[len(rows)-1]
        sv := fmt.Sprintf("%v", last[sortField])
        cd := CursorData{SortField: sortField, SortValue: sv}
        if insertionOrderField != "" {
            cd.IOField = insertionOrderField
            cd.IOValue = fmt.Sprintf("%v", last[insertionOrderField])
        }
        encoded := encodeCursor(cd)
        nextCursorStr = &encoded
    }

    metadata := requestcontext.GetMetadata(p.Context)
    reqId := ""
    if metadata != nil {
        reqId = metadata.ReqID
    }

    return map[string]any{
        FieldItems:      rows,
        FieldNextCursor: nextCursorStr,
        FieldHasNextPage: hasNextPage,
        FieldTimeCost:   int(time.Since(startTime).Milliseconds()),
        FieldReqId:      reqId,
    }, nil
}

func (m *graphqlModelResolver) createListPageField(modelType graphql.Type) (*graphql.Field, error) {
    args := m.inputTypeGenerator.GenerateListPageArgs(m.model)
    resultType := m.createListPageResultType(modelType)

    return &graphql.Field{
        Type: resultType,
        Args: args,
        Resolve: func(p graphql.ResolveParams) (interface{}, error) {
            result, err := m.executeListPage(p)
            if err != nil {
                logfacade.GetLogger(p.Context).Error(p.Context, "executeListPage fail", logfacade.Err(err))
                return nil, err
            }
            return result, err
        },
    }, nil
}
```

- [ ] **Step 4: 在 createRootQuery 注册 listPage**

找到 `createRootQuery` 方法，在 `countField` 创建后、`rootQuery` 构建前加：

```go
listPageField, err := m.createListPageField(modelType)
if err != nil {
    return nil, err
}
```

在 `graphql.Fields{}` 中加：

```go
OperationListPage: listPageField,
```

- [ ] **Step 5: 编译验证**

```bash
cd modelcraft-backend
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add internal/domain/modelruntime/
git commit -m "feat(modelruntime): add listPage cursor pagination resolver"
```

---

## Task 10: 集成验证 — 本地跑通 listPage

**Files:** 无新文件，验证已有实现

- [ ] **Step 1: 启动后端**

```bash
cd modelcraft-backend
just run force=true
```

- [ ] **Step 2: 用 curl 测试 listPage（第一页）**

```bash
TOKEN="<your-end-user-token>"
ORG="luke_5l0o"
PROJECT="luke"
DB="your-db"
MODEL="products"

curl -s -X POST \
  "http://localhost:8080/end-user/graphql/org/$ORG/project/$PROJECT/db/$DB/model/$MODEL" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ listPage(sortField: \"price\", sortDirection: \"asc\", limit: 5) { items { id price } nextCursor hasNextPage timeCost } }"
  }' | jq .
```

预期：返回 `items`（≤5 条）、`hasNextPage`（bool）、`nextCursor`（有下一页时为字符串）

- [ ] **Step 3: 测试第二页（带 after）**

将上一步的 `nextCursor` 值填入 `after`：

```bash
CURSOR="<nextCursor from step 2>"

curl -s -X POST \
  "http://localhost:8080/end-user/graphql/org/$ORG/project/$PROJECT/db/$DB/model/$MODEL" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"query\": \"{ listPage(sortField: \\\"price\\\", sortDirection: \\\"asc\\\", limit: 5, after: \\\"$CURSOR\\\") { items { id price } nextCursor hasNextPage } }\"
  }" | jq .
```

预期：返回第 6~10 条，price 均大于等于第一页最后一条的 price

- [ ] **Step 4: Commit（如有 fix）**

```bash
git commit -am "fix(modelruntime): listPage integration fixes"
```

---

## Task 11: 管理态前端 — insertionOrderField 配置 UI

**Files:**
- Sync: `modelcraft-front/contract/`
- Modify: 模型配置页的 meta 表单组件

- [ ] **Step 1: 同步 contract**

```bash
# 在 modelcraft-front 目录，使用 front-contract-pull skill 同步
```

- [ ] **Step 2: 生成前端类型**

```bash
cd modelcraft-front
npm run generate
```

预期：`src/generated/graphql.ts` 中 `Model` 加了 `insertionOrderField?: Maybe<string>`

- [ ] **Step 3: 找到 displayField 的配置 UI 组件**

```bash
grep -rn "displayField\|DisplayField" modelcraft-front/src/app --include="*.tsx" -l | head -5
```

找到渲染 `displayField` 选择器和警告提示的组件文件。

- [ ] **Step 4: 仿照 displayField 加 insertionOrderField 选择器**

在同一个组件中，紧跟 `displayField` 配置块后，加 `insertionOrderField` 的 Select 组件：

```tsx
{/* 插入序字段 */}
<div className="space-y-1.5">
  <Label>插入序字段</Label>
  <Select
    value={form.insertionOrderField ?? ''}
    onValueChange={(v) => setForm((f) => ({ ...f, insertionOrderField: v || null }))}
  >
    <SelectTrigger>
      <SelectValue placeholder="未配置" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem value="">未配置</SelectItem>
      {fields.map((f) => (
        <SelectItem key={f.name} value={f.name}>
          {f.name}
        </SelectItem>
      ))}
    </SelectContent>
  </Select>
  {!form.insertionOrderField && (
    <p className="flex items-center gap-1.5 text-xs text-amber-600">
      <span className="shrink-0">⚠</span>
      未配置插入序字段，<code className="rounded bg-muted px-1 font-mono">listPage</code> 分页结果稳定性无法保证。
    </p>
  )}
  {form.insertionOrderField && (
    <p className="text-xs text-muted-foreground">
      listPage 将使用 <code className="font-mono">{form.insertionOrderField}</code> 作为 cursor tiebreaker，确保分页稳定。
    </p>
  )}
</div>
```

- [ ] **Step 5: 更新 updateModelMeta mutation 调用**

找到调用 `updateModelMeta` 的地方，在 variables 中加入 `insertionOrderField: form.insertionOrderField`。

- [ ] **Step 6: 前端 lint**

```bash
cd modelcraft-front
npm run lint
```

预期：0 errors

- [ ] **Step 7: Commit**

```bash
git add modelcraft-front/
git commit -m "feat(front): add insertionOrderField config UI with stability warning"
```

---

## Task 12: 前端 runtime-query-builder — 加 buildListPageQuery

**Files:**
- Modify: `modelcraft-front/src/api-client/runtime-query/runtime-query-builder.ts`
- Modify: `modelcraft-front/src/api-client/runtime-query/runtime-query-builder.test.ts`

- [ ] **Step 1: 写失败测试**

打开 `runtime-query-builder.test.ts`，加：

```typescript
describe('buildListPageQuery', () => {
  it('should build a valid listPage query with required fields', () => {
    const query = buildListPageQuery('Product', ['id', 'price', 'name'])
    const printed = print(query)
    expect(printed).toContain('listPage')
    expect(printed).toContain('sortField')
    expect(printed).toContain('sortDirection')
    expect(printed).toContain('after')
    expect(printed).toContain('limit')
    expect(printed).toContain('nextCursor')
    expect(printed).toContain('hasNextPage')
    expect(printed).toContain('items')
  })
})
```

- [ ] **Step 2: 运行确认失败**

```bash
cd modelcraft-front
npx jest runtime-query-builder --testNamePattern="buildListPageQuery"
```

- [ ] **Step 3: 实现 buildListPageQuery**

在 `runtime-query-builder.ts` 中加：

```typescript
/**
 * Build a listPage cursor pagination query.
 * Variables: sortField (String!), sortDirection (String!), limit (Int), after (String)
 */
export function buildListPageQuery(
  modelName: string,
  fields: string[] | FieldDefinition[] | (string | FieldDefinition)[]
): DocumentNode {
  const fieldSelections = buildFieldSelections(fields)
  const operationName = `${modelName}ListPage`

  return gql`
    query ${operationName}(
      $sortField: String!
      $sortDirection: String!
      $limit: Int
      $after: String
    ) {
      listPage(
        sortField: $sortField
        sortDirection: $sortDirection
        limit: $limit
        after: $after
      ) {
        items {
          ${fieldSelections}
        }
        nextCursor
        hasNextPage
        timeCost
        reqId
      }
    }
  `
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd modelcraft-front
npx jest runtime-query-builder --testNamePattern="buildListPageQuery"
```

- [ ] **Step 5: Lint 验证**

```bash
cd modelcraft-front
npm run lint
```

- [ ] **Step 6: Commit**

```bash
git add modelcraft-front/src/api-client/runtime-query/
git commit -m "feat(front): add buildListPageQuery for cursor pagination"
```

---

## Self-Review

### Spec Coverage Check

| 需求 | 覆盖任务 |
|------|---------|
| `list` 接口去掉 `hasNextPage` | ⚠️ **Gap** — 当前 plan 未包含此项。需补一个 Task 修改 `createFindManyResultType` 去掉 `hasNextPage` 字段，更新前端 `findMany` 用法 |
| `listPage` 始终生成 | Task 9 ✅ |
| cursor 编解码（base64 JSON） | Task 6 ✅ |
| 双字段 cursor（有插入序字段时） | Task 7 ✅ |
| 单字段 cursor（无插入序字段时） | Task 7 ✅ |
| 管理态 DB/Schema 加 `insertionOrderField` | Task 1–2 ✅ |
| Domain/App 层传递新字段 | Task 3–5 ✅ |
| 管理态前端配置 UI + 风险提示 | Task 11 ✅ |
| 前端 `buildListPageQuery` | Task 12 ✅ |
| 后端集成验证 | Task 10 ✅ |

### 补充 Task 13: list 去掉 hasNextPage

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/model_resolver.go`（`createFindManyResultType` 移除 `FieldHasNextPage`；`executeFindMany` 不再计算 hasNextPage）
- 前端所有使用 `findMany { hasNextPage }` 的地方删除该字段引用

> 这是一个破坏性变更，执行前确认前端没有依赖 `findMany.hasNextPage` 的功能代码：
> ```bash
> grep -rn "hasNextPage" modelcraft-front/src --include="*.ts" --include="*.tsx"
> ```
