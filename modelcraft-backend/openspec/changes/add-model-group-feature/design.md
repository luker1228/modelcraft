## Context

Models in a project have no organizational structure. The `tags []string` field on `ModelMeta` was a placeholder that was never exposed in GraphQL and has no real data. Users need to visually group models, especially for drag-and-drop reordering in the UI.

The system already has a strong pattern for project-scoped entities (model_enums, models) that we follow here. No FK constraints are used for project-scoped resources — consistent with the project's explicit policy.

## Goals / Non-Goals

**Goals:**
- Introduce a `model_groups` table scoped to `(org_name, project_slug)`
- Support full CRUD on groups with name validation
- Support drag-and-drop reordering via lexicographic fractional indexing
- Expose a virtual "ungrouped" group for models with no group assignment
- Remove the unused `tags` field from models
- Follow TDD: tests written before implementation

**Non-Goals:**
- Nested groups or group hierarchies
- Cross-project group sharing
- Bulk move of models between groups
- Group-level permissions or visibility controls

## Decisions

### D1: Group as entity with ID reference on Model (not string label)

**Decision**: `models.group_id VARCHAR(36) NULL` referencing `model_groups.id`.

**Alternatives considered**:
- Store group name directly on model — simpler, but rename requires updating all models in that group (dual-write, atomicity concern).

**Rationale**: Rename is a required operation. ID reference means a single UPDATE on `model_groups` and all models instantly reflect the new name via join. Consistent with how enums are referenced (`enum_name` on fields uses a soft string ref, but groups need rename support unlike enums).

No FK constraint added — consistent with project policy of avoiding FK constraints for project-scoped resources.

### D2: Lexicographic fractional indexing for display_order

**Decision**: `display_order VARCHAR(255)` using a lexicographic midpoint algorithm.

**Alternatives considered**:
- `INT` sequential ordering — O(n) writes on every reorder, unacceptable for frequent drag-and-drop.
- `FLOAT/DOUBLE` midpoint — O(1) write but precision degrades after many reorders, eventually requires full renumber.

**Rationale**: Lexicographic strings (e.g., "a0", "a0V", "a1") always have a computable midpoint, never lose precision, and sort correctly with standard `ORDER BY`. Used by Figma, Linear, Notion. Single row write per drag. Extremely rare collision (identical keys) triggers a lazy background renumber.

### D3: Virtual "ungrouped" group — never stored in DB

**Decision**: Models with `group_id IS NULL` belong to the virtual ungrouped group. The API layer synthesizes a sentinel object with `id = "__ungrouped__"`.

**Alternatives considered**:
- Store a real "ungrouped" group record per project — creates bootstrapping complexity (must create on project creation, must prevent deletion/rename).

**Rationale**: Virtual sentinel is simpler. No DB row means no lifecycle management. Frontend distinguishes it via `isVirtual: true`. Always appended last in listings regardless of display_order.

### D4: Group name validation — application layer, not DB constraint

**Decision**: Validate `^[a-z][a-z0-9_]*$` (max 64 chars) in the domain/application layer.

**Rationale**: DB CHECK constraints are harder to introspect for error messages. Application-layer validation produces typed GraphQL errors (`InvalidGroupName`). Uniqueness is enforced by DB unique index `(org_name, project_slug, name)`.

### D5: Delete cascades models to ungrouped (not blocked)

**Decision**: On group delete, `UPDATE models SET group_id = NULL WHERE group_id = ?` then `DELETE` the group, all in one transaction.

**Alternatives considered**:
- Block deletion if models exist — more conservative but creates friction; users must manually empty a group before deleting.

**Rationale**: Moving to ungrouped is non-destructive. Models are never lost. Consistent with intuitive UX for label/group deletion.

### D6: SQL schema — single file, model domain

**Decision**: All changes in `db/schema/mysql/03_model_domain.sql`. `model_groups` table inserted between `models` and `field_definitions`. `tags` column removed from `models` definition. `group_id` added to `models`.

**Rationale**: `model_groups` is part of the model domain, same file as `model_enums`. Atlas handles declarative schema diffing — no separate migration SQL needed.

## Risks / Trade-offs

- **Lexicographic collision** (extremely rare): Two groups end up with identical `display_order`. Mitigation: detect on reorder; if collision, renumber all groups in project (lazy, infrequent).
- **Soft reference integrity**: `group_id` has no FK constraint. A deleted group could leave orphaned `group_id` values if delete logic has a bug. Mitigation: delete operation always runs in a single transaction (SET NULL then DELETE).
- **tags removal**: If any future code path reads `tags`, it will break at compile time (Go struct field removed). Mitigation: grep for usages before removal; field was never exposed in GraphQL so surface area is minimal.

## Open Questions

- Should `ListGroups` include model count per group (without loading full models)? Useful for UI badges. Can be added as a follow-up.
- Should groups support a `title` field (display name separate from the technical `name`)? Currently out of scope — name serves as display label.
