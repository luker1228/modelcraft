# 5.1 模型设计（设计态）

> 代码位置：`internal/domain/modeldesign/`

## 核心实体

### DataModel — 模型定义

```
internal/domain/modeldesign/model.go

DataModel
├── ModelMeta
│   ├── ID               string
│   ├── ModelLocator     (OrgName, ProjectSlug, DatabaseName, ModelName)
│   ├── Title            string
│   ├── Description      string
│   ├── StorageType      string
│   ├── Version          int64
│   ├── Status           string
│   ├── GroupID          *string          // 所属模型分组
│   ├── DeploymentStatus DeploymentStatus // 部署状态
│   └── LastSyncAt       *time.Time
└── Fields  []*FieldDefinition
```

**ModelLocator** 是模型的定位坐标：`{OrgName}.{ProjectSlug}.{DatabaseName}.{ModelName}`

```go
type ModelLocator struct {
    project.ProjectScope  // 嵌入项目作用域（包含 OrgName + ProjectSlug）
    DatabaseName string
    ModelName    string
}
```

通过嵌入 `ProjectScope` 值对象确保 OrgName 和 ProjectSlug 始终同时存在且有效。

### FieldDefinition — 字段定义

```
internal/domain/modeldesign/field_definition.go

FieldDefinition
├── ModelID      string
├── ModelLocator *ModelLocator
├── Name         string        // 字段名（snake_case）
├── Title        string        // 显示名
├── Type         string        // boolean | number | string | array | object
├── Format       string        // text | json | markdown | uuid | date | datetime
│                              // integer | float | richtext | time 等
├── Required     bool
├── ReadOnly     bool
├── MaxLength    *int
├── Order        string        // 字典序分数索引（lexicographic fractional index）
└── ... （枚举关联、外键关联等）
```

### EnumDefinition — 枚举类型

```
internal/domain/modeldesign/enum_definition.go

EnumDefinition
├── ID           string
├── ProjectScope              // 嵌入项目作用域
│   ├── OrgName  string
│   └── ProjectSlug string
├── Name         string       // 英文标识，项目内唯一
├── Title        string
├── Description  string
├── Options      []EnumOption // {Code, Label, Order}
└── IsMultiSelect bool        // 是否支持多选
```

通过嵌入 `ProjectScope` 确保枚举定义始终属于某个项目。

### ModelGroup — 模型分组

```
internal/domain/modeldesign/model_group.go

ModelGroup
├── ID           string
├── ProjectScope              // 嵌入项目作用域
│   ├── OrgName  string
│   └── ProjectSlug string
├── Name         string       // 分组名称
├── DisplayOrder string       // 字典序分数索引
├── CreatedAt    time.Time
└── UpdatedAt    time.Time
```

通过嵌入 `ProjectScope` 确保分组始终属于某个项目。

### FieldEnumAssociation — 字段枚举关联

```
internal/domain/modeldesign/field_enum_association.go

FieldEnumAssociation
├── ModelID       string
├── FieldName     string
├── ProjectScope              // 嵌入项目作用域
│   ├── OrgName   string
│   └── ProjectSlug string
├── EnumName      string      // 关联的枚举类型
├── DatabaseName  string
├── CreatedAt     time.Time
└── UpdatedAt     time.Time
```

表示模型字段与枚举定义的多对一关系。通过嵌入 `ProjectScope` 确保字段、枚举、模型都在同一项目范围内。

### ModelGroup — 模型分组

```
internal/domain/modeldesign/model_group.go

ModelGroup
├── ID          string
├── OrgName     string
├── ProjectSlug string
├── Name        string
└── Order       string   // 字典序分数索引
```

### LogicalForeignKey — 逻辑外键（关联关系）

```
internal/domain/modeldesign/logical_foreign_key.go

LogicalForeignKey
├── SourceModel   ModelLocator
├── SourceField   string
├── TargetModel   ModelLocator
├── TargetField   string
└── （成对创建，源端和目标端各一条记录）
```

## ProjectScope 嵌入模式

多个实体使用 Go 的 **结构体嵌入** 特性复用 `ProjectScope`，避免 OrgName/ProjectSlug 字段重复定义：

```go
// internal/domain/project/project_scope.go
type ProjectScope struct {
    OrgName     string
    ProjectSlug string
}

// 嵌入示例（ModelLocator）
type ModelLocator struct {
    project.ProjectScope  // ← 嵌入
    DatabaseName string
    ModelName    string
}

// 嵌入后，以下字段被提升到外层
locator.OrgName      // ✅ 可直接访问
locator.ProjectSlug  // ✅ 可直接访问
locator.DatabaseName // ✅ 直接字段
locator.ModelName    // ✅ 直接字段

// Validate 方法同样继承
err := locator.Validate() // ✅ 首先验证 ProjectScope，再验证其他字段
```

**好处**：
- 避免 OrgName/ProjectSlug 的重复定义
- 确保一致的验证逻辑
- 类型安全的项目上下文

## Schema 同步机制

```
设计态定义（DataModel）
        │
        ▼
  Schema Compare
  (comparison_service.go)
        │
        ├── 发现 diff（新增字段、修改类型、删除字段）
        │
        ▼
  生成 DDL（actual_schema.go）
        │
        ▼
  Apply 到目标 MySQL
        │
        ▼
  更新 DeploymentStatus / LastSyncAt
```

## 相关文件

- `internal/domain/project/project_scope.go` — ProjectScope 值对象
- `internal/domain/modeldesign/model.go` — DataModel & ModelLocator
- `internal/domain/modeldesign/field_definition.go` — 字段定义
- `internal/domain/modeldesign/enum_definition.go` — 枚举定义
- `internal/domain/modeldesign/field_enum_association.go` — 字段枚举关联
- `internal/domain/modeldesign/logical_foreign_key.go` — 逻辑外键
- `internal/domain/modeldesign/model_group.go` — 模型分组
- `internal/domain/modeldesign/comparison_service.go` — Schema 比对
- `internal/domain/modeldesign/actual_schema.go` — 目标库实际 Schema 读取
- `internal/domain/modeldesign/field_service.go` — 字段操作业务逻辑
- `internal/domain/modeldesign/type_mapper.go` — 字段类型 → SQL 类型映射
