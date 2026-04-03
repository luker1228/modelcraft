# Python Testing Guidelines

## Purpose

This specification defines guidelines and best practices for Python automated testing in the ModelCraft project. It ensures test isolation, prevents resource conflicts, and maintains clean test data through proper cleanup mechanisms.

## Requirements

### Requirement: Unique Test Resource Naming

Test resources (such as clusters, projects, models) SHALL use unique identifiers to prevent conflicts between test runs.

#### Scenario: Using timestamp suffix for uniqueness
- **WHEN** a test creates a named resource
- **THEN** the resource name MUST include a timestamp suffix or random ID
- **AND** the format SHALL be `{base_name}_{timestamp}` or `{base_name}_{random_id}`

#### Scenario: Using UUID for uniqueness
- **WHEN** a test needs a globally unique identifier
- **THEN** the test MAY use UUID4 or a shortened version
- **AND** the identifier SHALL be unique across concurrent test runs

**Implementation Pattern:**
```python
import time
import uuid

def generate_unique_name(base_name: str) -> str:
    """Generate a unique name using timestamp."""
    timestamp = int(time.time() * 1000)
    return f"{base_name}_{timestamp}"

def generate_unique_id() -> str:
    """Generate a unique ID using UUID."""
    return str(uuid.uuid4())[:8]
```

---

### Requirement: Test Resource Cleanup

All test-created resources SHALL be cleaned up after test execution to prevent database pollution and resource leaks.

#### Scenario: Using pytest fixtures for cleanup
- **WHEN** a test module creates resources
- **THEN** a module-scoped fixture SHALL track created resource IDs
- **AND** the fixture SHALL clean up resources in its teardown phase

#### Scenario: Cleanup on test failure
- **WHEN** a test fails during execution
- **THEN** cleanup SHALL still occur via fixture teardown
- **AND** cleanup errors SHALL be logged but not cause additional test failures

#### Scenario: Order of cleanup
- **WHEN** multiple resources are created with dependencies
- **THEN** cleanup SHALL occur in reverse creation order
- **AND** child resources SHALL be deleted before parent resources

**Implementation Pattern:**
```python
import pytest
from typing import List
from gql import gql

@pytest.fixture(scope="module")
def created_resources(graphql_client) -> List[str]:
    """
    Track created resources for cleanup.
    
    Usage:
        def test_create_something(graphql_client, created_resources):
            result = create_resource(graphql_client, data)
            created_resources.append(result["id"])  # Track for cleanup
    """
    resources = []
    yield resources
    
    # Cleanup after all module tests complete
    if resources:
        print(f"\n🧹 Cleaning up {len(resources)} test resources...")
        
        DELETE_MUTATION = gql("""
            mutation DeleteResource($id: ID!) {
                deleteResource(id: $id) {
                    success
                }
            }
        """)
        
        # Cleanup in reverse order (LIFO)
        for resource_id in reversed(resources):
            try:
                result = graphql_client.execute(
                    DELETE_MUTATION, 
                    variable_values={"id": resource_id}
                )
                if result.get("deleteResource", {}).get("success"):
                    print(f"  ✅ Deleted: {resource_id}")
                else:
                    print(f"  ⚠️  Failed to delete: {resource_id}")
            except Exception as e:
                print(f"  ❌ Error deleting {resource_id}: {e}")
```

---

### Requirement: Test Data Factory Functions

Test data creation SHALL use factory functions with sensible defaults to ensure consistency and reduce boilerplate.

#### Scenario: Factory function with defaults
- **WHEN** creating test input data
- **THEN** a factory function SHALL provide sensible default values
- **AND** all defaults SHALL be overridable via function parameters

#### Scenario: Auto-generated unique names
- **WHEN** a factory function is called without a name parameter
- **THEN** it SHALL auto-generate a unique name using timestamp or random ID

**Implementation Pattern:**
```python
from typing import Optional
import time

def build_test_input(
    name: Optional[str] = None,
    title: Optional[str] = None,
    **kwargs
) -> dict:
    """
    Build test input with sensible defaults.
    
    Args:
        name: Resource name (auto-generated if not provided)
        title: Display title (defaults to name if not provided)
        **kwargs: Additional fields to include
    
    Returns:
        dict: Input data ready for GraphQL mutation
    """
    if name is None:
        timestamp = int(time.time() * 1000)
        name = f"test-resource-{timestamp}"
    
    if title is None:
        title = f"Test Resource {name}"
    
    result = {
        "name": name,
        "title": title,
    }
    result.update(kwargs)
    return result
```

---

### Requirement: Prerequisite Resource Setup

Tests requiring prerequisite resources (e.g., a project before creating a cluster) SHALL use fixtures to ensure prerequisites exist.

#### Scenario: Ensuring prerequisite exists
- **WHEN** a test requires a parent resource
- **THEN** a fixture SHALL check if the resource exists
- **AND** create it if not found
- **AND** return the resource identifier for use in tests

**Implementation Pattern:**
```python
@pytest.fixture(scope="module")
def default_project(graphql_client) -> str:
    """
    Ensure a 'default' project exists for testing.
    Creates the project if it doesn't exist.
    Returns the project ID.
    """
    project_id = "default"
    
    # Try to get existing project
    try:
        result = graphql_client.execute(GET_PROJECT, {"id": project_id})
        if result.get("project", {}).get("project"):
            return project_id
    except Exception:
        pass
    
    # Create if not exists
    try:
        graphql_client.execute(CREATE_PROJECT, {
            "input": {
                "id": project_id,
                "title": "Default Test Project",
            }
        })
    except Exception:
        pass  # May already exist from concurrent test
    
    return project_id
```

---

## File Organization

### Test Directory Structure
```
tests/
├── conftest.py              # Root fixtures (URLs, env setup)
├── design/                  # Design-time API tests
│   ├── conftest.py          # Design-specific fixtures
│   ├── common/              # Shared utilities
│   │   ├── graphql_client.py
│   │   └── test_data.py     # Factory functions
│   ├── cluster/
│   │   └── test_cluster_graphql.py
│   └── project/
│       └── test_project_graphql.py
└── runtime/                 # Runtime API tests
    └── ...
```

### Fixture Scope Guidelines

| Scope | Use Case |
|-------|----------|
| `function` | Isolated state per test |
| `class` | Shared state within test class |
| `module` | Shared state within file, cleanup after all tests |
| `session` | Global setup (e.g., server health check) |

---

## Best Practices Checklist

- [ ] All resource names include timestamp or random suffix
- [ ] Created resources are tracked in cleanup fixtures
- [ ] Factory functions provide sensible defaults
- [ ] Prerequisites are ensured via fixtures
- [ ] Cleanup handles errors gracefully
- [ ] Tests are independent and can run in any order
