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
func ChiJWTAuthMiddleware(config *JWTAuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.SkipValidation {
				next.ServeHTTP(w, r)
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
