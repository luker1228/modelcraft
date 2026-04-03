## ADDED Requirements

### Requirement: Model exposes group field
The `Model` GraphQL type SHALL include a `group: ModelGroup!` field. This field SHALL never be null — models with no explicit group assignment SHALL return the virtual ungrouped group (`id = "__ungrouped__"`, `isVirtual = true`).

#### Scenario: Model with assigned group returns group
- **WHEN** a client queries a model that belongs to a group
- **THEN** the `group` field SHALL return the `ModelGroup` with the correct `id`, `name`, and `isVirtual = false`

#### Scenario: Model without group returns virtual ungrouped
- **WHEN** a client queries a model that has no group assignment
- **THEN** the `group` field SHALL return the virtual ungrouped group with `id = "__ungrouped__"`, `name = "ungrouped"`, and `isVirtual = true`

## REMOVED Requirements

### Requirement: Model tags field
**Reason**: The `tags` field on `Model` was an unused placeholder with no real data and no GraphQL exposure. It is replaced by the structured group feature.
**Migration**: No migration required — the field was never exposed in the GraphQL API and contains no data.
