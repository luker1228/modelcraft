# Enum Relation Support Specification

## Purpose
Enables runtime queries to fetch enum label information alongside the stored code value through auto-generated enum label fields.

## Requirements

### Requirement: Auto-generated Enum Label Field

The system SHALL automatically generate a `{fieldName}Label` field for each enum field in the runtime GraphQL schema.

#### Scenario: Single-select enum generates label field

- **WHEN** a field has format type `ENUM`
- **AND** the runtime GraphQL schema is generated
- **THEN** a `{fieldName}Label` field is added to the output type
- **AND** the field returns an `EnumLabel` type containing code, label, and description

#### Scenario: Multi-select enum generates label field

- **WHEN** a field has format type `ENUM_ARRAY`
- **AND** the runtime GraphQL schema is generated
- **THEN** a `{fieldName}Label` field is added to the output type
- **AND** the field returns a list of `EnumLabel` types (one for each enum code in the array)

### Requirement: EnumLabel Scalar Type

The system SHALL provide an `EnumLabel` scalar type structure containing code, label, and optional description.

#### Scenario: EnumLabel contains all required fields

- **WHEN** inspecting the `EnumLabel` type structure
- **THEN** it contains `code: String` field for the stored identifier
- **AND** it contains `label: String` field for the human-readable display value
- **AND** it contains `description: String` (optional) field for additional context

#### Scenario: EnumLabel serialization returns object

- **WHEN** the GraphQL executor serializes an `EnumLabel` value
- **THEN** it returns a JSON object with `code`, `label`, and `description` keys
- **AND** the `description` is omitted if null

### Requirement: Label Field Resolution from Enum Definition

The system SHALL resolve enum label fields using the loaded enum definition associated with the field.

#### Scenario: Single-select label resolution

- **WHEN** a single-select enum field with value "ACTIVE" is queried
- **AND** the field's enum definition has `{code: "ACTIVE", label: "Active", description: "User is active"}`
- **THEN** the `statusLabel` field returns `{code: "ACTIVE", label: "Active", description: "User is active"}`

#### Scenario: Multi-select label resolution

- **WHEN** a multi-select enum field with value `["ADMIN", "USER"]` is queried
- **AND** the field's enum definition has options for both codes
- **THEN** the `tagsLabel` field returns an array of label objects
- **AND** the array order matches the order in the stored array
- **AND** each element contains full code, label, and description

#### Scenario: Invalid code returns null

- **WHEN** an enum field contains a code that doesn't exist in the associated enum definition
- **THEN** the label field resolves to `null`
- **AND** the query still returns the code field with the invalid value

### Requirement: Enum Definition Loading for Labels

The system SHALL load enum definitions when generating the runtime schema to support label resolution.

#### Scenario: Enum definition loaded at schema generation

- **WHEN** the runtime schema generator processes a model with enum fields
- **THEN** each enum field's associated enum definition is loaded
- **AND** the `Enum` field in `RuntimeField` is populated with the full definition
- **AND** no additional queries are needed at runtime for label resolution

#### Scenario: Cached enum definitions across queries

- **WHEN** multiple queries request label fields from the same enum
- **THEN** the enum definition is loaded once and cached
- **AND** subsequent label resolutions use the cached definition

### Requirement: Label Field is Read-Only

The system SHALL ensure enum label fields are read-only in the runtime schema.

#### Scenario: Label field not available in input types

- **WHEN** examining create/update input types for a model with enum fields
- **THEN** label fields (e.g., `{fieldName}Label`) are not included in input types
- **AND** only the code field is available for updates

#### Scenario: Label field in mutation resolve params

- **WHEN** a mutation includes a return type with enum fields
- **AND** the result includes the label field
- **THEN** the label field is resolved from the enum definition
- **AND** the label field cannot be directly set by the mutation

### Requirement: GraphQL Query with Enum Labels

The system SHALL enable querying enum labels alongside field values in a single GraphQL query.

#### Scenario: Query single enum field with label

```graphql
- WHEN** executing query:
```graphql
query {
  findManyUser {
    id
    status
    statusLabel {
      code
      label
      description
    }
  }
}
```
- **THEN** each user record contains:
  - `id`: the user identifier
  - `status`: the stored enum code (e.g., "ACTIVE")
  - `statusLabel`: object `{code: "ACTIVE", label: "Active", description: "User is active"}`

#### Scenario: Query multiple enum fields with labels

```graphql
- WHEN** executing query:
```graphql
query {
  findManyArticle {
    id
    title
    status
    statusLabel {
      label
    }
    category
    categoryLabel {
      label
    }
  }
}
```
- **THEN** each article record contains all requested fields
- **AND** each label field is resolved independently from its enum definition

#### Scenario: Enum array with labels

```graphql
- WHEN** executing query:
```graphql
query {
  findManyUser {
    id
    tags
    tagsLabel {
      code
      label
    }
  }
}
```
- **THEN** `tags` returns array of codes `["ADMIN", "MODERATOR"]`
- **AND** `tagsLabel` returns array of label objects matching the codes in order

### Requirement: Performance of Label Resolution

The system SHALL resolve enum labels efficiently without additional database queries.

#### Scenario: Label resolution uses cached enum definition

- **WHEN** a query requests enum labels
- **THEN** label resolution uses the enum definition cached at schema generation time
- **AND** no additional database queries are made for enum definitions

#### Scenario: Unrequested labels are not resolved

- **WHEN** a query requests enum fields but does not include label fields
- **THEN** no label resolution occurs
- **AND** query performance is identical to before label support was added

### Requirement: Error Handling for Label Resolution

The system SHALL handle errors gracefully during label resolution without failing the entire query.

#### Scenario: Missing enum definition returns null

- **WHEN** a field references an enum but the enum definition is not loaded
- **THEN** the label field resolves to `null`
- **AND** the query completes successfully returning the code field

#### Scenario: Enum definition without matching code

- **WHEN** the stored code doesn't match any option in the enum definition
- **THEN** the label field resolves to `null`
- **AND** the query completes successfully returning the invalid code value

### Requirement: Label Field in Aggregates

The system SHALL determine whether enum label fields are included in aggregate queries.

#### Scenario: Label fields not in count with select

- **WHEN** executing a count query with select parameter
- **THEN** select choices do not include label fields
- **AND** only the code field can be selected

#### Scenario: Label fields not in aggregate operations

- **WHEN** executing an aggregate query (_sum, _avg, _min, _max)
- **THEN** label fields cannot be used in aggregate operations
- **AND** only the code field (which stores string) can be aggregated

## Cross-References
- **Related Spec**: `openspec/changes/refactor-enum-code-label/specs/enum-code-label-rename/spec.md`
- **Base Spec**: `openspec/specs/modeldesign-field-types/spec.md` (Enum Data Storage requirement)
