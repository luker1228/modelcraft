package modeldesign

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/lexorder"

	"github.com/google/uuid"
)

// ModelGroupAppService handles use cases for model group management.
type ModelGroupAppService struct {
	groupRepo modeldesign.ModelGroupRepository
	modelRepo modeldesign.ModelRepository
	txManager repository.TxManager
}

// NewModelGroupAppService creates a new ModelGroupAppService.
func NewModelGroupAppService(
	groupRepo modeldesign.ModelGroupRepository,
	modelRepo modeldesign.ModelRepository,
	txManager repository.TxManager,
) *ModelGroupAppService {
	return &ModelGroupAppService{
		groupRepo: groupRepo,
		modelRepo: modelRepo,
		txManager: txManager,
	}
}

// CreateGroup validates and persists a new model group.
// Returns GroupAlreadyExists if a group with the same name exists in the project.
// Returns ParamInvalid if the name fails validation.
func (s *ModelGroupAppService) CreateGroup(
	ctx context.Context,
	cmd CreateGroupCommand,
) (*modeldesign.ModelGroup, error) {
	if err := modeldesign.ValidateGroupName(cmd.Name); err != nil {
		return nil, err
	}

	existing, err := s.groupRepo.FindByName(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.Name)
	if err != nil && !shared.IsNotFoundError(err) {
		return nil, err
	}
	if existing != nil {
		return nil, bizerrors.NewError(bizerrors.GroupAlreadyExists, cmd.Name)
	}

	tail, err := s.groupRepo.GetTailDisplayOrder(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, err
	}
	order, err := lexorder.Midpoint(tail, "")
	if err != nil {
		return nil, bizerrors.Wrapf(err, "compute display order")
	}

	group := &modeldesign.ModelGroup{
		ID: uuid.NewString(),
		ProjectScope: project.ProjectScope{
			OrgName:     cmd.OrgName,
			ProjectSlug: cmd.ProjectSlug,
		},
		Name:         cmd.Name,
		DisplayOrder: order,
	}
	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

// RenameGroup changes the display name of an existing group.
// Returns GroupNotFound if the group does not exist.
// Returns GroupAlreadyExists if another group in the project already has the new name.
// Returns ParamInvalid if the new name fails validation.
func (s *ModelGroupAppService) RenameGroup(
	ctx context.Context,
	cmd RenameGroupCommand,
) (*modeldesign.ModelGroup, error) {
	group, err := s.groupRepo.FindByID(ctx, cmd.GroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, bizerrors.NewError(bizerrors.GroupNotFound, cmd.GroupID)
	}

	if err := modeldesign.ValidateGroupName(cmd.NewName); err != nil {
		return nil, err
	}

	conflict, err := s.groupRepo.FindByName(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.NewName)
	if err != nil && !shared.IsNotFoundError(err) {
		return nil, err
	}
	if conflict != nil && conflict.ID != cmd.GroupID {
		return nil, bizerrors.NewError(bizerrors.GroupAlreadyExists, cmd.NewName)
	}

	group.Name = cmd.NewName
	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

// DeleteGroup removes a group and cascades its models to ungrouped in a transaction.
// Returns GroupNotFound if the group does not exist.
func (s *ModelGroupAppService) DeleteGroup(ctx context.Context, groupID string) error {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return bizerrors.NewError(bizerrors.GroupNotFound, groupID)
	}

	if s.txManager != nil {
		return s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
			txGroupRepo := repository.NewSqlModelGroupRepository(q)
			if txErr := txGroupRepo.UpdateModelsGroup(ctx, groupID, nil); txErr != nil {
				return txErr
			}
			return txGroupRepo.Delete(ctx, groupID)
		})
	}

	// No TxManager (used in unit tests with mocked repos).
	if err := s.groupRepo.UpdateModelsGroup(ctx, groupID, nil); err != nil {
		return err
	}
	return s.groupRepo.Delete(ctx, groupID)
}

// ReorderGroup moves a group to a new position within the project's ordered list.
// Pass nil for AfterGroupID to move the group to the head.
// Returns GroupNotFound if the target group does not exist.
func (s *ModelGroupAppService) ReorderGroup(
	ctx context.Context,
	cmd ReorderGroupCommand,
) error {
	group, err := s.groupRepo.FindByID(ctx, cmd.GroupID)
	if err != nil {
		return err
	}
	if group == nil {
		return bizerrors.NewError(bizerrors.GroupNotFound, cmd.GroupID)
	}

	groups, err := s.groupRepo.ListByProject(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return err
	}

	prev, next := computeNeighbors(groups, cmd.GroupID, cmd.AfterGroupID)
	newOrder, err := lexorder.Midpoint(prev, next)
	if err != nil {
		// Collision: renumber all groups.
		return s.renumberAll(ctx, groups)
	}

	group.DisplayOrder = newOrder
	return s.groupRepo.Update(ctx, group)
}

// computeNeighbors returns the display_order values of the group immediately before
// and after the desired insertion position.
// afterGroupID == nil means insert at head.
func computeNeighbors(
	groups []*modeldesign.ModelGroup,
	movingID string,
	afterGroupID *string,
) (prev, next string) {
	// Build a list excluding the group being moved.
	others := make([]*modeldesign.ModelGroup, 0, len(groups))
	for _, g := range groups {
		if g.ID != movingID {
			others = append(others, g)
		}
	}

	if afterGroupID == nil {
		// Move to head: prev="" (before first), next=first group if exists.
		if len(others) > 0 {
			return "", others[0].DisplayOrder
		}
		return "", ""
	}

	// Find the group with afterGroupID and return its order and the next one's order.
	for i, g := range others {
		if g.ID == *afterGroupID {
			p := g.DisplayOrder
			n := ""
			if i+1 < len(others) {
				n = others[i+1].DisplayOrder
			}
			return p, n
		}
	}
	// afterGroupID not found — insert at tail.
	if len(others) > 0 {
		return others[len(others)-1].DisplayOrder, ""
	}
	return "", ""
}

// renumberAll reassigns evenly-spaced display_order values to all groups.
func (s *ModelGroupAppService) renumberAll(
	ctx context.Context,
	groups []*modeldesign.ModelGroup,
) error {
	orders, err := lexorder.Renumber(len(groups))
	if err != nil {
		return bizerrors.Wrapf(err, "renumber groups")
	}
	for i, g := range groups {
		g.DisplayOrder = orders[i]
		if updErr := s.groupRepo.Update(ctx, g); updErr != nil {
			return updErr
		}
	}
	return nil
}

// MoveModelToGroup assigns a model to a group, or clears its group assignment.
// Returns GroupNotFound if the specified group does not exist.
func (s *ModelGroupAppService) MoveModelToGroup(
	ctx context.Context,
	cmd MoveModelToGroupCommand,
) error {
	if cmd.GroupID != nil {
		group, err := s.groupRepo.FindByID(ctx, *cmd.GroupID)
		if err != nil {
			return err
		}
		if group == nil {
			return bizerrors.NewError(bizerrors.GroupNotFound, *cmd.GroupID)
		}
	}

	model, err := s.modelRepo.GetByID(ctx, cmd.ModelID)
	if err != nil {
		return err
	}
	if model == nil {
		return bizerrors.NewError(bizerrors.ModelNotFound, cmd.ModelID)
	}

	model.GroupID = cmd.GroupID
	return s.modelRepo.Update(ctx, model)
}

// ListGroups returns all groups for a project ordered by display_order,
// with the virtual ungrouped sentinel appended at the end.
func (s *ModelGroupAppService) ListGroups(
	ctx context.Context,
	orgName, projectSlug string,
) ([]*modeldesign.ModelGroup, error) {
	groups, err := s.groupRepo.ListByProject(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}
	return append(groups, modeldesign.NewUngroupedGroup()), nil
}
