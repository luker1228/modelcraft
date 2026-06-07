# Constrained GraphQL Runtime ModelCraft Evidence Map

**Date:** 2026-06-07

| Claimed Concept | ModelCraft Evidence | Why It Matters |
|---|---|---|
| Design-time / runtime separation | `ai-metadata/backend/design/core-principles.md` | Supports the runtime-boundary generation story instead of static CRUD endpoints. |
| Runtime GraphQL as the access protocol | `ai-metadata/backend/design/core-principles.md` | Confirms the system uses GraphQL rather than REST as the runtime carrier. |
| Developer / EndUser access split | `ai-metadata/backend/development/developer-enduser-system.md` | Supports主体差异化访问 and permission-driven boundary generation. |
| AI-usable data access path | `ai-metadata/cli/README.md` | Shows the project already values schema discovery and structured access for agents. |
| GraphQL schema & runtime communities | `graphify-out/GRAPH_REPORT.md` | Supports that the codebase already has distinct runtime and schema-related architecture hubs. |
| Model-driven runtime docs effort | `docs/superpowers/specs/2026-06-06-model-scoped-runtime-api-doc-design.md` | Supports the claim that runtime ability exposure and query guidance are project-native concerns. |

## Review Notes

1. This file is not the patent itself.
2. This file exists to help later reviewers prove that the application is grounded in a real system.
3. If later code anchors are needed, add them after targeted source inspection rather than guessing.
