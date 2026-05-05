package middleware

import (
	"context"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
)

type endUserContextKeyType string

// endUserContextKey is the context key for end-user identity.
const endUserContextKey endUserContextKeyType = "end_user_identity"

// issuerPlatform 是统一 token 体系的 issuer 标识（与 domain/auth.IssuerPlatform 一致）。
const issuerPlatform = "mc-platform"

// EndUserIdentity represents the authenticated end-user identity.
type EndUserIdentity struct {
	EndUserID string `json:"endUserId"`
	Issuer    string `json:"issuer"` // 统一为 mc-platform
	Scope     string `json:"scope"`  // "org" | "project"
}

// IsEndUser 判断是否为可访问 Runtime 的身份（统一 token 体系后检查 mc-platform issuer）。
// 阶段 1：所有 mc-platform token 均可访问 runtime；scope 强制校验由阶段 2 引入。
func (e *EndUserIdentity) IsEndUser() bool {
	return e.Issuer == issuerPlatform
}

// IsDeveloper Deprecated：统一 token 体系后，developer/enduser 概念消失。
// 保留以兼容下游调用，将在阶段 2 重命名为 HasOrgScope。
func (e *EndUserIdentity) IsDeveloper() bool {
	return e.Issuer == issuerPlatform
}

// RuntimeAuthMiddleware validates JWT for Runtime endpoints.
// Only accepts tokens with iss="mc-platform"（统一 Token 体系后，所有 token 均使用此 issuer）。
type RuntimeAuthMiddleware struct {
	jwtValidator JWTValidator
	logger       logfacade.Logger
}

// JWTValidator defines the interface for JWT validation.
type JWTValidator interface {
	Validate(token string) (map[string]interface{}, error)
}

// NewRuntimeAuthMiddleware creates a new RuntimeAuthMiddleware.
func NewRuntimeAuthMiddleware(validator JWTValidator, logger logfacade.Logger) *RuntimeAuthMiddleware {
	return &RuntimeAuthMiddleware{
		jwtValidator: validator,
		logger:       logger,
	}
}

// Middleware returns the HTTP middleware handler.
func (m *RuntimeAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 1. Extract JWT from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn(ctx, "Missing Authorization header")
			http.Error(w, `{"error": "Unauthorized: Missing token"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			m.logger.Warn(ctx, "Invalid Authorization header format")
			http.Error(w, `{"error": "Unauthorized: Invalid token format"}`, http.StatusUnauthorized)
			return
		}
		token := parts[1]

		// 2. Parse and validate JWT
		claims, err := m.jwtValidator.Validate(token)
		if err != nil {
			m.logger.Warn(ctx, "Invalid JWT", logfacade.Err(err))
			http.Error(w, `{"error": "Unauthorized: Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// 3. Validate issuer must be "mc-platform"
		issuer, ok := claims["iss"].(string)
		if !ok {
			m.logger.Warn(ctx, "Missing iss claim in JWT")
			http.Error(w, `{"error": "Unauthorized: Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		if issuer != issuerPlatform {
			m.logger.Warn(ctx, "Invalid JWT issuer for runtime endpoint",
				logfacade.String("issuer", issuer),
				logfacade.String("expected", issuerPlatform))
			http.Error(
				w,
				`{"error": "Unauthorized: Invalid issuer. Runtime endpoints require mc-platform JWT"}`,
				http.StatusUnauthorized,
			)
			return
		}

		// 4. Extract endUserId from claims
		endUserID, ok := claims["user_id"].(string)
		if !ok || endUserID == "" {
			m.logger.Warn(ctx, "Missing user_id claim in JWT")
			http.Error(w, `{"error": "Unauthorized: Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		// 5. Inject end-user identity into context（含 scope 字段）
		scope, _ := claims["scope"].(string)
		identity := &EndUserIdentity{
			EndUserID: endUserID,
			Issuer:    issuer,
			Scope:     scope,
		}
		ctx = context.WithValue(ctx, endUserContextKey, identity)

		m.logger.Debug(ctx, "EndUser authenticated",
			logfacade.String("endUserId", endUserID),
			logfacade.String("issuer", issuer))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetEndUserIdentity retrieves the end-user identity from context.
// Returns nil if no identity is found.
func GetEndUserIdentity(ctx context.Context) *EndUserIdentity {
	identity, ok := ctx.Value(endUserContextKey).(*EndUserIdentity)
	if !ok {
		return nil
	}
	return identity
}

// GetEndUserID retrieves the end-user ID from context.
// Returns empty string if no identity is found.
func GetEndUserID(ctx context.Context) string {
	identity := GetEndUserIdentity(ctx)
	if identity == nil {
		return ""
	}
	return identity.EndUserID
}
