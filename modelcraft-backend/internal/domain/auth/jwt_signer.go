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
		issuer = string(IssuerLegacy)
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
		issuer:     "modelcraft",
		ttl:        defaultAccessTokenTTL,
	}, nil
}

// IssueAccessToken signs a short-lived ES256 JWT for the given user.
func (s *JWTSigner) IssueAccessToken(userID, userName string) (string, error) {
	now := time.Now()
	claims := &UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	_ = userName // reserved for future claims enrichment
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(s.privateKey)
}

// TTLSeconds returns the access token TTL in seconds.
func (s *JWTSigner) TTLSeconds() int {
	return int(s.ttl.Seconds())
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
