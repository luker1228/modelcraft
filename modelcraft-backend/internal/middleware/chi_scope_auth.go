package middleware

import (
	"encoding/json"
	"net/http"
)

const headerTokenScope = "X-Token-Scope" //nolint:gosec // not a credential, this is an HTTP header name

// RequireScope 返回一个 chi 中间件，校验请求的 X-Token-Scope header 是否在 allowedScopes 中。
//
// 设计原则：防止向上越权，不限制向下调用。
//   - Org 路由只传入 "org"：阻止 scope=project token 访问 org 管理接口
//   - Project/Runtime 路由传入 "org", "project"：上级（org）和平级（project）均可访问
//
// 若 header 为空（SkipValidation 开发模式或 X-Internal-Token BFF 路径），直接放行。
// 若 header 不在允许列表，返回 403 JSON 错误。
func RequireScope(allowedScopes ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedScopes))
	for _, s := range allowedScopes {
		allowed[s] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scope := r.Header.Get(headerTokenScope)
			if scope == "" {
				// header 为空：SkipValidation / InternalToken 模式，放行
				next.ServeHTTP(w, r)
				return
			}
			if !allowed[scope] {
				writeScopeError(w, http.StatusForbidden, "INSUFFICIENT_SCOPE",
					"token scope is not allowed for this endpoint")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeScopeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
