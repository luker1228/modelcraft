## ADDED Requirements

### Requirement: The system SHALL manage enum field associations through a dedicated association table

The system SHALL use a dedicated `model_field_enum_associations` table to record relationships between model fields and enum definitions, supporting foreign key constraints and cascade operations.

#### Scenario: Creating an enum field with association to existing enum

- **WHEN** defining a field with format "ENUM" or "ENUM_ARRAY" and `enumConfig.connectEnum = true`
- **AND** providing a valid `enumConfig.enumName` that references an existing enum
- **THEN** a record is created in the `model_field_enum_associations` table
- **AND** the association includes `cluster_name`, `database_name`, `model_id`, `field_name`, and `enum_id`
- **AND** the field can query the full enum details through the association

#### Scenario: Creating an enum field with new enum creation

- **WHEN** defining a field with format "ENUM" or "ENUM_ARRAY" and `enumConfig.connectEnum = false`
- **AND** providing `enumConfig.enumName`, `enumConfig.options`, and optional `enumConfig.description`
- **THEN** a new enum definition is created with the provided name and options
- **AND** a record is created in the `model_field_enum_associations` table linking the field to the new enum
- **AND** if an enum with the same name already exists, the validation fails with error "enum name already exists"

#### Scenario: Attempting to connect to non-existent enum

- **WHEN** defining a field with `enumConfig.connectEnum = true`
- **AND** the `enumConfig.enumName` references a non-existent enum
- **THEN** the validation fails with error "referenced enum does not exist"
- **AND** the field definition is not created

#### Scenario: Deleting a field with enum association

- **WHEN** a field with an enum association is deleted
- **THEN** the associated record in `model_field_enum_associations` is automatically deleted via CASCADE constraint
- **AND** the enum definition itself remains unchanged (not deleted)

#### Scenario: Attempting to delete an enum referenced by fields

- **WHEN** attempting to delete an enum definition
- **AND** the enum has associations in the `model_field_enum_associations` table
- **THEN** the deletion fails with error "enum is referenced by fields"
- **AND** the enum definition is not deleted
- **AND** the error message lists the fields referencing this enum

#### Scenario: Querying field with enum association

- **WHEN** querying a field definition through GraphQL
- **AND** the field has an enum association
- **THEN** the field's `enum` field is populated with the full enum definition
- **AND** the enum details include ID, name, title, description, options, and isMultiSelect

### Requirement: GraphQL API SHALL provide enumConfig for enum field configuration

The system SHALL provide a dedicated `enumConfig` input type for configuring enum fields, replacing the `enum` field in `validationConfig`.

#### Scenario: Adding enum field with enumConfig in GraphQL

- **WHEN** calling `addFields` mutation with a field of format "ENUM" or "ENUM_ARRAY"
- **THEN** the `AddFieldInput` accepts an `enumConfig` input object
- **AND** `enumConfig` has the following fields:
  - `enumName: String!` - the enum name (for connection or creation)
  - `options: [EnumOptionInput!]` - array of enum options (required when `connectEnum = false`)
  - `description: String` - optional enum description
  - `connectEnum: Boolean!` - whether to connect existing enum (true) or create new (false)

#### Scenario: Creating field with connectEnum true

- **WHEN** `enumConfig.connectEnum = true`
- **THEN** the system looks up an existing enum by `enumConfig.enumName`
- **AND** if found, creates an association in `model_field_enum_associations`
- **AND** if not found, returns error "enum not found: {enumName}"
- **AND** `enumConfig.options` and `enumConfig.description` are ignored

#### Scenario: Creating field with connectEnum false

- **WHEN** `enumConfig.connectEnum = false`
- **THEN** the system creates a new enum definition with `enumConfig.enumName`
- **AND** uses `enumConfig.options` to populate the enum options
- **AND** uses `enumConfig.description` for the enum description
- **AND** creates an association in `model_field_enum_associations` linking to the new enum
- **AND** if an enum with the same name exists, returns error "enum name already exists"

#### Scenario: EnumOptionInput structure

- **WHEN** providing enum options in `enumConfig.options`
- **THEN** each `EnumOptionInput` has the following fields:
  - `key: String!` - the option key for storage
  - `value: String!` - the option display value
  - `order: Int!` - the display order
  - `description: String` - optional option description

## REMOVED Requirements

### Requirement: Field Enum Association by Name

**Reason**: This requirement is being replaced with the MODIFIED version above that uses a dedicated association table instead of storing `enumName` directly in the field definition.

**Migration**: Existing `field_definitions.enum_name` data will be migrated to the new `model_field_enum_associations` table via database migration script.

#### Scenario: Defining a field with format "ENUM" and a valid enumName (REMOVED)

- This scenario is replaced by "Creating an enum field with association to existing enum" in the MODIFIED requirement

#### Scenario: Defining a field with an enumName that references a non-existent enum (REMOVED)

- This scenario is replaced by "Attempting to connect to non-existent enum" in the MODIFIED requirement

#### Scenario: Defining an enum field with both enumName and enumValues set (REMOVED)

- This scenario remains valid but is now checked against `enumConfig` vs `validationConfig.enumValues`

#### Scenario: Defining a field with format "ENUM" or "ENUM_ARRAY" without enumName or enumValues (REMOVED)

- This scenario remains valid but now checks for `enumConfig` vs `validationConfig.enumValues`
