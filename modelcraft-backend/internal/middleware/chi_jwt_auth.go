package middleware

import (
	"modelcraft/internal/domain/auth"
	"modelcraft/pkg/ctxutils"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthConfig holds configuration for the JWT authentication middleware.
type JWTAuthConfig struct {
	ModelCraftSecret []byte
	SkipValidation   bool
	APIKeyVerifier   APIKeyVerifier
	// InternalToken allows BFF server-side callers to authenticate via X-Internal-Token header,
	// bypassing the requirement for a user JWT. When set, requests carrying a matching
	// X-Internal-Token are granted access without a userID in context.
	InternalToken string
}

// writeJSONError is a helper for writing JSON error responses.
func writeJSONError(w http.ResponseWriter, status int, errMsg, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + errMsg + `"}`))
}

// ChiJWTAuthMiddleware validates JWT tokens or API keys from the Authorization header.
// Token path: "Bearer <jwt>" — HMAC-SHA256 signed ModelCraft JWT
// API Key path: "Bearer mc_*" — hashed lookup (A6 will inject real verifier)
// Internal Token path: "X-Internal-Token" header — BFF server-side callers bypass JWT requirement
func ChiJWTAuthMiddleware(config *JWTAuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.SkipValidation {
				next.ServeHTTP(w, r)
				return
			}

			if tryInternalTokenAuth(config, w, r, next) {
				return
			}

			token := extractBearerToken(r)
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// API Key path: mc_ prefix
			if strings.HasPrefix(token, "mc_") {
				if config.APIKeyVerifier == nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				userID, err := config.APIKeyVerifier.VerifyAPIKey(r.Context(), token)
				if err != nil || userID == "" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				ctx := ctxutils.SetUserID(r.Context(), userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// JWT path: HMAC-SHA256 ModelCraft JWT
			claims, err := validateModelCraftJWT(config.ModelCraftSecret, token)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := ctxutils.SetUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// tryInternalTokenAuth checks for X-Internal-Token header and handles the request if present.
// Returns true if the request was handled (either accepted or rejected), false if not an internal token request.
func tryInternalTokenAuth(config *JWTAuthConfig, w http.ResponseWriter, r *http.Request, next http.Handler) bool {
	if config.InternalToken == "" {
		return false
	}
	provided := r.Header.Get("X-Internal-Token")
	if provided == "" {
		return false
	}
	if !matchInternalToken(config.InternalToken, provided) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return true
	}
	next.ServeHTTP(w, r)
	return true
}

// extractBearerToken extracts the Bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// validateModelCraftJWT validates an HMAC-SHA256 signed ModelCraft JWT.
func validateModelCraftJWT(secret []byte, tokenString string) (*auth.UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &auth.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	claims, ok := token.Claims.(*auth.UserClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}
	return claims, nil
}
