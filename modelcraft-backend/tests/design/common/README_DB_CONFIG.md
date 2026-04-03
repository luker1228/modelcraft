# Test Database Configuration Guide

This guide explains how to configure database connections for integration tests.

## Overview

All integration tests use a centralized database configuration located in `common/db_config.py`. This eliminates hardcoded connection parameters and allows easy configuration via environment variables.

## Default Configuration

Default test database connection settings:

```python
host: "localhost"
port: 3307
username: "root"
password: "modelcraft123"
```

## Configuration Methods

### Method 1: Use Default Configuration (Recommended)

Most tests can use the default configuration without any setup:

```python
from design.common.test_data import build_cluster_input

# Uses default test database configuration automatically
input_data = build_cluster_input(
    project_name="my-project",
    name="test-cluster"
)
```

### Method 2: Override via Environment Variables

Set environment variables to override default values:

```bash
# In .env file or shell
export TEST_DB_HOST="your-host"
export TEST_DB_PORT="3306"
export TEST_DB_USER="your-username"
export TEST_DB_PASSWORD="your-password"

# Run tests
pytest tests/design/cluster/
```

### Method 3: Override in Code

For specific test cases, override connection parameters directly:

```python
from design.common.test_data import build_cluster_input

# Override specific parameters
input_data = build_cluster_input(
    project_name="my-project",
    name="test-cluster",
    host="custom-host",
    port=3306
)
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TEST_DB_HOST` | `localhost` | Database host |
| `TEST_DB_PORT` | `3307` | Database port |
| `TEST_DB_USER` | `root` | Database username |
| `TEST_DB_PASSWORD` | `modelcraft123` | Database password |

## Usage Examples

### Creating a Cluster

```python
from design.common.test_data import build_cluster_input

# Automatic configuration
cluster_input = build_cluster_input(
    project_name="test-project",
    name="test-cluster"
)

# Manual override for specific test
cluster_input = build_cluster_input(
    project_name="test-project",
    name="test-cluster",
    host="custom-host",
    port=3306
)
```

### Accessing Configuration Directly

```python
from design.common.db_config import get_test_db_config

# Get configuration object
config = get_test_db_config()

print(f"Host: {config.host}")
print(f"Port: {config.port}")
print(f"Username: {config.username}")

# Convert to dictionary
config_dict = config.to_dict()
```

## Files Updated

The following files have been updated to use the unified configuration:

### Test Files
- `tests/design/cluster/test_cluster_graphql.py`
- `tests/design/cluster/test_project_cluster_one_to_one.py`
- `tests/design/model/test_model_graphql.py`
- `tests/design/model/test_model_errors.py`
- `tests/design/model/conftest.py`

### Configuration Files
- `tests/design/common/db_config.py` - New centralized configuration
- `tests/design/common/test_data.py` - Updated `build_cluster_input()` to use config

## Benefits

1. **Single Source of Truth**: All database connection settings in one place
2. **Environment Flexibility**: Easy to switch between different test environments
3. **Backward Compatible**: Existing tests continue to work without changes
4. **Override Support**: Can override defaults when needed for specific tests
5. **Maintainability**: No duplicated hardcoded values across test files

## Testing

To verify the configuration is working:

```bash
# Run cluster tests
cd modelcraft-go
pytest tests/design/cluster/test_cluster_graphql.py -v

# Run model tests
pytest tests/design/model/test_model_graphql.py -v

# Run with custom database
TEST_DB_HOST=custom-host TEST_DB_PORT=3306 pytest tests/design/cluster/ -v
```

## Troubleshooting

### Connection Failed

If you see `DatabaseConnectionFailed` errors:

1. Verify MySQL is running: `docker ps | grep mysql`
2. Check port mapping: Should be `0.0.0.0:3307->3306/tcp`
3. Test connection manually:
   ```bash
   mysql -h localhost -P 3307 -u root -pmodelcraft123
   ```

### Environment Variables Not Loading

If environment variables aren't being picked up:

1. Ensure `.env` file is in the correct location
2. Check pytest is loading the `.env` file (see root `conftest.py`)
3. Try setting variables directly in the command line

### Reset Configuration

To force reload configuration from environment:

```python
from design.common.db_config import reset_test_db_config

reset_test_db_config()
config = get_test_db_config()  # Will reload from environment
```

## Best Practices

1. **Use defaults for most tests**: Don't specify connection parameters unless necessary
2. **Document overrides**: Add comments explaining why you're overriding defaults
3. **Clean up resources**: Use fixtures like `created_clusters` for automatic cleanup
4. **Environment-specific configs**: Use `.env` files for different environments (dev, CI, etc.)

## Related Documentation

- [Test Data Builders](./test_data.py) - Helper functions for creating test data
- [Test Fixtures](../conftest.py) - Shared test fixtures
- [Assertions](./assertions.py) - Test assertion helpers
