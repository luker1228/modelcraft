# Spec: Casdoor Provider Implementation

## Overview

This specification defines the Casdoor-specific implementation of the `AuthProvider` interface. Casdoor is an open-source Identity and Access Management (IAM) platform that supports OAuth 2.0, OIDC, and multiple identity sources (LDAP, WeCom, GitHub, etc.).

## Context

Casdoor is chosen as the default authentication provider for ModelCraft because:
- Go-native implementation (low resource usage ~100MB vs Keycloak ~1GB)
- Built-in support for Chinese enterprise identity sources (WeCom, DingTalk, LDAP)
- Comprehensive OAuth 2.0 / OIDC support
- Official Go SDK available

Casdoor issues JWT tokens signed with RS256 (RSA-SHA256) using X.509 certificates.

## ADDED Requirements

### Requirement: Casdoor Provider Implementation

The system MUST implement the AuthProvider interface for Casdoor.

#### Scenario: Provider initializes with valid configuration

**Given** a Casdoor configuration containing:
```json
{
  "endpoint": "https://casdoor.example.com",
  "client_id": "abc123",
  "organization": "tenant1",
  "application": "modelcraft",
  "certificate": "-----BEGIN CERTIFICATE-----\nMIIE...\n-----END CERTIFICATE-----"
}
```
**When** the system calls `NewCasdoorProvider(config)`
**Then** the provider is created with fields:
- `endpoint` = "https://casdoor.example.com"
- `clientID` = "abc123"
- `organization` = "tenant1"
- `certificate` = full PEM string
- `publicKey` = nil (lazily parsed on first use)

#### Scenario: Provider returns RS256 signing method

**Given** a CasdoorProvider instance
**When** the system calls `GetSigningMethod()`
**Then** the method returns `"RS256"`

#### Scenario: Provider returns type identifier

**Given** a CasdoorProvider instance
**When** the system calls `Type()`
**Then** the method returns `"casdoor"`

---

### Requirement: Public Key Extraction from Certificate

Casdoor uses X.509 certificates for JWT signing. The provider MUST extract the RSA public key from the PEM-encoded certificate.

#### Scenario: Parse valid PEM certificate on first use

**Given** a CasdoorProvider with a valid X.509 certificate in PEM format
**When** the system calls `GetPublicKey(ctx)` for the first time
**Then** the method decodes the PEM block using `pem.Decode()`
**And** parses the certificate using `x509.ParseCertificate()`
**And** extracts the RSA public key from `cert.PublicKey`
**And** caches the key in `provider.publicKey` field
**And** returns the *rsa.PublicKey

#### Scenario: Return cached key on subsequent calls

**Given** a CasdoorProvider with cached publicKey != nil
**When** the system calls `GetPublicKey(ctx)` again
**Then** the method immediately returns the cached `provider.publicKey`
**And** does not re-parse the certificate

#### Scenario: Reject invalid PEM format

**Given** a CasdoorProvider with certificate = "invalid-not-pem"
**When** the system calls `GetPublicKey(ctx)`
**Then** the method returns error "failed to decode PEM block"
**And** publicKey remains nil

#### Scenario: Reject invalid certificate

**Given** a CasdoorProvider with a PEM block that is not a valid X.509 certificate
**When** the system calls `GetPublicKey(ctx)`
**Then** the method returns error "failed to parse certificate: ..."
**And** publicKey remains nil

#### Scenario: Reject non-RSA public key

**Given** a CasdoorProvider with a certificate containing an ECDSA public key
**When** the system calls `GetPublicKey(ctx)`
**Then** the method returns error "certificate public key is not RSA"
**And** publicKey remains nil

---

### Requirement: Claims Parsing from Casdoor JWT

Casdoor JWT tokens contain specific claim fields that MUST be mapped to UnifiedClaims.

#### Scenario: Parse standard Casdoor claims

**Given** a Casdoor JWT with claims:
```json
{
  "sub": "built-in/alice",
  "name": "Alice Wang",
  "owner": "tenant1",
  "email": "alice@example.com",
  "iss": "https://casdoor.example.com",
  "aud": ["abc123"],
  "exp": 1738368000,
  "iat": 1738364400
}
```
**When** the provider calls `ParseClaims(mapClaims)`
**Then** the method returns UnifiedClaims with:
- `UserID` = "built-in/alice" (from `sub`)
- `Username` = "Alice Wang" (from `name`)
- `Email` = "alice@example.com" (from `email`)
- `Organization` = "tenant1" (from `owner`)
- `ExpiresAt` = time from `exp` claim
- `Raw` = original mapClaims

#### Scenario: Handle missing optional email claim

**Given** a Casdoor JWT with claims missing `email` field
**When** the provider calls `ParseClaims(mapClaims)`
**Then** the method returns UnifiedClaims with `Email` = ""
**And** no error is returned

#### Scenario: Reject missing required sub claim

**Given** a Casdoor JWT with claims missing `sub` field
**When** the provider calls `ParseClaims(mapClaims)`
**Then** the method returns error "missing required claim: sub"

#### Scenario: Reject missing required owner claim

**Given** a Casdoor JWT with claims missing `owner` field
**When** the provider calls `ParseClaims(mapClaims)`
**Then** the method returns error "missing required claim: owner"

#### Scenario: Reject missing required name claim

**Given** a Casdoor JWT with claims missing `name` field
**When** the provider calls `ParseClaims(mapClaims)`
**Then** the method returns error "missing required claim: name"

---

### Requirement: Configuration Validation

The system MUST validate Casdoor configuration before creating a provider.

#### Scenario: Validate complete configuration

**Given** a Casdoor configuration with all required fields:
- endpoint = "https://casdoor.example.com"
- client_id = "abc123"
- organization = "tenant1"
- certificate = valid PEM string
**When** the system calls `ValidateCasdoorConfig(config)`
**Then** the method returns `nil` (no error)

#### Scenario: Reject missing endpoint

**Given** a Casdoor configuration with empty `endpoint`
**When** the system calls `ValidateCasdoorConfig(config)`
**Then** the method returns error "endpoint is required"

#### Scenario: Reject missing client_id

**Given** a Casdoor configuration with empty `client_id`
**When** the system calls `ValidateCasdoorConfig(config)`
**Then** the method returns error "client_id is required"

#### Scenario: Reject missing organization

**Given** a Casdoor configuration with empty `organization`
**When** the system calls `ValidateCasdoorConfig(config)`
**Then** the method returns error "organization is required"

#### Scenario: Reject missing certificate

**Given** a Casdoor configuration with empty `certificate`
**When** the system calls `ValidateCasdoorConfig(config)`
**Then** the method returns error "certificate is required"

#### Scenario: Reject malformed certificate at validation

**Given** a Casdoor configuration with certificate = "not-a-pem"
**When** the system calls `ValidateCasdoorConfig(config)`
**Then** the method returns error "certificate is not valid PEM format"

---

### Requirement: Logging and Debugging

The provider MUST log initialization and errors for troubleshooting.

#### Scenario: Log provider initialization

**Given** a CasdoorProvider being created
**When** `NewCasdoorProvider(config)` is called
**Then** the system logs at INFO level:
- "Initializing Casdoor provider"
- "Endpoint: https://casdoor.example.com"
- "Organization: tenant1"

#### Scenario: Log public key parsing

**Given** a CasdoorProvider parsing certificate for the first time
**When** `GetPublicKey(ctx)` successfully parses the certificate
**Then** the system logs at DEBUG level:
- "Parsed Casdoor certificate, RSA key size: 2048 bits"

#### Scenario: Log parsing errors

**Given** a CasdoorProvider with invalid certificate
**When** `GetPublicKey(ctx)` fails to parse
**Then** the system logs at ERROR level:
- "Failed to parse Casdoor certificate" with error details

---

## Design Notes

**Certificate Format**: Casdoor uses standard X.509 certificates in PEM format. These can be downloaded from Casdoor Admin UI under Certs section.

**Key Caching**: Public key is cached after first parse to avoid repeated cryptographic operations on every request. Certificate rotation requires restarting the application or invalidating the provider cache.

**Claims Mapping**:
- Casdoor's `owner` field = ModelCraft's `Organization` (for multi-tenancy)
- Casdoor's `sub` field = ModelCraft's `UserID` (unique user identifier)
- Casdoor's `name` field = ModelCraft's `Username` (display name)

**Error Handling**: All errors during key parsing return descriptive messages to aid troubleshooting. Certificate format errors should be caught at config validation time when possible.

---

## Dependencies

**External Libraries**:
- `crypto/x509` - Certificate parsing
- `encoding/pem` - PEM decoding
- `crypto/rsa` - RSA public key type
- `github.com/golang-jwt/jwt/v5` - JWT claims type

**Casdoor Setup**:
- Casdoor must be deployed (via Docker Compose, see deployment guide)
- Organization and Application must be created in Casdoor Admin UI
- Certificate must be obtained from Casdoor (Certs → cert-built-in)

---

## Testing Strategy

**Unit Tests**:
- Test valid certificate parsing with real X.509 test certificate
- Test invalid PEM rejection
- Test claims parsing with complete and partial claim sets
- Test configuration validation (all required fields)
- Mock certificate for test reproducibility

**Integration Tests**:
- Test with real Casdoor instance (docker-compose)
- Obtain real JWT token from Casdoor test endpoint
- Verify signature validation passes with real token
- Test certificate rotation (update cert and verify provider picks it up)

**Test Data**:
- Generate test RSA key pair and self-signed certificate for unit tests
- Store in `internal/infrastructure/auth/testdata/`
- Generate test JWT signed with test key for integration tests

---

## Example Casdoor JWT Payload

```json
{
  "owner": "tenant1",
  "name": "Alice Wang",
  "createdTime": "2026-01-15T10:30:00Z",
  "updatedTime": "2026-01-31T14:22:00Z",
  "id": "built-in/alice",
  "type": "normal-user",
  "displayName": "Alice Wang",
  "avatar": "https://casdoor.example.com/avatar/alice.jpg",
  "email": "alice@example.com",
  "phone": "+86-13800138000",
  "isAdmin": false,
  "isGlobalAdmin": false,
  "iss": "https://casdoor.example.com",
  "sub": "built-in/alice",
  "aud": ["abc123"],
  "exp": 1738368000,
  "iat": 1738364400,
  "nbf": 1738364400
}
```

**Mapping**:
- `sub` → `UserID`
- `name` → `Username`
- `email` → `Email`
- `owner` → `Organization`
- `exp` → `ExpiresAt`

---

## Open Questions

1. **Certificate Rotation**: How frequently do Casdoor certificates rotate?
   - Current design: Manual rotation requires app restart or cache invalidation
   - Future: Periodic refetch from Casdoor JWKS endpoint?

2. **Multi-Application Support**: Should a single provider support multiple Casdoor applications?
   - Current design: One application per provider (one-to-one mapping)
   - Use case: Different projects using different Casdoor apps in same org

3. **Fallback Certificate**: Should the provider support multiple certificates (for rotation overlap)?
   - Concern: During rotation, old tokens may still be valid with old cert
   - Solution: Accept array of certificates, try all for verification?
