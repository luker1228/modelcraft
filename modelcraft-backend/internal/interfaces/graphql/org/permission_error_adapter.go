package orggraphql

import (
	"modelcraft/internal/interfaces/graphql/org/generated"
	"modelcraft/pkg/bizerrors"
)

func newOrgResourceNotFound(message string, resourceType generated.ResourceType) *generated.ResourceNotFound {
	return &generated.ResourceNotFound{
		Message:      message,
		ResourceType: resourceType,
	}
}

// convertToCreateCustomRoleError converts bizerrors to GraphQL CreateCustomRoleError
func convertToCreateCustomRoleError(bizErr *bizerrors.BusinessError) generated.CreateCustomRoleError {
	switch {
	case bizErr.Info().IsConflictError():
		return &generated.PermissionRoleAlreadyExists{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Use a different role name or update the existing role"),
		}
	case bizErr.Info().IsParamInvalidError():
		return &generated.InvalidInput{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Check role name format and org_name validity"),
		}
	}
	return nil
}

// convertToUpdateRoleError converts bizerrors to GraphQL UpdatePermissionRoleError
func convertToUpdateRoleError(bizErr *bizerrors.BusinessError) generated.UpdatePermissionRoleError {
	switch {
	case bizErr.Info().IsNotFoundError():
		return newOrgResourceNotFound(bizErr.Msg(), generated.ResourceTypePermissionRole)
	case bizErr.Info().IsOperationDeniedError():
		return &generated.PermissionSystemRoleCannotBeModified{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("System roles (owner, admin, editor, viewer) cannot be modified"),
		}
	case bizErr.Info().IsConflictError():
		return &generated.PermissionRoleAlreadyExists{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("A role with this name already exists in the organization"),
		}
	case bizErr.Info().IsParamInvalidError():
		return &generated.InvalidInput{
			Message:    bizErr.Msg(),
			Suggestion: nil,
		}
	}
	return nil
}

// convertToDeleteRoleError converts bizerrors to GraphQL DeletePermissionRoleError
func convertToDeleteRoleError(bizErr *bizerrors.BusinessError) generated.DeletePermissionRoleError {
	switch {
	case bizErr.Info().IsNotFoundError():
		return newOrgResourceNotFound(bizErr.Msg(), generated.ResourceTypePermissionRole)
	case bizErr.Info().IsOperationDeniedError():
		return &generated.PermissionSystemRoleCannotBeModified{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("System roles cannot be deleted"),
		}
	}
	return nil
}

// convertToRolePermissionError converts bizerrors to GraphQL RolePermissionError
func convertToRolePermissionError(bizErr *bizerrors.BusinessError) generated.RolePermissionError {
	switch {
	case bizErr.Info().IsNotFoundError():
		return newOrgResourceNotFound(bizErr.Msg(), generated.ResourceTypePermissionRole)
	case bizErr.Info().IsOperationDeniedError():
		return &generated.PermissionSystemRoleCannotBeModified{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Cannot modify system role permissions"),
		}
	case bizErr.Info().IsParamInvalidError():
		return &generated.InvalidInput{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Check permission format (e.g., 'resource:action')"),
		}
	}
	return nil
}

// convertToAssignRoleError converts bizerrors to GraphQL AssignRoleError
func convertToAssignRoleError(bizErr *bizerrors.BusinessError) generated.AssignRoleError {
	switch {
	case bizErr.Info().IsNotFoundError():
		if bizErr.Info().GetCode() == bizerrors.UserNotFound.GetCode() {
			return newOrgResourceNotFound(bizErr.Msg(), generated.ResourceTypePermissionUser)
		}
		return newOrgResourceNotFound(bizErr.Msg(), generated.ResourceTypePermissionRole)
	case bizErr.Info().IsParamInvalidError():
		return &generated.InvalidInput{
			Message:    bizErr.Msg(),
			Suggestion: nil,
		}
	}
	return nil
}

// convertToRevokeRoleError converts bizerrors to GraphQL RevokeRoleError
func convertToRevokeRoleError(bizErr *bizerrors.BusinessError) generated.RevokeRoleError {
	if bizErr.Info().IsNotFoundError() {
		if bizErr.Info().GetCode() == bizerrors.UserNotFound.GetCode() {
			return newOrgResourceNotFound(bizErr.Msg(), generated.ResourceTypePermissionUser)
		}
		return newOrgResourceNotFound(bizErr.Msg(), generated.ResourceTypePermissionRole)
	}
	return nil
}
