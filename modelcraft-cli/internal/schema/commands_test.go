package schema

import "testing"

func TestBuildCommandSchemaIncludesQueryFlags(t *testing.T) {
	doc := BuildCommandSchema()
	query, ok := doc.Commands["query"]
	if !ok {
		t.Fatalf("query command missing")
	}
	if _, ok := query.Flags["take"]; !ok {
		t.Fatalf("query --take flag missing")
	}
}

