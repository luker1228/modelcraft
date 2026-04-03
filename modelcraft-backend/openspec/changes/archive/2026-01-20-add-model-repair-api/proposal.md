# Change: Add Model Repair API

## Why
Currently, models can become out of sync with their underlying database resources through various scenarios (manual database operations, failed deployments, external schema changes). When this happens, there is no mechanism to detect and repair the drift between the model definition in the platform database and the actual table structure in the customer database. Users need a way to validate and repair models that have missing or mismatched underlying resources.

## What Changes
- Add a new GraphQL mutation `repairModel` to detect and repair model schema drift
- Implement three repair modes:
  1. **Dry run mode**: Detect issues without making any changes
  2. **Additive mode**: Add missing tables and fields without deleting existing resources
  3. **Full sync mode**: Add missing resources and optionally remove extra fields (with safety checks)
- Add domain service to compare model definitions with actual database schema
- Create comprehensive result payload showing detected issues, executed operations, and before/after comparison
- Reuse existing components: SchemaIntrospector, DDLConverter, TypeMapper, DeploymentImpl

## Impact
- Affected specs: modeldesign-repair (new spec capability)
- Affected code:
  - `api/graph/schema/model.graphql` - Add repairModel mutation types
  - `internal/domain/modeldesign/` - Add comparison service and change types
  - `internal/app/modeldesign/` - Add repair use case service
  - `internal/infrastructure/database/` - Enhance schema introspection for comparison
  - `internal/interfaces/graphql/model.resolvers.go` - Add repairModel resolver
