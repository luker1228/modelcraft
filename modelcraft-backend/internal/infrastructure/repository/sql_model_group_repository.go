package repository

import (
	"context"
	"database/sql"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// ModelGroupToDomain converts a dbgen.ModelGroup row to a domain ModelGroup.
func ModelGroupToDomain(row dbgen.ModelGroup) *modeldesign.ModelGroup {
	var createdAt, updatedAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}
	return &modeldesign.ModelGroup{
		ID: row.ID,
		ProjectScope: project.ProjectScope{
			OrgName:     row.OrgName,
			ProjectSlug: row.ProjectSlug,
		},
		Name:         row.Name,
		DisplayOrder: row.DisplayOrder,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

// ModelGroupToCreateParams converts a domain ModelGroup to dbgen create params.
func ModelGroupToCreateParams(g *modeldesign.ModelGroup) dbgen.CreateModelGroupParams {
	return dbgen.CreateModelGroupParams{
		ID:           g.ID,
		OrgName:      g.OrgName,
		ProjectSlug:  g.ProjectSlug,
		Name:         g.Name,
		DisplayOrder: g.DisplayOrder,
	}
}

// SqlModelGroupRepository is the sqlc-based implementation of modeldesign.ModelGroupRepository.
type SqlModelGroupRepository struct {
	q dbgen.Querier
}

// NewSqlModelGroupRepository creates a new SqlModelGroupRepository.
func NewSqlModelGroupRepository(q dbgen.Querier) modeldesign.ModelGroupRepository {
	return &SqlModelGroupRepository{q: q}
}

// Create persists a new group.
func (r *SqlModelGroupRepository) Create(ctx context.Context, group *modeldesign.ModelGroup) error {
	return ExecWithErrorHandling(func() error {
		return r.q.CreateModelGroup(ctx, ModelGroupToCreateParams(group))
	})
}

// FindByID retrieves a group by its unique identifier.
// Returns nil, nil if the group does not exist.
func (r *SqlModelGroupRepository) FindByID(ctx context.Context, id string) (*modeldesign.ModelGroup, error) {
	var row dbgen.ModelGroup
	err := QueryWithSQLErrorHandling(func() error {
		var e error
		row, e = r.q.GetModelGroupByID(ctx, id)
		return e
	})
	if err != nil {
		if IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("model group not found by id: " + id)
		}
		return nil, err
	}
	return ModelGroupToDomain(row), nil
}

// FindByName retrieves a group by its name within a project scope.
// Returns nil, shared.NewNotFoundError if the group does not exist.
func (r *SqlModelGroupRepository) FindByName(
	ctx context.Context, orgName, projectSlug, name string,
) (*modeldesign.ModelGroup, error) {
	var row dbgen.ModelGroup
	err := QueryWithSQLErrorHandling(func() error {
		var e error
		row, e = r.q.GetModelGroupByName(ctx, dbgen.GetModelGroupByNameParams{
			OrgName:     orgName,
			ProjectSlug: projectSlug,
			Name:        name,
		})
		return e
	})
	if err != nil {
		if IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("model group not found: " + name)
		}
		return nil, err
	}
	return ModelGroupToDomain(row), nil
}

// ListByProject retrieves all groups in a project ordered by display_order ascending.
func (r *SqlModelGroupRepository) ListByProject(
	ctx context.Context, orgName, projectSlug string,
) ([]*modeldesign.ModelGroup, error) {
	var rows []dbgen.ModelGroup
	if err := QueryWithSQLErrorHandling(func() error {
		var e error
		rows, e = r.q.ListModelGroupsByProject(ctx, dbgen.ListModelGroupsByProjectParams{
			OrgName:     orgName,
			ProjectSlug: projectSlug,
		})
		return e
	}); err != nil {
		return nil, err
	}

	groups := make([]*modeldesign.ModelGroup, len(rows))
	for i, row := range rows {
		groups[i] = ModelGroupToDomain(row)
	}
	return groups, nil
}

// Update persists changes to an existing group (name, display_order).
func (r *SqlModelGroupRepository) Update(ctx context.Context, group *modeldesign.ModelGroup) error {
	return ExecWithErrorHandling(func() error {
		return r.q.UpdateModelGroup(ctx, dbgen.UpdateModelGroupParams{
			Name:         group.Name,
			DisplayOrder: group.DisplayOrder,
			ID:           group.ID,
		})
	})
}

// Delete removes a group by ID.
func (r *SqlModelGroupRepository) Delete(ctx context.Context, id string) error {
	return ExecWithErrorHandling(func() error {
		return r.q.DeleteModelGroup(ctx, id)
	})
}

// UpdateModelsGroup sets group_id = newGroupID on all models currently assigned to groupID.
// Pass nil for newGroupID to move models to ungrouped.
func (r *SqlModelGroupRepository) UpdateModelsGroup(
	ctx context.Context, groupID string, newGroupID *string,
) error {
	return ExecWithErrorHandling(func() error {
		return r.q.UpdateModelsGroupID(ctx, dbgen.UpdateModelsGroupIDParams{
			GroupID:   PtrToNullStr(newGroupID),
			GroupID_2: sql.NullString{String: groupID, Valid: true},
		})
	})
}

// GetTailDisplayOrder returns the largest display_order value among groups in the project,
// or an empty string if no groups exist.
func (r *SqlModelGroupRepository) GetTailDisplayOrder(
	ctx context.Context, orgName, projectSlug string,
) (string, error) {
	var order string
	err := QueryWithSQLErrorHandling(func() error {
		var e error
		order, e = r.q.GetTailModelGroupDisplayOrder(ctx, dbgen.GetTailModelGroupDisplayOrderParams{
			OrgName:     orgName,
			ProjectSlug: projectSlug,
		})
		return e
	})
	if err != nil {
		if IsNotFoundError(err) {
			return "", nil
		}
		return "", bizerrors.Wrapf(err, "failed to get tail display order")
	}
	return order, nil
}

// compile-time interface check
var _ modeldesign.ModelGroupRepository = (*SqlModelGroupRepository)(nil)
