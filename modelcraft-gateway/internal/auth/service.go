package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT payload for access tokens issued by the gateway.
type Claims struct {
	UserID   string `json:"sub"`
	Username string `json:"username"`
	Email    string `json:"email"`
	OrgName  string `json:"orgName,omitempty"`
	jwt.RegisteredClaims
}

// Service handles JWT issuance, verification, and refresh-token management.
type Service struct {
	secret            []byte
	accessTokenTTL    time.Duration
	refreshTokenTTL   time.Duration
	refreshCookieName string
}

func NewService(secret string, accessTTL, refreshTTL time.Duration, cookieName string) *Service {
	return &Service{
		secret:            []byte(secret),
		accessTokenTTL:    accessTTL,
		refreshTokenTTL:   refreshTTL,
		refreshCookieName: cookieName,
	}
}

// IssueAccessToken signs a short-lived JWT access token.
func (s *Service) IssueAccessToken(userID, username, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// VerifyAccessToken parses and validates an access token, returning its claims.
func (s *Service) VerifyAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// GenerateRefreshToken creates a cryptographically random opaque refresh token.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
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

// GetRefreshCookie reads the refresh token from the request cookie.
func (s *Service) GetRefreshCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(s.refreshCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
