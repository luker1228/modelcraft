package middleware

import (
	"modelcraft/pkg/ctxutils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// okHandler is a test handler that records whether it was called.
func okHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	})
}

// requestWithPermissions builds a request whose context carries the given permissions.
func requestWithPermissions(permissions []string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if permissions != nil {
		ctx := r.Context()
		ctx = ctxutils.SetContextValue(ctx, ctxutils.ContextKeyPermissions, permissions)
		r = r.WithContext(ctx)
	}
	return r
}

func TestChiRequirePermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		required    string
		wantStatus  int
		wantCalled  bool
	}{
		{
			name:        "exact permission grants access",
			permissions: []string{"project:create"},
			required:    "project:create",
			wantStatus:  http.StatusOK,
			wantCalled:  true,
		},
		{
			name:        "wildcard permission grants access",
			permissions: []string{"project:*"},
			required:    "project:create",
			wantStatus:  http.StatusOK,
			wantCalled:  true,
		},
		{
			name:        "missing permission denies access",
			permissions: []string{"project:read"},
			required:    "project:create",
			wantStatus:  http.StatusForbidden,
			wantCalled:  false,
		},
		{
			name:        "nil permissions denies access",
			permissions: nil,
			required:    "project:create",
			wantStatus:  http.StatusForbidden,
			wantCalled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			w := httptest.NewRecorder()
			r := requestWithPermissions(tt.permissions)

			ChiRequirePermission(tt.required)(okHandler(&called)).ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantCalled, called)
		})
	}
}

func TestChiRequireAnyPermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		required    []string
		wantStatus  int
		wantCalled  bool
	}{
		{
			name:        "has first permission",
			permissions: []string{"project:create"},
			required:    []string{"project:create", "project:delete"},
			wantStatus:  http.StatusOK,
			wantCalled:  true,
		},
		{
			name:        "has second permission",
			permissions: []string{"project:delete"},
			required:    []string{"project:create", "project:delete"},
			wantStatus:  http.StatusOK,
			wantCalled:  true,
		},
		{
			name:        "has none of required permissions",
			permissions: []string{"model:read"},
			required:    []string{"project:create", "project:delete"},
			wantStatus:  http.StatusForbidden,
			wantCalled:  false,
		},
		{
			name:        "nil permissions denies access",
			permissions: nil,
			required:    []string{"project:create"},
			wantStatus:  http.StatusForbidden,
			wantCalled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			w := httptest.NewRecorder()
			r := requestWithPermissions(tt.permissions)

			ChiRequireAnyPermission(tt.required...)(okHandler(&called)).ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantCalled, called)
		})
	}
}

func TestChiRequireAllPermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		required    []string
		wantStatus  int
		wantCalled  bool
	}{
		{
			name:        "has all required permissions",
			permissions: []string{"project:create", "model:read"},
			required:    []string{"project:create", "model:read"},
			wantStatus:  http.StatusOK,
			wantCalled:  true,
		},
		{
			name:        "missing one permission denies access",
			permissions: []string{"project:create"},
			required:    []string{"project:create", "model:read"},
			wantStatus:  http.StatusForbidden,
			wantCalled:  false,
		},
		{
			name:        "nil permissions denies access",
			permissions: nil,
			required:    []string{"project:create"},
			wantStatus:  http.StatusForbidden,
			wantCalled:  false,
		},
		{
			name:        "global wildcard satisfies all",
			permissions: []string{"*"},
			required:    []string{"project:create", "model:delete", "cluster:manage"},
			wantStatus:  http.StatusOK,
			wantCalled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			w := httptest.NewRecorder()
			r := requestWithPermissions(tt.permissions)

			ChiRequireAllPermissions(tt.required...)(okHandler(&called)).ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantCalled, called)
		})
	}
}
