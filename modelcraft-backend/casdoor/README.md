# Casdoor Integration Guide

This directory contains the configuration files for the Casdoor authentication service integrated with ModelCraft.

## Overview

Casdoor is an open-source Identity and Access Management (IAM) / Single-Sign-On (SSO) platform that provides:

- OAuth 2.0, OIDC, SAML authentication protocols
- Multi-tenancy support via Organizations
- Integration with 50+ identity providers (LDAP, WeChat Work, DingTalk, GitHub, Google, etc.)
- User management and JWT token issuance
- Low resource footprint (~100MB memory)

## Architecture

```
┌─────────────┐     ┌──────────────────┐     ┌──────────────────┐
│   Client    │────►│   Casdoor        │────►│   ModelCraft     │
│  (Browser)  │     │   (Port 8000)    │     │   (Port 8080)    │
│             │◄────│   - User Auth    │     │   - GraphQL API  │
│             │     │   - JWT Issue    │     │   - JWT Verify   │
└─────────────┘     └──────────────────┘     └──────────────────┘
                            │
                            ▼
                    ┌──────────────────┐
                    │  Casdoor MySQL   │
                    │  (Internal DB)   │
                    └──────────────────┘
```

## Quick Start

### 1. Start Casdoor Service

Start Casdoor along with other ModelCraft services:

```bash
# Start all services (including Casdoor)
docker compose up -d

# Or start only Casdoor and its database
docker compose up -d casdoor casdoor-db
```

### 2. Access Casdoor Admin Console

Open your browser and navigate to:

```
http://localhost:8000
```

**Default Admin Credentials:**
- Username: `admin`
- Password: `123`

**⚠️ IMPORTANT:** Change the default password immediately after first login!

### 3. Configure Casdoor

#### Step 1: Create an Organization (Tenant)

Organizations in Casdoor represent tenants in a multi-tenant system.

1. Navigate to: **Organizations** → **Add**
2. Fill in the form:
   - **Name**: `modelcraft-default` (or your organization name)
   - **Display Name**: `ModelCraft Default Organization`
   - **Website URL**: `http://localhost:8080`
3. Click **Save**

#### Step 2: Create an Application

Applications represent systems that integrate with Casdoor.

1. Navigate to: **Applications** → **Add**
2. Fill in the form:
   - **Name**: `modelcraft`
   - **Display Name**: `ModelCraft Runtime`
   - **Organization**: Select the organization you created
   - **Redirect URLs**: `http://localhost:8080/api/auth/callback`
   - **Grant Types**: Check `Authorization Code` and `Refresh Token`
   - **Token Format**: Select `JWT`
   - **Token Expire In**: `7200` (2 hours)
3. Click **Save**

After saving, you will see:
- **Client ID**: Copy this value
- **Client Secret**: Click "Show" and copy this value

#### Step 3: Get Certificate (Public Key)

Casdoor uses X.509 certificates to sign JWTs. You need to obtain the public key for JWT verification.

1. Navigate to: **Certs** → **cert-built-in**
2. Click **Copy Certificate**
3. The certificate is in PEM format:
   ```
   -----BEGIN CERTIFICATE-----
   MIIEpAIBAAKCAQEA...
   -----END CERTIFICATE-----
   ```

### 4. Configure ModelCraft

Update your `.env` file with the Casdoor configuration:

```bash
# Copy from .env.example if you haven't already
cp .env.example .env

# Edit .env and update the following variables:
CASDOOR_ENDPOINT=http://casdoor:8000  # Use service name in Docker
CASDOOR_CLIENT_ID=your-client-id-from-step-2
CASDOOR_CLIENT_SECRET=your-client-secret-from-step-2
CASDOOR_ORGANIZATION=modelcraft-default  # From Step 1
CASDOOR_APPLICATION=modelcraft  # From Step 2
CASDOOR_CERTIFICATE="-----BEGIN CERTIFICATE-----
MIIEpAIBAAKCAQEA...
-----END CERTIFICATE-----"
```

**Note:** For the certificate, you can:
- Use literal `\n` for newlines, or
- Use actual line breaks in the `.env` file (as shown above)

### 5. Restart ModelCraft

After updating the configuration, restart the ModelCraft service:

```bash
docker compose restart modelcraft
```

## Configuration Files

### `conf/app.conf`

This is the main Casdoor configuration file. Key settings:

- **httpport**: HTTP server port (default: 8000)
- **dataSourceName**: MySQL database connection string
- **origin**: CORS allowed origin (ModelCraft URL)
- **runmode**: `dev` for development, `prod` for production

You can override these settings using environment variables in `docker-compose.yml`.

## Authentication Flow

### 1. User Login Flow

```
┌──────┐     ┌──────────┐     ┌─────────┐     ┌────────────┐
│ User │     │  Client  │     │ Casdoor │     │ ModelCraft │
└──┬───┘     └────┬─────┘     └────┬────┘     └─────┬──────┘
   │              │                │                │
   │  1. Login    │                │                │
   │─────────────►│                │                │
   │              │ 2. Redirect    │                │
   │              │───────────────►│                │
   │              │                │                │
   │  3. Enter Credentials         │                │
   │──────────────────────────────►│                │
   │              │                │                │
   │              │ 4. Auth Code   │                │
   │◄─────────────│◄───────────────│                │
   │              │                │                │
   │              │ 5. Exchange for JWT             │
   │              │───────────────►│                │
   │              │                │                │
   │              │ 6. JWT Token   │                │
   │              │◄───────────────│                │
   │              │                │                │
   │              │ 7. API Request with JWT         │
   │              │────────────────────────────────►│
   │              │                │                │
   │              │                │   8. Verify JWT│
   │              │                │        ↓       │
   │              │                │   9. Response  │
   │◄─────────────│◄───────────────────────────────│
   │              │                │                │
```

### 2. JWT Verification

ModelCraft verifies JWT tokens using the Casdoor certificate:

1. Extract `Bearer` token from `Authorization` header
2. Verify signature using the RSA public key from Casdoor certificate
3. Parse JWT claims:
   - `sub`: User ID (e.g., `built-in/alice`)
   - `name`: Display name
   - `owner`: Organization (tenant identifier)
   - `email`: User email (optional)
   - `exp`: Expiration timestamp
4. Convert to `UnifiedClaims` for internal use
5. Store user context for the request

## Advanced Configuration

### Multi-Tenancy

Casdoor supports multi-tenancy through Organizations. Each organization has:

- Isolated user pool
- Independent identity providers
- Separate branding and customization

To create multiple tenants, simply create multiple Organizations in Casdoor.

### Identity Provider Integration

Casdoor can integrate with external identity providers:

#### LDAP Integration

1. Navigate to: **Providers** → **Add**
2. Select **Type**: `LDAP`
3. Fill in LDAP server details:
   - **Server**: `ldap://ldap.example.com:389`
   - **Base DN**: `dc=example,dc=com`
   - **Admin**: `cn=admin,dc=example,dc=com`
   - **Admin Password**: Your LDAP admin password
4. Click **Save**

#### WeChat Work (企业微信) Integration

1. Navigate to: **Providers** → **Add**
2. Select **Type**: `WeChatWork`
3. Fill in WeChat Work credentials:
   - **Client ID**: Your Corp ID
   - **Client Secret**: Your App Secret
4. Click **Save**

### Enabling Authentication for Projects

ModelCraft supports per-project authentication configuration. To enable authentication for a specific project:

1. Ensure `auth.runtime.enabled: true` in `configs/config.yaml`
2. Create a project-specific auth configuration via the ModelCraft API:

```bash
curl -X POST http://localhost:8080/api/projects/{projectId}/auth-config \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "casdoor",
    "enabled": true,
    "config": {
      "endpoint": "http://casdoor:8000",
      "client_id": "your-client-id",
      "client_secret": "your-client-secret",
      "organization": "your-organization",
      "application": "modelcraft",
      "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
    }
  }'
```

## Troubleshooting

### Casdoor Service Not Starting

**Check logs:**
```bash
docker compose logs casdoor
```

**Common issues:**
- Database connection failure: Ensure `casdoor-db` is healthy
- Port conflict: Change `CASDOOR_PORT` in `.env` if port 8000 is occupied

### JWT Verification Fails

**Error:** `invalid token` or `signature verification failed`

**Solutions:**
1. Verify the certificate is correctly copied (including BEGIN/END markers)
2. Ensure the certificate matches the Application in Casdoor
3. Check JWT expiration time
4. Verify the `CASDOOR_ENDPOINT` is accessible from ModelCraft container

**Test JWT manually:**
```bash
# Decode JWT (without verification)
echo "YOUR_JWT_TOKEN" | cut -d. -f2 | base64 -d | jq
```

### CORS Errors

**Error:** `CORS policy: No 'Access-Control-Allow-Origin' header`

**Solution:** Update `origin` in `casdoor/conf/app.conf` to match your client URL.

## Security Best Practices

1. **Change Default Password**: Immediately change the `admin` password after first login
2. **Use HTTPS**: In production, always use HTTPS for both Casdoor and ModelCraft
3. **Secure Secrets**: Store `CASDOOR_CLIENT_SECRET` and `CASDOOR_CERTIFICATE` in secure vaults
4. **Token Expiration**: Use short-lived access tokens (1-2 hours) and refresh tokens
5. **CORS Configuration**: Limit `origin` to trusted domains only
6. **Network Isolation**: In production, place Casdoor and its database in a private network

## Production Deployment

### Environment Variables for Production

Update your production `.env` file:

```bash
# Use HTTPS endpoint
CASDOOR_ENDPOINT=https://casdoor.yourcompany.com

# Use strong database password
CASDOOR_DB_PASSWORD=your-strong-password-here

# Update CORS origin
# Edit casdoor/conf/app.conf:
# origin = https://modelcraft.yourcompany.com
```

### Enable HTTPS

Use a reverse proxy (nginx/Traefik) or update `docker-compose.yml` to add TLS certificates:

```yaml
casdoor:
  image: casbin/casdoor:latest
  environment:
    - RUNNING_IN_DOCKER=true
  volumes:
    - ./casdoor/conf:/conf
    - ./certs:/certs  # Mount TLS certificates
  # Add TLS configuration in app.conf
```

## References

- [Casdoor Official Documentation](https://casdoor.org/docs/overview)
- [Casdoor GitHub](https://github.com/casdoor/casdoor)
- [Casdoor Go SDK](https://github.com/casdoor/casdoor-go-sdk)
- [ModelCraft Runtime Auth Design](../docs/runtime-auth-design.md)

## Support

For issues related to:
- **Casdoor setup**: Check [Casdoor Issues](https://github.com/casdoor/casdoor/issues)
- **ModelCraft integration**: Check project documentation or create an issue in ModelCraft repository
