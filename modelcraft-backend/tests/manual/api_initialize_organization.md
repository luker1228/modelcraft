# POST /api/orgs/initialize API - Testing Guide

## Overview

This API endpoint automatically generates a unique organization name and creates an organization with the authenticated user as the owner.

## Endpoint

```
POST /api/orgs/initialize
```

## Authentication

Requires JWT authentication. The API uses the `user_id` from the JWT token.

## Request

### Headers

```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

### Body (Optional)

```json
{
  "namePrefix": "mycompany"  // Optional: custom prefix (1-12 chars, lowercase, letters/numbers/hyphens)
}
```

If no body is provided or `namePrefix` is omitted, a fully random 12-character name will be generated.

## Response

### Success (200 OK)

```json
{
  "requestId": "req-abc123...",
  "success": true,
  "message": "Organization initialized successfully",
  "organization": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "mycompany_a3k9m2",
    "displayName": "mycompany_a3k9m2"
  },
  "membership": {
    "orgId": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "user-uuid",
    "role": "owner"
  }
}
```

### Error Responses

#### 400 Bad Request - Invalid Prefix

```json
{
  "error": "name prefix must start with lowercase letter and contain only lowercase letters, numbers, and hyphens"
}
```

#### 401 Unauthorized - Missing JWT

```json
{
  "error": "User ID not found in token"
}
```

#### 500 Internal Server Error

```json
{
  "error": "Failed to initialize organization"
}
```

## Testing with cURL

### Test 1: Initialize with no prefix (fully random name)

```bash
curl -X POST http://localhost:8080/api/orgs/initialize \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

Expected result: Organization created with 12-character random name (e.g., "a3k9m2x7p5q1")

### Test 2: Initialize with custom prefix

```bash
curl -X POST http://localhost:8080/api/orgs/initialize \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"namePrefix": "mycompany"}'
```

Expected result: Organization created with name format "mycompany_a3k9m2"

### Test 3: Invalid prefix (uppercase)

```bash
curl -X POST http://localhost:8080/api/orgs/initialize \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"namePrefix": "MyCompany"}'
```

Expected result: 400 Bad Request with error message about lowercase requirement

### Test 4: Invalid prefix (too long)

```bash
curl -X POST http://localhost:8080/api/orgs/initialize \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"namePrefix": "verylongcompanyname"}'
```

Expected result: 400 Bad Request with error message about max 12 characters

## Name Format Validation

### Valid Prefixes

- `mycompany` - lowercase letters
- `my-company` - lowercase letters with hyphens
- `company123` - lowercase letters with numbers
- `a` - single character (min 1 char)
- `my-company-1` - mixed (max 12 chars)

### Invalid Prefixes

- `MyCompany` - contains uppercase
- `my_company` - contains underscore
- `my company` - contains space
- `myverylongcompanyname` - exceeds 12 characters
- `123company` - starts with number

## Generated Name Format

### Without Prefix

Format: `[a-z][a-z0-9]{11}`

Examples:
- `a3k9m2x7p5q1`
- `t8t6hicj6dxk`
- `xjb4k9p2a1zn`

### With Prefix

Format: `{prefix}_[a-z0-9]{6}`

Examples:
- `mycompany_a3k9m2`
- `my-org_x7k2j9`
- `company1_b5n8p4`

## Uniqueness Guarantee

The API ensures uniqueness through:

1. **Retry Mechanism**: Up to 3 attempts with different random suffixes
2. **Database Check**: Checks `organizations.name` uniqueness before creation
3. **Collision Handling**: Generates new random suffix on collision

With 6-character random suffix (36^6 = 2.2 billion combinations), collision rate is < 0.0001% with 10M organizations.

## Testing Sequence

### 1. Get JWT Token

```bash
# First, get a JWT token via login
curl -X POST http://localhost:8080/api/auth/token \
  -H "Content-Type: application/json" \
  -d '{"code": "YOUR_OAUTH_CODE"}'
```

Save the `accessToken` from the response.

### 2. Initialize Organization

```bash
# Use the token to initialize organization
export JWT_TOKEN="YOUR_ACCESS_TOKEN"

curl -X POST http://localhost:8080/api/orgs/initialize \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"namePrefix": "myorg"}'
```

### 3. Verify Organization Created

```bash
# Check the organization was created by listing user's memberships
curl -X GET http://localhost:8080/api/user/me \
  -H "Authorization: Bearer $JWT_TOKEN"
```

Should show the new organization in the `memberships` array.

## Integration Test Checklist

- [ ] Initialize with no prefix succeeds
- [ ] Initialize with valid prefix succeeds
- [ ] Invalid prefix (uppercase) returns 400
- [ ] Invalid prefix (too long) returns 400
- [ ] Invalid prefix (starts with number) returns 400
- [ ] Missing JWT returns 401
- [ ] Generated name is unique
- [ ] User becomes owner of new organization
- [ ] Membership record is created
- [ ] Organization appears in user's memberships list
- [ ] Retry works on collision (simulate with database lock)

## Known Limitations

1. **Max Retries**: If all 3 attempts fail (extremely rare), the API returns 500 error
2. **Prefix Length**: Maximum 12 characters to keep total name length reasonable
3. **No Custom Full Name**: Users cannot specify the full organization name directly (by design)
4. **Owner Role Only**: Initial user is always assigned the "owner" role

## Future Enhancements

- [ ] Support for custom full name (POST /api/org/create)
- [ ] Support for organization templates
- [ ] Configurable retry count
- [ ] Configurable suffix length
- [ ] Metrics for collision rate monitoring
