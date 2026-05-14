package schema

import "testing"

func TestBuildCommandSchemaIncludesRunCommand(t *testing.T) {
	doc := BuildCommandSchema()
	if _, ok := doc.Commands["run"]; !ok {
		t.Fatalf("run command missing from schema")
	}
	if _, ok := doc.Commands["describe"]; !ok {
		t.Fatalf("describe command missing from schema")
	}
	for _, removed := range []string{"query", "get", "count", "aggregate"} {
		if _, ok := doc.Commands[removed]; ok {
			t.Fatalf("removed command %q should not be in schema", removed)
		}
	}
}
