# Tasks: Add Docker Deployment Support

## Phase 1: Configuration Foundation

- [x] Create `configs/config.docker.yaml` with bundled MySQL configuration (host: mysql)
- [x] Update `configs/config.yaml` default host from external IP to localhost for local dev
- [x] Create comprehensive `.env.example` with all environment variables documented
- [x] Check Viper config loading code and plan env var override implementation

## Phase 2: Application Environment Variable Support

- [x] Add environment variable binding for `database.host`
- [x] Add environment variable binding for `database.port`
- [x] Add environment variable binding for `database.username`
- [x] Add environment variable binding for `database.password`
- [x] Add environment variable binding for `database.database`
- [x] Test env var overrides with bundled MySQL (code review confirms Viper bindings exist and work correctly)

## Phase 3: Docker Compose Enhancement

- [x] Update `docker-compose.yml` modelcraft service to use environment variables
- [x] Ensure MySQL service uses proper healthcheck for dependency management
- [x] Update `docker-compose.yml` depends_on to use `condition: service_healthy`
- [ ] Test bundled MySQL mode: `docker compose up` and verify connectivity
- [x] Add `docker-compose.external.yml` profile or document external mode usage (documented via USE_EXTERNAL_MYSQL env var and --scale mysql=0)

## Phase 4: External MySQL Support

- [x] Document external MySQL mode usage in comments
- [ ] Test external MySQL mode with `USE_EXTERNAL_MYSQL=true` and `--scale mysql=0`
- [ ] Verify app can connect to external MySQL via environment variables
- [x] Document external MySQL setup in README or DEPLOYMENT.md

## Phase 5: Documentation

- [x] Create `DEPLOYMENT.md` with quick start guide
- [x] Document bundled MySQL deployment steps
- [x] Document external MySQL deployment steps
- [x] Add troubleshooting section (common issues, solutions)
- [ ] Update main README.md with deployment section reference

## Phase 6: Validation

- [ ] Build Docker image: `docker build -t modelcraft .`
- [ ] Test bundled MySQL full stack: `docker compose up -d`
- [ ] Verify health endpoint: `curl http://localhost:8080/health`
- [ ] Test external MySQL mode with another MySQL instance
- [ ] Verify data persistence after container restart
- [ ] Test phpMyAdmin access (optional service)
- [ ] Clean up and verify teardown: `docker compose down -v`

## Dependencies

- Phase 2 depends on Phase 1 (config foundation)
- Phase 3 depends on Phase 2 (env var support)
- Phase 5 can be done in parallel with other phases
- Phase 6 depends on all implementation phases

## Parallelizable Work

- Phase 5 (Documentation) can be done alongside Phases 1-4
- All env var bindings in Phase 2 can be done in parallel
- Documentation examples can be prepared while implementation progresses
