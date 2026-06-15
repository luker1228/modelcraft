package middleware

import (
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"

	"modelcraft/pkg/httpheader"
)

// issuerPlatform 是统一 token 体系的 issuer 标识（与 domain/auth.IssuerPlatform 一致）。
const issuerPlatform = "mc-platform"

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
		authHeader := r.Header.Get(httpheader.Authorization)
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

		// 5. 从 Gateway 注入的 X-Token-Scope header 读取 scope 验证已移除。
		// 端点级鉴权将通过 aud 字段实现（后续迭代）。

		ctx = ctxutils.SetEndUserID(ctx, endUserID)
		ctx = ctxutils.SetUserType(ctx, ctxutils.UserTypeEndUser)

		m.logger.Debug(ctx, "EndUser authenticated",
			logfacade.String("endUserId", endUserID),
			logfacade.String("issuer", issuer))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
