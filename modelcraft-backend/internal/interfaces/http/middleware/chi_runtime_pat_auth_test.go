package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	appEnduser "modelcraft/internal/app/enduser"
	domainEnduser "modelcraft/internal/domain/enduser"
	"modelcraft/pkg/ctxutils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// stubTokenRepo is an in-memory implementation of domainEnduser.APITokenRepository for tests.
type stubTokenRepo struct {
	// byHash maps token hash → stored token (nil = not found / error)
	byHash map[string]*domainEnduser.APIToken
	errOn  string // hash that triggers a lookup error
}

func (s *stubTokenRepo) Save(_ context.Context, _ *domainEnduser.APIToken) error {
	return nil
}

func (s *stubTokenRepo) FindByHash(_ context.Context, hash string) (*domainEnduser.APIToken, error) {
	if s.errOn != "" && hash == s.errOn {
		return nil, fmt.Errorf("lookup error")
	}
	tok, ok := s.byHash[hash]
	if !ok {
		return nil, nil //nolint:nilnil
	}
	return tok, nil
}

func (s *stubTokenRepo) ListByUser(_ context.Context, _, _ string) ([]*domainEnduser.APIToken, error) {
	return nil, nil
}

func (s *stubTokenRepo) SoftDelete(_ context.Context, _, _, _ string) error {
	return nil
}

func (s *stubTokenRepo) UpdateLastUsed(_ context.Context, _ string, _ time.Time) error {
	return nil
}

// hashOf computes SHA-256 hex of s (mirrors auth.HashToken).
func hashOf(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// buildAPITokenServiceFromStub wires the stub repo into the real APITokenService.
func buildAPITokenServiceFromStub(repo domainEnduser.APITokenRepository) *appEnduser.APITokenService {
	return appEnduser.NewAPITokenService(repo)
}

func TestChiRuntimePATMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		repo           *stubTokenRepo
		wantStatus     int
		wantNextCalled bool
		wantEndUserID  string // empty means no identity expected
		wantCtxUserID  string // expected ctxutils UserID (only set for valid-token case)
		wantCtxOrgName string // expected ctxutils OrgName (only set for valid-token case)
	}{
		{
			name:           "non-PAT Bearer passes through unchanged",
			authHeader:     "Bearer eyJhbGciOiJFUzI1NiJ9.payload.sig",
			repo:           &stubTokenRepo{byHash: map[string]*domainEnduser.APIToken{}},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantEndUserID:  "",
		},
		{
			name:       "mc_pat_ token invalid → 401",
			authHeader: "Bearer mc_pat_invalid",
			repo: &stubTokenRepo{
				errOn: hashOf("invalid"),
			},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
			wantEndUserID:  "",
		},
		{
			name:       "mc_pat_ token valid → identity injected, next called",
			authHeader: "Bearer mc_pat_valid",
			repo: &stubTokenRepo{
				byHash: map[string]*domainEnduser.APIToken{
					hashOf("valid"): {
						ID:        "tok-1",
						OrgName:   "acme",
						EndUserID: "eu-42",
						CreatedAt: time.Now(),
					},
				},
			},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantEndUserID:  "eu-42",
			wantCtxUserID:  "eu-42",
			wantCtxOrgName: "acme",
		},
		{
			name:           "no Authorization header passes through",
			authHeader:     "", // don't set the header
			repo:           &stubTokenRepo{byHash: map[string]*domainEnduser.APIToken{}},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantEndUserID:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := buildAPITokenServiceFromStub(tc.repo)

			nextCalled := false
			var gotIdentity *EndUserIdentity

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				gotIdentity = GetEndUserIdentity(r.Context())

				// Verify ctxutils fields are set when identity is expected
				if tc.wantCtxUserID != "" {
					if uid, err := ctxutils.GetUserIDFromContext(r.Context()); err != nil || uid != tc.wantCtxUserID {
						t.Errorf("ctxutils UserID: got %q (err %v), want %q", uid, err, tc.wantCtxUserID)
					}
				}
				if tc.wantCtxOrgName != "" {
					if org, err := ctxutils.GetOrgNameFromContext(r.Context()); err != nil || org != tc.wantCtxOrgName {
						t.Errorf("ctxutils OrgName: got %q (err %v), want %q", org, err, tc.wantCtxOrgName)
					}
				}

				w.WriteHeader(http.StatusOK)
			})

			handler := ChiRuntimePATMiddleware(svc, nil)(next)

			req := httptest.NewRequest(
				http.MethodPost,
				"/end-user/graphql/org/acme/project/p/db/d/model/m",
				nil,
			)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", rr.Code, tc.wantStatus)
			}
			if nextCalled != tc.wantNextCalled {
				t.Errorf("nextCalled: got %v, want %v", nextCalled, tc.wantNextCalled)
			}
			if tc.wantEndUserID != "" {
				if gotIdentity == nil {
					t.Fatal("expected EndUserIdentity in context, got nil")
				}
				if gotIdentity.EndUserID != tc.wantEndUserID {
					t.Errorf("EndUserID: got %q, want %q", gotIdentity.EndUserID, tc.wantEndUserID)
				}
			} else if nextCalled && gotIdentity != nil {
				// non-PAT pass-through: identity must NOT be injected
				t.Errorf("expected no EndUserIdentity for non-PAT request, got %+v", gotIdentity)
			}
		})
	}
}
