package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireScope_AllowedScope(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := RequireScope("org")
	handler := mw(next)

	req := httptest.NewRequest(http.MethodPost, "/graphql/org/acme", nil)
	req.Header.Set("X-Token-Scope", "org")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.True(t, handlerCalled, "next handler should be called when scope is allowed")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireScope_DeniedScope(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Org 路由只允许 "org"，scope=project 应被拒绝（防止向上越权）
	mw := RequireScope("org")
	handler := mw(next)

	req := httptest.NewRequest(http.MethodPost, "/graphql/org/acme", nil)
	req.Header.Set("X-Token-Scope", "project") // scope=project 不能访问 org 管理路由
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, handlerCalled, "next handler must NOT be called when scope is denied")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "INSUFFICIENT_SCOPE")
}

func TestRequireScope_EmptyHeader(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := RequireScope("org")
	handler := mw(next)

	req := httptest.NewRequest(http.MethodPost, "/graphql/org/acme", nil)
	// No X-Token-Scope header — simulates SkipValidation / X-Internal-Token mode
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.True(t, handlerCalled, "next handler should be called when X-Token-Scope header is absent")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireScope_MultiScopeAllowed(t *testing.T) {
	// Project/Runtime 路由接受 org 和 project 两种 scope（上级可向下调用）
	for _, scope := range []string{"org", "project"} {
		t.Run("scope="+scope, func(t *testing.T) {
			handlerCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			mw := RequireScope("org", "project")
			handler := mw(next)

			req := httptest.NewRequest(http.MethodPost, "/graphql/org/acme/project/my-proj", nil)
			req.Header.Set("X-Token-Scope", scope)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.True(t, handlerCalled, "scope=%s should be allowed for project routes", scope)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
