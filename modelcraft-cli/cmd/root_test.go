package cmd

import (
	"bytes"
	"errors"
	"testing"

	"modelcraft-cli/internal/output"

	"github.com/spf13/cobra"
)

func TestExecuteWritesErrorEnvelopeAndExitCodeForArgumentFailure(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	code := Execute(BuildInfo{}, []string{"version", "extra"}, stdout, stderr)

	if code != 2 {
		t.Fatalf("Execute() exit code = %d, want 2", code)
	}

	wantStdout := "{\"ok\":false,\"error\":{\"code\":\"INVALID_ARGUMENT\",\"message\":\"Unexpected argument or subcommand.\",\"retryable\":true,\"suggestion\":\"Run 'mc version --help' to inspect valid arguments and flags.\",\"details\":{\"parserError\":\"unknown command \\\"extra\\\" for \\\"mc version\\\"\"}}}\n"
	if stdout.String() != wantStdout {
		t.Fatalf("stdout mismatch\nwant: %s\ngot:  %s", wantStdout, stdout.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestExecuteWritesInvalidArgumentEnvelopeForUnknownFlag(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	code := Execute(BuildInfo{}, []string{"version", "--bad"}, stdout, stderr)

	if code != 2 {
		t.Fatalf("Execute() exit code = %d, want 2", code)
	}

	wantStdout := "{\"ok\":false,\"error\":{\"code\":\"INVALID_ARGUMENT\",\"message\":\"Invalid flag or flag value.\",\"retryable\":true,\"suggestion\":\"Run 'mc version --help' to inspect valid arguments and flags.\",\"details\":{\"parserError\":\"unknown flag: --bad\"}}}\n"
	if stdout.String() != wantStdout {
		t.Fatalf("stdout mismatch\nwant: %s\ngot:  %s", wantStdout, stdout.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestExecuteWritesMissingRequiredFlagEnvelope(t *testing.T) {
	root := &cobra.Command{Use: "mc", SilenceUsage: true, SilenceErrors: true}
	createCmd := &cobra.Command{
		Use:  "create",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	createCmd.Flags().String("name", "", "resource name")
	if err := createCmd.MarkFlagRequired("name"); err != nil {
		t.Fatalf("MarkFlagRequired() error = %v", err)
	}
	root.AddCommand(createCmd)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	code := executeCommand(root, []string{"create"}, stdout, stderr)

	if code != 2 {
		t.Fatalf("executeCommand() exit code = %d, want 2", code)
	}

	wantStdout := "{\"ok\":false,\"error\":{\"code\":\"MISSING_REQUIRED_FLAG\",\"message\":\"Missing required flag.\",\"retryable\":true,\"suggestion\":\"Run 'mc create --help' to inspect valid arguments and flags.\",\"details\":{\"parserError\":\"required flag(s) \\\"name\\\" not set\"}}}\n"
	if stdout.String() != wantStdout {
		t.Fatalf("stdout mismatch\nwant: %s\ngot:  %s", wantStdout, stdout.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestNormalizeErrorClassifiesUnknownFlag(t *testing.T) {
	err := normalizeError("mc version", errors.New("unknown flag: --bad"))

	got, ok := err.(*output.CLIError)
	if !ok {
		t.Fatalf("normalizeError() type = %T, want *output.CLIError", err)
	}

	if got.Code != "INVALID_ARGUMENT" {
		t.Fatalf("Code = %q, want INVALID_ARGUMENT", got.Code)
	}
	if got.Message != "Invalid flag or flag value." {
		t.Fatalf("Message = %q, want stable message", got.Message)
	}
	if got.Details["parserError"] != "unknown flag: --bad" {
		t.Fatalf("parserError = %v, want raw parser text", got.Details["parserError"])
	}
}

func TestNormalizeErrorClassifiesInvalidFlagValue(t *testing.T) {
	err := normalizeError("mc version", errors.New("invalid argument \"oops\" for \"--take\" flag: strconv.Atoi: parsing \"oops\": invalid syntax"))

	got, ok := err.(*output.CLIError)
	if !ok {
		t.Fatalf("normalizeError() type = %T, want *output.CLIError", err)
	}

	if got.Code != "INVALID_ARGUMENT" {
		t.Fatalf("Code = %q, want INVALID_ARGUMENT", got.Code)
	}
	if got.Message != "Invalid flag or flag value." {
		t.Fatalf("Message = %q, want stable message", got.Message)
	}
	if got.Details["parserError"] == nil {
		t.Fatal("expected parserError detail")
	}
}

func TestExecuteCommandWritesUnknownErrorEnvelopeForNonCLIError(t *testing.T) {
	root := &cobra.Command{
		Use:           "mc",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("plain failure")
		},
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	code := executeCommand(root, nil, stdout, stderr)

	if code != 7 {
		t.Fatalf("executeCommand() exit code = %d, want 7", code)
	}

	wantStdout := "{\"ok\":false,\"error\":{\"code\":\"UNKNOWN_ERROR\",\"message\":\"plain failure\",\"retryable\":false,\"suggestion\":\"Inspect stderr or rerun with --verbose.\"}}\n"
	if stdout.String() != wantStdout {
		t.Fatalf("stdout mismatch\nwant: %s\ngot:  %s", wantStdout, stdout.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}
