## 1. Database Schema Changes

- [x] 1.1 Create `model_field_enum_associations` table with foreign key constraints
- [x] 1.2 Remove `enum_name` column from `field_definitions` table
- [x] 1.3 Create data migration script to migrate existing enum_name data to association table

## 2. Domain Layer Implementation

- [x] 2.1 Remove `EnumName` field from `FieldDefinition` struct in `internal/domain/modeldesign/field_definition.go`
- [x] 2.2 Create `FieldEnumAssociation` entity in `internal/domain/modeldesign/field_enum_association.go`
- [x] 2.3 Create `FieldEnumAssociationRepository` interface in `internal/domain/modeldesign/field_enum_association_repository.go`
- [x] 2.4 Update `validateEnumField()` method to check for enumConfig instead of enumName

## 3. Infrastructure Layer Implementation

- [x] 3.1 Remove `EnumName` field mapping from `FieldDefinitionPO` in `internal/infrastructure/repository/field_definition_model.go`
- [x] 3.2 Create `FieldEnumAssociationPO` model in `internal/infrastructure/repository/field_enum_association_model.go`
- [x] 3.3 Implement `GormFieldEnumAssociationRepository` in `internal/infrastructure/repository/field_enum_association_repository.go`
- [x] 3.4 Add CRUD operations for field-enum associations

## 4. Application Layer Changes

- [ ] 4.1 Update `FieldService.CreateFields()` to handle enumConfig parameter
- [ ] 4.2 Implement logic to create or connect enums based on `connectEnum` flag
- [ ] 4.3 Create association records in `model_field_enum_associations` table
- [ ] 4.4 Update `FieldService.DeleteField()` to rely on CASCADE delete for associations
- [x] 4.5 Update `EnumService.DeleteEnum()` to check for field associations before deletion

## 5. GraphQL Schema Updates

- [x] 5.1 Remove `enum` field from `ValidationConfigInput` in `api/graph/schema/field.graphql`
- [x] 5.2 Add `EnumConfigInput` type with fields: enumName, options, description, connectEnum
- [x] 5.3 Add `EnumOptionInput` type with fields: key, value, order, description (already exists in enum.graphql)
- [x] 5.4 Add `enumConfig` field to `AddFieldInput`
- [x] 5.5 Run `make generate-gql` to regenerate GraphQL code

## 6. GraphQL Resolver Updates

- [x] 6.1 Update `addFields` resolver to handle `enumConfig` input (via mapper)
- [ ] 6.2 Implement enum creation logic when `connectEnum = false`
- [ ] 6.3 Implement enum connection logic when `connectEnum = true`
- [ ] 6.4 Update field query resolvers to load enum via association table
- [ ] 6.5 Update error handling for enum-related validations

## 7. Mapper Layer Updates

- [x] 7.1 Remove `enumName` mapping from `FieldMapper` in `internal/interfaces/mapper/field_mapper.go`
- [x] 7.2 Update DTO conversion logic to handle enumConfig
- [x] 7.3 Create mapper for `FieldEnumAssociation` if needed

## 8. Testing

- [ ] 8.1 Update unit tests for `FieldDefinition` validation
- [ ] 8.2 Add unit tests for `FieldEnumAssociation` entity
- [ ] 8.3 Add integration tests for field creation with enumConfig
- [ ] 8.4 Add tests for enum connection (connectEnum=true) scenarios
- [ ] 8.5 Add tests for enum creation (connectEnum=false) scenarios
- [ ] 8.6 Add tests for field deletion with cascade association cleanup
- [ ] 8.7 Add tests for enum deletion prevention when referenced by fields
- [ ] 8.8 Update GraphQL integration tests in `tests/automated/`

## 9. Documentation Updates

- [ ] 9.1 Update API documentation for enumConfig usage
- [ ] 9.2 Document migration guide for clients using old API
- [ ] 9.3 Update CHANGELOG.md with breaking changes

## Summary

✅ **Completed**: Domain Layer, Infrastructure Layer, GraphQL Schema, Mapper Layer, Build verification
⏳ **Remaining**: Application Layer enum handling logic, Field query enum loading, Tests, Documentation