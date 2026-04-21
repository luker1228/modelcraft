package adapter

import (
	"context"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

// FKErrorAdapter handles conversion of domain errors to GraphQL errors for FK operations
type FKErrorAdapter struct {
	ctx    context.Context
	logger logfacade.Logger
}

// NewFKErrorAdapter creates a new FK error adapter
func NewFKErrorAdapter(ctx context.Context) *FKErrorAdapter {
	return &FKErrorAdapter{
		ctx:    ctx,
		logger: logfacade.GetLogger(ctx),
	}
}

// ConvertToCreateResult converts a business error to CreateLogicalForeignKeyResult union type.
func (a *FKErrorAdapter) ConvertToCreateResult(err *bizerrors.BusinessError) generated.CreateLogicalForeignKeyResult {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.FKColumnsNotFound.GetCode():
		return &generated.FKColumnsNotFoundError{Message: err.Msg()}
	case bizerrors.FKFieldCountMismatch.GetCode():
		return &generated.FKFieldCountMismatchError{Message: err.Msg()}
	default:
		a.logger.Errorf(a.ctx, "Unknown FK create error code: %s", err.Info().GetCode())
		return &generated.FKColumnsNotFoundError{Message: err.Msg()}
	}
}

// ConvertToDeleteResult converts a business error to DeleteLogicalForeignKeyResult union type.
func (a *FKErrorAdapter) ConvertToDeleteResult(err *bizerrors.BusinessError) generated.DeleteLogicalForeignKeyResult {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.FKNotFound.GetCode():
		return &generated.FKNotFoundError{Message: err.Msg()}
	case bizerrors.FKPairHasRelateFields.GetCode():
		return &generated.FKPairHasRelateFieldsError{Message: err.Msg()}
	case bizerrors.FKNotDeletable.GetCode():
		return &generated.FKNotDeletableError{Message: err.Msg()}
	default:
		a.logger.Errorf(a.ctx, "Unknown FK delete error code: %s", err.Info().GetCode())
		return &generated.FKNotFoundError{Message: err.Msg()}
	}
}
