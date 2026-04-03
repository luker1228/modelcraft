# modeldesign-field-types Specification Delta

## ADDED Requirements

### Requirement: Enum Field Type Support

The system SHALL support enum field types with centralized enum definitions for single-select and multi-select scenarios.

#### Scenario: Create enum definition with valid options

- **WHEN** creating an enum definition with name, title, and valid options array
- **THEN** the enum definition is created successfully
- **AND** each option has a non-empty key and value
- **AND** the enum is assigned a unique ID and globally unique name

#### Scenario: Enum option key cannot be empty

- **WHEN** creating or updating an enum definition with an option that has an empty or whitespace-only key
- **THEN** the validation fails with error "enum option key cannot be empty"
- **AND** the enum definition is not created or updated

#### Scenario: Enum option value cannot be empty

- **WHEN** creating or updating an enum definition with an option that has an empty or whitespace-only value
- **THEN** the validation fails with error "enum option value cannot be empty"
- **AND** the enum definition is not created or updated

#### Scenario: Enum must have at least one option

- **WHEN** creating an enum definition with an empty options array
- **THEN** the validation fails with error "enum must have at least one option"
- **AND** the enum definition is not created

#### Scenario: Enum option keys must be unique within definition

- **WHEN** creating an enum definition with duplicate option keys
- **THEN** the validation fails with error indicating "duplicate enum option key"
- **AND** the enum definition is not created

#### Scenario: Enum name must be globally unique

- **WHEN** creating an enum definition with a name that already exists
- **THEN** the validation fails with error "enum name already exists"
- **AND** the enum definition is not created

### Requirement: Enum Field Format Types

The system SHALL provide ENUM and ENUM_ARRAY format types for single-select and multi-select enum fields.

#### Scenario: ENUM format type for single-select

- **WHEN** a field is defined with format type "ENUM"
- **THEN** the field has SchemaType "string" and Format "ENUM"
- **AND** the field title is "枚举(单选)"
- **AND** the field requires either an enumName or enumValues

#### Scenario: ENUM_ARRAY format type for multi-select

- **WHEN** a field is defined with format type "ENUM_ARRAY"
- **THEN** the field has SchemaType "array" and Format "ENUM_ARRAY"
- **AND** the field title is "枚举(多选)"
- **AND** the field requires either an enumName or enumValues

### Requirement: Field Enum Association by Name

The system SHALL allow fields to reference centralized enum definitions through an enumName field using the enum's business identifier.

#### Scenario: Associate field with enum definition by name

- **WHEN** defining a field with format "ENUM" and a valid enumName
- **THEN** the field is associated with the referenced enum definition
- **AND** the field can query the full enum details including options
- **AND** the association uses the enum's name (business identifier) not ID

#### Scenario: Enum reference validation

- **WHEN** defining a field with an enumName that references a non-existent enum
- **THEN** the validation fails with error "referenced enum does not exist"
- **AND** the field definition is not created

#### Scenario: Cannot use both enumName and enumValues

- **WHEN** defining an enum field with both enumName and enumValues set
- **THEN** the validation fails with error "cannot use both enumName and enumValues"
- **AND** the field definition is not created

#### Scenario: Enum field must have enum source

- **WHEN** defining a field with format "ENUM" or "ENUM_ARRAY" without enumName or enumValues
- **THEN** the validation fails with error "enum field must have either enumName or enumValues"
- **AND** the field definition is not created

### Requirement: Enum Data Storage

The system SHALL store enum keys in physical fields and provide full enum information through logical fields.

#### Scenario: Single-select enum stores key value

- **WHEN** a field with format "ENUM" stores data
- **THEN** the physical field stores the enum option key as a string value
- **AND** the logical field returns the full enum option object with key and value

#### Scenario: Multi-select enum stores key array

- **WHEN** a field with format "ENUM_ARRAY" stores data
- **THEN** the physical field stores the enum option keys as a JSON array
- **AND** the logical field returns an array of full enum option objects with keys and values

#### Scenario: Invalid enum key rejected for single-select

- **WHEN** storing a value in an ENUM field with a key not in the enum options
- **THEN** the validation fails with error "invalid enum key"
- **AND** the value is not stored

#### Scenario: Invalid enum keys rejected for multi-select

- **WHEN** storing values in an ENUM_ARRAY field with keys not all in the enum options
- **THEN** the validation fails with error "invalid enum keys"
- **AND** the values are not stored

### Requirement: Enum Definition Management

The system SHALL provide operations to create, update, query, and delete enum definitions.

#### Scenario: Create enum definition with unique name

- **WHEN** creating an enum definition with name "UserStatus", title "用户状态", and valid options
- **THEN** the enum definition is created with a unique ID
- **AND** the enum name is globally unique
- **AND** the enum can be referenced by fields using its name

#### Scenario: Update enum options

- **WHEN** updating an enum definition's options array
- **THEN** the new options are validated for non-empty keys and values
- **AND** the enum definition is updated if validation passes
- **AND** all fields referencing this enum use the new options immediately

#### Scenario: Delete enum definition with field references

- **WHEN** deleting an enum definition that is referenced by one or more fields
- **THEN** the deletion fails with error "enum is referenced by fields"
- **AND** the enum definition is not deleted
- **AND** the error message lists the fields referencing this enum

#### Scenario: Delete unreferenced enum definition

- **WHEN** deleting an enum definition that is not referenced by any field
- **THEN** the enum definition is deleted successfully

#### Scenario: Query enum definition by name

- **WHEN** querying an enum definition by its name
- **THEN** the full enum definition is returned including ID, name, title, description, and options array

### Requirement: Enum Options Structure

The system SHALL store enum options as a JSON array with key, value, order, and optional description fields.

#### Scenario: Enum option contains required fields

- **WHEN** an enum option is defined
- **THEN** it contains a non-empty "key" field for storage
- **AND** it contains a non-empty "value" field for display
- **AND** it contains an "order" field for display sorting

#### Scenario: Enum option contains optional description

- **WHEN** an enum option is defined with a description
- **THEN** the description is stored in the option
- **AND** the description is returned when querying the enum

#### Scenario: Enum options are ordered by order field

- **WHEN** querying enum options
- **THEN** the options are returned sorted by the "order" field in ascending order
- **AND** the order determines the display sequence in UI

### Requirement: Backward Compatibility with EnumValues

The system SHALL maintain backward compatibility with the existing EnumValues field in ValidationConfig.

#### Scenario: Existing EnumValues fields continue to work

- **WHEN** a field uses ValidationConfig.EnumValues without enumName
- **THEN** the field validation uses the EnumValues array
- **AND** the field behavior is unchanged from before

#### Scenario: Simple enum without centralized definition

- **WHEN** creating a field with format "ENUM" and ValidationConfig.EnumValues set
- **THEN** the field uses the simple enum values
- **AND** no enum definition is required

#### Scenario: Choose between simple and centralized enum

- **WHEN** deciding whether to use EnumValues or enumName
- **THEN** use EnumValues for simple, non-reusable, field-specific enums
- **AND** use enumName for reusable, centrally managed enums shared across fields

### Requirement: Enum Multi-Select Configuration

The system SHALL support configuring whether an enum definition allows multi-select through an isMultiSelect flag.

#### Scenario: Single-select enum configuration

- **WHEN** an enum definition has isMultiSelect set to false
- **THEN** fields referencing this enum with format "ENUM" are valid
- **AND** fields referencing this enum with format "ENUM_ARRAY" fail validation with error "enum does not support multi-select"

#### Scenario: Multi-select enum configuration

- **WHEN** an enum definition has isMultiSelect set to true
- **THEN** fields referencing this enum with format "ENUM_ARRAY" are valid
- **AND** fields referencing this enum with format "ENUM" are also valid for selecting single values

#### Scenario: Enum format matches configuration

- **WHEN** validating a field definition with enum reference
- **THEN** if field format is "ENUM_ARRAY", the referenced enum must have isMultiSelect true
- **AND** if field format is "ENUM", the enum can have isMultiSelect true or false

### Requirement: Enum Import and Export Support

The system SHALL support importing and exporting models with enum field references using enum names for portability.

#### Scenario: Export model with enum field references

- **WHEN** exporting a model that has fields with enumName references
- **THEN** the export includes the field's enumName (not enumId)
- **AND** the export includes all referenced enum definitions
- **AND** the enum definitions include name, title, description, options, and isMultiSelect

#### Scenario: Import model with enum definitions

- **WHEN** importing a model with enum field references
- **THEN** the system first imports the enum definitions by name
- **AND** if an enum with the same name exists, it is reused
- **AND** if an enum with the same name does not exist, it is created
- **AND** the field enumName references are established after enums are imported

#### Scenario: Import handles enum name conflicts

- **WHEN** importing an enum definition with a name that already exists but different options
- **THEN** the system provides options to skip, overwrite, or rename the enum
- **AND** the user can choose the appropriate conflict resolution strategy
- **AND** field references are updated according to the chosen strategy
