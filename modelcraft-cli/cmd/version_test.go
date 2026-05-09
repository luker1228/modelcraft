package cmd

import (
	"bytes"
	"testing"
)

func TestVersionCommandPrintsInjectedMetadata(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	cmd := NewRootCommand(BuildInfo{
		Version:   "v0.1.0",
		Commit:    "abc1234",
		BuildTime: "2026-05-09T12:00:00Z",
	})

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := buf.String()
	want := "{\"ok\":true,\"data\":{\"version\":\"v0.1.0\",\"commit\":\"abc1234\",\"buildTime\":\"2026-05-09T12:00:00Z\"}}\n"
	if got != want {
		t.Fatalf("version output mismatch\nwant: %s\ngot:  %s", want, got)
	}
}
