package client

import (
	"bytes"
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

type loginRequest struct {
	OrgName  string `json:"orgName"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	OrgName      string `json:"orgName"`
	RefreshToken string `json:"refreshToken"`
}

type logoutRequest struct {
	OrgName      string `json:"orgName"`
	RefreshToken string `json:"refreshToken"`
}

type selectProjectRequest struct {
	OrgName      string `json:"orgName"`
	ProjectSlug  string `json:"projectSlug"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type meResponse struct {
	EndUser map[string]any `json:"endUser"`
}

func (c AuthClient) Login(ctx context.Context, server, org, username, password string) (*config.Credentials, error) {
	var creds config.Credentials
	err := c.postJSON(ctx, joinURL(server, "/api/cli/end-user/auth/login"), loginRequest{
		OrgName:  org,
		Username: username,
		Password: password,
	}, &creds)
	if err != nil {
		return nil, err
	}

	creds.Server = server
	creds.OrgName = org
	return &creds, nil
}

func (c AuthClient) Refresh(ctx context.Context, server, orgName, refreshToken string) (*config.Credentials, error) {
	var creds config.Credentials
	err := c.postJSON(ctx, joinURL(server, "/api/cli/end-user/auth/refresh"), refreshRequest{
		OrgName:      orgName,
		RefreshToken: refreshToken,
	}, &creds)
	if err != nil {
		return nil, err
	}
	creds.Server = server
	creds.OrgName = orgName
	return &creds, nil
}

func (c AuthClient) Logout(ctx context.Context, server, orgName, refreshToken string) error {
	return c.postJSON(ctx, joinURL(server, "/api/cli/end-user/auth/logout"), logoutRequest{
		OrgName:      orgName,
		RefreshToken: refreshToken,
	}, nil)
}

func (c AuthClient) SelectProject(ctx context.Context, server, orgName, projectSlug, refreshToken string) error {
	return c.postJSON(ctx, joinURL(server, "/api/cli/end-user/auth/select-project"), selectProjectRequest{
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		RefreshToken: refreshToken,
	}, nil)
}

func (c AuthClient) Me(ctx context.Context, server, accessToken string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, joinURL(server, "/api/cli/end-user/auth/me"), nil)
	if err != nil {
		return nil, output.NewCLIError("INVALID_ARGUMENT", "Failed to build request.", false, "Verify command arguments and retry.", nil)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.client().Do(req)
	if err != nil {
		return nil, output.NewCLIError("SERVICE_UNAVAILABLE", "Gateway is unreachable.", true, "Check network connectivity and retry.", nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, decodeUpstreamError(resp)
	}

	var payload meResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, output.NewCLIError("INVALID_UPSTREAM", "Gateway returned invalid JSON.", false, "Inspect the upstream service output.", nil)
	}
	return payload.EndUser, nil
}

func (c AuthClient) postJSON(ctx context.Context, url string, reqBody any, out any) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return output.NewCLIError("INVALID_ARGUMENT", "Failed to serialize request payload.", false, "Verify command arguments and retry.", nil)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return output.NewCLIError("INVALID_ARGUMENT", "Failed to build request.", false, "Verify command arguments and retry.", nil)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client().Do(req)
	if err != nil {
		return output.NewCLIError("SERVICE_UNAVAILABLE", "Gateway is unreachable.", true, "Check network connectivity and retry.", nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return decodeUpstreamError(resp)
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return output.NewCLIError("INVALID_UPSTREAM", "Gateway returned invalid JSON.", false, "Inspect the upstream service output.", nil)
	}
	return nil
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
