## Why

`ModelRelation` currently encodes bidirectional semantics in a single record — a `MANY_TO_ONE` row implicitly owns both sides of the relationship through `target_relation_name`, a loose string coupling. This makes queries asymmetric (`WHERE model_id = ? OR target_model_id = ?`), the direction of a relation ambiguous (is `source_fields` the FK column or the referenced column?), and deletion logic fragile.

The root problem is that `ModelRelation` tries to represent two things at once: (1) the physical FK column mapping, and (2) the logical navigation fields visible on each model. Separating these concerns leads to a cleaner, more queryable design.

## What Changes

- **Removed**: `ModelRelation` entity, table, and all associated repository/service code
- **New**: `LogicalForeignKey` entity — represents a single FK mapping as two symmetric rows sharing a `pair_id`
  - `direction: normal` — the model that owns the FK columns (`source_fields`)
  - `direction: reverse` — the referenced model (mirrored, redundant for self-contained reads)
  - All `FindByModel` queries become `WHERE model_id = ?` (no OR)
- **Changed**: `FieldDefinition` gains two new nullable columns:
  - `belongs_to_fk_id` — FK column fields (e.g. `order.userId`) reference the `LogicalForeignKey` row for their model; deletion of this field cascades to delete the FK (after relation fields are gone)
  - `relate_fk_id` — RELATION-format fields (e.g. `order.user`, `user.orders`) reference the `LogicalForeignKey` row for their model; these are optional and independently creatable/deletable
- **New**: Explicit two-step creation flow: create FK first, then optionally attach RELATION fields
- **New**: `DeleteLogicalForeignKey(pair_id)` — pair_id is the only exposed deletion granularity; single-side deletion is not permitted

## Capabilities

### New Capabilities

- `logical-foreign-key-management`: Create and delete logical foreign keys (pair-wise); query FKs by model; inspect which fields belong to or relate to a FK

### Modified Capabilities

- `field-management`: `AddField` with `format=RELATION` now requires a pre-existing `relate_fk_id`; `belongs_to_fk_id` is set when creating FK columns; deletion order is enforced by dependency checks
- `model-management`: Model deletion cascades to its `LogicalForeignKey` rows via DB CASCADE; orphaned reverse rows from the deleted model's pairs are cleaned up by the application layer

## Impact

- `db/schema/mysql/03_model_domain.sql` — remove `model_relations` table; add `logical_foreign_keys` table; add `belongs_to_fk_id` and `relate_fk_id` columns to `field_definitions`
- `internal/domain/modeldesign/` — remove `ModelRelation`, `ModelRelationRepository`; add `LogicalForeignKey`, `LogicalForeignKeyRepository`
- `internal/infrastructure/` — remove `SqlModelRelationRepository`; add `SqlLogicalForeignKeyRepository`
- `internal/app/modeldesign/` — remove relation use cases; add FK creation/deletion use cases
- `api/graph/schema/field.graphql` — remove `RelationType` enum, `RelationConfig` type/input; add `LogicalForeignKey` type, `CreateLogicalForeignKeyInput`
- GraphQL generated code — regeneration required
