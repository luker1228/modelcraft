# API Contract First Workflow

`api/` is the single source of truth for contracts shared by backend and frontend.

## Scope

- `graph/` : GraphQL contracts (Org + Project domains)
- `openapi/` : REST/OpenAPI contracts
- `graph/project/schema/runtime-json-schema-contract.md` : Runtime JSON Schema protocol used by `model-editor`

## Team Workflow (Required)

1. Define or update contract in `api/` first.
2. Review contract changes with frontend/backend together.
3. Implement backend and frontend against the same contract.
4. If `api/` changed in backend repo, run subtree push:
   - `git subtree push --prefix=api contracts main`
5. Frontend syncs latest contract with `front-contract-pull` skill.

## Change Rules

- Contract changes are authoritative; implementation follows contract.
- Any breaking change must include migration notes and compatibility strategy.
- Do not edit frontend `contract/` directly; always sync from backend `api/`.

## Quick Links

- `graph/project/schema/runtime-json-schema-contract.md`
- `graph/project/schema/`
- `openapi/README.md`
