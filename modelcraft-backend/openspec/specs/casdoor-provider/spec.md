# casdoor-provider Specification

## Purpose
TBD - created by archiving change implement-casdoor-runtime-auth. Update Purpose after archive.
## Requirements
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

