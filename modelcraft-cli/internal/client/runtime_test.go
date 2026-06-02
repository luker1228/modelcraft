package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunBuildsModelScopedEndpointAndPassesQueryThrough(t *testing.T) {
	const gqlBody = `{ findMany(take: 3) { id name } }`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/end-user/graphql/org/acme/project/sales/db/maindb/model/users" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "findMany") {
			t.Fatalf("body missing query: %s", string(body))
		}
		_, _ = w.Write([]byte(`{"data":{"findMany":[{"id":"1"},{"id":"2"},{"id":"3"}]}}`))
	}))
	defer srv.Close()

	c := GraphQLClient{HTTPClient: srv.Client()}
	result, err := c.Run(context.Background(), srv.URL, "acme", "sales", "maindb", "users", "token", gqlBody)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	items, _ := result["findMany"].([]any)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
}

func TestRunReturnsGraphQLErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"some error","path":["findMany"]}],"data":null}`))
	}))
	defer srv.Close()

	c := GraphQLClient{HTTPClient: srv.Client()}
	_, err := c.Run(context.Background(), srv.URL, "acme", "sales", "maindb", "users", "token", `{ findMany { id } }`)
	if err == nil {
		t.Fatal("expected error from GraphQL errors, got nil")
	}
}
