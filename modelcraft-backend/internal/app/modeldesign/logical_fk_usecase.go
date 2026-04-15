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

// CreateLogicalForeignKey validates and atomically creates both normal and reverse FK rows.
//
// Validation:
//  1. source_fields and target_fields must have the same count
//  2. All source_fields must exist on the model with ModelID
//  3. All target_fields must exist on the model with RefModelID
func (s *LogicalFKAppService) CreateLogicalForeignKey(
	ctx context.Context,
	cmd CreateLogicalForeignKeyCommand,
) (*modeldesign.LogicalForeignKey, error) {
	// 1. Validate field count parity
	if len(cmd.SourceFields) == 0 {
		return nil, bizerrors.NewError(bizerrors.FKColumnsNotFound, "source_fields cannot be empty")
	}
	if len(cmd.SourceFields) != len(cmd.TargetFields) {
		return nil, bizerrors.NewError(bizerrors.FKFieldCountMismatch, "")
	}

	// 2. Validate source fields exist on ModelID
	sourceModel, err := s.modelRepo.GetByID(ctx, cmd.ModelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil {
		return nil, err
	}
	if err := validateFieldsExistOnModel(sourceModel, cmd.SourceFields); err != nil {
		return nil, bizerrors.NewError(bizerrors.FKColumnsNotFound, err.Error())
	}

	// 3. Validate target fields exist on RefModelID
	refModel, err := s.modelRepo.GetByID(ctx, cmd.RefModelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil {
		return nil, err
	}
	if err := validateFieldsExistOnModel(refModel, cmd.TargetFields); err != nil {
		return nil, bizerrors.NewError(bizerrors.FKColumnsNotFound, err.Error())
	}

	// 4. Generate IDs
	normalID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, fmt.Errorf("CreateLogicalForeignKey: generate normalID: %w", err)
	}
	reverseID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, fmt.Errorf("CreateLogicalForeignKey: generate reverseID: %w", err)
	}
	pairID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, fmt.Errorf("CreateLogicalForeignKey: generate pairID: %w", err)
	}

	now := time.Now()

	// 5. Create normal row (ModelID owns FK columns)
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
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 6. Create reverse row (RefModelID mirrors the FK)
	reverseRow := &modeldesign.LogicalForeignKey{
		ID:           reverseID,
		PairID:       pairID,
		OrgName:      cmd.OrgName,
		Direction:    modeldesign.DirectionReverse,
		ModelID:      cmd.RefModelID,
		ModelName:    refModel.ModelName,
		RefModelID:   cmd.ModelID,
		RefModelName: sourceModel.ModelName,
		SourceFields: cmd.TargetFields, // mirrored
		TargetFields: cmd.SourceFields, // mirrored
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 7. Save both rows atomically
	if err := s.txManager.WithTx(ctx, func(txCtx context.Context, q dbgen.Querier) error {
		txFKRepo := s.fkRepoFactory(q)
		if err := txFKRepo.Save(txCtx, normalRow); err != nil {
			return fmt.Errorf("save normal FK row: %w", err)
		}
		if err := txFKRepo.Save(txCtx, reverseRow); err != nil {
			return fmt.Errorf("save reverse FK row: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return normalRow, nil
}

// DeleteLogicalForeignKey deletes a FK pair by pair_id.
//
// Pre-condition: both FK rows (normal and reverse) must have zero relate_fk_id references.
func (s *LogicalFKAppService) DeleteLogicalForeignKey(
	ctx context.Context,
	cmd DeleteLogicalForeignKeyCommand,
) error {
	// 1. Find the FK pair
	rows, err := s.fkRepo.FindByPairID(ctx, cmd.OrgName, cmd.PairID)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return bizerrors.NewError(bizerrors.FKNotFound, cmd.PairID)
	}

	// 2. Check that no RELATION fields reference either FK row
	for _, row := range rows {
		relateFields, err := s.fkRepo.FindByRelateField(ctx, cmd.OrgName, row.ID)
		if err != nil {
			return fmt.Errorf("DeleteLogicalForeignKey: check relate fields for %s: %w", row.ID, err)
		}
		if len(relateFields) > 0 {
			return bizerrors.NewError(bizerrors.FKPairHasRelateFields, row.ID)
		}
	}

	// 3. Delete the pair
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
