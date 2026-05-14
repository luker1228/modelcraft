package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCatalogDatabasesUsesProjectEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql/end-user/org/acme/project/sales" {
			t.Fatalf("path = %s, want project catalog endpoint", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"modelDatabaseCatalog":{"data":{"databases":[{"name":"maindb"}]}}}}`))
	}))
	defer srv.Close()

	c := GraphQLClient{HTTPClient: srv.Client()}
	items, err := c.CatalogDatabases(context.Background(), srv.URL, "acme", "sales", "token")
	if err != nil {
		t.Fatalf("CatalogDatabases() error = %v", err)
	}
	if len(items) != 1 || items[0] != "maindb" {
		t.Fatalf("unexpected databases: %+v", items)
	}
}

func TestCatalogModelsUsesDatabaseFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"models":{"items":[{"name":"users"}]}}}`))
	}))
	defer srv.Close()

	c := GraphQLClient{HTTPClient: srv.Client()}
	items, err := c.CatalogModels(context.Background(), srv.URL, "acme", "sales", "maindb", "token")
	if err != nil {
		t.Fatalf("CatalogModels() error = %v", err)
	}
	if len(items) != 1 || items[0].Name != "users" {
		t.Fatalf("unexpected models: %+v", items)
	}
}
