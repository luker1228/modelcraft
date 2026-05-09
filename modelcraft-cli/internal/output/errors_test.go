package output

import (
	"bytes"
	"errors"
	"testing"
)

func TestWriteSuccessCompactJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := WriteSuccess(buf, "json", true, map[string]any{"version": "v0.1.0"}, nil); err != nil {
		t.Fatalf("WriteSuccess() error = %v", err)
	}

	want := "{\"ok\":true,\"data\":{\"version\":\"v0.1.0\"}}\n"
	if buf.String() != want {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestCLIErrorExitCodeAndEnvelope(t *testing.T) {
	err := NewCLIError(
		"NO_PROJECT_CONTEXT",
		"No project context is selected.",
		true,
		"Use --project <slug> or run 'mc auth switch-project <slug>'.",
		map[string]any{"availableProjects": []string{"sales"}},
	)
	if code := ExitCode(err); code != 5 {
		t.Fatalf("ExitCode() = %d, want 5", code)
	}

	buf := new(bytes.Buffer)
	if writeErr := WriteError(buf, "json", true, err); writeErr != nil {
		t.Fatalf("WriteError() error = %v", writeErr)
	}

	want := "{\"ok\":false,\"error\":{\"code\":\"NO_PROJECT_CONTEXT\",\"message\":\"No project context is selected.\",\"retryable\":true,\"suggestion\":\"Use --project <slug> or run 'mc auth switch-project <slug>'.\",\"details\":{\"availableProjects\":[\"sales\"]}}}\n"
	if buf.String() != want {
		t.Fatalf("unexpected envelope: %s", buf.String())
	}

	if !errors.Is(err, ErrCLI) {
		t.Fatalf("expected CLI sentinel")
	}
}

func TestNonCLIErrorMapsToServerErrorExitCode(t *testing.T) {
	err := errors.New("plain failure")

	if code := ExitCode(err); code != 7 {
		t.Fatalf("ExitCode() = %d, want 7", code)
	}
}

func TestWriteSuccessRejectsUnsupportedFormat(t *testing.T) {
	buf := new(bytes.Buffer)

	err := WriteSuccess(buf, "yaml", true, map[string]any{"version": "v0.1.0"}, nil)
	if err == nil {
		t.Fatal("WriteSuccess() error = nil, want unsupported format error")
	}
}

func TestWriteErrorRejectsUnsupportedFormat(t *testing.T) {
	buf := new(bytes.Buffer)

	err := WriteError(buf, "yaml", true, errors.New("plain failure"))
	if err == nil {
		t.Fatal("WriteError() error = nil, want unsupported format error")
	}
}
