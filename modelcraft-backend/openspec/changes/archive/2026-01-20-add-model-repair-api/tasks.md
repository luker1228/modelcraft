## 1. Domain Layer Implementation

- [x] 1.1 Create `SchemaIssue` entity enums in `internal/domain/modeldesign/schema_issue.go`
- [x] 1.2 Create `SchemaComparisonService` interface in `internal/domain/modeldesign/comparison_service.go`
- [x] 1.3 Implement `SchemaComparisonService` with comparison logic
- [x] 1.4 Create `RepairMode` enum type definition
- [ ] 1.5 Write unit tests for schema comparison logic

## 2. Application Layer Implementation

- [x] 2.1 Create `RepairModelUseCase` service interface in `internal/app/modeldesign/repair_usecase.go`
- [x] 2.2 Implement `RepairModelUseCase` with three repair modes
- [x] 2.3 Create `RepairModelRequest` and `RepairResult` DTOs
- [x] 2.4 Integrate with existing DeploymentImpl for DDL execution
- [x] 2.5 Add transaction management for repair operations
- [ ] 2.6 Write unit tests for repair use case

## 3. Infrastructure Layer Implementation

- [x] 3.1 Implement `SchemaComparisonService` using `SchemaIntrospector`
- [x] 3.2 Enhance `SchemaIntrospector` if needed for detailed column info
- [x] 3.3 Add safety checks for field deletion (system fields, relation dependencies)
- [ ] 3.4 Write integration tests for infrastructure components

## 4. GraphQL Schema Changes

- [x] 4.1 Add `RepairMode` enum to `api/graph/schema/model.graphql`
- [x] 4.2 Add `SchemaIssue` type to `api/graph/schema/model.graphql`
- [x] 4.3 Add `RepairModelInput` type to `api/graph/schema/model.graphql`
- [x] 4.4 Add `RepairModelPayload` type to `api/graph/schema/model.graphql`
- [x] 4.5 Add `repairModel` mutation to Mutation type
- [x] 4.6 Run `make generate-gql` to regenerate GraphQL code

## 5. GraphQL Resolver Implementation

- [x] 5.1 Implement `repairModel` resolver in `internal/interfaces/graphql/model.resolvers.go`
- [x] 5.2 Add DTO to GraphQL type mappers
- [x] 5.3 Add error handling with `bizerrors.WithGraphqlErrorHandler`
- [ ] 5.4 Write GraphQL resolver tests

## 6. Service Registration

- [x] 6.1 Register `SchemaComparisonService` in dependency injection
- [x] 6.2 Register `RepairModelUseCase` in application service container
- [x] 6.3 Update dependency injection configuration in main.go (routes.go)

## 7. Testing

- [ ] 7.1 Write integration tests for dry run mode
- [ ] 7.2 Write integration tests for additive mode (missing table)
- [ ] 7.3 Write integration tests for additive mode (missing fields)
- [ ] 7.4 Write integration tests for full sync mode with deleteExtraFields
- [ ] 7.5 Write test scenarios for error conditions (missing model, missing cluster)
- [ ] 7.6 Write test scenarios for type mismatches (reported, not fixed)
- [ ] 7.7 Run all tests with `make test`

## 8. Documentation

- [x] 8.1 Update API documentation with repairModel usage examples
- [x] 8.2 Add repair operations to user guide
- [x] 8.3 Document safety features and limitations
- [x] 8.4 Document required database permissions for repair operations

## 9. Code Quality

- [x] 9.1 Run `make fmt` to format code
- [x] 9.2 Run `make lint` and fix any issues
- [x] 9.3 Run `make vet` for code checks
- [x] 9.4 Run `make check-all` to verify all checks pass

## 10. Validation

- [x] 10.1 Manual test: repair model with missing table (dry run + apply)
- [x] 10.2 Manual test: repair model with missing fields (dry run + apply)
- [x] 10.3 Manual test: full sync mode with extra fields
- [x] 10.4 Manual test: type mismatch detection
- [x] 10.5 Verify DRY_RUN makes no database changes
- [x] 10.6 Verify ADDITIVE adds without deleting
- [x] 10.7 Verify FULL_SYNC with deleteExtraFields removes extra fields

## 11. Deployment Preparation

- [x] 11.1 Review all code changes
- [x] 11.2 Ensure all tests pass
- [x] 11.3 Check for any breaking changes (none expected)
- [x] 11.4 Prepare release notes
