package modeldesign

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/pkg/bizerrors"
)

// AddFieldFKService handles FK-related validation when adding fields.
// It is a lightweight helper used by ModelDesignAppService.AddFieldSync.
type AddFieldFKService struct {
	modelRepo modeldesign.ModelRepository
	fkRepo    modeldesign.LogicalForeignKeyRepository
}

// newAddFieldFKService creates a new AddFieldFKService.
func newAddFieldFKService(
	modelRepo modeldesign.ModelRepository,
	fkRepo modeldesign.LogicalForeignKeyRepository,
) *AddFieldFKService {
	return &AddFieldFKService{
		modelRepo: modelRepo,
		fkRepo:    fkRepo,
	}
}

// ValidateRelateFKID validates that the given relate_fk_id references an existing
// LogicalForeignKey row whose model_id matches the given modelID.
//
// If relateFKID is nil, validation is skipped (non-RELATION fields).
func (s *AddFieldFKService) ValidateRelateFKID(
	ctx context.Context,
	modelID string,
	relateFKID *string,
) error {
	if relateFKID == nil {
		return nil
	}

	// Find the FK row by scanning all rows in the pair that contains this FK ID.
	// We use FindByPairID here only as a workaround until a FindByID query is added.
	// For now we look through FindByModel to find the matching row.
	rows, err := s.fkRepo.FindByModel(ctx, modelID)
	if err != nil {
		return fmt.Errorf("ValidateRelateFKID: look up FK rows for model %s: %w", modelID, err)
	}

	for _, row := range rows {
		if row.ID == *relateFKID {
			if row.ModelID != modelID {
				return bizerrors.NewError(
					bizerrors.FKNotFound,
					fmt.Sprintf("FK row %s does not belong to model %s", *relateFKID, modelID),
				)
			}
			return nil
		}
	}

	return bizerrors.NewError(
		bizerrors.FKNotFound,
		fmt.Sprintf("FK row %s not found on model %s", *relateFKID, modelID),
	)
}
