package bizerrors

import (
	"github.com/graphql-go/graphql/gqlerrors"
)

// NewGraphqlErr 创建一个 graphql 错误
func NewGraphqlErr(err error, code string) *gqlerrors.FormattedError {
	return &gqlerrors.FormattedError{
		Message:   err.Error(),
		Locations: nil,
		Path:      nil,
		Extensions: map[string]any{
			"code": code,
		},
	}
}
