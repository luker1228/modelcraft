# Implementation Tasks

## 1. Database Layer

- [x] 1.1 Create migration script for `model_enums` table (SKIPPED - migrations not required)
- [x] 1.2 Create migration script to alter `field_definitions` table (SKIPPED - migrations not required)
- [x] 1.3 Test migration scripts on development database (SKIPPED)
- [x] 1.4 Prepare rollback migration scripts (SKIPPED)

## 2. Domain Layer - Enum Definition

- [x] 2.1 Create `internal/domain/modeldesign/enum_definition.go`
  - [x] 2.1.1 Define `EnumOption` value object with Key, Value, Order, Description fields
  - [x] 2.1.2 Implement `EnumOption.Validate()` with non-empty key/value validation
  - [x] 2.1.3 Define `EnumDefinition` aggregate root with ID, Name, Title, Description, Options, IsMultiSelect
  - [x] 2.1.4 Implement `EnumDefinition.Validate()` with name, options, and uniqueness checks
- [x] 2.2 Create `internal/domain/modeldesign/enum_repository.go` interface
  - [x] 2.2.1 Define `Create(enum *EnumDefinition) error`
  - [x] 2.2.2 Define `Update(enum *EnumDefinition) error`
  - [x] 2.2.3 Define `Delete(name string) error`
  - [x] 2.2.4 Define `FindByName(name string) (*EnumDefinition, error)`
  - [x] 2.2.5 Define `FindByID(id string) (*EnumDefinition, error)`
  - [x] 2.2.6 Define `List() ([]*EnumDefinition, error)`
  - [x] 2.2.7 Define `IsReferencedByFields(name string) (bool, []string, error)`
- [ ] 2.3 Write unit tests for enum domain models

## 3. Domain Layer - Field Definition Extension

- [x] 3.1 Update `internal/domain/modeldesign/field_definition.go`
  - [x] 3.1.1 Add `EnumName *string` field to FieldDefinition struct
  - [x] 3.1.2 Add `Enum *EnumDefinition` field to FieldDefinition struct
  - [x] 3.1.3 Update `Validate()` to check enum field requirements
  - [x] 3.1.4 Add validation for enumName and enumValues mutual exclusivity
  - [x] 3.1.5 Add validation for enum format type matching
- [ ] 3.2 Update field validation tests

## 4. Infrastructure Layer - Repository Implementation

- [x] 4.1 Create `internal/infrastructure/repository/enum_definition_model.go`
  - [x] 4.1.1 Define sqlc model matching `model_enums` table schema
  - [x] 4.1.2 Implement JSON marshaling/unmarshaling for options field
- [x] 4.2 Create `internal/infrastructure/repository/enum_definition_repository.go`
  - [x] 4.2.1 Implement `Create` method with name uniqueness check
  - [x] 4.2.2 Implement `Update` method with validation
  - [x] 4.2.3 Implement `Delete` method with reference checking
  - [x] 4.2.4 Implement `FindByName` method
  - [x] 4.2.5 Implement `FindByID` method
  - [x] 4.2.6 Implement `List` method
  - [x] 4.2.7 Implement `IsReferencedByFields` by querying field_definitions
- [x] 4.3 Update `internal/infrastructure/repository/field_definition_model.go`
  - [x] 4.3.1 Add `EnumName *string` field to sqlc model
- [x] 4.4 Update `internal/infrastructure/repository/field_definition_repository.go`
  - [x] 4.4.1 Update queries to include enum_name field (handled by sqlc auto-mapping)
  - [x] 4.4.2 Implement method to load enum details when querying fields (deferred to service layer)
- [ ] 4.5 Write repository integration tests

## 5. Field Type System Extension

- [x] 5.1 Update `internal/domain/modeldesign/field_definition.go`
  - [x] 5.1.1 Add `FormatEnum FormatType = "ENUM"` constant
  - [x] 5.1.2 Add `FormatEnumArray FormatType = "ENUM_ARRAY"` constant
  - [x] 5.1.3 Register enum types in fieldTypeMap initialization
  - [x] 5.1.4 Set ENUM SchemaType to string, title to "枚举(单选)"
  - [x] 5.1.5 Set ENUM_ARRAY SchemaType to array, title to "枚举(多选)"
- [ ] 5.2 Update field type tests

## 6. Application Layer - Enum Service

- [x] 6.1 Create `internal/app/modeldesign/enum_service.go`
  - [x] 6.1.1 Implement `CreateEnum(name, title, description string, options []EnumOption, isMultiSelect bool) (*EnumDefinition, error)`
  - [x] 6.1.2 Implement `UpdateEnum(name string, title, description *string, options []EnumOption) error`
  - [x] 6.1.3 Implement `DeleteEnum(name string) error` with reference check
  - [x] 6.1.4 Implement `GetEnum(name string) (*EnumDefinition, error)`
  - [x] 6.1.5 Implement `ListEnums() ([]*EnumDefinition, error)`
  - [x] 6.1.6 Implement `GetEnumReferences(name string) ([]string, error)`
- [ ] 6.2 Write service unit tests

## 7. GraphQL API Layer

- [x] 7.1 Update GraphQL schema definition
  - [x] 7.1.1 Add `EnumOption` type with key, value, order, description fields
  - [x] 7.1.2 Add `EnumDefinition` type
  - [x] 7.1.3 Add `enumName` field to FieldDefinition type
  - [x] 7.1.4 Add enum queries: `enum(name: String!)`, `enums`
  - [x] 7.1.5 Add enum mutations: `createEnum`, `updateEnum`, `deleteEnum`
  - [x] 7.1.6 Add ENUM and ENUM_ARRAY to FormatType enum
- [x] 7.2 Update `internal/interfaces/graphql/adapter/field_mapper.go`
  - [x] 7.2.1 Map `enumName` field in DTO conversions
  - [x] 7.2.2 Handle enum format types in convertFormatType2Domain
  - [x] 7.2.3 Handle enum format types in convertFormatTypeDomain2Model
- [x] 7.3 Create enum resolvers
  - [x] 7.3.1 Implement query resolvers (Enum, Enums, EnumReferences)
  - [x] 7.3.2 Implement mutation resolvers (CreateEnum, UpdateEnum, DeleteEnum)
- [x] 7.4 Update resolver dependency injection
  - [x] 7.4.1 Add EnumService to Resolver struct
- [x] 7.5 Update DTO layer
  - [x] 7.5.1 Add EnumName field to FieldDefinitionDTO
- [x] 7.6 Run `make generate-gql` to regenerate GraphQL code (COMPLETED - GraphQL code generated and fixed)
- [ ] 7.7 Write GraphQL integration tests

## 8. Field Validation Logic

- [x] 8.1 Create enum value validator
  - [x] 8.1.1 Implement single-select enum key validation (ValidateEnumValue)
  - [x] 8.1.2 Implement multi-select enum keys validation (ValidateEnumArrayValue)
  - [x] 8.1.3 Handle both enumName and enumValues validation paths
- [x] 8.2 Update field validation service to use enum validator
  - [x] 8.2.1 Add ValidateEnumField to check enum field configuration
  - [x] 8.2.2 Integrate ValidateEnumField into ValidateAll
- [ ] 8.3 Write validation tests for various enum scenarios

## 9. Import/Export Support

- [ ] 9.1 Update model export logic
  - [ ] 9.1.1 Include enumName in field export (not enumId)
  - [ ] 9.1.2 Collect and export all referenced enum definitions
  - [ ] 9.1.3 Export enum definitions with name, title, description, options, isMultiSelect
- [ ] 9.2 Update model import logic
  - [ ] 9.2.1 Import enum definitions first, keyed by name
  - [ ] 9.2.2 Check for existing enums by name
  - [ ] 9.2.3 Implement conflict resolution strategies (skip/overwrite/rename)
  - [ ] 9.2.4 Import fields with enumName references
- [ ] 9.3 Write import/export integration tests

## 10. Documentation

- [ ] 10.1 Update API documentation
  - [ ] 10.1.1 Document enum definition APIs
  - [ ] 10.1.2 Document enum field usage
  - [ ] 10.1.3 Document import/export with enums
- [ ] 10.2 Create migration guide
  - [ ] 10.2.1 Guide for migrating from enumValues to enumName
  - [ ] 10.2.2 Document when to use each approach
- [ ] 10.3 Update design documentation with implementation details

## 11. Testing and Validation

- [ ] 11.1 Run all unit tests
- [ ] 11.2 Run all integration tests
- [ ] 11.3 Test enum CRUD operations via GraphQL
- [ ] 11.4 Test field creation with enum references
- [ ] 11.5 Test data storage and retrieval with enum fields
- [ ] 11.6 Test validation of invalid enum keys
- [ ] 11.7 Test enum deletion with field references
- [ ] 11.8 Test model import/export with enums
- [ ] 11.9 Test backward compatibility with existing enumValues
- [ ] 11.10 Performance testing for enum queries

## 12. Deployment

- [ ] 12.1 Run database migrations in staging environment
- [ ] 12.2 Deploy code to staging
- [ ] 12.3 Verify enum functionality in staging
- [ ] 12.4 Run database migrations in production
- [ ] 12.5 Deploy code to production
- [ ] 12.6 Monitor for errors and performance issues
