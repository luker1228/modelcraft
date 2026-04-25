package enduser

import (
	"context"
	"strings"

	domainenduser "modelcraft/internal/domain/enduser"
	"modelcraft/internal/domain/shared"
	infrrepo "modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
)

// EndUserProjectAccessAppService handles end-user project access management use cases.
type EndUserProjectAccessAppService struct {
	privateDBManager PrivateDBManager
}

// NewEndUserProjectAccessAppService creates a new EndUserProjectAccessAppService.
func NewEndUserProjectAccessAppService(privateDBManager PrivateDBManager) *EndUserProjectAccessAppService {
	return &EndUserProjectAccessAppService{privateDBManager: privateDBManager}
}

// GrantProjectAccess grants a permission bundle to one end-user in project scope.
func (s *EndUserProjectAccessAppService) GrantProjectAccess(
	ctx context.Context,
	cmd GrantEndUserProjectAccessCommand,
) (*EndUserProjectAccessDTO, error) {
	if msg := validateProjectAccessScope(cmd.OrgName, cmd.ProjectSlug); msg != "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, msg)
	}
	if strings.TrimSpace(cmd.EndUserID) == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "endUserId is required")
	}
	if strings.TrimSpace(cmd.PermissionBundleID) == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "permissionBundleId is required")
	}

	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBError(ctx, cmd.ProjectSlug, err)
	}

	userRepo := infrrepo.NewSqlEndUserRepository(db, cmd.OrgName, cmd.ProjectSlug)
	accessRepo := infrrepo.NewSqlEndUserProjectAccessRepository(db, cmd.OrgName, cmd.ProjectSlug)

	user, err := userRepo.GetByID(ctx, cmd.OrgName, cmd.EndUserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.EndUserID)
	}

	exists, err := accessRepo.PermissionBundleExists(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.PermissionBundleID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if !exists {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.EndUserPermissionBundleNotFound,
			cmd.PermissionBundleID,
		)
	}

	accessID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to generate project access id")
	}

	access, err := domainenduser.NewEndUserProjectAccess(
		accessID,
		cmd.OrgName,
		cmd.ProjectSlug,
		cmd.EndUserID,
		cmd.PermissionBundleID,
		cmd.GrantedBy,
	)
	if err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, err.Error())
	}

	if err = accessRepo.Grant(ctx, access); err != nil {
		if shared.IsRepoError(err, shared.ErrTypeDuplicatedKey) || shared.IsDuplicateKeyError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.Conflict, "end user project access already exists")
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	created, err := accessRepo.GetByID(ctx, cmd.OrgName, cmd.ProjectSlug, accessID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if created == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.NotFound, accessID)
	}

	return s.toAccessDTO(created), nil
}

// UpdateProjectAccess updates the permission bundle of one existing access grant.
func (s *EndUserProjectAccessAppService) UpdateProjectAccess(
	ctx context.Context,
	cmd UpdateEndUserProjectAccessCommand,
) (*EndUserProjectAccessDTO, error) {
	if msg := validateProjectAccessScope(cmd.OrgName, cmd.ProjectSlug); msg != "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, msg)
	}
	if strings.TrimSpace(cmd.AccessID) == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "accessId is required")
	}
	if strings.TrimSpace(cmd.PermissionBundleID) == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "permissionBundleId is required")
	}

	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBError(ctx, cmd.ProjectSlug, err)
	}

	accessRepo := infrrepo.NewSqlEndUserProjectAccessRepository(db, cmd.OrgName, cmd.ProjectSlug)

	exists, err := accessRepo.PermissionBundleExists(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.PermissionBundleID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if !exists {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.EndUserPermissionBundleNotFound,
			cmd.PermissionBundleID,
		)
	}

	access, err := accessRepo.GetByID(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.AccessID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if access == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.NotFound, cmd.AccessID)
	}

	if err = accessRepo.UpdatePermissionBundle(
		ctx,
		cmd.OrgName,
		cmd.ProjectSlug,
		cmd.AccessID,
		cmd.PermissionBundleID,
	); err != nil {
		if shared.IsRepoError(err, shared.ErrTypeNoRowsAffected) || shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.NotFound, cmd.AccessID)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	updated, err := accessRepo.GetByID(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.AccessID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if updated == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.NotFound, cmd.AccessID)
	}

	return s.toAccessDTO(updated), nil
}

// RevokeProjectAccess revokes one project access grant by access ID.
func (s *EndUserProjectAccessAppService) RevokeProjectAccess(
	ctx context.Context,
	cmd RevokeEndUserProjectAccessCommand,
) error {
	if msg := validateProjectAccessScope(cmd.OrgName, cmd.ProjectSlug); msg != "" {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, msg)
	}
	if strings.TrimSpace(cmd.AccessID) == "" {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "accessId is required")
	}

	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return s.convertDBError(ctx, cmd.ProjectSlug, err)
	}

	accessRepo := infrrepo.NewSqlEndUserProjectAccessRepository(db, cmd.OrgName, cmd.ProjectSlug)
	if err = accessRepo.Revoke(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.AccessID); err != nil {
		if shared.IsRepoError(err, shared.ErrTypeNoRowsAffected) || shared.IsNotFoundError(err) {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.NotFound, cmd.AccessID)
		}
		return bizerrors.ConvertRepositoryError(ctx, err)
	}

	return nil
}

// ListProjectAccess lists project access grants with cursor pagination.
func (s *EndUserProjectAccessAppService) ListProjectAccess(
	ctx context.Context,
	cmd ListEndUserProjectAccessCommand,
) (*ListEndUserProjectAccessResult, error) {
	if msg := validateProjectAccessScope(cmd.OrgName, cmd.ProjectSlug); msg != "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, msg)
	}

	first := cmd.First
	if first <= 0 {
		first = 20
	}
	if first > 100 {
		first = 100
	}

	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBError(ctx, cmd.ProjectSlug, err)
	}

	accessRepo := infrrepo.NewSqlEndUserProjectAccessRepository(db, cmd.OrgName, cmd.ProjectSlug)
	items, totalCount, err := accessRepo.ListWithTotal(ctx, domainenduser.ListEndUserProjectAccessQuery{
		OrgName:     cmd.OrgName,
		ProjectSlug: cmd.ProjectSlug,
		Search:      cmd.Search,
		First:       first,
		After:       cmd.After,
	})
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	dtos := make([]*EndUserProjectAccessDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, s.toAccessDTO(item))
	}

	var endCursor string
	if len(dtos) > 0 {
		endCursor = dtos[len(dtos)-1].ID
	}

	return &ListEndUserProjectAccessResult{
		Items:       dtos,
		TotalCount:  totalCount,
		HasNextPage: len(dtos) == first,
		EndCursor:   endCursor,
	}, nil
}

func (s *EndUserProjectAccessAppService) convertDBError(ctx context.Context, projectSlug string, err error) error {
	if shared.IsNotFoundError(err) {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ProjectNotFound, projectSlug)
	}
	return bizerrors.ConvertRepositoryError(ctx, err)
}

func (s *EndUserProjectAccessAppService) toAccessDTO(
	entity *domainenduser.EndUserProjectAccess,
) *EndUserProjectAccessDTO {
	if entity == nil {
		return nil
	}

	var userDTO *EndUserDTO
	if entity.EndUser != nil {
		userDTO = &EndUserDTO{
			ID:          entity.EndUser.ID,
			Username:    entity.EndUser.Username,
			IsForbidden: entity.EndUser.IsForbidden,
			CreatedBy:   entity.EndUser.CreatedBy,
			CreatedAt:   entity.EndUser.CreatedAt,
			UpdatedAt:   entity.EndUser.UpdatedAt,
		}
	}

	return &EndUserProjectAccessDTO{
		ID:                   entity.ID,
		EndUser:              userDTO,
		PermissionBundleID:   entity.PermissionBundleID,
		PermissionBundleName: entity.PermissionName,
		GrantedBy:            entity.GrantedBy,
		GrantedAt:            entity.GrantedAt,
	}
}

func validateProjectAccessScope(orgName, projectSlug string) string {
	if strings.TrimSpace(orgName) == "" {
		return "orgName is required"
	}
	if strings.TrimSpace(projectSlug) == "" {
		return "projectSlug is required"
	}
	return ""
}
