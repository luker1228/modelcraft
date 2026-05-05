package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT payload for access tokens issued by the backend auth service.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Service handles JWT verification and refresh-token cookie management.
// Token signing is fully delegated to the backend auth service (ES256).
// The gateway only holds the EC public key for verification.
type Service struct {
	publicKey                *ecdsa.PublicKey
	// Deprecated: endUserJWTSecret 用于 HMAC 端用户 token 验证，已不再使用。
	// 将在阶段 3 Schema 清理时移除。
	endUserJWTSecret         []byte
	refreshTokenTTL          time.Duration
	refreshCookieName        string
	endUserRefreshCookieName string
}

// NewService parses a PEM-encoded EC public key and creates a Service.
func NewService(publicKeyPEM string, refreshTTL time.Duration, cookieName string, endUserCookieName string, endUserSecret string) (*Service, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("auth: failed to decode public key PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("auth: parse public key: %w", err)
	}

	ecPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("auth: public key is not an EC key")
	}

	return &Service{
		publicKey:                ecPub,
		endUserJWTSecret:         []byte(endUserSecret),
		refreshTokenTTL:          refreshTTL,
		refreshCookieName:        cookieName,
		endUserRefreshCookieName: endUserCookieName,
	}, nil
}

// VerifyAccessToken parses and validates an ES256 access token from the backend.
func (s *Service) VerifyAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// SetRefreshCookie writes the refresh token as an httpOnly, Secure, SameSite=Strict cookie.
func (s *Service) SetRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.refreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // set to true in production (HTTPS)
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(s.refreshTokenTTL.Seconds()),
	})
}

// SetEndUserRefreshCookie writes the end-user refresh token cookie.
func (s *Service) SetEndUserRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.endUserRefreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // set to true in production (HTTPS)
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(s.refreshTokenTTL.Seconds()),
	})
}

// ClearRefreshCookie expires the refresh cookie immediately.
func (s *Service) ClearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.refreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// ClearEndUserRefreshCookie expires the end-user refresh cookie immediately.
func (s *Service) ClearEndUserRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.endUserRefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// GetRefreshCookie reads the refresh token from the request cookie.
func (s *Service) GetRefreshCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(s.refreshCookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

// Deprecated: EndUserClaims 是 HMAC 端用户 token 的 payload 结构，已被 PlatformClaims（backend）替代。
// 此类型将在阶段 3 Schema 清理时删除。
type EndUserClaims struct {
	UserID  string `json:"user_id"`
	OrgName string `json:"org_name"`
	jwt.RegisteredClaims
}

// Deprecated: VerifyEndUserAccessToken 使用 HMAC-SHA256 验证端用户 token。
// 端用户 token 已在 Backend Token 核心统一阶段迁移至 ES256（mc-platform issuer）。
// 请改用 VerifyAccessToken 进行验证。此方法将在阶段 3 Schema 清理时删除。
func (s *Service) VerifyEndUserAccessToken(tokenString string) (*EndUserClaims, error) {
	var claims EndUserClaims

	if len(s.endUserJWTSecret) == 0 {
		// Dev mode: no secret configured — decode without verifying signature.
		p := jwt.NewParser()
		token, _, err := p.ParseUnverified(tokenString, &claims)
		if err != nil {
			return nil, fmt.Errorf("end-user token parse (unverified): %w", err)
		}
		c, ok := token.Claims.(*EndUserClaims)
		if !ok {
			return nil, errors.New("end-user token: invalid claims type")
		}
		return c, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.endUserJWTSecret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := token.Claims.(*EndUserClaims)
	if !ok || !token.Valid {
		return nil, errors.New("end-user token: invalid claims")
	}
	return c, nil
}
func (s *Service) GetEndUserRefreshCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(s.endUserRefreshCookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}
