package middleware

import (
	"context"
	"net/http"
	"strings"

	"modelcraft/pkg/logfacade"
)

// EndUserContextKey is the context key for end-user identity.
const EndUserContextKey = "end_user_identity"

// EndUserIdentity represents the authenticated end-user identity.
type EndUserIdentity struct {
	EndUserID string `json:"endUserId"`
	Issuer    string `json:"issuer"`
}

// IsEndUser returns true if the identity is a valid end-user (mc-enduser issuer).
func (e *EndUserIdentity) IsEndUser() bool {
	return e.Issuer == "mc-enduser"
}

// IsDeveloper returns true if the identity is a developer (mc-developer issuer).
func (e *EndUserIdentity) IsDeveloper() bool {
	return e.Issuer == "mc-developer"
}

// RuntimeAuthMiddleware validates JWT for Runtime endpoints.
// It only accepts tokens with iss="mc-enduser" (EndUser JWT).
// Developer JWTs (iss="mc-developer") are rejected with 401.
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

		// 3. Validate issuer must be "mc-enduser"
		issuer, ok := claims["iss"].(string)
		if !ok {
			m.logger.Warn(ctx, "Missing iss claim in JWT")
			http.Error(w, `{"error": "Unauthorized: Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		if issuer != "mc-enduser" {
			m.logger.Warn(ctx, "Invalid JWT issuer for runtime endpoint",
				logfacade.String("issuer", issuer),
				logfacade.String("expected", "mc-enduser"))
			http.Error(w, `{"error": "Unauthorized: Invalid issuer. Runtime endpoints only accept EndUser JWT (mc-enduser)"}`, http.StatusUnauthorized)
			return
		}

		// 4. Extract endUserId from claims
		endUserID, ok := claims["user_id"].(string)
		if !ok || endUserID == "" {
			m.logger.Warn(ctx, "Missing user_id claim in JWT")
			http.Error(w, `{"error": "Unauthorized: Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		// 5. Inject end-user identity into context
		identity := &EndUserIdentity{
			EndUserID: endUserID,
			Issuer:    issuer,
		}
		ctx = context.WithValue(ctx, EndUserContextKey, identity)

		m.logger.Debug(ctx, "EndUser authenticated",
			logfacade.String("endUserId", endUserID),
			logfacade.String("issuer", issuer))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetEndUserIdentity retrieves the end-user identity from context.
// Returns nil if no identity is found.
func GetEndUserIdentity(ctx context.Context) *EndUserIdentity {
	identity, ok := ctx.Value(EndUserContextKey).(*EndUserIdentity)
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
