# Change: 重构枚举字段关联机制 (Refactor Enum Field Association)

## Why

当前枚举字段关联设计存在以下问题:

1. **外键设计不当**: `field_definitions` 表直接存储 `enum_name` 字段关联枚举,但这是一个业务标识而非外键,无法使用数据库外键约束保证引用完整性
2. **缺少关联管理**: 没有专门的关联表记录模型字段与枚举的关系,无法追踪枚举被哪些字段使用
3. **级联删除缺失**: 删除字段时无法自动清理枚举关联关系
4. **GraphQL API 设计不明确**: `validationConfig` 中混合了枚举配置(`enum`字段),职责不清晰

类比 `model_relations` 表的设计,我们需要一个专门的关联表来管理模型字段与枚举之间的多对一关系。

## What Changes

### 数据库层面
- **新增关联表** `model_field_enum_associations`:记录 `cluster_name`, `database_name`, `model_id`, `field_name`, `enum_id`
- **外键约束**: `(model_id, field_name)` 作为联合外键关联 `field_definitions` 表,支持级联删除
- **外键约束**: `enum_id` 关联 `model_enums` 表,防止删除被引用的枚举
- **移除字段**: 从 `field_definitions` 表删除 `enum_name` 字段

### Domain 层面
- **移除字段**: `FieldDefinition.EnumName` 字段删除
- **保留字段**: `FieldDefinition.Enum` 字段保留用于查询时加载枚举详情
- **新增实体**: `ModelFieldEnumAssociation` 聚合根,包含关联的完整信息

### GraphQL API 层面
- **移除**: `AddFieldInput.validationConfig.enum` 字段
- **新增**: `AddFieldInput.enumConfig` 输入类型
  - `enumName: String` - 枚举名称(用于关联或创建)
  - `options: [EnumOptionInput!]` - 枚举选项(创建新枚举时使用)
  - `description: String` - 枚举描述
  - `connectEnum: Boolean!` - 决定是关联现有枚举(true)还是创建新枚举(false)
- **行为变更**:
  - 当 `connectEnum=true` 时,使用 `enumName` 关联现有枚举,如果不存在则报错
  - 当 `connectEnum=false` 时,使用 `enumName`、`options`、`description` 创建新枚举,如果名称已存在则报错
  - 创建关联记录存储在 `model_field_enum_associations` 表

## Impact

### 受影响的规范 (Affected Specs)
- `modeldesign-field-types`: 重构枚举字段关联机制

### 受影响的代码 (Affected Code)

#### Domain Layer
- `internal/domain/modeldesign/field_definition.go`: 删除 `EnumName` 字段
- `internal/domain/modeldesign/field_enum_association.go`: 新建关联实体
- `internal/domain/modeldesign/field_enum_association_repository.go`: 新建关联仓储接口

#### Infrastructure Layer
- `internal/infrastructure/repository/field_definition_model.go`: 删除 `EnumName` 字段映射
- `internal/infrastructure/repository/field_enum_association_model.go`: 新建关联数据模型
- `internal/infrastructure/repository/field_enum_association_repository.go`: 实现关联仓储

#### Application Layer
- `internal/app/modeldesign/field_service.go`: 调整字段创建逻辑,处理枚举关联
- `internal/app/modeldesign/enum_service.go`: 调整枚举删除逻辑,检查关联

#### Interface Layer
- `api/graph/schema/field.graphql`: 调整 GraphQL schema
- `internal/interfaces/graphql/field.resolvers.go`: 调整解析器逻辑
- `internal/interfaces/mapper/field_mapper.go`: 调整字段映射器

### 破坏性变更 (Breaking Changes)
**BREAKING**: GraphQL API 变更
- `AddFieldInput.validationConfig.enum` 字段移除
- 客户端需要调整为使用新的 `AddFieldInput.enumConfig`

**BREAKING**: Domain 模型变更
- `FieldDefinition.EnumName` 字段删除
- 需要通过关联表查询枚举关联关系

### 兼容性说明
- 现有使用 `ValidationConfig.EnumValues` 的字段继续工作(简单枚举值,不使用中心化枚举)
