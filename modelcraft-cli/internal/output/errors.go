package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var ErrCLI = errors.New("cli error")

type CLIError struct {
	Code       string
	Message    string
	Retryable  bool
	Suggestion string
	Details    map[string]any
}

func (e *CLIError) Error() string { return e.Code + ": " + e.Message }

func (e *CLIError) Unwrap() error { return ErrCLI }

func NewCLIError(code, message string, retryable bool, suggestion string, details map[string]any) *CLIError {
	return &CLIError{
		Code:       code,
		Message:    message,
		Retryable:  retryable,
		Suggestion: suggestion,
		Details:    details,
	}
}

func ExitCode(err error) int {
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		return 7
	}

	switch cliErr.Code {
	case "INVALID_JSON_FLAG", "INVALID_ARGUMENT", "MISSING_REQUIRED_FLAG":
		return 2
	case "UNAUTHENTICATED", "TOKEN_EXPIRED", "INVALID_CREDENTIALS":
		return 3
	case "PERMISSION_DENIED":
		return 4
	case "NO_PROJECT_CONTEXT", "MODEL_NOT_FOUND", "DATABASE_NOT_FOUND", "NOT_FOUND", "PROJECT_NOT_FOUND":
		return 5
	case "TAKE_EXCEEDS_LIMIT", "INVALID_RESOURCE_PATH":
		return 6
	default:
		return 7
	}
}

func WriteError(w io.Writer, format string, compact bool, err error) error {
	if err := validateFormat(format); err != nil {
		return err
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		cliErr = NewCLIError(
			"UNKNOWN_ERROR",
			fmt.Sprintf("%v", err),
			false,
			"Inspect stderr or rerun with --verbose.",
			nil,
		)
	}

	type errorEnvelope struct {
		Code       string         `json:"code"`
		Message    string         `json:"message"`
		Retryable  bool           `json:"retryable"`
		Suggestion string         `json:"suggestion"`
		Details    map[string]any `json:"details,omitempty"`
	}

	payload := struct {
		OK    bool          `json:"ok"`
		Error errorEnvelope `json:"error"`
	}{
		OK: false,
		Error: errorEnvelope{
			Code:       cliErr.Code,
			Message:    cliErr.Message,
			Retryable:  cliErr.Retryable,
			Suggestion: cliErr.Suggestion,
			Details:    cliErr.Details,
		},
	}
	return writeJSON(w, compact, payload)
}

func validateFormat(format string) error {
	if format != "json" {
		return fmt.Errorf("unsupported output format %q", format)
	}

	return nil
}

func writeJSON(w io.Writer, compact bool, payload any) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if !compact {
		encoder.SetIndent("", "  ")
	} else {
		encoder.SetIndent("", "")
	}
	return encoder.Encode(payload)
}
