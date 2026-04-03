# Change: Replace Makefile with Taskfile

## Why

The project currently uses GNU Make (Makefile) for build automation, which has several limitations:
- Make syntax is arcane and error-prone (tabs vs spaces, shell escaping issues)
- Limited cross-platform support (requires GNU Make on non-Unix systems)
- No built-in variable templating or task dependencies visualization
- Tests already use Taskfile (tests/Taskfile.yml), creating inconsistent tooling

Taskfile (go-task) provides a modern, YAML-based alternative with better ergonomics, clearer syntax, and consistent behavior across platforms. It's already proven in the tests directory and aligns with Go ecosystem conventions.

## What Changes

- Create root `Taskfile.yml` to replace `Makefile`
- Organize tasks into three main categories:
  1. **Application tasks**: build, run, dev, Docker operations
  2. **Testing tasks**: unit tests, integration tests, automated tests
  3. **Database tasks**: migrations, schema sync, data cleanup
- Integrate with existing `tests/Taskfile.yml` (call test tasks from root)
- Preserve all existing Make target names for backward compatibility
- Add improved task descriptions and help documentation
- Remove `Makefile` after migration is complete

## Impact

### Affected Specs
- `deployment` - Docker and local deployment commands
- `database-migration` - Database management commands
- New spec: `build-automation` - Build, test, and code quality tasks

### Affected Code
- Root `Makefile` (436 lines) → replaced by `Taskfile.yml`
- `tests/Taskfile.yml` (205 lines) → integrated with root Taskfile
- CI/CD scripts may need updating if they call `make` commands
- Documentation (`CLAUDE.md`, `docs/`) needs command updates

### Breaking Changes
None - all existing `make` commands will have equivalent `task` commands with the same names.

### Migration Path
1. Install go-task: `go install github.com/go-task/task/v3/cmd/task@latest`
2. Use `task` instead of `make`: `task build` instead of `make build`
3. Backward compatibility: Keep `Makefile` as a thin wrapper initially (optional)
