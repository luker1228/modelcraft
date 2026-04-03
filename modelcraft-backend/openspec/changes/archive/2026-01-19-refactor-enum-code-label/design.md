# Design: Refactor Enum Feature - Code/Label Rename and Relation Support

## Overview

This document describes the architectural decisions and implementation approach for refactoring the enum feature to use `code`/`label` terminology and adding support for enum relation queries.

## Current State Analysis

### Existing Enum Structure
```go
// internal/domain/modeldesign/enum_definition.go
type EnumOption struct {
    Key         string `json:"key"`
    Value       string `json:"value"`
    Order       int32  `json:"order"`
    Description string `json:"description,omitempty"`
}

type EnumDefinition struct {
    ID            string       `json:"id"`
    ProjectID     string       `json:"projectId"`
    Name          string       `json:"name"`
    Title         string       `json:"title"`
    Options       []EnumOption `json:"options"`
    IsMultiSelect bool         `json:"isMultiSelect"`
    // ...
}
```

### Current Data Flow
1. **Design-Time**: Enum definitions stored in `model_enums` table with JSON `options` array
2. **Field Association**: `model_field_enum_associations` links fields to enum definitions
3. **Runtime**: Enum fields store the key (code) as string in physical database columns
4. **Query**: Runtime queries return only the code; clients must call design-time API to get labels

### Current GraphQL Schema
```graphql
type EnumOption {
  key: String!
  value: String!
  order: Int!
  description: String
}

type EnumDefinition {
  id: ID!
  projectId: ID!
  name: String!
  title: String!
  options: [EnumOption!]!
  # ...
}
```

## Proposed Changes

### Phase 1: Rename key/value to code/label

#### Domain Layer Changes
```go
// internal/domain/modeldesign/enum_definition.go
type EnumOption struct {
    Code        string `json:"code"`        // Renamed from Key
    Label       string `json:"label"`       // Renamed from Value
    Order       int32  `json:"order"`
    Description string `json:"description,omitempty"`
}
```



#### GraphQL Schema Changes
```graphql
api/graph/schema/enum.graphql
api/graph/schema/field.graphql
```

Changes:
- `EnumOption.key` → `EnumOption.code`
- `EnumOption.value` → `EnumOption.label`
- `EnumOptionInput.key` → `EnumOptionInput.code`
- `EnumOptionInput.value` → `EnumOptionInput.label`

#### Affected Files
1. `internal/domain/modeldesign/enum_definition.go`
2. `internal/infrastructure/repository/enum_definition_model.go`
3. `internal/app/modeldesign/enum_service.go`
4. `internal/interfaces/graphql/adapter/enum_mapper.go`
5. `internal/interfaces/http/dtos/field_definition.go`
6. `api/graph/schema/enum.graphql`
7. `api/graph/schema/field.graphql`
8. Test files: All files referencing `Key` or `Value` in enum context

### Phase 2: Enum Relation Support

#### Problem Statement
Currently, runtime queries for enum fields return only the code:
```graphql
query {
  findManyUser {
    status  # Returns: "ACTIVE"
  }
}
```

To get the label, clients must:
1. Know the enum definition name (from design-time API)
2. Call design-time API to get full enum definition
3. Look up the code to get the label

This is inefficient and requires two API calls.

#### Proposed Solution
Add automatic enum label field generation in runtime schema:

```graphql
query {
  findManyUser {
    status         # Returns: "ACTIVE"
    statusLabel {  # Auto-generated field
      code        # Returns: "ACTIVE"
      label       # Returns: "Active"
      description # Returns: "User is active"
    }
  }
}
```

#### Implementation Design

##### 1. Enum Label Value Object
```go
// internal/modelruntime/enum_label.go
type EnumLabel struct {
    Code        string `json:"code"`
    Label       string `json:"label"`
    Description string `json:"description,omitempty"`
}
```

##### 2. Runtime Field Extension
For each enum field in a model, automatically generate a `{fieldName}Label` field.

```go
// internal/domain/modelruntime/graphqlschema_manager.go
func (g *GraphQLSchemaGenerator) generateFieldEnumLabel(
    field *RuntimeField,
) *graphql.Field {
    // Field name: fieldName + "Label"
    labelFieldName := field.Name + "Label"

    return &graphql.Field{
        Type:        g.getEnumLabelScalarType(),
        Description: fmt.Sprintf("Enum label for %s field", field.Name),
        Resolve: g.resolveEnumLabel(field),
    }
}

func (g *GraphQLSchemaGenerator) resolveEnumLabel(field *RuntimeField) graphql.FieldResolveFn {
    return func(p graphql.ResolveParams) (interface{}, error) {
        // Get source value (the enum code)
        source := p.Source.(map[string]interface{})
        codeValue, ok := source[field.Name]
        if !ok || codeValue == nil {
            return nil, nil
        }

        code, ok := codeValue.(string)
        if !ok {
            return nil, nil
        }

        // Get enum definition from field
        enumDef := field.Enum
        if enumDef == nil {
            return nil, nil
        }

        // Find option by code
        for _, opt := range enumDef.Options {
            if opt.Code == code {
                return map[string]interface{}{
                    "code":        opt.Code,
                    "label":       opt.Label,
                    "description": opt.Description,
                }, nil
            }
        }

        return nil, nil
    }
}
```

##### 3. Custom Scalar for Enum Label
```go
// internal/domain/modelruntime/graphql_scalars.go
var QLEnumLabel = graphql.NewScalar(graphql.ScalarConfig{
    Name:        "EnumLabel",
    Description: "Scalar type representing an enum label with code, label, and optional description",
    Serialize: func(value interface{}) interface{} {
        if value == nil {
            return nil
        }
        return value // Already in map format
    },
    ParseValue: func(value interface{}) interface{} {
        return value
    },
    ParseLiteral: func(valueAST ast.Value) interface{} {
        return nil // Read-only field
    },
})
```

##### 4. Modified RuntimeField Loading
When loading a model, the `Enum` field in `FieldDefinition` must be populated:

```go
// internal/domain/modeldesign/field_builder.go or similar
func ensureEnumLoaded(field *FieldDefinition, enumRepo EnumRepository) error {
    if field.Type.Format != FormatEnum && field.Type.Format != FormatEnumArray {
        return nil
    }

    // Load enum definition if not already loaded
    if field.Enum == nil {
        // Get enum name from metadata or association
        enumName, ok := field.Metadata["enumName"].(string)
        if !ok {
            return nil // No enum association
        }

        // Load enum definition
        enumDef, err := enumRepo.FindByName(field.ModelLocator.ProjectID, enumName)
        if err != nil {
            return err
        }

        field.Enum = enumDef
    }

    return nil
}
```

#### Data Flow with Enum Labels

```
┌─────────────────┐
│   GraphQL Query │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────┐
│ Runtime Schema Generator     │
│ - Auto-generates <field>Label │
└────────┬────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│ Query Executor              │
│ - Fetches physical data     │
│ - Resolves label from enum  │
└────────┬────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│ Customer Database           │
│ - Stores enum code only     │
└─────────────────────────────┘
```

#### Performance Considerations

1. **Enum Definition Caching**:
   - Cache enum definitions at model load time
   - Cache shared across all model instances

2. **Single-Query Resolution**:
   - Labels resolved in-memory from cached enum definition
   - No additional database queries for labels

3. **Optional Loading**:
   - Labels only loaded when field is queried
   - No impact on queries that don't request labels

#### Backward Compatibility

For ENUM_ARRAY fields, the label field should return an array of labels:

```graphql
query {
  findManyUser {
    tags         # Returns: ["ADMIN", "USER"]
    tagsLabel {  # Auto-generated field
      code        # Returns: "ADMIN"
      label       # Returns: "Administrator"
      description # Returns: "System administrator"
    }
  }
}
```

## Implementation Order

### Phase 1: Code/Label Rename (sequential)
1. Update domain model `EnumOption` struct
2. Update repository models and converters
3. Update mappers (DTOs, GraphQL adapters)
4. Regenerate GraphQL schema (`make generate-gql`)
5. Update tests
6. Verify all tests pass

### Phase 2: Enum Relation Support (sequential)
1. Add `EnumLabel` scalar type
2. Update runtime schema generator to create label fields
3. Implement label resolution logic
4. Update model loading to include enum definitions
5. Add integration tests for label queries
6. Update documentation

## Testing Strategy

### Unit Tests
- Enum option validation with `code`/`label` fields
- Enum definition JSON serialization/deserialization
- Mapper conversions between layers

### Integration Tests
- Create enum with new field names via GraphQL
- Query enum definition via GraphQL
- Runtime query with enum label field
- ENUM_ARRAY label field returns array



## API Breaking Changes

### Breaking Changes
- GraphQL field names: `EnumOption.key` → `EnumOption.code`, `value` → `label`
- JSON field names in enum definitions: `key` → `code`, `value` → `label`

### Migration Guide for Clients
```diff
# Before
enum(name: "UserStatus") {
  options {
    key      # Returns: "ACTIVE"
    value    # Returns: "Active"
  }
}

# After
enum(name: "UserStatus") {
  options {
    code     # Returns: "ACTIVE"
    label    # Returns: "Active"
  }
}
```

### New Feature (Non-Breaking)
```graphql
# New label field in runtime queries
query {
  findManyUser {
    status        # Existing: returns "ACTIVE"
    statusLabel { # New: returns {code, label, description}
      code        # "ACTIVE"
      label       # "Active"
      description # "User is active"
    }
  }
}
```

## Rollback Plan

If issues arise during deployment:

1. **Revert Code**: Rollback git commit with code changes
2. **Fallback API**: Consider maintaining deprecated field names with deprecation warnings

## Future Considerations

1. **i18n Support**: Extend label field to support multiple languages
   ```go
   type EnumOptionLabel struct {
       Code  string            `json:"code"`
       Labels map[string]string `json:"labels"` // "en": "Active", "zh": "激活"
   }
   ```

2. **Enum Versioning**: Support for enum definition versioning to handle schema evolution

3. **Enum Metadata**: Additional metadata on enum options (icon, color, etc.)

## References
- Current enum spec: `openspec/specs/modeldesign-field-types/spec.md`
- GraphQL generation: internal use of `gqlgen v0.17.83`
