package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"modelcraft-cli/internal/output"
)

type GraphQLClient struct {
	HTTPClient *http.Client
}

type graphQLRequest struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName string         `json:"operationName,omitempty"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string         `json:"message"`
		Path    []any          `json:"path"`
		Ext     map[string]any `json:"extensions"`
	} `json:"errors"`
}

// extractOperation parses the query string and returns the operation type
// ("query", "mutation", "subscription") and the named operation name.
// Returns ("", "") for anonymous operations.
func extractOperation(query string) (opType, opName string) {
	re := regexp.MustCompile(`(?i)(query|mutation|subscription)\s+(\w+)`)
	m := re.FindStringSubmatch(query)
	if len(m) >= 3 {
		return strings.ToLower(m[1]), m[2]
	}
	return "", ""
}

func (c GraphQLClient) Execute(ctx context.Context, endpoint, token, query string, variables map[string]any, out any) error {
	opType, opName := extractOperation(query)

	reqBody := graphQLRequest{Query: query, Variables: variables}
	if opName != "" {
		reqBody.OperationName = opName
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return output.NewCLIError("INVALID_ARGUMENT", "Failed to serialize GraphQL request.", false, "Check command arguments and retry.", nil)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return output.NewCLIError("INVALID_ARGUMENT", "Failed to build GraphQL request.", false, "Check command arguments and retry.", nil)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	// Set X-Action header when the query has a named operation — required by the
	// backend's ChiGraphQLActionMiddleware on project-level GraphQL routes.
	if opType != "" && opName != "" {
		req.Header.Set("X-Action", opType+":"+opName)
	}

	resp, err := c.client().Do(req)
	if err != nil {
		return output.NewCLIError("SERVICE_UNAVAILABLE", "Gateway is unreachable.", true, "Check network connectivity and retry.", nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return decodeUpstreamError(resp)
	}

	var payload graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return output.NewCLIError("INVALID_UPSTREAM", "Gateway returned invalid JSON.", false, "Inspect the upstream service output.", nil)
	}
	if len(payload.Errors) > 0 {
		message := payload.Errors[0].Message
		code := "GRAPHQL_ERROR"
		if extCode, ok := payload.Errors[0].Ext["code"].(string); ok && extCode != "" {
			code = strings.ToUpper(strings.ReplaceAll(extCode, "-", "_"))
		}
		if message == "" {
			message = "GraphQL request failed."
		}
		return output.NewCLIError(code, message, false, "Inspect GraphQL error details and retry.", map[string]any{"path": payload.Errors[0].Path})
	}
	if out == nil {
		return nil
	}
	if len(payload.Data) == 0 {
		return output.NewCLIError("INVALID_UPSTREAM", "GraphQL response is missing data.", false, "Inspect the upstream service output.", nil)
	}
	if err := json.Unmarshal(payload.Data, out); err != nil {
		return output.NewCLIError("INVALID_UPSTREAM", fmt.Sprintf("Failed to decode GraphQL data: %v", err), false, "Inspect the upstream service output.", nil)
	}

	return nil
}

func (c GraphQLClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

