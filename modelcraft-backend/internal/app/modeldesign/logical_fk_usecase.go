package modeldesign

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"time"
)

// LogicalFKAppService provides application-layer use cases for logical foreign keys.
type LogicalFKAppService struct {
	fkRepo        modeldesign.LogicalForeignKeyRepository
	fkRepoFactory func(q dbgen.Querier) modeldesign.LogicalForeignKeyRepository
	modelRepo     modeldesign.ModelRepository
	txManager     repository.TxManager
}

// NewLogicalFKAppService creates a new LogicalFKAppService.
func NewLogicalFKAppService(
	fkRepo modeldesign.LogicalForeignKeyRepository,
	modelRepo modeldesign.ModelRepository,
	txManager repository.TxManager,
) *LogicalFKAppService {
	return &LogicalFKAppService{
		fkRepo:        fkRepo,
		fkRepoFactory: repository.NewSqlLogicalForeignKeyRepository,
		modelRepo:     modelRepo,
		txManager:     txManager,
	}
}

// CreateLogicalForeignKey validates and creates FK rows.
// By default it creates a bidirectional pair (normal + reverse).
// When cmd.CreateMode is UNIDIRECTIONAL, only the normal row is created.
func (s *LogicalFKAppService) CreateLogicalForeignKey(
	ctx context.Context,
	cmd CreateLogicalForeignKeyCommand,
) (*modeldesign.LogicalForeignKey, error) {
	if len(cmd.SourceFields) == 0 {
		return nil, bizerrors.NewError(bizerrors.FKColumnsNotFound, "source_fields cannot be empty")
	}
	if len(cmd.SourceFields) != len(cmd.TargetFields) {
		return nil, bizerrors.NewError(bizerrors.FKFieldCountMismatch, "")
	}

	sourceModel, err := s.modelRepo.GetByID(ctx, cmd.ModelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil {
		return nil, err
	}
	if err := validateFieldsExistOnModel(sourceModel, cmd.SourceFields); err != nil {
		return nil, bizerrors.NewError(bizerrors.FKColumnsNotFound, err.Error())
	}

	refModel, err := s.modelRepo.GetByID(ctx, cmd.RefModelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil {
		return nil, err
	}
	if err := validateFieldsExistOnModel(refModel, cmd.TargetFields); err != nil {
		return nil, bizerrors.NewError(bizerrors.FKColumnsNotFound, err.Error())
	}

	normalID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, fmt.Errorf("CreateLogicalForeignKey: generate normalID: %w", err)
	}
	pairID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, fmt.Errorf("CreateLogicalForeignKey: generate pairID: %w", err)
	}

	now := time.Now()
	normalRow := &modeldesign.LogicalForeignKey{
		ID:           normalID,
		PairID:       pairID,
		OrgName:      cmd.OrgName,
		Direction:    modeldesign.DirectionNormal,
		ModelID:      cmd.ModelID,
		ModelName:    sourceModel.ModelName,
		RefModelID:   cmd.RefModelID,
		RefModelName: refModel.ModelName,
		SourceFields: cmd.SourceFields,
		TargetFields: cmd.TargetFields,
		IsDeletable:  true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	createMode := cmd.CreateMode
	if createMode == "" {
		createMode = modeldesign.FKCreateModeBidirectional
	}

	if err := s.txManager.WithTx(ctx, func(txCtx context.Context, q dbgen.Querier) error {
		txFKRepo := s.fkRepoFactory(q)
		if err := txFKRepo.Save(txCtx, normalRow); err != nil {
			return fmt.Errorf("save normal FK row: %w", err)
		}

		if createMode == modeldesign.FKCreateModeBidirectional {
			reverseID, idErr := bizutils.GenerateUUIDV7()
			if idErr != nil {
				return fmt.Errorf("CreateLogicalForeignKey: generate reverseID: %w", idErr)
			}
			reverseRow := &modeldesign.LogicalForeignKey{
				ID:           reverseID,
				PairID:       pairID,
				OrgName:      cmd.OrgName,
				Direction:    modeldesign.DirectionReverse,
				ModelID:      cmd.RefModelID,
				ModelName:    refModel.ModelName,
				RefModelID:   cmd.ModelID,
				RefModelName: sourceModel.ModelName,
				SourceFields: cmd.TargetFields,
				TargetFields: cmd.SourceFields,
				IsDeletable:  true,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := txFKRepo.Save(txCtx, reverseRow); err != nil {
				return fmt.Errorf("save reverse FK row: %w", err)
			}
		}

		if err := txFKRepo.BindBelongsToFields(
			txCtx,
			cmd.OrgName,
			cmd.ModelID,
			normalID,
			cmd.SourceFields,
		); err != nil {
			return fmt.Errorf("bind source fields to belongs_to_fk_id: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return normalRow, nil
}

// CreateSystemEndUserRefFK creates owner(END_USER_REF) -> users.id unidirectional FK and marks it undeletable.
func (s *LogicalFKAppService) CreateSystemEndUserRefFK(
	ctx context.Context,
	orgName string,
	model *modeldesign.DataModel,
) error {
	owner := model.GetOwnerField()
	if owner == nil {
		return nil
	}
	if owner.BelongsToFKID != nil && *owner.BelongsToFKID != "" {
		return nil
	}

	normalID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return fmt.Errorf("CreateSystemEndUserRefFK: generate normalID: %w", err)
	}
	pairID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return fmt.Errorf("CreateSystemEndUserRefFK: generate pairID: %w", err)
	}

	row := &modeldesign.LogicalForeignKey{
		ID:              normalID,
		PairID:          pairID,
		OrgName:         orgName,
		Direction:       modeldesign.DirectionNormal,
		ModelID:         model.ID,
		ModelName:       model.ModelName,
		RefModelID:      "", // END_USER_REF points to private users table, not metadata model table
		RefModelName:    "users",
		RefDatabaseName: fmt.Sprintf("mc_private_%s", model.ProjectSlug),
		RefTableName:    "users",
		SourceFields:    []string{owner.Name},
		TargetFields:    []string{"id"},
		IsDeletable:     false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.fkRepo.Save(ctx, row); err != nil {
		return fmt.Errorf("CreateSystemEndUserRefFK: save FK row: %w", err)
	}
	if err := s.fkRepo.BindBelongsToFields(ctx, orgName, model.ID, normalID, []string{owner.Name}); err != nil {
		return fmt.Errorf("CreateSystemEndUserRefFK: bind owner belongs_to_fk_id: %w", err)
	}
	owner.BelongsToFKID = &normalID
	return nil
}

// DeleteLogicalForeignKey deletes a FK pair by pair_id.
func (s *LogicalFKAppService) DeleteLogicalForeignKey(
	ctx context.Context,
	cmd DeleteLogicalForeignKeyCommand,
) error {
	rows, err := s.fkRepo.FindByPairID(ctx, cmd.OrgName, cmd.PairID)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return bizerrors.NewError(bizerrors.FKNotFound, cmd.PairID)
	}

	for _, row := range rows {
		if !row.IsDeletable {
			return bizerrors.NewError(bizerrors.FKNotDeletable, row.ID)
		}
		relateFields, err := s.fkRepo.FindByRelateField(ctx, cmd.OrgName, row.ID)
		if err != nil {
			return fmt.Errorf("DeleteLogicalForeignKey: check relate fields for %s: %w", row.ID, err)
		}
		if len(relateFields) > 0 {
			return bizerrors.NewError(bizerrors.FKPairHasRelateFields, row.ID)
		}
	}

	return s.fkRepo.DeleteByPairID(ctx, cmd.OrgName, cmd.PairID)
}

// ListLogicalForeignKeys returns all FK rows for a given model.
func (s *LogicalFKAppService) ListLogicalForeignKeys(
	ctx context.Context,
	cmd ListLogicalForeignKeysCommand,
) ([]*modeldesign.LogicalForeignKey, error) {
	return s.fkRepo.FindByModel(ctx, cmd.OrgName, cmd.ModelID)
}

// validateFieldsExistOnModel checks that all fieldNames exist in the model's fields.
func validateFieldsExistOnModel(model *modeldesign.DataModel, fieldNames []string) error {
	fieldSet := make(map[string]bool, len(model.Fields))
	for _, f := range model.Fields {
		fieldSet[f.Name] = true
	}
	for _, name := range fieldNames {
		if !fieldSet[name] {
			return fmt.Errorf("field '%s' not found on model '%s'", name, model.ModelName)
		}
	}
	return nil
}
