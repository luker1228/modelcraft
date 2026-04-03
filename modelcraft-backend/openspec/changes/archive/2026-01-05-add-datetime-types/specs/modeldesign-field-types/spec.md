# modeldesign-field-types Specification

## Purpose
Defines validation rules and constraints for Date, Time, and DateTime field types in ModelCraft field definitions, including format validation and range constraints.

## ADDED Requirements

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
