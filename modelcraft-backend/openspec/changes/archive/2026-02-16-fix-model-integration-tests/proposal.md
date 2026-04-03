# Proposal: Fix Model Integration Tests

## Problem

The model integration tests in `tests/design/model/` are failing with multiple critical issues:

1. **GraphQL Query Syntax Errors**: Duplicate parameter names in GraphQL queries (e.g., `projectName` appears twice)
2. **Fixture Data Structure Errors**: Test fixtures incorrectly track and cleanup resources
3. **Cluster Reference Errors**: Tests attempt to create models referencing non-existent clusters
4. **Cleanup Logic Errors**: Resource cleanup uses incorrect identifiers (ID vs name)
5. **Fixture Return Value Mismatches**: Tests expect fields that fixtures don't provide

### Test Execution Results

Running `pytest tests/design/model/ -v` shows:
- **16 failures** out of 23 tests
- **5 errors** (fixture setup failures)
- **2 passes** (only non-cluster-dependent tests pass)

### Critical Error Patterns

**Error Pattern 1: TypeError in fixture setup**
```
TypeError: 'NoneType' object is not subscriptable
  at test_model fixture line 243
```
The fixture tries to access `model["projectName"]` when model creation failed and returned None.

**Error Pattern 2: Cluster not found**
```
AssertionError: assert {'__typename': 'InvalidModelInput',
                         'message': '数据库集群不存在: model-test-cluster-*'} is None
```
Models are created with cluster references, but the clusters don't actually exist in the database.

**Error Pattern 3: KeyError in tests**
```
KeyError: 'projectName'
```
Tests expect `default_project["projectName"]` but the fixture returns `default_project["name"]`.

**Error Pattern 4: Cleanup failures**
```
⚠ Failed to cleanup cluster default/019c6465-*:
   {'message': 'database cluster not found: 019c6465-*'}
```
Cleanup attempts to delete clusters by ID, but the mutation expects name.

## Solution Overview

Fix all test failures by correcting:

1. **GraphQL Query Syntax**: Remove duplicate parameters in all queries
2. **Fixture Error Handling**: Add proper null checks and error validation
3. **Cluster Creation Flow**: Ensure clusters are actually created before model creation
4. **Cleanup Logic**: Use correct identifiers (name) for resource deletion
5. **Fixture Consistency**: Align fixture return values with test expectations

## Scope

**In Scope:**
- Fix GraphQL query definitions in test files
- Correct fixture tracking and cleanup logic
- Add validation for cluster creation success
- Standardize fixture return values
- Ensure all 23 tests pass

**Out of Scope:**
- Changes to application code (only test fixes)
- New test cases (only fix existing tests)
- Performance optimizations
- Test framework changes

## Impact

- **Test Reliability**: All model integration tests will pass consistently
- **Developer Experience**: Clear test failures when issues occur
- **CI/CD**: Automated tests can gate deployments effectively
- **No Breaking Changes**: Only test code modifications

## Dependencies

- Requires running ModelCraft server on localhost:8080
- Requires MySQL test database on port 3307
- Requires valid Casdoor test credentials in .env
- No code dependencies or API changes

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Tests still fail after fixes | High | Thoroughly analyze each failure pattern and test locally |
| Cleanup logic breaks | Medium | Validate cleanup with proper error handling and logging |
| Fixture changes break other tests | Low | Only modify model test fixtures, not shared fixtures |

## Success Criteria

- [ ] All 23 tests in `tests/design/model/` pass
- [ ] No fixture setup errors (TypeError, KeyError)
- [ ] No cluster reference errors in model creation
- [ ] Resource cleanup succeeds without warnings
- [ ] Tests can run repeatedly without conflicts
