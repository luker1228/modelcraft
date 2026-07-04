package modeldesign

import (
	"context"
	"fmt"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/ddlfactory"
	"modelcraft/pkg/logfacade"
	"strings"

	entity "modelcraft/internal/domain/modeldesign"
)

// RepairModelUseCase handles model repair operations
type RepairModelUseCase struct {
	modelRepo               entity.ModelRepository
	clusterManager          *repository.ClusterConnectionManager
	deployRepo              entity.DeployRepo
	schemaComparisonService entity.SchemaComparisonService
}

// NewRepairModelUseCase creates a new repair model use case
func NewRepairModelUseCase(
	modelRepo entity.ModelRepository,
	clusterManager *repository.ClusterConnectionManager,
	deployRepo entity.DeployRepo,
	schemaComparisonService entity.SchemaComparisonService,
) *RepairModelUseCase {
	return &RepairModelUseCase{
		modelRepo:               modelRepo,
		clusterManager:          clusterManager,
		deployRepo:              deployRepo,
		schemaComparisonService: schemaComparisonService,
	}
}

// RepairModel repairs a model based on the specified mode
func (s *RepairModelUseCase) RepairModel(
	ctx context.Context,
	projectID, modelID string,
	mode entity.RepairMode,
	deleteExtraFields bool,
) (*entity.RepairResult, error) {
	logger := logfacade.GetLogger(ctx)
	logger.Infof(
		ctx,
		"Starting model repair: modelID=%s, mode=%s, deleteExtraFields=%v",
		modelID, mode, deleteExtraFields,
	)

	// Get model with fields
	opts := entity.NewModelQueryOptions().WithFields()
	model, err := s.modelRepo.GetByID(ctx, modelID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	if model == nil {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}
	if model.IsReadOnly {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ManagedModelReadOnly, model.ModelName)
	}

	// Compare schema to detect issues
	issues, healthBefore, err := s.schemaComparisonService.CompareSchema(ctx, model, projectID, s.clusterManager)
	if err != nil {
		return nil, fmt.Errorf("failed to compare schema: %w", err)
	}

	result := &entity.RepairResult{
		Model:              model,
		ChangesApplied:     false,
		DetectedIssues:     issues,
		ExecutedDDL:        []string{},
		HealthStatusBefore: healthBefore,
		HealthStatusAfter:  healthBefore,
		ExtraFieldsRemoved: []string{},
		FieldsAdded:        []string{},
	}

	// In dry run mode, return without making changes
	if mode == entity.DryRun {
		logger.Infof(ctx, "Dry run mode: returning without making changes")
		return result, nil
	}

	// Execute repair operations
	executedDDL, fieldsAdded, fieldsRemoved, err := s.executeRepair(ctx, model, issues, mode, deleteExtraFields)
	if err != nil {
		return nil, fmt.Errorf("failed to execute repair: %w", err)
	}

	result.ExecutedDDL = executedDDL
	result.FieldsAdded = fieldsAdded
	result.ExtraFieldsRemoved = fieldsRemoved

	// Recalculate health after repair
	_, healthAfter, err := s.schemaComparisonService.CompareSchema(ctx, model, projectID, s.clusterManager)
	if err != nil {
		logger.Warnf(ctx, "Failed to calculate health after repair: %v", err)
		result.HealthStatusAfter = healthBefore // Keep previous health if check fails
	} else {
		result.HealthStatusAfter = healthAfter
	}

	result.ChangesApplied = len(executedDDL) > 0
	logger.Infof(ctx, "Model repair completed: changesApplied=%v, ddlExecuted=%d, fieldsAdded=%d, fieldsRemoved=%d",
		result.ChangesApplied, len(executedDDL), len(fieldsAdded), len(fieldsRemoved))

	return result, nil
}

// executeRepair executes the repair operations based on detected issues
func (s *RepairModelUseCase) executeRepair(
	ctx context.Context,
	model *entity.DataModel,
	issues []entity.SchemaIssue,
	mode entity.RepairMode,
	deleteExtraFields bool,
) ([]string, []string, []string, error) {
	logger := logfacade.GetLogger(ctx)
	var executedDDL []string
	var fieldsAdded []string
	var fieldsRemoved []string

	// Group issues by type
	tableMissing := false
	missingFields := []string{}

	for _, issue := range issues {
		switch issue.Type {
		case entity.TableMissing:
			tableMissing = true
		case entity.FieldMissing:
			missingFields = append(missingFields, issue.FieldName)
		case entity.FieldTypeMismatch, entity.FieldConstraintMismatch:
			// Type and constraint mismatches are not auto-fixed
			logger.Infof(
				ctx, "Skipping fix for issue type %s (field: %s): requires manual intervention",
				issue.Type,
				issue.FieldName,
			)
		case entity.DatabaseMissing, entity.ClusterNotFound:
			// Cannot fix missing database or cluster
			return nil, nil, nil, fmt.Errorf("cannot repair: %s", issue.Description)
		}
	}

	// Handle table missing - create the table
	if tableMissing {
		logger.Infof(ctx, "Table missing, creating table %s", model.ModelName)
		deployErr := s.deployRepo.DeployModelToCreate(ctx, model)
		if deployErr != nil {
			return nil, nil, nil, fmt.Errorf("failed to create table: %w", deployErr)
		}
		generatedDDL, genErr := s.generateCreateTableDDL(ctx, model)
		if genErr != nil {
			return nil, nil, nil, fmt.Errorf("failed to generate DDL: %w", genErr)
		}
		executedDDL = append(executedDDL, generatedDDL)
		// When table is created, all fields are added
		for _, field := range model.Fields {
			if !field.IsRelationField() {
				fieldsAdded = append(fieldsAdded, field.Name)
			}
		}
	}

	// Handle missing fields - add them
	if len(missingFields) > 0 {
		addDDL, added, addErr := s.addMissingFields(ctx, model, missingFields)
		if addErr != nil {
			return nil, nil, nil, addErr
		}
		executedDDL = append(executedDDL, addDDL...)
		fieldsAdded = append(fieldsAdded, added...)
	}

	// Handle extra fields in FULL_SYNC mode
	if mode == entity.FullSync && deleteExtraFields {
		removeDDL, removed, removeErr := s.removeExtraFields(ctx, model)
		if removeErr != nil {
			return nil, nil, nil, removeErr
		}
		executedDDL = append(executedDDL, removeDDL...)
		fieldsRemoved = append(fieldsRemoved, removed...)
	}

	return executedDDL, fieldsAdded, fieldsRemoved, nil
}

func (s *RepairModelUseCase) addMissingFields(
	ctx context.Context,
	model *entity.DataModel,
	missingFields []string,
) ([]string, []string, error) {
	logger := logfacade.GetLogger(ctx)
	logger.Infof(ctx, "Adding missing fields: %v", missingFields)

	fieldsToAdd := make([]*entity.FieldDefinition, 0, len(missingFields))
	for _, fieldName := range missingFields {
		field := model.GetField(fieldName)
		if field != nil && !field.IsRelationField() {
			fieldsToAdd = append(fieldsToAdd, field)
		}
	}

	if len(fieldsToAdd) == 0 {
		return nil, nil, nil
	}

	if err := s.deployRepo.DeployModelToAddFields(ctx, model, fieldsToAdd); err != nil {
		return nil, nil, fmt.Errorf("failed to add fields: %w", err)
	}
	generatedDDL, err := s.generateAddColumnsDDL(ctx, model, fieldsToAdd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate add columns DDL: %w", err)
	}

	addedFields := make([]string, 0, len(fieldsToAdd))
	for _, field := range fieldsToAdd {
		addedFields = append(addedFields, field.Name)
	}

	return []string{generatedDDL}, addedFields, nil
}

func (s *RepairModelUseCase) removeExtraFields(
	ctx context.Context,
	model *entity.DataModel,
) ([]string, []string, error) {
	logger := logfacade.GetLogger(ctx)

	extraFields, err := s.findExtraFields(ctx, model)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find extra fields: %w", err)
	}
	if len(extraFields) == 0 {
		return nil, nil, nil
	}

	logger.Infof(ctx, "Removing extra fields in FULL_SYNC mode: %v", extraFields)
	fieldsToRemove, blockedFields := s.filterSafeToRemoveFields(model, extraFields)
	for _, blocked := range blockedFields {
		logger.Infof(ctx, "Cannot remove field %s: protected or has dependencies", blocked)
	}
	if len(fieldsToRemove) == 0 {
		return nil, nil, nil
	}

	executedDDL := make([]string, 0, len(fieldsToRemove))
	fieldsRemoved := make([]string, 0, len(fieldsToRemove))
	for _, fieldName := range fieldsToRemove {
		sql := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", model.ModelName, fieldName)
		if execErr := s.execDDL(ctx, model, sql); execErr != nil {
			logger.Warnf(ctx, "Failed to remove field %s: %v", fieldName, execErr)
			continue
		}
		executedDDL = append(executedDDL, sql)
		fieldsRemoved = append(fieldsRemoved, fieldName)
	}

	return executedDDL, fieldsRemoved, nil
}

// generateCreateTableDDL generates CREATE TABLE DDL for a model
func (s *RepairModelUseCase) generateCreateTableDDL(ctx context.Context, model *entity.DataModel) (string, error) {
	mysqlBuilder := ddlfactory.NewMySQLDDLBuilder(ctx)
	typeMapper := entity.NewMySQLTypeMapper()

	// Convert all fields to FieldEntity
	fieldEntities := make([]*ddlfactory.FieldEntity, 0, len(model.Fields))
	for _, field := range model.Fields {
		if field.IsRelationField() {
			continue
		}

		mysqlType, err := typeMapper.MapToMySQL(field)
		if err != nil {
			return "", bizerrors.NewError(bizerrors.SystemError,
				fmt.Sprintf("failed to map field type: %v", err))
		}

		fieldEntity, err := s.buildFieldEntity(field, mysqlType)
		if err != nil {
			return "", err
		}
		fieldEntities = append(fieldEntities, fieldEntity)
	}

	tableEntity := &ddlfactory.TableEntity{
		Name:    model.ModelName,
		Fields:  fieldEntities,
		Engine:  "InnoDB",
		Charset: "utf8mb4",
		Comment: model.Title,
	}

	return mysqlBuilder.BuildCreateTable(tableEntity)
}

// generateAddColumnsDDL generates ADD COLUMN DDL for fields
func (s *RepairModelUseCase) generateAddColumnsDDL(
	ctx context.Context,
	model *entity.DataModel,
	fields []*entity.FieldDefinition,
) (string, error) {
	mysqlBuilder := ddlfactory.NewMySQLDDLBuilder(ctx)
	typeMapper := entity.NewMySQLTypeMapper()

	fieldEntities := make([]*ddlfactory.FieldEntity, 0, len(fields))
	for _, field := range fields {
		mysqlType, err := typeMapper.MapToMySQL(field)
		if err != nil {
			return "", bizerrors.NewError(bizerrors.SystemError,
				fmt.Sprintf("failed to map field type: %v", err))
		}

		fieldEntity, err := s.buildFieldEntity(field, mysqlType)
		if err != nil {
			return "", err
		}
		fieldEntities = append(fieldEntities, fieldEntity)
	}

	return mysqlBuilder.BuildAddColumns(model.ModelName, fieldEntities)
}

// buildFieldEntity builds a FieldEntity from a FieldDefinition and MySQL type string
func (s *RepairModelUseCase) buildFieldEntity(
	field *entity.FieldDefinition,
	mysqlType string,
) (*ddlfactory.FieldEntity, error) {
	fieldEntity := &ddlfactory.FieldEntity{
		Name: field.Name,
	}

	// Parse MySQL type string
	if err := parseMySQLTypeToField(fieldEntity, mysqlType); err != nil {
		return nil, err
	}

	fieldEntity.Nullable = !field.NonNull

	if field.IsPrimary {
		fieldEntity.Primary = true
		fieldEntity.Nullable = false
	}

	if field.Description != "" {
		fieldEntity.Comment = field.Description
	}

	return fieldEntity, nil
}

// parseMySQLTypeToField parses a MySQL type string and populates FieldEntity
func parseMySQLTypeToField(field *ddlfactory.FieldEntity, mysqlType string) error {
	switch {
	case len(mysqlType) >= 7 && mysqlType[:7] == "VARCHAR":
		field.Type = ddlfactory.VARCHAR
		if len(mysqlType) > 7 {
			var length int
			_, _ = fmt.Sscanf(mysqlType, "VARCHAR(%d)", &length)
			field.Length = &length
		}
	case len(mysqlType) >= 4 && mysqlType[:4] == "CHAR":
		field.Type = ddlfactory.CHAR
		if len(mysqlType) > 4 {
			var length int
			_, _ = fmt.Sscanf(mysqlType, "CHAR(%d)", &length)
			field.Length = &length
		}
	case mysqlType == "TEXT":
		field.Type = ddlfactory.TEXT
	case mysqlType == "MEDIUMTEXT":
		field.Type = ddlfactory.TEXT
	case mysqlType == "LONGTEXT":
		field.Type = ddlfactory.TEXT
	case mysqlType == "INT":
		field.Type = ddlfactory.INT
	case mysqlType == "DOUBLE":
		field.Type = ddlfactory.DOUBLE
	case len(mysqlType) > 7 && mysqlType[:7] == "DECIMAL":
		field.Type = ddlfactory.DECIMAL
		var precision, scale int
		_, _ = fmt.Sscanf(mysqlType, "DECIMAL(%d,%d)", &precision, &scale)
		field.Precision = &precision
		field.Scale = &scale
	case mysqlType == "TINYINT(1)" || mysqlType == "BOOL":
		field.Type = ddlfactory.BOOL
	case mysqlType == "DATE":
		field.Type = ddlfactory.DATE
	case mysqlType == "DATETIME":
		field.Type = ddlfactory.DATETIME
	case mysqlType == "TIME":
		field.Type = ddlfactory.VARCHAR
		length := 8
		field.Length = &length
	case mysqlType == "JSON":
		field.Type = ddlfactory.JSON
	default:
		return fmt.Errorf("unsupported MySQL type: %s", mysqlType)
	}
	return nil
}

// findExtraFields finds columns that exist in the database but not in the model
func (s *RepairModelUseCase) findExtraFields(
	ctx context.Context,
	model *entity.DataModel,
) ([]string, error) {
	// Build map of expected field names
	expectedFields := make(map[string]bool)
	for _, field := range model.Fields {
		expectedFields[field.Name] = true
	}

	// Get actual table definition
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get orgName from context: %w", err)
	}
	tableDef, err := s.schemaComparisonService.GetTableDefinition(
		ctx,
		orgName,
		model.ProjectSlug,
		model.DatabaseName,
		model.ModelName,
		s.clusterManager,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get table definition: %w", err)
	}

	// Find extra fields
	var extraFields []string
	for _, col := range tableDef.Columns {
		if !expectedFields[col.Name] {
			extraFields = append(extraFields, col.Name)
		}
	}

	return extraFields, nil
}

// filterSafeToRemoveFields filters out protected fields (system fields)
func (s *RepairModelUseCase) filterSafeToRemoveFields(
	model *entity.DataModel,
	fieldNames []string,
) ([]string, []string) {
	systemFields := map[string]bool{
		"id":        true,
		"createdAt": true,
		"updatedAt": true,
	}

	safeToRemove := make([]string, 0, len(fieldNames))
	blocked := make([]string, 0, len(fieldNames))

	for _, name := range fieldNames {
		// Check if it's a system field
		if systemFields[name] {
			blocked = append(blocked, name)
			continue
		}

		safeToRemove = append(safeToRemove, name)
	}

	return safeToRemove, blocked
}

// execDDL executes a DDL statement against the customer database
func (s *RepairModelUseCase) execDDL(
	ctx context.Context,
	model *entity.DataModel,
	ddlStatement string,
) error {
	// Get connection
	// TODO: Get orgName from context
	orgName := "default"
	conn, err := s.clusterManager.GetConnectionWithDatabase(
		ctx,
		orgName,
		model.ProjectSlug,
		model.DatabaseName,
	)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Execute DDL
	_, err = conn.ExecContext(ctx, ddlStatement)
	if err != nil {
		return fmt.Errorf("failed to execute DDL: %s, error: %w", ddlStatement, err)
	}

	return nil
}

// ParseRepairMode parses a string to RepairMode
func ParseRepairMode(mode string) (entity.RepairMode, error) {
	switch strings.ToUpper(mode) {
	case string(entity.DryRun):
		return entity.DryRun, nil
	case string(entity.Additive):
		return entity.Additive, nil
	case string(entity.FullSync):
		return entity.FullSync, nil
	default:
		return "", fmt.Errorf("invalid repair mode: %s", mode)
	}
}
