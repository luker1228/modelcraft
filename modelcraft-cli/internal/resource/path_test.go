package resource

import "testing"

func TestParseDatabaseModelUsesCurrentProjectFallback(t *testing.T) {
	got, err := ParseModelPath("maindb.users", ParseContext{CurrentProject: "sales"})
	if err != nil {
		t.Fatalf("ParseModelPath() error = %v", err)
	}
	if got.Project != "sales" || got.Database != "maindb" || got.Model != "users" {
		t.Fatalf("unexpected path: %+v", got)
	}
}

func TestParseModelPathRejectsSingleSegment(t *testing.T) {
	_, err := ParseModelPath("users", ParseContext{CurrentProject: "sales"})
	if err == nil {
		t.Fatalf("expected single-segment rejection")
	}
}
