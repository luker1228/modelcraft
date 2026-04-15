package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"time"
)

// logicalFKQuerier defines the minimal querier interface needed by SqlLogicalForeignKeyRepository.
type logicalFKQuerier interface {
	CreateLogicalForeignKey(ctx context.Context, arg dbgen.CreateLogicalForeignKeyParams) error
	DeleteLogicalForeignKeyByPairID(ctx context.Context, arg dbgen.DeleteLogicalForeignKeyByPairIDParams) error
	FindLogicalForeignKeysByModelID(
		ctx context.Context, arg dbgen.FindLogicalForeignKeysByModelIDParams,
	) ([]dbgen.LogicalForeignKey, error)
	FindLogicalForeignKeysByPairID(
		ctx context.Context, arg dbgen.FindLogicalForeignKeysByPairIDParams,
	) ([]dbgen.LogicalForeignKey, error)
	GetLogicalForeignKeyByID(ctx context.Context, id string) (dbgen.LogicalForeignKey, error)
	FindFieldsByBelongsToFKID(
		ctx context.Context, arg dbgen.FindFieldsByBelongsToFKIDParams,
	) ([]dbgen.FindFieldsByBelongsToFKIDRow, error)
	FindFieldsByRelateFKID(
		ctx context.Context, arg dbgen.FindFieldsByRelateFKIDParams,
	) ([]dbgen.FindFieldsByRelateFKIDRow, error)
}

// SqlLogicalForeignKeyRepository is the sqlc-based implementation of modeldesign.LogicalForeignKeyRepository.
type SqlLogicalForeignKeyRepository struct {
	q logicalFKQuerier
}

// NewSqlLogicalForeignKeyRepository creates a SqlLogicalForeignKeyRepository.
func NewSqlLogicalForeignKeyRepository(q dbgen.Querier) modeldesign.LogicalForeignKeyRepository {
	return &SqlLogicalForeignKeyRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// Save creates a new logical foreign key record.
func (r *SqlLogicalForeignKeyRepository) Save(ctx context.Context, lf *modeldesign.LogicalForeignKey) error {
	params, err := LogicalForeignKeyToCreateParams(lf)
	if err != nil {
		return fmt.Errorf("Save logical foreign key: %w", err)
	}
	return r.q.CreateLogicalForeignKey(ctx, params)
}

// DeleteByPairID deletes both rows of a FK pair atomically, scoped to the given org.
func (r *SqlLogicalForeignKeyRepository) DeleteByPairID(ctx context.Context, orgName, pairID string) error {
	return r.q.DeleteLogicalForeignKeyByPairID(ctx, dbgen.DeleteLogicalForeignKeyByPairIDParams{
		PairID:  pairID,
		OrgName: orgName,
	})
}

// GetByID finds a single logical foreign key by its ID.
func (r *SqlLogicalForeignKeyRepository) GetByID(
	ctx context.Context, id string,
) (*modeldesign.LogicalForeignKey, error) {
	row, err := r.q.GetLogicalForeignKeyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return LogicalForeignKeyToDomain(row)
}

// FindByModel finds all logical foreign keys where model_id = modelID within the org.
func (r *SqlLogicalForeignKeyRepository) FindByModel(
	ctx context.Context, orgName, modelID string,
) ([]*modeldesign.LogicalForeignKey, error) {
	rows, err := r.q.FindLogicalForeignKeysByModelID(ctx, dbgen.FindLogicalForeignKeysByModelIDParams{
		OrgName: orgName,
		ModelID: modelID,
	})
	if err != nil {
		return nil, err
	}
	return logicalForeignKeyRowsToDomain(rows)
}

// FindByPairID finds both rows of a FK pair by pair_id within the org.
func (r *SqlLogicalForeignKeyRepository) FindByPairID(
	ctx context.Context, orgName, pairID string,
) ([]*modeldesign.LogicalForeignKey, error) {
	rows, err := r.q.FindLogicalForeignKeysByPairID(ctx, dbgen.FindLogicalForeignKeysByPairIDParams{
		OrgName: orgName,
		PairID:  pairID,
	})
	if err != nil {
		return nil, err
	}
	return logicalForeignKeyRowsToDomain(rows)
}

// FindByBelongsToField finds logical foreign keys referenced by the given belongs_to_fk_id within the org.
// lfID is the logical_foreign_key.id value stored in field_definitions.belongs_to_fk_id.
func (r *SqlLogicalForeignKeyRepository) FindByBelongsToField(
	ctx context.Context, orgName, lfID string,
) ([]*modeldesign.LogicalForeignKey, error) {
	fields, err := r.q.FindFieldsByBelongsToFKID(ctx, dbgen.FindFieldsByBelongsToFKIDParams{
		BelongsToFkID: sql.NullString{String: lfID, Valid: true},
		OrgName:       orgName,
	})
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return []*modeldesign.LogicalForeignKey{}, nil
	}
	// Return the FK row itself
	rows, err := r.FindByPairID(ctx, orgName, lfID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return []*modeldesign.LogicalForeignKey{}, nil
		}
		return nil, err
	}
	return rows, nil
}

// FindByRelateField finds logical foreign keys referenced by the given relate_fk_id within the org.
func (r *SqlLogicalForeignKeyRepository) FindByRelateField(
	ctx context.Context, orgName, lfID string,
) ([]*modeldesign.LogicalForeignKey, error) {
	fields, err := r.q.FindFieldsByRelateFKID(ctx, dbgen.FindFieldsByRelateFKIDParams{
		RelateFkID: sql.NullString{String: lfID, Valid: true},
		OrgName:    orgName,
	})
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return []*modeldesign.LogicalForeignKey{}, nil
	}
	rows, err := r.FindByPairID(ctx, orgName, lfID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return []*modeldesign.LogicalForeignKey{}, nil
		}
		return nil, err
	}
	return rows, nil
}

// LogicalForeignKeyToCreateParams converts a domain LogicalForeignKey to dbgen.CreateLogicalForeignKeyParams.
func LogicalForeignKeyToCreateParams(lf *modeldesign.LogicalForeignKey) (dbgen.CreateLogicalForeignKeyParams, error) {
	sourceFieldsJSON, err := json.Marshal(lf.SourceFields)
	if err != nil {
		return dbgen.CreateLogicalForeignKeyParams{}, fmt.Errorf("marshal source_fields: %w", err)
	}
	targetFieldsJSON, err := json.Marshal(lf.TargetFields)
	if err != nil {
		return dbgen.CreateLogicalForeignKeyParams{}, fmt.Errorf("marshal target_fields: %w", err)
	}
	sf := json.RawMessage(sourceFieldsJSON)
	tf := json.RawMessage(targetFieldsJSON)
	return dbgen.CreateLogicalForeignKeyParams{
		ID:           lf.ID,
		PairID:       lf.PairID,
		OrgName:      lf.OrgName,
		Direction:    dbgen.LogicalForeignKeysDirection(lf.Direction),
		ModelID:      lf.ModelID,
		ModelName:    lf.ModelName,
		RefModelID:   lf.RefModelID,
		RefModelName: lf.RefModelName,
		SourceFields: sf,
		TargetFields: tf,
	}, nil
}

// LogicalForeignKeyToDomain converts a dbgen.LogicalForeignKey row to a domain entity.
func LogicalForeignKeyToDomain(row dbgen.LogicalForeignKey) (*modeldesign.LogicalForeignKey, error) {
	var sourceFields []string
	if len(row.SourceFields) > 0 {
		if err := json.Unmarshal(row.SourceFields, &sourceFields); err != nil {
			return nil, fmt.Errorf("LogicalForeignKeyToDomain: unmarshal source_fields: %w", err)
		}
	}
	var targetFields []string
	if len(row.TargetFields) > 0 {
		if err := json.Unmarshal(row.TargetFields, &targetFields); err != nil {
			return nil, fmt.Errorf("LogicalForeignKeyToDomain: unmarshal target_fields: %w", err)
		}
	}

	var createdAt, updatedAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	return &modeldesign.LogicalForeignKey{
		ID:           row.ID,
		PairID:       row.PairID,
		OrgName:      row.OrgName,
		Direction:    modeldesign.LogicalFKDirection(row.Direction),
		ModelID:      row.ModelID,
		ModelName:    row.ModelName,
		RefModelID:   row.RefModelID,
		RefModelName: row.RefModelName,
		SourceFields: sourceFields,
		TargetFields: targetFields,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

// logicalForeignKeyRowsToDomain converts a slice of dbgen.LogicalForeignKey rows to domain entities.
func logicalForeignKeyRowsToDomain(rows []dbgen.LogicalForeignKey) ([]*modeldesign.LogicalForeignKey, error) {
	result := make([]*modeldesign.LogicalForeignKey, len(rows))
	for i, row := range rows {
		lf, err := LogicalForeignKeyToDomain(row)
		if err != nil {
			return nil, err
		}
		result[i] = lf
	}
	return result, nil
}

// compile-time interface check
var _ modeldesign.LogicalForeignKeyRepository = (*SqlLogicalForeignKeyRepository)(nil)
