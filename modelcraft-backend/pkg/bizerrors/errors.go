package bizerrors

import (
	"errors"
	"fmt"
	"io"
	"runtime"
)

var (
	New    = fmt.Errorf
	Errorf = fmt.Errorf
)

type withStack struct {
	err    error
	frames []uintptr
}

func (w *withStack) Error() string  { return w.err.Error() }
func (w *withStack) Unwrap() error  { return w.err }

func (w *withStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprint(s, w.err.Error())
			for _, pc := range w.frames {
				fn := runtime.FuncForPC(pc)
				if fn == nil {
					continue
				}
				file, line := fn.FileLine(pc)
				_, _ = fmt.Fprintf(s, "\n%s\n    %s:%d", fn.Name(), file, line)
			}
			return
		}
		_, _ = fmt.Fprint(s, w.err.Error())
	case 's':
		_, _ = io.WriteString(s, w.err.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", w.err.Error())
	}
}

func (w *withStack) StackTrace() []uintptr {
	return w.frames
}

func callers() []uintptr {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[:n]
}

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withStack{
		err:    fmt.Errorf("%s: %w", message, err),
		frames: callers(),
	}
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &withStack{
		err:    fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err),
		frames: callers(),
	}
}

func WithStack(err error) error {
	if err == nil {
		return nil
	}
	return &withStack{
		err:    err,
		frames: callers(),
	}
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
