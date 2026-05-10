package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQueryBuildsModelScopedRuntimeEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql/end-user/org/acme/project/sales/db/maindb/model/users" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req["query"] == nil {
			t.Fatalf("missing query")
		}
		_, _ = w.Write([]byte(`{"data":{"findMany":[{"id":"1"}]}}`))
	}))
	defer srv.Close()

	c := GraphQLClient{HTTPClient: srv.Client()}
	result, err := c.Query(context.Background(), srv.URL, "acme", "sales", "maindb", "users", "token", QueryOptions{Take: 1})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items len = %d, want 1", len(result.Items))
	}
}
