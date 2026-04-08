---
name: contract-sync
description: >
  Sync API contracts from the Go backend to the frontend. Use this skill when:
  - The user says "sync contract", "sync api", "update contract", "refresh contract"
  - Backend GraphQL schema (.graphql files) or OpenAPI specs (.yaml) have changed and the frontend needs updating
  - The user reports that `modelcraft-front/contract/` is out of date or missing files
  - Before running frontend code generation that depends on the contract files
  Do NOT trigger this skill for questions about what the contracts contain — only trigger on actual sync requests.
---

# Contract Sync

Copies API contracts from `modelcraft-backend/api/` into `modelcraft-front/contract/` so the frontend can use local schema files without reaching across submodule boundaries.

> **Note:** In a git-subtree workflow this step is replaced by `git subtree pull`. This script is the local-dev alternative — use it when you just need a quick sync without committing.

## What Gets Synced

```
modelcraft-backend/api/              →   modelcraft-front/contract/
├── graph/org/schema/*.graphql       →   graph/org/schema/
├── graph/project/schema/*.graphql   →   graph/project/schema/
└── openapi/*.yaml                   →   openapi/
    (openapi-root, auth, org, user, webhook, common)
```

**Excluded** (backend-only, not needed by frontend):
- `openapi/openapi.yaml` — generated bundle
- `openapi/oapi-codegen.yaml` — Go codegen config
- `openapi/examples/` — example payloads
- `openapi/README.md` — backend maintenance notes

## How to Run

```bash
bash .agents/skills/front/contract-sync/scripts/sync-contracts.sh
```

The script auto-detects the monorepo root (the directory containing both `modelcraft-front/` and `modelcraft-backend/`), so it works regardless of which directory you run it from.

## After Sync

Check the result:

```bash
find modelcraft-front/contract/ -type f | sort
```

Expected files:
- `graph/org/schema/` — 5–6 `.graphql` files (base, schema, project, user_management, permission, api_key, …)
- `graph/project/schema/` — 6–7 `.graphql` files (base, schema, model, field, enum, cluster, logical_foreign_key, …)
- `openapi/` — 6 `.yaml` files (openapi-root, auth, org, user, webhook, common)
