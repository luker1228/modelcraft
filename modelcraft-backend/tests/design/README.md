# Design-Time Tests

Design-time tests verify the "design phase" operations of ModelCraft:
- Project management (create, update, delete projects)
- Database cluster configuration
- Model design and schema definition
- Field and relation management

## Structure

```
design/
├── common/                    # Shared utilities (can be used by Runtime tests)
│   ├── graphql_client.py     # GraphQL client creation
│   ├── test_data.py          # Test data builders
│   ├── assertions.py         # Custom assertions
│   └── fixtures.py           # Shared fixtures
│
├── project/                   # Project management tests
│   ├── test_project_crud.py
│   └── test_project_isolation.py
│
├── cluster/                   # Database cluster tests
│   └── test_cluster_graphql.py
│
├── model/                     # Model design tests
│   ├── test_model_graphql.py
│   ├── test_jsonschema_export.py
│   └── test_schema_operations.py
│
├── compatibility/             # Backward compatibility tests
│   └── test_backward_compatibility.py
│
└── conftest.py               # Design test fixtures
```

## Running Design Tests

```bash
# Run all Design tests
pytest design/ -v

# Run specific domain tests
pytest design/project/ -v
pytest design/model/ -v

# Run with markers
pytest -m design -v
```

## Shared Utilities

The `design/common/` directory contains utilities that can be reused:

### GraphQL Client

```python
from tests.design.common.graphql_client import create_design_graphql_client

client = create_design_graphql_client("http://localhost:8080/org/modelcraft/design/graphql")
```

### Test Data Builders

```python
from tests.design.common.test_data import build_project_input, build_model_input

project = build_project_input(title="My Project")
model = build_model_input(cluster_id=1, name="User")
```

### Custom Assertions

```python
from tests.design.common.assertions import assert_graphql_success, assert_project_fields

assert_graphql_success(result, "createProject")
assert_project_fields(project, {"projectId": "test-123", "title": "Test"})
```

## Usage by Runtime Tests

Runtime tests can import and use Design utilities:

```python
# In runtime tests
from tests.design.common.graphql_client import create_design_graphql_client
from tests.design.common.test_data import build_model_input
```

## Writing New Design Tests

1. **Choose the right domain**: project/, cluster/, model/, compatibility/
2. **Use shared utilities**: Import from `design/common/`
3. **Use fixtures**: Leverage `graphql_client`, `created_projects`, etc.
4. **Follow naming**: `test_*.py` for files, `test_*` for functions
5. **Add cleanup**: Use fixtures to track created resources

### Example Test

```python
from tests.design.common.test_data import build_cluster_input
from tests.design.common.assertions import assert_graphql_success

def test_create_cluster(graphql_client, created_clusters):
    # Arrange
    input_data = build_cluster_input(name="test-cluster")

    # Act
    result = graphql_client.execute(CREATE_CLUSTER, variable_values={"input": input_data})

    # Assert
    assert_graphql_success(result, "createDatabaseCluster")
    cluster = result["createDatabaseCluster"]

    # Track for cleanup
    created_clusters.append(cluster["id"])
```

## Test Dependencies

Design tests should be independent and can run in any order. Each test:
- Creates its own test data
- Cleans up after itself (via fixtures)
- Does not depend on other tests
