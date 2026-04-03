# Design: Replace Makefile with Taskfile

## Context

The ModelCraft project uses GNU Make for build automation with 436 lines of Makefile containing 50+ targets. The tests directory already uses Taskfile (tests/Taskfile.yml) successfully. This creates tooling inconsistency and misses opportunities for better developer experience.

### Stakeholders
- Developers: Need consistent, easy-to-use build commands
- CI/CD: Requires reliable, reproducible build automation
- New contributors: Benefit from clearer, more readable build scripts

### Constraints
- Must maintain backward compatibility with existing workflows
- Must support Linux and macOS development environments
- Must integrate with existing Docker and testing infrastructure
- Cannot introduce complex dependencies

## Goals / Non-Goals

### Goals
- Replace Makefile with Taskfile for consistent tooling
- Improve readability and maintainability of build automation
- Provide better help documentation and task discovery
- Maintain 100% functional compatibility with existing commands
- Integrate root and tests task files cleanly

### Non-Goals
- Not changing functionality of any existing commands
- Not adding new features beyond task migration
- Not modifying CI/CD pipelines (only updating commands if needed)
- Not supporting Windows (project targets Linux/macOS)

## Decisions

### Decision 1: Use go-task/task v3
**Rationale**: go-task is the de facto standard task runner in the Go ecosystem, already proven in tests directory, with excellent documentation and active maintenance.

**Alternatives Considered**:
- Keep Makefile: Rejected due to syntax issues and inconsistency with tests/
- Just: Less mature, smaller ecosystem
- Mage: Requires Go code instead of declarative config

### Decision 2: Root Taskfile calls tests/Taskfile.yml
**Approach**: Use Taskfile's task calling syntax to delegate test-related tasks to tests/Taskfile.yml rather than duplicating them.

**Example**:
```yaml
test-python:
  desc: Run Python tests
  dir: tests
  cmds:
    - task: test-python
```

**Rationale**: Maintains separation of concerns, avoids duplication, respects existing test infrastructure.

### Decision 3: Organize tasks into logical groups
**Categories**:
1. **Application**: build, run, dev, start, stop, restart
2. **Docker**: docker-build, docker-run, docker-compose-*, docker-up, docker-clean
3. **Deployment**: deploy-local, deploy-docker, deploy-stop
4. **Testing**: test-unit*, test-python, auto-test*
5. **Database**: migrate-*, clean-data
6. **Code Quality**: fmt, lint, vet, check-all
7. **Tools**: install-*, generate-gql

**Rationale**: Mirrors the three categories mentioned in requirements (app, test, db) with logical additions. Makes help output more navigable.

### Decision 4: Preserve exact command names
All existing Makefile targets get identically-named Taskfile tasks (e.g., `make build` → `task build`).

**Rationale**: Zero learning curve, existing documentation and muscle memory remain valid, CI/CD changes are minimal.

### Decision 5: Use YAML variables for shared values
```yaml
vars:
  BINARY_NAME: modelcraft
  DOCKER_IMAGE: modelcraft:latest
  SERVER_PATH: ./cmd/server
  GQLGEN_VERSION: v0.17.83
```

**Rationale**: Taskfile's variable system is clearer than Make's, easier to override, better for templating.

### Decision 6: No Makefile wrapper (clean break)
After validation, remove Makefile entirely rather than keeping it as a thin wrapper.

**Rationale**:
- Taskfile installation is trivial (`go install`)
- Maintaining two build systems adds complexity
- Tests already require Taskfile, so dependency exists
- Clear migration path reduces confusion

**Alternative**: Could keep `Makefile` that just calls `task` targets for grace period. Rejected because it adds maintenance burden and delays full migration.

## Risks / Trade-offs

### Risk: Developers must install go-task
**Mitigation**:
- Add installation instructions to README and CLAUDE.md
- go-task is a single binary, easy to install via `go install`
- Tests already require it, so not a new dependency

### Risk: CI/CD pipeline breakage
**Mitigation**:
- Update CI/CD configs to use `task` instead of `make`
- Test in CI before merging
- Rollback plan: revert commit if issues arise

### Risk: Muscle memory with `make` commands
**Mitigation**:
- Keep exact same command names (`task build` vs `make build`)
- Document migration clearly
- Consider shell alias: `alias make=task` during transition

### Trade-off: YAML verbosity vs clarity
Taskfile YAML is more verbose than Make syntax, but significantly more readable and maintainable. The trade-off favors long-term maintainability over brevity.

## Migration Plan

### Phase 1: Implementation (1-2 days)
1. Create Taskfile.yml with all targets from Makefile
2. Test all tasks locally (build, test, deploy, docker, db)
3. Update documentation (CLAUDE.md, relevant docs/)

### Phase 2: Validation (1 day)
1. Run full test suite using Taskfile commands
2. Test Docker deployment workflows
3. Verify database migration tasks
4. Check task help output and discoverability

### Phase 3: Integration (1 day)
1. Update CI/CD configuration files (if any call `make`)
2. Test CI/CD pipelines with new commands
3. Update any automation scripts

### Phase 4: Cleanup (1 day)
1. Remove Makefile after all validation passes
2. Update .gitignore if needed
3. Announce migration to team
4. Monitor for issues

### Rollback Plan
If critical issues arise:
1. Revert Taskfile commit
2. Restore Makefile
3. Document issues for future attempt

## Implementation Notes

### Taskfile Structure

```yaml
version: "3"

vars:
  BINARY_NAME: modelcraft
  # ... other vars

includes:
  test:
    taskfile: ./tests/Taskfile.yml
    dir: ./tests

tasks:
  default:
    desc: Show help
    cmds:
      - task: help

  build:
    desc: Build the application
    cmds:
      - go build -o bin/{{.BINARY_NAME}} {{.SERVER_PATH}}

  # ... other tasks organized by category
```

### Task Calling Patterns

**Direct shell command**:
```yaml
build:
  cmds:
    - go build -o bin/{{.BINARY_NAME}} {{.SERVER_PATH}}
```

**Multi-step with variables**:
```yaml
test-unit-pkg:
  vars:
    PKG: '{{.PKG | default ""}}'
  preconditions:
    - sh: test -n "{{.PKG}}"
      msg: "Please specify package: task test-unit-pkg PKG=./internal/domain/project"
  cmds:
    - go test -v -race -timeout=5m {{.PKG}}
```

**Calling subtasks**:
```yaml
test-all:
  desc: Run all tests (Go + Python)
  cmds:
    - task: test-unit
    - task: test:test-python
```

### Help Output Design

```bash
$ task help

📋 ModelCraft Build Automation

🏗️  Application:
  build          Build the application
  run            Run the application
  dev            Run with hot reload (requires Air)

🐳 Docker:
  docker-build   Build Docker image
  docker-up      Start all services with Docker Compose

🧪 Testing:
  test-unit      Run Go unit tests
  test-all       Run all tests (Go + Python)

🗄️  Database:
  migrate-up     Apply database migrations
  clean-data     Clean test data

🎨 Code Quality:
  fmt            Format code
  lint           Run linter
  check-all      Run all checks
```

## Open Questions

None - implementation approach is straightforward based on existing Makefile and tests/Taskfile.yml patterns.
