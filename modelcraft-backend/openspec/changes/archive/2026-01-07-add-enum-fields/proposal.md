# Change: Add Centralized Enum Field Support

## Why

当前系统字段定义中的 `EnumValues` 实现为字段级别的硬编码枚举值列表,存在以下问题:

1. **缺乏集中管理**: 每个字段的枚举值独立存储,无法跨字段复用枚举定义
2. **维护困难**: 当枚举值需要变更时,需要修改所有使用该枚举的字段
3. **功能受限**: 无法支持枚举元数据(如显示名称、描述、排序等)
4. **无版本控制**: 无法追踪枚举值的变更历史

类比关联关系的设计,我们需要将枚举定义作为独立的元数据资源进行中心化管理。

## What Changes

- **新增枚举定义表** `model_enums`: 单表存储中心化的枚举元数据,使用JSON字段存储枚举选项
- **扩展字段定义表**: 在 `field_definitions` 表中新增 `enum_name` 字段,用于关联枚举定义(使用业务标识name而非ID)
- **扩展字段类型**: 新增 `ENUM` 和 `ENUM_ARRAY` 格式类型
- **建立关联机制**: 字段通过 `EnumName` 引用枚举定义,类似 `ModelName` 的设计模式,支持导入导出
- **逻辑字段支持**: 枚举字段作为逻辑字段,依赖物理字段存储key值
- **双重存储**: 物理字段存储key(单选为string,多选为JSON数组),逻辑字段提供完整枚举信息(包含value展示值)
- **严格校验**: 枚举选项的key和value必须非空,创建和更新时强制验证

## Impact

### 受影响的规范 (Affected Specs)
- `modeldesign-field-types`: 新增枚举字段类型和验证规则

### 受影响的代码 (Affected Code)
- `internal/domain/modeldesign/field_definition.go`: 新增 `EnumName` 和 `Enum` 字段
- `internal/domain/modeldesign/enum_definition.go`: 新建枚举领域模型
- `internal/domain/modeldesign/enum_repository.go`: 新建枚举仓储接口
- `internal/infrastructure/repository/enum_definition_repository.go`: 实现枚举仓储
- `internal/infrastructure/repository/enum_definition_model.go`: 枚举数据库模型
- `internal/infrastructure/repository/field_definition_model.go`: 新增 `EnumName` 字段
- `internal/app/modeldesign/enum_service.go`: 枚举应用服务
- `pkg/schema/core/field_type.go`: 新增 `FormatEnum` 和 `FormatEnumArray` 常量
- GraphQL schema 生成器: 支持枚举类型生成
- 字段验证逻辑: 新增枚举值验证

### 数据库变更
- 新建表: `model_enums`
- 修改表: `field_definitions` 新增 `enum_name` 字段

### 破坏性变更 (Breaking Changes)
**无破坏性变更**。现有的 `EnumValues` 字段保留用于简单场景,新的枚举引用机制为可选功能。

### 兼容性说明
- 现有使用 `EnumValues` 的字段继续工作
- 新字段可选择使用中心化枚举或简单枚举值
- 系统同时支持两种枚举实现方式
