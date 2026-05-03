package adapter

import (
	"context"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

// GroupErrorAdapter converts domain errors to GraphQL error types for group operations.
type GroupErrorAdapter struct {
	logger logfacade.Logger
}

// NewGroupErrorAdapter creates a new GroupErrorAdapter.
func NewGroupErrorAdapter(ctx context.Context) *GroupErrorAdapter {
	return &GroupErrorAdapter{
		logger: logfacade.GetLogger(ctx),
	}
}

// ConvertToCreateGroupError converts a business error to a CreateGroupError union type.
func (a *GroupErrorAdapter) ConvertToCreateGroupError(err *bizerrors.BusinessError) generated.CreateGroupError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.GroupAlreadyExists.GetCode():
		suggestion := "Please choose a different group name within this project"
		return &generated.GroupAlreadyExists{
			Message:    err.Msg(),
			Suggestion: &suggestion,
		}
	default:
		return &generated.InvalidGroupName{Message: err.Msg()}
	}
}

// ConvertToRenameGroupError converts a business error to a RenameGroupError union type.
func (a *GroupErrorAdapter) ConvertToRenameGroupError(err *bizerrors.BusinessError) generated.RenameGroupError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.GroupNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeGroup}
	case bizerrors.GroupAlreadyExists.GetCode():
		suggestion := "Please choose a different group name within this project"
		return &generated.GroupAlreadyExists{
			Message:    err.Msg(),
			Suggestion: &suggestion,
		}
	default:
		return &generated.InvalidGroupName{Message: err.Msg()}
	}
}

// ConvertToDeleteGroupError converts a business error to a DeleteGroupError union type.
func (a *GroupErrorAdapter) ConvertToDeleteGroupError(err *bizerrors.BusinessError) generated.DeleteGroupError {
	if err == nil {
		return nil
	}
	return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeGroup}
}

// ConvertToReorderGroupError converts a business error to a ReorderGroupError union type.
func (a *GroupErrorAdapter) ConvertToReorderGroupError(err *bizerrors.BusinessError) generated.ReorderGroupError {
	if err == nil {
		return nil
	}
	return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeGroup}
}

// ConvertToMoveModelToGroupError converts a business error to a MoveModelToGroupError union type.
func (a *GroupErrorAdapter) ConvertToMoveModelToGroupError(
	err *bizerrors.BusinessError,
) generated.MoveModelToGroupError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.ModelNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeModel}
	default:
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeGroup}
	}
}
