# Runtime Tests

Runtime tests verify the "runtime phase" operations of ModelCraft:
- Data querying through GraphQL
- Model data CRUD operations
- Query filtering and pagination
- Aggregations and computed fields

**Important**: Runtime tests often depend on models being created in the Design phase.

## Structure

```
runtime/
├── query/                     # Data query tests
│   └── test_user_graphql.py
│
├── integration/               # End-to-end integration tests
│   └── test_modelcraft_client.py
│
└── conftest.py               # Runtime test fixtures
```

## Running Runtime Tests

```bash
# Run all Runtime tests
pytest runtime/ -v

# Run specific category tests
pytest runtime/query/ -v
pytest runtime/integration/ -v

# Run with markers
pytest -m runtime -v
```

## Dependencies on Design

Runtime tests can use Design utilities for setup:

```python
# Import Design utilities
from tests.design.common.graphql_client import create_design_graphql_client
from tests.design.common.test_data import build_model_input, build_cluster_input

# Use them to set up test data
def sample_model_setup(graphql_client):
    cluster = create_cluster(...)  # Design operation
    model = create_model(...)      # Design operation
    return model  # Use in Runtime tests
```

## Fixtures

Runtime tests have access to:

### From Root conftest.py
- `test_config`: Test configuration
- `base_url`: API base URL
- `design_graphql_url`: Design GraphQL endpoint
- `db_config`: Database configuration

### From design/conftest.py
- `graphql_client`: GraphQL client (can be used for setup)

### From runtime/conftest.py
- `runtime_graphql_client`: GraphQL client for Runtime API
- `sample_model_setup`: Example fixture that creates a test model

## Writing New Runtime Tests

1. **Set up Design phase first**: Create models/clusters needed for testing
2. **Use Runtime GraphQL queries**: Query data, not design operations
3. **Import Design utilities**: Reuse test data builders and assertions
4. **Handle missing data**: Use `pytest.skip()` if data doesn't exist

### Example Test

```python
from tests.design.common.assertions import assert_graphql_success

def test_query_user_data(runtime_graphql_client, sample_model_setup):
    # Arrange
    model_id = sample_model_setup["id"]  # Model created by Design fixture

    # Act: Runtime query
    result = runtime_graphql_client.execute(
        QUERY_USER_DATA,
        variable_values={"modelId": model_id}
    )

    # Assert
    assert_graphql_success(result, "findMany")
```

## Integration Tests

Integration tests in `runtime/integration/` verify complete workflows:
- Design phase: Create project → cluster → model
- Runtime phase: Query data, execute operations
- Cleanup: Remove test resources

These tests ensure the full ModelCraft pipeline works end-to-end.

## Test Independence

Like Design tests, Runtime tests should:
- Set up their own test data (via Design operations)
- Clean up after themselves
- Not depend on other tests
- Be able to run in any order
