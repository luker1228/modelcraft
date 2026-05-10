package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type QueryOptions struct {
	Where   json.RawMessage
	Select  []string
	OrderBy json.RawMessage
	Take    int
	Skip    int
}

type QueryResult struct {
	Items []map[string]any `json:"items"`
}

type GetOptions struct {
	Where  json.RawMessage
	Select []string
}

type CountOptions struct {
	Where json.RawMessage
}

type AggregateOptions struct {
	Where  json.RawMessage
	Fields []string
}

func (c GraphQLClient) Query(ctx context.Context, server, org, project, db, model, token string, opts QueryOptions) (QueryResult, error) {
	endpoint := modelEndpoint(server, org, project, db, model)
	query := `query FindMany($where: JSON, $select: [String!], $orderBy: JSON, $take: Int, $skip: Int) { findMany(where: $where, select: $select, orderBy: $orderBy, take: $take, skip: $skip) }`
	variables := map[string]any{
		"where":   opts.Where,
		"select":  opts.Select,
		"orderBy": opts.OrderBy,
		"take":    opts.Take,
		"skip":    opts.Skip,
	}

	var data struct {
		FindMany []map[string]any `json:"findMany"`
	}
	if err := c.Execute(ctx, endpoint, token, query, variables, &data); err != nil {
		return QueryResult{}, err
	}

	return QueryResult{Items: data.FindMany}, nil
}

func (c GraphQLClient) Get(ctx context.Context, server, org, project, db, model, token string, opts GetOptions) (map[string]any, error) {
	endpoint := modelEndpoint(server, org, project, db, model)
	query := `query FindUnique($where: JSON!, $select: [String!]) { findUnique(where: $where, select: $select) }`
	variables := map[string]any{"where": opts.Where, "select": opts.Select}

	var data struct {
		FindUnique map[string]any `json:"findUnique"`
	}
	if err := c.Execute(ctx, endpoint, token, query, variables, &data); err != nil {
		return nil, err
	}
	return data.FindUnique, nil
}

func (c GraphQLClient) Count(ctx context.Context, server, org, project, db, model, token string, opts CountOptions) (int, error) {
	endpoint := modelEndpoint(server, org, project, db, model)
	query := `query Count($where: JSON) { count(where: $where) }`
	variables := map[string]any{"where": opts.Where}

	var data struct {
		Count int `json:"count"`
	}
	if err := c.Execute(ctx, endpoint, token, query, variables, &data); err != nil {
		return 0, err
	}
	return data.Count, nil
}

func (c GraphQLClient) Aggregate(ctx context.Context, server, org, project, db, model, token string, opts AggregateOptions) (map[string]any, error) {
	endpoint := modelEndpoint(server, org, project, db, model)
	query := `query Aggregate($where: JSON, $fields: [String!]) { aggregate(where: $where, fields: $fields) }`
	variables := map[string]any{"where": opts.Where, "fields": opts.Fields}

	var data struct {
		Aggregate map[string]any `json:"aggregate"`
	}
	if err := c.Execute(ctx, endpoint, token, query, variables, &data); err != nil {
		return nil, err
	}
	return data.Aggregate, nil
}

func modelEndpoint(server, org, project, db, model string) string {
	return fmt.Sprintf("%s/graphql/end-user/org/%s/project/%s/db/%s/model/%s", strings.TrimRight(server, "/"), org, project, db, model)
}
