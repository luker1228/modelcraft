# GraphQL Error Optimization Proposal - Example Schema

This document provides example GraphQL schema definitions to illustrate the proposed changes.

## Proposed Schema Structure

```graphql
# ============================================
# Error Interface and Types
# ============================================

# Common interface for all user-facing errors
interface Error {
  message: String!
}

# Specific error types for Project operations
type ProjectAlreadyExists implements Error {
  message: String!
  suggestion: String
}

type ProjectNotFound implements Error {
  message: String!
}

type InvalidProjectInput implements Error {
  message: String!
  suggestion: String
}

type CannotDeleteDefaultProject implements Error {
  message: String!
}

# ============================================
# Error Unions (per mutation/query)
# ============================================

union GetProjectError = ProjectNotFound
union CreateProjectError = ProjectAlreadyExists | InvalidProjectInput
union UpdateProjectError = ProjectNotFound | InvalidProjectInput
union DeleteProjectError = ProjectNotFound | CannotDeleteDefaultProject
union ArchiveProjectError = ProjectNotFound
union ActivateProjectError = ProjectNotFound

# ============================================
# Updated Payload Types
# ============================================

type GetProjectPayload {
  project: Project              # Nullable - null on error
  error: GetProjectError        # NEW - single optional error
}

type CreateProjectPayload {
  project: Project              # Nullable - null on error
  error: CreateProjectError     # NEW - single optional error
}

type UpdateProjectPayload {
  project: Project              # Nullable - null on error
  error: UpdateProjectError     # NEW - single optional error
}

type DeleteProjectPayload {
  success: Boolean!             # Backward compatible
  error: DeleteProjectError     # NEW - single optional error
}

type ArchiveProjectPayload {
  success: Boolean!             # Backward compatible
  error: ArchiveProjectError    # NEW - single optional error
}

type ActivateProjectPayload {
  success: Boolean!             # Backward compatible
  error: ActivateProjectError   # NEW - single optional error
}
```

## Example GraphQL Queries/Mutations

### Creating a project (success case)

```graphql
mutation {
  createProject(input: {
    id: "my-project"
    title: "My Project"
    description: "A test project"
  }) {
    project {
      id
      title
    }
    error {
      __typename
      ... on Error {
        message
      }
    }
  }
}

# Response:
{
  "data": {
    "createProject": {
      "project": {
        "id": "my-project",
        "title": "My Project"
      },
      "error": null
    }
  }
}
```

### Creating a project (conflict error case)

```graphql
mutation {
  createProject(input: {
    id: "default"  # Already exists
    title: "Duplicate Project"
  }) {
    project {
      id
      title
    }
    error {
      __typename
      ... on ProjectAlreadyExists {
        message
        suggestion
      }
      ... on InvalidProjectInput {
        message
        suggestion
      }
    }
  }
}

# Response:
{
  "data": {
    "createProject": {
      "project": null,
      "error": {
        "__typename": "ProjectAlreadyExists",
        "message": "Project already exists: default",
        "suggestion": "Please use a different project ID"
      }
    }
  }
}
```

### Updating a project (not found error case)

```graphql
mutation {
  updateProject(
    id: "non-existent-project"
    input: { title: "New Title" }
  ) {
    success
    project {
      id
      title
    }
    error {
      __typename
      ... on ProjectNotFound {
        message
      }
    }
  }
}

# Response:
{
  "data": {
    "updateProject": {
      "success": false,
      "project": null,
      "error": {
        "__typename": "ProjectNotFound",
        "message": "Project not found: non-existent-project"
      }
    }
  }
}
```

### Deleting default project (operation denied case)

```graphql
mutation {
  deleteProject(id: "default") {
    success
    error {
      __typename
      ... on CannotDeleteDefaultProject {
        message
      }
      ... on ProjectNotFound {
        message
      }
    }
  }
}

# Response:
{
  "data": {
    "deleteProject": {
      "success": false,
      "error": {
        "__typename": "CannotDeleteDefaultProject",
        "message": "Cannot delete the default project"
      }
    }
  }
}
```

### Query Project Example

```graphql
query {
  project(id: "my-project") {
    project {
      id
      title
    }
    error {
      __typename
      ... on ProjectNotFound {
        message
      }
    }
  }
}

# Response (success):
{
  "data": {
    "project": {
      "project": {
        "id": "my-project",
        "title": "My Project"
      },
      "error": null
    }
  }
}

# Response (not found):
{
  "data": {
    "project": {
      "project": null,
      "error": {
        "__typename": "ProjectNotFound",
        "message": "Project not found: my-project"
      }
    }
  }
}
```

## Client Usage Patterns

### TypeScript Example (with codegen)

```typescript
// Generated types from GraphQL schema
import {
  CreateProjectMutation,
  CreateProjectError
} from './generated/graphql';

async function createProject(input: CreateProjectInput) {
  const result = await client.mutate<CreateProjectMutation>({
    mutation: CREATE_PROJECT,
    variables: { input }
  });

  const { project, errors } = result.data.createProject;

  // Check for errors first
  if (errors.length > 0) {
    errors.forEach(error => {
      switch (error.__typename) {
        case 'ProjectAlreadyExists':
          console.error(`Project already exists: ${error.message}`);
          if (error.suggestion) {
            console.log(`Suggestion: ${error.suggestion}`);
          }
          break;
        case 'InvalidProjectInput':
          console.error(`Invalid input: ${error.message}`);
          if (error.suggestion) {
            console.log(`Suggestion: ${error.suggestion}`);
          }
          break;
      }
    });
    return null;
  }

  return project;
}
```

### Backward Compatible Client (Old Pattern)

```graphql
# Old clients can still use this pattern
mutation {
  createProject(input: {...}) {
    project {
      id
      title
    }
  }
}
```

## Benefits Summary

1. **Type Safety**: Clients know exactly what errors to expect per mutation
2. **Better UX**: Specific error messages with context (IDs, suggestions)
3. **Retry Logic**: Clients can distinguish transient vs permanent errors
4. **Introspection**: GraphQL schema documents all possible errors
5. **Backward Compatible**: Existing clients continue to work
6. **Testable**: Each error type can be unit tested
7. **Consistent**: All errors implement Error interface
8. **Scalable**: Pattern can be applied to Model, Cluster, Enum domains
