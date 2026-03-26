---
name: graphql-client
description: Create and manage frontend GraphQL client definitions in src/graphql/. Use when adding new queries, mutations, or GraphQL operations for the ModelCraft frontend. Covers the conventions for writing gql template literals, field selection, error handling, file organization, and barrel exports. Definitions derive from contract/graph/schema/*.graphql. Triggers on phrases like "add graphql query", "create mutation", "add graphql operation", "write query for", "new mutation for".
---

# GraphQL Client Definitions

All frontend GraphQL operations live in `src/graphql/`. Only define what's actually needed — skip unused schema fields.

## File Organization

```
src/graphql/
├── index.ts                # barrel re-export
├── queries/
│   ├── index.ts            # barrel re-export
│   ├── project.ts          # GET_PROJECTS, GET_PROJECT, LIST_TABLES
│   ├── model.ts            # GET_MODELS, GET_MODEL, GET_MODEL_BY_NAME, ...
│   ├── cluster.ts          # GET_CLUSTER, LIST_DATABASES
│   ├── enum.ts             # GET_ENUMS, GET_ENUM, GET_ENUM_REFERENCES
│   └── user.ts             # GET_ME, GET_MY_ORGANIZATIONS, GET_ORGANIZATION_MEMBERS, GET_ROLES
└── mutations/
    ├── index.ts            # barrel re-export
    ├── project.ts          # CREATE_PROJECT, UPDATE_PROJECT, DELETE_PROJECT, ...
    ├── model.ts            # CREATE_MODEL, UPDATE_MODEL, DELETE_MODEL, ...
    ├── cluster.ts          # TEST_CLUSTER_CONNECTION
    ├── enum.ts             # CREATE_ENUM, UPDATE_ENUM, DELETE_ENUM
    └── user.ts             # UPDATE_ORGANIZATION, CREATE_ROLE, DELETE_ROLE
```

## Contract Schema Source

Contract schemas at `contract/graph/schema/*.graphql` define the API surface.

| Contract File | Domain | Key Types |
|---|---|---|
| `schema.graphql` | Root | `Query`, `Mutation` |
| `base.graphql` | Shared | Scalars (`Int64`, `Date`, `Time`), `Node`, `PageInfo` |
| `model.graphql` | Model/Group | `DataModel`, `ModelGroup`, repair/sync/import |
| `field.graphql` | Field | `FieldDefinition`, `DbColumnInfo`, `ValidationConfig` |
| `enum.graphql` | Enum | `EnumDefinition`, `EnumOption` |
| `logical_foreign_key.graphql` | FK | `LogicalForeignKey`, `ForeignKeyDirection` |
| `project.graphql` | Project/Cluster | `Project`, `DatabaseCluster`, errors |
| `user_management.graphql` | User/Org | `Organization`, `Role`, `CurrentUser` |
| `permission.graphql` | Permissions | `PermissionRole`, `PermissionDef` |

Always read the relevant contract file before writing a new query/mutation.

## Naming Conventions

**Exported constant**: `SCREAMING_SNAKE_CASE`
```
GET_MODEL, CREATE_MODEL, UPDATE_MODEL, DELETE_MODEL
GET_MODELS, GET_MODEL_BY_NAME, TEST_CLUSTER_CONNECTION
```

**Operation name inside gql**: `PascalCase`
```
query GetModel($projectSlug: String!, $id: ID!) { ... }
mutation CreateModel($input: CreateModelInput!) { ... }
```

## Code Pattern

```typescript
import { gql } from '@apollo/client'

export const GET_MODEL = gql`
  query GetModel($projectSlug: String!, $id: ID!, $withActualSchema: Boolean) {
    model(projectSlug: $projectSlug, id: $id, withActualSchema: $withActualSchema) {
      model {
        id
        name
        description
        # only select fields the UI actually uses
      }
      error {
        __typename
        ... on ModelNotFound { message suggestion }
        ... on InvalidModelInput { message fields }
      }
    }
  }
`
```

### Error Handling (Union Payloads)

All mutations return a union payload with `error { __typename ... on SpecificError { ... } }`.

Common error types per domain:
- **NotFound**: `ProjectNotFound`, `ModelNotFound`, `EnumNotFound`, `ClusterNotFound`, `GroupNotFound`
- **AlreadyExists**: `ModelAlreadyExists`, `EnumAlreadyExists`, `RoleAlreadyExists`
- **InvalidInput**: `InvalidModelInput`, `InvalidEnumInput`, `InvalidClusterInput`
- **CannotDelete**: `CannotDeleteDeployedModel`, `CannotDeleteReferencedEnum`, `CannotDeleteDefaultProject`

### Mutations

```typescript
export const CREATE_MODEL = gql`
  mutation CreateModel($input: CreateModelInput!) {
    createModel(input: $input) {
      model {
        id
        name
      }
      error {
        __typename
        ... on ModelAlreadyExists { message suggestion }
        ... on InvalidModelInput { message fields }
      }
    }
  }
`
```

## Barrel Exports

Every new file must be added to the barrel `index.ts`:

```typescript
// src/graphql/queries/index.ts
export * from './project'
export * from './model'
export * from './cluster'
export * from './enum'
export * from './user'
```

Consumers import from the top level:
```typescript
import { GET_MODELS, CREATE_MODEL } from '@/graphql'
```

## Key Rules

1. **Only select needed fields** — don't request everything the schema offers
2. **No fragments** — inline field selection (no `fragment` keyword)
3. **No generated types** — Apollo Client infers types from gql literals
4. **Always handle errors** — include `error { __typename ... on ...Error { ... } }` in payload selection
5. **One domain per file** — project, model, cluster, enum, user
6. **Read the contract first** — check `contract/graph/schema/*.graphql` for the exact operation signature, input types, and available error types before writing
7. **Add to barrel export** — new files must be exported from `queries/index.ts` or `mutations/index.ts`

## GraphQL Endpoints

| Type | URL Pattern | Usage |
|---|---|---|
| Design-time | `/org/{orgName}/design/graphql` | Most queries/mutations |
| Runtime | `/org/{orgName}/project/{projectSlug}/db/{databaseName}/model/{modelName}` | Dynamic per-model queries |

Auth: `Authorization: Bearer {token}` from `localStorage`.
