package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"modelcraft-cli/internal/config"
	"modelcraft-cli/internal/output"
)

type AuthClient struct {
	HTTPClient *http.Client
}

type whoamiResponse struct {
	UserID  string `json:"userId"`
	OrgName string `json:"orgName"`
}

// Whoami calls GET /api/cli/end-user/auth/whoami with a PAT token (mc_pat_xxx)
// and returns the resolved credentials. The PAT is stored as AccessToken.
func (c AuthClient) Whoami(ctx context.Context, server, pat string) (*config.Credentials, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, joinURL(server, "/api/cli/end-user/auth/whoami"), nil)
	if err != nil {
		return nil, output.NewCLIError("INVALID_ARGUMENT", "Failed to build request.", false, "Verify command arguments and retry.", nil)
	}
	req.Header.Set("Authorization", "Bearer "+pat)

	resp, err := c.client().Do(req)
	if err != nil {
		return nil, output.NewCLIError("SERVICE_UNAVAILABLE", "Gateway is unreachable.", true, "Check network connectivity and retry.", nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, decodeUpstreamError(resp)
	}

	var payload whoamiResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, output.NewCLIError("INVALID_UPSTREAM", "Gateway returned invalid JSON.", false, "Inspect the upstream service output.", nil)
	}

	return &config.Credentials{
		Server:      server,
		OrgName:     payload.OrgName,
		UserID:      payload.UserID,
		AccessToken: pat,
	}, nil
}

func (c AuthClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func joinURL(server, path string) string {
	return strings.TrimRight(server, "/") + path
}

func decodeUpstreamError(resp *http.Response) error {
	var payload struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"requestId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return output.NewCLIError("UPSTREAM_ERROR", fmt.Sprintf("Gateway returned status %d.", resp.StatusCode), true, "Inspect gateway logs and retry.", map[string]any{"statusCode": resp.StatusCode})
	}

	code := "UPSTREAM_ERROR"
	message := payload.Message
	retryable := false
	suggestion := "Inspect upstream error details and retry."

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		code = "UNAUTHENTICATED"
		retryable = true
		suggestion = "Run 'mc auth login' and retry."
	case http.StatusForbidden:
		code = "PERMISSION_DENIED"
		suggestion = "Verify your project access and retry."
	case http.StatusNotFound:
		code = "NOT_FOUND"
		suggestion = "Verify the requested resource and retry."
	case http.StatusBadRequest:
		code = "INVALID_ARGUMENT"
		retryable = true
		suggestion = "Check command arguments and retry."
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		code = "SERVICE_UNAVAILABLE"
		retryable = true
		suggestion = "Gateway is unavailable. Retry shortly."
	}
	if message == "" {
		message = fmt.Sprintf("Gateway returned status %d.", resp.StatusCode)
	}

	details := map[string]any{"statusCode": resp.StatusCode}
	if payload.Code != "" {
		details["upstreamCode"] = payload.Code
	}
	if payload.RequestID != "" {
		details["requestId"] = payload.RequestID
	}

	return output.NewCLIError(code, message, retryable, suggestion, details)
}
