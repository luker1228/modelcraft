package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"modelcraft/pkg/logfacade"
)

// CasdoorProvider implements AuthProvider interface for Casdoor authentication.
type CasdoorProvider struct {
	endpoint     string
	clientID     string
	clientSecret string
	organization string
	application  string
	certificate  string
	publicKey    *rsa.PublicKey // Cached public key
	logger       logfacade.Logger
}

// CasdoorConfig holds the configuration for Casdoor provider.
type CasdoorConfig struct {
	Endpoint     string `json:"endpoint" mapstructure:"endpoint"`
	ClientID     string `json:"client_id" mapstructure:"client_id"`
	ClientSecret string `json:"client_secret" mapstructure:"client_secret"`
	Organization string `json:"organization" mapstructure:"organization"`
	Application  string `json:"application" mapstructure:"application"`
	Certificate  string `json:"certificate" mapstructure:"certificate"`
}

// NewCasdoorProvider creates a new Casdoor authentication provider.
func NewCasdoorProvider(config map[string]interface{}) (*CasdoorProvider, error) {
	logger := logfacade.GetLogger(context.Background())

	endpoint, _ := config["endpoint"].(string)
	clientID, _ := config["client_id"].(string)
	clientSecret, _ := config["client_secret"].(string)
	organization, _ := config["organization"].(string)
	application, _ := config["application"].(string)
	certificate, _ := config["certificate"].(string)

	provider := &CasdoorProvider{
		endpoint:     endpoint,
		clientID:     clientID,
		clientSecret: clientSecret,
		organization: organization,
		application:  application,
		certificate:  certificate,
		logger:       logger,
	}

	logger.Infof(context.Background(),
		"Initializing Casdoor provider: endpoint=%s, organization=%s", endpoint, organization)

	return provider, nil
}

// Type returns the provider type identifier.
func (p *CasdoorProvider) Type() string {
	return "casdoor"
}

// GetSigningMethod returns the JWT signing algorithm used by Casdoor.
func (p *CasdoorProvider) GetSigningMethod() string {
	return "RS256"
}

// GetPublicKey extracts and caches the RSA public key from the X.509 certificate.
func (p *CasdoorProvider) GetPublicKey(ctx context.Context) (interface{}, error) {
	// Return cached key if already parsed
	if p.publicKey != nil {
		return p.publicKey, nil
	}

	// Parse PEM-encoded certificate
	block, _ := pem.Decode([]byte(p.certificate))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from certificate")
	}

	// Parse X.509 certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		p.logger.Errorf(context.Background(), "Failed to parse Casdoor certificate: %v", err)
		return nil, fmt.Errorf("failed to parse X.509 certificate: %w", err)
	}

	// Extract RSA public key
	rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("certificate public key is not RSA")
	}

	// Cache the public key
	p.publicKey = rsaPublicKey

	p.logger.Infof(context.Background(), "Parsed Casdoor certificate, RSA key size: %d bits", rsaPublicKey.N.BitLen())

	return rsaPublicKey, nil
}
