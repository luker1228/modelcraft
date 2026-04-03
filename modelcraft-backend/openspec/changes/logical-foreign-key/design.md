## Context

`ModelRelation` encodes both sides of a foreign key relationship in one record, using `target_relation_name` (a plain string) to loosely couple the reverse navigation name. This design forces asymmetric queries, ambiguous field semantics, and brittle deletion ordering.

The replacement — `LogicalForeignKey` — is a pair of symmetric rows: one per participating model, bound by a shared `pair_id`. Physical FK columns reference their row via `belongs_to_fk_id`. Optional RELATION-format navigation fields reference their row via `relate_fk_id`. Both columns live on `FieldDefinition`.

## Goals / Non-Goals

**Goals:**
- Replace `ModelRelation` with `LogicalForeignKey` (pair-per-FK, direction-explicit)
- Make `FindByModel` a simple `WHERE model_id = ?` query
- Enforce creation order: FK columns must exist before FK can be created; FK must exist before RELATION fields can be created
- Enforce deletion order via application-layer checks (not DB constraints)
- Allow RELATION fields to be created independently per side (one or both)
- Expose `pair_id` as the only deletion granularity for FKs
- Follow TDD: tests written before implementation

**Non-Goals:**
- Many-to-many support in this change (through-table pattern deferred)
- Migrating existing `model_relations` data (handled separately)
- Changing runtime query execution logic

## Decisions

### D1: Two rows per FK, not one

**Decision**: Each FK is stored as two `logical_foreign_keys` rows sharing a `pair_id`. The `normal` row has `model_id` = the model owning the FK columns. The `reverse` row has `model_id` = the referenced model.

**Rationale**: Every row is "self-centered" — its `model_id` is always the model it belongs to. `FindByModel(X)` is always `WHERE model_id = X`. No OR queries. Deleting a model cascades cleanly to its own rows.

**Trade-off**: Redundant storage of `source_fields`/`target_fields` on the reverse row. Accepted — reads are self-contained, no join needed.

### D2: `source_fields`/`target_fields` redundantly stored on reverse row

**Decision**: Reverse row stores fields mirrored (`source_fields` = referenced columns, `target_fields` = FK columns), fully redundant with the normal row.

**Rationale**: Any reader of the reverse row can determine the full FK mapping without fetching its pair. Simplifies all query paths that start from the reverse side.

### D3: `belongs_to_fk_id` is set on both source-side and target-side FK columns

**Decision**: `order.userId` and `order.companyId` get `belongs_to_fk_id = lf-001 (normal)`. `user.id` and `user.companyId` get `belongs_to_fk_id = lf-002 (reverse)`. Each field points to the row whose `model_id` matches its own model.

**Rationale**: Consistent with the "self-centered" invariant. `FindByModel(user)` returns both the `lf-002` row and `user.id`/`user.companyId` fields referencing it — all via `WHERE model_id = user`.

### D4: RELATION fields are optional and independently creatable per side

**Decision**: Creating a FK does not automatically create any RELATION fields. The user explicitly calls `AddField` with `format=RELATION` and `relate_fk_id` set to the appropriate `lf-id`. Either side can be created alone.

**Rationale**: Not every FK needs a navigation field on both sides. Decoupling creation gives maximum flexibility. The FK itself is the canonical source of truth; RELATION fields are optional views over it.

### D5: `pair_id` is the only FK deletion granularity

**Decision**: `DeleteLogicalForeignKey(pair_id)` atomically deletes both rows. There is no `DeleteLogicalForeignKey(lf_id)`.

**Rationale**: A single-row pair is an orphan and an invalid state. Exposing only `pair_id` deletion makes it impossible to create orphans through the API.

**Pre-condition**: Both `lf-001` and `lf-002` must have zero `relate_fk_id` references before deletion is allowed.

### D6: Deletion order enforced by application layer, not DB constraints

**Decision**: No DB-level FK constraints between `field_definitions` and `logical_foreign_keys`. Order enforced by application-layer checks.

**Rationale**: Consistent with project policy of avoiding FK constraints. Application-layer checks produce typed errors.

Enforced order:
1. Delete RELATION fields (`relate_fk_id` references) first
2. Delete FK via `pair_id` (or proceed to step 3 directly)
3. Delete FK columns (`belongs_to_fk_id` references) — this triggers cascade deletion of the FK pair if still present

### D7: Model deletion cascades via DB CASCADE

**Decision**: `logical_foreign_keys` has `model_id` as a DB-level CASCADE FK to `models.id`. When a model is deleted, all its `logical_foreign_keys` rows are deleted automatically.

**Exception**: The reverse row (`model_id = other_model`) is NOT in the deleted model's cascade. The application layer must clean up orphaned reverse rows on model deletion — query by `ref_model_id` and delete any rows whose pair counterpart no longer exists.

### D8: No many-to-many in this change

**Decision**: `through_table` concept is deferred. This change only handles direct FK relationships (many-to-one / one-to-many).

## Data Model

```
logical_foreign_keys
─────────────────────────────────────────────────
id             VARCHAR(36)   PK
pair_id        VARCHAR(36)   NOT NULL  (shared by the two rows)
direction      ENUM('normal','reverse')  NOT NULL
model_id       VARCHAR(36)   NOT NULL  (FK → models.id CASCADE DELETE)
ref_model_id   VARCHAR(36)   NOT NULL
source_fields  JSON          NOT NULL  (FK columns on model_id's table for normal; referenced cols for reverse)
target_fields  JSON          NOT NULL  (referenced cols on ref_model_id's table for normal; FK cols for reverse)
created_at     DATETIME
updated_at     DATETIME

INDEX (pair_id)
INDEX (model_id)

field_definitions (new columns)
─────────────────────────────────────────────────
belongs_to_fk_id   VARCHAR(36)  NULL  (set on FK column fields)
relate_fk_id       VARCHAR(36)  NULL  (set on RELATION-format fields)
```

## Deletion Dependency Graph

```
RELATION field (relate_fk_id)
    must be deleted BEFORE
LogicalForeignKey pair
    must be deleted BEFORE  (or triggers cascade of)
FK column field (belongs_to_fk_id)
```

## Risks / Trade-offs

- **Orphaned reverse rows on model deletion**: Application layer must query `WHERE ref_model_id = deleted_model_id` and clean up. If missed, stale rows remain but cause no query-time errors (the model they point to is gone). Mitigation: wrap model deletion in a transaction with explicit cleanup step.
- **Redundant field data on reverse row**: If `source_fields` changes on the normal row (currently not supported post-creation), reverse row becomes stale. Mitigation: FK fields are immutable after creation — changing FK columns requires delete + recreate.
- **Migration of existing `model_relations` data**: Existing relation records must be migrated to the new schema. This is a one-time data migration handled outside this change.
