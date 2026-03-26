---
name: contract-sync
description: Sync Go backend API contracts (GraphQL schemas + OpenAPI specs) from modelcraft-go/api/ to modelcraft-front/contract/. Use when the user asks to sync contracts, copy API definitions, update frontend API contracts, or refresh the contract directory. Triggers on phrases like "sync contract", "copy api", "sync api", "update contracts".
---

# Contract Sync

Keep `modelcraft-front/contract/` in sync with `modelcraft-go/api/` so the frontend has a local copy of API contracts without cross-repo imports.

## Source Structure

GraphQL is now split into two endpoint namespaces (`org` and `project`):

```
modelcraft-go/api/
├── graph/
│   ├── org/schema/        # Org-scoped GraphQL endpoint
│   │   ├── schema.graphql
│   │   ├── base.graphql
│   │   ├── project.graphql
│   │   ├── user_management.graphql
│   │   └── permission.graphql
│   └── project/schema/    # Project-scoped GraphQL endpoint
│       ├── schema.graphql
│       ├── base.graphql
│       ├── model.graphql
│       ├── field.graphql
│       ├── enum.graphql
│       ├── cluster.graphql
│       └── logical_foreign_key.graphql
└── openapi/               # REST specs (.yaml) — auth, org, user, webhook
    ├── openapi-root.yaml  # Entry point (references modules)
    ├── common.yaml        # Shared schemas
    ├── auth.yaml          # /api/auth/*
    ├── org.yaml           # /api/org/*
    ├── user.yaml          # /api/user/*
    └── webhook.yaml       # /api/webhook/*
```

## Sync Workflow

Run the sync script:

```bash
bash .codebuddy/skills/contract-sync/scripts/sync-contracts.sh
```

The script:
1. Cleans `contract/` entirely (no stale files)
2. Copies `graph/org/schema/*.graphql` → `contract/graph/org/schema/`
3. Copies `graph/project/schema/*.graphql` → `contract/graph/project/schema/`
4. Copies `openapi/` module YAMLs → `contract/openapi/`
5. Excludes generated files: `openapi.yaml` (bundled), `oapi-codegen.yaml` (codegen config)

## After Sync

Verify the contract directory contains the expected files:

```bash
find contract/ -type f
```

Expected output: `.graphql` files under `contract/graph/org/schema/` and `contract/graph/project/schema/`, plus module `.yaml` files under `contract/openapi/`.
