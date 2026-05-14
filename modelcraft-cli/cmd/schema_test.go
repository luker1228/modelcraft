package cmd

import (
	"bytes"
	"testing"
)

func TestSchemaCommandsReturnsLocalSchema(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	code := Execute(BuildInfo{}, []string{"schema", "commands"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("Execute() code = %d, stdout=%s", code, stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"commands"`)) {
		t.Fatalf("missing commands payload: %s", stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"run"`)) {
		t.Fatalf("missing run command: %s", stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"describe"`)) {
		t.Fatalf("missing describe command: %s", stdout.String())
	}
}
