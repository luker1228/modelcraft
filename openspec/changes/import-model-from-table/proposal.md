## Why

When users set up a new project, they often already have existing database tables that should become models. Currently there is no way to import tables directly — users must manually create models and re-enter every field. This wastes time and introduces mistakes when the database schema is the source of truth.

## What Changes

- Add a **"导入模型" (Import Model)** button to the model editor sidebar, next to the existing "新建模型" button.
- Add a new `listTables` GraphQL query that returns the list of tables in a given database, filtered to exclude tables that already have a corresponding model.
- Add a new `ImportModelDialog` frontend component: user selects a table from the list and clicks "导入" to auto-create the model via the existing `reverseEngineerModel` mutation.

## Capabilities

### New Capabilities

- `list-tables`: Backend GraphQL query `listTables(input: ListTablesInput!)` that queries `INFORMATION_SCHEMA.TABLES` from the project's connected database cluster and returns unimported table names.
- `import-model-dialog`: Frontend dialog component that lets users browse available tables and trigger model import via `reverseEngineerModel`.

### Modified Capabilities

(none)

## Impact

- **Backend**: New GraphQL schema types (`TableInfo`, `ListTablesInput`), new query in `project.graphql`, new App Service method in `ReverseEngineerAppService`, new resolver in `project.resolvers.go`. Requires `just generate-gql` after schema changes.
- **Frontend**: New GraphQL query file, new `ImportModelDialog` component, minor edit to `model-editor/page.tsx` to add the import button.
- **No breaking changes.** All existing APIs and UI remain unchanged.
