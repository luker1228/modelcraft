package middleware

import (
	"context"
	"modelcraft/internal/domain/rls"
	"modelcraft/pkg/httpheader"
	"net/http"
	"strings"
)

type rlsContextKey struct{}

// RLSContextMiddleware 从 X-MC-Auth-* Header 提取 UserContext 注入 context
type RLSContextMiddleware struct{}

// NewRLSContextMiddleware 创建 RLSContextMiddleware
func NewRLSContextMiddleware() *RLSContextMiddleware {
	return &RLSContextMiddleware{}
}

// Middleware 提取 X-MC-Auth-* headers 并注入 context
func (m *RLSContextMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uc := &rls.UserContext{
			UserID:   strings.TrimSpace(r.Header.Get(httpheader.XMCAuthUserID)),
			UserName: strings.TrimSpace(r.Header.Get(httpheader.XMCAuthUserName)),
		}

		rolesStr := strings.TrimSpace(r.Header.Get(httpheader.XMCAuthRoles))
		if rolesStr != "" {
			parts := strings.Split(rolesStr, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					uc.Roles = append(uc.Roles, p)
				}
			}
		}

		ctx := context.WithValue(r.Context(), rlsContextKey{}, uc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserContext 从 context 获取 UserContext
func GetUserContext(ctx context.Context) *rls.UserContext {
	uc, ok := ctx.Value(rlsContextKey{}).(*rls.UserContext)
	if !ok {
		return nil
	}
	return uc
}
