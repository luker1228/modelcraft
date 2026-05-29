package modeldatabase

import (
	"context"
	appmodeldesign "modelcraft/internal/app/modeldesign"
	domainmodel "modelcraft/internal/domain/modeldesign"
)

type ImportGroupService struct {
	groupRepo domainmodel.ModelGroupRepository
	groupApp  *appmodeldesign.ModelGroupAppService
}

func NewImportGroupService(
	groupRepo domainmodel.ModelGroupRepository,
	groupApp *appmodeldesign.ModelGroupAppService,
) *ImportGroupService {
	return &ImportGroupService{groupRepo: groupRepo, groupApp: groupApp}
}

func (s *ImportGroupService) EnsureImportGroup(
	ctx context.Context,
	orgName, projectSlug string,
) (*domainmodel.ModelGroup, error) {
	group, err := s.groupRepo.FindByName(ctx, orgName, projectSlug, importGroupName)
	if err != nil {
		return nil, err
	}
	if group != nil {
		return group, nil
	}
	return s.groupApp.CreateGroup(ctx, appmodeldesign.CreateGroupCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Name:        importGroupName,
	})
}

func (s *ImportGroupService) MoveModelToGroup(ctx context.Context, modelID string, groupID *string) error {
	return s.groupApp.MoveModelToGroup(ctx, appmodeldesign.MoveModelToGroupCommand{
		ModelID: modelID,
		GroupID: groupID,
	})
}
