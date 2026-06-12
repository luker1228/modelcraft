package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRLSContextMiddleware_AllHeaders(t *testing.T) {
	var capturedCtx context.Context
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(200)
	})

	mw := NewRLSContextMiddleware()
	handler := mw.Middleware(next)

	req := httptest.NewRequest("POST", "/api/data", nil)
	req.Header.Set("X-MC-User-ID", "user_123")
	req.Header.Set("X-MC-User-Name", "zhangsan")
	req.Header.Set("X-MC-User-Roles", "admin, manager")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	uc := GetUserContext(capturedCtx)
	if uc == nil {
		t.Fatal("expected UserContext in context, got nil")
	}
	if uc.UserID != "user_123" {
		t.Errorf("expected UserID 'user_123', got %q", uc.UserID)
	}
	if uc.UserName != "zhangsan" {
		t.Errorf("expected UserName 'zhangsan', got %q", uc.UserName)
	}
	if len(uc.Roles) != 2 || uc.Roles[0] != "admin" || uc.Roles[1] != "manager" {
		t.Errorf("expected Roles [admin, manager], got %v", uc.Roles)
	}
}

func TestRLSContextMiddleware_NoHeaders(t *testing.T) {
	var capturedCtx context.Context
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
	})

	mw := NewRLSContextMiddleware()
	handler := mw.Middleware(next)

	req := httptest.NewRequest("POST", "/api/data", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	uc := GetUserContext(capturedCtx)
	if uc == nil {
		t.Fatal("expected UserContext in context, got nil")
	}
	if uc.UserID != "" {
		t.Errorf("expected empty UserID, got %q", uc.UserID)
	}
	if len(uc.Roles) != 0 {
		t.Errorf("expected empty Roles, got %v", uc.Roles)
	}
}

func TestRLSContextMiddleware_EmptyRoles(t *testing.T) {
	var capturedCtx context.Context
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
	})

	mw := NewRLSContextMiddleware()
	handler := mw.Middleware(next)

	req := httptest.NewRequest("POST", "/api/data", nil)
	req.Header.Set("X-MC-User-Roles", "  , ,  ")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	uc := GetUserContext(capturedCtx)
	if uc == nil {
		t.Fatal("expected UserContext in context, got nil")
	}
	if len(uc.Roles) != 0 {
		t.Errorf("expected empty Roles, got %v", uc.Roles)
	}
}
