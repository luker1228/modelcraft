## Context

ModelCraft is a model design tool where users define data models backed by a MySQL database. Each project connects to a database cluster. Currently, users must create models manually — entering name, title, and each field by hand.

The codebase already has `reverseEngineerModel` mutation (backend: `ReverseEngineerAppService`) that, given a table name, introspects the database and creates a model automatically. However, there is no UI entry point that lets users browse and pick an existing table to import.

The backend has `ListDatabases` as a direct reference pattern: `project.graphql` schema → `project.resolvers.go` resolver → cluster app service.

## Goals / Non-Goals

**Goals:**
- Add `listTables` GraphQL query that returns tables in a given database, excluding those already imported as models
- Add frontend "导入模型" button + dialog that calls `listTables` then `reverseEngineerModel`
- Reuse existing `reverseEngineerModel` mutation without modification

**Non-Goals:**
- Bulk import of multiple tables in a single operation
- Preview or editing of field mappings before import
- Support for databases outside the project's configured cluster
- Any changes to the `reverseEngineerModel` backend logic

## Decisions

### D1: New query in `project.graphql`, not `model.graphql`

`listTables` is an infrastructure query (reads from the database cluster connection), not a model domain operation. Placing it in `project.graphql` alongside `listDatabases` keeps the schema organized by concern.

**Alternative considered**: Add to `model.graphql` since it supports model creation. Rejected — `listDatabases` already lives in `project.graphql` and `listTables` is symmetric with it.

### D2: Implement in `ReverseEngineerAppService`, not a new service

`ReverseEngineerAppService` already holds `clusterManager` and `modelRepo` — the two dependencies needed for `listTables`. Adding a new method there avoids introducing another service and keeps related logic colocated.

**Alternative considered**: New `TableIntrospectionService`. Rejected — YAGNI; the existing service already has everything needed.

### D3: `excludeExisting` defaults to `true` on the backend

The UI always wants to show only unimported tables. Defaulting to `true` means the frontend doesn't need to pass the flag explicitly, and a future admin use-case could pass `false` to see all tables.

### D4: Frontend calls `reverseEngineerModel` with `ddlStatement: ""`

The existing mutation accepts either `ddlStatement` or `tableName` (validated in `validateRequest`). Passing an empty `ddlStatement` and a `tableName` triggers the database introspection path. No backend changes needed.

**Note**: The current GraphQL schema marks `ddlStatement` as `String!` (non-null). This means the frontend must pass `""` — an empty string — to signal the table-name introspection path. This is an existing API convention.

## Risks / Trade-offs

- **Large table count**: If a database has hundreds of tables, the list could be long. → Mitigation: add client-side search/filter in the dialog.
- **Stale list**: If a model is created outside the dialog between the time the user opens it and clicks import, `reverseEngineerModel` will fail with `ModelAlreadyExists`. → Mitigation: surface the error as a toast and close the dialog; the new model will appear in the sidebar.
- **`generate-gql` must be run**: After schema changes, `just generate-gql` must be run or the resolver won't compile. → Documented in tasks.

## Open Questions

(none — all decisions resolved)
