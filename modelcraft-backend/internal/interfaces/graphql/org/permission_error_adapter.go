package orggraphql

import (
	"modelcraft/internal/interfaces/graphql/org/generated"
	"modelcraft/pkg/bizerrors"
)

// convertToCreateCustomRoleError converts bizerrors to GraphQL CreateCustomRoleError
func convertToCreateCustomRoleError(bizErr *bizerrors.BusinessError) generated.CreateCustomRoleError {
	switch {
	case bizErr.Info().IsConflictError():
		return &generated.PermissionRoleAlreadyExists{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Use a different role name or update the existing role"),
		}
	case bizErr.Info().IsParamInvalidError():
		return &generated.PermissionInvalidInput{
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
		return &generated.PermissionRoleNotFound{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Verify the role ID exists"),
		}
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
		return &generated.PermissionInvalidInput{
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
		return &generated.PermissionRoleNotFound{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Verify the role ID exists"),
		}
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
		return &generated.PermissionRoleNotFound{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Verify the role ID exists"),
		}
	case bizErr.Info().IsOperationDeniedError():
		return &generated.PermissionSystemRoleCannotBeModified{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Cannot modify system role permissions"),
		}
	case bizErr.Info().IsParamInvalidError():
		return &generated.PermissionInvalidInput{
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
		if bizErr.Info().GetCode() == "NOT_FOUND.USER" {
			return &generated.PermissionUserNotFound{
				Message:    bizErr.Msg(),
				Suggestion: strPtr("Verify the user ID exists"),
			}
		}
		return &generated.PermissionRoleNotFound{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Verify the role ID exists"),
		}
	case bizErr.Info().IsParamInvalidError():
		return &generated.PermissionInvalidInput{
			Message:    bizErr.Msg(),
			Suggestion: nil,
		}
	}
	return nil
}

// convertToRevokeRoleError converts bizerrors to GraphQL RevokeRoleError
func convertToRevokeRoleError(bizErr *bizerrors.BusinessError) generated.RevokeRoleError {
	if bizErr.Info().IsNotFoundError() {
		if bizErr.Info().GetCode() == "NOT_FOUND.USER" {
			return &generated.PermissionUserNotFound{
				Message:    bizErr.Msg(),
				Suggestion: strPtr("Verify the user ID exists"),
			}
		}
		return &generated.PermissionRoleNotFound{
			Message:    bizErr.Msg(),
			Suggestion: strPtr("Verify the role ID exists"),
		}
	}
	return nil
}
