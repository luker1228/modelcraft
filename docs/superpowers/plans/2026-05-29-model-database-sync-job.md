# Model Database Sync Job Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an async "sync database" job that imports missing tables as models, syncs schema for existing models, and exposes job status for polling from the database management page.

**Architecture:** Extend the existing `modeldatabase` module with a persisted sync-job entity and repository, then add an app-layer orchestrator that reuses `ReverseEngineerAppService.ImportModel` and `ModelDesignAppService.SyncModelSchemaFromJSON`. Frontend starts a job from the databases table and polls the job status until completion.

**Tech Stack:** Go, gqlgen GraphQL, sqlc, MySQL schema files, Next.js App Router, Apollo Client, shadcn/ui

---

### Task 1: Backend Contract And Persistence

**Files:**
- Modify: `modelcraft-backend/api/graph/project/schema/database.graphql`
- Modify: `modelcraft-backend/db/schema/mysql/16_model_database.sql`
- Modify: `modelcraft-backend/db/queries/model_database.sql`
- Test: `modelcraft-backend/internal/infrastructure/repository/*_test.go`

- [ ] Add sync-job GraphQL types, query, and mutation to `database.graphql`.
- [ ] Add `model_database_sync_job` table definition to `16_model_database.sql`.
- [ ] Add sqlc queries for create/get/update/list-running job operations in `model_database.sql`.
- [ ] Run the generated-code commands and confirm generated GraphQL/sqlc outputs update cleanly.

### Task 2: Backend Orchestration

**Files:**
- Modify: `modelcraft-backend/internal/domain/modeldatabase/*.go`
- Modify: `modelcraft-backend/internal/app/modeldatabase/*.go`
- Modify: `modelcraft-backend/internal/infrastructure/repository/*.go`
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/*.go`
- Modify: `modelcraft-backend/internal/interfaces/http/routes.go`
- Test: `modelcraft-backend/internal/app/modeldatabase/*_test.go`

- [ ] Write failing app-layer tests for starting a sync job, rejecting duplicate running jobs, and continuing after per-table failures.
- [ ] Add sync-job domain types and repository interfaces.
- [ ] Implement repository persistence and job status updates.
- [ ] Implement app service methods to start a job, execute it in background, ensure the "数据库导入" group exists, and return job snapshots.
- [ ] Wire new GraphQL resolvers and service dependencies.
- [ ] Run targeted backend tests and `go build ./...`.

### Task 3: Frontend Polling UI

**Files:**
- Modify: `modelcraft-front/src/api-client/project/model-database-graphql-docs.ts`
- Modify: `modelcraft-front/src/web/hooks/model-database/use-model-databases.ts`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/page.tsx`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/_components/*`

- [ ] Add GraphQL documents and generated typings for starting/querying sync jobs.
- [ ] Add hooks to start a sync job and poll job status.
- [ ] Add the row action, confirmation dialog, progress badge, and result dialog to the databases page.
- [ ] Verify the page handles loading, partial success, and completed states cleanly.

### Task 4: Verification

**Files:**
- Test only

- [ ] Run targeted backend tests for the new app/repository code.
- [ ] Run GraphQL/sqlc generation and confirm no dirty generated diffs are broken.
- [ ] Run frontend type/lint checks for the touched files.
- [ ] Summarize any residual risks, especially around process-local background execution.
