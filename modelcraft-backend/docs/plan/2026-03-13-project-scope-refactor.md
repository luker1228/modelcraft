# Plan: 引入 ProjectScope，统一项目上下文

**日期**: 2026-03-13

## 目标

在 `internal/domain/project` 包下新建 `ProjectScope` 值对象，嵌入所有需要项目上下文的实体，替换各实体中独立的 `OrgName + ProjectSlug` 字段组合，确保数据完整性（OrgName 不为空）。

---

## Step 1: 新建 ProjectScope

**新建文件**: `internal/domain/project/project_scope.go`

```go
// ProjectScope 项目作用域，表示某个组织下的某个项目。
// 用于需要完整项目上下文的实体，确保数据完整性。
type ProjectScope struct {
    OrgName     string
    ProjectSlug string
}

func (s *ProjectScope) Validate() error {
    if s.OrgName == "" {
        return bizerrors.New("org name is required")
    }
    if s.ProjectSlug == "" {
        return bizerrors.New("project slug is required")
    }
    return nil
}

// NewProjectScope 创建并验证 ProjectScope
func NewProjectScope(orgName, projectSlug string) (ProjectScope, error) { ... }

// GetFullPath 返回 "orgName.projectSlug"
func (s *ProjectScope) GetFullPath() string { ... }
```

**新建测试文件**: `internal/domain/project/project_scope_test.go`

测试覆盖 `ProjectScope` 不合法的场景：
- `OrgName` 为空
- `ProjectSlug` 为空
- 两者均为空
- 合法场景

---

## Step 2: 改造 ModelLocator

**文件**: `internal/domain/modeldesign/model.go`

**改前**:
```go
type ModelLocator struct {
    ProjectSlug  string
    ModelName    string
    DatabaseName string
}
```

**改后**:
```go
type ModelLocator struct {
    project.ProjectScope        // 嵌入，包含 OrgName + ProjectSlug
    DatabaseName string
    ModelName    string
}
```

变更点：
- `Validate()`：委托 `ProjectScope.Validate()`，追加 `DatabaseName`、`ModelName` 校验
- `NewModelLocator()`：签名增加 `orgName` 参数
- `String()` / `GetFullPath()`：路径变为 `orgName.projectSlug.databaseName.modelName`
- `GetDatabasePath()`：路径变为 `orgName.projectSlug.databaseName`
- `DataModel.Validate()`：增加 `OrgName` 不为空校验

**更新测试**: `internal/domain/modeldesign/model_test.go`

新增/更新测试场景：
- `TestModelLocator_Validate`：增加 `OrgName` 为空的用例
- `TestNewModelLocator`：签名变更，增加 `orgName` 参数
- `TestModelLocator_GetFullPath`：期望值变为 4 段路径
- `TestModelLocator_GetDatabasePath`：期望值变为 3 段路径
- `TestDataModel_Validate`：增加 `OrgName` 为空的用例

---

## Step 3: 改造 EnumDefinition

**文件**: `internal/domain/modeldesign/enum_definition.go`

**改前**:
```go
type EnumDefinition struct {
    OrgName     string
    ProjectSlug string
    ...
}
```

**改后**:
```go
type EnumDefinition struct {
    project.ProjectScope   // 替换 OrgName + ProjectSlug
    ...
}
```

- `Validate()`：委托 `ProjectScope.Validate()`

---

## Step 4: 改造 ModelGroup

**文件**: `internal/domain/modeldesign/model_group.go`

```go
type ModelGroup struct {
    project.ProjectScope   // 替换 OrgName + ProjectSlug
    ...
}
```

---

## Step 5: 改造 FieldEnumAssociation

**文件**: `internal/domain/modeldesign/field_enum_association.go`

注意：当前只有 `ProjectSlug`，无 `OrgName`，本次补齐。

```go
type FieldEnumAssociation struct {
    project.ProjectScope   // 替换 ProjectSlug，补充 OrgName
    ...
}
```

- `Validate()`：委托 `ProjectScope.Validate()`
- `NewFieldEnumAssociation()`：签名增加 `orgName` 参数

---

## Step 6: 修复调用方编译错误

逐层修复：

| 层 | 文件 | 主要变更 |
|----|------|---------|
| Repository | `internal/infrastructure/repository/sql_modeldesign_repository.go` | 构建实体时补充 `OrgName` |
| Application | `internal/app/modeldesign/` 各文件 | `NewModelLocator` 签名变更 |
| Application | `internal/app/modelruntime/graphql_app.go` | `NewModelLocator` 签名变更 |

---

## Step 7: 更新相关测试

- `internal/infrastructure/repository/modeldesign_convert_test.go`
- `internal/domain/modeldesign/model_test.go`（Step 2 已列出）

---

## 关键文件汇总

| 文件 | 操作 |
|------|------|
| `internal/domain/project/project_scope.go` | **新建** |
| `internal/domain/project/project_scope_test.go` | **新建** |
| `internal/domain/modeldesign/model.go` | 改造 ModelLocator |
| `internal/domain/modeldesign/model_test.go` | 更新测试 |
| `internal/domain/modeldesign/enum_definition.go` | 替换 OrgName+ProjectSlug |
| `internal/domain/modeldesign/model_group.go` | 替换 OrgName+ProjectSlug |
| `internal/domain/modeldesign/field_enum_association.go` | 替换，补充 OrgName |
| `internal/infrastructure/repository/sql_modeldesign_repository.go` | 修复构建调用 |
| `internal/app/modeldesign/` 各文件 | 修复构建调用 |
| `internal/app/modelruntime/graphql_app.go` | 修复构建调用 |
