## Why

Models in a project currently have no organizational structure beyond their name — as projects grow, users need a way to visually group and categorize models. The existing `tags` field on models was a placeholder that was never exposed or used, and should be replaced with a proper grouping mechanism.

## What Changes

- **New**: `model_groups` table — project-scoped group entities with lexicographic ordering for drag-and-drop reordering
- **New**: `group_id` column on `models` table — soft reference to a group (`NULL` = ungrouped)
- **Removed**: `tags` column from `models` table — unused placeholder, dropped from schema and domain
- **New**: Full CRUD operations for groups: create, rename, delete, reorder
- **New**: Move model to/from group
- **New**: Virtual "ungrouped" group — always exists, never stored, appended last in listings
- **New**: GraphQL API for group management and model-group assignment
- Group names: lowercase letters, numbers, underscore only; must start with a letter; unique within project

## Capabilities

### New Capabilities

- `model-group-management`: Create, rename, delete, and reorder groups within a project; move models between groups; list groups with their models; virtual ungrouped group behavior

### Modified Capabilities

- `project-management`: Models now carry a `group` field in responses; `tags` field removed from model domain and API

## Impact

- `db/schema/mysql/03_model_domain.sql` — modified (remove `tags`, add `group_id`, add `model_groups` table)
- `internal/domain/modeldesign/` — `ModelMeta` loses `tags`, gains `groupID`; new `ModelGroup` domain entity
- `internal/infrastructure/` — new `ModelGroupRepository`; `ModelRepository` updated for `group_id`
- `internal/app/` — new use cases for group CRUD and model assignment
- `api/graph/schema/` — new `ModelGroup` type, mutations, queries
- GraphQL generated code — regeneration required after schema changes
