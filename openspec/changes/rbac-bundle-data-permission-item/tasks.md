## 1. Schema And Storage Redesign

- [x] 1.1 Rewrite `modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql` so `end_user_data_permissions` stores only custom permissions and add a new bundle data permission item table with `UNIQUE(bundle_id, model_id)`
- [x] 1.2 Rewrite `modelcraft-backend/db/schema/mysql/14_rbac_bundle_snapshots.sql` so snapshots store item-level payloads (`modelId`, `grantType`, `preset`, `customPermissionId`, `sortOrder`) instead of permission ID arrays
- [x] 1.3 Update SQL queries and generated db accessors for bundle item CRUD, bundle detail reads, authz expansion, snapshot persistence, and reference checks on custom permissions

## 2. Backend Domain And Application Changes

- [ ] 2.1 Refactor RBAC domain models/repositories to separate preset template definition, custom permission entity, and bundle data permission item
- [ ] 2.2 Remove preset-as-permission persistence flow and implement bundle item bind/replace flows for preset and custom paths
- [ ] 2.3 Update authz aggregation to resolve effective policies from bundle items, expanding preset items at runtime and loading custom permission entities by reference
- [ ] 2.4 Add deletion/update guards so referenced custom permissions cannot be removed in a way that leaves dangling bundle items

## 3. GraphQL Contract Redesign

- [ ] 3.1 Redesign `modelcraft-backend/api/graph/project/schema/rbac.graphql` around item-centric types and mutations, using `item` naming rather than `grant`
- [ ] 3.2 Replace old preset apply/add flows with explicit bind preset item / bind custom item mutations and item-based bundle detail fields
- [ ] 3.3 Regenerate GraphQL code and update resolvers/adapters/error mapping to the new contract

## 4. Frontend Bundle Management Updates

- [ ] 4.1 Update bundle detail queries, generated types, and state mapping to consume data permission items and item-based snapshots
- [ ] 4.2 Redesign the add-permission dialog into two explicit paths: bind preset item and bind custom item
- [ ] 4.3 Surface replace semantics in the UI when a bundle/model already has a configured item and verify the list always renders one item per model

## 5. Verification

- [ ] 5.1 Add backend tests for preset/custom item binding, bundle-model uniqueness, runtime preset expansion, and custom permission reference protection
- [ ] 5.2 Add snapshot and rollback tests covering item payload persistence for both preset and custom cases
- [ ] 5.3 Add frontend tests or interaction verification for item rendering, mode switching, and replace warnings
