package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appapitoken "modelcraft/internal/app/apitoken"
	appauth "modelcraft/internal/app/auth"
	domainenduser "modelcraft/internal/domain/enduser"
	httpmiddleware "modelcraft/internal/interfaces/http/middleware"
)

type stubAPITokenRepo struct {
	token         *domainenduser.APIToken
	lastUpdatedID string
}

func (s *stubAPITokenRepo) Save(ctx context.Context, token *domainenduser.APIToken) error {
	return nil
}

func (s *stubAPITokenRepo) FindByHash(ctx context.Context, hash string) (*domainenduser.APIToken, error) {
	if s.token != nil && s.token.TokenHash == hash {
		return s.token, nil
	}
	return nil, nil
}

func (s *stubAPITokenRepo) ListByUser(
	ctx context.Context, orgName, endUserID string,
) ([]*domainenduser.APIToken, error) {
	return nil, nil
}

func (s *stubAPITokenRepo) SoftDelete(ctx context.Context, id, orgName, endUserID string) error {
	return nil
}

func (s *stubAPITokenRepo) UpdateLastUsed(ctx context.Context, id string, at time.Time) error {
	s.lastUpdatedID = id
	return nil
}

func TestHandlePATWhoamiReturnsIdentity(t *testing.T) {
	raw := "plain-token"
	repo := &stubAPITokenRepo{
		token: &domainenduser.APIToken{
			ID:        "token-1",
			OrgName:   "acme",
			EndUserID: "user-1",
			TokenHash: appauth.HashToken(raw),
			CreatedAt: time.Now(),
		},
	}
	h := &Handler{
		apiTokenSvc: appapitoken.NewAPITokenService(repo),
		isOrgAdminFn: func(ctx context.Context, orgName, userID string) (bool, error) {
			if orgName != "acme" {
				t.Fatalf("unexpected orgName: %s", orgName)
			}
			if userID != "user-1" {
				t.Fatalf("unexpected userID: %s", userID)
			}
			return true, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/tenant/auth/whoami", nil)
	req.Header.Set("Authorization", "Bearer mc_pat_"+raw)

	rr := httptest.NewRecorder()
	h.HandlePATWhoami(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got := body["requestId"]; got != "req-1" {
		t.Fatalf("expected requestId req-1, got %#v", got)
	}
	if got := body["userId"]; got != "user-1" {
		t.Fatalf("expected userId user-1, got %#v", got)
	}
	if got := body["orgName"]; got != "acme" {
		t.Fatalf("expected orgName acme, got %#v", got)
	}
	if got := body["isAdmin"]; got != true {
		t.Fatalf("expected isAdmin true, got %#v", got)
	}
}

func TestHandlePATWhoamiRejectsInvalidToken(t *testing.T) {
	h := &Handler{
		apiTokenSvc: appapitoken.NewAPITokenService(&stubAPITokenRepo{}),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/tenant/auth/whoami", nil)
	req.Header.Set("Authorization", "Bearer mc_pat_invalid")

	rr := httptest.NewRecorder()
	h.HandlePATWhoami(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	errPayload, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error payload, got %#v", body["error"])
	}
	if got := errPayload["code"]; got != "UNAUTHENTICATED" {
		t.Fatalf("expected error code UNAUTHENTICATED, got %#v", got)
	}
}

var _ httpmiddleware.IsOrgAdminFn = func(ctx context.Context, orgName, userID string) (bool, error) {
	return false, nil
}
