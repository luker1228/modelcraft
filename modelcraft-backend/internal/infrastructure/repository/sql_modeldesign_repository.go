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

// ModelToDomain converts a dbgen.Model row to a domain DataModel (without fields).
func ModelToDomain(row dbgen.Model) *modeldesign.DataModel {
	var createdAt, updatedAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	return &modeldesign.DataModel{
		ModelMeta: modeldesign.ModelMeta{
			ID: row.ID,
			ModelLocator: modeldesign.ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     row.OrgName,
					ProjectSlug: row.ProjectSlug,
				},
				ModelName:    row.Name,
				DatabaseName: row.DatabaseName,
			},
			Title:            row.Title,
			Description:      row.Description.String,
			StorageType:      row.StorageType,
			DisplayField:     sqlerr.NullStrToPtr(row.DisplayField),
			Version:          row.Version.Int64,
			Status:           row.Status.String,
			GroupID:          sqlerr.NullStrToPtr(row.GroupID),
			DeploymentStatus: modeldesign.DeploymentStatus(row.DeploymentStatus.String),
			LastSyncAt:       sqlerr.NullTimeToPtr(row.LastSyncAt),
			SyncError:        row.SyncError.String,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		},
	}
}

// ModelToCreateParams converts a domain DataModel to dbgen.CreateModelParams.
func ModelToCreateParams(m *modeldesign.DataModel, orgName string) dbgen.CreateModelParams {
	return dbgen.CreateModelParams{
		ID:               m.ID,
		OrgName:          orgName,
		ProjectSlug:      m.ProjectSlug,
		Name:             m.ModelName,
		Title:            m.Title,
		Description:      sql.NullString{String: m.Description, Valid: m.Description != ""},
		StorageType:      m.StorageType,
		DatabaseName:     m.DatabaseName,
		DisplayField:     sqlerr.PtrToNullStr(m.DisplayField),
		Version:          sql.NullInt64{Int64: m.Version, Valid: m.Version != 0},
		Status:           sql.NullString{String: m.Status, Valid: m.Status != ""},
		GroupID:          sqlerr.PtrToNullStr(m.GroupID),
		DeploymentStatus: sql.NullString{String: string(m.DeploymentStatus), Valid: string(m.DeploymentStatus) != ""},
		LastSyncAt:       sqlerr.PtrToNullTime(m.LastSyncAt),
		SyncError:        sql.NullString{String: m.SyncError, Valid: m.SyncError != ""},
	}
}

// ModelToUpdateParams converts a domain DataModel to dbgen.UpdateModelParams.
func ModelToUpdateParams(m *modeldesign.DataModel) dbgen.UpdateModelParams {
	return dbgen.UpdateModelParams{
		Title:            m.Title,
		Description:      sql.NullString{String: m.Description, Valid: m.Description != ""},
		DisplayField:     sqlerr.PtrToNullStr(m.DisplayField),
		Status:           sql.NullString{String: m.Status, Valid: m.Status != ""},
		GroupID:          sqlerr.PtrToNullStr(m.GroupID),
		DeploymentStatus: sql.NullString{String: string(m.DeploymentStatus), Valid: string(m.DeploymentStatus) != ""},
		Version:          sql.NullInt64{Int64: m.Version, Valid: m.Version != 0},
		ID:               m.ID,
	}
}

// ModelToUpdateWithVersionParams converts a domain DataModel to dbgen.UpdateModelWithVersionParams.
func ModelToUpdateWithVersionParams(
	m *modeldesign.DataModel, originalVersion int64,
) dbgen.UpdateModelWithVersionParams {
	return dbgen.UpdateModelWithVersionParams{
		Title:            m.Title,
		Description:      sql.NullString{String: m.Description, Valid: m.Description != ""},
		DisplayField:     sqlerr.PtrToNullStr(m.DisplayField),
		Status:           sql.NullString{String: m.Status, Valid: m.Status != ""},
		GroupID:          sqlerr.PtrToNullStr(m.GroupID),
		DeploymentStatus: sql.NullString{String: string(m.DeploymentStatus), Valid: string(m.DeploymentStatus) != ""},
		ID:               m.ID,
		Version:          sql.NullInt64{Int64: originalVersion, Valid: true},
	}
}

// FieldDefinitionToDomain converts a dbgen.FieldDefinition row to a domain FieldDefinition.
// Returns an error if JSON fields cannot be unmarshalled or if the format is unknown.
func FieldDefinitionToDomain(row dbgen.FieldDefinition) (*modeldesign.FieldDefinition, error) {
	fieldType, err := modeldesign.NewFieldFormat(modeldesign.FormatType(row.Format))
	if err != nil {
		return nil, fmt.Errorf("FieldDefinitionToDomain: %w", err)
	}

	var validation *modeldesign.ValidationConfig
	if row.Validation != nil && len(*row.Validation) > 0 {
		validation = &modeldesign.ValidationConfig{}
		if err := json.Unmarshal(*row.Validation, validation); err != nil {
			return nil, fmt.Errorf("FieldDefinitionToDomain: unmarshal validation: %w", err)
		}
	}

	var metadata map[string]any
	if row.Metadata != nil && len(*row.Metadata) > 0 {
		if err := json.Unmarshal(*row.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("FieldDefinitionToDomain: unmarshal metadata: %w", err)
		}
	}
	var createdAt, updatedAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	return &modeldesign.FieldDefinition{
		ModelID: row.ModelID,
		ModelLocator: &modeldesign.ModelLocator{
			ProjectScope: project.ProjectScope{
				OrgName:     row.OrgName,
				ProjectSlug: row.ProjectSlug,
			},
			ModelName:    row.ModelName,
			DatabaseName: row.DatabaseName,
		},
		Name:          row.Name,
		EnumName:      row.EnumName.String,
		Title:         row.Title,
		Description:   row.Description.String,
		Type:          fieldType,
		NonNull:       sqlerr.NullBoolToBool(row.NonNull),
		Required:      sqlerr.NullBoolToBool(row.Required),
		IsUnique:      sqlerr.NullBoolToBool(row.IsUnique),
		IsPrimary:     sqlerr.NullBoolToBool(row.IsPrimary),
		IsDeprecated:  sqlerr.NullBoolToBool(row.IsDeprecated),
		Status:        modeldesign.StatusType(row.Status),
		Validation:    validation,
		DisplayOrder:  row.DisplayOrder,
		Metadata:      metadata,
		BelongsToFKID: sqlerr.NullStrToPtr(row.BelongsToFkID),
		RelateFKID:    sqlerr.NullStrToPtr(row.RelateFkID),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

// FieldDefinitionToCreateParams converts a domain FieldDefinition to dbgen.CreateFieldDefinitionParams.
// Returns an error if JSON marshalling of validation/metadata fails.
func FieldDefinitionToCreateParams(
	fd *modeldesign.FieldDefinition, orgName string,
) (dbgen.CreateFieldDefinitionParams, error) {
	validationJSON, err := marshalJSON(fd.Validation)
	if err != nil {
		return dbgen.CreateFieldDefinitionParams{},
			fmt.Errorf("FieldDefinitionToCreateParams: marshal validation: %w", err)
	}

	metadataJSON, err := marshalJSON(fd.Metadata)
	if err != nil {
		return dbgen.CreateFieldDefinitionParams{},
			fmt.Errorf("FieldDefinitionToCreateParams: marshal metadata: %w", err)
	}

	return dbgen.CreateFieldDefinitionParams{
		ModelID:       fd.ModelID,
		OrgName:       orgName,
		ProjectSlug:   fd.ModelLocator.ProjectSlug,
		ModelName:     fd.ModelLocator.ModelName,
		DatabaseName:  fd.ModelLocator.DatabaseName,
		Name:          fd.Name,
		EnumName:      sql.NullString{String: fd.EnumName, Valid: fd.EnumName != ""},
		Title:         fd.Title,
		Description:   sql.NullString{String: fd.Description, Valid: fd.Description != ""},
		Format:        string(fd.Type.Format),
		NonNull:       sqlerr.BoolToNullBool(fd.NonNull),
		Required:      sqlerr.BoolToNullBool(fd.Required),
		IsUnique:      sqlerr.BoolToNullBool(fd.IsUnique),
		IsPrimary:     sqlerr.BoolToNullBool(fd.IsPrimary),
		IsDeprecated:  sqlerr.BoolToNullBool(fd.IsDeprecated),
		Status:        string(fd.Status),
		Validation:    ptrJSON(validationJSON),
		DisplayOrder:  fd.DisplayOrder,
		Metadata:      ptrJSON(metadataJSON),
		RelateFkID:    sqlerr.PtrToNullStr(fd.RelateFKID),
		BelongsToFkID: sqlerr.PtrToNullStr(fd.BelongsToFKID),
	}, nil
}

// FieldDefinitionToUpdateParams converts a domain FieldDefinition to dbgen.UpdateFieldParams.
func FieldDefinitionToUpdateParams(fd *modeldesign.FieldDefinition) (dbgen.UpdateFieldParams, error) {
	validationJSON, err := marshalJSON(fd.Validation)
	if err != nil {
		return dbgen.UpdateFieldParams{}, fmt.Errorf("FieldDefinitionToUpdateParams: marshal validation: %w", err)
	}

	metadataJSON, err := marshalJSON(fd.Metadata)
	if err != nil {
		return dbgen.UpdateFieldParams{}, fmt.Errorf("FieldDefinitionToUpdateParams: marshal metadata: %w", err)
	}

	return dbgen.UpdateFieldParams{
		Title:        fd.Title,
		Description:  sql.NullString{String: fd.Description, Valid: fd.Description != ""},
		NonNull:      sqlerr.BoolToNullBool(fd.NonNull),
		Required:     sqlerr.BoolToNullBool(fd.Required),
		IsUnique:     sqlerr.BoolToNullBool(fd.IsUnique),
		IsPrimary:    sqlerr.BoolToNullBool(fd.IsPrimary),
		IsDeprecated: sqlerr.BoolToNullBool(fd.IsDeprecated),
		Status:       string(fd.Status),
		Validation:   ptrJSON(validationJSON),
		DisplayOrder: fd.DisplayOrder,
		Metadata:     ptrJSON(metadataJSON),
		ModelID:      fd.ModelID,
		Name:         fd.Name,
	}, nil
}

// marshalJSON returns the JSON encoding of v, or nil if v is nil or marshals to JSON null.
func marshalJSON(v any) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// json.Marshal on nil maps/slices returns "null"; treat as no value
	if string(b) == "null" {
		return nil, nil
	}
	return json.RawMessage(b), nil
}

// ptrJSON returns a pointer to a json.RawMessage, or nil if the value is nil.
func ptrJSON(v json.RawMessage) *json.RawMessage {
	if v == nil {
		return nil
	}
	return &v
}

// SqlModelDesignRepository is the sqlc-based implementation of modeldesign.ModelRepository.
type SqlModelDesignRepository struct {
	q dbgen.Querier
}

// NewSqlModelDesignRepository creates a SqlModelDesignRepository.
func NewSqlModelDesignRepository(q dbgen.Querier) modeldesign.ModelRepository {
	return &SqlModelDesignRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// Save creates a new model and its fields.
func (r *SqlModelDesignRepository) Save(ctx context.Context, orgName string, model *modeldesign.DataModel) error {
	if err := r.q.CreateModel(ctx, ModelToCreateParams(model, orgName)); err != nil {
		return err
	}

	return r.AddFields(ctx, orgName, model.Fields)
}

// GetByID retrieves a model by ID, optionally loading its fields.
// Returns nil, shared.NewNotFoundError if the model does not exist.
// GetByID retrieves a model by its ID.
// NOTE: Fields (FieldDefinition list) are NOT loaded by default.
// If you need field data (e.g., for validation or FK checks), you must explicitly pass the WithFields option:
//
//	repo.GetByID(ctx, id, modeldesign.NewModelQueryOptions().WithFields())
func (r *SqlModelDesignRepository) GetByID(
	ctx context.Context, id string, options ...*modeldesign.ModelQueryOptions,
) (*modeldesign.DataModel, error) {
	row, err := r.q.GetModelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if row.ID == "" {
		return nil, shared.NewNotFoundError("model not found by id: " + id)
	}

	m := ModelToDomain(row)

	if len(options) > 0 {
		opt := modeldesign.ApplyOptions(options)
		if opt.GetFields {
			fields, err := r.getFieldsForModel(ctx, m.ID)
			if err != nil {
				return m, err
			}
			m.Fields = fields
		}
	}

	return m, nil
}

// GetByName retrieves a model by org name, database name, model name, and project slug.
// Returns nil, shared.NewNotFoundError if the model does not exist.
func (r *SqlModelDesignRepository) GetByName(
	ctx context.Context,
	orgName, databaseName, name, projectSlug string,
	opts ...*modeldesign.ModelQueryOptions,
) (*modeldesign.DataModel, error) {
	row, err := r.q.GetModelByName(ctx, dbgen.GetModelByNameParams{
		OrgName:      orgName,
		DatabaseName: databaseName,
		Name:         name,
		ProjectSlug:  projectSlug,
	})
	if err != nil {
		return nil, err
	}

	if row.ID == "" {
		return nil, shared.NewNotFoundError("model not found: " + name)
	}

	m := ModelToDomain(row)

	if len(opts) > 0 {
		opt := modeldesign.ApplyOptions(opts)
		if opt.GetFields {
			fields, err := r.getFieldsForModel(ctx, m.ID)
			if err != nil {
				return m, err
			}
			m.Fields = fields
		}
	}

	return m, nil
}

// Update updates an existing model's mutable fields.
func (r *SqlModelDesignRepository) Update(ctx context.Context, model *modeldesign.DataModel) error {
	result, err := r.q.UpdateModel(ctx, ModelToUpdateParams(model))
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "Model not found or not updated")
	}

	return nil
}

// Delete removes a model by ID.
func (r *SqlModelDesignRepository) Delete(ctx context.Context, id string) error {
	return r.q.DeleteModel(ctx, id)
}

// Query returns a paginated, filtered list of models plus total count.
func (r *SqlModelDesignRepository) Query(
	ctx context.Context, queryObj modeldesign.ModelQuery,
) ([]modeldesign.DataModel, int, error) {
	nameFilter, nameArg := nullableTrickArgs(queryObj.Name)
	titleFilter, titleArg := nullableTrickArgs(queryObj.Title)

	statusArg := sql.NullString{}
	if queryObj.Status != "" {
		statusArg = sql.NullString{String: queryObj.Status, Valid: true}
	}

	storageTypeArg := queryObj.StorageType

	var limit, offset int32
	if queryObj.PageSize > 0 && queryObj.Page > 0 {
		limit = int32(queryObj.PageSize)
		offset = int32((queryObj.Page - 1) * queryObj.PageSize)
	} else {
		limit = 1000
		offset = 0
	}

	rows, err := r.q.ListModels(ctx, dbgen.ListModelsParams{
		OrgName:      queryObj.OrgName,
		ProjectSlug:  queryObj.ProjectSlug,
		DatabaseName: queryObj.DatabaseName,
		Column4:      nameFilter,
		CONCAT:       nameArg,
		Column6:      titleFilter,
		CONCAT_2:     titleArg,
		Column8:      nullableStatusFilter(statusArg),
		Status:       statusArg,
		Column10:     nullableStorageTypeFilter(storageTypeArg),
		StorageType:  storageTypeArg,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.q.CountModels(ctx, dbgen.CountModelsParams{
		OrgName:      queryObj.OrgName,
		ProjectSlug:  queryObj.ProjectSlug,
		DatabaseName: queryObj.DatabaseName,
		Column4:      nameFilter,
		CONCAT:       nameArg,
		Column6:      titleFilter,
		CONCAT_2:     titleArg,
		Column8:      nullableStatusFilter(statusArg),
		Status:       statusArg,
		Column10:     nullableStorageTypeFilter(storageTypeArg),
		StorageType:  storageTypeArg,
	})
	if err != nil {
		return nil, 0, err
	}

	models := make([]modeldesign.DataModel, len(rows))
	for i, row := range rows {
		models[i] = *ModelToDomain(row)
	}

	return models, int(total), nil
}

// ListDatabaseCatalog returns distinct database names for a project with pagination.
func (r *SqlModelDesignRepository) ListDatabaseCatalog(
	ctx context.Context,
	orgName, projectSlug, search string,
	page, pageSize int,
) ([]string, int, error) {
	var limit, offset int32
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	limit = int32(pageSize)
	offset = int32((page - 1) * pageSize)

	searchFilter, searchArg := nullableTrickArgs(search)
	rows, err := r.q.ListModelDatabases(ctx, dbgen.ListModelDatabasesParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Column3:     searchFilter,
		CONCAT:      searchArg,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.q.CountModelDatabases(ctx, dbgen.CountModelDatabasesParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Column3:     searchFilter,
		CONCAT:      searchArg,
	})
	if err != nil {
		return nil, 0, err
	}

	databases := make([]string, 0, len(rows))
	for _, name := range rows {
		if name == "" {
			continue
		}
		databases = append(databases, name)
	}
	return databases, int(total), nil
}

// GetAll returns all models without pagination.
func (r *SqlModelDesignRepository) GetAll(ctx context.Context) ([]modeldesign.DataModel, error) {
	rows, err := r.q.GetAllModels(ctx)
	if err != nil {
		return nil, err
	}

	models := make([]modeldesign.DataModel, len(rows))
	for i, row := range rows {
		models[i] = *ModelToDomain(row)
	}

	return models, nil
}

// UpdateWithVersion updates a model only if the current version matches, returning rows affected.
func (r *SqlModelDesignRepository) UpdateWithVersion(
	ctx context.Context, model *modeldesign.DataModel, originalVersion int64,
) (int64, error) {
	result, err := r.q.UpdateModelWithVersion(ctx, ModelToUpdateWithVersionParams(model, originalVersion))
	if err != nil {
		return 0, err
	}

	rows, _ := result.RowsAffected()
	return rows, nil
}

// FindByDeploymentStatus returns all models matching any of the given deployment statuses.
func (r *SqlModelDesignRepository) FindByDeploymentStatus(
	ctx context.Context, statuses ...modeldesign.DeploymentStatus,
) ([]modeldesign.DataModel, error) {
	sqlStatuses := make([]sql.NullString, len(statuses))
	for i, s := range statuses {
		sqlStatuses[i] = sql.NullString{String: string(s), Valid: true}
	}

	rows, err := r.q.FindModelsByDeploymentStatus(ctx, sqlStatuses)
	if err != nil {
		return nil, err
	}

	models := make([]modeldesign.DataModel, len(rows))
	for i, row := range rows {
		models[i] = *ModelToDomain(row)
	}

	return models, nil
}

// AddFields bulk-inserts field definitions for a model.
func (r *SqlModelDesignRepository) AddFields(
	ctx context.Context, orgName string, fields []*modeldesign.FieldDefinition,
) error {
	for _, field := range fields {
		params, err := FieldDefinitionToCreateParams(field, orgName)
		if err != nil {
			return fmt.Errorf("AddFields: convert field %q: %w", field.Name, err)
		}
		if err := r.q.CreateFieldDefinition(ctx, params); err != nil {
			return err
		}
	}
	return nil
}

// GetTailFieldDisplayOrder returns the largest display_order value among fields in the model,
// or an empty string if no fields exist.
func (r *SqlModelDesignRepository) GetTailFieldDisplayOrder(
	ctx context.Context, modelID string,
) (string, error) {
	order, err := r.q.GetTailFieldDisplayOrder(ctx, modelID)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return "", nil
		}
		return "", bizerrors.Wrapf(err, "get tail field display order for model %s", modelID)
	}
	return order, nil
}

// AddRelationField creates a field definition and its associated model relation.
func (r *SqlModelDesignRepository) AddRelationField(
	ctx context.Context, orgName string, field *modeldesign.FieldDefinition,
) error {
	params, err := FieldDefinitionToCreateParams(field, orgName)
	if err != nil {
		return fmt.Errorf("AddRelationField: convert field: %w", err)
	}
	return r.q.CreateFieldDefinition(ctx, params)
}

// GetFieldByModelID retrieves a single field definition by model ID and field name.
// Returns ErrRecordNotFound if the field does not exist.
func (r *SqlModelDesignRepository) GetFieldByModelID(
	ctx context.Context, modelID, name string,
) (*modeldesign.FieldDefinition, error) {
	row, err := r.q.GetFieldByModelIDAndName(ctx, dbgen.GetFieldByModelIDAndNameParams{
		ModelID: modelID,
		Name:    name,
	})
	if err != nil {
		return nil, err
	}

	if row.ModelID == "" {
		return nil, shared.NewNotFoundError("field not found: " + modelID + "." + name)
	}

	return FieldDefinitionToDomain(row)
}

// GetFieldsByModelID returns all field definitions for a model, ordered by display_order.
func (r *SqlModelDesignRepository) GetFieldsByModelID(
	ctx context.Context, modelID string,
) ([]*modeldesign.FieldDefinition, error) {
	return r.getFieldsForModel(ctx, modelID)
}

// UpdateField updates a field definition by model ID and name.
func (r *SqlModelDesignRepository) UpdateField(
	ctx context.Context, field *modeldesign.FieldDefinition,
) error {
	params, err := FieldDefinitionToUpdateParams(field)
	if err != nil {
		return fmt.Errorf("UpdateField: convert field: %w", err)
	}

	result, err := r.q.UpdateField(ctx, params)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "Field definition not found or not updated")
	}

	return nil
}

// BulkUpdateFields updates multiple field definitions sequentially.
func (r *SqlModelDesignRepository) BulkUpdateFields(
	ctx context.Context, fields []*modeldesign.FieldDefinition,
) error {
	for _, field := range fields {
		if err := r.UpdateField(ctx, field); err != nil {
			return err
		}
	}
	return nil
}

// DeleteFields deletes named fields from a model. Returns ErrTypeNoRowsAffected if none were deleted.
func (r *SqlModelDesignRepository) DeleteFields(
	ctx context.Context, modelID string, names []string,
) error {
	result, err := r.q.DeleteFieldsByNames(ctx, dbgen.DeleteFieldsByNamesParams{
		ModelID: modelID,
		Names:   names,
	})
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "Field definition not found")
	}

	return nil
}

// BulkDeleteFields deletes fields from multiple models.
func (r *SqlModelDesignRepository) BulkDeleteFields(
	ctx context.Context, requests ...modeldesign.DeleteFieldRequest,
) error {
	for _, req := range requests {
		if len(req.Name) == 0 {
			continue
		}

		result, err := r.q.DeleteFieldsByNames(ctx, dbgen.DeleteFieldsByNamesParams{
			ModelID: req.ModelId,
			Names:   req.Name,
		})
		if err != nil {
			return err
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected,
				fmt.Sprintf("No fields found for model %q with names: %v", req.ModelId, req.Name))
		}
	}
	return nil
}

// UpdateFieldsStatus bulk-updates the status of named fields in a model.
func (r *SqlModelDesignRepository) UpdateFieldsStatus(
	ctx context.Context, requests ...modeldesign.UpdateFieldsStatusRequest,
) error {
	for _, req := range requests {
		if len(req.Name) == 0 {
			continue
		}

		if err := r.q.UpdateFieldsStatus(ctx, dbgen.UpdateFieldsStatusParams{
			Status:  string(req.Status),
			ModelID: req.ModelId,
			Names:   req.Name,
		}); err != nil {
			return err
		}
	}
	return nil
}

// getFieldsForModel fetches and converts all field definitions for a model.
func (r *SqlModelDesignRepository) getFieldsForModel(
	ctx context.Context, modelID string,
) ([]*modeldesign.FieldDefinition, error) {
	rows, err := r.q.GetFieldsByModelID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	fields := make([]*modeldesign.FieldDefinition, len(rows))
	for i, row := range rows {
		fd, err := FieldDefinitionToDomain(row)
		if err != nil {
			return nil, fmt.Errorf("getFieldsForModel: convert field %q: %w", row.Name, err)
		}
		fields[i] = fd
	}
	return fields, nil
}

// --- nullable trick helpers ---

// nullableTrickArgs returns the two args needed for: (? IS NULL OR col LIKE CONCAT('%', ?, '%'))
// When s is empty, both args are nil (so the condition is skipped).
func nullableTrickArgs(s string) (interface{}, interface{}) {
	if s == "" {
		return nil, nil
	}
	return s, s
}

// nullableStatusFilter returns nil or the NullString value for IS NULL OR status = ?
func nullableStatusFilter(ns sql.NullString) interface{} {
	if !ns.Valid {
		return nil
	}
	return ns.String
}

// nullableStorageTypeFilter returns nil or the string value for IS NULL OR storage_type = ?
func nullableStorageTypeFilter(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Compile-time interface satisfaction checks.
var (
	_ modeldesign.ModelRepository = (*SqlModelDesignRepository)(nil)
)
