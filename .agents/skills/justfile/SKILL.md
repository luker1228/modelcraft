---
name: justfile
description: >
  Execute and manage justfile recipes for the ModelCraft project. Use when the user wants to:
  (1) Run build tasks (build, build-prod, build-all),
  (2) Start or stop the application (run, dev, start, stop, restart),
  (3) Run unit or integration tests (test-unit, test-coverage, test-unit-pkg),
  (4) Check code quality (lint, lint-fix, check-all),
  (5) Generate code (generate-gql, generate-oapi, generate-sqlc),
  (6) Manage the database (db up, db create, db drop, db login),
  (7) Work with Docker (docker-up, docker-compose-up, docker-clean),
  (8) Deploy environments (deploy-infra, deploy-app, deploy-all),
  (9) Manage ports (port-kill, port-check),
  (10) Manage environments (env-list, env-switch, env-create),
  (11) Ask "how do I run X", "run the tests", "build the project", "apply migrations", etc.
---

# Justfile Skill

Run `just <recipe>` from the project root. Use `just --list` to see all recipes.

## Reference Documentation

**Full command reference is in ai-metadata:**
- [justfile-guide.md](../../../ai-metadata/backend/tools/justfile-guide.md) - Complete command reference with all recipes organized by category

## Critical Restrictions

- **NEVER edit `api/openapi/openapi.yaml` directly** - edit module files then `just generate-oapi`
- **Use `just clean-gql` with caution** - may lose custom resolver code

## Common Workflows

**Start development:**
```bash
just run                     # normal run
just run force=true          # kill port 8080 first
just run stdout=true         # console output
just dev                     # hot reload mode
```

**Run tests:**
```bash
just test-unit               # all unit tests
just test-unit-pkg ./internal/domain/project  # specific package
just test-coverage           # coverage check (requires 95%)
```

**Code quality before commit:**
```bash
just lint                    # run linter
just lint-fix                # auto-fix
just check-all               # lint + test
```

**Lint workflow (when `just lint` reports issues):**
1. Run `just lint-fix` first to auto-fix fixable issues
2. Re-run `just lint` to verify remaining issues
3. Manually fix any issues that `lint-fix` could not resolve

**After modifying GraphQL schema (`api/graph/schema/*.graphql`):**
```bash
just generate-gql
```

**After modifying OpenAPI modules (`api/openapi/*.yaml`):**
```bash
just generate-oapi
```

**Apply DB schema changes:**
```bash
just db                      # apply schema (default)
just db up .env.autotest     # for test DB
```

**Deployment:**
```bash
just deploy-all              # start all services (infra + app)
just deploy-infra            # start MySQL, Redis only
just deploy-app              # start application only
```

## Parameter Syntax

just uses `key=value` syntax:
- `just run force=true stdout=true` - kill port 8080 and log to console
- `just test-unit-pkg ./internal/domain/project`
- `just port-kill 3000`
- `just db up .env.autotest`
- `just test-user-cleanup <user_id>`
- `just log-cat <request_id>` - filter logs by request ID

## Full Command Reference

See [ai-metadata/backend/tools/justfile-guide.md](../../../ai-metadata/backend/tools/justfile-guide.md) for the complete list of all available recipes organized by category.
