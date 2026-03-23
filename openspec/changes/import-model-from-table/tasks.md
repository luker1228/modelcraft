## 1. Backend — GraphQL Schema

- [x] 1.1 Add `TableInfo` type to `api/graph/schema/project.graphql`
- [x] 1.2 Add `ListTablesInput` input type to `api/graph/schema/project.graphql`
- [x] 1.3 Add `listTables` query to `extend type Query` in `api/graph/schema/project.graphql`
- [x] 1.4 Run `just generate-gql` to regenerate gqlgen code

## 2. Backend — App Service

- [x] 2.1 Create `internal/app/modeldesign/list_tables_app.go` with `ListTables` method on `ReverseEngineerAppService`
- [x] 2.2 Implement INFORMATION_SCHEMA query to fetch BASE TABLE names for the given database
- [x] 2.3 Implement `excludeExisting` logic: query existing model names and filter them out

## 3. Backend — GraphQL Resolver

- [x] 3.1 Check `resolver.go` for `ReverseEngineerService` field; add it if missing and wire in `resolver_factory.go`
- [x] 3.2 Add `ListTables` resolver to `internal/interfaces/graphql/project.resolvers.go`
- [x] 3.3 Map `[]string` result to `[]*generated.TableInfo`

## 4. Frontend — GraphQL Query

- [x] 4.1 Add `LIST_TABLES` query (with `ListTablesInput`) to `src/graphql/queries/project.ts` (create file if it doesn't exist)

## 5. Frontend — ImportModelDialog Component

- [x] 5.1 Create `src/components/model-editor/ImportModelDialog.tsx`
- [x] 5.2 On dialog open, call `listTables` with current `projectSlug` and `databaseName`; show loading state
- [x] 5.3 Render searchable table list; selected row uses `#dadee5` background
- [x] 5.4 Show empty state when list is empty
- [x] 5.5 On "导入" confirm, call `REVERSE_ENGINEER_MODEL` mutation with `tableName` and `ddlStatement: ""`
- [x] 5.6 On success: close dialog, show success toast, refetch `GET_MODELS`
- [x] 5.7 On error: show error toast, keep dialog open
- [x] 5.8 Disable "导入" button when nothing selected or mutation is in flight; show loading spinner on button during mutation

## 6. Frontend — Wire Button in Page

- [x] 6.1 Import `ImportModelDialog` in `src/app/org/[orgName]/projects/[projectSlug]/model-editor/page.tsx`
- [x] 6.2 Add `importDialogOpen` state and "导入模型" button next to "新建模型" button
- [x] 6.3 Pass required props (`projectSlug`, `databaseName`, `open`, `onOpenChange`, `onSuccess`) to the dialog
