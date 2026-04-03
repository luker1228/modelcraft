# modeldesign-field-types Specification

## Purpose
TBD - created by archiving change add-datetime-types. Update Purpose after archive.
## Requirements
### Requirement: Date Format Validation

The system SHALL validate Date field values conform to ISO 8601 date format.

#### Scenario: Valid ISO 8601 date format

- **WHEN** a Date field value is provided in the format `YYYY-MM-DD` (e.g., "2024-01-15")
- **THEN** the validation passes
- **AND** the date is accepted for storage

#### Scenario: Invalid date format rejected

- **WHEN** a Date field value does not match `YYYY-MM-DD` format (e.g., "15/01/2024", "2024-1-5", "2024-01-15T14:30:00Z")
- **THEN** the validation fails with a format error
- **AND** the error message indicates the expected format is `YYYY-MM-DD`

#### Scenario: Invalid date values rejected

- **WHEN** a Date field value has invalid date components (e.g., "2024-13-01", "2024-02-30")
- **THEN** the validation fails with an invalid date error
- **AND** the error message indicates the date is not valid

### Requirement: Time Format Validation

The system SHALL validate Time field values conform to 24-hour time format.

#### Scenario: Valid 24-hour time format

- **WHEN** a Time field value is provided in the format `HH:MM:SS` (e.g., "14:30:00", "00:00:00", "23:59:59")
- **THEN** the validation passes
- **AND** the time is accepted for storage

#### Scenario: Invalid time format rejected

- **WHEN** a Time field value does not match `HH:MM:SS` format (e.g., "2:30 PM", "14:30", "14:30:00.000")
- **THEN** the validation fails with a format error
- **AND** the error message indicates the expected format is `HH:MM:SS`

#### Scenario: Invalid time values rejected

- **WHEN** a Time field value has invalid time components (e.g., "25:00:00", "14:60:00", "14:30:60")
- **THEN** the validation fails with an invalid time error
- **AND** the error message indicates the time is not valid

### Requirement: DateTime Format Validation

The system SHALL validate DateTime field values conform to ISO 8601 date-time format with timezone.

#### Scenario: Valid ISO 8601 datetime format

- **WHEN** a DateTime field value is provided in ISO 8601 format with timezone (e.g., "2024-01-15T14:30:00Z", "2024-01-15T14:30:00+08:00", "2024-01-15T14:30:00.123Z")
- **THEN** the validation passes
- **AND** the datetime is accepted for storage

#### Scenario: Invalid datetime format rejected

- **WHEN** a DateTime field value does not match ISO 8601 format (e.g., "2024-01-15 14:30:00", "2024-01-15T14:30:00", "01/15/2024 14:30:00")
- **THEN** the validation fails with a format error
- **AND** the error message indicates the expected format is ISO 8601 with timezone

#### Scenario: Invalid datetime values rejected

- **WHEN** a DateTime field value has invalid components (e.g., "2024-13-01T14:30:00Z", "2024-01-32T14:30:00Z", "2024-01-15T25:00:00Z")
- **THEN** the validation fails with an invalid datetime error
- **AND** the error message indicates the datetime is not valid

### Requirement: Date Range Validation

The system SHALL support minimum and maximum date constraints for Date and DateTime fields.

#### Scenario: Extend ValidationConfig with date range fields

- **WHEN** defining a Date or DateTime field with range constraints
- **THEN** the ValidationConfig accepts `minDate` and `maxDate` fields in ISO 8601 format
- **AND** both fields are optional
- **AND** `minDate` must be less than or equal to `maxDate` if both are specified

#### Scenario: Date value within valid range

- **WHEN** a Date field has `minDate: "2024-01-01"` and `maxDate: "2024-12-31"`
- **AND** a date value "2024-06-15" is provided
- **THEN** the validation passes
- **AND** the date is accepted

#### Scenario: Date value below minimum rejected

- **WHEN** a Date field has `minDate: "2024-01-01"`
- **AND** a date value "2023-12-31" is provided
- **THEN** the validation fails with a range error
- **AND** the error message indicates the date must be on or after "2024-01-01"

#### Scenario: Date value above maximum rejected

- **WHEN** a Date field has `maxDate: "2024-12-31"`
- **AND** a date value "2025-01-01" is provided
- **THEN** the validation fails with a range error
- **AND** the error message indicates the date must be on or before "2024-12-31"

#### Scenario: DateTime range validation

- **WHEN** a DateTime field has `minDate: "2024-01-01T00:00:00Z"` and `maxDate: "2024-12-31T23:59:59Z"`
- **AND** a datetime value is provided
- **THEN** the validation checks if the datetime is within the specified range
- **AND** rejects values outside the range with appropriate error messages

### Requirement: Time Range Validation

The system SHALL support minimum and maximum time constraints for Time fields.

#### Scenario: Extend ValidationConfig with time range fields

- **WHEN** defining a Time field with range constraints
- **THEN** the ValidationConfig accepts `minTime` and `maxTime` fields in `HH:MM:SS` format
- **AND** both fields are optional
- **AND** `minTime` must be less than or equal to `maxTime` if both are specified

#### Scenario: Time value within valid range

- **WHEN** a Time field has `minTime: "09:00:00"` and `maxTime: "18:00:00"`
- **AND** a time value "14:30:00" is provided
- **THEN** the validation passes
- **AND** the time is accepted

#### Scenario: Time value below minimum rejected

- **WHEN** a Time field has `minTime: "09:00:00"`
- **AND** a time value "08:59:59" is provided
- **THEN** the validation fails with a range error
- **AND** the error message indicates the time must be at or after "09:00:00"

#### Scenario: Time value above maximum rejected

- **WHEN** a Time field has `maxTime: "18:00:00"`
- **AND** a time value "18:00:01" is provided
- **THEN** the validation fails with a range error
- **AND** the error message indicates the time must be at or before "18:00:00"

#### Scenario: Time range spans midnight

- **WHEN** a Time field has `minTime: "22:00:00"` and `maxTime: "06:00:00"` (overnight range)
- **THEN** the validation accepts this configuration
- **AND** values between 22:00:00-23:59:59 or 00:00:00-06:00:00 pass validation
- **AND** values between 06:00:01-21:59:59 fail validation

### Requirement: ValidationConfig Structure Extension

The ValidationConfig struct SHALL be extended to support date and time range constraints.

#### Scenario: Add date range fields to ValidationConfig

- **WHEN** the ValidationConfig struct is defined
- **THEN** it includes `MinDate *string` field with JSON tag `minDate,omitempty`
- **AND** it includes `MaxDate *string` field with JSON tag `maxDate,omitempty`
- **AND** both fields are pointers to allow nil (no constraint)

#### Scenario: Add time range fields to ValidationConfig

- **WHEN** the ValidationConfig struct is defined
- **THEN** it includes `MinTime *string` field with JSON tag `minTime,omitempty`
- **AND** it includes `MaxTime *string` field with JSON tag `maxTime,omitempty`
- **AND** both fields are pointers to allow nil (no constraint)

#### Scenario: Backward compatibility

- **WHEN** existing field definitions without date/time constraints are loaded
- **THEN** the new fields default to nil
- **AND** no validation errors occur for existing definitions
- **AND** existing behavior is unchanged

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

