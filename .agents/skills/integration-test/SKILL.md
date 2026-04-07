---
name: integration-test
description: >
  Run ModelCraft design-time integration tests using pytest with HTML report generation.
  Use when the user requests to: (1) Run integration tests for design-time modules
  (tests/design/), (2) Test specific modules like project, model, field, cluster, or enum,
  (3) Generate test reports, (4) Validate design-time functionality. This skill handles
  test execution and reporting without automatically fixing failures.
---

# Integration Test

Run ModelCraft design-time integration tests with pytest and generate HTML reports.

## Quick Start

**Check current environment:**
```bash
task env:current
```

**Switch environment (if needed):**
```bash
task env:list                    # List available environments
task env:switch NAME=autotest    # Switch to autotest environment
```

**Test specific module:**
```bash
scripts/run_integration_test.py tests/design/project
```

**Test all design-time modules:**
```bash
scripts/run_integration_test.py tests/design/
```

**Clear all test reports:**
```bash
scripts/run_integration_test.py --clear
```

> HTTP server always starts automatically on port **10001** after tests run. Any existing server on that port is killed first.

## Workflow

When user requests integration testing:

1. **Identify the action** from user's request:
   - Run tests: specific module or all modules
   - Clear reports: user mentions "clear", "delete reports", "clean up"
   - Env file: user mentions `.env.autotest` or specific env file

2. **For test execution:**
   - Identify test scope: specific module or all modules
   - **Check current environment:**
     ```bash
     task env:current
     ```
   - **If user wants to switch environment:**
     - Use `task env:list` to show available environments
     - Switch with `task env:switch NAME=<env_name>` (e.g., `task env:switch NAME=autotest`)
     - After switching, `.env` will be a symlink to the target environment file
   - **IMPORTANT:** The test script always reads from `.env` (the active symlink), so no need to pass `--env-file` parameter

3. **For clearing reports:**
   ```bash
   scripts/run_integration_test.py --clear
   ```
   - Script will list all reports and ask for confirmation
   - User must confirm with 'y' or 'yes'

4. **Run the test script:**
   ```bash
   scripts/run_integration_test.py <test_path>
   ```
   HTTP server on port 10001 starts automatically. Any previous server on that port is killed first.

5. **Present results** to the user:
   - Show pytest terminal output (passed/failed counts)
   - **Always provide the Remote Access URL** (port 10001)
   - If `--clear` used, show deletion confirmation and results
   - **CRITICAL: If tests fail, ONLY report the failures. DO NOT attempt to fix them unless explicitly requested.**

6. **Report format:**
   ```
   Test Results for {module}:
   - Total: X tests
   - Passed: Y tests
   - Failed: Z tests

   Remote Access:
   - URL: http://{server_ip}:10001/test_report_{module}_{timestamp}.html
   - All reports: http://{server_ip}:10001/

   [If failed] Failed tests:
   - test_name_1: error message
   - test_name_2: error message
   ```

## Available Modules

See [references/test_structure.md](references/test_structure.md) for complete test directory structure and available modules.

Common modules:
- `tests/design/project` - Project domain tests
- `tests/design/model` - Model domain tests
- `tests/design/field` - Field domain tests
- `tests/design/cluster` - Cluster domain tests
- `tests/design/enum` - Enum domain tests

## Script Options

Environment
```bash
PROJECT_PATH=./claude/skills/integration-test/
```

```bash
$PROJECT_PATH/scripts/run_integration_test.py [test_path] [--no-verbose] [--clear]
```

- `test_path`: Optional when using `--clear`. Path to test directory or file
- `--no-verbose`: Optional. Disable verbose pytest output
- `--clear`: Optional. Clear all test reports (requires user confirmation)

**Note:** Environment is controlled via `.env` symlink. Use `task env:switch NAME=<env>` before running tests.

## Important Notes

- **Report-Only Mode**: This skill reports test results. Do NOT fix failing tests unless explicitly asked.
- **HTML Reports**: All reports are saved to `tests/reports/` with timestamps
- **HTTP Server**: Always starts on fixed port **10001** after every test run
  - Any existing process on port 10001 is killed before starting
  - Runs in background (non-blocking)
  - All reports in the directory are accessible at `http://{server_ip}:10001/`
- **Prerequisites**: Server must be running on localhost:8080 for tests to pass
- **Test User**: `tests/common/test_user_setup.py` runs automatically before every test run to ensure the test user exists in the database. If setup fails, tests are aborted.

## Examples

**User:** "Run integration tests for project module"
**Action:** 
1. Check current environment: `task env:current`
2. If needed, switch: `task env:switch NAME=autotest`
3. Execute: `scripts/run_integration_test.py tests/design/project`
4. Report results and Remote Access URL at port 10001

**User:** "Test all design-time modules"
**Action:** Execute `scripts/run_integration_test.py tests/design/`, report results and Remote Access URL at port 10001

**User:** "Clear all test reports" or "Delete old reports"
**Action:** Execute `scripts/run_integration_test.py --clear` and show confirmation prompt

**User:** "Run integration tests with .env.autotest" or "用 autotest 环境跑测试"
**Action:** 
1. Execute: `task env:switch NAME=autotest`
2. Execute: `scripts/run_integration_test.py tests/design/`
3. Report results and Remote Access URL at port 10001

**User:** "Why is test_create_project failing?"
**Action:** Read the HTML report or pytest output to identify the failure reason and explain to user
