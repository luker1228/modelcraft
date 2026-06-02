package client

import (
	"context"
	"fmt"
	"strings"

	"modelcraft-cli/internal/config"
)

type CatalogModel struct {
	Name string `json:"name"`
}

func (c GraphQLClient) CatalogProjects(_ context.Context, creds config.Credentials) ([]config.AccessibleProject, error) {
	return creds.Projects, nil
}

func (c GraphQLClient) CatalogDatabases(ctx context.Context, server, org, project, token string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/end-user/graphql/org/%s/project/%s", strings.TrimRight(server, "/"), org, project)
	query := `query CatalogDatabases { modelDatabaseCatalog(input: {}) { data { databases { name } } } }`

	var data struct {
		ModelDatabaseCatalog struct {
			Data struct {
				Databases []struct {
					Name string `json:"name"`
				} `json:"databases"`
			} `json:"data"`
		} `json:"modelDatabaseCatalog"`
	}
	if err := c.Execute(ctx, endpoint, token, query, nil, &data); err != nil {
		return nil, err
	}

	items := make([]string, 0, len(data.ModelDatabaseCatalog.Data.Databases))
	for _, db := range data.ModelDatabaseCatalog.Data.Databases {
		items = append(items, db.Name)
	}
	return items, nil
}

func (c GraphQLClient) CatalogModels(ctx context.Context, server, org, project, database, token string) ([]CatalogModel, error) {
	endpoint := fmt.Sprintf("%s/end-user/graphql/org/%s/project/%s", strings.TrimRight(server, "/"), org, project)
	query := `query CatalogModels($database: String!) { models(input: {databaseName: $database}) { items { name } } }`

	var data struct {
		Models struct {
			Items []CatalogModel `json:"items"`
		} `json:"models"`
	}
	if err := c.Execute(ctx, endpoint, token, query, map[string]any{"database": database}, &data); err != nil {
		return nil, err
	}
	return data.Models.Items, nil
}
