package repository

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"

	bizerrors "modelcraft/pkg/bizerrors"
)

// SqlModelRuntimeRepository is the sqlc-based implementation of modelruntime.ModelRepository.
type SqlModelRuntimeRepository struct {
	q dbgen.Querier
}

// NewSqlModelRuntimeRepository creates a new SqlModelRuntimeRepository.
func NewSqlModelRuntimeRepository(q dbgen.Querier) modelruntime.ModelRepository {
	return &SqlModelRuntimeRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// GetByID retrieves a runtime model with its fields by model ID.
func (r *SqlModelRuntimeRepository) GetByID(ctx context.Context, id string) (*modelruntime.RuntimeModel, error) {
	row, err := r.q.GetModelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	runtimeModel := DbgenModelToRuntimeModel(row)
	fields, err := r.getFields(ctx, row.ID, row.OrgName, row.ProjectSlug)
	if err != nil {
		return nil, err
	}
	runtimeModel.Fields = fields

	return runtimeModel, nil
}

// GetByName retrieves a runtime model by model locator.
func (r *SqlModelRuntimeRepository) GetByName(
	ctx context.Context, modelLocator *modeldesign.ModelLocator,
) (*modelruntime.RuntimeModel, error) {
	row, err := r.q.GetModelByName(ctx, dbgen.GetModelByNameParams{
		OrgName:      modelLocator.OrgName,
		DatabaseName: modelLocator.DatabaseName,
		Name:         modelLocator.ModelName,
		ProjectSlug:  modelLocator.ProjectSlug,
	})
	if err != nil {
		return nil, err
	}

	runtimeModel := DbgenModelToRuntimeModel(row)
	fields, err := r.getFields(ctx, row.ID, row.OrgName, row.ProjectSlug)
	if err != nil {
		return nil, err
	}
	runtimeModel.Fields = fields

	return runtimeModel, nil
}

// getFields retrieves all fields for a model, enriched with relation and enum info.
func (r *SqlModelRuntimeRepository) getFields(
	ctx context.Context, modelID, orgName, projectSlug string,
) (map[string]*modelruntime.RuntimeField, error) {
	// Fetch field definitions
	fieldRows, err := r.q.GetFieldsByModelID(ctx, modelID)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "getFields: fetch fields")
	}

	// Collect enum names from fields
	enumNames := make([]string, 0)
	for _, f := range fieldRows {
		if f.EnumName.Valid && f.EnumName.String != "" {
			enumNames = append(enumNames, f.EnumName.String)
		}
	}

	// Fetch enum definitions if needed
	enumMap := make(map[string]*modeldesign.EnumDefinition)
	if len(enumNames) > 0 {
		enumRows, enumErr := r.q.GetEnumsByNames(ctx, dbgen.GetEnumsByNamesParams{
			OrgName:     orgName,
			ProjectSlug: projectSlug,
			Names:       enumNames,
		})
		if enumErr != nil {
			return nil, bizerrors.Wrapf(enumErr, "getFields: fetch enums")
		}
		for _, row := range enumRows {
			ed, convErr := EnumDefinitionToDomain(row)
			if convErr != nil {
				return nil, bizerrors.Wrapf(convErr, "getFields: convert enum")
			}
			enumMap[row.Name] = ed
		}
	}

	// Assemble RuntimeField map
	fieldsMap := make(map[string]*modelruntime.RuntimeField, len(fieldRows))
	for _, fieldRow := range fieldRows {
		fd, convErr := FieldDefinitionToDomain(fieldRow)
		if convErr != nil {
			return nil, bizerrors.Wrapf(convErr, "getFields: convert field")
		}

		// Attach enum definition if present
		if fieldRow.EnumName.Valid && fieldRow.EnumName.String != "" {
			if ed, ok := enumMap[fieldRow.EnumName.String]; ok {
				fd.Enum = ed
			}
		}

		fieldsMap[fieldRow.Name] = fd
	}

	return fieldsMap, nil
}

// DbgenModelToRuntimeModel converts a dbgen.Model row to a modelruntime.RuntimeModel.
func DbgenModelToRuntimeModel(row dbgen.Model) *modelruntime.RuntimeModel {
	var displayField *string
	if row.DisplayField.Valid && row.DisplayField.String != "" {
		displayField = &row.DisplayField.String
	}
	return &modelruntime.RuntimeModel{
		ID:           row.ID,
		OrgName:      row.OrgName,
		ProjectSlug:  row.ProjectSlug,
		Name:         row.Name,
		Title:        row.Title,
		Description:  row.Description.String,
		DatabaseName: row.DatabaseName,
		CreatedVia:   modeldesign.ModelCreationSource(row.CreatedVia),
		DisplayField: displayField,
	}
}

// compile-time interface check
var _ modelruntime.ModelRepository = (*SqlModelRuntimeRepository)(nil)
