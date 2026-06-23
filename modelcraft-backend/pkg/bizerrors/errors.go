package bizerrors

import (
	"errors"
	"fmt"
)

var (
	New    = fmt.Errorf
	Errorf = fmt.Errorf
)

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

func WithStack(err error) error {
	return err
}

func Cause(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

var (
	Is      = errors.Is
	As      = errors.As
	Unwrap  = errors.Unwrap
)
