package projectgraphql

import (
	"modelcraft/internal/interfaces/graphql/project/adapter"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
)

func convertToAddFieldsError(err error, errorAdapter *adapter.ModelErrorAdapter) generated.AddFieldsError {
	if err == nil {
		return nil
	}
	if bizErr, ok := err.(*bizerrors.BusinessError); ok {
		return errorAdapter.ConvertToAddFieldsError(bizErr)
	}
	msg := err.Error()
	return &generated.InvalidInput{Message: msg}
}
