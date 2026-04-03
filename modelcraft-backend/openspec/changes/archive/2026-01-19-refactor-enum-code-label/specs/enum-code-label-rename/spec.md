# Enum Code/Label Rename Specification

## Purpose
Renames enum option fields from `key`/`value` to `code`/`label` to align with industry-standard terminology and improve semantic clarity.

## Requirements

### Requirement: Enum Option Code Field

The system SHALL rename the `key` field of enum options to `code` across all layers.

#### Scenario: EnumOption struct uses Code field

- **WHEN** inspecting the `EnumOption` struct in the domain layer
- **THEN** it contains a `Code string` field with JSON tag `"code"`
- **AND** the `Key` field no longer exists

#### Scenario: Enum option validation requires non-empty code

- **WHEN** creating or updating an enum option
- **AND** the `Code` field is empty or whitespace-only
- **THEN** validation fails with error "enum option code cannot be empty"
- **AND** the enum definition is not created or updated

#### Scenario: GraphQL EnumOption type uses code field

- **WHEN** querying the GraphQL `EnumOption` type
- **THEN** the `code` field is available as required string type
- **AND** the `key` field is not available

### Requirement: Enum Option Label Field

The system SHALL rename the `value` field of enum options to `label` across all layers.

#### Scenario: EnumOption struct uses Label field

- **WHEN** inspecting the `EnumOption` struct in the domain layer
- **THEN** it contains a `Label string` field with JSON tag `"label"`
- **AND** the `Value` field no longer exists

#### Scenario: Enum option validation requires non-empty label

- **WHEN** creating or updating an enum option
- **AND** the `Label` field is empty or whitespace-only
- **THEN** validation fails with error "enum option label cannot be empty"
- **AND** the enum definition is not created or updated

#### Scenario: GraphQL EnumOption type uses label field

- **WHEN** querying the GraphQL `EnumOption` type
- **THEN** the `label` field is available as required string type
- **AND** the `value` field is not available

### Requirement: EnumOptionInput Reflects Code/Label

The system SHALL update input types to use `code` and `label` instead of `key` and `value`.

#### Scenario: EnumOptionInput for mutation

- **WHEN** using `EnumOptionInput` in GraphQL mutations (createEnum, updateEnum)
- **THEN** the input accepts `code: String!` and `label: String!` fields
- **AND** the `key` and `value` fields are not accepted

#### Scenario: DTOs use code and label fields

- **WHEN** using `EnumOptionDTO` in HTTP layer
- **THEN** the DTO contains `Code string` and `Label string` fields
- **AND** JSON serialization uses `"code"` and `"label"` keys

### Requirement: Mapper Conversions Handle New Field Names

The system SHALL update all mapper functions to use `code`/`label` terminology.

#### Scenario: GraphQL adapter converts to domain enum option

- **WHEN** converting GraphQL `EnumOption` to domain `EnumOption`
- **THEN** the `code` field maps to domain `Code` field
- **AND** the `label` field maps to domain `Label` field
- **AND** the `order` and `description` fields map unchanged

#### Scenario: GraphQL adapter converts from domain enum option

- **WHEN** converting domain `EnumOption` to GraphQL `EnumOption`
- **THEN** the domain `Code` field maps to GraphQL `code` field
- **AND** the domain `Label` field maps to GraphQL `label` field
- **AND** the `order` and `description` fields map unchanged

### Requirement: Existing Data Migration

The system SHALL migrate existing enum data from `key`/`value` to `code`/`label` structure.

#### Scenario: Migration updates stored enum definitions

- **WHEN** running the data migration
- **AND** existing `model_enums` records contain `options` JSON with `"key"` and `"value"` fields
- **THEN** all `"key"` fields are renamed to `"code"`
- **AND** all `"value"` fields are renamed to `"label"`
- **AND** the `order` and `description` fields remain unchanged

#### Scenario: Migration validates data integrity

- **WHEN** running the data migration
- **AND** any enum option has an empty `key` (which becomes `code`)
- **THEN** the migration fails with a descriptive error
- **AND** no data is modified for that enum

#### Scenario: Migration is idempotent

- **WHEN** running the data migration multiple times
- **AND** the first run successfully translates `key`/`value` to `code`/`label`
- **THEN** subsequent runs do not modify data that already uses `code`/`label`
- **AND** the migration completes successfully

### Requirement: Enum Definition Methods Use Code/Label

The system SHALL update utility methods on `EnumDefinition` to use new terminology.

#### Scenario: GetOptionByCode returns enum option

- **WHEN** calling `GetOptionByCode(code string)` on an enum definition
- **THEN** it returns the option matching the provided `code`
- **AND** returns error if no option matches
- **AND** the returned option contains `Code`, `Label`, `Order`, and `Description`

#### Scenario: HasOptionCode checks existence

- **WHEN** calling `HasOptionCode(code string)` on an enum definition
- **THEN** it returns `true` if an option with matching `code` exists
- **AND** returns `false` if no option matches

#### Scenario: ValidateCodes checks multiple codes

- **WHEN** calling `ValidateCodes([]string)` with multiple codes
- **THEN** it validates all codes exist in enum options
- **AND** returns error with details of first invalid code
- **AND** the error message uses "enum code" terminology

### Requirement: Documentation Reflects New Terminology

The system SHALL update all documentation to use `code`/`label` terminology.

#### Scenario: API docs describe code field

- **WHEN** reading API documentation for enum options
- **THEN** the `code` field is described as "programmatic identifier stored in database"
- **AND** examples show `code` (not `key`)

#### Scenario: API docs describe label field

- **WHEN** reading API documentation for enum options
- **THEN** the `label` field is described as "human-readable display value"
- **AND** examples show `label` (not `value`)

### Requirement: Tests Use New Terminology

The system SHALL update all tests to reference `code`/`label` fields.

#### Scenario: Unit tests create options with code and label

- **WHEN** writing unit tests for enum functionality
- **THEN** options are created with `Code` and `Label` fields
- **AND** assertions check `Code` and `Label` values
- **AND** no references to `Key` or `Value` remain

#### Scenario: Integration tests query code and label fields

- **WHEN** running GraphQL integration tests
- **THEN** queries request `code` and `label` fields on `EnumOption`
- **AND** assertions verify these fields contain correct values
