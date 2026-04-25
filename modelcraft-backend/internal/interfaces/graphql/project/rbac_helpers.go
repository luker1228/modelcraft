package projectgraphql

import (
	"errors"
	"modelcraft/pkg/bizerrors"
)

func toBizErr(err error) *bizerrors.BusinessError {
	var be *bizerrors.BusinessError
	if errors.As(err, &be) {
		return be
	}
	return bizerrors.NewError(bizerrors.SystemError, err.Error())
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
