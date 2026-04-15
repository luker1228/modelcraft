package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// EnumDefinitionToDomain converts a dbgen.ModelEnum row to a domain EnumDefinition.
func EnumDefinitionToDomain(row dbgen.ModelEnum) (*modeldesign.EnumDefinition, error) {
	var options []modeldesign.EnumOption
	if err := json.Unmarshal(row.Options, &options); err != nil {
		return nil, fmt.Errorf("EnumDefinitionToDomain: unmarshal options: %w", err)
	}

	var createdAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}

	var updatedAt time.Time
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	return &modeldesign.EnumDefinition{
		ID: row.ID,
		ProjectScope: project.ProjectScope{
			OrgName:     row.OrgName,
			ProjectSlug: row.ProjectSlug,
		},
		Name:          row.Name,
		DisplayName:   row.DisplayName,
		Description:   row.Description.String,
		Options:       options,
		IsMultiSelect: sqlerr.NullBoolToBool(row.IsMultiSelect),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

// EnumDefinitionToCreateParams converts a domain EnumDefinition to dbgen create params.
func EnumDefinitionToCreateParams(ed *modeldesign.EnumDefinition) (dbgen.CreateEnumDefinitionParams, error) {
	optionsJSON, err := json.Marshal(ed.Options)
	if err != nil {
		return dbgen.CreateEnumDefinitionParams{}, fmt.Errorf("EnumDefinitionToCreateParams: marshal options: %w", err)
	}

	return dbgen.CreateEnumDefinitionParams{
		ID:            ed.ID,
		OrgName:       ed.OrgName,
		ProjectSlug:   ed.ProjectSlug,
		Name:          ed.Name,
		DisplayName:   ed.DisplayName,
		Description:   sql.NullString{String: ed.Description, Valid: ed.Description != ""},
		Options:       json.RawMessage(optionsJSON),
		IsMultiSelect: sqlerr.BoolToNullBool(ed.IsMultiSelect),
	}, nil
}

// FieldEnumAssociationToDomain converts a dbgen.ModelFieldEnumAssociation row to a domain entity.
func FieldEnumAssociationToDomain(row dbgen.ModelFieldEnumAssociation) *modeldesign.FieldEnumAssociation {
	var createdAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}

	var updatedAt time.Time
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	return &modeldesign.FieldEnumAssociation{
		ModelID:   row.ModelID,
		FieldName: row.FieldName,
		ProjectScope: project.ProjectScope{
			OrgName:     row.OrgName,
			ProjectSlug: row.ProjectSlug,
		},
		EnumName:     row.EnumName,
		DatabaseName: row.DatabaseName,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

// FieldEnumAssociationToCreateParams converts a domain FieldEnumAssociation to dbgen create params.
func FieldEnumAssociationToCreateParams(
	assoc *modeldesign.FieldEnumAssociation,
) dbgen.CreateFieldEnumAssociationParams {
	return dbgen.CreateFieldEnumAssociationParams{
		ModelID:      assoc.ModelID,
		FieldName:    assoc.FieldName,
		OrgName:      assoc.OrgName,
		ProjectSlug:  assoc.ProjectSlug,
		EnumName:     assoc.EnumName,
		DatabaseName: assoc.DatabaseName,
	}
}

// SqlEnumRepository is the sqlc-based implementation of modeldesign.EnumRepository.
type SqlEnumRepository struct {
	q dbgen.Querier
}

// NewSqlEnumRepository creates a new SqlEnumRepository backed by the given sqlc Querier.
func NewSqlEnumRepository(q dbgen.Querier) modeldesign.EnumRepository {
	return &SqlEnumRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// Create creates a new enum definition.
// It validates the enum and checks for name uniqueness within the project before persisting.
func (r *SqlEnumRepository) Create(enum *modeldesign.EnumDefinition) error {
	ctx := context.Background()

	if err := enum.Validate(); err != nil {
		return bizerrors.Wrapf(err, "enum validation failed")
	}

	exists, err := r.ExistsByName(enum.OrgName, enum.ProjectSlug, enum.Name)
	if err != nil {
		return bizerrors.Wrapf(err, "failed to check enum existence")
	}
	if exists {
		return shared.NewRepositoryError(
			shared.ErrTypeDuplicatedKey,
			fmt.Sprintf("enum name already exists in project: %s", enum.Name),
		)
	}

	params, err := EnumDefinitionToCreateParams(enum)
	if err != nil {
		return bizerrors.Wrapf(err, "failed to convert enum to create params")
	}

	return r.q.CreateEnumDefinition(ctx, params)
}

// Update updates the mutable fields of an existing enum definition identified by org, project, and name.
func (r *SqlEnumRepository) Update(enum *modeldesign.EnumDefinition) error {
	ctx := context.Background()

	if err := enum.Validate(); err != nil {
		return bizerrors.Wrapf(err, "enum validation failed")
	}

	optionsJSON, err := json.Marshal(enum.Options)
	if err != nil {
		return bizerrors.Wrapf(err, "failed to marshal enum options")
	}

	return r.q.UpdateEnum(ctx, dbgen.UpdateEnumParams{
		DisplayName:   enum.DisplayName,
		Description:   sql.NullString{String: enum.Description, Valid: enum.Description != ""},
		Options:       json.RawMessage(optionsJSON),
		IsMultiSelect: sqlerr.BoolToNullBool(enum.IsMultiSelect),
		OrgName:       enum.OrgName,
		ProjectSlug:   enum.ProjectSlug,
		Name:          enum.Name,
	})
}

// Delete removes an enum definition identified by org, project, and name.
// It returns an error if the enum is still referenced by any model fields.
func (r *SqlEnumRepository) Delete(orgName, projectSlug, name string) error {
	isReferenced, fieldNames, err := r.IsReferencedByFields(orgName, projectSlug, name)
	if err != nil {
		return bizerrors.Wrapf(err, "failed to check enum references")
	}
	if isReferenced {
		return bizerrors.Errorf("enum is referenced by fields: %v", fieldNames)
	}

	return r.q.DeleteEnum(context.Background(), dbgen.DeleteEnumParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Name:        name,
	})
}

// FindByName retrieves an enum definition by org, project, and name.
// Returns a NOT_FOUND repository error if the enum does not exist.
func (r *SqlEnumRepository) FindByName(orgName, projectSlug, name string) (*modeldesign.EnumDefinition, error) {
	ctx := context.Background()

	row, err := r.q.GetEnumByName(ctx, dbgen.GetEnumByNameParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Name:        name,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("enum not found: " + name)
		}
		return nil, err
	}

	return EnumDefinitionToDomain(row)
}

// FindByID retrieves an enum definition by its unique ID.
// Returns nil, shared.NewNotFoundError if the enum does not exist.
func (r *SqlEnumRepository) FindByID(id string) (*modeldesign.EnumDefinition, error) {
	ctx := context.Background()

	row, err := r.q.GetEnumByID(ctx, id)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("enum not found by id: " + id)
		}
		return nil, err
	}

	return EnumDefinitionToDomain(row)
}

// List returns all enum definitions for the given org and project, ordered by name.
func (r *SqlEnumRepository) List(orgName, projectSlug string) ([]*modeldesign.EnumDefinition, error) {
	ctx := context.Background()

	rows, err := r.q.ListEnums(ctx, dbgen.ListEnumsParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	enums := make([]*modeldesign.EnumDefinition, 0, len(rows))
	for _, row := range rows {
		ed, err := EnumDefinitionToDomain(row)
		if err != nil {
			return nil, bizerrors.Wrapf(err, "failed to convert enum row")
		}
		enums = append(enums, ed)
	}

	return enums, nil
}

// IsReferencedByFields checks whether the enum is referenced by any field definitions in the project.
// Returns true and the list of referencing field identifiers (model.field) if referenced.
func (r *SqlEnumRepository) IsReferencedByFields(orgName, projectSlug, name string) (bool, []string, error) {
	ctx := context.Background()

	rows, err := r.q.GetEnumReferencesByName(ctx, dbgen.GetEnumReferencesByNameParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		EnumName:    name,
	})
	if err != nil {
		return false, nil, bizerrors.Wrapf(err, "failed to check field associations")
	}

	if len(rows) == 0 {
		return false, nil, nil
	}

	fieldNames := make([]string, 0, len(rows))
	for _, row := range rows {
		fieldNames = append(fieldNames, fmt.Sprintf("%s.%s", row.ModelName, row.FieldName))
	}

	return true, fieldNames, nil
}

// ExistsByName checks whether an enum with the given name exists in the specified org and project.
func (r *SqlEnumRepository) ExistsByName(orgName, projectSlug, name string) (bool, error) {
	ctx := context.Background()

	count, err := r.q.ExistsEnumByName(ctx, dbgen.ExistsEnumByNameParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Name:        name,
	})
	if err != nil {
		return false, bizerrors.Wrapf(err, "failed to check enum existence")
	}

	return count > 0, nil
}

// SqlFieldEnumAssociationRepository is the sqlc-based implementation of modeldesign.FieldEnumAssociationRepository.
type SqlFieldEnumAssociationRepository struct {
	q dbgen.Querier
}

// NewSqlFieldEnumAssociationRepository creates a new SqlFieldEnumAssociationRepository
// backed by the given sqlc Querier.
func NewSqlFieldEnumAssociationRepository(q dbgen.Querier) modeldesign.FieldEnumAssociationRepository {
	return &SqlFieldEnumAssociationRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// Create creates a new field-enum association after validating the input.
func (r *SqlFieldEnumAssociationRepository) Create(
	ctx context.Context,
	association *modeldesign.FieldEnumAssociation,
) error {
	if err := association.Validate(); err != nil {
		return bizerrors.Wrapf(err, "invalid field enum association")
	}

	return r.q.CreateFieldEnumAssociation(ctx, FieldEnumAssociationToCreateParams(association))
}

// FindByField retrieves the enum association for a specific model field.
// Returns ErrRecordNotFound if no association exists for the given model and field.
func (r *SqlFieldEnumAssociationRepository) FindByField(
	ctx context.Context,
	modelID, fieldName string,
) (*modeldesign.FieldEnumAssociation, error) {
	row, err := r.q.GetFieldEnumAssociationByField(ctx, dbgen.GetFieldEnumAssociationByFieldParams{
		ModelID:   modelID,
		FieldName: fieldName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("field enum association not found: " + modelID + "." + fieldName)
		}
		return nil, err
	}

	return FieldEnumAssociationToDomain(row), nil
}

// FindByEnumName retrieves all field associations for a given enum within an org and project.
func (r *SqlFieldEnumAssociationRepository) FindByEnumName(
	ctx context.Context,
	orgName, projectSlug, enumName string,
) ([]*modeldesign.FieldEnumAssociation, error) {
	rows, err := r.q.GetFieldEnumAssociationsByEnumName(ctx, dbgen.GetFieldEnumAssociationsByEnumNameParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		EnumName:    enumName,
	})
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to find associations by enum_name")
	}

	associations := make([]*modeldesign.FieldEnumAssociation, len(rows))
	for i, row := range rows {
		associations[i] = FieldEnumAssociationToDomain(row)
	}

	return associations, nil
}

// FindByModelID retrieves all field-enum associations for a given model.
func (r *SqlFieldEnumAssociationRepository) FindByModelID(
	ctx context.Context,
	modelID string,
) ([]*modeldesign.FieldEnumAssociation, error) {
	rows, err := r.q.GetFieldEnumAssociationsByModelID(ctx, modelID)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to find associations by model_id")
	}

	associations := make([]*modeldesign.FieldEnumAssociation, len(rows))
	for i, row := range rows {
		associations[i] = FieldEnumAssociationToDomain(row)
	}

	return associations, nil
}

// Delete removes a field-enum association identified by model ID and field name.
func (r *SqlFieldEnumAssociationRepository) Delete(ctx context.Context, modelID, fieldName string) error {
	return r.q.DeleteFieldEnumAssociation(ctx, dbgen.DeleteFieldEnumAssociationParams{
		ModelID:   modelID,
		FieldName: fieldName,
	})
}

// DeleteByModelID removes all field-enum associations for a given model.
func (r *SqlFieldEnumAssociationRepository) DeleteByModelID(ctx context.Context, modelID string) error {
	return r.q.DeleteFieldEnumAssociationsByModelID(ctx, modelID)
}
