package adapter

import (
	"context"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
	"strings"
)

// RbacErrorAdapter converts BusinessError to RBAC GraphQL union errors.
type RbacErrorAdapter struct {
	ctx    context.Context
	logger logfacade.Logger
}

// NewRbacErrorAdapter creates a new RbacErrorAdapter.
func NewRbacErrorAdapter(ctx context.Context) *RbacErrorAdapter {
	return &RbacErrorAdapter{ctx: ctx, logger: logfacade.GetLogger(ctx)}
}

func (a *RbacErrorAdapter) logUnknown(op string, err *bizerrors.BusinessError) {
	a.logger.Errorf(a.ctx, "unknown RBAC error [%s]: %s", op, err.Info().GetCode())
}

// ConvertToCreatePermissionError maps to CreateEndUserPermissionError union.
func (a *RbacErrorAdapter) ConvertToCreatePermissionError(
	err *bizerrors.BusinessError,
) generated.CreateEndUserPermissionError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.ModelNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeModel}
	case bizerrors.EndUserRowScopeFieldMissing.GetCode():
		requiredField := "owner"
		return &generated.RowScopeFieldMissing{
			Message:            err.Msg(),
			MissingField:       requiredField,
			RequiredByRowScope: generated.RowScopeTypeAll,
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("createPermission", err)
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToUpdatePermissionError maps to UpdateEndUserPermissionError union.
func (a *RbacErrorAdapter) ConvertToUpdatePermissionError(
	err *bizerrors.BusinessError,
) generated.UpdateEndUserPermissionError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserPermission}
	case bizerrors.EndUserRowScopeFieldMissing.GetCode():
		return &generated.RowScopeFieldMissing{
			Message:            err.Msg(),
			MissingField:       "owner",
			RequiredByRowScope: generated.RowScopeTypeAll,
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("updatePermission", err)
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToDeletePermissionError maps to DeleteEndUserPermissionError union.
func (a *RbacErrorAdapter) ConvertToDeletePermissionError(
	err *bizerrors.BusinessError,
) generated.DeleteEndUserPermissionError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserPermission}
	case bizerrors.EndUserPermissionInUse.GetCode():
		return &generated.EndUserPermissionInUse{Message: err.Msg()}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("deletePermission", err)
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserPermission}
	}
}

// ConvertToCreateBundleError maps to CreateEndUserPermissionBundleError union.
func (a *RbacErrorAdapter) ConvertToCreateBundleError(
	err *bizerrors.BusinessError,
) generated.CreateEndUserPermissionBundleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionBundleAlreadyExists.GetCode(), bizerrors.Conflict.GetCode():
		return &generated.EndUserPermissionBundleAlreadyExists{Message: err.Msg()}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("createBundle", err)
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToUpdateBundleError maps to UpdateEndUserPermissionBundleError union.
func (a *RbacErrorAdapter) ConvertToUpdateBundleError(
	err *bizerrors.BusinessError,
) generated.UpdateEndUserPermissionBundleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionBundleNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("updateBundle", err)
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToDeleteBundleError maps to DeleteEndUserPermissionBundleError union.
func (a *RbacErrorAdapter) ConvertToDeleteBundleError(
	err *bizerrors.BusinessError,
) generated.DeleteEndUserPermissionBundleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionBundleNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.EndUserPermissionBundleInUse.GetCode():
		return &generated.EndUserPermissionBundleInUse{Message: err.Msg()}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("deleteBundle", err)
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	}
}

// ConvertToAddPermissionToBundleError maps to AddEndUserPermissionToBundleError union.
func (a *RbacErrorAdapter) ConvertToAddPermissionToBundleError(
	err *bizerrors.BusinessError,
) generated.AddEndUserPermissionToBundleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.EndUserPermissionNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserPermission}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("addPermToBundle", err)
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToAddPresetToBundleError maps to AddEndUserPresetToBundleError union.
func (a *RbacErrorAdapter) ConvertToAddPresetToBundleError(
	err *bizerrors.BusinessError,
) generated.AddEndUserPresetToBundleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.ModelNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeModel}
	case bizerrors.EndUserPresetRequiresOwnerField.GetCode(), bizerrors.EndUserRowScopeFieldMissing.GetCode():
		suggestion := "请先在模型中创建 END_USER_REF 字段，然后重试绑定预设"
		preset := detectPresetFromMessage(err.Msg())
		return &generated.PresetRequiresOwnerField{
			Message:    err.Msg(),
			Preset:     preset,
			Suggestion: &suggestion,
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("addPresetToBundle", err)
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToRemovePermissionFromBundleError maps to RemoveEndUserPermissionFromBundleError union.
func (a *RbacErrorAdapter) ConvertToRemovePermissionFromBundleError(
	err *bizerrors.BusinessError,
) generated.RemoveEndUserPermissionFromBundleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.EndUserPermissionNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserPermission}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("removePermFromBundle", err)
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	}
}

// ConvertToCreateRoleError maps to CreateEndUserRoleError union.
func (a *RbacErrorAdapter) ConvertToCreateRoleError(
	err *bizerrors.BusinessError,
) generated.CreateEndUserRoleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserRoleAlreadyExists.GetCode(), bizerrors.Conflict.GetCode():
		return &generated.EndUserRoleAlreadyExists{Message: err.Msg()}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("createRole", err)
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToUpdateRoleError maps to UpdateEndUserRoleError union.
func (a *RbacErrorAdapter) ConvertToUpdateRoleError(
	err *bizerrors.BusinessError,
) generated.UpdateEndUserRoleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserRoleNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	case bizerrors.EndUserImplicitRoleCannotBeModified.GetCode():
		return &generated.EndUserImplicitRoleCannotBeModified{Message: err.Msg()}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("updateRole", err)
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToDeleteRoleError maps to DeleteEndUserRoleError union.
func (a *RbacErrorAdapter) ConvertToDeleteRoleError(
	err *bizerrors.BusinessError,
) generated.DeleteEndUserRoleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserRoleNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	case bizerrors.EndUserImplicitRoleCannotBeModified.GetCode():
		return &generated.EndUserImplicitRoleCannotBeModified{Message: err.Msg()}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("deleteRole", err)
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	}
}

// ConvertToAssignBundleToRoleError maps to AssignBundleToEndUserRoleError union.
func (a *RbacErrorAdapter) ConvertToAssignBundleToRoleError(
	err *bizerrors.BusinessError,
) generated.AssignBundleToEndUserRoleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserRoleNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("assignBundleToRole", err)
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	}
}

// ConvertToRevokeBundleFromRoleError maps to RevokeBundleFromEndUserRoleError union.
func (a *RbacErrorAdapter) ConvertToRevokeBundleFromRoleError(
	err *bizerrors.BusinessError,
) generated.RevokeBundleFromEndUserRoleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserRoleNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	default:
		a.logUnknown("revokeBundleFromRole", err)
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	}
}

// ConvertToAssignBundleToUserError maps to AssignBundleToEndUserError union.
func (a *RbacErrorAdapter) ConvertToAssignBundleToUserError(
	err *bizerrors.BusinessError,
) generated.AssignBundleToEndUserError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserNotFoundInProject.GetCode(), bizerrors.EndUserNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserInProject}
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.UserBundleAlreadyAssigned.GetCode():
		return &generated.UserBundleAlreadyAssigned{Message: err.Msg()}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("assignBundleToUser", err)
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserInProject}
	}
}

// ConvertToRevokeBundleFromUserError maps to RevokeBundleFromEndUserError union.
func (a *RbacErrorAdapter) ConvertToRevokeBundleFromUserError(
	err *bizerrors.BusinessError,
) generated.RevokeBundleFromEndUserError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserNotFoundInProject.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserInProject}
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	default:
		a.logUnknown("revokeBundleFromUser", err)
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserInProject}
	}
}

// ConvertToAssignRoleToUserError maps to AssignEndUserRoleError union.
func (a *RbacErrorAdapter) ConvertToAssignRoleToUserError(
	err *bizerrors.BusinessError,
) generated.AssignEndUserRoleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserNotFoundInProject.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserInProject}
	case bizerrors.EndUserRoleNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	case bizerrors.EndUserCannotAssignImplicitRole.GetCode():
		return &generated.EndUserCannotAssignImplicitRole{Message: err.Msg()}
	case bizerrors.UserRoleAlreadyAssigned.GetCode():
		return &generated.UserRoleAlreadyAssigned{Message: err.Msg()}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("assignRoleToUser", err)
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	}
}

// ConvertToRevokeRoleFromUserError maps to RevokeEndUserRoleError union.
func (a *RbacErrorAdapter) ConvertToRevokeRoleFromUserError(
	err *bizerrors.BusinessError,
) generated.RevokeEndUserRoleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserNotFoundInProject.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserInProject}
	case bizerrors.EndUserRoleNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	default:
		a.logUnknown("revokeRoleFromUser", err)
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserRole}
	}
}

// ConvertToApplyPresetPolicyError maps to ApplyEndUserPresetPolicyError union.
func (a *RbacErrorAdapter) ConvertToApplyPresetPolicyError(
	err *bizerrors.BusinessError,
) generated.ApplyEndUserPresetPolicyError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.EndUserPresetRequiresOwnerField.GetCode(), bizerrors.EndUserRowScopeFieldMissing.GetCode():
		suggestion := "请先在模型中创建 END_USER_REF 字段，然后重试应用预设"
		preset := detectPresetFromMessage(err.Msg())
		return &generated.PresetRequiresOwnerField{
			Message:    err.Msg(),
			Preset:     preset,
			Suggestion: &suggestion,
		}
	case bizerrors.PresetDeleteBlockedByBundle.GetCode():
		suggestion := "请先解绑引用该预设的权限包，或迁移后再重试"
		return &generated.PresetDeleteBlockedByBundle{
			Message:    err.Msg(),
			Suggestion: &suggestion,
		}
	case bizerrors.ModelNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeModel}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("applyPresetPolicy", err)
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	}
}

// ConvertToRestoreBundleError maps to RestoreEndUserPermissionBundleError union.
func (a *RbacErrorAdapter) ConvertToRestoreBundleError(
	err *bizerrors.BusinessError,
) generated.RestoreEndUserPermissionBundleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionBundleNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.EndUserPermissionBundleSnapshotNotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundleSnapshot,
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("restoreBundle", err)
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	}
}

// ConvertToBindPresetItemToBundleError maps to BindPresetItemToBundleError union.
func (a *RbacErrorAdapter) ConvertToBindPresetItemToBundleError(
	err *bizerrors.BusinessError,
) generated.BindPresetItemToBundleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.ModelNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeModel}
	case bizerrors.EndUserPresetRequiresOwnerField.GetCode(), bizerrors.EndUserRowScopeFieldMissing.GetCode():
		suggestion := "请先在模型中创建 END_USER_REF 字段，然后重试绑定预设"
		preset := detectPresetFromMessage(err.Msg())
		return &generated.PresetRequiresOwnerField{
			Message:    err.Msg(),
			Preset:     preset,
			Suggestion: &suggestion,
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("bindPresetItemToBundle", err)
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToBindCustomItemToBundleError maps to BindCustomItemToBundleError union.
func (a *RbacErrorAdapter) ConvertToBindCustomItemToBundleError(
	err *bizerrors.BusinessError,
) generated.BindCustomItemToBundleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.EndUserPermissionNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUserPermission}
	case bizerrors.ModelNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeModel}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("bindCustomItemToBundle", err)
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToRemoveDataPermissionItemError maps to RemoveDataPermissionItemFromBundleError union.
func (a *RbacErrorAdapter) ConvertToRemoveDataPermissionItemError(
	err *bizerrors.BusinessError,
) generated.RemoveDataPermissionItemFromBundleError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserPermissionBundleNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	case bizerrors.ModelNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeModel}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logUnknown("removeDataPermissionItem", err)
		return &generated.ResourceNotFound{
			Message:      err.Msg(),
			ResourceType: generated.ResourceTypeEndUserPermissionBundle,
		}
	}
}

func detectPresetFromMessage(message string) generated.EndUserPermissionPreset {
	if strings.Contains(message, string(generated.EndUserPermissionPresetReadAllWriteOwner)) {
		return generated.EndUserPermissionPresetReadAllWriteOwner
	}
	if strings.Contains(message, string(generated.EndUserPermissionPresetReadWriteOwner)) {
		return generated.EndUserPermissionPresetReadWriteOwner
	}
	if strings.Contains(message, string(generated.EndUserPermissionPresetReadAll)) {
		return generated.EndUserPermissionPresetReadAll
	}
	if strings.Contains(message, string(generated.EndUserPermissionPresetReadWriteAll)) {
		return generated.EndUserPermissionPresetReadWriteAll
	}
	return generated.EndUserPermissionPresetReadWriteOwner
}
