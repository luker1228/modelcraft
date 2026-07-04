package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"modelcraft/pkg/httpheader"
	"net/http"
	"regexp"
	"strings"
)

const (
	errCodeActionRequired = "ACTION_REQUIRED"
	errCodeActionInvalid  = "ACTION_INVALID"
	errCodeActionMismatch = "ACTION_MISMATCH"
)

type gqlRequestBody struct {
	Query         string `json:"query"`
	OperationName string `json:"operationName"`
}

// ChiGraphQLActionMiddleware validates the X-Action header for GraphQL POST requests.
//
// Header format: "{type}:{operationName}", e.g. "query:GetProjects", "mutation:CreateModel".
//
// Validation rules (in order):
//  1. X-Action header must be present.
//  2. Format must be {type}:{operationName} where type is query/mutation/subscription.
//  3. GraphQL body must contain a non-empty operationName.
//  4. operationName in body must equal X-Action operationName.
//  5. The operation type declared in the query string must equal X-Action type.
//
// GET requests (GraphQL Playground) are skipped without validation.
func ChiGraphQLActionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			if errCode, errMsg := validateGraphQLAction(r); errCode != "" {
				writeJSONError(w, http.StatusBadRequest, errMsg, errCode)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// validateGraphQLAction performs all X-Action validation and body restoration.
// Returns (errorCode, errorMessage) on failure, ("", "") on success.
func validateGraphQLAction(r *http.Request) (string, string) {
	xAction := r.Header.Get(httpheader.XAction)
	if xAction == "" {
		return errCodeActionRequired, "X-Action header is required"
	}

	headerType, headerName, ok := parseXAction(xAction)
	if !ok {
		return errCodeActionInvalid, `X-Action must be "{type}:{operationName}"`
	}

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return errCodeActionInvalid, "failed to read request body"
	}
	r.Body = io.NopCloser(bytes.NewReader(rawBody))

	var gqlBody gqlRequestBody
	if err := json.Unmarshal(rawBody, &gqlBody); err != nil {
		return errCodeActionInvalid, "invalid GraphQL JSON body"
	}

	if gqlBody.OperationName == "" {
		return errCodeActionRequired, "GraphQL operationName is required"
	}
	if gqlBody.OperationName != headerName {
		return errCodeActionMismatch, "X-Action operationName does not match GraphQL operationName"
	}

	opType := extractGraphQLOpType(gqlBody.Query, gqlBody.OperationName)
	if opType == "" {
		return errCodeActionInvalid, "could not find operation type in GraphQL query"
	}
	if opType != headerType {
		return errCodeActionMismatch, "X-Action type does not match GraphQL operation type"
	}

	return "", ""
}

// parseXAction parses "type:operationName" and validates the type.
// Returns (type, operationName, ok).
func parseXAction(xAction string) (string, string, bool) {
	parts := strings.SplitN(xAction, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	opType := strings.ToLower(parts[0])
	switch opType {
	case "query", "mutation", "subscription":
		return opType, parts[1], true
	default:
		return "", "", false
	}
}

// extractGraphQLOpType finds the operation type (query/mutation/subscription)
// for the named operation in the query string.
// Returns lowercase type string, or "" if not found.
func extractGraphQLOpType(query, operationName string) string {
	pattern := `(?i)(query|mutation|subscription)\s+` + regexp.QuoteMeta(operationName) + `[\s({]`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	m := re.FindStringSubmatch(query)
	if len(m) >= 2 {
		return strings.ToLower(m[1])
	}
	return ""
}
