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
	endpoint := fmt.Sprintf("%s/graphql/end-user/org/%s/project/%s", strings.TrimRight(server, "/"), org, project)
	query := `query CatalogDatabases { modelDatabaseCatalog(input: {}) { items { name } } }`

	var data struct {
		ModelDatabaseCatalog struct {
			Items []struct {
				Name string `json:"name"`
			} `json:"items"`
		} `json:"modelDatabaseCatalog"`
	}
	if err := c.Execute(ctx, endpoint, token, query, nil, &data); err != nil {
		return nil, err
	}

	items := make([]string, 0, len(data.ModelDatabaseCatalog.Items))
	for _, item := range data.ModelDatabaseCatalog.Items {
		items = append(items, item.Name)
	}
	return items, nil
}

func (c GraphQLClient) CatalogModels(ctx context.Context, server, org, project, database, token string) ([]CatalogModel, error) {
	endpoint := fmt.Sprintf("%s/graphql/end-user/org/%s/project/%s", strings.TrimRight(server, "/"), org, project)
	query := `query CatalogModels($database: String!) { modelCatalog(input: {databaseName: $database}) { items { name } } }`

	var data struct {
		ModelCatalog struct {
			Items []CatalogModel `json:"items"`
		} `json:"modelCatalog"`
	}
	if err := c.Execute(ctx, endpoint, token, query, map[string]any{"database": database}, &data); err != nil {
		return nil, err
	}
	return data.ModelCatalog.Items, nil
}
