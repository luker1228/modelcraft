# Tasks: Refactor Enum Feature - Code/Label Rename and Relation Support

## Database Impact Overview

### Phase 1: Code/Label Rename

| 表 | DDL 变更 | 数据兼容性 | 说明 |
|---|---|---|---|
| `model_enums` | ❌ 无 | ⚠️ 数据格式不兼容 | `options` JSON 字段从 `{"key": ..., "value": ...}` 变为 `{"code": ..., "label": ...}` |
| `model_field_enum_associations` | ❌ 无 | ✅ 兼容 | 仅存储关联关系，不受影响 |

**重要说明：Phase 1 是 BREAKING CHANGE**
- 数据库表结构不需要变更（无 DDL）
- 但 `model_enums.options` 存储的 JSON 格式变更，现有数据不兼容
- 部署后需要用户重新创建枚举定义
- 或者用户需要自行编写数据迁移脚本将 `key`/`value` 改为 `code`/`label`

### Phase 2: Enum Label Virtual Field

| 表 | DDL 变更 | 数据兼容性 | 说明 |
|---|---|---|---|
| `field_definitions` | ✅ 新增列 | ✅ 兼容 | 新增 `enum_label_config` JSON 列存储配置 |
| `model_enums` | ❌ 无 | ✅ 兼容 | 不涉及枚举定义变更 |
| `model_field_enum_associations` | ❌ 无 | ✅ 兼容 | 不涉及字段-枚举关联变更 |

**Phase 2 数据库变更（需要新增列）：**
```sql
-- 新增列 DDL
ALTER TABLE `field_definitions`
ADD COLUMN `enum_label_config` JSON NULL COMMENT '枚举标签虚拟字段配置（ENUM_LABEL 格式使用）';

-- 示例数据（新增 ENUM_LABEL 虚拟字段时）
UPDATE field_definitions
SET enum_label_config = '{"sourceField": "status"}'
WHERE model_id = 'model-uuid' AND name = 'statusLabel';
```

**重要说明：Phase 2 是 NEW FEATURE，兼容现有数据**
- 需要执行 DDL 变更：向 `field_definitions` 表新增 `enum_label_config` 列
- 新列可 NULL，不影响现有数据
- 仅 ENUM_LABEL 格式的字段会填充该列
- 虚拟字段仅存在于设计态字段定义和运行时 schema 生成，不会创建物理数据库列

## Phase 1: Code/Label Rename (设计态 / Design-Time)

> 本阶段修改设计态相关代码，影响枚举定义的创建和管理 API。

### Task 1.1: Update Domain Layer Model
- [x] Update `EnumOption` struct in `internal/domain/modeldesign/enum_definition.go`
  - Rename `Key` field to `Code` with JSON tag `"code"`
  - Rename `Value` field to `Label` with JSON tag `"label"`
  - Keep `Order`, `Description`, and other fields unchanged
- [x] Update validation methods to use new field names
  - Update `Validate()` method error messages
  - Update `GetOption(key)` → `GetOptionByCode(code)`
  - Update `HasOption(key)` → `HasOptionCode(code)`
  - Update `ValidateKeys(keys)` → `ValidateCodes(keys)`
- [x] Update utility methods on `EnumDefinition` struct
- [x] Update `Update()` method to accept options with code/label

### Task 1.2: Update Repository Layer
- [x] Update `EnumDefinitionPO` struct fields in `internal/infrastructure/repository/enum_definition_model.go`
  - Keep JSON serialization unchanged (will handle in converter)
- [x] Update `ToEnumDefinition()` converter to map code/label from JSON
- [x] Update `FromEnumDefinition()` converter to serialize code/label to JSON
- [x] Update repository implementations that reference enum options

### Task 1.3: Update Application Service Layer
- [x] Update `enum_service.go` service methods
  - Update references to `Key`/`Value` to use `Code`/`Label`
  - Update error messages to use new terminology
- [x] Update enum validation methods

### Task 1.4: Update HTTP Layer DTOs
- [x] Update `EnumOptionDTO` in `internal/interfaces/http/dtos/field_definition.go`
  - Rename `Key` to `Code` with JSON tag `"code"`
  - Rename `Value` to `Label` with JSON tag `"label"`
- [x] Update any other DTOs that reference enum options

### Task 1.5: Update GraphQL Type Mappers
- [x] Update `internal/interfaces/graphql/adapter/enum_mapper.go`
  - Update `ConvertEnumDefinitionToGraphQL()` to use code/label
- [x] Update `internal/interfaces/graphql/adapter/field_mapper.go`
  - Update `convertEnumConfigInput2DTO()` to map code/label
- [x] Update any other GraphQL mappers that reference enum options

### Task 1.6: Update GraphQL Schema Files
- [x] Update `api/graph/schema/enum.graphql`
  - Rename `EnumOption.key` to `code`
  - Rename `EnumOption.value` to `label`
  - Rename `EnumOptionInput.key` to `code`
  - Rename `EnumOptionInput.value` to `label`
- [x] Update `api/graph/schema/field.graphql` if enum options are referenced

### Task 1.7: Regenerate GraphQL Code
- [x] Run `make generate-gql` to regenerate GraphQL code
- [x] Verify `internal/interfaces/graphql/generated/model_gen.go` has updated fields
- [x] Verify `internal/interfaces/graphql/generated/generated.go` has updated resolvers

### Task 1.8: Update GraphQL Resolvers
- [x] Update `internal/interfaces/graphql/enum.resolvers.go`
  - Update references to `Key`/`Value` to use `Code`/`Label`
- [x] Update any other resolvers that handle enum options

### Task 1.9: Update HTTP Handlers
- [x] Update `internal/interfaces/http/handlers/design/enum_handler.go`
  - Update request/response handling to use new field names

### Task 1.10: Update Unit Tests
- [x] Update `internal/domain/modeldesign/enum_definition_test.go` (if exists)
  - Update test cases to use `Code`/`Label` fields
- [x] Update `internal/infrastructure/repository/enum_definition_repository_test.go` (if exists)
- [x] Update `internal/app/modeldesign/enum_service_test.go` (if exists)
- [x] Update `internal/interfaces/graphql/adapter/enum_mapper_test.go` (if exists)
- [x] Update any other test files that reference enum options
- [x] Verify all unit tests pass

### Task 1.11: Update JSONSchema Generator
- [x] Update `internal/domain/modeldesign/jsonschema_generator.go`
  - Update enum option serialization to use code/label
- [x] Update `internal/domain/modeldesign/jsonschema_parser.go`
  - Update enum option deserialization to handle code/label
- [x] Update related tests if any

### Task 1.12: Update Field Validator
- [x] Update `internal/domain/modeldesign/field_validator.go`
  - Update any enum-related validations to use new terminology

### Task 1.13: Update Documentation
- [x] Update API documentation comments in code
- [x] Update any inline documentation referencing `key`/`value`
- [x] No need to create new docs files unless required
- [x] Note: This is a **BREAKING CHANGE** - existing enum data stored with `key`/`value` JSON will need to be recreated

**Dependency**: Task 1.14 depends on all previous Phase 1 tasks

### Task 1.14: Phase 1 Verification
- [x] Run all tests: `make test`
- [x] Run integration tests: pytest in `tests/automated/`
- [x] Verify GraphQL schema validation
- [x] Manual test: create enum, query enum, verify code/label returned
- [x] Clean up any remaining references to `Key`/`Value` in enum context

## Phase 2: Enum Label Virtual Field (设计态 + 运行态 / Design-Time + Runtime)

> 本阶段实现枚举标签虚拟字段功能。类似于 Relation 字段的设计，用户需要**显式配置**来创建一个虚拟字段，而不是自动生成隐藏字段。
>
> 功能说明：
> - 用户在设计态显式添加一个虚拟字段（如 `statusLabel`），配置为连接到枚举源字段（如 `status`）
> - 虚拟字段在设计态可见，用户可以看到并配置它
> - 虚拟字段在运行时查询时返回 `{ code, label, description }` 结构
> - 单选枚举源字段返回单个对象，多选枚举源字段返回对象数组
> - 虚拟字段仅用于查询（只读），不在输入类型（create/update）中出现
>
> 示例配置：
> ```json
> {
>   "name": "statusLabel",
>   "title": "状态标签",
>   "format": "ENUM_LABEL",
>   "enumLabelConfig": {
>     "sourceField": "status"
>   }
> }
> ```

### Task 2.0: Database Schema Migration (数据库 / Database)
- [x] Create migration file for adding `enum_label_config` column to `field_definitions` table
  ```sql
  ALTER TABLE `field_definitions`
  ADD COLUMN `enum_label_config` JSON NULL COMMENT '枚举标签虚拟字段配置（ENUM_LABEL 格式使用）';
  ```
- [x] Update `db/schema/mysql/03_model_domain.sql` to include the new column definition
- [x] Run migration against development database
- [x] Verify column is added correctly with `DESCRIBE field_definitions;`

### Task 2.1: Add ENUM_LABEL Format Type (设计态 / Design-Time)
- [x] Add `FormatEnumLabel FormatType = "ENUM_LABEL"` to `internal/domain/modeldesign/field_definition.go`
- [x] Add to `fieldTypeMap` with appropriate `SchemaType` and `Title`
- [x] Update `api/graph/schema/field.graphql` to include `ENUM_LABEL` in `FormatType` enum

### Task 2.2: Add EnumLabelConfig Domain Model (设计态 / Design-Time)
- [x] Create `EnumLabelConfig` struct in `internal/domain/modeldesign/field_definition.go`
  ```go
  type EnumLabelConfig struct {
      SourceField string `json:"sourceField"` // 源字段名称（枚举字段）
  }
  ```
- [x] Add `EnumLabelConfig` field to `FieldDefinition` struct
  ```go
  EnumLabelConfig *EnumLabelConfig `json:"enumLabelConfig,omitempty"`
  ```
- [x] Add validation method for `EnumLabelConfig`:
  - Validate `sourceField` is not empty
  - Validate source field exists in the model
  - Validate source field is an Enum field (ENUM or ENUM_ARRAY format)
  - Validate source field has enumName in metadata

### Task 2.3: Add EnumLabel DTOs (设计态 / Design-Time)
- [x] Add `EnumLabelConfigInput` to `api/graph/schema/field.graphql`
  ```graphql
  input EnumLabelConfigInput {
    sourceField: String!
  }
  ```
- [x] Add `enumLabelConfig: EnumLabelConfigInput` to `AddFieldInput` in schema
- [x] Regenerate GraphQL code with `make generate-gql`

### Task 2.4: Update Field Service (设计态 / Design-Time)
- [x] Update `field_service.go` to handle `ENUM_LABEL` field creation
- [x] Validate enum label field configuration
- [x] Ensure enum label fields are read-only (not part of input types in runtime schema)

### Task 2.5: Update Field Mapper (设计态 / Design-Time)
- [x] Update `field_mapper.go` to map `EnumLabelConfigInput` to domain `EnumLabelConfig`
- [x] Update mappers in HTTP layer DTOs

### Task 2.6: Update Resolvers (设计态 / Design-Time)
- [x] Update enum label field resolvers in `field.resolvers.go` to handle the new config

### Task 2.7: Add EnumLabel Value Object (运行态 / Runtime)
- [x] Create `EnumLabel` struct in `internal/modelruntime/enum_label.go`
  ```go
  type EnumLabel struct {
      Code        string `json:"code"`
      Label       string `json:"label"`
      Description string `json:"description,omitempty"`
  }
  ```

### Task 2.8: Add EnumLabel GraphQL Object Type (运行态 / Runtime)
- [x] Add `EnumLabel` GraphQL type to runtime schema generation
  ```go
  // 在 schema manager 中添加
  typeEnumLabel := graphql.NewObject(graphql.ObjectConfig{
      Name: "EnumLabel",
      Fields: graphql.Fields{
          "code": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
          "label": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
          "description": &graphql.Field{Type: graphql.String},
      },
  })
  ```

### Task 2.9: Update Runtime Schema Field Generation (运行态 / Runtime)
- [x] Update `generateQueryFields()` in runtime schema manager to handle `ENUM_LABEL` format
- [x] For enum label fields:
  - Get source field from `EnumLabelConfig.SourceField`
  - Single select source → return single EnumLabel
  - Multi select source → return []EnumLabel
- [x] Ensure enum label fields are NOT added to input types (create, update)

### Task 2.10: Implement Enum Label Resolver (运行态 / Runtime)
- [x] Implement label field resolver in `internal/modelruntime/model_resolver.go`
  - For single select: query enum definition by `sourceField.Metadata["enumName"]`, find option by code
  - For multi select: query enum definition, find options by codes
  - Handle missing enum definition gracefully (return null)
  - Handle invalid code gracefully (return null for that item)
- [x] Add enum definition caching to avoid repeated queries

### Task 2.11: Update Runtime Field Loading (运行态 / Runtime)
- [x] Ensure `EnumLabelConfig` is loaded when creating `RuntimeField`
- [x] Pass enum label config to schema generation

### Task 2.12: Write Unit Tests
- [x] Test enum label field validation:
  - Valid config with valid source field
  - Empty source field
  - Source field doesn't exist
  - Source field is not an enum field
  - Source field has no enumName
- [x] Test enum label resolver:
  - Single-select with valid code
  - Single-select with invalid code
  - Multi-select with multiple codes
  - Multi-select with invalid code
  - Missing enum definition

### Task 2.13: Write Integration Tests
- [x] Test complete flow:
  1. Create enum definition
  2. Create model with enum field
  3. Create enum label virtual field pointing to enum field
  4. Add record with enum value
  5. Query with label field, verify code/label/description
- [x] Test with single-select enum
- [x] Test with multi-select enum
- [x] Verify label field is read-only (not in create/update input types)
- [x] Verify label field appears in schema introspection

### Task 2.14: Performance Testing
- [x] Benchmark query time with label field vs without
- [x] Verify enum definition caching works
- [x] Test with models containing multiple enum label fields

### Task 2.15: Update Documentation
- [x] Add inline documentation for enum label virtual field
- [x] Document how to create enum label field in design-time
- [x] Add example GraphQL queries with enum label fields
- [x] Update Design-Time API docs if any

**Dependency**: Task 2.16 depends on all previous Phase 2 tasks

### Task 2.16: Phase 2 Verification
- [x] Run all tests: `make test`
- [x] Run integration tests: pytest in `tests/automated/`
- [x] Manual test:
  1. Create enum with code/label options
  2. Create model with enum field
  3. Create enum label virtual field
  4. Create records with enum values
  5. Query with label field, verify results
- [x] Verify no regressions in existing functionality
- [x] Clean up any temporary code or debugging output

## Final Tasks

### Task 3.1: Full System Verification
- [x] Run complete test suite (unit + integration)
- [x] Manual end-to-end test:
  1. Create enum with code/label (Phase 1)
  2. Create model with enum field
  3. Create enum label virtual field (Phase 2)
  4. Add records with enum values
  5. Query with original enum field and label field
- [x] Verify both Phase 1 and Phase 2 work together
- [x] Verify no regressions in non-enum functionality

### Task 3.2: Code Review
- [x] Review all changes for consistency
- [x] Ensure no hardcoded `key`/`value` references remain
- [x] Verify error messages use correct terminology
- [x] Check for TODO comments or incomplete features
- [x] Verify enum label virtual field follows same patterns as Relation field

### Task 3.3: Git Commit Preparation
- [x] Stage all changes
- [x] Create comprehensive commit message describing both phases
- [x] Ensure commit follows project conventions
