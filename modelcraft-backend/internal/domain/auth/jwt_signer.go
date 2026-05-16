package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AudienceTenant is the aud value for tenant (developer/admin) tokens.
const AudienceTenant = "tenant"

// AudienceEndUser is the aud value for end-user tokens.
const AudienceEndUser = "end_user"

// ApisixConsumerKey is the APISIX jwt-auth Consumer key embedded in every
// platform token. It must match the Consumer definition in APISIX config.
const ApisixConsumerKey = "mcuser"

const defaultAccessTokenTTL = time.Hour

// JWTSigner signs access tokens using ES256 (ECDSA P-256).
// The private key stays exclusively in the auth service; the gateway
// only needs the corresponding public key to verify tokens.
type JWTSigner struct {
	privateKey *ecdsa.PrivateKey
	issuer     string
	ttl        time.Duration
}

// NewJWTSignerFromPEM parses a PEM-encoded EC private key and returns a JWTSigner.
// The key must use the P-256 curve (openssl ecparam -name prime256v1 ...).
func NewJWTSignerFromPEM(
	pemKey string,
	issuer string,
	ttl time.Duration,
) (*JWTSigner, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, errors.New("jwt_signer: failed to decode PEM block")
	}

	var key *ecdsa.PrivateKey
	var err error

	switch block.Type {
	case "EC PRIVATE KEY":
		key, err = x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		parsed, parseErr := x509.ParsePKCS8PrivateKey(block.Bytes)
		if parseErr != nil {
			return nil, fmt.Errorf("jwt_signer: parse PKCS8: %w", parseErr)
		}
		var ok bool
		key, ok = parsed.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("jwt_signer: key is not an EC private key")
		}
	default:
		return nil, fmt.Errorf("jwt_signer: unsupported PEM block type: %s", block.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("jwt_signer: parse EC key: %w", err)
	}
	if key.Curve != elliptic.P256() {
		return nil, errors.New("jwt_signer: only P-256 (ES256) is supported")
	}

	if ttl == 0 {
		ttl = defaultAccessTokenTTL
	}
	if issuer == "" {
		issuer = string(IssuerPlatform)
	}

	return &JWTSigner{privateKey: key, issuer: issuer, ttl: ttl}, nil
}

// GenerateDevSigner creates an ephemeral in-memory signer for local development.
// The generated key is NOT persisted; tokens are invalid after restart.
func GenerateDevSigner() (*JWTSigner, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("jwt_signer: generate dev key: %w", err)
	}
	return &JWTSigner{
		privateKey: key,
		issuer:     string(IssuerPlatform),
		ttl:        defaultAccessTokenTTL,
	}, nil
}

// IssueAccessToken 为指定用户签发短效 ES256 JWT（PlatformClaims 格式）。
// orgName 不能为空。aud 字段用于标识受众类型（tenant / end_user），由调用方传入。
// endUserAdminIDs 仅在 tenant token 中携带，映射 orgName → 该 org 的 end-user super-admin ID。
func (s *JWTSigner) IssueAccessToken(
	userID, orgName string,
	aud jwt.ClaimStrings,
	endUserAdminIDs map[string]string,
) (string, error) {
	if orgName == "" {
		return "", errors.New("jwt_signer: orgName is required")
	}
	now := time.Now()
	claims := &PlatformClaims{
		UserID:          userID,
		OrgName:         orgName,
		Key:             ApisixConsumerKey,
		EndUserAdminIDs: endUserAdminIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(IssuerPlatform),
			Subject:   userID,
			Audience:  aud,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(s.privateKey)
}

// TTLSeconds returns the access token TTL in seconds.
func (s *JWTSigner) TTLSeconds() int {
	return int(s.ttl.Seconds())
}

// ParsePlatformClaims parses and validates an ES256 JWT using the signer's own public key.
// Returns the PlatformClaims if valid, or an error if the token is invalid, expired, or malformed.
func (s *JWTSigner) ParsePlatformClaims(tokenStr string) (*PlatformClaims, error) {
	claims := &PlatformClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("jwt_signer: unexpected signing method: %v", t.Header["alg"])
		}
		return &s.privateKey.PublicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt_signer: parse platform claims: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("jwt_signer: token is not valid")
	}
	if err := claims.Validate(); err != nil {
		return nil, fmt.Errorf("jwt_signer: invalid claims: %w", err)
	}
	return claims, nil
}

// PublicKeyPEM returns the PEM-encoded public key so the gateway can load it at startup.
func (s *JWTSigner) PublicKeyPEM() (string, error) {
	der, err := x509.MarshalPKIXPublicKey(&s.privateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("jwt_signer: marshal public key: %w", err)
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}
	return string(pem.EncodeToMemory(block)), nil
}
